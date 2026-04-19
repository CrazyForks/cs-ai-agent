package services

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/config"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/repositories"
	"cs-agent/internal/wxwork"

	"github.com/mlogclub/simple/sqls"
	wxmessage "github.com/silenceper/wechat/v2/work/message"
)

var WxWorkNotifyService = newWxWorkNotifyService()

type wxWorkMessageSender interface {
	SendText(request wxmessage.SendTextRequest) (*wxmessage.SendResponse, error)
}

type wxWorkNotifyRecipients struct {
	ToUsers   []string
	ToParties []string
	ToTags    []string
}

type wxWorkNotifyService struct {
	senderFactory func() (wxWorkMessageSender, error)
}

func newWxWorkNotifyService() *wxWorkNotifyService {
	return &wxWorkNotifyService{
		senderFactory: func() (wxWorkMessageSender, error) {
			if !wxwork.Enabled() || wxwork.GetWorkCli() == nil {
				return nil, fmt.Errorf("wxwork is not enabled")
			}
			return wxwork.GetWorkCli().GetMessage(), nil
		},
	}
}

func (s *wxWorkNotifyService) NotifyConversationAssigned(conversationID, assigneeID int64, reason string) {
	if conversationID <= 0 {
		return
	}
	conversation := ConversationService.Get(conversationID)
	if conversation == nil {
		return
	}
	if err := s.sendToAssigneeOrDefault(assigneeID, "会话分配提醒", s.buildConversationAssignedBody(conversation, assigneeID, reason)); err != nil {
		slog.Warn("send wxwork conversation assignment notify failed",
			"conversation_id", conversationID,
			"assignee_id", assigneeID,
			"error", err,
		)
	}
}

func (s *wxWorkNotifyService) NotifyTicketCreated(ticketID int64) {
	if ticketID <= 0 {
		return
	}
	ticket := TicketService.Get(ticketID)
	if ticket == nil {
		return
	}
	if err := s.sendToAssigneeOrDefault(ticket.CurrentAssigneeID, "工单创建提醒", s.buildTicketCreatedBody(ticket)); err != nil {
		slog.Warn("send wxwork ticket created notify failed",
			"ticket_id", ticketID,
			"assignee_id", ticket.CurrentAssigneeID,
			"error", err,
		)
	}
}

func (s *wxWorkNotifyService) NotifyTicketAssigned(ticketID, assigneeID int64, reason string) {
	if ticketID <= 0 || assigneeID <= 0 {
		return
	}
	ticket := TicketService.Get(ticketID)
	if ticket == nil {
		return
	}
	if err := s.sendToAssigneeOrDefault(assigneeID, "工单指派提醒", s.buildTicketAssignedBody(ticket, assigneeID, reason)); err != nil {
		slog.Warn("send wxwork ticket assigned notify failed",
			"ticket_id", ticketID,
			"assignee_id", assigneeID,
			"error", err,
		)
	}
}

func (s *wxWorkNotifyService) Enabled() bool {
	if !wxwork.Enabled() {
		return false
	}
	return config.Current().WxWork.Notify.Enabled
}

func (s *wxWorkNotifyService) sendToAssigneeOrDefault(assigneeID int64, title, body string) error {
	if !s.Enabled() {
		return nil
	}
	recipients := s.resolveRecipientsByUserIDs([]int64{assigneeID})
	if recipients.empty() {
		recipients = s.defaultRecipients()
	}
	if recipients.empty() {
		return nil
	}
	return s.sendText(title, body, recipients)
}

func (s *wxWorkNotifyService) sendText(title, body string, recipients wxWorkNotifyRecipients) error {
	if !s.Enabled() {
		return nil
	}
	content := s.buildTextContent(title, body)
	if content == "" {
		return nil
	}
	sender, err := s.senderFactory()
	if err != nil {
		return err
	}
	cfg := config.Current().WxWork
	req := wxmessage.SendTextRequest{
		SendRequestCommon: &wxmessage.SendRequestCommon{
			ToUser:                 strings.Join(recipients.ToUsers, "|"),
			ToParty:                strings.Join(recipients.ToParties, "|"),
			ToTag:                  strings.Join(recipients.ToTags, "|"),
			AgentID:                strings.TrimSpace(cfg.AgentID),
			Safe:                   boolToInt(cfg.Notify.Safe),
			EnableDuplicateCheck:   boolToInt(cfg.Notify.EnableDuplicateCheck),
			DuplicateCheckInterval: s.normalizeDuplicateCheckInterval(cfg.Notify.DuplicateCheckInterval),
		},
		Text: wxmessage.TextField{Content: content},
	}
	_, err = sender.SendText(req)
	return err
}

func (s *wxWorkNotifyService) resolveRecipientsByUserIDs(userIDs []int64) wxWorkNotifyRecipients {
	userIDs = uniqueInt64s(userIDs)
	if len(userIDs) == 0 {
		return wxWorkNotifyRecipients{}
	}
	cfg := config.Current().WxWork
	identities := repositories.UserIdentityRepository.Find(sqls.DB(), sqls.NewCnd().
		Eq("provider", enums.ThirdProviderWxWork).
		Eq("provider_corp_id", strings.TrimSpace(cfg.CorpID)).
		Eq("status", enums.StatusOk).
		In("user_id", userIDs).
		Asc("id"))
	recipients := wxWorkNotifyRecipients{}
	for i := range identities {
		if receiver := strings.TrimSpace(identities[i].ProviderUserID); receiver != "" {
			recipients.ToUsers = append(recipients.ToUsers, receiver)
		}
	}
	recipients.ToUsers = uniqueStrings(recipients.ToUsers)
	return recipients
}

func (s *wxWorkNotifyService) defaultRecipients() wxWorkNotifyRecipients {
	cfg := config.Current().WxWork.Notify
	return wxWorkNotifyRecipients{
		ToUsers:   uniqueStrings(cfg.ToUsers),
		ToParties: uniqueStrings(cfg.ToParties),
		ToTags:    uniqueStrings(cfg.ToTags),
	}
}

func (s *wxWorkNotifyService) buildConversationAssignedBody(conversation *models.Conversation, assigneeID int64, reason string) string {
	if conversation == nil {
		return ""
	}
	lines := []string{
		fmt.Sprintf("会话ID: #%d", conversation.ID),
		fmt.Sprintf("会话主题: %s", defaultIfBlank(conversation.Subject, "-")),
		fmt.Sprintf("接入渠道: %s", enums.GetExternalSourceLabel(conversation.ExternalSource)),
		fmt.Sprintf("当前状态: %s", enums.GetIMConversationStatusLabel(conversation.Status)),
		fmt.Sprintf("处理人: %s", s.resolveUserLabel(assigneeID)),
	}
	if strings.TrimSpace(reason) != "" {
		lines = append(lines, fmt.Sprintf("分配原因: %s", strings.TrimSpace(reason)))
	}
	lines = append(lines, fmt.Sprintf("时间: %s", time.Now().Format("2006-01-02 15:04:05")))
	return strings.Join(lines, "\n")
}

func (s *wxWorkNotifyService) buildTicketCreatedBody(ticket *models.Ticket) string {
	if ticket == nil {
		return ""
	}
	lines := []string{
		fmt.Sprintf("工单号: %s", defaultIfBlank(ticket.TicketNo, fmt.Sprintf("#%d", ticket.ID))),
		fmt.Sprintf("工单标题: %s", defaultIfBlank(ticket.Title, "-")),
		fmt.Sprintf("工单来源: %s", defaultIfBlank(string(ticket.Source), "-")),
		fmt.Sprintf("当前状态: %s", enums.GetTicketStatusLabel(ticket.Status)),
	}
	if ticket.CurrentAssigneeID > 0 {
		lines = append(lines, fmt.Sprintf("处理人: %s", s.resolveUserLabel(ticket.CurrentAssigneeID)))
	}
	lines = append(lines, fmt.Sprintf("时间: %s", time.Now().Format("2006-01-02 15:04:05")))
	return strings.Join(lines, "\n")
}

func (s *wxWorkNotifyService) buildTicketAssignedBody(ticket *models.Ticket, assigneeID int64, reason string) string {
	if ticket == nil {
		return ""
	}
	lines := []string{
		fmt.Sprintf("工单号: %s", defaultIfBlank(ticket.TicketNo, fmt.Sprintf("#%d", ticket.ID))),
		fmt.Sprintf("工单标题: %s", defaultIfBlank(ticket.Title, "-")),
		fmt.Sprintf("当前状态: %s", enums.GetTicketStatusLabel(ticket.Status)),
		fmt.Sprintf("处理人: %s", s.resolveUserLabel(assigneeID)),
	}
	if strings.TrimSpace(reason) != "" {
		lines = append(lines, fmt.Sprintf("指派原因: %s", strings.TrimSpace(reason)))
	}
	lines = append(lines, fmt.Sprintf("时间: %s", time.Now().Format("2006-01-02 15:04:05")))
	return strings.Join(lines, "\n")
}

func (s *wxWorkNotifyService) resolveUserLabel(userID int64) string {
	if userID <= 0 {
		return "-"
	}
	user := UserService.Get(userID)
	if user == nil {
		return fmt.Sprintf("用户#%d", userID)
	}
	if nickname := strings.TrimSpace(user.Nickname); nickname != "" {
		return nickname
	}
	if username := strings.TrimSpace(user.Username); username != "" {
		return username
	}
	return fmt.Sprintf("用户#%d", userID)
}

func (s *wxWorkNotifyService) buildTextContent(title, body string) string {
	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)
	switch {
	case title == "" && body == "":
		return ""
	case title == "":
		return truncateRunes(body, 1024)
	case body == "":
		return truncateRunes(title, 1024)
	default:
		return truncateRunes(title+"\n\n"+body, 1024)
	}
}

func (s *wxWorkNotifyService) normalizeDuplicateCheckInterval(value int) int {
	if value <= 0 {
		return 1800
	}
	if value > 14400 {
		return 14400
	}
	return value
}

func (r wxWorkNotifyRecipients) empty() bool {
	return len(r.ToUsers) == 0 && len(r.ToParties) == 0 && len(r.ToTags) == 0
}

func uniqueStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	ret := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		item := strings.TrimSpace(value)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		ret = append(ret, item)
	}
	return ret
}

func uniqueInt64s(values []int64) []int64 {
	if len(values) == 0 {
		return nil
	}
	ret := make([]int64, 0, len(values))
	seen := make(map[int64]struct{}, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		ret = append(ret, value)
	}
	return ret
}

func truncateRunes(value string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= max {
		return string(runes)
	}
	return string(runes[:max])
}

func defaultIfBlank(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		return value
	}
	return strings.TrimSpace(fallback)
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
