package order

import (
	"time"
)

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
	ShipDeadline    *time.Time `json:"ship_deadline"` // 发货截止时间
	// ========== 佣金相关字段 ==========
	AccrualsForSale         float64    `gorm:"type:decimal(10,2);default:0" json:"accruals_for_sale"`
	SaleCommission          float64    `gorm:"type:decimal(10,2);default:0" json:"sale_commission"`
	ProcessingAndDelivery   float64    `gorm:"type:decimal(10,2);default:0" json:"processing_and_delivery"`
	RefundsAndCancellations float64    `gorm:"type:decimal(10,2);default:0" json:"refunds_and_cancellations"`
	ServicesAmount          float64    `gorm:"type:decimal(10,2);default:0" json:"services_amount"`
	CompensationAmount      float64    `gorm:"type:decimal(10,2);default:0" json:"compensation_amount"`
	MoneyTransfer           float64    `gorm:"type:decimal(10,2);default:0" json:"money_transfer"`
	OthersAmount            float64    `gorm:"type:decimal(10,2);default:0" json:"others_amount"`
	ProfitAmount            float64    `gorm:"type:decimal(10,2);default:0" json:"profit_amount"`
	CommissionCurrency      string     `gorm:"size:10" json:"commission_currency"`
	CommissionSyncedAt      *time.Time `json:"commission_synced_at"`
	RawData                 string     `gorm:"type:longtext" json:"-"`
	// ========== 物流费用估算字段 ==========
	EstimatedShippingFee      float64     `gorm:"type:decimal(10,2);default:0" json:"estimated_shipping_fee"` // 估算物流费用
	EstimatedShippingCurrency string      `gorm:"size:10" json:"estimated_shipping_currency"`                 // 物流费用货币
	ShippingTemplateID        uint        `gorm:"default:0" json:"shipping_template_id"`                      // 使用的运费模版ID
	ShippingEstimatedAt       *time.Time  `json:"shipping_estimated_at"`                                      // 物流费用估算时间
	CreatedAt                 time.Time   `json:"created_at"`
	UpdatedAt                 time.Time   `json:"updated_at"`
	Items                     []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
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
	// 物流费用估算
	EstimatedShippingFee      float64 `gorm:"type:decimal(10,2);default:0" json:"estimated_shipping_fee"` // 单品估算物流费用
	EstimatedShippingCurrency string  `gorm:"size:10" json:"estimated_shipping_currency"`                 // 物流费用货币
}

// TableName 指定表名
func (OrderItem) TableName() string {
	return "order_items"
}

// OrderDailyStat 订单每日统计模型
type OrderDailyStat struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"uniqueIndex:idx_user_date_currency;not null" json:"user_id"`
	StatDate    time.Time `gorm:"uniqueIndex:idx_user_date_currency;type:date;not null" json:"stat_date"`
	Currency    string    `gorm:"size:10;not null;default:'RUB';uniqueIndex:idx_user_date_currency" json:"currency"`
	OrderCount  int64     `gorm:"default:0" json:"order_count"`
	OrderAmount float64   `gorm:"type:decimal(15,2);default:0" json:"order_amount"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (OrderDailyStat) TableName() string {
	return "order_daily_stats"
}

// CommissionData 佣金数据
type CommissionData struct {
	AccrualsForSale         float64
	SaleCommission          float64
	ProcessingAndDelivery   float64
	RefundsAndCancellations float64
	ServicesAmount          float64
	CompensationAmount      float64
	MoneyTransfer           float64
	OthersAmount            float64
	ProfitAmount            float64
	CommissionCurrency      string
}
