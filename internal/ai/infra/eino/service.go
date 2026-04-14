package eino

import (
	"context"

	runtimeexecutor "cs-agent/internal/ai/runtime/executor"
)

type RuntimeExecutor struct {
	inner *runtimeexecutor.Service
}

func NewRuntimeExecutor() *RuntimeExecutor {
	return &RuntimeExecutor{
		inner: runtimeexecutor.NewService(),
	}
}

func (s *RuntimeExecutor) ExecuteRun(ctx context.Context, req RunInput) (*RunResult, error) {
	if s == nil || s.inner == nil {
		return nil, nil
	}
	return s.inner.ExecuteRun(ctx, runtimeexecutor.RunInput(req))
}

func (s *RuntimeExecutor) ExecuteResume(ctx context.Context, req ResumeInput) (*RunResult, error) {
	if s == nil || s.inner == nil {
		return nil, nil
	}
	return s.inner.ExecuteResume(ctx, runtimeexecutor.ResumeInput(req))
}
