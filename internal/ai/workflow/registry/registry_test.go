package registry

import "testing"

func TestDefaultRegistryExposesStartOutputs(t *testing.T) {
	spec, ok := DefaultRegistry().Get(NodeTypeStart)
	if !ok {
		t.Fatalf("start node spec not found")
	}
	if !hasVariable(spec.OutputSchema, "userMessage", VariableTypeString) {
		t.Fatalf("expected start output userMessage:string, got %#v", spec.OutputSchema)
	}
	if !hasVariable(spec.OutputSchema, "knowledgeBaseIds", VariableTypeIntegerArray) {
		t.Fatalf("expected start output knowledgeBaseIds:array<int>, got %#v", spec.OutputSchema)
	}
}

func TestDefaultRegistryExposesKnowledgeRetrieveInputsAndOutputs(t *testing.T) {
	spec, ok := DefaultRegistry().Get(NodeTypeKnowledgeRetrieve)
	if !ok {
		t.Fatalf("knowledge_retrieve node spec not found")
	}
	if !hasRequiredVariable(spec.InputSchema, "query", VariableTypeString) {
		t.Fatalf("expected knowledge_retrieve required input query:string, got %#v", spec.InputSchema)
	}
	if !hasVariable(spec.OutputSchema, "items", VariableTypeObjectArray) {
		t.Fatalf("expected knowledge_retrieve output items:array<object>, got %#v", spec.OutputSchema)
	}
}

func TestDefaultRegistryExposesSendReplyRequiredInput(t *testing.T) {
	spec, ok := DefaultRegistry().Get(NodeTypeSendReply)
	if !ok {
		t.Fatalf("send_reply node spec not found")
	}
	if !hasRequiredVariable(spec.InputSchema, "replyText", VariableTypeString) {
		t.Fatalf("expected send_reply required input replyText:string, got %#v", spec.InputSchema)
	}
	if !hasVariable(spec.OutputSchema, "sent", VariableTypeBoolean) {
		t.Fatalf("expected send_reply output sent:boolean, got %#v", spec.OutputSchema)
	}
}

func hasRequiredVariable(items []VariableSpec, name string, variableType VariableType) bool {
	for _, item := range items {
		if item.Name == name && item.Type == variableType && item.Required {
			return true
		}
	}
	return false
}

func hasVariable(items []VariableSpec, name string, variableType VariableType) bool {
	for _, item := range items {
		if item.Name == name && item.Type == variableType {
			return true
		}
	}
	return false
}
