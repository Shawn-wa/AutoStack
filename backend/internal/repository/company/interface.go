package company

import (
	"context"
)

// CompanyRepository 企业仓储接口
type CompanyRepository interface {
	// FindByID 根据ID查找企业
	FindByID(ctx context.Context, id uint) (*Company, error)
	// Create 创建企业
	Create(ctx context.Context, company *Company) error
	// Update 更新企业
	Update(ctx context.Context, company *Company) error
	// UpdateFields 更新指定字段
	UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error
	// List 查询企业列表
	List(ctx context.Context) ([]Company, error)
}
