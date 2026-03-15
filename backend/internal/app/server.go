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
	"autostack/internal/modules/product"
	"autostack/internal/modules/project"
	"autostack/internal/modules/shipping"
	"autostack/internal/modules/template"
	"autostack/internal/modules/user"
	companyRepo "autostack/internal/repository/company"
	shippingRepo "autostack/internal/repository/shipping"
	userRepo "autostack/internal/repository/user"
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
		&companyRepo.Company{},
		&user.User{},
		&order.PlatformAuth{},
		&order.Order{},
		&order.OrderItem{},
		&order.OrdersRequestLog{},
		&order.CashFlowStatement{},
		&order.OrderDailyStat{},
		&product.Product{},
		&product.PlatformProduct{},
		&product.ProductMapping{},
		&product.PlatformSyncTask{},
		&product.StockInOrder{},
		&product.StockInOrderItem{},
		&product.Warehouse{},
		&product.WarehouseCenterInventory{},
		&product.ProductSupplier{},
		&shippingRepo.ShippingTemplate{},
		&shippingRepo.ShippingTemplateRule{},
		&shippingRepo.ProductShippingTemplate{},
		&shippingRepo.PlatformProductShippingTemplate{},
	); err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %w", err)
	}

	// 初始化模块 Handler（依赖注入）
	// 注意：初始化顺序很重要，user 必须在 auth 之前
	user.InitHandler(database.GetDB())
	order.InitHandler(database.GetDB())
	product.InitHandler(database.GetDB())
	shipping.InitService(
		shippingRepo.NewShippingTemplateRepository(database.GetDB()),
		shippingRepo.NewShippingTemplateRuleRepository(database.GetDB()),
		shippingRepo.NewProductShippingTemplateRepository(database.GetDB()),
		shippingRepo.NewPlatformProductShippingTemplateRepository(database.GetDB()),
	)

	// 初始化默认超级管理员（需在 user.InitHandler 之后）
	if err := user.InitDefaultSuperAdmin(); err != nil {
		return nil, fmt.Errorf("初始化超级管理员失败: %w", err)
	}

	// 初始化认证服务（需在 user.InitHandler 之后）
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

			// 管理员接口（需 user:view 权限）
			admin := authorized.Group("/admin")
			admin.Use(middleware.RequireAnyPermission(userRepo.PermUserView))
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
				// 手动触发订单走势统计
				admin.POST("/trigger-trend-stats", func(c *gin.Context) {
					scheduler.TriggerTrendStats()
					c.JSON(200, gin.H{"message": "订单走势统计任务已触发，请查看日志"})
				})
				// 手动触发同步任务扫描
				admin.POST("/trigger-sync-tasks", func(c *gin.Context) {
					scheduler.TriggerSyncTasks()
					c.JSON(200, gin.H{"message": "同步任务扫描已触发，请查看日志"})
				})
			}

			// 项目管理（暂无独立菜单，不做权限控制）
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

			// 订单管理模块（包含仪表盘统计、平台授权、订单、现金流）
			orderGroup := authorized.Group("/order")
			orderGroup.Use(middleware.RequireAnyPermission(
				userRepo.PermDashboardView,
				userRepo.PermPlatformAuthView,
				userRepo.PermOrderView,
				userRepo.PermReportView,
			))
			{
				// 仪表盘统计
				orderGroup.GET("/dashboard/stats", order.GetDashboardStats)
				orderGroup.GET("/dashboard/recent-orders", order.GetRecentOrders)
				orderGroup.GET("/dashboard/trend", order.GetOrderTrend)
				orderGroup.GET("/stats/summary", order.GetOrderSummary)
				orderGroup.POST("/dashboard/init", order.InitDashboardStats)
				orderGroup.POST("/dashboard/refresh", order.RefreshDashboardStats)

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
				orderGroup.POST("/auths/:id/sync-cashflow", order.SyncCashFlow)
				orderGroup.POST("/auths/:id/mutual-settlement", order.GetMutualSettlement)

				// 订单管理
				orderGroup.GET("/orders", order.ListOrders)
				orderGroup.GET("/orders/:id", order.GetOrder)
				orderGroup.POST("/orders/:id/sync", order.SyncSingleOrder)
				orderGroup.POST("/orders/:id/sync-commission", order.SyncOrderCommission)

				// 现金流报表
				orderGroup.GET("/cashflow", order.ListCashFlow)
				orderGroup.GET("/cashflow/:id", order.GetCashFlow)
			}

			// 产品管理模块
			productGroup := authorized.Group("/product")
			productGroup.Use(middleware.RequireAnyPermission(
				userRepo.PermProductView,
				userRepo.PermWarehouseView,
			))
			{
				// 本地产品
				productGroup.GET("/products", product.ListProducts)
				productGroup.POST("/products", product.CreateProduct)
				productGroup.PUT("/products/:id", product.UpdateProduct)
				productGroup.DELETE("/products/:id", product.DeleteProduct)
				productGroup.GET("/products/:id/suppliers", product.GetProductSuppliers) // 获取产品的供应商列表
				productGroup.POST("/init", product.InitProducts)                         // 根据平台SKU初始化本地产品

				// 供应商/采购信息
				productGroup.GET("/suppliers", product.ListSuppliers)
				productGroup.POST("/suppliers", product.CreateSupplier)
				productGroup.PUT("/suppliers/:id", product.UpdateSupplier)
				productGroup.DELETE("/suppliers/:id", product.DeleteSupplier)
				productGroup.PUT("/suppliers/batch", product.BatchUpdateSuppliers)             // 批量更新供应商
				productGroup.GET("/suppliers/export-template", product.ExportSupplierTemplate) // 导出导入模板
				productGroup.POST("/suppliers/import", product.ImportSuppliers)                // 导入供应商数据

				// 产品带供应商信息
				productGroup.GET("/products-with-supplier", product.ListProductsWithSupplier)

				// 平台产品
				productGroup.GET("/platform-products", product.ListPlatformProducts)
				productGroup.POST("/sync", product.SyncPlatformProducts)
				productGroup.POST("/sync-direct", product.SyncPlatformProductsDirect) // 直接同步，不走任务队列
				productGroup.POST("/map", product.MapProduct)
				productGroup.DELETE("/map/:id", product.UnmapProduct)

				// 同步任务
				productGroup.GET("/sync-tasks", product.ListSyncTasks)
				productGroup.POST("/sync-tasks/trigger", product.TriggerSyncTasks)

				// 入库单
				productGroup.GET("/stock-in-orders", product.ListStockInOrders)
				productGroup.POST("/stock-in-orders", product.CreateStockInOrder)
				productGroup.GET("/stock-in-orders/:id", product.GetStockInOrder)
				productGroup.GET("/stock-in-template", product.ExportStockInTemplate)
				productGroup.POST("/stock-in-import", product.ImportStockInOrders)

				// 仓库
				productGroup.GET("/warehouses", product.ListWarehouses)
				productGroup.GET("/warehouses/available", product.ListAvailableWarehouses) // 获取当前用户可用仓库
				productGroup.GET("/warehouses/all", product.ListAllWarehouses)
			productGroup.POST("/warehouses", product.CreateWarehouse)
			productGroup.PUT("/warehouses/:id", product.UpdateWarehouse)

			// 库存
				productGroup.GET("/inventory", product.ListInventory)
				productGroup.PUT("/inventory", product.UpdateInventory)
				productGroup.POST("/inventory/init", product.InitInventory)
			}

			// 物流管理模块
			shippingGroup := authorized.Group("/shipping")
			shippingGroup.Use(middleware.RequireAnyPermission(userRepo.PermShippingView))
			{
				// 运费模板
				shippingGroup.GET("/templates", shipping.ListTemplates)
				shippingGroup.GET("/templates/all", shipping.ListAllTemplates) // 获取所有启用模板（下拉选择）
				shippingGroup.POST("/templates", shipping.CreateTemplate)
				shippingGroup.GET("/templates/:id", shipping.GetTemplate)
				shippingGroup.PUT("/templates/:id", shipping.UpdateTemplate)
				shippingGroup.DELETE("/templates/:id", shipping.DeleteTemplate)

				// 运费规则
				shippingGroup.GET("/templates/:id/rules", shipping.GetTemplateRules)
				shippingGroup.POST("/templates/:id/rules", shipping.CreateRule)
				shippingGroup.PUT("/templates/:id/rules/:ruleId", shipping.UpdateRule)
				shippingGroup.DELETE("/templates/:id/rules/:ruleId", shipping.DeleteRule)

				// 运费计算
				shippingGroup.POST("/calculate", shipping.CalculateShippingHandler)
				shippingGroup.POST("/calculate/batch", shipping.BatchCalculateShippingHandler)

				// 本地产品运费模版绑定
				shippingGroup.POST("/product-templates", shipping.BindProductShippingTemplate)
				shippingGroup.DELETE("/product-templates/:id", shipping.UnbindProductShippingTemplate)
				shippingGroup.GET("/products/:productId/templates", shipping.GetProductShippingTemplates)
				shippingGroup.PUT("/products/:productId/default-template", shipping.SetProductDefaultShippingTemplate)

				// 平台产品运费模版绑定
				shippingGroup.POST("/platform-product-templates", shipping.BindPlatformProductShippingTemplate)
				shippingGroup.DELETE("/platform-product-templates/:id", shipping.UnbindPlatformProductShippingTemplate)
				shippingGroup.GET("/platform-products/:platformProductId/templates", shipping.GetPlatformProductShippingTemplates)
				shippingGroup.PUT("/platform-products/:platformProductId/default-template", shipping.SetPlatformProductDefaultShippingTemplate)
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
