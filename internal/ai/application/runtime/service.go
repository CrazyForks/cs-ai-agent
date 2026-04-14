package runtime

import (
	"context"

	runtimeeino "cs-agent/internal/ai/infra/eino"
)

type Service struct {
	runtime *runtimeeino.RuntimeExecutor
	catalog *toolCatalog
	prepare *prepareService
}

func NewService() *Service {
	catalog := newToolCatalog()
	return &Service{
		runtime: runtimeeino.NewRuntimeExecutor(),
		catalog: catalog,
		prepare: newPrepareService(catalog),
	}
}

func (s *Service) Run(ctx context.Context, req Request) (*Summary, error) {
	if s == nil || s.runtime == nil || s.prepare == nil {
		return nil, nil
	}
	selectedSkill, skillReason, skillTrace, skillErr := s.prepare.selectSkill(ctx, req)
	req.SelectedSkill = selectedSkill
	req.SkillRouteReason = skillReason
	req.SkillRouteTrace = skillTrace
	if req.SelectedSkill != nil {
		req.SelectedSkill = cloneSkillDefinition(req.SelectedSkill)
	}
	if err := s.prepare.prepareToolsForRun(&req); err != nil {
		return nil, err
	}
	summary, err := s.runtime.ExecuteRun(ctx, runtimeeino.RunInput{
		Conversation:     req.Conversation,
		UserMessage:      req.UserMessage,
		AIAgent:          req.AIAgent,
		AIConfig:         req.AIConfig,
		SelectedSkill:    req.SelectedSkill,
		SkillRouteReason: req.SkillRouteReason,
		SkillRouteTrace:  req.SkillRouteTrace,
		CheckPointID:     req.CheckPointID,
		ToolSet:          req.ToolSet,
	})
	if err != nil {
		ret := toSummary(summary)
		if ret != nil && skillErr != nil && ret.PlanReason == "" {
			ret.PlanReason = "skill_failed_fallback_runtime"
		}
		return ret, err
	}
	ret := toSummary(summary)
	if ret != nil && skillErr != nil && ret.PlanReason == "" {
		ret.PlanReason = "skill_failed_fallback_runtime"
	}
	return ret, nil
}

func (s *Service) Resume(ctx context.Context, req ResumeRequest) (*Summary, error) {
	if s == nil || s.runtime == nil || s.prepare == nil {
		return nil, nil
	}
	if err := s.prepare.prepareToolsForResume(&req); err != nil {
		return nil, err
	}
	summary, err := s.runtime.ExecuteResume(ctx, runtimeeino.ResumeInput{
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
