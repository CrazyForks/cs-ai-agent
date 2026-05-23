package dashboard

import (
	"cs-agent/internal/builders"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/services"

	"cs-agent/internal/pkg/httpx/params"
	"github.com/gin-gonic/gin"
	"github.com/mlogclub/simple/web"
)

type AgentRunLogController struct {
	Ctx *gin.Context
}

func (c *AgentRunLogController) AnyList() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionConversationView); err != nil {
		return web.JsonError(err)
	}

	cnd := params.NewPagedSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "conversationId"},
		params.QueryFilter{ParamName: "messageId"},
		params.QueryFilter{ParamName: "aiAgentId"},
		params.QueryFilter{ParamName: "plannedAction"},
		params.QueryFilter{ParamName: "plannedSkillCode", Op: params.Like},
		params.QueryFilter{ParamName: "graphToolCode"},
		params.QueryFilter{ParamName: "interruptType"},
		params.QueryFilter{ParamName: "resumeSource"},
		params.QueryFilter{ParamName: "finalStatus"},
		params.QueryFilter{ParamName: "handoffReason", Op: params.Like},
		params.QueryFilter{ParamName: "finalAction"},
		params.QueryFilter{ParamName: "userMessage", Op: params.Like},
	).Desc("id")
	if hitlStatus, _ := params.Get(c.Ctx, "hitlStatus"); hitlStatus != "" && hitlStatus != "all" {
		cnd = services.AgentRunLogService.ApplyHITLStatusFilter(cnd, hitlStatus)
	}
	queryParams := params.NewQueryParams(c.Ctx)
	queryParams.Cnd = *cnd
	list, paging := services.AgentRunLogService.FindPageByParams(queryParams)
	results := make([]response.AgentRunLogResponse, 0, len(list))
	for _, item := range list {
		results = append(results, builders.BuildAgentRunLog(&item))
	}
	return web.JsonData(&web.PageResult{Results: results, Page: paging})
}

func (c *AgentRunLogController) GetBy(id int64) *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionConversationView); err != nil {
		return web.JsonError(err)
	}

	item := services.AgentRunLogService.Get(id)
	if item == nil {
		return web.JsonErrorMsg("Agent 运行日志不存在")
	}
	return web.JsonData(builders.BuildAgentRunLog(item))
}
