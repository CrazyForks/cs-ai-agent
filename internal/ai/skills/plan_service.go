package skills

import (
	"context"
	"strings"
)

func newPlanService() *planService {
	return &planService{
		loader: newContextLoader(),
	}
}

type planService struct {
	loader *contextLoader
}

// BuildExecutionPlan 构建当前请求的 Skill 执行计划。
func (s *planService) BuildExecutionPlan(execCtx context.Context, ctx RuntimeContext) (*ExecutionPlan, error) {
	if s.loader == nil {
		s.loader = newContextLoader()
	}
	aiAgent, aiConfig, err := s.loader.loadAIAgentWithConfig(ctx.AIAgentID)

	skill, matchReason, routeTrace, err := MatchSkill(execCtx, ctx, aiAgent, aiConfig)
	if err != nil {
		return nil, err
	}

	return &ExecutionPlan{
		AIAgent:     aiAgent,
		AIConfig:    aiConfig,
		Skill:       skill,
		MatchReason: strings.TrimSpace(matchReason),
		RouteTrace:  routeTrace,
	}, nil
}
