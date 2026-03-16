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
		&user.Permission{},
		&user.RolePermission{},
		&user.RolePermissionBinding{},
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
				userGroup.GET("/permission-routes", user.GetPermissionRoutes)
				userGroup.PUT("/profile", user.UpdateProfile)
				userGroup.PUT("/password", user.ChangePassword)
			}

			// 管理员接口（需 user:view 权限）
			admin := authorized.Group("/admin")
			admin.Use(middleware.RequireAnyPermission(userRepo.PermUserView, userRepo.PermRoleView))
			{
				admin.GET("/permissions", user.GetPermissions)
				admin.GET("/role-permissions", user.GetRolePermissions)
				admin.PUT("/role-permissions/:role", middleware.RequirePermission("route:system.roles:update"), user.UpdateRolePermissions)
				admin.GET("/users", user.ListUsers)
				admin.POST("/users", middleware.RequirePermission("route:system.users:create"), user.CreateUser)
				admin.GET("/users/:id", user.GetUser)
				admin.PUT("/users/:id", middleware.RequirePermission("route:system.users:update"), user.UpdateUser)
				admin.DELETE("/users/:id", middleware.RequirePermission("route:system.users:delete"), user.DeleteUser)
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
				// 手动触发 company_id 迁移（幂等、可重复调用）
				admin.POST("/migrations/company-id", user.RunCompanyIDMigration)
			}

			// 项目管理（暂无独立菜单，不做权限控制）
			projects := authorized.Group("/projects")
			{
				projects.GET("", middleware.RequirePermission("route:system.projects:read"), project.ListProjects)
				projects.POST("", middleware.RequirePermission("route:system.projects:create"), project.CreateProject)
				projects.GET("/:id", middleware.RequirePermission("route:system.projects:read"), project.GetProject)
				projects.PUT("/:id", middleware.RequirePermission("route:system.projects:update"), project.UpdateProject)
				projects.DELETE("/:id", middleware.RequirePermission("route:system.projects:delete"), project.DeleteProject)
			}

			// 部署管理
			deployments := authorized.Group("/deployments")
			{
				deployments.GET("", middleware.RequirePermission("route:system.deployments:read"), deployment.ListDeployments)
				deployments.POST("", middleware.RequirePermission("route:system.deployments:create"), deployment.CreateDeployment)
				deployments.GET("/:id", middleware.RequirePermission("route:system.deployments:read"), deployment.GetDeployment)
				deployments.POST("/:id/start", middleware.RequirePermission("route:system.deployments:update"), deployment.StartDeployment)
				deployments.POST("/:id/stop", middleware.RequirePermission("route:system.deployments:update"), deployment.StopDeployment)
			}

			// 模板管理
			templates := authorized.Group("/templates")
			{
				templates.GET("", middleware.RequirePermission("route:system.templates:read"), template.ListTemplates)
				templates.POST("", middleware.RequirePermission("route:system.templates:create"), template.CreateTemplate)
				templates.GET("/:id", middleware.RequirePermission("route:system.templates:read"), template.GetTemplate)
			}

			// 订单管理模块（包含仪表盘统计、平台授权、订单、现金流）
			orderGroup := authorized.Group("/order")
			orderGroup.Use(middleware.RequireAnyPermission(
				"route:dashboard.home:read",
				"route:order.platform_auths:read",
				"route:order.orders:read",
				"route:order.cashflow:read",
				"route:order.settlement:read",
			))
			{
				// 仪表盘统计
				orderGroup.GET("/dashboard/stats", middleware.RequirePermission("route:dashboard.home:read"), order.GetDashboardStats)
				orderGroup.GET("/dashboard/recent-orders", middleware.RequirePermission("route:dashboard.home:read"), order.GetRecentOrders)
				orderGroup.GET("/dashboard/trend", middleware.RequirePermission("route:dashboard.home:read"), order.GetOrderTrend)
				orderGroup.GET("/stats/summary", middleware.RequirePermission("route:product.order_summary:read"), order.GetOrderSummary)
				orderGroup.POST("/dashboard/init", middleware.RequirePermission("route:dashboard.home:update"), order.InitDashboardStats)
				orderGroup.POST("/dashboard/refresh", middleware.RequirePermission("route:dashboard.home:update"), order.RefreshDashboardStats)

				// 平台列表
				orderGroup.GET("/platforms", middleware.RequirePermission("route:order.platform_auths:read"), order.ListPlatforms)

				// 平台授权管理
				orderGroup.GET("/auths", middleware.RequirePermission("route:order.platform_auths:read"), order.ListAuths)
				orderGroup.POST("/auths", middleware.RequirePermission("route:order.platform_auths:create"), order.CreateAuth)
				orderGroup.PUT("/auths/:id", middleware.RequirePermission("route:order.platform_auths:update"), order.UpdateAuth)
				orderGroup.DELETE("/auths/:id", middleware.RequirePermission("route:order.platform_auths:delete"), order.DeleteAuth)
				orderGroup.POST("/auths/:id/test", middleware.RequirePermission("route:order.platform_auths:update"), order.TestAuth)
				orderGroup.POST("/auths/:id/sync", middleware.RequirePermission("route:order.platform_auths:update"), order.SyncOrders)
				orderGroup.POST("/auths/:id/sync-commission", middleware.RequirePermission("route:order.platform_auths:update"), order.SyncCommission)
				orderGroup.POST("/auths/:id/sync-cashflow", middleware.RequirePermission("route:order.platform_auths:update"), order.SyncCashFlow)
				orderGroup.POST("/auths/:id/mutual-settlement", middleware.RequirePermission("route:order.platform_auths:update"), order.GetMutualSettlement)

				// 订单管理
				orderGroup.GET("/orders", middleware.RequirePermission("route:order.orders:read"), order.ListOrders)
				orderGroup.GET("/orders/:id", middleware.RequirePermission("route:order.orders:read"), order.GetOrder)
				orderGroup.POST("/orders/:id/sync", middleware.RequirePermission("route:order.orders:update"), order.SyncSingleOrder)
				orderGroup.POST("/orders/:id/sync-commission", middleware.RequirePermission("route:order.orders:update"), order.SyncOrderCommission)

				// 现金流报表
				orderGroup.GET("/cashflow", middleware.RequirePermission("route:order.cashflow:read"), order.ListCashFlow)
				orderGroup.GET("/cashflow/:id", middleware.RequirePermission("route:order.cashflow:read"), order.GetCashFlow)
			}

			// 产品管理模块
			productGroup := authorized.Group("/product")
			productGroup.Use(middleware.RequireAnyPermission(
				"route:product.local_products:read",
				"route:product.platform_products:read",
				"route:product.order_summary:read",
				"route:warehouse.list:read",
				"route:warehouse.inventory:read",
				"route:warehouse.stock_in_orders:read",
			))
			{
				// 本地产品
				productGroup.GET("/products", middleware.RequirePermission("route:product.local_products:read"), product.ListProducts)
				productGroup.POST("/products", middleware.RequirePermission("route:product.local_products:create"), product.CreateProduct)
				productGroup.PUT("/products/:id", middleware.RequirePermission("route:product.local_products:update"), product.UpdateProduct)
				productGroup.DELETE("/products/:id", middleware.RequirePermission("route:product.local_products:delete"), product.DeleteProduct)
				productGroup.GET("/products/:id/suppliers", middleware.RequirePermission("route:product.local_products:read"), product.GetProductSuppliers) // 获取产品的供应商列表
				productGroup.POST("/init", middleware.RequirePermission("route:product.local_products:update"), product.InitProducts)                       // 根据平台SKU初始化本地产品

				// 供应商/采购信息
				productGroup.GET("/suppliers", middleware.RequirePermission("route:product.local_products:read"), product.ListSuppliers)
				productGroup.POST("/suppliers", middleware.RequirePermission("route:product.local_products:create"), product.CreateSupplier)
				productGroup.PUT("/suppliers/:id", middleware.RequirePermission("route:product.local_products:update"), product.UpdateSupplier)
				productGroup.DELETE("/suppliers/:id", middleware.RequirePermission("route:product.local_products:delete"), product.DeleteSupplier)
				productGroup.PUT("/suppliers/batch", middleware.RequirePermission("route:product.local_products:update"), product.BatchUpdateSuppliers)           // 批量更新供应商
				productGroup.GET("/suppliers/export-template", middleware.RequirePermission("route:product.local_products:read"), product.ExportSupplierTemplate) // 导出导入模板
				productGroup.POST("/suppliers/import", middleware.RequirePermission("route:product.local_products:update"), product.ImportSuppliers)              // 导入供应商数据

				// 产品带供应商信息
				productGroup.GET("/products-with-supplier", middleware.RequirePermission("route:product.local_products:read"), product.ListProductsWithSupplier)

				// 平台产品
				productGroup.GET("/platform-products", middleware.RequirePermission("route:product.platform_products:read"), product.ListPlatformProducts)
				productGroup.POST("/sync", middleware.RequirePermission("route:product.platform_products:update"), product.SyncPlatformProducts)
				productGroup.POST("/sync-direct", middleware.RequirePermission("route:product.platform_products:update"), product.SyncPlatformProductsDirect) // 直接同步，不走任务队列
				productGroup.POST("/map", middleware.RequirePermission("route:product.platform_products:update"), product.MapProduct)
				productGroup.DELETE("/map/:id", middleware.RequirePermission("route:product.platform_products:delete"), product.UnmapProduct)

				// 同步任务
				productGroup.GET("/sync-tasks", middleware.RequirePermission("route:product.platform_products:read"), product.ListSyncTasks)
				productGroup.POST("/sync-tasks/trigger", middleware.RequirePermission("route:product.platform_products:update"), product.TriggerSyncTasks)

				// 入库单
				productGroup.GET("/stock-in-orders", middleware.RequirePermission("route:warehouse.stock_in_orders:read"), product.ListStockInOrders)
				productGroup.POST("/stock-in-orders", middleware.RequirePermission("route:warehouse.stock_in_orders:create"), product.CreateStockInOrder)
				productGroup.GET("/stock-in-orders/:id", middleware.RequirePermission("route:warehouse.stock_in_orders:read"), product.GetStockInOrder)
				productGroup.GET("/stock-in-template", middleware.RequirePermission("route:warehouse.stock_in_orders:read"), product.ExportStockInTemplate)
				productGroup.POST("/stock-in-import", middleware.RequirePermission("route:warehouse.stock_in_orders:update"), product.ImportStockInOrders)

				// 仓库
				productGroup.GET("/warehouses", middleware.RequirePermission("route:warehouse.list:read"), product.ListWarehouses)
				productGroup.GET("/warehouses/available", middleware.RequirePermission("route:warehouse.list:read"), product.ListAvailableWarehouses) // 获取当前用户可用仓库
				productGroup.GET("/warehouses/all", middleware.RequirePermission("route:warehouse.list:read"), product.ListAllWarehouses)
				productGroup.POST("/warehouses", middleware.RequirePermission("route:warehouse.list:create"), product.CreateWarehouse)
				productGroup.PUT("/warehouses/:id", middleware.RequirePermission("route:warehouse.list:update"), product.UpdateWarehouse)

				// 库存
				productGroup.GET("/inventory", middleware.RequirePermission("route:warehouse.inventory:read"), product.ListInventory)
				productGroup.PUT("/inventory", middleware.RequirePermission("route:warehouse.inventory:update"), product.UpdateInventory)
				productGroup.POST("/inventory/init", middleware.RequirePermission("route:warehouse.inventory:update"), product.InitInventory)
			}

			// 物流管理模块
			shippingGroup := authorized.Group("/shipping")
			shippingGroup.Use(middleware.RequireAnyPermission("route:shipping.templates:read"))
			{
				// 运费模板
				shippingGroup.GET("/templates", middleware.RequirePermission("route:shipping.templates:read"), shipping.ListTemplates)
				shippingGroup.GET("/templates/all", middleware.RequirePermission("route:shipping.templates:read"), shipping.ListAllTemplates) // 获取所有启用模板（下拉选择）
				shippingGroup.POST("/templates", middleware.RequirePermission("route:shipping.templates:create"), shipping.CreateTemplate)
				shippingGroup.GET("/templates/:id", middleware.RequirePermission("route:shipping.templates:read"), shipping.GetTemplate)
				shippingGroup.PUT("/templates/:id", middleware.RequirePermission("route:shipping.templates:update"), shipping.UpdateTemplate)
				shippingGroup.DELETE("/templates/:id", middleware.RequirePermission("route:shipping.templates:delete"), shipping.DeleteTemplate)

				// 运费规则
				shippingGroup.GET("/templates/:id/rules", middleware.RequirePermission("route:shipping.templates:read"), shipping.GetTemplateRules)
				shippingGroup.POST("/templates/:id/rules", middleware.RequirePermission("route:shipping.templates:update"), shipping.CreateRule)
				shippingGroup.PUT("/templates/:id/rules/:ruleId", middleware.RequirePermission("route:shipping.templates:update"), shipping.UpdateRule)
				shippingGroup.DELETE("/templates/:id/rules/:ruleId", middleware.RequirePermission("route:shipping.templates:delete"), shipping.DeleteRule)

				// 运费计算
				shippingGroup.POST("/calculate", middleware.RequirePermission("route:shipping.templates:read"), shipping.CalculateShippingHandler)
				shippingGroup.POST("/calculate/batch", middleware.RequirePermission("route:shipping.templates:read"), shipping.BatchCalculateShippingHandler)

				// 本地产品运费模版绑定
				shippingGroup.POST("/product-templates", middleware.RequirePermission("route:shipping.templates:update"), shipping.BindProductShippingTemplate)
				shippingGroup.DELETE("/product-templates/:id", middleware.RequirePermission("route:shipping.templates:delete"), shipping.UnbindProductShippingTemplate)
				shippingGroup.GET("/products/:productId/templates", middleware.RequirePermission("route:shipping.templates:read"), shipping.GetProductShippingTemplates)
				shippingGroup.PUT("/products/:productId/default-template", middleware.RequirePermission("route:shipping.templates:update"), shipping.SetProductDefaultShippingTemplate)

				// 平台产品运费模版绑定
				shippingGroup.POST("/platform-product-templates", middleware.RequirePermission("route:shipping.templates:update"), shipping.BindPlatformProductShippingTemplate)
				shippingGroup.DELETE("/platform-product-templates/:id", middleware.RequirePermission("route:shipping.templates:delete"), shipping.UnbindPlatformProductShippingTemplate)
				shippingGroup.GET("/platform-products/:platformProductId/templates", middleware.RequirePermission("route:shipping.templates:read"), shipping.GetPlatformProductShippingTemplates)
				shippingGroup.PUT("/platform-products/:platformProductId/default-template", middleware.RequirePermission("route:shipping.templates:update"), shipping.SetPlatformProductDefaultShippingTemplate)
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
