package eino

import (
	"context"
	"cs-agent/internal/ai/runtime/executor"
)

// TODO 这个不要了，直接使用executor不行吗？
type RuntimeExecutor struct {
	inner *executor.Service
}

func NewRuntimeExecutor() *RuntimeExecutor {
	return &RuntimeExecutor{
		inner: executor.NewService(),
	}
}

func (s *RuntimeExecutor) ExecuteRun(ctx context.Context, req executor.RunInput) (*executor.RunResult, error) {
	if s == nil || s.inner == nil {
		return nil, nil
	}
	return s.inner.ExecuteRun(ctx, executor.RunInput(req))
}

func (s *RuntimeExecutor) ExecuteResume(ctx context.Context, req executor.ResumeInput) (*executor.RunResult, error) {
	if s == nil || s.inner == nil {
		return nil, nil
	}
	return s.inner.ExecuteResume(ctx, executor.ResumeInput(req))
}
