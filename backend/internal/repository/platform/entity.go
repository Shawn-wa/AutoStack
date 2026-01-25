package platform

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

// 请求日志接口类型常量
const (
	RequestTypeOrderList   = "OrderList"   // 订单列表
	RequestTypeFinance     = "Finance"     // 财务/佣金
	RequestTypeTestConnect = "TestConnect" // 测试连接
	RequestTypeCashFlow    = "CashFlow"    // 现金流报表
)

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

// OrdersRequestLog 订单请求日志模型
type OrdersRequestLog struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	PlatformAuthID uint      `gorm:"index;not null" json:"platform_auth_id"`
	Platform       string    `gorm:"size:50;not null;index" json:"platform"`
	RequestType    string    `gorm:"size:50;not null;index" json:"request_type"`
	RequestURL     string    `gorm:"size:500;not null" json:"request_url"`
	RequestMethod  string    `gorm:"size:10;not null" json:"request_method"`
	RequestHeaders string    `gorm:"type:text" json:"-"`
	RequestBody    string    `gorm:"type:longtext" json:"request_body"`
	ResponseStatus int       `gorm:"not null" json:"response_status"`
	ResponseBody   string    `gorm:"type:longtext" json:"response_body"`
	Duration       int64     `gorm:"not null" json:"duration"`
	ErrorMessage   string    `gorm:"size:1000" json:"error_message"`
	CreatedAt      time.Time `gorm:"index" json:"created_at"`
}

// TableName 指定表名
func (OrdersRequestLog) TableName() string {
	return "orders_request_log"
}

// CashFlowStatement 现金流报表模型
type CashFlowStatement struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	UserID         uint       `gorm:"index;not null" json:"user_id"`
	PlatformAuthID uint       `gorm:"index;not null" json:"platform_auth_id"`
	Platform       string     `gorm:"size:50;not null;index" json:"platform"`
	PeriodBegin    *time.Time `gorm:"uniqueIndex:idx_auth_period" json:"period_begin"`
	PeriodEnd      *time.Time `json:"period_end"`
	CurrencyCode   string     `gorm:"size:10" json:"currency_code"`
	// ========== 金额字段 ==========
	OrdersAmount                float64   `gorm:"type:decimal(12,2);default:0" json:"orders_amount"`
	ReturnsAmount               float64   `gorm:"type:decimal(12,2);default:0" json:"returns_amount"`
	CommissionAmount            float64   `gorm:"type:decimal(12,2);default:0" json:"commission_amount"`
	ServicesAmount              float64   `gorm:"type:decimal(12,2);default:0" json:"services_amount"`
	ItemDeliveryAndReturnAmount float64   `gorm:"type:decimal(12,2);default:0" json:"item_delivery_and_return_amount"`
	SyncedAt                    time.Time `json:"synced_at"`
	CreatedAt                   time.Time `json:"created_at"`
	UpdatedAt                   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (CashFlowStatement) TableName() string {
	return "cash_flow_statements"
}
