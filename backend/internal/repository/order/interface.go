package order

import (
	"context"
	"time"
)

// OrderRepository 订单仓储接口
type OrderRepository interface {
	// FindByID 根据ID查找订单
	FindByID(ctx context.Context, id uint) (*Order, error)
	// FindByIDAndUserID 根据ID和用户ID查找订单
	FindByIDAndUserID(ctx context.Context, id, userID uint) (*Order, error)
	// FindByPlatformOrderNo 根据平台订单号查找订单
	FindByPlatformOrderNo(ctx context.Context, orderNo string) (*Order, error)
	// List 分页查询订单列表
	List(ctx context.Context, query *OrderQuery) ([]Order, int64, error)
	// Create 创建订单
	Create(ctx context.Context, order *Order) error
	// Update 更新订单
	Update(ctx context.Context, order *Order) error
	// UpdateFields 更新指定字段
	UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error
	// UpdateByPlatformOrderNo 根据平台订单号更新字段
	UpdateByPlatformOrderNo(ctx context.Context, orderNo string, authID uint, fields map[string]interface{}) (int64, error)
	// CountByUserID 统计用户订单总数
	CountByUserID(ctx context.Context, userID uint) (int64, error)
	// CountByStatus 统计指定状态的订单数
	CountByStatus(ctx context.Context, userID uint, status string) (int64, error)
	// CountByStatuses 统计多个状态的订单数
	CountByStatuses(ctx context.Context, userID uint, statuses []string) (int64, error)
	// CountToday 统计今日订单数
	CountToday(ctx context.Context, userID uint) (int64, error)
	// SumAmountByCurrency 按币种汇总金额
	SumAmountByCurrency(ctx context.Context, userID uint, excludeStatus string) ([]CurrencyAmount, error)
	// SumCommission 汇总佣金
	SumCommission(ctx context.Context, userID uint, status string) (*CommissionSummary, error)
	// ListByAuthIDAndTimeRange 根据授权ID和时间范围查询订单
	ListByAuthIDAndTimeRange(ctx context.Context, authID uint, since, to time.Time) ([]Order, error)
	// ListByAuthIDStatusAndTimeRange 根据授权ID、状态和时间范围查询订单
	ListByAuthIDStatusAndTimeRange(ctx context.Context, authID uint, status string, since, to time.Time) ([]Order, error)
	// GetRecentOrders 获取最近订单
	GetRecentOrders(ctx context.Context, userID uint, limit int) ([]Order, error)
}

// OrderItemRepository 订单商品仓储接口
type OrderItemRepository interface {
	// Create 创建订单商品
	Create(ctx context.Context, item *OrderItem) error
	// BatchCreate 批量创建订单商品
	BatchCreate(ctx context.Context, items []OrderItem) error
	// ListByOrderID 根据订单ID查询商品列表
	ListByOrderID(ctx context.Context, orderID uint) ([]OrderItem, error)
}

// OrderDailyStatRepository 订单每日统计仓储接口
type OrderDailyStatRepository interface {
	// FindByUserDateCurrency 根据用户ID、日期和币种查找
	FindByUserDateCurrency(ctx context.Context, userID uint, date time.Time, currency string) (*OrderDailyStat, error)
	// Create 创建统计记录
	Create(ctx context.Context, stat *OrderDailyStat) error
	// Update 更新统计记录
	Update(ctx context.Context, stat *OrderDailyStat) error
	// UpdateFields 更新指定字段
	UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error
	// ListByUserAndDateRange 根据用户和日期范围查询
	ListByUserAndDateRange(ctx context.Context, userID uint, start, end time.Time, currency string) ([]OrderDailyStat, error)
	// DeleteByUserID 删除用户的所有统计记录
	DeleteByUserID(ctx context.Context, userID uint) error
	// GetDailyStats 获取每日统计数据（原始查询）
	GetDailyStats(ctx context.Context, userID uint, startDate, endDate time.Time) ([]DailyStat, error)
}
