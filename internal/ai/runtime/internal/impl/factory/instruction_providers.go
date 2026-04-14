package factory

import (
	"os"
	"path/filepath"
	"strings"

	einoadapter "cs-agent/internal/ai/runtime/internal/impl/adapter"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/toolx"
)

type ProjectInstructionProvider struct {
	fileName string
}

// TODO 这个要读取AGENTS.md文件，后面考虑还要不要
func NewProjectInstructionProvider() *ProjectInstructionProvider {
	return &ProjectInstructionProvider{fileName: "AGENTS.md"}
}

func (p *ProjectInstructionProvider) Resolve() string {
	if text := p.loadFromFile(); text != "" {
		return text
	}
	return strings.TrimSpace(DefaultProjectInstruction)
}

func (p *ProjectInstructionProvider) loadFromFile() string {
	path := p.resolvePath()
	if path == "" {
		return ""
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func (p *ProjectInstructionProvider) resolvePath() string {
	fileName := "AGENTS.md"
	if p != nil && strings.TrimSpace(p.fileName) != "" {
		fileName = strings.TrimSpace(p.fileName)
	}
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for {
		candidate := filepath.Join(dir, fileName)
		if stat, statErr := os.Stat(candidate); statErr == nil && !stat.IsDir() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

type ToolAppendixProvider struct{}

func NewToolAppendixProvider() *ToolAppendixProvider {
	return &ToolAppendixProvider{}
}

func (p *ToolAppendixProvider) Build(selectedSkill *models.SkillDefinition, toolDefinitions []einoadapter.MCPToolDefinition, extraToolCodes map[string]string) []string {
	appendixParts := make([]string, 0, 2)
	if skillInstruction := buildSelectedSkillActivationInstruction(selectedSkill); skillInstruction != "" {
		appendixParts = append(appendixParts, skillInstruction)
	}
	toolCodes := make([]string, 0, len(toolDefinitions)+len(extraToolCodes))
	for _, item := range toolDefinitions {
		toolCodes = append(toolCodes, item.ToolCode)
	}
	for _, item := range extraToolCodes {
		toolCodes = append(toolCodes, item)
	}
	appendixParts = append(appendixParts, toolx.BuildToolAppendicesForCodes(len(toolDefinitions) > 0, toolCodes)...)
	return appendixParts
}
