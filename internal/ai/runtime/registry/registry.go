package registry

import (
	"strings"

	"cs-agent/internal/pkg/toolx"

	einotool "github.com/cloudwego/eino/components/tool"
)

type Registry struct {
	tools []Tool
}

func NewRegistry(tools ...Tool) *Registry {
	return &Registry{
		tools: tools,
	}
}

func (r *Registry) Resolve(ctx Context) (*ToolSet, error) {
	ret := &ToolSet{
		Tools:     make([]einotool.BaseTool, 0, len(r.tools)),
		ToolCodes: make(map[string]string),
	}
	allowedToolCodes := makeAllowedToolCodeSet(ctx.AllowedToolCodes)
	for _, toolDef := range r.tools {
		if toolDef == nil || !toolDef.Enabled(ctx) {
			continue
		}
		toolCode := strings.TrimSpace(toolDef.Code())
		if len(allowedToolCodes) > 0 {
			if _, ok := allowedToolCodes[toolCode]; !ok && !isAlwaysAllowedToolCode(toolCode) {
				continue
			}
		}
		tool, err := toolDef.Build(ctx)
		if err != nil {
			return nil, err
		}
		if tool == nil {
			continue
		}
		toolName := strings.TrimSpace(toolDef.Name())
		if toolName == "" || toolCode == "" {
			continue
		}
		ret.Tools = append(ret.Tools, tool)
		ret.ToolCodes[toolName] = toolCode
	}
	return ret, nil
}

func makeAllowedToolCodeSet(input []string) map[string]struct{} {
	if len(input) == 0 {
		return nil
	}
	ret := make(map[string]struct{}, len(input))
	for _, item := range input {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		ret[item] = struct{}{}
	}
	return ret
}

func isAlwaysAllowedToolCode(toolCode string) bool {
	return strings.TrimSpace(toolCode) == toolx.GraphHandoffConversationToolCode
}
