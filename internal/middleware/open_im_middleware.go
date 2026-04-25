package middleware

import (
	"cs-agent/internal/pkg/irisx"
	"cs-agent/internal/pkg/openidentity"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
)

// ChannelContextMiddleware 校验 X-Channel-Id / channelId 对应启用 web 渠道
func ChannelContextMiddleware(ctx iris.Context) {
	ext, err := openidentity.GetExternalInfo(ctx)
	if err != nil {
		ctx.StopExecution()
		_ = ctx.JSON(web.JsonError(err))
		return
	}
	irisx.SetOpenImExternalInfo(ctx, ext)
	ctx.Next()
}
