package runtime

import (
	"context"

	"cs-agent/internal/ai/runtime/executor"
	"cs-agent/internal/pkg/utils"
)

type Service struct {
	runtime *executor.Service
	catalog *toolCatalog
	prepare *prepareService
}

func NewService() *Service {
	catalog := newToolCatalog()
	return &Service{
		runtime: executor.NewService(),
		catalog: catalog,
		prepare: newPrepareService(catalog),
	}
}

func (s *Service) Run(ctx context.Context, req Request) (*Summary, error) {
	if s == nil || s.runtime == nil || s.prepare == nil {
		return nil, nil
	}
	req.UserMessage.Content = utils.BuildRuntimeMessageText(req.UserMessage.MessageType, req.UserMessage.Content)
	if err := s.prepare.prepareToolsForRun(&req); err != nil {
		return nil, err
	}
	summary, err := s.runtime.ExecuteRun(ctx, executor.RunInput{
		Conversation: req.Conversation,
		UserMessage:  req.UserMessage,
		AIAgent:      req.AIAgent,
		AIConfig:     req.AIConfig,
		CheckPointID: req.CheckPointID,
		ToolSet:      req.ToolSet,
	})
	if err != nil {
		return toSummary(summary), err
	}
	return toSummary(summary), nil
}

func (s *Service) Resume(ctx context.Context, req ResumeRequest) (*Summary, error) {
	if s == nil || s.runtime == nil || s.prepare == nil {
		return nil, nil
	}
	if err := s.prepare.prepareToolsForResume(&req); err != nil {
		return nil, err
	}
	summary, err := s.runtime.ExecuteResume(ctx, executor.ResumeInput{
		Conversation: req.Conversation,
		AIAgent:      req.AIAgent,
		AIConfig:     req.AIConfig,
		CheckPointID: req.CheckPointID,
		ResumeData:   req.ResumeData,
		ToolSet:      req.ToolSet,
	})
	if err != nil {
		return toSummary(summary), err
	}
	return toSummary(summary), nil
}
