package product

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"autostack/internal/repository"
)

// ========== ProductRepository 实现 ==========

type gormProductRepository struct {
	db *gorm.DB
}

// NewProductRepository 创建产品仓储实例
func NewProductRepository(db *gorm.DB) ProductRepository {
	return &gormProductRepository{db: db}
}

func (r *gormProductRepository) getDB(ctx context.Context) *gorm.DB {
	return repository.GetDB(ctx, r.db)
}

func (r *gormProductRepository) FindByID(ctx context.Context, id uint) (*Product, error) {
	var product Product
	if err := r.getDB(ctx).First(&product, id).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *gormProductRepository) FindBySKU(ctx context.Context, sku string) (*Product, error) {
	var product Product
	if err := r.getDB(ctx).Where("sku = ?", sku).First(&product).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *gormProductRepository) List(ctx context.Context, query *ProductQuery) ([]Product, int64, error) {
	var products []Product
	var total int64
	db := r.getDB(ctx)

	q := db.Model(&Product{})
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		q = q.Where("sku LIKE ? OR name LIKE ?", like, like)
	}
	if query.WarehouseID > 0 {
		q = q.Where("wid = ?", query.WarehouseID)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	if err := q.Order("id DESC").Offset(offset).Limit(query.PageSize).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *gormProductRepository) Create(ctx context.Context, product *Product) error {
	return r.getDB(ctx).Create(product).Error
}

func (r *gormProductRepository) Update(ctx context.Context, product *Product) error {
	return r.getDB(ctx).Save(product).Error
}

func (r *gormProductRepository) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	return r.getDB(ctx).Model(&Product{}).Where("id = ?", id).Updates(fields).Error
}

func (r *gormProductRepository) Delete(ctx context.Context, id uint) error {
	return r.getDB(ctx).Delete(&Product{}, id).Error
}

func (r *gormProductRepository) CountBySKU(ctx context.Context, sku string) (int64, error) {
	var count int64
	if err := r.getDB(ctx).Model(&Product{}).Where("sku = ?", sku).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *gormProductRepository) UpdateStock(ctx context.Context, id uint, delta int) error {
	return r.getDB(ctx).Model(&Product{}).Where("id = ?", id).
		Update("stock", gorm.Expr("stock + ?", delta)).Error
}

func (r *gormProductRepository) FindAll(ctx context.Context) ([]Product, error) {
	var products []Product
	if err := r.getDB(ctx).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

// ========== PlatformProductRepository 实现 ==========

type gormPlatformProductRepository struct {
	db *gorm.DB
}

// NewPlatformProductRepository 创建平台产品仓储实例
func NewPlatformProductRepository(db *gorm.DB) PlatformProductRepository {
	return &gormPlatformProductRepository{db: db}
}

func (r *gormPlatformProductRepository) getDB(ctx context.Context) *gorm.DB {
	return repository.GetDB(ctx, r.db)
}

func (r *gormPlatformProductRepository) FindByID(ctx context.Context, id uint) (*PlatformProduct, error) {
	var product PlatformProduct
	if err := r.getDB(ctx).First(&product, id).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *gormPlatformProductRepository) FindByAccountAndUniqueCode(ctx context.Context, accountID uint, uniqueCode string) (*PlatformProduct, error) {
	var product PlatformProduct
	if err := r.getDB(ctx).Where("platform_account_id = ? AND unique_code = ?", accountID, uniqueCode).First(&product).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *gormPlatformProductRepository) List(ctx context.Context, query *PlatformProductQuery) ([]PlatformProduct, int64, error) {
	var products []PlatformProduct
	var total int64
	db := r.getDB(ctx)

	q := db.Model(&PlatformProduct{})
	if query.PlatformAuthID > 0 {
		q = q.Where("platform_auth_id = ?", query.PlatformAuthID)
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		// 使用子查询搜索本地SKU，避免JOIN导致的重复记录和性能问题
		q = q.Where(`(
			platform_sku LIKE ? OR 
			name LIKE ? OR 
			id IN (
				SELECT pm.platform_product_id 
				FROM product_mappings pm 
				JOIN products p ON p.id = pm.product_id 
				WHERE p.sku LIKE ?
			)
		)`, like, like, like)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	if err := q.Preload("ProductMapping.Product").Order("id DESC").Offset(offset).Limit(query.PageSize).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *gormPlatformProductRepository) Save(ctx context.Context, product *PlatformProduct) error {
	return r.getDB(ctx).Save(product).Error
}

func (r *gormPlatformProductRepository) ListByAuthID(ctx context.Context, authID uint) ([]PlatformProduct, error) {
	var products []PlatformProduct
	if err := r.getDB(ctx).Where("platform_auth_id = ?", authID).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

// ========== ProductMappingRepository 实现 ==========

type gormProductMappingRepository struct {
	db *gorm.DB
}

// NewProductMappingRepository 创建产品映射仓储实例
func NewProductMappingRepository(db *gorm.DB) ProductMappingRepository {
	return &gormProductMappingRepository{db: db}
}

func (r *gormProductMappingRepository) getDB(ctx context.Context) *gorm.DB {
	return repository.GetDB(ctx, r.db)
}

func (r *gormProductMappingRepository) FindByCompositeKey(ctx context.Context, wid, accountID, productID, platformProductID uint) (*ProductMapping, error) {
	var mapping ProductMapping
	if err := r.getDB(ctx).Where(
		"wid = ? AND platform_account_id = ? AND product_id = ? AND platform_product_id = ?",
		wid, accountID, productID, platformProductID,
	).First(&mapping).Error; err != nil {
		return nil, err
	}
	return &mapping, nil
}

func (r *gormProductMappingRepository) FindByPlatformProductID(ctx context.Context, platformProductID uint) (*ProductMapping, error) {
	var mapping ProductMapping
	if err := r.getDB(ctx).Where("platform_product_id = ?", platformProductID).First(&mapping).Error; err != nil {
		return nil, err
	}
	return &mapping, nil
}

func (r *gormProductMappingRepository) Create(ctx context.Context, mapping *ProductMapping) error {
	return r.getDB(ctx).Create(mapping).Error
}

func (r *gormProductMappingRepository) DeleteByPlatformProductID(ctx context.Context, platformProductID uint) error {
	return r.getDB(ctx).Where("platform_product_id = ?", platformProductID).Delete(&ProductMapping{}).Error
}

func (r *gormProductMappingRepository) CountByProductID(ctx context.Context, productID uint) (int64, error) {
	var count int64
	if err := r.getDB(ctx).Model(&ProductMapping{}).Where("product_id = ?", productID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// ========== SyncTaskRepository 实现 ==========

type gormSyncTaskRepository struct {
	db *gorm.DB
}

// NewSyncTaskRepository 创建同步任务仓储实例
func NewSyncTaskRepository(db *gorm.DB) SyncTaskRepository {
	return &gormSyncTaskRepository{db: db}
}

func (r *gormSyncTaskRepository) getDB(ctx context.Context) *gorm.DB {
	return repository.GetDB(ctx, r.db)
}

func (r *gormSyncTaskRepository) FindByID(ctx context.Context, id uint) (*PlatformSyncTask, error) {
	var task PlatformSyncTask
	if err := r.getDB(ctx).First(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *gormSyncTaskRepository) FindPendingOrRunning(ctx context.Context, authID uint, taskType string) (*PlatformSyncTask, error) {
	var task PlatformSyncTask
	err := r.getDB(ctx).Where(
		"platform_auth_id = ? AND task_type = ? AND status IN ?",
		authID, taskType, []string{SyncTaskStatusPending, SyncTaskStatusRunning},
	).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *gormSyncTaskRepository) FindPendingTasks(ctx context.Context, lockTimeout time.Duration, limit int) ([]PlatformSyncTask, error) {
	var tasks []PlatformSyncTask
	now := time.Now()
	lockExpired := now.Add(-lockTimeout)

	if err := r.getDB(ctx).Where(
		"(status = ? OR (status = ? AND locked_at < ?)) AND retry_count < max_retry",
		SyncTaskStatusPending, SyncTaskStatusRunning, lockExpired,
	).Order("priority DESC, created_at ASC").Limit(limit).Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *gormSyncTaskRepository) List(ctx context.Context, query *SyncTaskQuery) ([]PlatformSyncTask, int64, error) {
	var tasks []PlatformSyncTask
	var total int64
	db := r.getDB(ctx)

	q := db.Model(&PlatformSyncTask{})
	if query.Status != "" {
		q = q.Where("status = ?", query.Status)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	if err := q.Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

func (r *gormSyncTaskRepository) Create(ctx context.Context, task *PlatformSyncTask) error {
	return r.getDB(ctx).Create(task).Error
}

func (r *gormSyncTaskRepository) UpdatePriority(ctx context.Context, id uint, priority int) error {
	return r.getDB(ctx).Model(&PlatformSyncTask{}).Where("id = ?", id).Update("priority", priority).Error
}

func (r *gormSyncTaskRepository) UpdateStatus(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.getDB(ctx).Model(&PlatformSyncTask{}).Where("id = ?", id).Updates(updates).Error
}

func (r *gormSyncTaskRepository) DeleteOldTasks(ctx context.Context, before time.Time, statuses []string) (int64, error) {
	result := r.getDB(ctx).Where("created_at < ? AND status IN ?", before, statuses).Delete(&PlatformSyncTask{})
	return result.RowsAffected, result.Error
}

func (r *gormSyncTaskRepository) TryLock(ctx context.Context, id uint, lockID string, lockTimeout time.Duration) (bool, error) {
	now := time.Now()
	lockExpired := now.Add(-lockTimeout)

	result := r.getDB(ctx).Model(&PlatformSyncTask{}).
		Where("id = ? AND (status = ? OR (status = ? AND locked_at < ?))",
			id, SyncTaskStatusPending, SyncTaskStatusRunning, lockExpired).
		Updates(map[string]interface{}{
			"status":    SyncTaskStatusRunning,
			"locked_at": now,
			"locked_by": lockID,
		})

	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// ========== ProductSupplierRepository 实现 ==========

type gormProductSupplierRepository struct {
	db *gorm.DB
}

// NewProductSupplierRepository 创建产品供应商仓储实例
func NewProductSupplierRepository(db *gorm.DB) ProductSupplierRepository {
	return &gormProductSupplierRepository{db: db}
}

func (r *gormProductSupplierRepository) getDB(ctx context.Context) *gorm.DB {
	return repository.GetDB(ctx, r.db)
}

func (r *gormProductSupplierRepository) FindByID(ctx context.Context, id uint) (*ProductSupplier, error) {
	var supplier ProductSupplier
	if err := r.getDB(ctx).Preload("Product").First(&supplier, id).Error; err != nil {
		return nil, err
	}
	return &supplier, nil
}

func (r *gormProductSupplierRepository) FindByProductID(ctx context.Context, productID uint) ([]ProductSupplier, error) {
	var suppliers []ProductSupplier
	if err := r.getDB(ctx).Where("product_id = ?", productID).
		Order("is_default DESC, id ASC").Find(&suppliers).Error; err != nil {
		return nil, err
	}
	return suppliers, nil
}

func (r *gormProductSupplierRepository) FindDefaultByProductID(ctx context.Context, productID uint) (*ProductSupplier, error) {
	var supplier ProductSupplier
	if err := r.getDB(ctx).Where("product_id = ? AND is_default = ?", productID, true).
		First(&supplier).Error; err != nil {
		return nil, err
	}
	return &supplier, nil
}

func (r *gormProductSupplierRepository) FindDefaultByProductIDs(ctx context.Context, productIDs []uint) ([]ProductSupplier, error) {
	if len(productIDs) == 0 {
		return nil, nil
	}
	var suppliers []ProductSupplier
	if err := r.getDB(ctx).Where("product_id IN ? AND is_default = ?", productIDs, true).
		Find(&suppliers).Error; err != nil {
		return nil, err
	}
	return suppliers, nil
}

func (r *gormProductSupplierRepository) FindByProductAndName(ctx context.Context, productID uint, supplierName string) (*ProductSupplier, error) {
	var supplier ProductSupplier
	if err := r.getDB(ctx).Where("product_id = ? AND supplier_name = ?", productID, supplierName).
		First(&supplier).Error; err != nil {
		return nil, err
	}
	return &supplier, nil
}

func (r *gormProductSupplierRepository) List(ctx context.Context, query *ProductSupplierQuery) ([]ProductSupplier, int64, error) {
	var suppliers []ProductSupplier
	var total int64
	db := r.getDB(ctx)

	q := db.Model(&ProductSupplier{})

	if query.ProductID > 0 {
		q = q.Where("product_id = ?", query.ProductID)
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		q = q.Where("supplier_name LIKE ?", like)
	}
	if query.Status != "" {
		q = q.Where("status = ?", query.Status)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	if err := q.Preload("Product").Order("is_default DESC, id DESC").
		Offset(offset).Limit(query.PageSize).Find(&suppliers).Error; err != nil {
		return nil, 0, err
	}

	return suppliers, total, nil
}

func (r *gormProductSupplierRepository) Create(ctx context.Context, supplier *ProductSupplier) error {
	return r.getDB(ctx).Create(supplier).Error
}

func (r *gormProductSupplierRepository) Update(ctx context.Context, supplier *ProductSupplier) error {
	return r.getDB(ctx).Save(supplier).Error
}

func (r *gormProductSupplierRepository) Delete(ctx context.Context, id uint) error {
	return r.getDB(ctx).Delete(&ProductSupplier{}, id).Error
}

func (r *gormProductSupplierRepository) SetDefault(ctx context.Context, productID uint, supplierID uint) error {
	db := r.getDB(ctx)
	// 先取消该产品所有供应商的默认状态
	if err := db.Model(&ProductSupplier{}).Where("product_id = ?", productID).
		Update("is_default", false).Error; err != nil {
		return err
	}
	// 再设置指定供应商为默认
	return db.Model(&ProductSupplier{}).Where("id = ?", supplierID).
		Update("is_default", true).Error
}

func (r *gormProductSupplierRepository) CountByProductID(ctx context.Context, productID uint) (int64, error) {
	var count int64
	if err := r.getDB(ctx).Model(&ProductSupplier{}).Where("product_id = ?", productID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// ========== 错误定义 ==========

var (
	ErrProductNotFound         = errors.New("产品不存在")
	ErrPlatformProductNotFound = errors.New("平台产品不存在")
	ErrMappingNotFound         = errors.New("映射关系不存在")
	ErrTaskNotFound            = errors.New("任务不存在")
	ErrSupplierNotFound        = errors.New("供应商不存在")
)
