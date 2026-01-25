package order

import "time"

// OrderQuery 订单查询条件
type OrderQuery struct {
	Page           int
	PageSize       int
	UserID         uint
	Platform       string
	AuthID         uint
	Status         string   // 支持逗号分隔多状态
	Statuses       []string // 多状态列表
	Keyword        string
	StartTime      string
	EndTime        string
	OrderBy        string
}

// OrderItemQuery 订单商品查询条件
type OrderItemQuery struct {
	OrderID uint
}

// DailyStatQuery 每日统计查询条件
type DailyStatQuery struct {
	UserID    uint
	StartDate time.Time
	EndDate   time.Time
	Currency  string
}

// CurrencyAmount 币种金额
type CurrencyAmount struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
}

// CommissionSummary 佣金汇总
type CommissionSummary struct {
	TotalProfit     float64
	TotalCommission float64
	TotalServiceFee float64
}

// DailyStat 每日统计数据
type DailyStat struct {
	Date     time.Time
	Currency string
	Count    int64
	Amount   float64
}
