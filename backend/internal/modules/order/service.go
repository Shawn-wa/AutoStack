package order

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"gorm.io/gorm"

	"autostack/internal/commonBase/database"
	"autostack/internal/modules/apiClient/platform"
	"autostack/internal/modules/apiClient/platform/ozon"
)

var (
	ErrAuthNotFound       = errors.New("授权不存在")
	ErrOrderNotFound      = errors.New("订单不存在")
	ErrPlatformNotFound   = errors.New("平台不存在")
	ErrInvalidCredentials = errors.New("凭证无效")
)

// Service 订单服务
type Service struct{}

// NewService 创建服务实例
func NewService() *Service {
	return &Service{}
}

// GetAllPlatformsInfo 获取所有平台信息
func (s *Service) GetAllPlatformsInfo() []PlatformInfo {
	return GetAllPlatforms()
}

// CreateAuth 创建平台授权
func (s *Service) CreateAuth(userID uint, req *CreateAuthRequest) (*PlatformAuth, error) {
	db := database.GetDB()

	// 检查平台是否支持
	adapter := GetAdapter(req.Platform)
	if adapter == nil {
		return nil, ErrPlatformNotFound
	}

	// 序列化凭证
	credBytes, err := json.Marshal(req.Credentials)
	if err != nil {
		return nil, err
	}

	// 加密凭证
	encryptedCreds, err := Encrypt(string(credBytes))
	if err != nil {
		return nil, err
	}

	auth := &PlatformAuth{
		UserID:      userID,
		Platform:    req.Platform,
		ShopName:    req.ShopName,
		Credentials: encryptedCreds,
		Status:      AuthStatusActive,
	}

	if err := db.Create(auth).Error; err != nil {
		return nil, err
	}

	return auth, nil
}

// GetAuthByID 根据ID获取授权
func (s *Service) GetAuthByID(id uint, userID uint) (*PlatformAuth, error) {
	db := database.GetDB()

	var auth PlatformAuth
	if err := db.Where("id = ? AND user_id = ?", id, userID).First(&auth).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAuthNotFound
		}
		return nil, err
	}

	return &auth, nil
}

// ListAuths 获取授权列表
func (s *Service) ListAuths(userID uint, page, pageSize int) ([]PlatformAuth, int64, error) {
	db := database.GetDB()

	var auths []PlatformAuth
	var total int64

	query := db.Model(&PlatformAuth{}).Where("user_id = ?", userID)
	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&auths).Error; err != nil {
		return nil, 0, err
	}

	return auths, total, nil
}

// maskValue 脱敏单个值，显示最后4位
func maskValue(value string) string {
	if len(value) <= 4 {
		return "****"
	}
	return "****" + value[len(value)-4:]
}

// GetMaskedCredentials 获取脱敏后的凭证
func (s *Service) GetMaskedCredentials(auth *PlatformAuth) map[string]string {
	result := make(map[string]string)

	// 解密凭证
	decrypted, err := Decrypt(auth.Credentials)
	if err != nil {
		return result
	}

	var credentials map[string]string
	if err := json.Unmarshal([]byte(decrypted), &credentials); err != nil {
		return result
	}

	// 脱敏处理
	for key, value := range credentials {
		result[key] = maskValue(value)
	}

	return result
}

// UpdateAuth 更新授权
func (s *Service) UpdateAuth(id uint, userID uint, req *UpdateAuthRequest) (*PlatformAuth, error) {
	db := database.GetDB()

	auth, err := s.GetAuthByID(id, userID)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})

	if req.ShopName != "" {
		updates["shop_name"] = req.ShopName
	}

	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if len(req.Credentials) > 0 {
		credBytes, err := json.Marshal(req.Credentials)
		if err != nil {
			return nil, err
		}
		encryptedCreds, err := Encrypt(string(credBytes))
		if err != nil {
			return nil, err
		}
		updates["credentials"] = encryptedCreds
	}

	if len(updates) > 0 {
		if err := db.Model(auth).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	return auth, nil
}

// DeleteAuth 删除授权
func (s *Service) DeleteAuth(id uint, userID uint) error {
	db := database.GetDB()

	result := db.Where("id = ? AND user_id = ?", id, userID).Delete(&PlatformAuth{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrAuthNotFound
	}

	return nil
}

// TestAuth 测试授权连接
func (s *Service) TestAuth(id uint, userID uint) error {
	auth, err := s.GetAuthByID(id, userID)
	if err != nil {
		return err
	}

	adapter := GetAdapter(auth.Platform)
	if adapter == nil {
		return ErrPlatformNotFound
	}

	// 解密凭证
	credentials, err := Decrypt(auth.Credentials)
	if err != nil {
		return ErrInvalidCredentials
	}

	// 尝试使用带日志的方法
	if adapterWithLog, ok := adapter.(PlatformAdapterWithLog); ok {
		return adapterWithLog.TestConnectionWithLog(credentials, auth.ID)
	}
	return adapter.TestConnection(credentials)
}

// SyncOrders 同步订单
func (s *Service) SyncOrders(id uint, userID uint, since, to time.Time) (*SyncOrdersResponse, error) {
	db := database.GetDB()

	auth, err := s.GetAuthByID(id, userID)
	if err != nil {
		return nil, err
	}

	adapter := GetAdapter(auth.Platform)
	if adapter == nil {
		return nil, ErrPlatformNotFound
	}

	// 解密凭证
	credentials, err := Decrypt(auth.Credentials)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// 调用适配器同步订单（优先使用带日志的方法）
	var orders []*Order
	if adapterWithLog, ok := adapter.(PlatformAdapterWithLog); ok {
		orders, err = adapterWithLog.SyncOrdersWithLog(credentials, since, to, auth.ID)
	} else {
		orders, err = adapter.SyncOrders(credentials, since, to)
	}
	if err != nil {
		log.Printf("[SyncOrders] 同步失败: %v", err)
		return nil, err
	}

	log.Printf("[SyncOrders] 从平台获取到 %d 条订单", len(orders))

	result := &SyncOrdersResponse{}

	// 保存订单
	for _, ord := range orders {
		ord.UserID = userID
		ord.PlatformAuthID = auth.ID

		// 检查订单是否已存在
		var existingOrder Order
		err := db.Where("platform_order_no = ?", ord.PlatformOrderNo).First(&existingOrder).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新订单
			if err := db.Create(ord).Error; err != nil {
				log.Printf("[SyncOrders] 创建订单失败: %v, 订单号: %s", err, ord.PlatformOrderNo)
				continue
			}
			result.Created++
		} else if err == nil {
			// 更新现有订单
			updates := map[string]interface{}{
				"status":          ord.Status,
				"platform_status": ord.PlatformStatus,
				"total_amount":    ord.TotalAmount,
				"ship_time":       ord.ShipTime,
				"ship_deadline":   ord.ShipDeadline,
			}
			if err := db.Model(&existingOrder).Updates(updates).Error; err != nil {
				continue
			}
			result.Updated++
		}
		result.Total++
	}

	// 更新最后同步时间
	now := time.Now()
	db.Model(auth).Update("last_sync_at", &now)

	// 同步佣金信息（异步执行，避免阻塞订单同步）
	go func() {
		// 收集本次同步的订单号
		postingNumbers := make([]string, 0, len(orders))
		for _, ord := range orders {
			postingNumbers = append(postingNumbers, ord.PlatformOrderNo)
		}

		if len(postingNumbers) == 0 {
			return
		}

		// 优先使用 GetCommissionsForOrders 接口（逐个订单获取）
		var commissions map[string]*CommissionData
		var err error

		if adapterWithOrders, ok := adapter.(PlatformAdapterWithOrders); ok {
			commissions, err = adapterWithOrders.GetCommissionsForOrders(credentials, postingNumbers, auth.ID)
		} else {
			// 兼容旧接口
			commissions, err = adapter.GetCommissions(credentials, since, to)
		}

		if err != nil {
			log.Printf("[SyncOrders] 佣金同步失败: %v", err)
			return
		}

		nowSync := time.Now()
		for postingNumber, commData := range commissions {
			db.Model(&Order{}).Where("platform_order_no = ?", postingNumber).Updates(map[string]interface{}{
				"accruals_for_sale":         commData.AccrualsForSale,
				"sale_commission":           commData.SaleCommission,
				"processing_and_delivery":   commData.ProcessingAndDelivery,
				"refunds_and_cancellations": commData.RefundsAndCancellations,
				"services_amount":           commData.ServicesAmount,
				"compensation_amount":       commData.CompensationAmount,
				"money_transfer":            commData.MoneyTransfer,
				"others_amount":             commData.OthersAmount,
				"profit_amount":             commData.ProfitAmount,
				"commission_currency":       commData.CommissionCurrency,
				"commission_synced_at":      &nowSync,
			})
		}
		log.Printf("[SyncOrders] 佣金同步完成: 更新 %d 条订单", len(commissions))
	}()

	return result, nil
}

// SyncOrderCommission 同步单个订单的佣金
func (s *Service) SyncOrderCommission(userID, orderID uint) (*Order, error) {
	db := database.GetDB()

	// 获取订单信息
	var ord Order
	if err := db.Where("id = ? AND user_id = ?", orderID, userID).First(&ord).Error; err != nil {
		return nil, ErrOrderNotFound
	}

	// 获取授权信息
	var auth PlatformAuth
	if err := db.Where("id = ? AND user_id = ?", ord.PlatformAuthID, userID).First(&auth).Error; err != nil {
		return nil, ErrAuthNotFound
	}

	// 获取适配器
	adapter := GetAdapter(auth.Platform)
	if adapter == nil {
		return nil, ErrPlatformNotFound
	}

	// 解密凭证
	credentials, err := Decrypt(auth.Credentials)
	if err != nil {
		return nil, fmt.Errorf("凭证解密失败: %w", err)
	}

	var commData *CommissionData

	// 使用单订单佣金接口（v3/finance/transaction/totals）
	if adapterWithLog, ok := adapter.(PlatformAdapterWithLog); ok {
		commData, err = adapterWithLog.GetSingleOrderCommission(credentials, ord.PlatformOrderNo, auth.ID)
		if err != nil {
			log.Printf("[SyncOrderCommission] 获取单订单佣金失败: %v", err)
			// 失败时不回退，直接返回空
			return &ord, nil
		}
	}

	// 更新订单佣金数据
	if commData != nil {
		now := time.Now()
		db.Model(&ord).Updates(map[string]interface{}{
			"accruals_for_sale":         commData.AccrualsForSale,
			"sale_commission":           commData.SaleCommission,
			"processing_and_delivery":   commData.ProcessingAndDelivery,
			"refunds_and_cancellations": commData.RefundsAndCancellations,
			"services_amount":           commData.ServicesAmount,
			"compensation_amount":       commData.CompensationAmount,
			"money_transfer":            commData.MoneyTransfer,
			"others_amount":             commData.OthersAmount,
			"profit_amount":             commData.ProfitAmount,
			"commission_currency":       commData.CommissionCurrency,
			"commission_synced_at":      &now,
		})
	}

	// 重新加载更新后的订单
	db.Where("id = ?", orderID).Preload("Items").First(&ord)

	return &ord, nil
}

// SyncSingleOrder 同步单个订单信息（从平台获取最新状态）
// 优先使用订单详情接口（/v3/posting/fbs/get），效率更高
func (s *Service) SyncSingleOrder(userID, orderID uint) (*Order, error) {
	db := database.GetDB()

	// 获取订单信息
	var ord Order
	if err := db.Where("id = ? AND user_id = ?", orderID, userID).First(&ord).Error; err != nil {
		return nil, ErrOrderNotFound
	}

	// 获取授权信息
	var auth PlatformAuth
	if err := db.Where("id = ? AND user_id = ?", ord.PlatformAuthID, userID).First(&auth).Error; err != nil {
		return nil, ErrAuthNotFound
	}

	// 获取适配器
	adapter := GetAdapter(auth.Platform)
	if adapter == nil {
		return nil, ErrPlatformNotFound
	}

	// 解密凭证
	credentials, err := Decrypt(auth.Credentials)
	if err != nil {
		return nil, fmt.Errorf("凭证解密失败: %w", err)
	}

	var matchedOrder *Order

	// 优先使用订单详情接口（更高效）
	if adapterWithDetail, ok := adapter.(PlatformAdapterWithOrderDetail); ok {
		matchedOrder, err = adapterWithDetail.GetOrderDetail(credentials, ord.PlatformOrderNo, auth.ID)
		if err != nil {
			log.Printf("[SyncSingleOrder] 订单详情接口失败，回退到列表接口: %v", err)
			matchedOrder = nil // 清空，使用回退逻辑
		}
	}

	// 如果订单详情接口不可用或失败，回退到列表接口
	if matchedOrder == nil {
		// 使用订单时间前后各1天作为同步范围
		var since, to time.Time
		if ord.OrderTime != nil {
			since = ord.OrderTime.Add(-24 * time.Hour)
			to = ord.OrderTime.Add(24 * time.Hour)
		} else {
			to = time.Now()
			since = to.AddDate(0, 0, -7)
		}

		var orders []*Order
		if adapterWithLog, ok := adapter.(PlatformAdapterWithLog); ok {
			orders, err = adapterWithLog.SyncOrdersWithLog(credentials, since, to, auth.ID)
		} else {
			orders, err = adapter.SyncOrders(credentials, since, to)
		}
		if err != nil {
			return nil, fmt.Errorf("从平台获取订单失败: %w", err)
		}

		// 查找匹配的订单
		for _, o := range orders {
			if o.PlatformOrderNo == ord.PlatformOrderNo {
				matchedOrder = o
				break
			}
		}

		if matchedOrder == nil {
			return nil, fmt.Errorf("平台未返回该订单信息")
		}
	}

	// 更新订单信息
	updates := map[string]interface{}{
		"status":          matchedOrder.Status,
		"platform_status": matchedOrder.PlatformStatus,
		"total_amount":    matchedOrder.TotalAmount,
		"ship_time":       matchedOrder.ShipTime,
		"ship_deadline":   matchedOrder.ShipDeadline,
	}
	if err := db.Model(&ord).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("更新订单失败: %w", err)
	}

	// 重新加载更新后的订单
	db.Where("id = ?", orderID).Preload("Items").First(&ord)

	log.Printf("[SyncSingleOrder] 订单 %s 同步成功，状态: %s -> %s", ord.PlatformOrderNo, ord.Status, matchedOrder.Status)

	return &ord, nil
}

// ListOrders 获取订单列表
func (s *Service) ListOrders(userID uint, req *OrderListRequest) ([]Order, int64, error) {
	db := database.GetDB()

	var orders []Order
	var total int64

	query := db.Model(&Order{}).Where("user_id = ?", userID)

	// 应用过滤条件
	if req.Platform != "" {
		query = query.Where("platform = ?", req.Platform)
	}
	if req.AuthID > 0 {
		query = query.Where("platform_auth_id = ?", req.AuthID)
	}
	if req.Status != "" {
		// 支持多状态筛选，逗号分隔
		if strings.Contains(req.Status, ",") {
			statuses := strings.Split(req.Status, ",")
			query = query.Where("status IN ?", statuses)
		} else {
			query = query.Where("status = ?", req.Status)
		}
	}
	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		query = query.Where("platform_order_no LIKE ? OR recipient_name LIKE ?", keyword, keyword)
	}
	// 时间过滤 - 直接拼接 SQL 避免 GORM 的参数时区转换
	if req.StartTime != "" {
		startTimeStr, _ := url.QueryUnescape(req.StartTime)
		if startTimeStr == "" {
			startTimeStr = req.StartTime
		}
		if len(startTimeStr) == 10 {
			startTimeStr = startTimeStr + " 00:00:00"
		}
		query = query.Where(fmt.Sprintf("order_time >= '%s'", startTimeStr))
	}
	if req.EndTime != "" {
		endTimeStr, _ := url.QueryUnescape(req.EndTime)
		if endTimeStr == "" {
			endTimeStr = req.EndTime
		}
		if len(endTimeStr) == 10 {
			endTimeStr = endTimeStr + " 23:59:59"
		}
		query = query.Where(fmt.Sprintf("order_time <= '%s'", endTimeStr))
	}

	// 发货截止时间筛选
	if req.DeadlineFilter != "" {
		now := time.Now()
		switch req.DeadlineFilter {
		case "overdue":
			// 已逾期：截止时间已过且订单未发货
			query = query.Where("ship_deadline IS NOT NULL AND ship_deadline < ? AND status IN ?",
				now, []string{OrderStatusPending, OrderStatusReadyToShip})
		case "within_1d":
			// 1天内：截止时间在当前到1天后之间
			deadline := now.Add(24 * time.Hour)
			query = query.Where("ship_deadline IS NOT NULL AND ship_deadline >= ? AND ship_deadline < ? AND status IN ?",
				now, deadline, []string{OrderStatusPending, OrderStatusReadyToShip})
		case "within_3d":
			// 3天内：截止时间在当前到3天后之间
			deadline := now.Add(72 * time.Hour)
			query = query.Where("ship_deadline IS NOT NULL AND ship_deadline >= ? AND ship_deadline < ? AND status IN ?",
				now, deadline, []string{OrderStatusPending, OrderStatusReadyToShip})
		}
	}

	query.Count(&total)

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Items").Offset(offset).Limit(pageSize).Order("order_time DESC").Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// GetOrderByID 根据ID获取订单详情
func (s *Service) GetOrderByID(id uint, userID uint) (*Order, error) {
	db := database.GetDB()

	var ord Order
	if err := db.Preload("Items").Where("id = ? AND user_id = ?", id, userID).First(&ord).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}

	return &ord, nil
}

// SyncCommission 同步佣金信息（使用 transaction/totals 逐个订单获取）
func (s *Service) SyncCommission(userID, authID uint, since, to time.Time) (*SyncCommissionResponse, error) {
	db := database.GetDB()

	auth, err := s.GetAuthByID(authID, userID)
	if err != nil {
		return nil, err
	}

	adapter := GetAdapter(auth.Platform)
	if adapter == nil {
		return nil, fmt.Errorf("平台 %s 适配器未找到", auth.Platform)
	}

	// 解密凭证
	credentials, err := Decrypt(auth.Credentials)
	if err != nil {
		return nil, fmt.Errorf("凭证解密失败: %w", err)
	}

	// 先从数据库获取该时间范围内的订单列表
	var orders []Order
	query := db.Model(&Order{}).
		Where("platform_auth_id = ?", authID).
		Where("order_time >= ? AND order_time <= ?", since, to)
	if err := query.Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("查询订单失败: %w", err)
	}

	// 提取订单号列表
	postingNumbers := make([]string, 0, len(orders))
	for _, ord := range orders {
		postingNumbers = append(postingNumbers, ord.PlatformOrderNo)
	}

	log.Printf("[SyncCommission] 找到 %d 个订单需要同步佣金", len(postingNumbers))

	if len(postingNumbers) == 0 {
		return &SyncCommissionResponse{Total: 0, Updated: 0}, nil
	}

	// 使用新的 GetCommissionsForOrders 方法逐个获取佣金
	var commissions map[string]*CommissionData
	if adapterWithOrders, ok := adapter.(PlatformAdapterWithOrders); ok {
		commissions, err = adapterWithOrders.GetCommissionsForOrders(credentials, postingNumbers, auth.ID)
	} else {
		// 兼容旧接口
		if adapterWithLog, ok := adapter.(PlatformAdapterWithLog); ok {
			commissions, err = adapterWithLog.GetCommissionsWithLog(credentials, since, to, auth.ID)
		} else {
			commissions, err = adapter.GetCommissions(credentials, since, to)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("获取佣金数据失败: %w", err)
	}

	result := &SyncCommissionResponse{
		Total: len(postingNumbers),
	}

	// 批量更新订单佣金
	now := time.Now()
	for postingNumber, commData := range commissions {
		updateResult := db.Model(&Order{}).
			Where("platform_order_no = ? AND platform_auth_id = ?", postingNumber, authID).
			Updates(map[string]interface{}{
				"accruals_for_sale":         commData.AccrualsForSale,
				"sale_commission":           commData.SaleCommission,
				"processing_and_delivery":   commData.ProcessingAndDelivery,
				"refunds_and_cancellations": commData.RefundsAndCancellations,
				"services_amount":           commData.ServicesAmount,
				"compensation_amount":       commData.CompensationAmount,
				"money_transfer":            commData.MoneyTransfer,
				"others_amount":             commData.OthersAmount,
				"profit_amount":             commData.ProfitAmount,
				"commission_currency":       commData.CommissionCurrency,
				"commission_synced_at":      &now,
			})
		if updateResult.RowsAffected > 0 {
			result.Updated++
		}
	}

	return result, nil
}

// SaveRequestLog 保存请求日志
func SaveRequestLog(logEntry *OrdersRequestLog) error {
	db := database.GetDB()
	return db.Create(logEntry).Error
}

// CashFlowSyncResult 现金流同步结果
type CashFlowSyncResult struct {
	Total   int `json:"total"`
	Created int `json:"created"`
	Updated int `json:"updated"`
	Skipped int `json:"skipped"`
}

// SyncCashFlowStatements 同步现金流报表
func (s *Service) SyncCashFlowStatements(authID, userID uint, since, to time.Time) (*CashFlowSyncResult, error) {
	db := database.GetDB()
	result := &CashFlowSyncResult{}

	// 获取授权信息
	var auth PlatformAuth
	if err := db.First(&auth, authID).Error; err != nil {
		return nil, fmt.Errorf("获取授权信息失败: %w", err)
	}

	if auth.UserID != userID {
		return nil, fmt.Errorf("无权访问该授权")
	}

	// 获取平台适配器
	baseAdapter := GetAdapter(auth.Platform)
	if baseAdapter == nil {
		return nil, ErrPlatformNotFound
	}

	// 检查是否支持现金流接口
	adapter, ok := baseAdapter.(PlatformAdapterWithCashFlow)
	if !ok {
		return nil, fmt.Errorf("平台 %s 不支持现金流报表", auth.Platform)
	}

	// 解密凭证
	credentials, err := Decrypt(auth.Credentials)
	if err != nil {
		return nil, fmt.Errorf("解密凭证失败: %w", err)
	}

	// 获取现金流报表
	cashFlows, err := adapter.GetCashFlowStatements(credentials, since, to, authID)
	if err != nil {
		return nil, fmt.Errorf("获取现金流报表失败: %w", err)
	}

	result.Total = len(cashFlows)

	// 保存到数据库
	for _, cf := range cashFlows {
		if cf.PeriodBegin == nil {
			result.Skipped++
			continue
		}

		// 检查是否已存在（使用 platform_auth_id + period_begin 作为唯一标识）
		var existing CashFlowStatement
		err := db.Where("platform_auth_id = ? AND period_begin = ?", authID, cf.PeriodBegin).First(&existing).Error

		if err == nil {
			// 更新现有记录
			existing.OrdersAmount = cf.OrdersAmount
			existing.ReturnsAmount = cf.ReturnsAmount
			existing.CommissionAmount = cf.CommissionAmount
			existing.ServicesAmount = cf.ServicesAmount
			existing.ItemDeliveryAndReturnAmount = cf.ItemDeliveryAndReturnAmount
			existing.CurrencyCode = cf.CurrencyCode
			existing.PeriodEnd = cf.PeriodEnd
			existing.SyncedAt = time.Now()
			if err := db.Save(&existing).Error; err != nil {
				continue
			}
			result.Updated++
		} else {
			// 创建新记录
			cf.UserID = userID
			cf.PlatformAuthID = authID
			cf.Platform = auth.Platform
			cf.SyncedAt = time.Now()
			if err := db.Create(&cf).Error; err != nil {
				continue
			}
			result.Created++
		}
	}

	return result, nil
}

// ListCashFlowStatements 查询现金流报表列表
func (s *Service) ListCashFlowStatements(userID uint, authID uint, page, pageSize int) ([]CashFlowStatement, int64, error) {
	db := database.GetDB()
	var statements []CashFlowStatement
	var total int64

	query := db.Model(&CashFlowStatement{}).Where("user_id = ?", userID)
	if authID > 0 {
		query = query.Where("platform_auth_id = ?", authID)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("period_end DESC").Offset(offset).Limit(pageSize).Find(&statements).Error; err != nil {
		return nil, 0, err
	}

	return statements, total, nil
}

// GetCashFlowStatement 获取现金流报表详情
func (s *Service) GetCashFlowStatement(id, userID uint) (*CashFlowStatement, error) {
	db := database.GetDB()
	var statement CashFlowStatement

	if err := db.Where("id = ? AND user_id = ?", id, userID).First(&statement).Error; err != nil {
		return nil, err
	}

	return &statement, nil
}

// GetMutualSettlement 获取结算报告
// API: POST /v1/finance/mutual-settlement
func (s *Service) GetMutualSettlement(authID, userID uint, since, to time.Time) (interface{}, error) {
	db := database.GetDB()

	// 获取授权信息
	var auth PlatformAuth
	if err := db.First(&auth, authID).Error; err != nil {
		return nil, ErrAuthNotFound
	}

	if auth.UserID != userID {
		return nil, fmt.Errorf("无权访问该授权")
	}

	// 目前只支持 Ozon
	if auth.Platform != PlatformOzon {
		return nil, fmt.Errorf("平台 %s 暂不支持结算报告", auth.Platform)
	}

	// 解密凭证
	credentials, err := Decrypt(auth.Credentials)
	if err != nil {
		return nil, fmt.Errorf("解密凭证失败: %w", err)
	}

	// 解析凭证并创建客户端
	creds, err := ozon.ParseCredentials(credentials)
	if err != nil {
		return nil, fmt.Errorf("解析凭证失败: %w", err)
	}

	// 创建带日志记录器的客户端
	logger := platform.NewRequestLogger(db, authID, PlatformOzon)
	client := ozon.NewClient(creds, logger)
	financeAPI := ozon.NewFinanceAPI(client)

	// 调用结算报告接口（支持异步报告轮询）
	result, err := financeAPI.GetMutualSettlementWithReport(since, to, 15)
	if err != nil {
		return nil, fmt.Errorf("获取结算报告失败: %w", err)
	}

	return result, nil
}

// GetDashboardStats 获取仪表盘统计数据
func (s *Service) GetDashboardStats(userID uint) (*DashboardStatsResponse, error) {
	db := database.GetDB()
	stats := &DashboardStatsResponse{
		Currency: "RUB",
	}

	// 订单统计查询基础条件
	orderQuery := db.Model(&Order{}).Where("user_id = ?", userID)

	// 总订单数
	orderQuery.Count(&stats.TotalOrders)

	// 已签收订单数
	db.Model(&Order{}).Where("user_id = ? AND status = ?", userID, OrderStatusDelivered).Count(&stats.DeliveredOrders)

	// 待处理订单数（待处理+待发货状态）
	db.Model(&Order{}).Where("user_id = ? AND status IN ?", userID, []string{OrderStatusPending, OrderStatusReadyToShip}).Count(&stats.PendingOrders)

	// 已发货订单数
	db.Model(&Order{}).Where("user_id = ? AND status = ?", userID, OrderStatusShipped).Count(&stats.ShippedOrders)

	// 即将超时订单数（待处理+待发货状态，且发货截止时间距当前不足1天）
	now := time.Now()
	deadline := now.Add(24 * time.Hour) // 1天后
	db.Model(&Order{}).Where(
		"user_id = ? AND status IN ? AND ship_deadline IS NOT NULL AND ship_deadline < ?",
		userID, []string{OrderStatusPending, OrderStatusReadyToShip}, deadline,
	).Count(&stats.TimeoutOrders)

	// 今日订单数（使用本地时区的今天零点）
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	db.Model(&Order{}).Where("user_id = ? AND order_time >= ?", userID, today).Count(&stats.TodayOrders)

	// 订单总金额（所有非取消订单，按币种分别统计）
	var totalAmounts []CurrencyAmount
	db.Model(&Order{}).
		Select("currency, COALESCE(SUM(total_amount), 0) as amount").
		Where("user_id = ? AND status != ?", userID, OrderStatusCancelled).
		Group("currency").
		Scan(&totalAmounts)
	stats.TotalAmounts = totalAmounts

	// 佣金统计（已签收订单）
	var commissionStats struct {
		TotalProfit     float64
		TotalCommission float64
		TotalServiceFee float64
	}
	db.Model(&Order{}).
		Select(`
			COALESCE(SUM(profit_amount), 0) as total_profit,
			COALESCE(SUM(sale_commission), 0) as total_commission,
			COALESCE(SUM(services_amount), 0) as total_service_fee
		`).
		Where("user_id = ? AND status = ?", userID, OrderStatusDelivered).
		Scan(&commissionStats)
	stats.TotalProfit = commissionStats.TotalProfit
	stats.TotalCommission = commissionStats.TotalCommission
	stats.TotalServiceFee = commissionStats.TotalServiceFee

	// 授权统计
	db.Model(&PlatformAuth{}).Where("user_id = ?", userID).Count(&stats.TotalAuths)
	db.Model(&PlatformAuth{}).Where("user_id = ? AND status = ?", userID, AuthStatusActive).Count(&stats.ActiveAuths)

	return stats, nil
}

// GetRecentOrders 获取最近订单
func (s *Service) GetRecentOrders(userID uint, limit int) ([]RecentOrderResponse, error) {
	db := database.GetDB()

	if limit <= 0 {
		limit = 10
	}

	var orders []Order
	if err := db.Where("user_id = ?", userID).
		Order("order_time DESC").
		Limit(limit).
		Find(&orders).Error; err != nil {
		return nil, err
	}

	result := make([]RecentOrderResponse, len(orders))
	for i, order := range orders {
		result[i] = RecentOrderResponse{
			ID:              order.ID,
			PlatformOrderNo: order.PlatformOrderNo,
			Status:          order.Status,
			TotalAmount:     order.TotalAmount,
			Currency:        order.Currency,
			OrderTime:       order.OrderTime,
		}
	}

	return result, nil
}

// InitOrderTrendStats 初始化订单走势统计数据（如果不存在则计算）
// forceUpdate: 为true时强制重新计算，忽略已有数据
func (s *Service) InitOrderTrendStats(userID uint, forceUpdate bool) error {
	db := database.GetDB()

	// 检查是否有统计数据
	var count int64
	db.Model(&OrderDailyStat{}).Where("user_id = ?", userID).Count(&count)

	if count > 0 && !forceUpdate {
		// 已有数据且非强制更新，无需初始化
		return nil
	}

	// 强制更新时，先删除旧数据
	if forceUpdate && count > 0 {
		db.Where("user_id = ?", userID).Delete(&OrderDailyStat{})
		log.Printf("[Service] 用户 %d 强制刷新，已清除旧统计数据", userID)
	}

	// 执行统计（最近30天）
	log.Printf("[Service] 用户 %d %s订单走势统计...", userID, map[bool]string{true: "刷新", false: "首次访问，初始化"}[forceUpdate])

	days := 30
	endDate := time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour)
	startDate := endDate.AddDate(0, 0, -days)

	type DailyStats struct {
		Date     time.Time
		Currency string
		Count    int64
		Amount   float64
	}

	var dailyStats []DailyStats
	err := db.Model(&Order{}).
		Select(`
			DATE(order_time) as date, 
			currency,
			COUNT(*) as count, 
			COALESCE(SUM(total_amount), 0) as amount
		`).
		Where("user_id = ? AND order_time >= ? AND order_time < ?", userID, startDate, endDate).
		Group("DATE(order_time), currency").
		Order("date ASC").
		Scan(&dailyStats).Error

	if err != nil {
		return err
	}

	log.Printf("[Service] 用户 %d 共查询到 %d 条日期统计数据", userID, len(dailyStats))

	// 存储统计数据
	for _, stat := range dailyStats {
		newStat := OrderDailyStat{
			UserID:      userID,
			StatDate:    stat.Date,
			Currency:    stat.Currency,
			OrderCount:  stat.Count,
			OrderAmount: stat.Amount,
		}
		db.Create(&newStat)
	}

	log.Printf("[Service] 用户 %d 订单走势统计初始化完成，共 %d 条记录", userID, len(dailyStats))
	return nil
}

// GetOrderTrend 获取订单趋势数据（优先从统计表读取，回退到实时查询）
func (s *Service) GetOrderTrend(userID uint, days int, currency string) (*OrderTrendResponse, error) {
	db := database.GetDB()

	if days <= 0 {
		days = 7
	}
	// 如果未指定币种，默认为 RUB
	if currency == "" {
		currency = "RUB"
	}

	// 计算起始日期（当天往前推days天）
	endDate := time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour) // 明天0点
	startDate := endDate.AddDate(0, 0, -days)                          // N天前

	// 优先从 order_daily_stats 表读取
	var cachedStats []OrderDailyStat
	err := db.Where("user_id = ? AND currency = ? AND stat_date >= ? AND stat_date < ?", userID, currency, startDate, endDate).
		Order("stat_date ASC").
		Find(&cachedStats).Error

	// 构建日期映射
	statsMap := make(map[string]OrderTrendItem)

	if err == nil && len(cachedStats) > 0 {
		// 使用缓存数据
		for _, stat := range cachedStats {
			date := stat.StatDate.Format("2006-01-02")
			statsMap[date] = OrderTrendItem{
				Date:   date,
				Count:  stat.OrderCount,
				Amount: stat.OrderAmount,
			}
		}
	} else {
		// 回退到实时查询
		type DailyStats struct {
			Date   string
			Count  int64
			Amount float64
		}

		var dailyStats []DailyStats
		err = db.Model(&Order{}).
			Select("DATE(order_time) as date, COUNT(*) as count, COALESCE(SUM(total_amount), 0) as amount").
			Where("user_id = ? AND currency = ? AND order_time >= ? AND order_time < ?", userID, currency, startDate, endDate).
			Group("DATE(order_time)").
			Order("date ASC").
			Scan(&dailyStats).Error

		if err != nil {
			return nil, err
		}

		for _, stat := range dailyStats {
			statsMap[stat.Date] = OrderTrendItem{
				Date:   stat.Date,
				Count:  stat.Count,
				Amount: stat.Amount,
			}
		}
	}

	// 填充所有日期（确保连续）
	items := make([]OrderTrendItem, days)
	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i).Format("2006-01-02")
		if stat, ok := statsMap[date]; ok {
			items[i] = stat
		} else {
			items[i] = OrderTrendItem{
				Date:   date,
				Count:  0,
				Amount: 0,
			}
		}
	}

	return &OrderTrendResponse{Items: items}, nil
}

// orderSummaryRaw 原始查询结果
type orderSummaryRaw struct {
	LocalSKU     string  `gorm:"column:local_sku"`
	ProductName  string  `gorm:"column:product_name"`
	PlatformSKU  string  `gorm:"column:platform_sku"`
	PlatformName string  `gorm:"column:platform_name"`
	PlatformImg  string  `gorm:"column:platform_img"`
	Status       string  `gorm:"column:status"`
	Quantity     int     `gorm:"column:quantity"`
	Amount       float64 `gorm:"column:amount"`
	Currency     string  `gorm:"column:currency"`
}

// GetOrderSummary 获取订单汇总（按本地SKU合并）
func (s *Service) GetOrderSummary(userID uint, req *OrderSummaryRequest) ([]OrderSummaryItem, error) {
	db := database.GetDB()
	var rawItems []orderSummaryRaw

	query := db.Table("order_items oi").
		Select(`
			COALESCE(p.sku, '') as local_sku,
			COALESCE(p.name, '') as product_name,
			oi.sku as platform_sku,
			COALESCE(pp.name, '') as platform_name,
			COALESCE(pp.image, '') as platform_img,
			o.status,
			SUM(oi.quantity) as quantity,
			SUM(oi.price * oi.quantity) as amount,
			oi.currency
		`).
		Joins("JOIN orders o ON oi.order_id = o.id").
		Joins("LEFT JOIN platform_products pp ON pp.platform_sku = oi.sku AND pp.platform_auth_id = o.platform_auth_id").
		Joins("LEFT JOIN product_mappings pm ON pm.platform_product_id = pp.id").
		Joins("LEFT JOIN products p ON p.id = pm.product_id").
		Where("o.user_id = ?", userID).
		Group("local_sku, product_name, platform_sku, platform_name, platform_img, o.status, oi.currency")

	// 过滤条件
	if req.Platform != "" {
		query = query.Where("o.platform = ?", req.Platform)
	}
	if req.AuthID > 0 {
		query = query.Where("o.platform_auth_id = ?", req.AuthID)
	}
	if req.StartTime != "" {
		startTimeStr, _ := url.QueryUnescape(req.StartTime)
		if startTimeStr == "" {
			startTimeStr = req.StartTime
		}
		if len(startTimeStr) == 10 {
			startTimeStr = startTimeStr + " 00:00:00"
		}
		query = query.Where("o.order_time >= ?", startTimeStr)
	}
	if req.EndTime != "" {
		endTimeStr, _ := url.QueryUnescape(req.EndTime)
		if endTimeStr == "" {
			endTimeStr = req.EndTime
		}
		if len(endTimeStr) == 10 {
			endTimeStr = endTimeStr + " 23:59:59"
		}
		query = query.Where("o.order_time <= ?", endTimeStr)
	}
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		query = query.Where("(p.sku LIKE ? OR p.name LIKE ? OR oi.sku LIKE ?)", like, like, like)
	}
	if req.Status != "" {
		query = query.Where("o.status = ?", req.Status)
	}

	if err := query.Scan(&rawItems).Error; err != nil {
		return nil, err
	}

	// 按本地SKU聚合
	skuMap := make(map[string]*OrderSummaryItem)
	skuOrder := []string{} // 保持顺序

	for _, raw := range rawItems {
		key := raw.LocalSKU
		if key == "" {
			key = "_unmapped_" + raw.PlatformSKU // 未关联的按平台SKU区分
		}

		item, exists := skuMap[key]
		if !exists {
			item = &OrderSummaryItem{
				LocalSKU:         raw.LocalSKU,
				ProductName:      raw.ProductName,
				PlatformSKUs:     []string{},
				PlatformProducts: []OrderSummaryPlatformSKU{},
				Quantity:         0,
				Amount:           0,
				Currency:         raw.Currency,
				StatusDetails:    []OrderSummaryStatusDetail{},
			}
			skuMap[key] = item
			skuOrder = append(skuOrder, key)
		}

		// 累加总数量和总金额
		item.Quantity += raw.Quantity
		item.Amount += raw.Amount

		// 添加平台SKU和详情（去重）
		found := false
		for _, sku := range item.PlatformSKUs {
			if sku == raw.PlatformSKU {
				found = true
				break
			}
		}
		if !found {
			item.PlatformSKUs = append(item.PlatformSKUs, raw.PlatformSKU)
			item.PlatformProducts = append(item.PlatformProducts, OrderSummaryPlatformSKU{
				SKU:   raw.PlatformSKU,
				Name:  raw.PlatformName,
				Image: raw.PlatformImg,
			})
		}

		// 添加或更新状态明细（使用中文状态名称）
		statusLabel := GetOrderStatusName(raw.Status)
		statusFound := false
		for i, sd := range item.StatusDetails {
			if sd.Status == statusLabel {
				item.StatusDetails[i].Quantity += raw.Quantity
				item.StatusDetails[i].Amount += raw.Amount
				statusFound = true
				break
			}
		}
		if !statusFound {
			item.StatusDetails = append(item.StatusDetails, OrderSummaryStatusDetail{
				Status:   statusLabel,
				Quantity: raw.Quantity,
				Amount:   raw.Amount,
			})
		}
	}

	// 按顺序输出结果
	result := make([]OrderSummaryItem, 0, len(skuOrder))
	for _, key := range skuOrder {
		result = append(result, *skuMap[key])
	}

	// 查询系统可用库存
	if len(result) > 0 {
		skus := make([]string, 0, len(result))
		for _, item := range result {
			if item.LocalSKU != "" {
				skus = append(skus, item.LocalSKU)
			}
		}

		if len(skus) > 0 {
			// 查询库存（按SKU汇总所有仓库的可用库存）
			type StockResult struct {
				SKU            string
				AvailableStock int
			}
			var stocks []StockResult
			db.Table("warehouse_center_inventory").
				Select("sku, SUM(available_stock) as available_stock").
				Where("sku IN ?", skus).
				Group("sku").
				Scan(&stocks)

			// 构建库存映射
			stockMap := make(map[string]int)
			for _, s := range stocks {
				stockMap[s.SKU] = s.AvailableStock
			}

			// 更新结果
			for i := range result {
				if stock, ok := stockMap[result[i].LocalSKU]; ok {
					result[i].AvailableStock = stock
				}
			}
		}
	}

	return result, nil
}
