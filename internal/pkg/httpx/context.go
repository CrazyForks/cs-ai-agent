package httpx

import (
	"cs-agent/internal/pkg/httpx/params"
	"cs-agent/internal/pkg/openidentity"

	"github.com/gin-gonic/gin"
	"github.com/mlogclub/simple/common/strs"
)

const (
	ctxKeyExternalUser = "externalUser"
)

func SetExternalUser(ctx *gin.Context, ext *openidentity.ExternalUser) {
	ctx.Set(ctxKeyExternalUser, ext)
}

func GetExternalUser(ctx *gin.Context) *openidentity.ExternalUser {
	v, _ := ctx.Get(ctxKeyExternalUser)
	ext, _ := v.(*openidentity.ExternalUser)
	return ext
}

func GetChannelID(ctx *gin.Context) string {
	if channelID := ctx.GetHeader("X-Channel-ID"); strs.IsNotBlank(channelID) {
		return channelID
	}
	if channelID, _ := params.Get(ctx, "channelId"); strs.IsNotBlank(channelID) {
		return channelID
	}
	return ""
}
