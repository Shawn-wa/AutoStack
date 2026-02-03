package shipping

import (
	"context"
)

// ShippingTemplateRepository 运费模板仓库接口
type ShippingTemplateRepository interface {
	Create(ctx context.Context, template *ShippingTemplate) error
	Update(ctx context.Context, template *ShippingTemplate) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*ShippingTemplate, error)
	FindByIDWithRules(ctx context.Context, id uint) (*ShippingTemplate, error)
	List(ctx context.Context, query *TemplateQuery) ([]ShippingTemplate, int64, error)
	ListAll(ctx context.Context) ([]ShippingTemplate, error)
}

// ShippingTemplateRuleRepository 运费模板规则仓库接口
type ShippingTemplateRuleRepository interface {
	Create(ctx context.Context, rule *ShippingTemplateRule) error
	BatchCreate(ctx context.Context, rules []ShippingTemplateRule) error
	Update(ctx context.Context, rule *ShippingTemplateRule) error
	Delete(ctx context.Context, id uint) error
	DeleteByTemplateID(ctx context.Context, templateID uint) error
	FindByID(ctx context.Context, id uint) (*ShippingTemplateRule, error)
	FindByTemplateID(ctx context.Context, templateID uint) ([]ShippingTemplateRule, error)
	FindMatchingRule(ctx context.Context, templateID uint, toRegion string, weight int) (*ShippingTemplateRule, error)
}

// ProductShippingTemplateRepository 本地产品运费模版关联仓库接口
type ProductShippingTemplateRepository interface {
	Create(ctx context.Context, pst *ProductShippingTemplate) error
	Update(ctx context.Context, pst *ProductShippingTemplate) error
	Delete(ctx context.Context, id uint) error
	DeleteByProductID(ctx context.Context, productID uint) error
	FindByID(ctx context.Context, id uint) (*ProductShippingTemplate, error)
	FindByProductID(ctx context.Context, productID uint) ([]ProductShippingTemplate, error)
	FindDefaultByProductID(ctx context.Context, productID uint) (*ProductShippingTemplate, error)
	SetDefault(ctx context.Context, productID uint, shippingTemplateID uint) error
}

// PlatformProductShippingTemplateRepository 平台产品运费模版关联仓库接口
type PlatformProductShippingTemplateRepository interface {
	Create(ctx context.Context, ppst *PlatformProductShippingTemplate) error
	Update(ctx context.Context, ppst *PlatformProductShippingTemplate) error
	Delete(ctx context.Context, id uint) error
	DeleteByPlatformProductID(ctx context.Context, platformProductID uint) error
	FindByID(ctx context.Context, id uint) (*PlatformProductShippingTemplate, error)
	FindByPlatformProductID(ctx context.Context, platformProductID uint) ([]PlatformProductShippingTemplate, error)
	FindDefaultByPlatformProductID(ctx context.Context, platformProductID uint) (*PlatformProductShippingTemplate, error)
	SetDefault(ctx context.Context, platformProductID uint, shippingTemplateID uint) error
}
