package factory

import (
	"context"

	einoadapter "cs-agent/internal/ai/runtime/internal/impl/adapter"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/toolx"

	"github.com/cloudwego/eino/adk"
	einoskill "github.com/cloudwego/eino/adk/middlewares/skill"
)

type SkillMiddlewareService struct{}

func NewSkillMiddlewareService() *SkillMiddlewareService {
	return &SkillMiddlewareService{}
}

func (s *SkillMiddlewareService) Build(
	ctx context.Context,
	selectedSkill *models.SkillDefinition,
	toolDefinitions []einoadapter.MCPToolDefinition,
) (adk.ChatModelAgentMiddleware, error) {
	backend, err := newSelectedSkillBackend(selectedSkill, toolDefinitions)
	if err != nil {
		return nil, err
	}
	toolName := toolx.BuiltinSkill.Name
	return einoskill.NewMiddleware(ctx, &einoskill.Config{
		Backend:       backend,
		SkillToolName: &toolName,
		UseChinese:    true,
	})
}
