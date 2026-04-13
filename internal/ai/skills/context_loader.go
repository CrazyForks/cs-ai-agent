package skills

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
)

func newContextLoader() *contextLoader {
	return &contextLoader{}
}

type contextLoader struct{}

func (l *contextLoader) loadAIAgentWithConfig(aiAgentID int64) (*models.AIAgent, *models.AIConfig, error) {
	if aiAgentID <= 0 {
		return nil, nil, errorsx.InvalidParam("AIAgentID不能为空")
	}
	aiAgent := repositories.AIAgentRepository.Get(sqls.DB(), aiAgentID)
	if aiAgent == nil {
		return nil, nil, errorsx.InvalidParam("AI Agent不存在")
	}
	aiConfig := repositories.AIConfigRepository.Get(sqls.DB(), aiAgent.AIConfigID)
	if aiConfig == nil {
		return nil, nil, errorsx.InvalidParam("AI Agent关联的AI配置不存在")
	}
	return aiAgent, aiConfig, nil
}
