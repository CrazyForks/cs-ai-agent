package api

import (
	"cs-agent/internal/pkg/openidentity"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
)

type CustomerController struct {
	Ctx iris.Context
}

func (c *CustomerController) PostSession_exchange() *web.JsonResult {
	channel := services.ChannelService.GetEnabledChannel(c.Ctx)
	if channel == nil {
		return web.JsonErrorMsg("接入渠道不存在或已停用")
	}
	externalUser, err := openidentity.GetExternalUser(c.Ctx, services.ChannelService.GetUserTokenSecret(channel))
	if err != nil {
		return web.JsonError(err)
	}
	resp, err := services.CustomerSessionService.Exchange(channel, *externalUser)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(resp)
}
