package app

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"autostack/internal/commonBase/database"
	"autostack/internal/commonBase/handler"
	"autostack/internal/commonBase/middleware"
	"autostack/internal/config"
	"autostack/internal/modules/auth"
	"autostack/internal/modules/deployment"
	"autostack/internal/modules/order"
	_ "autostack/internal/modules/order/platforms" // 注册平台适配器
	"autostack/internal/modules/project"
	"autostack/internal/modules/template"
	"autostack/internal/modules/user"
	"autostack/internal/scheduler"
)

// Server API服务器
type Server struct {
	config *config.Config
	router *gin.Engine
}

// NewServer 创建服务器实例
func NewServer(cfg *config.Config) (*Server, error) {
	gin.SetMode(cfg.Server.Mode)
	router := gin.Default()

	// 初始化数据库
	if err := database.InitDB(&cfg.Database); err != nil {
		return nil, fmt.Errorf("数据库初始化失败: %w", err)
	}

	// 自动迁移表结构
	if err := database.AutoMigrate(
		&user.User{},
		&order.PlatformAuth{},
		&order.Order{},
		&order.OrderItem{},
		&order.OrdersRequestLog{},
	); err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %w", err)
	}

	// 初始化默认超级管理员
	if err := user.InitDefaultSuperAdmin(); err != nil {
		return nil, fmt.Errorf("初始化超级管理员失败: %w", err)
	}

	// 初始化认证服务
	auth.InitService(cfg.JWT.Secret, cfg.JWT.ExpireHour)

	// 初始化加密模块
	if err := order.InitCrypto(cfg.Crypto.SecretKey); err != nil {
		return nil, fmt.Errorf("初始化加密模块失败: %w", err)
	}

	server := &Server{
		config: cfg,
		router: router,
	}

	server.setupRoutes()
	return server, nil
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// 中间件
	s.router.Use(middleware.Cors())

	// 健康检查
	s.router.GET("/health", handler.Health)

	// API v1
	v1 := s.router.Group("/api/v1")
	{
		// 公开接口
		v1.POST("/auth/login", auth.Login)
		v1.POST("/auth/register", auth.Register)

		// 需要认证的接口
		authorized := v1.Group("/")
		authorized.Use(middleware.JWTAuth(s.config.JWT.Secret))
		{
			// 用户个人信息管理
			userGroup := authorized.Group("/user")
			{
				userGroup.GET("/profile", user.GetProfile)
				userGroup.PUT("/profile", user.UpdateProfile)
				userGroup.PUT("/password", user.ChangePassword)
			}

			// 管理员接口
			admin := authorized.Group("/admin")
			admin.Use(middleware.RequireAdmin())
			{
				admin.GET("/permissions", user.GetPermissions)
				admin.GET("/users", user.ListUsers)
				admin.POST("/users", user.CreateUser)
				admin.GET("/users/:id", user.GetUser)
				admin.PUT("/users/:id", user.UpdateUser)
				admin.DELETE("/users/:id", user.DeleteUser)
				// 手动触发同步任务
				admin.POST("/trigger-sync", func(c *gin.Context) {
					scheduler.TriggerSync()
					c.JSON(200, gin.H{"message": "同步任务已触发，请查看日志"})
				})
			}

			// 项目管理
			projects := authorized.Group("/projects")
			{
				projects.GET("", project.ListProjects)
				projects.POST("", project.CreateProject)
				projects.GET("/:id", project.GetProject)
				projects.PUT("/:id", project.UpdateProject)
				projects.DELETE("/:id", project.DeleteProject)
			}

			// 部署管理
			deployments := authorized.Group("/deployments")
			{
				deployments.GET("", deployment.ListDeployments)
				deployments.POST("", deployment.CreateDeployment)
				deployments.GET("/:id", deployment.GetDeployment)
				deployments.POST("/:id/start", deployment.StartDeployment)
				deployments.POST("/:id/stop", deployment.StopDeployment)
			}

			// 模板管理
			templates := authorized.Group("/templates")
			{
				templates.GET("", template.ListTemplates)
				templates.POST("", template.CreateTemplate)
				templates.GET("/:id", template.GetTemplate)
			}

			// 订单管理模块
			orderGroup := authorized.Group("/order")
			{
				// 平台列表
				orderGroup.GET("/platforms", order.ListPlatforms)

				// 平台授权管理
				orderGroup.GET("/auths", order.ListAuths)
				orderGroup.POST("/auths", order.CreateAuth)
				orderGroup.PUT("/auths/:id", order.UpdateAuth)
				orderGroup.DELETE("/auths/:id", order.DeleteAuth)
				orderGroup.POST("/auths/:id/test", order.TestAuth)
				orderGroup.POST("/auths/:id/sync", order.SyncOrders)
				orderGroup.POST("/auths/:id/sync-commission", order.SyncCommission)

				// 订单管理
				orderGroup.GET("/orders", order.ListOrders)
				orderGroup.GET("/orders/:id", order.GetOrder)
				orderGroup.POST("/orders/:id/sync-commission", order.SyncOrderCommission)
			}
		}
	}
}

// Run 启动服务器
func (s *Server) Run() error {
	// 启动定时任务调度器
	scheduler.Start()

	addr := fmt.Sprintf(":%s", s.config.Server.Port)
	fmt.Printf("🚀 AutoStack 服务启动于 http://localhost%s\n", addr)
	return s.router.Run(addr)
}

// Stop 停止服务器
func (s *Server) Stop() {
	scheduler.Stop()
}
