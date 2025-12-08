package api

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"autostack/internal/config"
	"autostack/internal/handler"
	"autostack/internal/middleware"
)

// Server API服务器
type Server struct {
	config *config.Config
	router *gin.Engine
}

// NewServer 创建服务器实例
func NewServer(cfg *config.Config) *Server {
	gin.SetMode(cfg.Server.Mode)
	router := gin.Default()

	server := &Server{
		config: cfg,
		router: router,
	}

	server.setupRoutes()
	return server
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
		v1.POST("/auth/login", handler.Login)
		v1.POST("/auth/register", handler.Register)

		// 需要认证的接口
		authorized := v1.Group("/")
		authorized.Use(middleware.JWTAuth(s.config.JWT.Secret))
		{
			// 项目管理
			projects := authorized.Group("/projects")
			{
				projects.GET("", handler.ListProjects)
				projects.POST("", handler.CreateProject)
				projects.GET("/:id", handler.GetProject)
				projects.PUT("/:id", handler.UpdateProject)
				projects.DELETE("/:id", handler.DeleteProject)
			}

			// 部署管理
			deployments := authorized.Group("/deployments")
			{
				deployments.GET("", handler.ListDeployments)
				deployments.POST("", handler.CreateDeployment)
				deployments.GET("/:id", handler.GetDeployment)
				deployments.POST("/:id/start", handler.StartDeployment)
				deployments.POST("/:id/stop", handler.StopDeployment)
			}

			// 模板管理
			templates := authorized.Group("/templates")
			{
				templates.GET("", handler.ListTemplates)
				templates.POST("", handler.CreateTemplate)
				templates.GET("/:id", handler.GetTemplate)
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

