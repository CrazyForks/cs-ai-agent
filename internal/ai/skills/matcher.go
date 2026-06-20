package skills

import (
	"context"

	"agent-desk/internal/models"
	"agent-desk/internal/pkg/enums"
	"agent-desk/internal/pkg/errorsx"
)

type intentTriggerConfig struct {
	Intents []string `json:"intents"`
}

// MatchSkill 对单个 SkillDefinition 执行命中判断。
func MatchSkill(execCtx context.Context, ctx RuntimeContext) (*models.SkillDefinition, string, *RouteTrace, error) {
	loader := newCandidateLoader()
	if ctx.ManualSkillDefinitionID > 0 {
		skill := loader.findManualSkillDefinition(ctx.ManualSkillDefinitionID)
		if skill == nil || skill.Status != enums.StatusOk {
			return nil, "", nil, errorsx.InvalidParamI18n("error.e0054")
		}
		return skill, "manual_skill_id", &RouteTrace{
			Status:          "manual_selected",
			SelectedSkillID: skill.ID,
		}, nil
	}

	candidates := loader.loadCandidateSkills(ctx.AIAgent)
	trace := &RouteTrace{
		Status:            "started",
		CandidateSkillIDs: make([]int64, 0, len(candidates)),
	}
	for _, item := range candidates {
		trace.CandidateSkillIDs = append(trace.CandidateSkillIDs, item.ID)
	}
	if len(candidates) == 0 {
		trace.Status = "no_candidate"
		return nil, "no_enabled_skill_bound", trace, nil
	}

	selected, routeTrace, err := routeSkillWithLLM(execCtx, ctx, candidates)
	if routeTrace != nil {
		trace.Status = routeTrace.Status
		trace.SelectedSkillID = routeTrace.SelectedSkillID
		trace.RawDecision = routeTrace.RawDecision
		trace.LatencyMs = routeTrace.LatencyMs
		trace.Error = routeTrace.Error
	}
	if err != nil {
		if trace.Error == "" {
			trace.Error = err.Error()
		}
		return nil, "route_error", trace, err
	}
	if selected == nil {
		if trace.Status == "started" {
			trace.Status = "not_matched"
		}
		return nil, "route_none", trace, nil
	}
	return selected, "llm_route", trace, nil
}
