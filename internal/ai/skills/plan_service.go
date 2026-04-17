package skills

import (
	"context"
	"strings"

	"cs-agent/internal/pkg/errorsx"
)

func newPlanService() *planService {
	return &planService{}
}

type planService struct{}

// BuildExecutionPlan 构建当前请求的 Skill 执行计划。
func (s *planService) BuildExecutionPlan(execCtx context.Context, ctx RuntimeContext) (*ExecutionPlan, error) {
	if ctx.AIAgent == nil {
		return nil, errorsx.InvalidParam("AIAgent不能为空")
	}
	if ctx.AIConfig == nil {
		return nil, errorsx.InvalidParam("AIConfig不能为空")
	}
	skill, matchReason, routeTrace, err := MatchSkill(execCtx, ctx, ctx.AIAgent, ctx.AIConfig)
	if err != nil {
		return nil, err
	}

	return &ExecutionPlan{
		AIAgent:     ctx.AIAgent,
		AIConfig:    ctx.AIConfig,
		Skill:       skill,
		MatchReason: strings.TrimSpace(matchReason),
		RouteTrace:  routeTrace,
	}, nil
}
