package platform

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"autostack/internal/repository"
)

// ========== PlatformAuthRepository 实现 ==========

type gormPlatformAuthRepository struct {
	db *gorm.DB
}

// NewPlatformAuthRepository 创建平台授权仓储实例
func NewPlatformAuthRepository(db *gorm.DB) PlatformAuthRepository {
	return &gormPlatformAuthRepository{db: db}
}

func (r *gormPlatformAuthRepository) getDB(ctx context.Context) *gorm.DB {
	return repository.GetDB(ctx, r.db)
}

func (r *gormPlatformAuthRepository) FindByID(ctx context.Context, id uint) (*PlatformAuth, error) {
	var auth PlatformAuth
	if err := r.getDB(ctx).First(&auth, id).Error; err != nil {
		return nil, err
	}
	return &auth, nil
}

func (r *gormPlatformAuthRepository) FindByIDAndUserID(ctx context.Context, id, userID uint) (*PlatformAuth, error) {
	var auth PlatformAuth
	if err := r.getDB(ctx).Where("id = ? AND user_id = ?", id, userID).First(&auth).Error; err != nil {
		return nil, err
	}
	return &auth, nil
}

func (r *gormPlatformAuthRepository) List(ctx context.Context, query *PlatformAuthQuery) ([]PlatformAuth, int64, error) {
	var auths []PlatformAuth
	var total int64
	db := r.getDB(ctx)

	q := db.Model(&PlatformAuth{})
	if query.UserID > 0 {
		q = q.Where("user_id = ?", query.UserID)
	}
	if query.Platform != "" {
		q = q.Where("platform = ?", query.Platform)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	if err := q.Offset(offset).Limit(query.PageSize).Order("id DESC").Find(&auths).Error; err != nil {
		return nil, 0, err
	}

	return auths, total, nil
}

func (r *gormPlatformAuthRepository) ListActive(ctx context.Context) ([]PlatformAuth, error) {
	var auths []PlatformAuth
	if err := r.getDB(ctx).Where("status = ?", AuthStatusActive).Find(&auths).Error; err != nil {
		return nil, err
	}
	return auths, nil
}

func (r *gormPlatformAuthRepository) ListByUserID(ctx context.Context, userID uint) ([]PlatformAuth, error) {
	var auths []PlatformAuth
	if err := r.getDB(ctx).Where("user_id = ?", userID).Order("id DESC").Find(&auths).Error; err != nil {
		return nil, err
	}
	return auths, nil
}

func (r *gormPlatformAuthRepository) Create(ctx context.Context, auth *PlatformAuth) error {
	return r.getDB(ctx).Create(auth).Error
}

func (r *gormPlatformAuthRepository) Update(ctx context.Context, auth *PlatformAuth) error {
	return r.getDB(ctx).Save(auth).Error
}

func (r *gormPlatformAuthRepository) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	return r.getDB(ctx).Model(&PlatformAuth{}).Where("id = ?", id).Updates(fields).Error
}

func (r *gormPlatformAuthRepository) Delete(ctx context.Context, id, userID uint) (int64, error) {
	result := r.getDB(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&PlatformAuth{})
	return result.RowsAffected, result.Error
}

func (r *gormPlatformAuthRepository) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	var count int64
	if err := r.getDB(ctx).Model(&PlatformAuth{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *gormPlatformAuthRepository) CountActiveByUserID(ctx context.Context, userID uint) (int64, error) {
	var count int64
	if err := r.getDB(ctx).Model(&PlatformAuth{}).Where("user_id = ? AND status = ?", userID, AuthStatusActive).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// ========== RequestLogRepository 实现 ==========

type gormRequestLogRepository struct {
	db *gorm.DB
}

// NewRequestLogRepository 创建请求日志仓储实例
func NewRequestLogRepository(db *gorm.DB) RequestLogRepository {
	return &gormRequestLogRepository{db: db}
}

func (r *gormRequestLogRepository) getDB(ctx context.Context) *gorm.DB {
	return repository.GetDB(ctx, r.db)
}

func (r *gormRequestLogRepository) Create(ctx context.Context, log *OrdersRequestLog) error {
	return r.getDB(ctx).Create(log).Error
}

func (r *gormRequestLogRepository) List(ctx context.Context, query *RequestLogQuery) ([]OrdersRequestLog, int64, error) {
	var logs []OrdersRequestLog
	var total int64
	db := r.getDB(ctx)

	q := db.Model(&OrdersRequestLog{})
	if query.PlatformAuthID > 0 {
		q = q.Where("platform_auth_id = ?", query.PlatformAuthID)
	}
	if query.RequestType != "" {
		q = q.Where("request_type = ?", query.RequestType)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	if err := q.Offset(offset).Limit(query.PageSize).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *gormRequestLogRepository) ListByAuthID(ctx context.Context, authID uint, limit int) ([]OrdersRequestLog, error) {
	var logs []OrdersRequestLog
	if err := r.getDB(ctx).Where("platform_auth_id = ?", authID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// ========== CashFlowRepository 实现 ==========

type gormCashFlowRepository struct {
	db *gorm.DB
}

// NewCashFlowRepository 创建现金流仓储实例
func NewCashFlowRepository(db *gorm.DB) CashFlowRepository {
	return &gormCashFlowRepository{db: db}
}

func (r *gormCashFlowRepository) getDB(ctx context.Context) *gorm.DB {
	return repository.GetDB(ctx, r.db)
}

func (r *gormCashFlowRepository) FindByID(ctx context.Context, id uint) (*CashFlowStatement, error) {
	var statement CashFlowStatement
	if err := r.getDB(ctx).First(&statement, id).Error; err != nil {
		return nil, err
	}
	return &statement, nil
}

func (r *gormCashFlowRepository) FindByIDAndUserID(ctx context.Context, id, userID uint) (*CashFlowStatement, error) {
	var statement CashFlowStatement
	if err := r.getDB(ctx).Where("id = ? AND user_id = ?", id, userID).First(&statement).Error; err != nil {
		return nil, err
	}
	return &statement, nil
}

func (r *gormCashFlowRepository) FindByAuthAndPeriod(ctx context.Context, authID uint, periodBegin time.Time) (*CashFlowStatement, error) {
	var statement CashFlowStatement
	if err := r.getDB(ctx).Where("platform_auth_id = ? AND period_begin = ?", authID, periodBegin).First(&statement).Error; err != nil {
		return nil, err
	}
	return &statement, nil
}

func (r *gormCashFlowRepository) List(ctx context.Context, query *CashFlowQuery) ([]CashFlowStatement, int64, error) {
	var statements []CashFlowStatement
	var total int64
	db := r.getDB(ctx)

	q := db.Model(&CashFlowStatement{})
	if query.UserID > 0 {
		q = q.Where("user_id = ?", query.UserID)
	}
	if query.PlatformAuthID > 0 {
		q = q.Where("platform_auth_id = ?", query.PlatformAuthID)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	if err := q.Order("period_end DESC").Offset(offset).Limit(query.PageSize).Find(&statements).Error; err != nil {
		return nil, 0, err
	}

	return statements, total, nil
}

func (r *gormCashFlowRepository) Create(ctx context.Context, statement *CashFlowStatement) error {
	return r.getDB(ctx).Create(statement).Error
}

func (r *gormCashFlowRepository) Update(ctx context.Context, statement *CashFlowStatement) error {
	return r.getDB(ctx).Save(statement).Error
}

// ========== 错误定义 ==========

var (
	ErrAuthNotFound     = errors.New("授权不存在")
	ErrCashFlowNotFound = errors.New("现金流报表不存在")
)
