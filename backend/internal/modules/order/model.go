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
	// 佣金相关字段（来自 Ozon finance/transaction/totals API）
	// 参考文档: https://docs.ozon.ru/api/seller/#operation/FinanceAPI_FinanceTransactionTotalV3
	AccrualsForSale         float64    `gorm:"type:decimal(10,2);default:0;comment:销售收入-卖家因销售商品获得的收入金额" json:"accruals_for_sale"`
	SaleCommission          float64    `gorm:"type:decimal(10,2);default:0;comment:销售佣金-平台从销售中收取的佣金费用" json:"sale_commission"`
	ProcessingAndDelivery   float64    `gorm:"type:decimal(10,2);default:0;comment:加工和配送-商品处理和配送的费用" json:"processing_and_delivery"`
	RefundsAndCancellations float64    `gorm:"type:decimal(10,2);default:0;comment:退款和取消-退款及取消订单相关费用" json:"refunds_and_cancellations"`
	ServicesAmount          float64    `gorm:"type:decimal(10,2);default:0;comment:服务费-平台服务费用" json:"services_amount"`
	CompensationAmount      float64    `gorm:"type:decimal(10,2);default:0;comment:补偿金额-平台补偿给卖家的金额" json:"compensation_amount"`
	MoneyTransfer           float64    `gorm:"type:decimal(10,2);default:0;comment:资金转账-资金转账相关" json:"money_transfer"`
	OthersAmount            float64    `gorm:"type:decimal(10,2);default:0;comment:其他金额-其他杂项费用" json:"others_amount"`
	ProfitAmount            float64    `gorm:"type:decimal(10,2);default:0;comment:订单利润额-所有费用项汇总后的最终利润" json:"profit_amount"`
	CommissionCurrency      string     `gorm:"size:10;comment:佣金货币-费用结算使用的货币类型" json:"commission_currency"`
	CommissionSyncedAt      *time.Time `gorm:"comment:佣金同步时间-最后一次从平台同步佣金数据的时间" json:"commission_synced_at"`
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
	AccrualsForSale         float64 // 销售收入
	SaleCommission          float64 // 销售佣金
	ProcessingAndDelivery   float64 // 加工和配送费
	RefundsAndCancellations float64 // 退款和取消
	ServicesAmount          float64 // 服务费
	CompensationAmount      float64 // 补偿金额
	MoneyTransfer           float64 // 资金转账
	OthersAmount            float64 // 其他金额
	ProfitAmount            float64 // 订单利润额
	CommissionCurrency      string  // 佣金货币（结算货币）
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
