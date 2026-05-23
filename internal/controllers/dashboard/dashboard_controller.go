package dashboard

import (
	"cs-agent/internal/services"

	"cs-agent/internal/pkg/httpx/params"
	"github.com/gin-gonic/gin"
	"github.com/mlogclub/simple/web"
)

type DashboardController struct {
	Ctx *gin.Context
}

func (c *DashboardController) GetOverview() *web.JsonResult {
	rangeValue, _ := params.Get(c.Ctx, "range")
	return web.JsonData(services.DashboardService.GetOverview(rangeValue))
}
