package compiler

import (
	"testing"

	"agent-desk/internal/ai/workflow/dsl"
	workflowregistry "agent-desk/internal/ai/workflow/registry"
	"agent-desk/internal/pkg/toolx"
)

func TestCompileMapsWorkflowNodesToGraphTools(t *testing.T) {
	result := Compile(dsl.Definition{
		EntryNodeID: "start",
		Nodes: []dsl.Node{
			{ID: "start", Type: workflowregistry.NodeTypeStart, Name: "Start"},
			{ID: "analyze", Type: workflowregistry.NodeTypeAnalyzeConversation, Name: "Analyze"},
			{ID: "draft", Type: workflowregistry.NodeTypePrepareTicketDraft, Name: "Draft"},
			{ID: "create", Type: workflowregistry.NodeTypeCreateTicket, Name: "Create"},
			{ID: "handoff", Type: workflowregistry.NodeTypeHandoffToHuman, Name: "Handoff"},
		},
	})
	want := []string{
		toolx.GraphAnalyzeConversation.Code,
		toolx.GraphPrepareTicketDraft.Code,
		toolx.GraphCreateTicketConfirm.Code,
		toolx.GraphHandoffConversation.Code,
	}
	if len(result.ToolCodes) != len(want) {
		t.Fatalf("expected %d tool codes, got %d: %#v", len(want), len(result.ToolCodes), result.ToolCodes)
	}
	for i, item := range want {
		if result.ToolCodes[i] != item {
			t.Fatalf("tool code[%d] = %s, want %s", i, result.ToolCodes[i], item)
		}
	}
	if result.Appendix == "" {
		t.Fatalf("expected workflow appendix")
	}
}
