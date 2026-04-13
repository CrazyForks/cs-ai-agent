package runtime

import (
	"context"

	runtimeapp "cs-agent/internal/ai/runtime/app"
)

type Service struct {
	app *runtimeapp.Service
}

func NewService() *Service {
	return &Service{
		app: runtimeapp.NewService(),
	}
}

func (s *Service) Run(ctx context.Context, req Request) (*Summary, error) {
	if s == nil || s.app == nil {
		return nil, nil
	}
	return s.app.Run(ctx, req)
}

func (s *Service) Resume(ctx context.Context, req ResumeRequest) (*Summary, error) {
	if s == nil || s.app == nil {
		return nil, nil
	}
	return s.app.Resume(ctx, req)
}
