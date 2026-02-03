package shipping

import (
	"time"
)

// ShippingTemplate 运费模板
type ShippingTemplate struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`                // 模板名称
	Carrier     string    `gorm:"size:100" json:"carrier"`                      // 物流商名称
	FromRegion  string    `gorm:"size:100" json:"from_region"`                  // 发货区域
	Description string    `gorm:"size:500" json:"description"`                  // 描述
	Status      string    `gorm:"size:20;default:'active';index" json:"status"` // 状态: active/inactive
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 关联
	Rules []ShippingTemplateRule `gorm:"foreignKey:TemplateID" json:"rules,omitempty"`
}

// TableName 指定表名
func (ShippingTemplate) TableName() string {
	return "shipping_templates"
}

// ShippingTemplateRule 运费模板规则
type ShippingTemplateRule struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	TemplateID      uint      `gorm:"index;not null" json:"template_id"`                    // 模板ID
	ToRegion        string    `gorm:"size:100;not null" json:"to_region"`                   // 收货区域/国家
	MinWeight       int       `gorm:"default:0" json:"min_weight"`                          // 最小重量(g)
	MaxWeight       int       `gorm:"default:0" json:"max_weight"`                          // 最大重量(g)，0表示不限
	FirstWeight     int       `gorm:"default:0" json:"first_weight"`                        // 首重(g)
	FirstPrice      float64   `gorm:"type:decimal(10,2);default:0" json:"first_price"`      // 首重费用
	AdditionalUnit  int       `gorm:"default:100" json:"additional_unit"`                   // 续重单位(g)，默认100g
	AdditionalPrice float64   `gorm:"type:decimal(10,2);default:0" json:"additional_price"` // 续重单价
	Currency        string    `gorm:"size:10;default:'CNY'" json:"currency"`                // 货币
	EstimatedDays   int       `gorm:"default:0" json:"estimated_days"`                      // 预估时效(天)
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName 指定表名
func (ShippingTemplateRule) TableName() string {
	return "shipping_template_rules"
}

// 模板状态枚举
const (
	TemplateStatusActive   = "active"   // 启用
	TemplateStatusInactive = "inactive" // 停用
)
