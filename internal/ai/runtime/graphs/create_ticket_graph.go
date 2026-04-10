package graphs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/services"

	componenttool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type CreateTicketGraphState struct {
	Request request.CreateTicketFromConversationRequest
}

type CreateTicketGraphInterruptInfo struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func init() {
	schema.RegisterName[CreateTicketGraphState]("cs_agent_create_ticket_graph_state")
	schema.RegisterName[CreateTicketGraphInterruptInfo]("cs_agent_create_ticket_graph_interrupt_info")
}

type CreateTicketGraph struct {
	conversation *models.Conversation
	aiAgent      *models.AIAgent
}

type Decision string

const (
	DecisionConfirm Decision = "confirm"
	DecisionCancel  Decision = "cancel"
)

func NewCreateTicketGraph(conversation *models.Conversation, aiAgent *models.AIAgent) *CreateTicketGraph {
	return &CreateTicketGraph{
		conversation: conversation,
		aiAgent:      aiAgent,
	}
}

func (g *CreateTicketGraph) Run(ctx context.Context, argumentsInJSON string) (string, error) {
	if g == nil || g.conversation == nil || g.aiAgent == nil {
		return "", fmt.Errorf("create ticket graph not initialized")
	}
	wasInterrupted, hasState, state := componenttool.GetInterruptState[CreateTicketGraphState](ctx)
	if !wasInterrupted {
		req, err := g.buildCreateRequest(argumentsInJSON)
		if err != nil {
			return "", err
		}
		info := CreateTicketGraphInterruptInfo{
			Type:    "ticket_creation_confirmation",
			Message: g.buildConfirmationPrompt(req),
		}
		return "", componenttool.StatefulInterrupt(ctx, info, CreateTicketGraphState{Request: req})
	}
	if !hasState {
		return "", fmt.Errorf("create ticket graph state missing")
	}
	isResumeTarget, hasData, resumeText := componenttool.GetResumeContext[string](ctx)
	if !isResumeTarget {
		info := CreateTicketGraphInterruptInfo{
			Type:    "ticket_creation_confirmation",
			Message: g.buildConfirmationPrompt(state.Request),
		}
		return "", componenttool.StatefulInterrupt(ctx, info, state)
	}
	if !hasData {
		info := CreateTicketGraphInterruptInfo{
			Type:    "ticket_creation_confirmation",
			Message: "请回复“确认”或“取消”。",
		}
		return "", componenttool.StatefulInterrupt(ctx, info, state)
	}
	decision := ParseConfirmationDecision(resumeText)
	switch decision {
	case DecisionConfirm:
		item, err := services.TicketService.CreateFromConversation(state.Request, g.buildAIPrincipal())
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("工单已创建，工单号：%s，标题：%s。", strings.TrimSpace(item.TicketNo), strings.TrimSpace(item.Title)), nil
	case DecisionCancel:
		return "已取消本次工单创建。", nil
	default:
		info := CreateTicketGraphInterruptInfo{
			Type:    "ticket_creation_confirmation",
			Message: "我需要你的明确确认，请直接回复“确认”或“取消”。",
		}
		return "", componenttool.StatefulInterrupt(ctx, info, state)
	}
}

func (g *CreateTicketGraph) buildCreateRequest(argumentsInJSON string) (request.CreateTicketFromConversationRequest, error) {
	req := request.CreateTicketFromConversationRequest{
		ConversationID:     g.conversation.ID,
		SyncToConversation: true,
	}
	raw := make(map[string]any)
	if strings.TrimSpace(argumentsInJSON) != "" {
		if err := json.Unmarshal([]byte(argumentsInJSON), &raw); err != nil {
			return req, fmt.Errorf("invalid create ticket arguments: %w", err)
		}
	}
	req.Title = strings.TrimSpace(getStringValue(raw, "title"))
	req.Description = strings.TrimSpace(getStringValue(raw, "description"))
	req.Priority = getInt64Value(raw, "priority")
	req.Severity = int(getInt64Value(raw, "severity"))
	if req.Title == "" {
		req.Title = strings.TrimSpace(g.conversation.Subject)
	}
	if req.Description == "" {
		req.Description = strings.TrimSpace(g.conversation.LastMessageSummary)
	}
	if strings.TrimSpace(req.Title) == "" {
		return req, fmt.Errorf("ticket title is required")
	}
	return req, nil
}

func (g *CreateTicketGraph) buildConfirmationPrompt(req request.CreateTicketFromConversationRequest) string {
	return fmt.Sprintf("我准备为你创建工单。\n标题：%s\n描述：%s\n请直接回复“确认”或“取消”。",
		strings.TrimSpace(req.Title), strings.TrimSpace(req.Description))
}

func (g *CreateTicketGraph) buildAIPrincipal() *dto.AuthPrincipal {
	username := "AI"
	if strings.TrimSpace(g.aiAgent.Name) != "" {
		username = strings.TrimSpace(g.aiAgent.Name)
	}
	return &dto.AuthPrincipal{
		UserID:   0,
		Username: username,
		Nickname: username,
	}
}

func getStringValue(data map[string]any, key string) string {
	if len(data) == 0 {
		return ""
	}
	value, ok := data[key]
	if !ok {
		return ""
	}
	text, _ := value.(string)
	return text
}

func getInt64Value(data map[string]any, key string) int64 {
	if len(data) == 0 {
		return 0
	}
	value, ok := data[key]
	if !ok {
		return 0
	}
	switch v := value.(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	case int:
		return int64(v)
	default:
		return 0
	}
}

func ParseConfirmationDecision(value string) Decision {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}
	confirmWords := []string{"确认", "是", "好的", "可以", "ok", "yes", "继续", "同意"}
	for _, item := range confirmWords {
		if strings.Contains(value, item) {
			return DecisionConfirm
		}
	}
	cancelWords := []string{"取消", "不用", "不需要", "算了", "no"}
	for _, item := range cancelWords {
		if strings.Contains(value, item) {
			return DecisionCancel
		}
	}
	return ""
}
