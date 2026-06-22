package builders

import (
	"testing"

	workflowregistry "agent-desk/internal/ai/workflow/registry"
)

func TestBuildAIWorkflowNodeSpecsIncludesVariableContracts(t *testing.T) {
	specs := BuildAIWorkflowNodeSpecs(workflowregistry.DefaultRegistry().List())

	var startFound bool
	var sendReplyFound bool
	for _, spec := range specs {
		switch spec.Type {
		case workflowregistry.NodeTypeStart:
			startFound = true
			if !hasResponseVariable(spec.OutputSchema, "userMessage") {
				t.Fatalf("expected start output userMessage, got %#v", spec.OutputSchema)
			}
		case workflowregistry.NodeTypeSendReply:
			sendReplyFound = true
			if !hasResponseVariable(spec.InputSchema, "replyText") {
				t.Fatalf("expected send_reply input replyText, got %#v", spec.InputSchema)
			}
		}
	}
	if !startFound || !sendReplyFound {
		t.Fatalf("expected start and send_reply specs in response")
	}
}

func hasResponseVariable(items []workflowregistry.VariableSpec, name string) bool {
	for _, item := range items {
		if item.Name == name {
			return true
		}
	}
	return false
}
