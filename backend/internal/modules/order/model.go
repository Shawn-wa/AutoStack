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
	ID              uint        `gorm:"primaryKey" json:"id"`
	UserID          uint        `gorm:"index;not null" json:"user_id"`
	PlatformAuthID  uint        `gorm:"index;not null" json:"platform_auth_id"`
	Platform        string      `gorm:"size:50;not null;index" json:"platform"`
	PlatformOrderNo string      `gorm:"size:100;not null;uniqueIndex" json:"platform_order_no"`
	Status          string      `gorm:"size:50;not null;index" json:"status"`
	PlatformStatus  string      `gorm:"size:50" json:"platform_status"`
	TotalAmount     float64     `gorm:"type:decimal(10,2)" json:"total_amount"`
	Currency        string      `gorm:"size:10" json:"currency"`
	RecipientName   string      `gorm:"size:100" json:"recipient_name"`
	RecipientPhone  string      `gorm:"size:50" json:"recipient_phone"`
	Country         string      `gorm:"size:100" json:"country"`
	Province        string      `gorm:"size:100" json:"province"`
	City            string      `gorm:"size:100" json:"city"`
	ZipCode         string      `gorm:"size:20" json:"zip_code"`
	Address         string      `gorm:"size:500" json:"address"`
	OrderTime       *time.Time  `json:"order_time"`
	ShipTime        *time.Time  `json:"ship_time"`
	RawData         string      `gorm:"type:text" json:"-"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
	Items           []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
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
