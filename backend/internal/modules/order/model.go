package order

import (
	"time"
)

// 平台常量
const (
	PlatformOzon = "ozon"
)

// 授权状态常量
const (
	AuthStatusDisabled = 0 // 禁用
	AuthStatusActive   = 1 // 正常
	AuthStatusExpired  = 2 // 授权失效
)

// 统一订单状态常量
const (
	OrderStatusPending     = "pending"       // 待处理
	OrderStatusReadyToShip = "ready_to_ship" // 待发货
	OrderStatusShipped     = "shipped"       // 已发货
	OrderStatusDelivered   = "delivered"     // 已签收
	OrderStatusCancelled   = "cancelled"     // 已取消
)

// Ozon 状态映射
var OzonStatusMap = map[string]string{
	"awaiting_packaging": OrderStatusPending,
	"awaiting_deliver":   OrderStatusReadyToShip,
	"delivering":         OrderStatusShipped,
	"delivered":          OrderStatusDelivered,
	"cancelled":          OrderStatusCancelled,
}

// PlatformAuth 平台授权模型
type PlatformAuth struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	UserID      uint       `gorm:"index;not null" json:"user_id"`
	Platform    string     `gorm:"size:50;not null" json:"platform"`
	ShopName    string     `gorm:"size:100;not null" json:"shop_name"`
	Credentials string     `gorm:"type:text;not null" json:"-"`
	Status      int        `gorm:"default:1" json:"status"`
	LastSyncAt  *time.Time `json:"last_sync_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// TableName 指定表名
func (PlatformAuth) TableName() string {
	return "platform_auths"
}

// Order 订单模型
type Order struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	UserID          uint       `gorm:"index;not null" json:"user_id"`
	PlatformAuthID  uint       `gorm:"index;not null" json:"platform_auth_id"`
	Platform        string     `gorm:"size:50;not null;index" json:"platform"`
	PlatformOrderNo string     `gorm:"size:100;not null;uniqueIndex" json:"platform_order_no"`
	Status          string     `gorm:"size:50;not null;index" json:"status"`
	PlatformStatus  string     `gorm:"size:50" json:"platform_status"`
	TotalAmount     float64    `gorm:"type:decimal(10,2)" json:"total_amount"`
	Currency        string     `gorm:"size:10" json:"currency"`
	RecipientName   string     `gorm:"size:100" json:"recipient_name"`
	RecipientPhone  string     `gorm:"size:50" json:"recipient_phone"`
	Country         string     `gorm:"size:100" json:"country"`
	Province        string     `gorm:"size:100" json:"province"`
	City            string     `gorm:"size:100" json:"city"`
	ZipCode         string     `gorm:"size:20" json:"zip_code"`
	Address         string     `gorm:"size:500" json:"address"`
	OrderTime       *time.Time `json:"order_time"`
	ShipTime        *time.Time `json:"ship_time"`
	// 佣金相关字段
	SaleCommission       float64     `gorm:"type:decimal(10,2);default:0" json:"sale_commission"`
	AccrualsForSale      float64     `gorm:"type:decimal(10,2);default:0" json:"accruals_for_sale"`
	DeliveryCharge       float64     `gorm:"type:decimal(10,2);default:0" json:"delivery_charge"`
	ReturnDeliveryCharge float64     `gorm:"type:decimal(10,2);default:0" json:"return_delivery_charge"`
	CommissionAmount     float64     `gorm:"type:decimal(10,2);default:0" json:"commission_amount"`
	CommissionCurrency   string      `gorm:"size:10" json:"commission_currency"`
	CommissionSyncedAt   *time.Time  `json:"commission_synced_at"`
	RawData              string      `gorm:"type:longtext" json:"-"`
	CreatedAt            time.Time   `json:"created_at"`
	UpdatedAt            time.Time   `json:"updated_at"`
	Items                []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
}

// TableName 指定表名
func (Order) TableName() string {
	return "orders"
}

// OrderItem 订单商品模型
type OrderItem struct {
	ID          uint    `gorm:"primaryKey" json:"id"`
	OrderID     uint    `gorm:"index;not null" json:"order_id"`
	PlatformSku string  `gorm:"size:100" json:"platform_sku"`
	Sku         string  `gorm:"size:100" json:"sku"`
	Name        string  `gorm:"size:255" json:"name"`
	Quantity    int     `gorm:"not null" json:"quantity"`
	Price       float64 `gorm:"type:decimal(10,2)" json:"price"`
	Currency    string  `gorm:"size:10" json:"currency"`
}

// TableName 指定表名
func (OrderItem) TableName() string {
	return "order_items"
}

// CommissionData 佣金数据
type CommissionData struct {
	SaleCommission       float64
	AccrualsForSale      float64
	DeliveryCharge       float64
	ReturnDeliveryCharge float64
	CommissionAmount     float64
	CommissionCurrency   string // 佣金货币（结算货币）
}

// 请求日志接口类型常量
const (
	RequestTypeOrderList   = "OrderList"   // 订单列表
	RequestTypeFinance     = "Finance"     // 财务/佣金
	RequestTypeTestConnect = "TestConnect" // 测试连接
)

// OrdersRequestLog 订单请求日志模型
type OrdersRequestLog struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	PlatformAuthID uint      `gorm:"index;not null" json:"platform_auth_id"`
	Platform       string    `gorm:"size:50;not null;index" json:"platform"`
	RequestType    string    `gorm:"size:50;not null;index" json:"request_type"` // OrderList, Finance, TestConnect
	RequestURL     string    `gorm:"size:500;not null" json:"request_url"`
	RequestMethod  string    `gorm:"size:10;not null" json:"request_method"` // GET, POST
	RequestHeaders string    `gorm:"type:text" json:"-"`                     // 请求头（脱敏后）
	RequestBody    string    `gorm:"type:longtext" json:"request_body"`      // 请求入参
	ResponseStatus int       `gorm:"not null" json:"response_status"`        // HTTP状态码
	ResponseBody   string    `gorm:"type:longtext" json:"response_body"`     // 响应出参
	Duration       int64     `gorm:"not null" json:"duration"`               // 请求耗时(毫秒)
	ErrorMessage   string    `gorm:"size:1000" json:"error_message"`         // 错误信息
	CreatedAt      time.Time `gorm:"index" json:"created_at"`
}

// TableName 指定表名
func (OrdersRequestLog) TableName() string {
	return "orders_request_log"
}
