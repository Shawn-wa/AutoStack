package platform

import (
	"context"
	"time"
)

// PlatformAuthRepository 平台授权仓储接口
type PlatformAuthRepository interface {
	// FindByID 根据ID查找授权
	FindByID(ctx context.Context, id uint) (*PlatformAuth, error)
	// FindByIDAndCompanyID 根据ID和企业ID查找授权
	FindByIDAndCompanyID(ctx context.Context, id, companyID uint) (*PlatformAuth, error)
	// List 分页查询授权列表
	List(ctx context.Context, query *PlatformAuthQuery) ([]PlatformAuth, int64, error)
	// ListActive 查询所有活跃的授权
	ListActive(ctx context.Context) ([]PlatformAuth, error)
	// ListByCompanyID 根据企业ID查询授权列表
	ListByCompanyID(ctx context.Context, companyID uint) ([]PlatformAuth, error)
	// Create 创建授权
	Create(ctx context.Context, auth *PlatformAuth) error
	// Update 更新授权
	Update(ctx context.Context, auth *PlatformAuth) error
	// UpdateFields 更新指定字段
	UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error
	// Delete 删除授权
	Delete(ctx context.Context, id, companyID uint) (int64, error)
	// CountByCompanyID 统计企业授权数
	CountByCompanyID(ctx context.Context, companyID uint) (int64, error)
	// CountActiveByCompanyID 统计企业活跃授权数
	CountActiveByCompanyID(ctx context.Context, companyID uint) (int64, error)
}

// RequestLogRepository 请求日志仓储接口
type RequestLogRepository interface {
	// Create 创建日志
	Create(ctx context.Context, log *OrdersRequestLog) error
	// List 分页查询日志
	List(ctx context.Context, query *RequestLogQuery) ([]OrdersRequestLog, int64, error)
	// ListByAuthID 根据授权ID查询日志
	ListByAuthID(ctx context.Context, authID uint, limit int) ([]OrdersRequestLog, error)
}

// CashFlowRepository 现金流仓储接口
type CashFlowRepository interface {
	// FindByID 根据ID查找
	FindByID(ctx context.Context, id uint) (*CashFlowStatement, error)
	// FindByIDAndCompanyID 根据ID和企业ID查找
	FindByIDAndCompanyID(ctx context.Context, id, companyID uint) (*CashFlowStatement, error)
	// FindByAuthAndPeriod 根据授权ID和周期开始时间查找
	FindByAuthAndPeriod(ctx context.Context, authID uint, periodBegin time.Time) (*CashFlowStatement, error)
	// List 分页查询
	List(ctx context.Context, query *CashFlowQuery) ([]CashFlowStatement, int64, error)
	// Create 创建
	Create(ctx context.Context, statement *CashFlowStatement) error
	// Update 更新
	Update(ctx context.Context, statement *CashFlowStatement) error
}
