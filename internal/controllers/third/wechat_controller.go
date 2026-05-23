package third

import (
	"cs-agent/internal/pkg/httpx/params"
	"cs-agent/internal/wxwork"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/silenceper/wechat/v2/work/kf"
)

type WechatController struct {
	Ctx *gin.Context
}

// GetCallback GET请求用于校验回调是否配置正确
func (c *WechatController) GetCallback() {
	cli, err := wxwork.GetWorkCli().GetKF()
	if err != nil {
		c.Ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	options := kf.SignatureOptions{}
	if err := params.ReadForm(c.Ctx, &options); err != nil {
		c.Ctx.AbortWithError(http.StatusUnauthorized, err)
		return
	}
	// 调用VerifyURL方法校验当前请求，如果合法则把解密后的内容作为响应返回给微信服务器
	echo, err := cli.VerifyURL(options)
	if err == nil {
		c.Ctx.String(http.StatusOK, echo)
	} else {
		c.Ctx.AbortWithError(http.StatusUnauthorized, err)
	}
}

// PostCallback POST请求用于接收回调
func (c *WechatController) PostCallback() {
	cli, err := wxwork.GetWorkCli().GetKF()
	if err != nil {
		c.Ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	var (
		message kf.CallbackMessage
		body    []byte
	)
	// 读取原始消息内容
	body, err = io.ReadAll(c.Ctx.Request.Body)
	if err != nil {
		c.Ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	// 解析原始数据
	message, err = cli.GetCallbackMessage(body)
	if err != nil {
		c.Ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if err := wxwork.ConsumeCallback(message); err != nil {
		c.Ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Ctx.String(http.StatusOK, "ok")
}
