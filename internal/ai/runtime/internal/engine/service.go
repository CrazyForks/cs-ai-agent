package engine

import (
	"context"

	runtimeeino "cs-agent/internal/ai/infra/eino"
)

type Service struct {
	executor *runtimeeino.RuntimeExecutor
}

func NewService() *Service {
	return &Service{
		executor: runtimeeino.NewRuntimeExecutor(),
	}
}

func (s *Service) Run(ctx context.Context, req Request) (*Summary, error) {
	return s.ExecuteRun(ctx, req)
}

func (s *Service) ExecuteRun(ctx context.Context, req RunInput) (*RunResult, error) {
	return s.executor.ExecuteRun(ctx, runtimeeino.RunInput(req))
}

func (s *Service) Resume(ctx context.Context, req ResumeRequest) (*Summary, error) {
	return s.ExecuteResume(ctx, req)
}

func (s *Service) ExecuteResume(ctx context.Context, req ResumeInput) (*RunResult, error) {
	return s.executor.ExecuteResume(ctx, runtimeeino.ResumeInput(req))
}
