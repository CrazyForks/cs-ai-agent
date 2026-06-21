package runtime

import (
	"encoding/json"
	"strings"

	"agent-desk/internal/ai/workflow/compiler"
	"agent-desk/internal/ai/workflow/dsl"
	"agent-desk/internal/models"
	"agent-desk/internal/pkg/enums"
	"agent-desk/internal/repositories"

	"github.com/mlogclub/simple/sqls"
)

func resolveAgentWorkflow(aiAgent models.AIAgent) (compiler.Result, bool) {
	if aiAgent.RuntimeMode != enums.AIAgentRuntimeModeWorkflow || aiAgent.WorkflowVersionID <= 0 {
		return compiler.Result{}, false
	}
	version := repositories.AIWorkflowVersionRepository.Get(sqls.DB(), aiAgent.WorkflowVersionID)
	if version == nil || version.Status != enums.StatusOk {
		return compiler.Result{}, false
	}
	var def dsl.Definition
	if err := json.Unmarshal([]byte(version.Definition), &def); err != nil {
		return compiler.Result{}, false
	}
	return compiler.Compile(def), true
}

func applyWorkflowInstruction(aiAgent models.AIAgent) models.AIAgent {
	result, ok := resolveAgentWorkflow(aiAgent)
	if !ok || strings.TrimSpace(result.Appendix) == "" {
		return aiAgent
	}
	prompt := strings.TrimSpace(aiAgent.SystemPrompt)
	appendix := strings.TrimSpace(result.Appendix)
	if prompt == "" {
		aiAgent.SystemPrompt = appendix
		return aiAgent
	}
	aiAgent.SystemPrompt = prompt + "\n\n" + appendix
	return aiAgent
}
