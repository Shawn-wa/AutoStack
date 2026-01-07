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

// 订单状态常量和映射已迁移至 status.go
// 使用 OrderStatusPending, OrderStatusReadyToShip 等常量
// 平台状态映射使用 RegisterPlatformStatusMapping 和 MapPlatformStatus 函数

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
	// ========== 佣金相关字段 ==========
	// Ozon API: POST /v3/finance/transaction/totals
	// 文档: https://docs.ozon.ru/api/seller/#operation/FinanceAPI_FinanceTransactionTotalV3
	AccrualsForSale         float64     `gorm:"type:decimal(10,2);default:0;comment:accruals_for_sale-销售收入，卖家因销售商品获得的收入金额（正数）" json:"accruals_for_sale"`
	SaleCommission          float64     `gorm:"type:decimal(10,2);default:0;comment:sale_commission-销售佣金，Ozon平台从销售中收取的佣金费用（负数）" json:"sale_commission"`
	ProcessingAndDelivery   float64     `gorm:"type:decimal(10,2);default:0;comment:processing_and_delivery-物流费用，商品处理、包装和配送的费用（负数）" json:"processing_and_delivery"`
	RefundsAndCancellations float64     `gorm:"type:decimal(10,2);default:0;comment:refunds_and_cancellations-退款退货，退款及取消订单产生的费用扣减" json:"refunds_and_cancellations"`
	ServicesAmount          float64     `gorm:"type:decimal(10,2);default:0;comment:services_amount-平台服务费，Ozon提供的增值服务费用（负数）" json:"services_amount"`
	CompensationAmount      float64     `gorm:"type:decimal(10,2);default:0;comment:compensation_amount-补偿金额，平台对卖家的补偿款项（正数）" json:"compensation_amount"`
	MoneyTransfer           float64     `gorm:"type:decimal(10,2);default:0;comment:money_transfer-资金划转，账户间资金转移记录" json:"money_transfer"`
	OthersAmount            float64     `gorm:"type:decimal(10,2);default:0;comment:others_amount-其他费用，未归类的其他杂项费用" json:"others_amount"`
	ProfitAmount            float64     `gorm:"type:decimal(10,2);default:0;comment:profit_amount-订单利润，所有费用项汇总后的最终利润金额" json:"profit_amount"`
	CommissionCurrency      string      `gorm:"size:10;comment:commission_currency-结算货币，佣金费用结算使用的货币代码" json:"commission_currency"`
	CommissionSyncedAt      *time.Time  `gorm:"comment:commission_synced_at-佣金同步时间，最后一次从平台同步佣金数据的时间" json:"commission_synced_at"`
	RawData                 string      `gorm:"type:longtext" json:"-"`
	CreatedAt               time.Time   `json:"created_at"`
	UpdatedAt               time.Time   `json:"updated_at"`
	Items                   []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
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
	RequestTypeCashFlow    = "CashFlow"    // 现金流报表
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

// CashFlowStatement 现金流报表模型
// Ozon API: POST /v1/finance/cash-flow-statement/list
// 文档: https://docs.ozon.ru/api/seller/#operation/FinanceAPI_FinanceCashFlowStatementList
type CashFlowStatement struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	UserID         uint       `gorm:"index;not null" json:"user_id"`
	PlatformAuthID uint       `gorm:"index;not null" json:"platform_auth_id"`
	Platform       string     `gorm:"size:50;not null;index" json:"platform"`
	PeriodBegin    *time.Time `gorm:"uniqueIndex:idx_auth_period;comment:period.begin-报告周期开始时间" json:"period_begin"`
	PeriodEnd      *time.Time `gorm:"comment:period.end-报告周期结束时间" json:"period_end"`
	CurrencyCode   string     `gorm:"size:10;comment:currency_code-货币代码，如RUB、USD等" json:"currency_code"`
	// ========== 金额字段 ==========
	OrdersAmount                float64   `gorm:"type:decimal(12,2);default:0;comment:orders_amount-订单销售金额，该周期内订单的总销售额（正数）" json:"orders_amount"`
	ReturnsAmount               float64   `gorm:"type:decimal(12,2);default:0;comment:returns_amount-退货金额，该周期内退货产生的金额（负数）" json:"returns_amount"`
	CommissionAmount            float64   `gorm:"type:decimal(12,2);default:0;comment:commission_amount-平台佣金，Ozon收取的销售佣金总额（负数）" json:"commission_amount"`
	ServicesAmount              float64   `gorm:"type:decimal(12,2);default:0;comment:services_amount-服务费用，平台提供的各类服务费用（负数）" json:"services_amount"`
	ItemDeliveryAndReturnAmount float64   `gorm:"type:decimal(12,2);default:0;comment:item_delivery_and_return_amount-物流费用，商品配送和退货物流费用（负数）" json:"item_delivery_and_return_amount"`
	SyncedAt                    time.Time `gorm:"comment:synced_at-数据同步时间，从平台拉取此报表的时间" json:"synced_at"`
	CreatedAt                   time.Time `json:"created_at"`
	UpdatedAt                   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (CashFlowStatement) TableName() string {
	return "cash_flow_statements"
}

// OrderDailyStat 订单每日统计模型
type OrderDailyStat struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"uniqueIndex:idx_user_date_currency;not null" json:"user_id"`
	StatDate    time.Time `gorm:"uniqueIndex:idx_user_date_currency;type:date;not null" json:"stat_date"`
	Currency    string    `gorm:"size:10;not null;default:'RUB';uniqueIndex:idx_user_date_currency" json:"currency"`
	OrderCount  int64     `gorm:"default:0;comment:当日订单数量" json:"order_count"`
	OrderAmount float64   `gorm:"type:decimal(15,2);default:0;comment:当日订单金额" json:"order_amount"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (OrderDailyStat) TableName() string {
	return "order_daily_stats"
}
