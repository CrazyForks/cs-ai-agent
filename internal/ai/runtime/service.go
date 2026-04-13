package runtime

import (
	"context"

	applicationruntime "cs-agent/internal/ai/application/runtime"
)

var Service = newService()

func newService() *service {
	return &service{
		app: applicationruntime.NewService(),
	}
}

type service struct {
	app *applicationruntime.Service
}

func (s *service) Run(ctx context.Context, req Request) (*Summary, error) {
	if s == nil || s.app == nil {
		return nil, nil
	}
	return s.app.Run(ctx, req)
}

func (s *service) Resume(ctx context.Context, req ResumeRequest) (*Summary, error) {
	if s == nil || s.app == nil {
		return nil, nil
	}
	return s.app.Resume(ctx, req)
}
