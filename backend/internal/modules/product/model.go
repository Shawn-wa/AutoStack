package product

import (
	"time"
)

// ========== 同步任务相关 ==========

// 任务类型枚举
const (
	SyncTaskTypeProduct    = "product"    // 产品同步
	SyncTaskTypeOrder      = "order"      // 订单同步
	SyncTaskTypeCommission = "commission" // 佣金同步
	SyncTaskTypeCashFlow   = "cash_flow"  // 现金流同步
)

// 任务状态枚举
const (
	SyncTaskStatusPending = "pending" // 待执行
	SyncTaskStatusRunning = "running" // 执行中
	SyncTaskStatusSuccess = "success" // 成功
	SyncTaskStatusFailed  = "failed"  // 失败
)

// PlatformSyncTask 平台同步任务
type PlatformSyncTask struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	PlatformAuthID uint       `gorm:"index;not null" json:"platform_auth_id"`
	TaskType       string     `gorm:"size:32;not null;index" json:"task_type"`
	Status         string     `gorm:"size:16;not null;default:pending;index" json:"status"`
	Priority       int        `gorm:"default:10;index" json:"priority"` // 优先级，值越大越优先，默认10
	RetryCount     int        `gorm:"default:0" json:"retry_count"`
	MaxRetry       int        `gorm:"default:5" json:"max_retry"`
	ErrorMessage   string     `gorm:"type:text" json:"error_message"`
	LockedAt       *time.Time `gorm:"index" json:"locked_at"`
	LockedBy       string     `gorm:"size:64" json:"locked_by"`
	StartedAt      *time.Time `json:"started_at"`
	FinishedAt     *time.Time `json:"finished_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// TableName 指定表名
func (PlatformSyncTask) TableName() string {
	return "platform_sync_tasks"
}

// Product 本地产品模型
type Product struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	SKU        string    `gorm:"size:100;not null;uniqueIndex" json:"sku"`
	Name       string    `gorm:"size:255;not null" json:"name"`
	Image      string    `gorm:"size:500" json:"image"`
	CostPrice  float64   `gorm:"type:decimal(10,2);default:0" json:"cost_price"`
	Weight     float64   `gorm:"type:decimal(10,2);default:0" json:"weight"`     // kg
	Dimensions string    `gorm:"size:50" json:"dimensions"`                      // L*W*H cm
	Stock      int       `gorm:"default:0" json:"stock"`                         // 库存数量
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Product) TableName() string {
	return "products"
}

// PlatformProduct 平台产品模型（同步数据）
type PlatformProduct struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Platform       string    `gorm:"size:50;not null;index:idx_platform_auth_sku" json:"platform"`
	PlatformAuthID uint      `gorm:"index:idx_platform_auth_sku;not null" json:"platform_auth_id"`
	PlatformSKU    string    `gorm:"size:100;not null;index:idx_platform_auth_sku" json:"platform_sku"`
	Name           string    `gorm:"size:500" json:"name"`
	Image          string    `gorm:"column:image;size:500" json:"image"` // 产品主图URL
	Stock          int       `gorm:"default:0" json:"stock"`
	Price          float64   `gorm:"type:decimal(10,2)" json:"price"`
	Currency       string    `gorm:"size:10" json:"currency"`
	Status         string    `gorm:"size:50" json:"status"`
	RawData        string    `gorm:"type:longtext" json:"-"` // 原始JSON数据
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// 关联
	ProductMapping *ProductMapping `gorm:"foreignKey:PlatformProductID" json:"product_mapping,omitempty"`
}

// TableName 指定表名
func (PlatformProduct) TableName() string {
	return "platform_products"
}

// ProductMapping 产品映射关系
type ProductMapping struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	PlatformProductID uint      `gorm:"uniqueIndex;not null" json:"platform_product_id"`
	ProductID         uint      `gorm:"index;not null" json:"product_id"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`

	// 关联
	Product         *Product         `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	PlatformProduct *PlatformProduct `gorm:"foreignKey:PlatformProductID" json:"-"`
}

// TableName 指定表名
func (ProductMapping) TableName() string {
	return "product_mappings"
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
	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// TableName 指定表名
func (StockInOrderItem) TableName() string {
	return "stock_in_order_items"
}

// ========== 仓库与库存相关 ==========

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
	Product   *Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Warehouse *Warehouse `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
}

// TableName 指定表名
func (WarehouseCenterInventory) TableName() string {
	return "warehouse_center_inventory"
}

// TotalStock 计算总库存
func (w *WarehouseCenterInventory) TotalStock() int {
	return w.AvailableStock + w.LockedStock + w.InTransitStock
}
