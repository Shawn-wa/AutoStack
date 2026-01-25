package inventory

import (
	"context"
)

// WarehouseRepository 仓库仓储接口
type WarehouseRepository interface {
	// FindByID 根据ID查找仓库
	FindByID(ctx context.Context, id uint) (*Warehouse, error)
	// FindByCode 根据编码查找仓库
	FindByCode(ctx context.Context, code string) (*Warehouse, error)
	// FindFirstActive 查找第一个启用的仓库
	FindFirstActive(ctx context.Context) (*Warehouse, error)
	// List 查询仓库列表
	List(ctx context.Context, query *WarehouseQuery) ([]Warehouse, error)
	// ListActive 查询所有启用的仓库
	ListActive(ctx context.Context) ([]Warehouse, error)
	// Create 创建仓库
	Create(ctx context.Context, warehouse *Warehouse) error
	// Update 更新仓库
	Update(ctx context.Context, warehouse *Warehouse) error
	// Count 统计仓库数量
	Count(ctx context.Context) (int64, error)
	// CountByCode 统计指定编码的仓库数量
	CountByCode(ctx context.Context, code string) (int64, error)
}

// InventoryRepository 库存仓储接口
type InventoryRepository interface {
	// FindByID 根据ID查找库存记录
	FindByID(ctx context.Context, id uint) (*WarehouseCenterInventory, error)
	// FindByProductAndWarehouse 根据产品和仓库查找库存
	FindByProductAndWarehouse(ctx context.Context, productID, warehouseID uint) (*WarehouseCenterInventory, error)
	// List 分页查询库存列表
	List(ctx context.Context, query *InventoryQuery) ([]WarehouseCenterInventory, int64, error)
	// Create 创建库存记录
	Create(ctx context.Context, inventory *WarehouseCenterInventory) error
	// Update 更新库存记录
	Update(ctx context.Context, inventory *WarehouseCenterInventory) error
	// UpdateFields 更新指定字段
	UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error
	// UpdateAvailableStock 更新可用库存（增量）
	UpdateAvailableStock(ctx context.Context, id uint, delta int) error
	// CountByProductAndWarehouse 统计产品在仓库的库存记录数
	CountByProductAndWarehouse(ctx context.Context, productID, warehouseID uint) (int64, error)
	// GetStockSummaryBySKUs 根据SKU列表获取库存汇总
	GetStockSummaryBySKUs(ctx context.Context, skus []string) ([]StockSummary, error)
}

// StockInOrderRepository 入库单仓储接口
type StockInOrderRepository interface {
	// FindByID 根据ID查找入库单
	FindByID(ctx context.Context, id uint) (*StockInOrder, error)
	// FindByIDWithDetails 根据ID查找入库单（包含关联数据）
	FindByIDWithDetails(ctx context.Context, id uint) (*StockInOrder, error)
	// List 分页查询入库单列表
	List(ctx context.Context, query *StockInOrderQuery) ([]StockInOrder, int64, error)
	// Create 创建入库单
	Create(ctx context.Context, order *StockInOrder) error
	// Update 更新入库单
	Update(ctx context.Context, order *StockInOrder) error
}

// StockInOrderItemRepository 入库单明细仓储接口
type StockInOrderItemRepository interface {
	// Create 创建入库单明细
	Create(ctx context.Context, item *StockInOrderItem) error
	// ListByOrderID 根据入库单ID查询明细
	ListByOrderID(ctx context.Context, orderID uint) ([]StockInOrderItem, error)
	// BatchCreate 批量创建明细
	BatchCreate(ctx context.Context, items []StockInOrderItem) error
}
