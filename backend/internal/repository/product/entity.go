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
	WID        uint      `gorm:"column:wid;index;default:0" json:"wid"`          // 仓库ID
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
	ID                uint      `gorm:"primaryKey" json:"id"`
	Platform          string    `gorm:"size:50;not null;index:idx_platform_auth_sku" json:"platform"`
	PlatformAuthID    uint      `gorm:"index:idx_platform_auth_sku;not null" json:"platform_auth_id"`
	PlatformSKU       string    `gorm:"size:100;not null;index:idx_platform_auth_sku" json:"platform_sku"`
	PlatformAccountID uint      `gorm:"index:idx_account_unique_code;not null;default:0" json:"platform_account_id"` // 关联 platform_auths.user_id
	UniqueCode        string    `gorm:"size:100;index:idx_account_unique_code" json:"unique_code"`                   // 产品唯一标识(offer_id)
	Name              string    `gorm:"size:500" json:"name"`
	Image             string    `gorm:"column:image;size:500" json:"image"` // 产品主图URL
	Stock             int       `gorm:"default:0" json:"stock"`
	Price             float64   `gorm:"type:decimal(10,2)" json:"price"`
	Currency          string    `gorm:"size:10" json:"currency"`
	Status            string    `gorm:"size:50" json:"status"`
	RawData           string    `gorm:"type:longtext" json:"-"` // 原始JSON数据
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`

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
	WID               uint      `gorm:"index:idx_mapping_unique,unique;not null;default:0" json:"wid"`                        // 仓库ID
	PlatformAccountID uint      `gorm:"index:idx_mapping_unique,unique;not null;default:0" json:"platform_account_id"`        // 授权账户ID (platform_auths.user_id)
	ProductID         uint      `gorm:"index:idx_mapping_unique,unique;index;not null" json:"product_id"`                     // 本地产品ID
	PlatformProductID uint      `gorm:"index:idx_mapping_unique,unique;index;not null" json:"platform_product_id"`            // 平台产品ID
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
