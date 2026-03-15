package company

import "time"

// Company 企业模型
type Company struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Status    int       `gorm:"default:1" json:"status"` // 1=活跃, 0=禁用
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Company) TableName() string {
	return "companies"
}

// 企业状态常量
const (
	StatusActive   = 1
	StatusDisabled = 0
)

// IsActive 判断企业是否激活
func (c *Company) IsActive() bool {
	return c.Status == StatusActive
}

const (
	// CompanyIDMin/Max 约束 company_id 固定为 11 位且以 90 开头
	CompanyIDMin uint = 90000000000
	CompanyIDMax uint = 90999999999
)

// IsStandardCompanyID 判断是否为新规范 company_id（11位，90开头）
func IsStandardCompanyID(id uint) bool {
	return id >= CompanyIDMin && id <= CompanyIDMax
}
