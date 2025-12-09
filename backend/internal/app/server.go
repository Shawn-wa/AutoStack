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
	"autostack/internal/modules/project"
	"autostack/internal/modules/template"
	"autostack/internal/modules/user"
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
	if err := database.AutoMigrate(&user.User{}); err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %w", err)
	}

	// 初始化默认超级管理员
	if err := user.InitDefaultSuperAdmin(); err != nil {
		return nil, fmt.Errorf("初始化超级管理员失败: %w", err)
	}

	// 初始化认证服务
	auth.InitService(cfg.JWT.Secret, cfg.JWT.ExpireHour)

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
		}
	}
}

// Run 启动服务器
func (s *Server) Run() error {
	addr := fmt.Sprintf(":%s", s.config.Server.Port)
	fmt.Printf("🚀 AutoStack 服务启动于 http://localhost%s\n", addr)
	return s.router.Run(addr)
}
