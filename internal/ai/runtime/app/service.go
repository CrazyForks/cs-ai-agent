package app

import (
	"context"

	applicationruntime "cs-agent/internal/ai/application/runtime"
)

type Service struct {
	app *applicationruntime.Service
}

func NewService() *Service {
	return &Service{
		app: applicationruntime.NewService(),
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
