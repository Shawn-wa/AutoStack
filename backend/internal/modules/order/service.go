package order

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"

	"autostack/internal/commonBase/database"
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

	if req.Credentials != nil && len(req.Credentials) > 0 {
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

	// 同步佣金信息（方案A：订单同步时同步佣金）
	go func() {
		// 异步获取佣金，避免阻塞订单同步
		commissions, err := adapter.GetCommissions(credentials, since, to)
		if err != nil {
			// 佣金获取失败不影响订单同步结果，仅记录日志
			return
		}

		nowSync := time.Now()
		for postingNumber, commData := range commissions {
			db.Model(&Order{}).Where("platform_order_no = ?", postingNumber).Updates(map[string]interface{}{
				"sale_commission":        commData.SaleCommission,
				"accruals_for_sale":      commData.AccrualsForSale,
				"delivery_charge":        commData.DeliveryCharge,
				"return_delivery_charge": commData.ReturnDeliveryCharge,
				"commission_amount":      commData.CommissionAmount,
				"commission_currency":    commData.CommissionCurrency,
				"commission_synced_at":   &nowSync,
			})
		}
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

	// 使用订单时间范围获取佣金（订单时间前后各7天）
	orderTime := time.Now()
	if ord.OrderTime != nil {
		orderTime = *ord.OrderTime
	}
	since := orderTime.AddDate(0, 0, -7)
	to := orderTime.AddDate(0, 0, 7)

	commissions, err := adapter.GetCommissions(credentials, since, to)
	if err != nil {
		return nil, fmt.Errorf("获取佣金失败: %w", err)
	}

	// 查找该订单的佣金数据
	if commData, exists := commissions[ord.PlatformOrderNo]; exists {
		now := time.Now()
		db.Model(&ord).Updates(map[string]interface{}{
			"sale_commission":        commData.SaleCommission,
			"accruals_for_sale":      commData.AccrualsForSale,
			"delivery_charge":        commData.DeliveryCharge,
			"return_delivery_charge": commData.ReturnDeliveryCharge,
			"commission_amount":      commData.CommissionAmount,
			"commission_currency":    commData.CommissionCurrency,
			"commission_synced_at":   &now,
		})

		// 重新加载更新后的订单
		db.Where("id = ?", orderID).Preload("Items").First(&ord)
	}

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
		query = query.Where("status = ?", req.Status)
	}
	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		query = query.Where("platform_order_no LIKE ? OR recipient_name LIKE ?", keyword, keyword)
	}
	if req.StartTime != "" {
		if t, err := time.Parse("2006-01-02", req.StartTime); err == nil {
			query = query.Where("order_time >= ?", t)
		}
	}
	if req.EndTime != "" {
		if t, err := time.Parse("2006-01-02", req.EndTime); err == nil {
			query = query.Where("order_time <= ?", t.Add(24*time.Hour))
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

// SyncCommission 同步佣金信息（方案B：独立同步接口）
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

	// 获取佣金数据（优先使用带日志的方法）
	var commissions map[string]*CommissionData
	if adapterWithLog, ok := adapter.(PlatformAdapterWithLog); ok {
		commissions, err = adapterWithLog.GetCommissionsWithLog(credentials, since, to, auth.ID)
	} else {
		commissions, err = adapter.GetCommissions(credentials, since, to)
	}
	if err != nil {
		return nil, fmt.Errorf("获取佣金数据失败: %w", err)
	}

	result := &SyncCommissionResponse{
		Total: len(commissions),
	}

	// 批量更新订单佣金
	now := time.Now()
	for postingNumber, commData := range commissions {
		updateResult := db.Model(&Order{}).
			Where("platform_order_no = ? AND platform_auth_id = ?", postingNumber, authID).
			Updates(map[string]interface{}{
				"sale_commission":        commData.SaleCommission,
				"accruals_for_sale":      commData.AccrualsForSale,
				"delivery_charge":        commData.DeliveryCharge,
				"return_delivery_charge": commData.ReturnDeliveryCharge,
				"commission_amount":      commData.CommissionAmount,
				"commission_currency":    commData.CommissionCurrency,
				"commission_synced_at":   &now,
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
