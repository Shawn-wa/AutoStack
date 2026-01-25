package inventory

import (
	"time"

	productRepo "autostack/internal/repository/product"
)

// ========== 仓库相关 ==========

// 仓库状态
const (
	WarehouseStatusActive   = "active"   // 启用
	WarehouseStatusInactive = "inactive" // 停用
)

// 仓库类型
const (
	WarehouseTypeLocal    = "local"    // 本地仓
	WarehouseTypeOverseas = "overseas" // 海外仓
	WarehouseTypeFBA      = "fba"      // FBA仓
	WarehouseTypeThird    = "third"    // 第三方仓
	WarehouseTypeVirtual  = "virtual"  // 虚拟仓
)

// Warehouse 仓库
type Warehouse struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Code      string    `gorm:"size:32;uniqueIndex;not null" json:"code"`   // 仓库编码
	Name      string    `gorm:"size:100;not null" json:"name"`              // 仓库名称
	Type      string    `gorm:"size:32;default:local;index" json:"type"`    // 仓库类型
	Address   string    `gorm:"size:255" json:"address"`                    // 仓库地址
	Status    string    `gorm:"size:16;default:active;index" json:"status"` // 状态
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Warehouse) TableName() string {
	return "warehouses"
}

// ========== 库存相关 ==========

// WarehouseCenterInventory 仓库库存明细
type WarehouseCenterInventory struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	ProductID      uint      `gorm:"index;not null" json:"product_id"`              // 本地产品ID
	WarehouseID    uint      `gorm:"index;not null" json:"warehouse_id"`            // 仓库ID
	SKU            string    `gorm:"size:100;not null;index" json:"sku"`            // 系统SKU
	AvailableStock int       `gorm:"default:0" json:"available_stock"`              // 可用库存
	LockedStock    int       `gorm:"default:0" json:"locked_stock"`                 // 锁定库存
	InTransitStock int       `gorm:"default:0" json:"in_transit_stock"`             // 在途库存
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// 关联
	Product   *productRepo.Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Warehouse *Warehouse           `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
}

// TableName 指定表名
func (WarehouseCenterInventory) TableName() string {
	return "warehouse_center_inventory"
}

// TotalStock 计算总库存
func (w *WarehouseCenterInventory) TotalStock() int {
	return w.AvailableStock + w.LockedStock + w.InTransitStock
}

// ========== 入库单相关 ==========

// 入库单状态
const (
	StockInStatusPending   = "pending"   // 待入库
	StockInStatusCompleted = "completed" // 已完成
	StockInStatusCancelled = "cancelled" // 已取消
)

// StockInOrder 入库单
type StockInOrder struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	OrderNo     string    `gorm:"size:32;uniqueIndex;not null" json:"order_no"` // 入库单号
	WarehouseID uint      `gorm:"index;not null" json:"warehouse_id"`           // 入库仓库
	Status      string    `gorm:"size:16;default:pending;index" json:"status"`
	Remark      string    `gorm:"size:500" json:"remark"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 关联
	Warehouse *Warehouse         `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
	Items     []StockInOrderItem `gorm:"foreignKey:StockInOrderID" json:"items,omitempty"`
}

// TableName 指定表名
func (StockInOrder) TableName() string {
	return "stock_in_orders"
}

// StockInOrderItem 入库单明细
type StockInOrderItem struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	StockInOrderID uint      `gorm:"index;not null" json:"stock_in_order_id"`
	ProductID      uint      `gorm:"index;not null" json:"product_id"`
	SKU            string    `gorm:"size:100;not null" json:"sku"`
	ProductName    string    `gorm:"size:255" json:"product_name"`
	Quantity       int       `gorm:"not null" json:"quantity"` // 入库数量
	CreatedAt      time.Time `json:"created_at"`

	// 关联
	Product *productRepo.Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// TableName 指定表名
func (StockInOrderItem) TableName() string {
	return "stock_in_order_items"
}
