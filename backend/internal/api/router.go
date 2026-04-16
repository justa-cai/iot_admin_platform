package api

import (
	"iot-admin/internal/api/handler"
	"iot-admin/internal/api/middleware"
	"iot-admin/internal/rule"
	"iot-admin/internal/store/sqlite"
	"iot-admin/internal/ws"
	"time"

	"github.com/gin-gonic/gin"
)

type Router struct {
	engine *gin.Engine
}

func NewRouter(
	cfg struct {
		JWTSecret   string
		JWTExpire   time.Duration
	},
	stores struct {
		UserStore      *sqlite.UserStore
		DeviceStore    *sqlite.DeviceStore
		GroupStore     *sqlite.GroupStore
		TagStore       *sqlite.TagStore
		MessageStore   *sqlite.MessageStore
		TelemetryStore *sqlite.TelemetryStore
		RuleStore      *sqlite.RuleStore
		FirmwareStore  *sqlite.FirmwareStore
		OTAStore       *sqlite.OTAStore
	},
	services struct {
		Hub        *ws.Hub
		RuleEngine *rule.Engine
	},
	uploadDir string,
) *Router {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())

	// Health check
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Auth routes (public)
	authHandler := handler.NewAuthHandler(stores.UserStore, cfg.JWTSecret, cfg.JWTExpire)
	r.POST("/api/v1/auth/login", authHandler.Login)
	r.POST("/api/v1/auth/register", authHandler.Register)

	// WebSocket
	r.GET("/api/v1/ws", middleware.OptionalAuth(cfg.JWTSecret), ws.HandleWebSocket(services.Hub))

	// Protected routes
	api := r.Group("/api/v1")
	api.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		// Auth profile
		api.GET("/auth/profile", authHandler.GetProfile)
		api.PUT("/auth/password", authHandler.ChangePassword)

		// Users (admin/operator)
		userHandler := handler.NewUserHandler(stores.UserStore)
		users := api.Group("/users")
		users.Use(middleware.RequireRole("admin"))
		{
			users.GET("", userHandler.List)
			users.GET("/:id", userHandler.Get)
			users.POST("", userHandler.Create)
			users.PUT("/:id", userHandler.Update)
			users.DELETE("/:id", userHandler.Delete)
			users.PUT("/:id/role", userHandler.UpdateRole)
			users.PUT("/:id/password", userHandler.ResetPassword)
		}

		// Devices
		deviceHandler := handler.NewDeviceHandler(stores.DeviceStore)
		devices := api.Group("/devices")
		{
			devices.GET("", deviceHandler.List)
			devices.POST("", deviceHandler.Create)
			devices.GET("/:id", deviceHandler.Get)
			devices.PUT("/:id", deviceHandler.Update)
			devices.DELETE("/:id", deviceHandler.Delete)
			devices.GET("/:id/status", deviceHandler.GetStatus)
		}

		// Groups
		groupHandler := handler.NewGroupHandler(stores.GroupStore)
		groups := api.Group("/groups")
		{
			groups.GET("", groupHandler.List)
			groups.POST("", groupHandler.Create)
			groups.GET("/:id", groupHandler.Get)
			groups.PUT("/:id", groupHandler.Update)
			groups.DELETE("/:id", groupHandler.Delete)
			groups.POST("/:id/devices", groupHandler.AddDevices)
			groups.DELETE("/:id/devices", groupHandler.RemoveDevices)
		}

		// Tags
		tagHandler := handler.NewTagHandler(stores.TagStore)
		tags := api.Group("/tags")
		{
			tags.GET("", tagHandler.List)
			tags.POST("", tagHandler.Create)
			tags.PUT("/:id", tagHandler.Update)
			tags.DELETE("/:id", tagHandler.Delete)
		}

		// Messages
		messageHandler := handler.NewMessageHandler(stores.MessageStore)
		messages := api.Group("/messages")
		{
			messages.POST("/publish", messageHandler.Publish)
			messages.GET("/history", messageHandler.History)
			messages.GET("/topics", messageHandler.Topics)
		}

		// Rules
		ruleHandler := handler.NewRuleHandler(stores.RuleStore)
		rules := api.Group("/rules")
		{
			rules.GET("", ruleHandler.List)
			rules.POST("", ruleHandler.Create)
			rules.GET("/:id", ruleHandler.Get)
			rules.PUT("/:id", ruleHandler.Update)
			rules.DELETE("/:id", ruleHandler.Delete)
			rules.PUT("/:id/enable", ruleHandler.SetEnabled)
			rules.GET("/:id/logs", ruleHandler.Logs)
		}

		// Dashboard
		dashboardHandler := handler.NewDashboardHandler(
			stores.DeviceStore, stores.MessageStore,
			stores.RuleStore, stores.UserStore,
			stores.TelemetryStore,
		)
		api.GET("/dashboard/stats", dashboardHandler.Stats)
		api.GET("/dashboard/throughput", dashboardHandler.Throughput)

		// Telemetry
		telemetryHandler := handler.NewTelemetryHandler(stores.TelemetryStore)
		api.GET("/telemetry/query", telemetryHandler.Query)
		api.GET("/telemetry/latest", telemetryHandler.Latest)

		// Firmware & OTA
		fwHandler := handler.NewFirmwareHandler(stores.FirmwareStore, stores.OTAStore, uploadDir)
		firmware := api.Group("/firmware")
		firmware.Use(middleware.RequireRole("admin", "operator"))
		{
			firmware.GET("", fwHandler.List)
			firmware.POST("/upload", fwHandler.Upload)
			firmware.GET("/:id/download", fwHandler.Download)
			firmware.DELETE("/:id", fwHandler.Delete)
		}

		ota := api.Group("/ota")
		{
			ota.POST("", middleware.RequireRole("admin", "operator"), fwHandler.CreateOTA)
			ota.POST("/batch", middleware.RequireRole("admin", "operator"), fwHandler.BatchCreateOTA)
			ota.GET("", fwHandler.ListOTA)
			ota.PUT("/:id/status", fwHandler.UpdateOTAStatus)
			ota.DELETE("/:id", middleware.RequireRole("admin", "operator"), fwHandler.DeleteOTA)
		}
	}

	return &Router{engine: r}
}

func (r *Router) Engine() *gin.Engine {
	return r.engine
}
