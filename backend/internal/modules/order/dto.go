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
	ID         uint       `json:"id"`
	Platform   string     `json:"platform"`
	ShopName   string     `json:"shop_name"`
	Status     int        `json:"status"`
	LastSyncAt *time.Time `json:"last_sync_at"`
	CreatedAt  string     `json:"created_at"`
	UpdatedAt  string     `json:"updated_at"`
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
	SaleCommission       float64             `json:"sale_commission"`
	AccrualsForSale      float64             `json:"accruals_for_sale"`
	DeliveryCharge       float64             `json:"delivery_charge"`
	ReturnDeliveryCharge float64             `json:"return_delivery_charge"`
	CommissionAmount     float64             `json:"commission_amount"`
	CommissionCurrency   string              `json:"commission_currency"`
	CommissionSyncedAt   *time.Time          `json:"commission_synced_at"`
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
