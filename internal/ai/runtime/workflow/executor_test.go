package workflow

import (
	"context"
	"testing"

	"agent-desk/internal/ai/workflow/dsl"
	workflowregistry "agent-desk/internal/ai/workflow/registry"
	"agent-desk/internal/models"
)

func TestExecutorRoutesByConditionEdge(t *testing.T) {
	executor := NewExecutor()
	result, err := executor.Execute(context.Background(), Input{
		Definition: conditionalReplyDefinition(),
		UserMessage: models.Message{
			Content: "vip",
		},
	})
	if err != nil {
		t.Fatalf("execute workflow: %v", err)
	}
	if result.ReplyText != "VIP reply" {
		t.Fatalf("unexpected reply: %q", result.ReplyText)
	}
	assertPath(t, result.NodePath, []string{"start_1", "condition_1", "vip_reply", "send_vip", "end_1"})
}

func TestExecutorUsesDefaultEdgeWhenConditionDoesNotMatch(t *testing.T) {
	executor := NewExecutor()
	result, err := executor.Execute(context.Background(), Input{
		Definition: conditionalReplyDefinition(),
		UserMessage: models.Message{
			Content: "normal",
		},
	})
	if err != nil {
		t.Fatalf("execute workflow: %v", err)
	}
	if result.ReplyText != "Normal reply" {
		t.Fatalf("unexpected reply: %q", result.ReplyText)
	}
	assertPath(t, result.NodePath, []string{"start_1", "condition_1", "normal_reply", "send_normal", "end_1"})
}

func conditionalReplyDefinition() dsl.Definition {
	return dsl.Definition{
		SchemaVersion: 1,
		EntryNodeID:   "start_1",
		Nodes: []dsl.Node{
			{ID: "start_1", Type: workflowregistry.NodeTypeStart, Name: "Start"},
			{ID: "condition_1", Type: workflowregistry.NodeTypeCondition, Name: "Route"},
			{ID: "vip_reply", Type: workflowregistry.NodeTypeLLMReply, Name: "VIP", Config: []byte(`{"staticReply":"VIP reply"}`)},
			{ID: "normal_reply", Type: workflowregistry.NodeTypeLLMReply, Name: "Normal", Config: []byte(`{"staticReply":"Normal reply"}`)},
			{ID: "send_vip", Type: workflowregistry.NodeTypeSendReply, Name: "Send VIP", Inputs: map[string]dsl.VariableSelector{
				"replyText": {NodeID: "vip_reply", Field: "replyText"},
			}},
			{ID: "send_normal", Type: workflowregistry.NodeTypeSendReply, Name: "Send Normal", Inputs: map[string]dsl.VariableSelector{
				"replyText": {NodeID: "normal_reply", Field: "replyText"},
			}},
			{ID: "end_1", Type: workflowregistry.NodeTypeEnd, Name: "End"},
		},
		Edges: []dsl.Edge{
			{ID: "edge_start_condition", Source: "start_1", Target: "condition_1"},
			{
				ID:     "edge_condition_vip",
				Source: "condition_1",
				Target: "vip_reply",
				Condition: &dsl.Condition{
					Left:     &dsl.VariableSelector{NodeID: "start_1", Field: "userMessage"},
					Operator: "eq",
					Right:    "vip",
				},
			},
			{ID: "edge_condition_default", Source: "condition_1", Target: "normal_reply"},
			{ID: "edge_vip_send", Source: "vip_reply", Target: "send_vip"},
			{ID: "edge_normal_send", Source: "normal_reply", Target: "send_normal"},
			{ID: "edge_send_vip_end", Source: "send_vip", Target: "end_1"},
			{ID: "edge_send_normal_end", Source: "send_normal", Target: "end_1"},
		},
	}
}

func assertPath(t *testing.T, got []string, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("unexpected path length: got %#v want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected path: got %#v want %#v", got, want)
		}
	}
}
