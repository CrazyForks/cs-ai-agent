package callbacks

import (
	"encoding/json"
	"sync"
)

type RuntimeTraceCollector struct {
	mu   sync.Mutex
	Data RuntimeTraceData
}

func NewRuntimeTraceCollector() *RuntimeTraceCollector {
	ret := &RuntimeTraceCollector{}
	ret.Data.Version = "v1"
	ret.Data.Status = "started"
	return ret
}

func (c *RuntimeTraceCollector) Marshal() string {
	if c == nil {
		return ""
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	buf, err := json.Marshal(c.Data)
	if err != nil {
		return ""
	}
	return string(buf)
}

func (c *RuntimeTraceCollector) SetTooling(staticToolCodes []string, dynamicToolCodes []string, toolSearchEnabled bool) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data.Input.StaticToolCodes = append([]string(nil), staticToolCodes...)
	c.Data.Input.DynamicToolCodes = append([]string(nil), dynamicToolCodes...)
	c.Data.Input.ToolSearchEnabled = toolSearchEnabled
}

func (c *RuntimeTraceCollector) SetInstructionSummary(summary InstructionTraceSummary) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data.Instruction.SectionTitles = append([]string(nil), summary.SectionTitles...)
	c.Data.Instruction.HasAgentRule = summary.HasAgentRule
	c.Data.Instruction.HasSkillRule = summary.HasSkillRule
	c.Data.Instruction.HasToolRule = summary.HasToolRule
}

func (c *RuntimeTraceCollector) SetSkillMiddleware(enabled bool, toolName string) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data.Skill.MiddlewareEnabled = enabled
	c.Data.Skill.MiddlewareToolName = toolName
}

type SkillMetadata struct {
	Code             string
	Name             string
	Description      string
	AllowedToolCodes []string
}

func (c *RuntimeTraceCollector) SetVisibleSkills(skills map[string]SkillMetadata) {
	if c == nil || len(skills) == 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	codes := make([]string, 0, len(skills))
	for code := range skills {
		if code == "" {
			continue
		}
		codes = append(codes, code)
	}
	c.Data.Skill.VisibleCodes = append([]string(nil), codes...)
}

func (c *RuntimeTraceCollector) ActivateSkill(skill SkillMetadata, routeReason string, routeTrace string) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data.Skill.Code = skill.Code
	c.Data.Skill.Name = skill.Name
	c.Data.Skill.Description = skill.Description
	c.Data.Skill.AllowedToolCodes = append([]string(nil), skill.AllowedToolCodes...)
	c.Data.Skill.RouteReason = routeReason
	c.Data.Skill.RouteTrace = routeTrace
}

func (c *RuntimeTraceCollector) SetFilteredToolCodes(toolCodes []string) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data.Skill.FilteredToolCodes = append([]string(nil), toolCodes...)
}

func (c *RuntimeTraceCollector) SetRetrieverSummary(summary RetrieverTraceSummary) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data.Retriever.TopK = summary.TopK
	c.Data.Retriever.ScoreThreshold = summary.ScoreThreshold
	c.Data.Retriever.ContextMaxTokens = summary.ContextMaxTokens
	c.Data.Retriever.MaxContextItems = summary.MaxContextItems
	c.Data.Retriever.Count = summary.HitCount
	c.Data.Retriever.ContextCount = summary.ContextCount
	c.Data.Retriever.EmbeddingMs = summary.EmbeddingMs
	c.Data.Retriever.VectorSearchMs = summary.VectorSearchMs
	c.Data.Retriever.HydrateMs = summary.HydrateMs
	c.Data.Retriever.Policies = append([]RetrieverPolicyTraceItem(nil), summary.Policies...)
}

func (c *RuntimeTraceCollector) AddToolItem(item ToolTraceItem) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data.Tools.Count++
	c.Data.Tools.Items = append(c.Data.Tools.Items, item)
}

func (c *RuntimeTraceCollector) AddToolSearchItem(item ToolSearchTraceItem) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data.ToolSearch.Count++
	c.Data.ToolSearch.Items = append(c.Data.ToolSearch.Items, item)
}

func (c *RuntimeTraceCollector) AddGraphToolItem(item GraphToolTraceItem) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data.GraphTools.Count++
	c.Data.GraphTools.Items = append(c.Data.GraphTools.Items, item)
}
