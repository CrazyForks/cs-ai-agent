package middleware

import (
	"cs-agent/internal/pkg/openidentity"
	"cs-agent/internal/services"
	"log/slog"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
)

func DashboardWsMiddleware(ctx iris.Context) {
	principal := services.AuthService.GetAuthPrincipal(ctx)
	if principal == nil {
		if _, err := services.AuthService.Authenticate(ctx); err != nil {
			_ = ctx.StopWithJSON(iris.StatusUnauthorized, web.JsonError(err))
			return
		}
		principal = services.AuthService.GetAuthPrincipal(ctx)
	}
	if err := services.WsService.UpgradeAdminConnection(ctx, principal); err != nil {
		slog.Error("upgrade admin websocket failed", "error", err, "path", ctx.Path())
		ctx.StopExecution()
		return
	}
}

func OpenImWsMiddleware(ctx iris.Context) {
	channel := services.ChannelService.GetEnabledChannel(ctx)
	if channel == nil {
		_ = ctx.StopWithJSON(iris.StatusBadRequest, web.JsonErrorMsg("接入渠道不存在或已停用"))
		return
	}

	principal := services.AuthService.GetAuthPrincipal(ctx)
	var external *openidentity.ExternalInfo
	if principal == nil {
		ext, err := openidentity.GetExternalInfo(ctx)
		if err != nil {
			_ = ctx.StopWithJSON(iris.StatusUnauthorized, web.JsonError(err))
			return
		}
		external = ext
	}
	if err := services.WsService.UpgradeUserConnection(ctx, principal, external); err != nil {
		slog.Error("upgrade open im websocket failed", "error", err, "path", ctx.Path(), "channelId", channel.ChannelID, "channel_id", channel.ID)
		ctx.StopExecution()
		return
	}
}
