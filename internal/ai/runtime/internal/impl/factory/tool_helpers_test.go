package factory

import (
	"testing"
)

func TestBuildInstructionTraceSummary(t *testing.T) {
	got := buildInstructionTraceSummary(InstructionAssemblySummary{
		SectionTitles:     []string{"项目级规则", "当前技能上下文"},
		HasProjectRule:    true,
		HasGovernanceRule: true,
		HasAgentRule:      true,
		HasSkillRule:      true,
		HasToolRule:       false,
	})

	if len(got.SectionTitles) != 2 {
		t.Fatalf("unexpected section titles: %#v", got.SectionTitles)
	}
	if !got.HasProjectRule || !got.HasGovernanceRule || !got.HasAgentRule || !got.HasSkillRule {
		t.Fatalf("unexpected summary flags: %#v", got)
	}
	if got.HasToolRule {
		t.Fatalf("expected HasToolRule false, got %#v", got)
	}
}
