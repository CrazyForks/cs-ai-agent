package executor

import (
	"context"

	publicexecutor "cs-agent/internal/ai/runtime/executor"
)

type Service struct {
	inner *publicexecutor.Service
}

func NewService() *Service {
	return &Service{
		inner: publicexecutor.NewService(),
	}
}

func (s *Service) ExecuteRun(ctx context.Context, req RunInput) (*RunResult, error) {
	if s == nil || s.inner == nil {
		return nil, nil
	}
	return s.inner.ExecuteRun(ctx, publicexecutor.RunInput(req))
}

func (s *Service) ExecuteResume(ctx context.Context, req ResumeInput) (*RunResult, error) {
	if s == nil || s.inner == nil {
		return nil, nil
	}
	return s.inner.ExecuteResume(ctx, publicexecutor.ResumeInput(req))
}
