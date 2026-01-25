package product

import (
	"context"
	"time"
)

// ProductRepository 产品仓储接口
type ProductRepository interface {
	// FindByID 根据ID查找产品
	FindByID(ctx context.Context, id uint) (*Product, error)
	// FindBySKU 根据SKU查找产品
	FindBySKU(ctx context.Context, sku string) (*Product, error)
	// List 分页查询产品列表
	List(ctx context.Context, query *ProductQuery) ([]Product, int64, error)
	// Create 创建产品
	Create(ctx context.Context, product *Product) error
	// Update 更新产品
	Update(ctx context.Context, product *Product) error
	// UpdateFields 更新指定字段
	UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error
	// Delete 删除产品
	Delete(ctx context.Context, id uint) error
	// CountBySKU 统计SKU数量
	CountBySKU(ctx context.Context, sku string) (int64, error)
	// UpdateStock 更新库存（增量）
	UpdateStock(ctx context.Context, id uint, delta int) error
	// FindAll 查询所有产品
	FindAll(ctx context.Context) ([]Product, error)
}

// PlatformProductRepository 平台产品仓储接口
type PlatformProductRepository interface {
	// FindByID 根据ID查找平台产品
	FindByID(ctx context.Context, id uint) (*PlatformProduct, error)
	// FindByAccountAndUniqueCode 根据账户ID和唯一编码查找
	FindByAccountAndUniqueCode(ctx context.Context, accountID uint, uniqueCode string) (*PlatformProduct, error)
	// List 分页查询平台产品列表
	List(ctx context.Context, query *PlatformProductQuery) ([]PlatformProduct, int64, error)
	// Save 保存平台产品（创建或更新）
	Save(ctx context.Context, product *PlatformProduct) error
	// ListByAuthID 根据授权ID查询所有平台产品
	ListByAuthID(ctx context.Context, authID uint) ([]PlatformProduct, error)
}

// ProductMappingRepository 产品映射仓储接口
type ProductMappingRepository interface {
	// FindByCompositeKey 根据复合键查找映射
	FindByCompositeKey(ctx context.Context, wid, accountID, productID, platformProductID uint) (*ProductMapping, error)
	// FindByPlatformProductID 根据平台产品ID查找映射
	FindByPlatformProductID(ctx context.Context, platformProductID uint) (*ProductMapping, error)
	// Create 创建映射
	Create(ctx context.Context, mapping *ProductMapping) error
	// DeleteByPlatformProductID 根据平台产品ID删除映射
	DeleteByPlatformProductID(ctx context.Context, platformProductID uint) error
	// CountByProductID 统计产品关联的映射数量
	CountByProductID(ctx context.Context, productID uint) (int64, error)
}

// SyncTaskRepository 同步任务仓储接口
type SyncTaskRepository interface {
	// FindByID 根据ID查找任务
	FindByID(ctx context.Context, id uint) (*PlatformSyncTask, error)
	// FindPendingOrRunning 查找待执行或执行中的任务
	FindPendingOrRunning(ctx context.Context, authID uint, taskType string) (*PlatformSyncTask, error)
	// FindPendingTasks 查找待处理的任务列表
	FindPendingTasks(ctx context.Context, lockTimeout time.Duration, limit int) ([]PlatformSyncTask, error)
	// List 分页查询任务列表
	List(ctx context.Context, query *SyncTaskQuery) ([]PlatformSyncTask, int64, error)
	// Create 创建任务
	Create(ctx context.Context, task *PlatformSyncTask) error
	// UpdatePriority 更新任务优先级
	UpdatePriority(ctx context.Context, id uint, priority int) error
	// UpdateStatus 更新任务状态
	UpdateStatus(ctx context.Context, id uint, updates map[string]interface{}) error
	// DeleteOldTasks 删除旧任务
	DeleteOldTasks(ctx context.Context, before time.Time, statuses []string) (int64, error)
	// TryLock 尝试锁定任务（乐观锁）
	TryLock(ctx context.Context, id uint, lockID string, lockTimeout time.Duration) (bool, error)
}
