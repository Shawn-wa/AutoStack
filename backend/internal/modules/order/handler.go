package order

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"autostack/pkg/response"
)

var orderService = NewService()

// ListPlatforms 获取支持的平台列表
func ListPlatforms(c *gin.Context) {
	platforms := orderService.GetAllPlatformsInfo()
	response.Success(c, http.StatusOK, "获取成功", platforms)
}

// ListAuths 获取授权列表
func ListAuths(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	auths, total, err := orderService.ListAuths(userID, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取授权列表失败")
		return
	}

	list := make([]AuthResponse, len(auths))
	for i, auth := range auths {
		list[i] = AuthResponse{
			ID:                auth.ID,
			Platform:          auth.Platform,
			ShopName:          auth.ShopName,
			Status:            auth.Status,
			MaskedCredentials: orderService.GetMaskedCredentials(&auth),
			LastSyncAt:        auth.LastSyncAt,
			CreatedAt:         auth.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:         auth.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	response.Success(c, http.StatusOK, "获取成功", AuthListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// CreateAuth 创建授权
func CreateAuth(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	var req CreateAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	auth, err := orderService.CreateAuth(userID, &req)
	if err != nil {
		if err == ErrPlatformNotFound {
			response.Error(c, http.StatusBadRequest, "不支持的平台")
			return
		}
		response.Error(c, http.StatusInternalServerError, "创建授权失败")
		return
	}

	response.Success(c, http.StatusCreated, "创建成功", AuthResponse{
		ID:         auth.ID,
		Platform:   auth.Platform,
		ShopName:   auth.ShopName,
		Status:     auth.Status,
		LastSyncAt: auth.LastSyncAt,
		CreatedAt:  auth.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:  auth.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

// UpdateAuth 更新授权
func UpdateAuth(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的ID")
		return
	}

	var req UpdateAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	auth, err := orderService.UpdateAuth(uint(id), userID, &req)
	if err != nil {
		if err == ErrAuthNotFound {
			response.Error(c, http.StatusNotFound, "授权不存在")
			return
		}
		response.Error(c, http.StatusInternalServerError, "更新授权失败")
		return
	}

	response.Success(c, http.StatusOK, "更新成功", AuthResponse{
		ID:         auth.ID,
		Platform:   auth.Platform,
		ShopName:   auth.ShopName,
		Status:     auth.Status,
		LastSyncAt: auth.LastSyncAt,
		CreatedAt:  auth.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:  auth.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

// DeleteAuth 删除授权
func DeleteAuth(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的ID")
		return
	}

	if err := orderService.DeleteAuth(uint(id), userID); err != nil {
		if err == ErrAuthNotFound {
			response.Error(c, http.StatusNotFound, "授权不存在")
			return
		}
		response.Error(c, http.StatusInternalServerError, "删除授权失败")
		return
	}

	response.Success(c, http.StatusOK, "删除成功", nil)
}

// TestAuth 测试授权连接
func TestAuth(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的ID")
		return
	}

	if err := orderService.TestAuth(uint(id), userID); err != nil {
		if err == ErrAuthNotFound {
			response.Error(c, http.StatusNotFound, "授权不存在")
			return
		}
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "连接成功", nil)
}

// SyncOrders 同步订单
func SyncOrders(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的ID")
		return
	}

	var req SyncOrdersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 默认同步最近7天
		req.Since = time.Now().AddDate(0, 0, -7).Format(time.RFC3339)
		req.To = time.Now().Format(time.RFC3339)
	}

	since, err := time.Parse(time.RFC3339, req.Since)
	if err != nil {
		since = time.Now().AddDate(0, 0, -7)
	}

	to, err := time.Parse(time.RFC3339, req.To)
	if err != nil {
		to = time.Now()
	}

	result, err := orderService.SyncOrders(uint(id), userID, since, to)
	if err != nil {
		if err == ErrAuthNotFound {
			response.Error(c, http.StatusNotFound, "授权不存在")
			return
		}
		response.Error(c, http.StatusInternalServerError, "同步订单失败: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "同步成功", result)
}

// ListOrders 获取订单列表
func ListOrders(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	var req OrderListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	orders, total, err := orderService.ListOrders(userID, &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取订单列表失败")
		return
	}

	list := make([]OrderResponse, len(orders))
	for i, ord := range orders {
		items := make([]OrderItemResponse, len(ord.Items))
		for j, item := range ord.Items {
			items[j] = OrderItemResponse{
				ID:          item.ID,
				PlatformSku: item.PlatformSku,
				Sku:         item.Sku,
				Name:        item.Name,
				Quantity:    item.Quantity,
				Price:       item.Price,
				Currency:    item.Currency,
			}
		}

		list[i] = OrderResponse{
			ID:                      ord.ID,
			Platform:                ord.Platform,
			PlatformOrderNo:         ord.PlatformOrderNo,
			Status:                  ord.Status,
			PlatformStatus:          ord.PlatformStatus,
			TotalAmount:             ord.TotalAmount,
			Currency:                ord.Currency,
			RecipientName:           ord.RecipientName,
			RecipientPhone:          ord.RecipientPhone,
			Country:                 ord.Country,
			Province:                ord.Province,
			City:                    ord.City,
			ZipCode:                 ord.ZipCode,
			Address:                 ord.Address,
			OrderTime:               ord.OrderTime,
			ShipTime:                ord.ShipTime,
			AccrualsForSale:         ord.AccrualsForSale,
			SaleCommission:          ord.SaleCommission,
			ProcessingAndDelivery:   ord.ProcessingAndDelivery,
			RefundsAndCancellations: ord.RefundsAndCancellations,
			ServicesAmount:          ord.ServicesAmount,
			CompensationAmount:      ord.CompensationAmount,
			MoneyTransfer:           ord.MoneyTransfer,
			OthersAmount:            ord.OthersAmount,
			ProfitAmount:            ord.ProfitAmount,
			CommissionCurrency:      ord.CommissionCurrency,
			CommissionSyncedAt:      ord.CommissionSyncedAt,
			Items:                   items,
			CreatedAt:               ord.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:               ord.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	response.Success(c, http.StatusOK, "获取成功", OrderListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// GetOrder 获取订单详情
func GetOrder(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的ID")
		return
	}

	ord, err := orderService.GetOrderByID(uint(id), userID)
	if err != nil {
		if err == ErrOrderNotFound {
			response.Error(c, http.StatusNotFound, "订单不存在")
			return
		}
		response.Error(c, http.StatusInternalServerError, "获取订单失败")
		return
	}

	items := make([]OrderItemResponse, len(ord.Items))
	for j, item := range ord.Items {
		items[j] = OrderItemResponse{
			ID:          item.ID,
			PlatformSku: item.PlatformSku,
			Sku:         item.Sku,
			Name:        item.Name,
			Quantity:    item.Quantity,
			Price:       item.Price,
			Currency:    item.Currency,
		}
	}

	response.Success(c, http.StatusOK, "获取成功", OrderResponse{
		ID:                      ord.ID,
		Platform:                ord.Platform,
		PlatformOrderNo:         ord.PlatformOrderNo,
		Status:                  ord.Status,
		PlatformStatus:          ord.PlatformStatus,
		TotalAmount:             ord.TotalAmount,
		Currency:                ord.Currency,
		RecipientName:           ord.RecipientName,
		RecipientPhone:          ord.RecipientPhone,
		Country:                 ord.Country,
		Province:                ord.Province,
		City:                    ord.City,
		ZipCode:                 ord.ZipCode,
		Address:                 ord.Address,
		OrderTime:               ord.OrderTime,
		ShipTime:                ord.ShipTime,
		AccrualsForSale:         ord.AccrualsForSale,
		SaleCommission:          ord.SaleCommission,
		ProcessingAndDelivery:   ord.ProcessingAndDelivery,
		RefundsAndCancellations: ord.RefundsAndCancellations,
		ServicesAmount:          ord.ServicesAmount,
		CompensationAmount:      ord.CompensationAmount,
		MoneyTransfer:           ord.MoneyTransfer,
		OthersAmount:            ord.OthersAmount,
		ProfitAmount:            ord.ProfitAmount,
		CommissionCurrency:      ord.CommissionCurrency,
		CommissionSyncedAt:      ord.CommissionSyncedAt,
		Items:                   items,
		CreatedAt:               ord.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:               ord.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

// SyncCommission 同步佣金
func SyncCommission(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的ID")
		return
	}

	var req SyncCommissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 默认同步最近30天
		req.Since = time.Now().AddDate(0, 0, -30).Format(time.RFC3339)
		req.To = time.Now().Format(time.RFC3339)
	}

	since, err := time.Parse(time.RFC3339, req.Since)
	if err != nil {
		since = time.Now().AddDate(0, 0, -30)
	}

	to, err := time.Parse(time.RFC3339, req.To)
	if err != nil {
		to = time.Now()
	}

	result, err := orderService.SyncCommission(userID, uint(id), since, to)
	if err != nil {
		if err == ErrAuthNotFound {
			response.Error(c, http.StatusNotFound, "授权不存在")
			return
		}
		if err == ErrPlatformNotFound {
			response.Error(c, http.StatusBadRequest, "不支持的平台")
			return
		}
		if err == ErrInvalidCredentials {
			response.Error(c, http.StatusBadRequest, "凭证无效，请重新配置")
			return
		}
		response.Error(c, http.StatusInternalServerError, "同步佣金失败: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "同步成功", result)
}

// SyncOrderCommission 同步单个订单的佣金
func SyncOrderCommission(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的订单ID")
		return
	}

	ord, err := orderService.SyncOrderCommission(userID, uint(id))
	if err != nil {
		if err == ErrOrderNotFound {
			response.Error(c, http.StatusNotFound, "订单不存在")
			return
		}
		if err == ErrAuthNotFound {
			response.Error(c, http.StatusNotFound, "授权不存在")
			return
		}
		if err == ErrPlatformNotFound {
			response.Error(c, http.StatusBadRequest, "不支持的平台")
			return
		}
		response.Error(c, http.StatusInternalServerError, "同步佣金失败: "+err.Error())
		return
	}

	// 转换为响应格式
	items := make([]OrderItemResponse, len(ord.Items))
	for i, item := range ord.Items {
		items[i] = OrderItemResponse{
			ID:          item.ID,
			PlatformSku: item.PlatformSku,
			Sku:         item.Sku,
			Name:        item.Name,
			Quantity:    item.Quantity,
			Price:       item.Price,
			Currency:    item.Currency,
		}
	}

	response.Success(c, http.StatusOK, "同步成功", OrderResponse{
		ID:                      ord.ID,
		Platform:                ord.Platform,
		PlatformOrderNo:         ord.PlatformOrderNo,
		Status:                  ord.Status,
		PlatformStatus:          ord.PlatformStatus,
		TotalAmount:             ord.TotalAmount,
		Currency:                ord.Currency,
		RecipientName:           ord.RecipientName,
		RecipientPhone:          ord.RecipientPhone,
		Country:                 ord.Country,
		Province:                ord.Province,
		City:                    ord.City,
		ZipCode:                 ord.ZipCode,
		Address:                 ord.Address,
		OrderTime:               ord.OrderTime,
		ShipTime:                ord.ShipTime,
		AccrualsForSale:         ord.AccrualsForSale,
		SaleCommission:          ord.SaleCommission,
		ProcessingAndDelivery:   ord.ProcessingAndDelivery,
		RefundsAndCancellations: ord.RefundsAndCancellations,
		ServicesAmount:          ord.ServicesAmount,
		CompensationAmount:      ord.CompensationAmount,
		MoneyTransfer:           ord.MoneyTransfer,
		OthersAmount:            ord.OthersAmount,
		ProfitAmount:            ord.ProfitAmount,
		CommissionCurrency:      ord.CommissionCurrency,
		CommissionSyncedAt:      ord.CommissionSyncedAt,
		Items:                   items,
		CreatedAt:               ord.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:               ord.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

// getUserID 获取当前用户ID
func getUserID(c *gin.Context) uint {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0
	}

	switch v := userID.(type) {
	case float64:
		return uint(v)
	case uint:
		return v
	case int:
		return uint(v)
	default:
		return 0
	}
}

// ========== 现金流报表相关 ==========

// SyncCashFlow 同步现金流报表
func SyncCashFlow(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的ID")
		return
	}

	var req SyncCashFlowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 默认同步最近90天
		req.Since = time.Now().AddDate(0, 0, -90).Format(time.RFC3339)
		req.To = time.Now().Format(time.RFC3339)
	}

	since, err := time.Parse(time.RFC3339, req.Since)
	if err != nil {
		since = time.Now().AddDate(0, 0, -90)
	}

	to, err := time.Parse(time.RFC3339, req.To)
	if err != nil {
		to = time.Now()
	}

	result, err := orderService.SyncCashFlowStatements(uint(id), userID, since, to)
	if err != nil {
		if err == ErrAuthNotFound {
			response.Error(c, http.StatusNotFound, "授权不存在")
			return
		}
		response.Error(c, http.StatusInternalServerError, "同步现金流报表失败: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "同步成功", result)
}

// ListCashFlow 获取现金流报表列表
func ListCashFlow(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	authID, _ := strconv.Atoi(c.DefaultQuery("auth_id", "0"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	statements, total, err := orderService.ListCashFlowStatements(userID, uint(authID), page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取现金流报表失败")
		return
	}

	list := make([]CashFlowResponse, len(statements))
	for i, s := range statements {
		list[i] = CashFlowResponse{
			ID:                          s.ID,
			PlatformAuthID:              s.PlatformAuthID,
			Platform:                    s.Platform,
			PeriodBegin:                 s.PeriodBegin,
			PeriodEnd:                   s.PeriodEnd,
			CurrencyCode:                s.CurrencyCode,
			OrdersAmount:                s.OrdersAmount,
			ReturnsAmount:               s.ReturnsAmount,
			CommissionAmount:            s.CommissionAmount,
			ServicesAmount:              s.ServicesAmount,
			ItemDeliveryAndReturnAmount: s.ItemDeliveryAndReturnAmount,
			SyncedAt:                    s.SyncedAt.Format("2006-01-02 15:04:05"),
		}
	}

	response.Success(c, http.StatusOK, "获取成功", CashFlowListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// GetCashFlow 获取现金流报表详情
func GetCashFlow(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的ID")
		return
	}

	statement, err := orderService.GetCashFlowStatement(uint(id), userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "现金流报表不存在")
		return
	}

	response.Success(c, http.StatusOK, "获取成功", CashFlowResponse{
		ID:                          statement.ID,
		PlatformAuthID:              statement.PlatformAuthID,
		Platform:                    statement.Platform,
		PeriodBegin:                 statement.PeriodBegin,
		PeriodEnd:                   statement.PeriodEnd,
		CurrencyCode:                statement.CurrencyCode,
		OrdersAmount:                statement.OrdersAmount,
		ReturnsAmount:               statement.ReturnsAmount,
		CommissionAmount:            statement.CommissionAmount,
		ServicesAmount:              statement.ServicesAmount,
		ItemDeliveryAndReturnAmount: statement.ItemDeliveryAndReturnAmount,
		SyncedAt:                    statement.SyncedAt.Format("2006-01-02 15:04:05"),
	})
}

// GetDashboardStats 获取仪表盘统计数据
func GetDashboardStats(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	stats, err := orderService.GetDashboardStats(userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取统计数据失败")
		return
	}

	response.Success(c, http.StatusOK, "获取成功", stats)
}

// GetRecentOrders 获取最近订单
func GetRecentOrders(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	orders, err := orderService.GetRecentOrders(userID, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取最近订单失败")
		return
	}

	response.Success(c, http.StatusOK, "获取成功", orders)
}

// GetOrderTrend 获取订单趋势数据
func GetOrderTrend(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	if days <= 0 || days > 30 {
		days = 7
	}
	currency := c.Query("currency")

	trend, err := orderService.GetOrderTrend(userID, days, currency)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取订单趋势失败")
		return
	}

	response.Success(c, http.StatusOK, "获取成功", trend)
}

// InitDashboardStats 初始化仪表盘统计数据（首次访问时调用）
func InitDashboardStats(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	// 异步初始化统计数据（不强制更新）
	go func() {
		_ = orderService.InitOrderTrendStats(userID, false)
	}()

	response.Success(c, http.StatusOK, "初始化任务已启动", nil)
}

// RefreshDashboardStats 刷新仪表盘统计数据（强制重新计算）
func RefreshDashboardStats(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	// 同步刷新统计数据（强制更新）
	err := orderService.InitOrderTrendStats(userID, true)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "刷新统计数据失败")
		return
	}

	response.Success(c, http.StatusOK, "统计数据已刷新", nil)
}

// GetMutualSettlement 获取结算报告
// API: POST /v1/finance/mutual-settlement
func GetMutualSettlement(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的ID")
		return
	}

	// 解析请求参数
	var req struct {
		Since string `json:"since"` // 开始时间 RFC3339
		To    string `json:"to"`    // 结束时间 RFC3339
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// 默认当月
		now := time.Now()
		req.Since = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format(time.RFC3339)
		req.To = now.Format(time.RFC3339)
	}

	since, err := time.Parse(time.RFC3339, req.Since)
	if err != nil {
		now := time.Now()
		since = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	}

	to, err := time.Parse(time.RFC3339, req.To)
	if err != nil {
		to = time.Now()
	}

	result, err := orderService.GetMutualSettlement(uint(id), userID, since, to)
	if err != nil {
		if err == ErrAuthNotFound {
			response.Error(c, http.StatusNotFound, "授权不存在")
			return
		}
		response.Error(c, http.StatusInternalServerError, "获取结算报告失败: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "获取成功", result)
}
