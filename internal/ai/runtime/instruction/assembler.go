package instruction

import "strings"

const defaultGovernanceInstruction = `
你正在一个有明确工程约束的客服系统中工作。
执行时必须严格遵守当前注入的项目规则、Agent 规则和技能规则。
如果存在工具白名单限制，只能调用当前允许的工具；信息不足时优先追问，不要伪造事实或跳过必要确认。
`

type Assembler struct {
	governanceInstruction string
}

type AssemblerInput struct {
	AgentInstruction      string
	GovernanceInstruction string
	SkillInstruction      string
	ToolAppendices        []string
	ProjectInstruction    string
}

type AssemblySummary struct {
	SectionTitles     []string
	HasProjectRule    bool
	HasGovernanceRule bool
	HasAgentRule      bool
	HasSkillRule      bool
	HasToolRule       bool
}

type AssemblyResult struct {
	Text    string
	Summary AssemblySummary
}

func NewAssembler() *Assembler {
	return &Assembler{governanceInstruction: strings.TrimSpace(defaultGovernanceInstruction)}
}

func (a *Assembler) Build(input AssemblerInput) string {
	return a.Assemble(input).Text
}

func (a *Assembler) Assemble(input AssemblerInput) AssemblyResult {
	parts := make([]string, 0, 5)
	summary := AssemblySummary{SectionTitles: make([]string, 0, 5)}
	projectInstruction := strings.TrimSpace(input.ProjectInstruction)
	if projectInstruction == "" {
		projectInstruction = strings.TrimSpace(DefaultProjectInstruction)
	}
	if projectInstruction != "" {
		parts = append(parts, buildInstructionSection("项目级规则", projectInstruction))
		summary.HasProjectRule = true
		summary.SectionTitles = append(summary.SectionTitles, "项目级规则")
	}
	governanceInstruction := strings.TrimSpace(input.GovernanceInstruction)
	if governanceInstruction == "" && a != nil {
		governanceInstruction = strings.TrimSpace(a.governanceInstruction)
	}
	if governanceInstruction != "" {
		parts = append(parts, buildInstructionSection("系统治理规则", governanceInstruction))
		summary.HasGovernanceRule = true
		summary.SectionTitles = append(summary.SectionTitles, "系统治理规则")
	}
	if agentInstruction := strings.TrimSpace(input.AgentInstruction); agentInstruction != "" {
		parts = append(parts, buildInstructionSection("Agent 规则", agentInstruction))
		summary.HasAgentRule = true
		summary.SectionTitles = append(summary.SectionTitles, "Agent 规则")
	}
	if skillInstruction := strings.TrimSpace(input.SkillInstruction); skillInstruction != "" {
		parts = append(parts, buildInstructionSection("当前技能上下文", skillInstruction))
		summary.HasSkillRule = true
		summary.SectionTitles = append(summary.SectionTitles, "当前技能上下文")
	}
	if appendix := buildToolAppendix(input.ToolAppendices); appendix != "" {
		parts = append(parts, buildInstructionSection("工具补充规则", appendix))
		summary.HasToolRule = true
		summary.SectionTitles = append(summary.SectionTitles, "工具补充规则")
	}
	return AssemblyResult{
		Text:    strings.TrimSpace(strings.Join(parts, "\n\n")),
		Summary: summary,
	}
}

func buildInstructionSection(title, body string) string {
	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)
	if body == "" {
		return ""
	}
	if title == "" {
		return body
	}
	return title + "：\n" + body
}

func buildToolAppendix(input []string) string {
	if len(input) == 0 {
		return ""
	}
	parts := make([]string, 0, len(input))
	for _, item := range input {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		parts = append(parts, item)
	}
	return strings.TrimSpace(strings.Join(parts, "\n\n"))
}
