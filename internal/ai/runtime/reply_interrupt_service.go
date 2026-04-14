package runtime

import (
	"context"
	"strings"

	applicationruntime "cs-agent/internal/ai/application/runtime"
	"cs-agent/internal/ai/runtime/graphs"
	"cs-agent/internal/models"
	svc "cs-agent/internal/services"
)

type replyInterruptService struct{}

func newReplyInterruptService() *replyInterruptService {
	return &replyInterruptService{}
}

func (s *replyInterruptService) ResumePendingInterrupt(ctx context.Context, owner *aiReplyService, conversation models.Conversation, message models.Message, aiAgent models.AIAgent,
	pendingInterrupt *models.ConversationInterrupt, trace *aiReplyTraceData, summaryRef **applicationruntime.Summary) error {
	if pendingInterrupt == nil || owner == nil || owner.executor == nil {
		return nil
	}
	summary, err := owner.executor.ResumePendingInterrupt(ctx, conversation, message, aiAgent, pendingInterrupt, trace)
	*summaryRef = summary
	if err != nil {
		if isCheckpointMissingError(err) {
			summary = expiredInterruptSummary()
			*summaryRef = summary
			trace.Status = "interrupt_expired"
			trace.FinalAction = "expired"
			replyMessage, expireErr := owner.commit.CommitAIReply(conversation, message, aiAgent, summary.ReplyText, trace, "ai_interrupt_expired")
			if expireErr != nil {
				return expireErr
			}
			lastResumeMessageID := int64(0)
			if replyMessage != nil {
				lastResumeMessageID = replyMessage.ID
			}
			if expireMarkErr := svc.ConversationInterruptService.MarkExpired(pendingInterrupt.ID, lastResumeMessageID); expireMarkErr != nil {
				return expireMarkErr
			}
			return nil
		}
		return err
	}
	if summary != nil && summary.Interrupted {
		return s.HandleInterruptedResume(owner, conversation, message, aiAgent, pendingInterrupt, summary, trace)
	}
	if summary != nil && strings.TrimSpace(summary.ReplyText) != "" {
		replyMessage, err := owner.commit.CommitAIReply(conversation, message, aiAgent, summary.ReplyText, trace, "ai_resume")
		if err != nil {
			return err
		}
		replyMessageID := int64(0)
		if replyMessage != nil {
			replyMessageID = replyMessage.ID
		}
		if graphs.IsCancellationReply(summary.ReplyText) {
			return svc.ConversationInterruptService.MarkCancelled(pendingInterrupt.ID, replyMessageID)
		}
		return svc.ConversationInterruptService.MarkResolved(pendingInterrupt.ID, replyMessageID)
	}
	return svc.ConversationInterruptService.MarkResolved(pendingInterrupt.ID, 0)
}

func (s *replyInterruptService) HandleInterruptedSummary(owner *aiReplyService, conversation models.Conversation, message models.Message, aiAgent models.AIAgent,
	summary *applicationruntime.Summary, trace *aiReplyTraceData) error {
	if owner == nil {
		return nil
	}
	pending := buildConversationInterrupt(conversation, message, aiAgent, summary)
	if err := svc.ConversationInterruptService.CreateOrUpdatePending(pending); err != nil {
		return err
	}
	pending = svc.ConversationInterruptService.GetByCheckPointID(summary.CheckPointID)
	replyText := resolveInterruptPrompt(summary)
	replyMessage, err := owner.commit.CommitAIReply(conversation, message, aiAgent, replyText, trace, "ai_interrupt")
	if err != nil {
		return err
	}
	if replyMessage != nil && pending != nil {
		return svc.ConversationInterruptService.MarkPendingAgain(pending.ID, pending.InterruptID, replyText, replyMessage.ID)
	}
	return nil
}

func (s *replyInterruptService) HandleInterruptedResume(owner *aiReplyService, conversation models.Conversation, message models.Message, aiAgent models.AIAgent,
	pendingInterrupt *models.ConversationInterrupt, summary *applicationruntime.Summary, trace *aiReplyTraceData) error {
	if pendingInterrupt == nil || owner == nil {
		return nil
	}
	replyText := resolveInterruptPrompt(summary)
	replyMessage, err := owner.commit.CommitAIReply(conversation, message, aiAgent, replyText, trace, "ai_interrupt_resume")
	if err != nil {
		return err
	}
	if replyMessage != nil {
		return svc.ConversationInterruptService.MarkPendingAgain(pendingInterrupt.ID, firstInterruptID(summary), replyText, replyMessage.ID)
	}
	return nil
}
