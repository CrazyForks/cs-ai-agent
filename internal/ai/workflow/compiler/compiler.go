package compiler

import (
	"fmt"
	"strings"

	"agent-desk/internal/ai/workflow/dsl"
	workflowregistry "agent-desk/internal/ai/workflow/registry"
	"agent-desk/internal/pkg/toolx"
)

type Result struct {
	ToolCodes []string
	Appendix  string
}

func Compile(def dsl.Definition) Result {
	toolCodes := make([]string, 0)
	lines := make([]string, 0, len(def.Nodes)+2)
	if strings.TrimSpace(def.EntryNodeID) != "" {
		lines = append(lines, fmt.Sprintf("Workflow entry node: %s.", strings.TrimSpace(def.EntryNodeID)))
	}
	for _, node := range def.Nodes {
		nodeType := strings.TrimSpace(node.Type)
		if code := graphToolCodeForNodeType(nodeType); code != "" {
			toolCodes = append(toolCodes, code)
		}
		nodeName := strings.TrimSpace(node.Name)
		if nodeName == "" {
			nodeName = strings.TrimSpace(node.ID)
		}
		if nodeName == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("- %s: %s", nodeName, nodeType))
	}
	appendix := ""
	if len(lines) > 0 {
		appendix = "Published customer-service workflow:\n" + strings.Join(lines, "\n")
	}
	return Result{
		ToolCodes: toolx.NormalizeToolCodes(toolCodes),
		Appendix:  appendix,
	}
}

func graphToolCodeForNodeType(nodeType string) string {
	switch strings.TrimSpace(nodeType) {
	case workflowregistry.NodeTypeAnalyzeConversation:
		return toolx.GraphAnalyzeConversation.Code
	case workflowregistry.NodeTypePrepareTicketDraft:
		return toolx.GraphPrepareTicketDraft.Code
	case workflowregistry.NodeTypeCreateTicket:
		return toolx.GraphCreateTicketConfirm.Code
	case workflowregistry.NodeTypeHandoffToHuman:
		return toolx.GraphHandoffConversation.Code
	default:
		return ""
	}
}
