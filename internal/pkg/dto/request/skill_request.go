package request

type SkillDefinitionListRequest struct {
	Name   string `json:"name"`
	Code   string `json:"code"`
	Status int    `json:"status"`
}

type CreateSkillDefinitionRequest struct {
	Code             string   `json:"code"`
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	Content          string   `json:"content"`
	Examples         []string `json:"examples"`
	AllowedToolCodes []string `json:"allowedToolCodes"`
	Priority         int      `json:"priority"`
	Remark           string   `json:"remark"`
}

type UpdateSkillDefinitionRequest struct {
	ID int64 `json:"id"`
	CreateSkillDefinitionRequest
}

type DeleteSkillDefinitionRequest struct {
	ID int64 `json:"id"`
}

type UpdateSkillDefinitionStatusRequest struct {
	ID     int64 `json:"id"`
	Status int   `json:"status"`
}

type SkillDebugRunRequest struct {
	AIAgentID      int64  `json:"aiAgentId"`
	ConversationID int64  `json:"conversationId"`
	SkillCode      string `json:"skillCode"`
	UserMessage    string `json:"userMessage"`
}
