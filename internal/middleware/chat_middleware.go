package middleware

import (
	"cs-agent/internal/pkg/irisx"
	"cs-agent/internal/pkg/openidentity"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
)

func ExternalInfoMiddleware(ctx iris.Context) {
	ext, err := openidentity.GetExternalInfo(ctx)
	if err != nil {
		ctx.StopExecution()
		_ = ctx.JSON(web.JsonError(err))
		return
	}
	irisx.SetExternalInfo(ctx, ext)
	ctx.Next()
}
