package factory

import (
	"strings"

	runtimetooling "cs-agent/internal/ai/runtime/tooling"
	"cs-agent/internal/models"
)

type InstructionService struct {
	assembler                     *InstructionAssembler
	projectInstructionProvider    *ProjectInstructionProvider
	governanceInstructionProvider *GovernanceInstructionProvider
	skillInstructionProvider      *SkillInstructionProvider
	toolAppendixProvider          *ToolAppendixProvider
}

func NewInstructionService(
	assembler *InstructionAssembler,
	projectProvider *ProjectInstructionProvider,
	governanceProvider *GovernanceInstructionProvider,
	skillProvider *SkillInstructionProvider,
	toolProvider *ToolAppendixProvider,
) *InstructionService {
	if assembler == nil {
		assembler = NewInstructionAssembler()
	}
	if projectProvider == nil {
		projectProvider = NewProjectInstructionProvider()
	}
	if governanceProvider == nil {
		governanceProvider = NewGovernanceInstructionProvider()
	}
	if skillProvider == nil {
		skillProvider = NewSkillInstructionProvider()
	}
	if toolProvider == nil {
		toolProvider = NewToolAppendixProvider()
	}
	return &InstructionService{
		assembler:                     assembler,
		projectInstructionProvider:    projectProvider,
		governanceInstructionProvider: governanceProvider,
		skillInstructionProvider:      skillProvider,
		toolAppendixProvider:          toolProvider,
	}
}

func (s *InstructionService) Build(
	aiAgent *models.AIAgent,
	selectedSkill *models.SkillDefinition,
	toolDefinitions []runtimetooling.MCPToolDefinition,
	extraToolCodes map[string]string,
) InstructionAssemblyResult {
	baseInstruction := ""
	if aiAgent != nil {
		baseInstruction = strings.TrimSpace(aiAgent.SystemPrompt)
	}
	projectInstruction := ""
	governanceInstruction := ""
	skillInstruction := ""
	toolAppendices := make([]string, 0)
	if s != nil && s.projectInstructionProvider != nil {
		projectInstruction = s.projectInstructionProvider.Resolve()
	}
	if s != nil && s.governanceInstructionProvider != nil {
		governanceInstruction = s.governanceInstructionProvider.Resolve()
	}
	if s != nil && s.skillInstructionProvider != nil {
		skillInstruction = s.skillInstructionProvider.Resolve(selectedSkill)
	}
	if s != nil && s.toolAppendixProvider != nil {
		toolAppendices = s.toolAppendixProvider.Build(toolDefinitions, extraToolCodes)
	}
	assembler := NewInstructionAssembler()
	if s != nil && s.assembler != nil {
		assembler = s.assembler
	}
	return assembler.Assemble(InstructionAssemblerInput{
		AgentInstruction:      baseInstruction,
		GovernanceInstruction: governanceInstruction,
		SkillInstruction:      skillInstruction,
		ToolAppendices:        toolAppendices,
		ProjectInstruction:    projectInstruction,
	})
}
