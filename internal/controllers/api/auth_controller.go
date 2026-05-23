package api

import (
	"cs-agent/internal/pkg/config"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/httpx/params"
	"cs-agent/internal/services"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mlogclub/simple/web"
)

type AuthController struct {
	Ctx *gin.Context
}

func (c *AuthController) PostLogin() *web.JsonResult {
	cfg := config.Current()
	req := request.LoginRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}

	ret, err := services.AuthService.Login(req, cfg.Auth, c.Ctx.ClientIP(), c.Ctx.GetHeader("User-Agent"))
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(ret)
}

func (c *AuthController) GetWxwork_login() {
	loginURL, err := services.WxWorkLoginService.BuildWxWorkLoginURL(c.Ctx.Query("next"))
	if err != nil {
		c.redirectWxWorkError(err.Error())
		return
	}
	c.Ctx.Redirect(http.StatusFound, loginURL)
}

func (c *AuthController) GetWxwork_qr_login() {
	loginURL, err := services.WxWorkLoginService.BuildWxWorkQRCodeLoginURL(c.Ctx.Query("next"))
	if err != nil {
		c.redirectWxWorkError(err.Error())
		return
	}
	c.Ctx.Redirect(http.StatusFound, loginURL)
}

func (c *AuthController) GetWxwork_callback() {
	cfg := config.Current()
	ticket, next, err := services.WxWorkLoginService.LoginByWxWork(
		c.Ctx.Query("code"),
		c.Ctx.Query("state"),
		cfg.Auth,
		c.Ctx.ClientIP(),
		c.Ctx.GetHeader("User-Agent"),
	)
	if err != nil {
		c.redirectWxWorkError(err.Error())
		return
	}
	c.Ctx.Redirect(http.StatusFound, "/dashboard/login/wxwork/callback?ticket="+url.QueryEscape(ticket)+"&next="+url.QueryEscape(next))
}

func (c *AuthController) PostWxwork_exchange() *web.JsonResult {
	req := request.WxWorkExchangeRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	ret, err := services.WxWorkLoginService.ExchangeWxWorkLoginTicket(req.Ticket)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(ret)
}

func (c *AuthController) PostLogout() *web.JsonResult {
	if err := services.AuthService.Logout(c.Ctx.GetHeader("Authorization")); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *AuthController) GetProfile() *web.JsonResult {
	ret, err := services.AuthService.CurrentProfile(c.Ctx)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(ret)
}

func (c *AuthController) redirectWxWorkError(message string) {
	if idx := strings.Index(message, ": "); idx >= 0 {
		message = message[idx+2:]
	}
	c.Ctx.Redirect(http.StatusFound, "/login?wxworkError="+url.QueryEscape(message))
}
