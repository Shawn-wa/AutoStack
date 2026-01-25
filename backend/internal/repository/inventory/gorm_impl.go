package inventory

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"autostack/internal/repository"
)

// ========== WarehouseRepository 实现 ==========

type gormWarehouseRepository struct {
	db *gorm.DB
}

// NewWarehouseRepository 创建仓库仓储实例
func NewWarehouseRepository(db *gorm.DB) WarehouseRepository {
	return &gormWarehouseRepository{db: db}
}

func (r *gormWarehouseRepository) getDB(ctx context.Context) *gorm.DB {
	return repository.GetDB(ctx, r.db)
}

func (r *gormWarehouseRepository) FindByID(ctx context.Context, id uint) (*Warehouse, error) {
	var warehouse Warehouse
	if err := r.getDB(ctx).First(&warehouse, id).Error; err != nil {
		return nil, err
	}
	return &warehouse, nil
}

func (r *gormWarehouseRepository) FindByCode(ctx context.Context, code string) (*Warehouse, error) {
	var warehouse Warehouse
	if err := r.getDB(ctx).Where("code = ?", code).First(&warehouse).Error; err != nil {
		return nil, err
	}
	return &warehouse, nil
}

func (r *gormWarehouseRepository) FindFirstActive(ctx context.Context) (*Warehouse, error) {
	var warehouse Warehouse
	if err := r.getDB(ctx).Where("status = ?", WarehouseStatusActive).First(&warehouse).Error; err != nil {
		return nil, err
	}
	return &warehouse, nil
}

func (r *gormWarehouseRepository) List(ctx context.Context, query *WarehouseQuery) ([]Warehouse, error) {
	var warehouses []Warehouse
	db := r.getDB(ctx)

	q := db.Model(&Warehouse{})
	if query != nil {
		if query.Type != "" && query.Type != "all" {
			q = q.Where("type = ?", query.Type)
		}
		if query.Status != "" {
			q = q.Where("status = ?", query.Status)
		}
	}

	if err := q.Order("id ASC").Find(&warehouses).Error; err != nil {
		return nil, err
	}
	return warehouses, nil
}

func (r *gormWarehouseRepository) ListActive(ctx context.Context) ([]Warehouse, error) {
	var warehouses []Warehouse
	if err := r.getDB(ctx).Where("status = ?", WarehouseStatusActive).Order("id ASC").Find(&warehouses).Error; err != nil {
		return nil, err
	}
	return warehouses, nil
}

func (r *gormWarehouseRepository) Create(ctx context.Context, warehouse *Warehouse) error {
	return r.getDB(ctx).Create(warehouse).Error
}

func (r *gormWarehouseRepository) Update(ctx context.Context, warehouse *Warehouse) error {
	return r.getDB(ctx).Save(warehouse).Error
}

func (r *gormWarehouseRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.getDB(ctx).Model(&Warehouse{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *gormWarehouseRepository) CountByCode(ctx context.Context, code string) (int64, error) {
	var count int64
	if err := r.getDB(ctx).Model(&Warehouse{}).Where("code = ?", code).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// ========== InventoryRepository 实现 ==========

type gormInventoryRepository struct {
	db *gorm.DB
}

// NewInventoryRepository 创建库存仓储实例
func NewInventoryRepository(db *gorm.DB) InventoryRepository {
	return &gormInventoryRepository{db: db}
}

func (r *gormInventoryRepository) getDB(ctx context.Context) *gorm.DB {
	return repository.GetDB(ctx, r.db)
}

func (r *gormInventoryRepository) FindByID(ctx context.Context, id uint) (*WarehouseCenterInventory, error) {
	var inventory WarehouseCenterInventory
	if err := r.getDB(ctx).First(&inventory, id).Error; err != nil {
		return nil, err
	}
	return &inventory, nil
}

func (r *gormInventoryRepository) FindByProductAndWarehouse(ctx context.Context, productID, warehouseID uint) (*WarehouseCenterInventory, error) {
	var inventory WarehouseCenterInventory
	if err := r.getDB(ctx).Where("product_id = ? AND warehouse_id = ?", productID, warehouseID).First(&inventory).Error; err != nil {
		return nil, err
	}
	return &inventory, nil
}

func (r *gormInventoryRepository) List(ctx context.Context, query *InventoryQuery) ([]WarehouseCenterInventory, int64, error) {
	var inventories []WarehouseCenterInventory
	var total int64
	db := r.getDB(ctx)

	q := db.Model(&WarehouseCenterInventory{})
	if query.WarehouseID > 0 {
		q = q.Where("warehouse_id = ?", query.WarehouseID)
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		q = q.Where("sku LIKE ?", like)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	if err := q.Preload("Product").Preload("Warehouse").
		Order("updated_at DESC").Offset(offset).Limit(query.PageSize).Find(&inventories).Error; err != nil {
		return nil, 0, err
	}

	return inventories, total, nil
}

func (r *gormInventoryRepository) Create(ctx context.Context, inventory *WarehouseCenterInventory) error {
	return r.getDB(ctx).Create(inventory).Error
}

func (r *gormInventoryRepository) Update(ctx context.Context, inventory *WarehouseCenterInventory) error {
	return r.getDB(ctx).Save(inventory).Error
}

func (r *gormInventoryRepository) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	return r.getDB(ctx).Model(&WarehouseCenterInventory{}).Where("id = ?", id).Updates(fields).Error
}

func (r *gormInventoryRepository) UpdateAvailableStock(ctx context.Context, id uint, delta int) error {
	return r.getDB(ctx).Model(&WarehouseCenterInventory{}).Where("id = ?", id).
		Update("available_stock", gorm.Expr("available_stock + ?", delta)).Error
}

func (r *gormInventoryRepository) CountByProductAndWarehouse(ctx context.Context, productID, warehouseID uint) (int64, error) {
	var count int64
	if err := r.getDB(ctx).Model(&WarehouseCenterInventory{}).
		Where("product_id = ? AND warehouse_id = ?", productID, warehouseID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *gormInventoryRepository) GetStockSummaryBySKUs(ctx context.Context, skus []string) ([]StockSummary, error) {
	var summaries []StockSummary
	if err := r.getDB(ctx).Table("warehouse_center_inventory").
		Select("sku, SUM(available_stock) as available_stock").
		Where("sku IN ?", skus).
		Group("sku").
		Scan(&summaries).Error; err != nil {
		return nil, err
	}
	return summaries, nil
}

// ========== StockInOrderRepository 实现 ==========

type gormStockInOrderRepository struct {
	db *gorm.DB
}

// NewStockInOrderRepository 创建入库单仓储实例
func NewStockInOrderRepository(db *gorm.DB) StockInOrderRepository {
	return &gormStockInOrderRepository{db: db}
}

func (r *gormStockInOrderRepository) getDB(ctx context.Context) *gorm.DB {
	return repository.GetDB(ctx, r.db)
}

func (r *gormStockInOrderRepository) FindByID(ctx context.Context, id uint) (*StockInOrder, error) {
	var order StockInOrder
	if err := r.getDB(ctx).First(&order, id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *gormStockInOrderRepository) FindByIDWithDetails(ctx context.Context, id uint) (*StockInOrder, error) {
	var order StockInOrder
	if err := r.getDB(ctx).Preload("Warehouse").Preload("Items").First(&order, id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *gormStockInOrderRepository) List(ctx context.Context, query *StockInOrderQuery) ([]StockInOrder, int64, error) {
	var orders []StockInOrder
	var total int64
	db := r.getDB(ctx)

	q := db.Model(&StockInOrder{})
	if query.Status != "" {
		q = q.Where("status = ?", query.Status)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	if err := q.Preload("Warehouse").Preload("Items").
		Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *gormStockInOrderRepository) Create(ctx context.Context, order *StockInOrder) error {
	return r.getDB(ctx).Create(order).Error
}

func (r *gormStockInOrderRepository) Update(ctx context.Context, order *StockInOrder) error {
	return r.getDB(ctx).Save(order).Error
}

// ========== StockInOrderItemRepository 实现 ==========

type gormStockInOrderItemRepository struct {
	db *gorm.DB
}

// NewStockInOrderItemRepository 创建入库单明细仓储实例
func NewStockInOrderItemRepository(db *gorm.DB) StockInOrderItemRepository {
	return &gormStockInOrderItemRepository{db: db}
}

func (r *gormStockInOrderItemRepository) getDB(ctx context.Context) *gorm.DB {
	return repository.GetDB(ctx, r.db)
}

func (r *gormStockInOrderItemRepository) Create(ctx context.Context, item *StockInOrderItem) error {
	return r.getDB(ctx).Create(item).Error
}

func (r *gormStockInOrderItemRepository) ListByOrderID(ctx context.Context, orderID uint) ([]StockInOrderItem, error) {
	var items []StockInOrderItem
	if err := r.getDB(ctx).Where("stock_in_order_id = ?", orderID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *gormStockInOrderItemRepository) BatchCreate(ctx context.Context, items []StockInOrderItem) error {
	if len(items) == 0 {
		return nil
	}
	return r.getDB(ctx).Create(&items).Error
}

// ========== 错误定义 ==========

var (
	ErrWarehouseNotFound  = errors.New("仓库不存在")
	ErrInventoryNotFound  = errors.New("库存记录不存在")
	ErrStockInOrderNotFound = errors.New("入库单不存在")
)
