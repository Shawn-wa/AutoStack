package shipping

import (
	"context"

	"gorm.io/gorm"

	"autostack/internal/repository"
)

// ========== ShippingTemplate Repository ==========

type gormShippingTemplateRepository struct {
	db *gorm.DB
}

func NewShippingTemplateRepository(db *gorm.DB) ShippingTemplateRepository {
	return &gormShippingTemplateRepository{db: db}
}

func (r *gormShippingTemplateRepository) getDB(ctx context.Context) *gorm.DB {
	return repository.GetDB(ctx, r.db)
}

func (r *gormShippingTemplateRepository) Create(ctx context.Context, template *ShippingTemplate) error {
	return r.getDB(ctx).Create(template).Error
}

func (r *gormShippingTemplateRepository) Update(ctx context.Context, template *ShippingTemplate) error {
	return r.getDB(ctx).Save(template).Error
}

func (r *gormShippingTemplateRepository) Delete(ctx context.Context, id uint) error {
	return r.getDB(ctx).Delete(&ShippingTemplate{}, id).Error
}

func (r *gormShippingTemplateRepository) FindByID(ctx context.Context, id uint) (*ShippingTemplate, error) {
	var template ShippingTemplate
	if err := r.getDB(ctx).First(&template, id).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *gormShippingTemplateRepository) FindByIDWithRules(ctx context.Context, id uint) (*ShippingTemplate, error) {
	var template ShippingTemplate
	if err := r.getDB(ctx).Preload("Rules").First(&template, id).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *gormShippingTemplateRepository) List(ctx context.Context, query *TemplateQuery) ([]ShippingTemplate, int64, error) {
	var templates []ShippingTemplate
	var total int64
	db := r.getDB(ctx)

	q := db.Model(&ShippingTemplate{})
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		q = q.Where("name LIKE ? OR carrier LIKE ?", like, like)
	}
	if query.Status != "" {
		q = q.Where("status = ?", query.Status)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	if err := q.Preload("Rules").Order("id DESC").Offset(offset).Limit(query.PageSize).Find(&templates).Error; err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

func (r *gormShippingTemplateRepository) ListAll(ctx context.Context) ([]ShippingTemplate, error) {
	var templates []ShippingTemplate
	if err := r.getDB(ctx).Where("status = ?", TemplateStatusActive).Order("name ASC").Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

// ========== ShippingTemplateRule Repository ==========

type gormShippingTemplateRuleRepository struct {
	db *gorm.DB
}

func NewShippingTemplateRuleRepository(db *gorm.DB) ShippingTemplateRuleRepository {
	return &gormShippingTemplateRuleRepository{db: db}
}

func (r *gormShippingTemplateRuleRepository) getDB(ctx context.Context) *gorm.DB {
	return repository.GetDB(ctx, r.db)
}

func (r *gormShippingTemplateRuleRepository) Create(ctx context.Context, rule *ShippingTemplateRule) error {
	return r.getDB(ctx).Create(rule).Error
}

func (r *gormShippingTemplateRuleRepository) BatchCreate(ctx context.Context, rules []ShippingTemplateRule) error {
	if len(rules) == 0 {
		return nil
	}
	return r.getDB(ctx).Create(&rules).Error
}

func (r *gormShippingTemplateRuleRepository) Update(ctx context.Context, rule *ShippingTemplateRule) error {
	return r.getDB(ctx).Save(rule).Error
}

func (r *gormShippingTemplateRuleRepository) Delete(ctx context.Context, id uint) error {
	return r.getDB(ctx).Delete(&ShippingTemplateRule{}, id).Error
}

func (r *gormShippingTemplateRuleRepository) DeleteByTemplateID(ctx context.Context, templateID uint) error {
	return r.getDB(ctx).Where("template_id = ?", templateID).Delete(&ShippingTemplateRule{}).Error
}

func (r *gormShippingTemplateRuleRepository) FindByID(ctx context.Context, id uint) (*ShippingTemplateRule, error) {
	var rule ShippingTemplateRule
	if err := r.getDB(ctx).First(&rule, id).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *gormShippingTemplateRuleRepository) FindByTemplateID(ctx context.Context, templateID uint) ([]ShippingTemplateRule, error) {
	var rules []ShippingTemplateRule
	if err := r.getDB(ctx).Where("template_id = ?", templateID).Order("min_weight ASC").Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

func (r *gormShippingTemplateRuleRepository) FindMatchingRule(ctx context.Context, templateID uint, toRegion string, weight int) (*ShippingTemplateRule, error) {
	var rule ShippingTemplateRule

	// 先精确匹配区域，再尝试通配符匹配
	q := r.getDB(ctx).Where("template_id = ?", templateID).
		Where("(to_region = ? OR to_region = '*')", toRegion).
		Where("min_weight <= ?", weight).
		Where("max_weight = 0 OR max_weight >= ?", weight).
		Order("CASE WHEN to_region = '*' THEN 1 ELSE 0 END, min_weight DESC")

	if err := q.First(&rule).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

// ========== ProductShippingTemplate Repository ==========

type gormProductShippingTemplateRepository struct {
	db *gorm.DB
}

func NewProductShippingTemplateRepository(db *gorm.DB) ProductShippingTemplateRepository {
	return &gormProductShippingTemplateRepository{db: db}
}

func (r *gormProductShippingTemplateRepository) getDB(ctx context.Context) *gorm.DB {
	return repository.GetDB(ctx, r.db)
}

func (r *gormProductShippingTemplateRepository) Create(ctx context.Context, pst *ProductShippingTemplate) error {
	return r.getDB(ctx).Create(pst).Error
}

func (r *gormProductShippingTemplateRepository) Update(ctx context.Context, pst *ProductShippingTemplate) error {
	return r.getDB(ctx).Save(pst).Error
}

func (r *gormProductShippingTemplateRepository) Delete(ctx context.Context, id uint) error {
	return r.getDB(ctx).Delete(&ProductShippingTemplate{}, id).Error
}

func (r *gormProductShippingTemplateRepository) DeleteByProductID(ctx context.Context, productID uint) error {
	return r.getDB(ctx).Where("product_id = ?", productID).Delete(&ProductShippingTemplate{}).Error
}

func (r *gormProductShippingTemplateRepository) FindByID(ctx context.Context, id uint) (*ProductShippingTemplate, error) {
	var pst ProductShippingTemplate
	if err := r.getDB(ctx).Preload("ShippingTemplate").First(&pst, id).Error; err != nil {
		return nil, err
	}
	return &pst, nil
}

func (r *gormProductShippingTemplateRepository) FindByProductID(ctx context.Context, productID uint) ([]ProductShippingTemplate, error) {
	var list []ProductShippingTemplate
	if err := r.getDB(ctx).Preload("ShippingTemplate").
		Where("product_id = ? AND status = ?", productID, TemplateStatusActive).
		Order("is_default DESC, sort_order ASC, id ASC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *gormProductShippingTemplateRepository) FindDefaultByProductID(ctx context.Context, productID uint) (*ProductShippingTemplate, error) {
	var pst ProductShippingTemplate
	// 优先查找标记为默认的，否则取排序最前的
	if err := r.getDB(ctx).Preload("ShippingTemplate").
		Where("product_id = ? AND status = ?", productID, TemplateStatusActive).
		Order("is_default DESC, sort_order ASC, id ASC").
		First(&pst).Error; err != nil {
		return nil, err
	}
	return &pst, nil
}

func (r *gormProductShippingTemplateRepository) SetDefault(ctx context.Context, productID uint, shippingTemplateID uint) error {
	// 先取消所有默认
	if err := r.getDB(ctx).Model(&ProductShippingTemplate{}).
		Where("product_id = ?", productID).
		Update("is_default", false).Error; err != nil {
		return err
	}
	// 设置新的默认
	return r.getDB(ctx).Model(&ProductShippingTemplate{}).
		Where("product_id = ? AND shipping_template_id = ?", productID, shippingTemplateID).
		Update("is_default", true).Error
}

// ========== PlatformProductShippingTemplate Repository ==========

type gormPlatformProductShippingTemplateRepository struct {
	db *gorm.DB
}

func NewPlatformProductShippingTemplateRepository(db *gorm.DB) PlatformProductShippingTemplateRepository {
	return &gormPlatformProductShippingTemplateRepository{db: db}
}

func (r *gormPlatformProductShippingTemplateRepository) getDB(ctx context.Context) *gorm.DB {
	return repository.GetDB(ctx, r.db)
}

func (r *gormPlatformProductShippingTemplateRepository) Create(ctx context.Context, ppst *PlatformProductShippingTemplate) error {
	return r.getDB(ctx).Create(ppst).Error
}

func (r *gormPlatformProductShippingTemplateRepository) Update(ctx context.Context, ppst *PlatformProductShippingTemplate) error {
	return r.getDB(ctx).Save(ppst).Error
}

func (r *gormPlatformProductShippingTemplateRepository) Delete(ctx context.Context, id uint) error {
	return r.getDB(ctx).Delete(&PlatformProductShippingTemplate{}, id).Error
}

func (r *gormPlatformProductShippingTemplateRepository) DeleteByPlatformProductID(ctx context.Context, platformProductID uint) error {
	return r.getDB(ctx).Where("platform_product_id = ?", platformProductID).Delete(&PlatformProductShippingTemplate{}).Error
}

func (r *gormPlatformProductShippingTemplateRepository) FindByID(ctx context.Context, id uint) (*PlatformProductShippingTemplate, error) {
	var ppst PlatformProductShippingTemplate
	if err := r.getDB(ctx).Preload("ShippingTemplate").First(&ppst, id).Error; err != nil {
		return nil, err
	}
	return &ppst, nil
}

func (r *gormPlatformProductShippingTemplateRepository) FindByPlatformProductID(ctx context.Context, platformProductID uint) ([]PlatformProductShippingTemplate, error) {
	var list []PlatformProductShippingTemplate
	if err := r.getDB(ctx).Preload("ShippingTemplate").
		Where("platform_product_id = ? AND status = ?", platformProductID, TemplateStatusActive).
		Order("is_default DESC, sort_order ASC, id ASC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *gormPlatformProductShippingTemplateRepository) FindDefaultByPlatformProductID(ctx context.Context, platformProductID uint) (*PlatformProductShippingTemplate, error) {
	var ppst PlatformProductShippingTemplate
	// 优先查找标记为默认的，否则取排序最前的
	if err := r.getDB(ctx).Preload("ShippingTemplate").
		Where("platform_product_id = ? AND status = ?", platformProductID, TemplateStatusActive).
		Order("is_default DESC, sort_order ASC, id ASC").
		First(&ppst).Error; err != nil {
		return nil, err
	}
	return &ppst, nil
}

func (r *gormPlatformProductShippingTemplateRepository) SetDefault(ctx context.Context, platformProductID uint, shippingTemplateID uint) error {
	// 先取消所有默认
	if err := r.getDB(ctx).Model(&PlatformProductShippingTemplate{}).
		Where("platform_product_id = ?", platformProductID).
		Update("is_default", false).Error; err != nil {
		return err
	}
	// 设置新的默认
	return r.getDB(ctx).Model(&PlatformProductShippingTemplate{}).
		Where("platform_product_id = ? AND shipping_template_id = ?", platformProductID, shippingTemplateID).
		Update("is_default", true).Error
}
