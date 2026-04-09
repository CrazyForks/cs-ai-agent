package skills

import (
	"context"
	"strings"
	"time"

	"cs-agent/internal/ai"
	"cs-agent/internal/pkg/errorsx"
)

func executeByPlan(ctx context.Context, plan *ExecutionPlan, runtimeCtx RuntimeContext) (string, *ExecutionTrace, error) {
	if plan == nil || plan.Skill == nil {
		return "", nil, nil
	}
	trace := &ExecutionTrace{
		Status:        "started",
		ExecutionMode: "content",
	}
	replyText, err := executeContent(ctx, plan, runtimeCtx, trace)
	return replyText, trace, err
}

func executeContent(ctx context.Context, plan *ExecutionPlan, runtimeCtx RuntimeContext, trace *ExecutionTrace) (string, error) {
	if plan == nil || plan.Skill == nil {
		return "", nil
	}
	if plan.AIConfig == nil {
		return "", errorsx.InvalidParam("Skill 关联的 AI 配置不可用")
	}
	systemPrompt := strings.TrimSpace(plan.Skill.Content)
	if systemPrompt == "" {
		return "", errorsx.InvalidParam("Skill Content 不能为空")
	}
	userPrompt := strings.TrimSpace(runtimeCtx.UserMessage)
	if userPrompt == "" {
		return "", errorsx.InvalidParam("用户消息不能为空")
	}
	promptTrace := &PromptTrace{Status: "started"}
	if trace != nil {
		trace.Prompt = promptTrace
	}
	startedAt := time.Now()
	result, err := ai.LLM.ChatWithConfig(ctx, plan.AIConfig, systemPrompt, userPrompt)
	promptTrace.LatencyMs = time.Since(startedAt).Milliseconds()
	if err != nil {
		promptTrace.Status = "error"
		promptTrace.Error = err.Error()
		if trace != nil {
			trace.Status = "error"
		}
		return "", err
	}
	promptTrace.Status = "ok"
	promptTrace.ModelName = result.ModelName
	promptTrace.PromptTokens = result.PromptTokens
	promptTrace.CompletionTokens = result.CompletionTokens
	if trace != nil {
		trace.Status = "ok"
	}
	return strings.TrimSpace(result.Content), nil
}
