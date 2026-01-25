package order

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"gorm.io/gorm"

	"autostack/internal/repository"
)

// ========== OrderRepository 实现 ==========

type gormOrderRepository struct {
	db *gorm.DB
}

// NewOrderRepository 创建订单仓储实例
func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &gormOrderRepository{db: db}
}

func (r *gormOrderRepository) getDB(ctx context.Context) *gorm.DB {
	return repository.GetDB(ctx, r.db)
}

func (r *gormOrderRepository) FindByID(ctx context.Context, id uint) (*Order, error) {
	var order Order
	if err := r.getDB(ctx).Preload("Items").First(&order, id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *gormOrderRepository) FindByIDAndUserID(ctx context.Context, id, userID uint) (*Order, error) {
	var order Order
	if err := r.getDB(ctx).Preload("Items").Where("id = ? AND user_id = ?", id, userID).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *gormOrderRepository) FindByPlatformOrderNo(ctx context.Context, orderNo string) (*Order, error) {
	var order Order
	if err := r.getDB(ctx).Where("platform_order_no = ?", orderNo).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *gormOrderRepository) List(ctx context.Context, query *OrderQuery) ([]Order, int64, error) {
	var orders []Order
	var total int64
	db := r.getDB(ctx)

	q := db.Model(&Order{}).Where("user_id = ?", query.UserID)

	// 应用过滤条件
	if query.Platform != "" {
		q = q.Where("platform = ?", query.Platform)
	}
	if query.AuthID > 0 {
		q = q.Where("platform_auth_id = ?", query.AuthID)
	}
	if query.Status != "" {
		if strings.Contains(query.Status, ",") {
			statuses := strings.Split(query.Status, ",")
			q = q.Where("status IN ?", statuses)
		} else {
			q = q.Where("status = ?", query.Status)
		}
	}
	if len(query.Statuses) > 0 {
		q = q.Where("status IN ?", query.Statuses)
	}
	if query.Keyword != "" {
		keyword := "%" + query.Keyword + "%"
		q = q.Where("platform_order_no LIKE ? OR recipient_name LIKE ?", keyword, keyword)
	}
	if query.StartTime != "" {
		startTimeStr, _ := url.QueryUnescape(query.StartTime)
		if startTimeStr == "" {
			startTimeStr = query.StartTime
		}
		if len(startTimeStr) == 10 {
			startTimeStr = startTimeStr + " 00:00:00"
		}
		q = q.Where(fmt.Sprintf("order_time >= '%s'", startTimeStr))
	}
	if query.EndTime != "" {
		endTimeStr, _ := url.QueryUnescape(query.EndTime)
		if endTimeStr == "" {
			endTimeStr = query.EndTime
		}
		if len(endTimeStr) == 10 {
			endTimeStr = endTimeStr + " 23:59:59"
		}
		q = q.Where(fmt.Sprintf("order_time <= '%s'", endTimeStr))
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	orderBy := "order_time DESC"
	if query.OrderBy != "" {
		orderBy = query.OrderBy
	}
	if err := q.Preload("Items").Offset(offset).Limit(query.PageSize).Order(orderBy).Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *gormOrderRepository) Create(ctx context.Context, order *Order) error {
	return r.getDB(ctx).Create(order).Error
}

func (r *gormOrderRepository) Update(ctx context.Context, order *Order) error {
	return r.getDB(ctx).Save(order).Error
}

func (r *gormOrderRepository) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	return r.getDB(ctx).Model(&Order{}).Where("id = ?", id).Updates(fields).Error
}

func (r *gormOrderRepository) UpdateByPlatformOrderNo(ctx context.Context, orderNo string, authID uint, fields map[string]interface{}) (int64, error) {
	result := r.getDB(ctx).Model(&Order{}).
		Where("platform_order_no = ? AND platform_auth_id = ?", orderNo, authID).
		Updates(fields)
	return result.RowsAffected, result.Error
}

func (r *gormOrderRepository) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	var count int64
	if err := r.getDB(ctx).Model(&Order{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *gormOrderRepository) CountByStatus(ctx context.Context, userID uint, status string) (int64, error) {
	var count int64
	if err := r.getDB(ctx).Model(&Order{}).Where("user_id = ? AND status = ?", userID, status).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *gormOrderRepository) CountByStatuses(ctx context.Context, userID uint, statuses []string) (int64, error) {
	var count int64
	if err := r.getDB(ctx).Model(&Order{}).Where("user_id = ? AND status IN ?", userID, statuses).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *gormOrderRepository) CountToday(ctx context.Context, userID uint) (int64, error) {
	var count int64
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if err := r.getDB(ctx).Model(&Order{}).Where("user_id = ? AND order_time >= ?", userID, today).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *gormOrderRepository) SumAmountByCurrency(ctx context.Context, userID uint, excludeStatus string) ([]CurrencyAmount, error) {
	var amounts []CurrencyAmount
	if err := r.getDB(ctx).Model(&Order{}).
		Select("currency, COALESCE(SUM(total_amount), 0) as amount").
		Where("user_id = ? AND status != ?", userID, excludeStatus).
		Group("currency").
		Scan(&amounts).Error; err != nil {
		return nil, err
	}
	return amounts, nil
}

func (r *gormOrderRepository) SumCommission(ctx context.Context, userID uint, status string) (*CommissionSummary, error) {
	var summary CommissionSummary
	if err := r.getDB(ctx).Model(&Order{}).
		Select(`
			COALESCE(SUM(profit_amount), 0) as total_profit,
			COALESCE(SUM(sale_commission), 0) as total_commission,
			COALESCE(SUM(services_amount), 0) as total_service_fee
		`).
		Where("user_id = ? AND status = ?", userID, status).
		Scan(&summary).Error; err != nil {
		return nil, err
	}
	return &summary, nil
}

func (r *gormOrderRepository) ListByAuthIDAndTimeRange(ctx context.Context, authID uint, since, to time.Time) ([]Order, error) {
	var orders []Order
	if err := r.getDB(ctx).
		Where("platform_auth_id = ? AND order_time >= ? AND order_time <= ?", authID, since, to).
		Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *gormOrderRepository) ListByAuthIDStatusAndTimeRange(ctx context.Context, authID uint, status string, since, to time.Time) ([]Order, error) {
	var orders []Order
	if err := r.getDB(ctx).
		Where("platform_auth_id = ? AND status = ? AND order_time >= ? AND order_time <= ?", authID, status, since, to).
		Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *gormOrderRepository) GetRecentOrders(ctx context.Context, userID uint, limit int) ([]Order, error) {
	var orders []Order
	if err := r.getDB(ctx).Where("user_id = ?", userID).
		Order("order_time DESC").
		Limit(limit).
		Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

// ========== OrderItemRepository 实现 ==========

type gormOrderItemRepository struct {
	db *gorm.DB
}

// NewOrderItemRepository 创建订单商品仓储实例
func NewOrderItemRepository(db *gorm.DB) OrderItemRepository {
	return &gormOrderItemRepository{db: db}
}

func (r *gormOrderItemRepository) getDB(ctx context.Context) *gorm.DB {
	return repository.GetDB(ctx, r.db)
}

func (r *gormOrderItemRepository) Create(ctx context.Context, item *OrderItem) error {
	return r.getDB(ctx).Create(item).Error
}

func (r *gormOrderItemRepository) BatchCreate(ctx context.Context, items []OrderItem) error {
	if len(items) == 0 {
		return nil
	}
	return r.getDB(ctx).Create(&items).Error
}

func (r *gormOrderItemRepository) ListByOrderID(ctx context.Context, orderID uint) ([]OrderItem, error) {
	var items []OrderItem
	if err := r.getDB(ctx).Where("order_id = ?", orderID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// ========== OrderDailyStatRepository 实现 ==========

type gormOrderDailyStatRepository struct {
	db *gorm.DB
}

// NewOrderDailyStatRepository 创建订单每日统计仓储实例
func NewOrderDailyStatRepository(db *gorm.DB) OrderDailyStatRepository {
	return &gormOrderDailyStatRepository{db: db}
}

func (r *gormOrderDailyStatRepository) getDB(ctx context.Context) *gorm.DB {
	return repository.GetDB(ctx, r.db)
}

func (r *gormOrderDailyStatRepository) FindByUserDateCurrency(ctx context.Context, userID uint, date time.Time, currency string) (*OrderDailyStat, error) {
	var stat OrderDailyStat
	if err := r.getDB(ctx).Where("user_id = ? AND stat_date = ? AND currency = ?", userID, date, currency).First(&stat).Error; err != nil {
		return nil, err
	}
	return &stat, nil
}

func (r *gormOrderDailyStatRepository) Create(ctx context.Context, stat *OrderDailyStat) error {
	return r.getDB(ctx).Create(stat).Error
}

func (r *gormOrderDailyStatRepository) Update(ctx context.Context, stat *OrderDailyStat) error {
	return r.getDB(ctx).Save(stat).Error
}

func (r *gormOrderDailyStatRepository) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	return r.getDB(ctx).Model(&OrderDailyStat{}).Where("id = ?", id).Updates(fields).Error
}

func (r *gormOrderDailyStatRepository) ListByUserAndDateRange(ctx context.Context, userID uint, start, end time.Time, currency string) ([]OrderDailyStat, error) {
	var stats []OrderDailyStat
	if err := r.getDB(ctx).Where("user_id = ? AND currency = ? AND stat_date >= ? AND stat_date < ?", userID, currency, start, end).
		Order("stat_date ASC").
		Find(&stats).Error; err != nil {
		return nil, err
	}
	return stats, nil
}

func (r *gormOrderDailyStatRepository) DeleteByUserID(ctx context.Context, userID uint) error {
	return r.getDB(ctx).Where("user_id = ?", userID).Delete(&OrderDailyStat{}).Error
}

func (r *gormOrderDailyStatRepository) GetDailyStats(ctx context.Context, userID uint, startDate, endDate time.Time) ([]DailyStat, error) {
	var stats []DailyStat
	if err := r.getDB(ctx).Table("orders").
		Select(`
			DATE(order_time) as date, 
			currency,
			COUNT(*) as count, 
			COALESCE(SUM(total_amount), 0) as amount
		`).
		Where("user_id = ? AND order_time >= ? AND order_time < ?", userID, startDate, endDate).
		Group("DATE(order_time), currency").
		Order("date ASC").
		Scan(&stats).Error; err != nil {
		return nil, err
	}
	return stats, nil
}

// ========== 错误定义 ==========

var (
	ErrOrderNotFound = errors.New("订单不存在")
)
