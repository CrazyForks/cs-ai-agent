package bootstrap

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cs-agent/internal/ai/mcps"
	_ "cs-agent/internal/ai/runtime"
	"cs-agent/internal/controllers/api"
	"cs-agent/internal/controllers/dashboard"
	"cs-agent/internal/controllers/third"
	"cs-agent/internal/middleware"
	"cs-agent/internal/pkg/config"
	"cs-agent/internal/pkg/ginx"
	"cs-agent/internal/services"

	"github.com/gin-gonic/gin"

	_ "cs-agent/internal/services/wx_callback_handlers"
)

func NewServer() (*gin.Engine, error) {
	cfg := config.Current()
	app := gin.New()
	app.Use(corsMiddleware())
	app.Use(gin.Recovery())
	app.Use(requestLogMiddleware())
	app.Use(maxBodySizeMiddleware(cfg.Storage.MaxRequestBodySizeBytes()))

	addRouter(app)

	app.StaticFS(cfg.Storage.Local.BaseURL, http.Dir(cfg.Storage.Local.Root))
	registerDashboardStatic(app, "web/out")

	return app, nil
}

func corsMiddleware() gin.HandlerFunc {
	allowHeaders := "Origin, Content-Type, Accept, Authorization, X-Requested-With, X-Guest-Id, X-Channel-Id, X-External-Id, X-External-Name, X-Customer-Session-Token, X-Customer-Session-Expires-At"
	exposeHeaders := "Content-Length, Content-Type, Authorization, X-Guest-Id, X-Channel-Id, X-External-Id, X-External-Name, X-Customer-Session-Token, X-Customer-Session-Expires-At"
	return func(ctx *gin.Context) {
		if isWebsocketUpgrade(ctx) {
			ctx.Next()
			return
		}
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Headers", allowHeaders)
		ctx.Header("Access-Control-Expose-Headers", exposeHeaders)
		ctx.Header("Access-Control-Max-Age", "600")
		if ctx.Request.Method == http.MethodOptions {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}
		ctx.Next()
	}
}

func requestLogMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		path := ctx.Request.URL.Path
		method := ctx.Request.Method
		ctx.Next()

		slog.Info("http request",
			"method", method,
			"path", path,
			"status", ctx.Writer.Status(),
			"elapsed", time.Since(start).Milliseconds(),
			"clientIp", ctx.ClientIP(),
		)
	}
}

func maxBodySizeMiddleware(limit int64) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, limit)
		ctx.Next()
	}
}

func isWebsocketUpgrade(ctx *gin.Context) bool {
	if !strings.EqualFold(ctx.GetHeader("Upgrade"), "websocket") {
		return false
	}
	return strings.Contains(strings.ToLower(ctx.GetHeader("Connection")), "upgrade")
}

func addRouter(app *gin.Engine) {
	app.Any("/api/mcp", gin.WrapH(mcps.NewHTTPHandler()))

	apiGroup := app.Group("/api")
	ginx.HandleController(apiGroup, "/auth", new(api.AuthController))
	ginx.HandleController(apiGroup, "/channel", new(api.ChannelController))
	ginx.HandleController(apiGroup, "/customer", new(api.CustomerController))
	ginx.HandleController(apiGroup, "/conversation", new(api.ConversationController), middleware.ExternalUserMiddleware)
	ginx.HandleController(apiGroup, "/message", new(api.MessageController), middleware.ExternalUserMiddleware)

	wsGroup := app.Group("/api/ws")
	wsGroup.GET("/dashboard", middleware.AuthMiddleware, services.WsService.HandleDashboardWS)
	wsGroup.GET("/dashboard/notification", middleware.AuthMiddleware, services.WsService.HandleDashboardNotificationWS)
	wsGroup.GET("/open", services.WsService.HandleOpenWS)

	dashboardGroup := app.Group("/api/dashboard", middleware.AuthMiddleware)
	ginx.HandleController(dashboardGroup, "/dashboard", new(dashboard.DashboardController))
	ginx.HandleController(dashboardGroup, "/user", new(dashboard.UserController))
	ginx.HandleController(dashboardGroup, "/company", new(dashboard.CompanyController))
	ginx.HandleController(dashboardGroup, "/customer", new(dashboard.CustomerController))
	ginx.HandleController(dashboardGroup, "/customer-contact", new(dashboard.CustomerContactController))
	ginx.HandleController(dashboardGroup, "/role", new(dashboard.RoleController))
	ginx.HandleController(dashboardGroup, "/permission", new(dashboard.PermissionController))
	ginx.HandleController(dashboardGroup, "/session", new(dashboard.SessionController))
	ginx.HandleController(dashboardGroup, "/tag", new(dashboard.TagController))
	ginx.HandleController(dashboardGroup, "/conversation", new(dashboard.ConversationController))
	ginx.HandleController(dashboardGroup, "/ticket", new(dashboard.TicketController))
	ginx.HandleController(dashboardGroup, "/notification", new(dashboard.NotificationController))
	ginx.HandleController(dashboardGroup, "/quick-reply", new(dashboard.QuickReplyController))
	ginx.HandleController(dashboardGroup, "/channel", new(dashboard.ChannelController))
	ginx.HandleController(dashboardGroup, "/agent", new(dashboard.AgentController))
	ginx.HandleController(dashboardGroup, "/agent-team", new(dashboard.AgentTeamController))
	ginx.HandleController(dashboardGroup, "/agent-team-schedule", new(dashboard.AgentTeamScheduleController))
	ginx.HandleController(dashboardGroup, "/ai-agent", new(dashboard.AIAgentController))
	ginx.HandleController(dashboardGroup, "/ai-config", new(dashboard.AIConfigController))
	ginx.HandleController(dashboardGroup, "/asset", new(dashboard.AssetController))
	ginx.HandleController(dashboardGroup, "/knowledge-base", new(dashboard.KnowledgeBaseController))
	ginx.HandleController(dashboardGroup, "/knowledge-document", new(dashboard.KnowledgeDocumentController))
	ginx.HandleController(dashboardGroup, "/knowledge-faq", new(dashboard.KnowledgeFAQController))
	ginx.HandleController(dashboardGroup, "/knowledge-retrieve", new(dashboard.KnowledgeRetrieveController))
	ginx.HandleController(dashboardGroup, "/knowledge-retrieve-log", new(dashboard.KnowledgeRetrieveLogController))
	ginx.HandleController(dashboardGroup, "/agent-run-log", new(dashboard.AgentRunLogController))
	ginx.HandleController(dashboardGroup, "/skill-definition", new(dashboard.SkillDefinitionController))
	ginx.HandleController(dashboardGroup, "/mcp", new(dashboard.MCPController))

	thirdGroup := app.Group("/api/third")
	ginx.HandleController(thirdGroup, "/wechat", new(third.WechatController))
}

func registerDashboardStatic(app *gin.Engine, root string) {
	app.NoRoute(func(ctx *gin.Context) {
		if strings.HasPrefix(ctx.Request.URL.Path, "/api/") {
			ctx.JSON(http.StatusNotFound, gin.H{"success": false, "message": "not found"})
			return
		}
		requestPath := filepath.Clean(strings.TrimPrefix(ctx.Request.URL.Path, "/"))
		if strings.HasPrefix(requestPath, "..") {
			ctx.Status(http.StatusBadRequest)
			return
		}
		if requestPath == "." {
			requestPath = "index.html"
		}
		fullPath := filepath.Join(root, requestPath)
		if stat, err := os.Stat(fullPath); err == nil && !stat.IsDir() {
			ctx.File(fullPath)
			return
		}
		ctx.File(filepath.Join(root, "index.html"))
	})
}
