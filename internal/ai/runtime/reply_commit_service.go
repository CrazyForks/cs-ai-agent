package runtime

import (
	"fmt"
	"strings"
	"time"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/repositories"
	svc "cs-agent/internal/services"

	"github.com/mlogclub/simple/sqls"
)

type replyCommitService struct{}

func newReplyCommitService() *replyCommitService {
	return &replyCommitService{}
}

func (s *replyCommitService) SendAIReply(conversation models.Conversation, message models.Message, aiAgent models.AIAgent,
	replyText string, trace *aiReplyTraceData, clientPrefix string) (*models.Message, error) {
	replyText = strings.TrimSpace(replyText)
	if replyText == "" {
		return nil, nil
	}
	commitStartedAt := time.Now()
	replyMessage, err := svc.MessageService.SendAIMessage(
		conversation.ID,
		aiAgent.ID,
		fmt.Sprintf("%s_%d", strings.TrimSpace(clientPrefix), message.ID),
		enums.IMMessageTypeText,
		replyText,
		"",
		s.buildAIPrincipal(aiAgent),
	)
	if trace != nil {
		trace.CommitMs = time.Since(commitStartedAt).Milliseconds()
		trace.ReplySent = err == nil && replyMessage != nil
		if replyMessage != nil {
			trace.ReplyMessageID = replyMessage.ID
		}
	}
	return replyMessage, err
}

func (s *replyCommitService) IncrementAIReplyRounds(conversationID int64, nextRounds int, aiAgentName string) error {
	return repositories.ConversationRepository.Updates(sqls.DB(), conversationID, map[string]any{
		"ai_reply_rounds":  nextRounds,
		"update_user_id":   0,
		"update_user_name": strings.TrimSpace(aiAgentName),
		"updated_at":       time.Now(),
	})
}

func (s *replyCommitService) buildAIPrincipal(aiAgent models.AIAgent) *dto.AuthPrincipal {
	username := "AI"
	if strings.TrimSpace(aiAgent.Name) != "" {
		username = aiAgent.Name
	}
	return &dto.AuthPrincipal{
		UserID:   0,
		Username: username,
		Nickname: username,
	}
}
