package api

import (
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type ChannelController struct {
	Ctx iris.Context
}

func (c *ChannelController) AnyConfig() *web.JsonResult {
	channel := services.ChannelService.GetEnabledChannel(c.Ctx)
	if channel == nil {
		return web.JsonErrorMsg("接入渠道未初始化")
	}
	cfg, externalSource, err := resolveWidgetConfig(channel.ChannelType, channel.ConfigJSON)
	if err != nil {
		return web.JsonError(err)
	}

	ret := response.WidgetConfigResponse{
		ChannelID:      channel.ChannelID,
		ChannelType:    channel.ChannelType,
		ExternalSource: externalSource,
		Title:          cfg.Title,
		Subtitle:       cfg.Subtitle,
		ThemeColor:     cfg.ThemeColor,
		Position:       cfg.Position,
		Width:          cfg.Width,
	}
	return web.JsonData(ret)
}

func (c *ChannelController) AnyWechat_mpOauthAuthorize() *web.JsonResult {
	channelID, _ := params.Get(c.Ctx, "channelId")
	returnPath, _ := params.Get(c.Ctx, "returnPath")
	redirectURL, err := services.ChannelService.BuildWechatMPOAuthURL(c.Ctx, channelID, returnPath)
	if err != nil {
		return web.JsonError(err)
	}
	c.Ctx.Redirect(redirectURL)
	return nil
}

func (c *ChannelController) AnyWechat_mpOauthCallback() *web.JsonResult {
	code, _ := params.Get(c.Ctx, "code")
	state, _ := params.Get(c.Ctx, "state")
	result, err := services.ChannelService.CompleteWechatMPOAuth(c.Ctx.Request().Context(), code, state)
	if err != nil {
		return web.JsonError(err)
	}
	c.Ctx.Redirect(services.BuildWechatMPChatRedirectURL(c.Ctx, result))
	return nil
}

type webLikeWidgetConfig struct {
	Title      string
	Subtitle   string
	ThemeColor string
	Position   string
	Width      string
}

func resolveWidgetConfig(channelType, rawConfig string) (*webLikeWidgetConfig, string, error) {
	switch channelType {
	case enums.ChannelTypeWeb:
		cfg, err := services.ChannelService.ParseWebChannelConfig(rawConfig)
		if err != nil {
			return nil, "", err
		}
		return &webLikeWidgetConfig{
			Title:      cfg.Title,
			Subtitle:   cfg.Subtitle,
			ThemeColor: cfg.ThemeColor,
			Position:   cfg.Position,
			Width:      cfg.Width,
		}, string(enums.ExternalSourceWebChat), nil
	case enums.ChannelTypeWechatMP:
		cfg, err := services.ChannelService.ParseWechatMPChannelConfig(rawConfig)
		if err != nil {
			return nil, "", err
		}
		return &webLikeWidgetConfig{
			Title:      cfg.Title,
			Subtitle:   cfg.Subtitle,
			ThemeColor: cfg.ThemeColor,
			Position:   cfg.Position,
			Width:      cfg.Width,
		}, string(enums.ExternalSourceWechatMP), nil
	default:
		return nil, "", errorsx.InvalidParam("该渠道不支持开放客服配置")
	}
}
