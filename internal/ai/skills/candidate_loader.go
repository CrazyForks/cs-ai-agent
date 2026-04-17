package skills

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
)

var newCandidateLoader = func() *candidateLoader {
	return &candidateLoader{}
}

type candidateLoader struct {
}

func (l *candidateLoader) findManualSkillDefinition(skillCode string) *models.SkillDefinition {
	return repositories.SkillDefinitionRepository.GetByCode(sqls.DB(), skillCode)
}

func (l *candidateLoader) loadCandidateSkills(aiAgent *models.AIAgent) []models.SkillDefinition {
	if aiAgent == nil {
		return nil
	}
	skillIDs := utils.SplitInt64s(aiAgent.SkillIDs)
	if len(skillIDs) == 0 {
		return nil
	}
	ret := make([]models.SkillDefinition, 0, len(skillIDs))
	for _, id := range skillIDs {
		// TODO 这里批量查询一下，批量查询返回数据的顺序需要保证和skillIDs一致
		skill := l.getSkillByID(id)
		if skill == nil || skill.Status != enums.StatusOk {
			continue
		}
		ret = append(ret, *skill)
	}
	return ret
}

func (l *candidateLoader) getSkillByID(id int64) *models.SkillDefinition {
	return repositories.SkillDefinitionRepository.Get(sqls.DB(), id)
}
