package runtime

import (
	"context"
	"fmt"
	"strings"

	applicationruntime "agent-desk/internal/ai/application/runtime"
	"agent-desk/internal/ai/runtime/graphs"
	"agent-desk/internal/models"
	svc "agent-desk/internal/services"
)

type runtimeReplyExecutor struct{}

type runtimeReplyRunInput struct {
	Conversation models.Conversation
	Message      models.Message
	AIAgent      models.AIAgent
}

type runtimeReplyResumeInput struct {
	Conversation     models.Conversation
	Message          models.Message
	AIAgent          models.AIAgent
	PendingInterrupt *models.ConversationInterrupt
}

func newRuntimeReplyExecutor() *runtimeReplyExecutor {
	return &runtimeReplyExecutor{}
}

func (e *runtimeReplyExecutor) Run(ctx context.Context, input runtimeReplyRunInput) (*applicationruntime.Summary, error) {
	aiConfig := svc.AIConfigService.Get(input.AIAgent.AIConfigID)
	if aiConfig == nil {
		return nil, fmt.Errorf("ai config is nil")
	}
	summary, err := Service.Run(ctx, applicationruntime.Request{
		Conversation: input.Conversation,
		UserMessage:  input.Message,
		AIAgent:      input.AIAgent,
		AIConfig:     *aiConfig,
	})
	return summary, err
}

func (e *runtimeReplyExecutor) ResumePendingInterrupt(ctx context.Context, input runtimeReplyResumeInput) (*applicationruntime.Summary, error) {
	if input.PendingInterrupt == nil {
		return nil, fmt.Errorf("pending interrupt is required")
	}
	aiConfig := svc.AIConfigService.Get(input.AIAgent.AIConfigID)
	if aiConfig == nil {
		return nil, fmt.Errorf("ai config is nil")
	}
	summary, err := Service.Resume(ctx, applicationruntime.ResumeRequest{
		Conversation: input.Conversation,
		UserMessage:  input.Message,
		AIAgent:      input.AIAgent,
		AIConfig:     *aiConfig,
		CheckPointID: strings.TrimSpace(input.PendingInterrupt.CheckPointID),
		ResumeData: map[string]string{
			strings.TrimSpace(input.PendingInterrupt.InterruptID): strings.TrimSpace(input.Message.Content),
		},
	})
	return summary, err
}

func expiredInterruptSummary() *applicationruntime.Summary {
	return &applicationruntime.Summary{
		Status:    "expired",
		ReplyText: graphs.ConfirmationExpiredReply,
	}
}
