package platforms

import (
	"log"
	"strconv"
	"time"

	"gorm.io/gorm"

	"autostack/internal/commonBase/database"
	"autostack/internal/modules/apiClient/platform"
	"autostack/internal/modules/apiClient/platform/ozon"
	"autostack/internal/modules/order"
)

// OzonAdapter Ozon平台适配器
type OzonAdapter struct{}

// GetPlatformName 获取平台标识
func (a *OzonAdapter) GetPlatformName() string {
	return order.PlatformOzon
}

// GetPlatformLabel 获取平台显示名称
func (a *OzonAdapter) GetPlatformLabel() string {
	return "Ozon"
}

// GetCredentialFields 获取凭证字段定义
func (a *OzonAdapter) GetCredentialFields() []order.CredentialField {
	return []order.CredentialField{
		{Key: "client_id", Label: "Client ID", Type: "text", Required: true},
		{Key: "api_key", Label: "API Key", Type: "password", Required: true},
		{Key: "settlement_currency", Label: "结算货币", Type: "text", Required: false},
	}
}

// ========== 基础接口实现 ==========

// TestConnection 测试连接
func (a *OzonAdapter) TestConnection(credentials string) error {
	return a.TestConnectionWithLog(credentials, 0)
}

// SyncOrders 同步订单
func (a *OzonAdapter) SyncOrders(credentials string, since, to time.Time) ([]*order.Order, error) {
	return a.SyncOrdersWithLog(credentials, since, to, 0)
}

// GetCommissions 获取佣金信息
func (a *OzonAdapter) GetCommissions(credentials string, since, to time.Time) (map[string]*order.CommissionData, error) {
	return a.GetCommissionsWithLog(credentials, since, to, 0)
}

// ========== 带日志接口实现 ==========

// TestConnectionWithLog 测试连接（带日志记录）
func (a *OzonAdapter) TestConnectionWithLog(credentials string, platformAuthID uint) error {
	client, err := a.createClient(credentials, platformAuthID)
	if err != nil {
		return err
	}

	orderAPI := ozon.NewOrderAPI(client)
	return orderAPI.TestConnection()
}

// SyncOrdersWithLog 同步订单（带日志记录）
// 注意：Ozon API 对时间范围有限制（约30天），超过限制会返回 PERIOD_IS_TOO_LONG 错误
// 此方法会自动将长时间范围拆分成多个30天批次进行请求
func (a *OzonAdapter) SyncOrdersWithLog(credentials string, since, to time.Time, platformAuthID uint) ([]*order.Order, error) {
	client, err := a.createClient(credentials, platformAuthID)
	if err != nil {
		return nil, err
	}

	orderAPI := ozon.NewOrderAPI(client)
	var allOrders []*order.Order

	// Ozon API 时间范围限制为30天，将长时间范围拆分成多个批次
	const maxDays = 30
	batchStart := since

	for batchStart.Before(to) {
		batchEnd := batchStart.AddDate(0, 0, maxDays)
		if batchEnd.After(to) {
			batchEnd = to
		}

		log.Printf("[OZON] 同步订单批次: %s ~ %s", batchStart.Format("2006-01-02"), batchEnd.Format("2006-01-02"))

		// 分页获取该批次的订单
		offset := 0
		limit := 100

		for {
			resp, err := orderAPI.GetOrderList(batchStart, batchEnd, offset, limit)
			if err != nil {
				return nil, err
			}

			log.Printf("[OZON] 同步订单: offset=%d, 获取到 %d 条订单", offset, len(resp.Result.Postings))

			if len(resp.Result.Postings) == 0 {
				break
			}

			for _, posting := range resp.Result.Postings {
				ord := a.convertToOrder(&posting)
				allOrders = append(allOrders, ord)
			}

			if len(resp.Result.Postings) < limit {
				break
			}

			offset += limit
		}

		// 移动到下一批次
		batchStart = batchEnd
	}

	log.Printf("[OZON] 同步完成: 共获取 %d 条订单", len(allOrders))
	return allOrders, nil
}

// GetCommissionsWithLog 获取佣金信息（带日志记录）
// 注意：此方法已废弃，保留以兼容接口定义
// 新逻辑使用 GetCommissionsForOrders 逐个订单获取佣金
func (a *OzonAdapter) GetCommissionsWithLog(credentials string, since, to time.Time, platformAuthID uint) (map[string]*order.CommissionData, error) {
	// 返回空结果，实际业务已改用 GetCommissionsForOrders
	return make(map[string]*order.CommissionData), nil
}

// GetCommissionsForOrders 获取指定订单的佣金信息（使用 transaction/totals 接口）
func (a *OzonAdapter) GetCommissionsForOrders(credentials string, postingNumbers []string, platformAuthID uint) (map[string]*order.CommissionData, error) {
	creds, err := ozon.ParseCredentials(credentials)
	if err != nil {
		return nil, err
	}

	client, err := a.createClient(credentials, platformAuthID)
	if err != nil {
		return nil, err
	}

	financeAPI := ozon.NewFinanceAPI(client)
	result := make(map[string]*order.CommissionData)

	for _, postingNumber := range postingNumbers {
		commData, err := financeAPI.GetSingleOrderCommission(postingNumber, creds.SettlementCurrency)
		if err != nil {
			log.Printf("[OZON] 获取订单 %s 佣金失败: %v", postingNumber, err)
			continue
		}

		result[postingNumber] = &order.CommissionData{
			AccrualsForSale:         commData.AccrualsForSale,
			SaleCommission:          commData.SaleCommission,
			ProcessingAndDelivery:   commData.ProcessingAndDelivery,
			RefundsAndCancellations: commData.RefundsAndCancellations,
			ServicesAmount:          commData.ServicesAmount,
			CompensationAmount:      commData.CompensationAmount,
			MoneyTransfer:           commData.MoneyTransfer,
			OthersAmount:            commData.OthersAmount,
			ProfitAmount:            commData.ProfitAmount,
			CommissionCurrency:      commData.CommissionCurrency,
		}
	}

	return result, nil
}

// GetSingleOrderCommission 获取单个订单的佣金信息
func (a *OzonAdapter) GetSingleOrderCommission(credentials string, postingNumber string, platformAuthID uint) (*order.CommissionData, error) {
	creds, err := ozon.ParseCredentials(credentials)
	if err != nil {
		return nil, err
	}

	client, err := a.createClient(credentials, platformAuthID)
	if err != nil {
		return nil, err
	}

	financeAPI := ozon.NewFinanceAPI(client)
	commData, err := financeAPI.GetSingleOrderCommission(postingNumber, creds.SettlementCurrency)
	if err != nil {
		return nil, err
	}

	return &order.CommissionData{
		AccrualsForSale:         commData.AccrualsForSale,
		SaleCommission:          commData.SaleCommission,
		ProcessingAndDelivery:   commData.ProcessingAndDelivery,
		RefundsAndCancellations: commData.RefundsAndCancellations,
		ServicesAmount:          commData.ServicesAmount,
		CompensationAmount:      commData.CompensationAmount,
		MoneyTransfer:           commData.MoneyTransfer,
		OthersAmount:            commData.OthersAmount,
		ProfitAmount:            commData.ProfitAmount,
		CommissionCurrency:      commData.CommissionCurrency,
	}, nil
}

// GetCashFlowStatements 获取现金流报表
func (a *OzonAdapter) GetCashFlowStatements(credentials string, since, to time.Time, platformAuthID uint) ([]order.CashFlowStatement, error) {
	client, err := a.createClient(credentials, platformAuthID)
	if err != nil {
		return nil, err
	}

	financeAPI := ozon.NewFinanceAPI(client)
	cashFlows, err := financeAPI.GetAllCashFlowStatements(since, to)
	if err != nil {
		return nil, err
	}

	var result []order.CashFlowStatement
	for _, cf := range cashFlows {
		statement := a.convertToCashFlowStatement(&cf)
		result = append(result, *statement)
	}

	return result, nil
}

// convertToCashFlowStatement 转换为统一现金流报表格式
func (a *OzonAdapter) convertToCashFlowStatement(cf *ozon.CashFlowItem) *order.CashFlowStatement {
	statement := &order.CashFlowStatement{
		CurrencyCode:                cf.CurrencyCode,
		OrdersAmount:                cf.OrdersAmount,
		ReturnsAmount:               cf.ReturnsAmount,
		CommissionAmount:            cf.CommissionAmount,
		ServicesAmount:              cf.ServicesAmount,
		ItemDeliveryAndReturnAmount: cf.ItemDeliveryAndReturnAmount,
	}

	// 解析周期时间
	if cf.Period.Begin != "" {
		if t, err := time.Parse(time.RFC3339, cf.Period.Begin); err == nil {
			statement.PeriodBegin = &t
		}
	}
	if cf.Period.End != "" {
		if t, err := time.Parse(time.RFC3339, cf.Period.End); err == nil {
			statement.PeriodEnd = &t
		}
	}

	return statement
}

// ========== 辅助方法 ==========

// createClient 创建OZON API客户端
func (a *OzonAdapter) createClient(credentials string, platformAuthID uint) (*ozon.Client, error) {
	creds, err := ozon.ParseCredentials(credentials)
	if err != nil {
		return nil, err
	}

	var db *gorm.DB
	if platformAuthID > 0 {
		db = database.GetDB()
	}

	logger := platform.NewRequestLogger(db, platformAuthID, order.PlatformOzon)
	return ozon.NewClient(creds, logger), nil
}

// convertToOrder 转换为统一订单格式
func (a *OzonAdapter) convertToOrder(posting *ozon.Posting) *order.Order {
	ord := &order.Order{
		Platform:        order.PlatformOzon,
		PlatformOrderNo: posting.PostingNumber,
		PlatformStatus:  posting.Status,
		Status:          a.mapStatus(posting.Status),
	}

	// 序列化原始数据
	// rawData, _ := json.Marshal(posting)
	// ord.RawData = string(rawData)

	// 解析订单时间
	if posting.InProcessAt != "" {
		if t, err := time.Parse(time.RFC3339, posting.InProcessAt); err == nil {
			ord.OrderTime = &t
		}
	}

	// 解析发货时间
	if posting.ShipmentDate != "" {
		if t, err := time.Parse(time.RFC3339, posting.ShipmentDate); err == nil {
			ord.ShipTime = &t
		}
	}

	// 收件人信息
	if posting.AddressInfo != nil {
		ord.RecipientName = posting.AddressInfo.Name
		ord.RecipientPhone = posting.AddressInfo.Phone
	}

	// 地址信息
	if posting.Customer != nil {
		ord.Country = posting.Customer.Address.Country
		ord.Province = posting.Customer.Address.Region
		ord.City = posting.Customer.Address.City
		ord.ZipCode = posting.Customer.Address.ZipCode
		ord.Address = posting.Customer.Address.AddressTail
	}

	// 商品信息
	var totalAmount float64
	for _, prod := range posting.Products {
		price, _ := strconv.ParseFloat(prod.Price, 64)
		item := order.OrderItem{
			PlatformSku: strconv.FormatInt(prod.Sku, 10),
			Sku:         prod.OfferID,
			Name:        prod.Name,
			Quantity:    prod.Quantity,
			Price:       price,
			Currency:    prod.CurrencyCode,
		}
		totalAmount += price * float64(prod.Quantity)
		ord.Items = append(ord.Items, item)
		if ord.Currency == "" && prod.CurrencyCode != "" {
			ord.Currency = prod.CurrencyCode
		}
	}
	ord.TotalAmount = totalAmount

	return ord
}

// mapStatus 映射状态（使用统一的状态映射机制）
func (a *OzonAdapter) mapStatus(platformStatus string) string {
	return order.MapPlatformStatus(order.PlatformOzon, platformStatus)
}

func init() {
	// 注册 Ozon 平台状态映射
	// Ozon 订单状态映射
	// 文档参考: https://docs.ozon.ru/api/seller/#operation/PostingAPI_GetFbsPostingListV3
	order.RegisterPlatformStatusMapping(order.PlatformOzon, map[string]string{
		"awaiting_packaging":  order.OrderStatusPending,     // 等待包装
		"awaiting_deliver":    order.OrderStatusReadyToShip, // 等待交付
		"ready_to_ship":       order.OrderStatusReadyToShip, // 准备发货
		"arbitration":         order.OrderStatusPending,     // 仲裁中
		"client_arbitration":  order.OrderStatusPending,     // 客户仲裁
		"delivering":          order.OrderStatusShipped,     // 配送中
		"driver_pickup":       order.OrderStatusShipped,     // 司机取货
		"delivered":           order.OrderStatusDelivered,   // 已送达
		"cancelled":           order.OrderStatusCancelled,   // 已取消
		"not_accepted":        order.OrderStatusCancelled,   // 未接受
		"sent_by_seller":      order.OrderStatusShipped,     // 卖家已发货
	})

	// 自动注册适配器
	order.RegisterAdapter(&OzonAdapter{})
}
