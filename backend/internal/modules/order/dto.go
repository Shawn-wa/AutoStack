package order

import "time"

// CredentialField 凭证字段定义
type CredentialField struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Type     string `json:"type"` // text, password
	Required bool   `json:"required"`
}

// PlatformInfo 平台信息
type PlatformInfo struct {
	Name   string            `json:"name"`
	Label  string            `json:"label"`
	Fields []CredentialField `json:"fields"`
}

// CreateAuthRequest 创建授权请求
type CreateAuthRequest struct {
	Platform    string            `json:"platform" binding:"required"`
	ShopName    string            `json:"shop_name" binding:"required"`
	Credentials map[string]string `json:"credentials" binding:"required"`
}

// UpdateAuthRequest 更新授权请求
type UpdateAuthRequest struct {
	ShopName    string            `json:"shop_name"`
	Credentials map[string]string `json:"credentials"`
	Status      *int              `json:"status"`
}

// AuthResponse 授权响应
type AuthResponse struct {
	ID                uint              `json:"id"`
	Platform          string            `json:"platform"`
	ShopName          string            `json:"shop_name"`
	Status            int               `json:"status"`
	MaskedCredentials map[string]string `json:"masked_credentials,omitempty"`
	LastSyncAt        *time.Time        `json:"last_sync_at"`
	CreatedAt         string            `json:"created_at"`
	UpdatedAt         string            `json:"updated_at"`
}

// AuthListResponse 授权列表响应
type AuthListResponse struct {
	List     []AuthResponse `json:"list"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

// SyncOrdersRequest 同步订单请求
type SyncOrdersRequest struct {
	Since string `json:"since"` // 开始时间 ISO8601
	To    string `json:"to"`    // 结束时间 ISO8601
}

// SyncOrdersResponse 同步订单响应
type SyncOrdersResponse struct {
	Total   int `json:"total"`
	Created int `json:"created"`
	Updated int `json:"updated"`
}

// OrderListRequest 订单列表请求
type OrderListRequest struct {
	Page      int    `form:"page" binding:"omitempty,min=1"`
	PageSize  int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Platform  string `form:"platform"`
	AuthID    uint   `form:"auth_id"`
	Status    string `form:"status"`
	Keyword   string `form:"keyword"`
	StartTime string `form:"start_time"`
	EndTime   string `form:"end_time"`
}

// OrderResponse 订单响应
type OrderResponse struct {
	ID              uint       `json:"id"`
	Platform        string     `json:"platform"`
	PlatformOrderNo string     `json:"platform_order_no"`
	Status          string     `json:"status"`
	PlatformStatus  string     `json:"platform_status"`
	TotalAmount     float64    `json:"total_amount"`
	Currency        string     `json:"currency"`
	RecipientName   string     `json:"recipient_name"`
	RecipientPhone  string     `json:"recipient_phone"`
	Country         string     `json:"country"`
	Province        string     `json:"province"`
	City            string     `json:"city"`
	ZipCode         string     `json:"zip_code"`
	Address         string     `json:"address"`
	OrderTime       *time.Time `json:"order_time"`
	ShipTime        *time.Time `json:"ship_time"`
	// 佣金信息
	AccrualsForSale         float64             `json:"accruals_for_sale"`
	SaleCommission          float64             `json:"sale_commission"`
	ProcessingAndDelivery   float64             `json:"processing_and_delivery"`
	RefundsAndCancellations float64             `json:"refunds_and_cancellations"`
	ServicesAmount          float64             `json:"services_amount"`
	CompensationAmount      float64             `json:"compensation_amount"`
	MoneyTransfer           float64             `json:"money_transfer"`
	OthersAmount            float64             `json:"others_amount"`
	ProfitAmount            float64             `json:"profit_amount"`
	CommissionCurrency      string              `json:"commission_currency"`
	CommissionSyncedAt      *time.Time          `json:"commission_synced_at"`
	Items                []OrderItemResponse `json:"items,omitempty"`
	CreatedAt            string              `json:"created_at"`
	UpdatedAt            string              `json:"updated_at"`
}

// OrderItemResponse 订单商品响应
type OrderItemResponse struct {
	ID          uint    `json:"id"`
	PlatformSku string  `json:"platform_sku"`
	Sku         string  `json:"sku"`
	Name        string  `json:"name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
	Currency    string  `json:"currency"`
}

// OrderListResponse 订单列表响应
type OrderListResponse struct {
	List     []OrderResponse `json:"list"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

// SyncCommissionRequest 佣金同步请求
type SyncCommissionRequest struct {
	Since string `json:"since"` // 开始时间 ISO8601
	To    string `json:"to"`    // 结束时间 ISO8601
}

// SyncCommissionResponse 佣金同步响应
type SyncCommissionResponse struct {
	Total   int `json:"total"`   // 处理的交易数
	Updated int `json:"updated"` // 更新的订单数
}

// ========== 现金流报表相关 ==========

// SyncCashFlowRequest 现金流同步请求
type SyncCashFlowRequest struct {
	Since string `json:"since"` // 开始时间 ISO8601
	To    string `json:"to"`    // 结束时间 ISO8601
}

// CashFlowResponse 现金流报表响应
type CashFlowResponse struct {
	ID                          uint       `json:"id"`
	PlatformAuthID              uint       `json:"platform_auth_id"`
	Platform                    string     `json:"platform"`
	PeriodBegin                 *time.Time `json:"period_begin"`
	PeriodEnd                   *time.Time `json:"period_end"`
	CurrencyCode                string     `json:"currency_code"`
	OrdersAmount                float64    `json:"orders_amount"`
	ReturnsAmount               float64    `json:"returns_amount"`
	CommissionAmount            float64    `json:"commission_amount"`
	ServicesAmount              float64    `json:"services_amount"`
	ItemDeliveryAndReturnAmount float64    `json:"item_delivery_and_return_amount"`
	SyncedAt                    string     `json:"synced_at"`
}

// CashFlowListResponse 现金流报表列表响应
type CashFlowListResponse struct {
	List     []CashFlowResponse `json:"list"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
}

// ========== 统计相关 ==========

// DashboardStatsResponse 仪表盘统计响应
type DashboardStatsResponse struct {
	// 订单统计
	TotalOrders     int64            `json:"total_orders"`     // 总订单数
	DeliveredOrders int64            `json:"delivered_orders"` // 已签收订单数
	PendingOrders   int64            `json:"pending_orders"`   // 待处理订单数
	TodayOrders     int64            `json:"today_orders"`     // 今日订单数
	TotalAmounts    []CurrencyAmount `json:"total_amounts"`    // 订单总金额（多币种）
	// 佣金统计
	TotalProfit     float64 `json:"total_profit"`      // 总利润
	TotalCommission float64 `json:"total_commission"`  // 总佣金
	TotalServiceFee float64 `json:"total_service_fee"` // 总服务费
	// 授权统计
	TotalAuths  int64 `json:"total_auths"`  // 总授权数
	ActiveAuths int64 `json:"active_auths"` // 活跃授权数
	// 货币
	Currency string `json:"currency"`
}

// CurrencyAmount 币种金额
type CurrencyAmount struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
}

// RecentOrderResponse 最近订单响应
type RecentOrderResponse struct {
	ID              uint       `json:"id"`
	PlatformOrderNo string     `json:"platform_order_no"`
	Status          string     `json:"status"`
	TotalAmount     float64    `json:"total_amount"`
	Currency        string     `json:"currency"`
	OrderTime       *time.Time `json:"order_time"`
}

// OrderTrendItem 订单趋势数据项
type OrderTrendItem struct {
	Date   string `json:"date"`   // 日期 YYYY-MM-DD
	Count  int64  `json:"count"`  // 订单数量
	Amount float64 `json:"amount"` // 订单金额
}

// OrderTrendResponse 订单趋势响应
type OrderTrendResponse struct {
	Items []OrderTrendItem `json:"items"`
}

// ========== 订单汇总相关 ==========

// OrderSummaryRequest 订单汇总请求
type OrderSummaryRequest struct {
	StartTime string `form:"start_time"`
	EndTime   string `form:"end_time"`
	AuthID    uint   `form:"auth_id"`
	Platform  string `form:"platform"`
	Keyword   string `form:"keyword"` // 搜索关键词（本地SKU/标题/平台SKU）
	Status    string `form:"status"`  // 订单状态筛选
}

// OrderSummaryStatusDetail 状态明细
type OrderSummaryStatusDetail struct {
	Status   string  `json:"status"`
	Quantity int     `json:"quantity"`
	Amount   float64 `json:"amount"`
}

// OrderSummaryPlatformSKU 平台SKU详情
type OrderSummaryPlatformSKU struct {
	SKU   string `json:"sku"`
	Name  string `json:"name"`
	Image string `json:"image"`
}

// OrderSummaryItem 订单汇总项（按本地SKU合并）
type OrderSummaryItem struct {
	LocalSKU         string                     `json:"local_sku"`
	ProductName      string                     `json:"product_name"`
	PlatformSKUs     []string                   `json:"platform_skus"`      // 保持兼容
	PlatformProducts []OrderSummaryPlatformSKU  `json:"platform_products"`  // 平台产品详情
	Quantity         int                        `json:"quantity"`
	Amount           float64                    `json:"amount"`
	Currency         string                     `json:"currency"`
	StatusDetails    []OrderSummaryStatusDetail `json:"status_details"`
	AvailableStock   int                        `json:"available_stock"`    // 系统可用库存
}
