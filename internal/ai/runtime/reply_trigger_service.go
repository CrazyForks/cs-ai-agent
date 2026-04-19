package runtime

import (
	"context"
	"log/slog"
	"strings"
	"time"

	applicationruntime "cs-agent/internal/ai/application/runtime"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
	svc "cs-agent/internal/services"
)

func (s *aiReplyService) resolveReplyTimeout(aiAgent models.AIAgent) time.Duration {
	if aiAgent.ReplyTimeoutSeconds <= 0 {
		return time.Duration(defaultAIReplyAsyncTimeoutSeconds) * time.Second
	}
	if aiAgent.ReplyTimeoutSeconds > maxAIReplyAsyncTimeoutSeconds {
		return time.Duration(maxAIReplyAsyncTimeoutSeconds) * time.Second
	}
	return time.Duration(aiAgent.ReplyTimeoutSeconds) * time.Second
}

func (s *aiReplyService) TriggerReplyAsync(conversation models.Conversation, message models.Message) {
	go func() {
		aiAgent := svc.AIAgentService.Get(conversation.AIAgentID)
		if aiAgent == nil || aiAgent.Status != enums.StatusOk {
			return
		}
		startedAt := time.Now()
		timeout := s.resolveReplyTimeout(*aiAgent)
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := s.TriggerReply(ctx, conversation, message, *aiAgent); err != nil {
			slog.Error("failed to trigger ai reply",
				"message_id", message.ID,
				"timeout_ms", timeout.Milliseconds(),
				"elapsed_ms", time.Since(startedAt).Milliseconds(),
				"error", err)
		}
	}()
}

func (s *aiReplyService) TriggerReply(ctx context.Context, conversation models.Conversation, message models.Message, aiAgent models.AIAgent) (retErr error) {
	startedAt := time.Now()
	trace := &aiReplyTraceData{Status: "started"}
	var summary *applicationruntime.Summary
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.eligibility != nil && !s.eligibility.CanReply(conversation, message, aiAgent) {
		return nil
	}
	defer func() {
		s.runlog.Write(replyRunLogInput{
			StartedAt:    startedAt,
			Message:      message,
			Conversation: conversation,
			AIAgent:      aiAgent,
			Question:     message.Content,
			RunErr:       retErr,
			Trace:        trace,
			Summary:      summary,
		})
	}()
	if pendingInterrupt := svc.ConversationInterruptService.FindLatestPendingByConversationID(conversation.ID); pendingInterrupt != nil {
		return s.resumePendingInterrupt(ctx, conversation, message, aiAgent, pendingInterrupt, trace, &summary)
	}
	return s.executeReply(ctx, conversation, message, aiAgent, trace, &summary)
}

func (s *aiReplyService) resumePendingInterrupt(ctx context.Context, conversation models.Conversation, message models.Message, aiAgent models.AIAgent,
	pendingInterrupt *models.ConversationInterrupt, trace *aiReplyTraceData, summaryRef **applicationruntime.Summary) error {
	return s.interrupts.ResumePendingInterrupt(ctx, s, conversation, message, aiAgent, pendingInterrupt, trace, summaryRef)
}

func (s *aiReplyService) executeReply(ctx context.Context, conversation models.Conversation, message models.Message, aiAgent models.AIAgent,
	trace *aiReplyTraceData, summaryRef **applicationruntime.Summary) error {
	summary, err := s.executor.Run(ctx, runtimeReplyRunInput{
		Conversation: conversation,
		Message:      message,
		AIAgent:      aiAgent,
		Trace:        trace,
	})
	if summaryRef != nil {
		*summaryRef = summary
	}
	if err != nil {
		return err
	}
	if summary != nil && summary.Interrupted {
		return s.interrupts.HandleInterruptedSummary(s, conversation, message, aiAgent, summary, trace)
	}
	if summary != nil && strings.TrimSpace(summary.ReplyText) != "" {
		replyMessage, err := s.commit.CommitAIReply(replyCommitInput{
			Conversation: conversation,
			Message:      message,
			AIAgent:      aiAgent,
			ReplyText:    summary.ReplyText,
			Trace:        trace,
			ClientPrefix: "ai_reply",
		})
		if err != nil {
			return err
		}
		trace.ReplySent = replyMessage != nil
	}
	return nil
}
