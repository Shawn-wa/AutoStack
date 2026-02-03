package shipping

import (
	"context"
	"errors"
	"math"

	shippingRepo "autostack/internal/repository/shipping"
)

var service *Service

// Service 运费模块服务
type Service struct {
	templateRepo     shippingRepo.ShippingTemplateRepository
	templateRuleRepo shippingRepo.ShippingTemplateRuleRepository
}

// InitService 初始化服务
func InitService(templateRepo shippingRepo.ShippingTemplateRepository, templateRuleRepo shippingRepo.ShippingTemplateRuleRepository) {
	service = &Service{
		templateRepo:     templateRepo,
		templateRuleRepo: templateRuleRepo,
	}
}

// ========== 运费模板服务 ==========

// CreateTemplate 创建运费模板
func (s *Service) CreateTemplate(req *CreateTemplateRequest) (*shippingRepo.ShippingTemplate, error) {
	ctx := context.Background()

	template := &shippingRepo.ShippingTemplate{
		Name:        req.Name,
		Carrier:     req.Carrier,
		FromRegion:  req.FromRegion,
		Description: req.Description,
		Status:      shippingRepo.TemplateStatusActive,
	}

	if err := s.templateRepo.Create(ctx, template); err != nil {
		return nil, err
	}

	// 创建规则
	if len(req.Rules) > 0 {
		var rules []shippingRepo.ShippingTemplateRule
		for _, r := range req.Rules {
			additionalUnit := r.AdditionalUnit
			if additionalUnit == 0 {
				additionalUnit = 100 // 默认100g
			}
			currency := r.Currency
			if currency == "" {
				currency = "CNY"
			}
			rules = append(rules, shippingRepo.ShippingTemplateRule{
				TemplateID:      template.ID,
				ToRegion:        r.ToRegion,
				MinWeight:       r.MinWeight,
				MaxWeight:       r.MaxWeight,
				FirstWeight:     r.FirstWeight,
				FirstPrice:      r.FirstPrice,
				AdditionalUnit:  additionalUnit,
				AdditionalPrice: r.AdditionalPrice,
				Currency:        currency,
				EstimatedDays:   r.EstimatedDays,
			})
		}
		if err := s.templateRuleRepo.BatchCreate(ctx, rules); err != nil {
			return nil, err
		}
	}

	return template, nil
}

// UpdateTemplate 更新运费模板
func (s *Service) UpdateTemplate(id uint, req *UpdateTemplateRequest) error {
	ctx := context.Background()

	template, err := s.templateRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if req.Name != "" {
		template.Name = req.Name
	}
	if req.Carrier != "" {
		template.Carrier = req.Carrier
	}
	if req.FromRegion != "" {
		template.FromRegion = req.FromRegion
	}
	if req.Description != "" {
		template.Description = req.Description
	}
	if req.Status != "" {
		template.Status = req.Status
	}

	return s.templateRepo.Update(ctx, template)
}

// DeleteTemplate 删除运费模板
func (s *Service) DeleteTemplate(id uint) error {
	ctx := context.Background()
	// 先删除规则
	if err := s.templateRuleRepo.DeleteByTemplateID(ctx, id); err != nil {
		return err
	}
	return s.templateRepo.Delete(ctx, id)
}

// GetTemplate 获取运费模板详情
func (s *Service) GetTemplate(id uint) (*shippingRepo.ShippingTemplate, error) {
	ctx := context.Background()
	return s.templateRepo.FindByIDWithRules(ctx, id)
}

// ListTemplates 获取运费模板列表
func (s *Service) ListTemplates(page, pageSize int, keyword, status string) ([]shippingRepo.ShippingTemplate, int64, error) {
	ctx := context.Background()
	return s.templateRepo.List(ctx, &shippingRepo.TemplateQuery{
		Page:     page,
		PageSize: pageSize,
		Keyword:  keyword,
		Status:   status,
	})
}

// ListAllTemplates 获取所有启用的运费模板
func (s *Service) ListAllTemplates() ([]shippingRepo.ShippingTemplate, error) {
	ctx := context.Background()
	return s.templateRepo.ListAll(ctx)
}

// ========== 运费规则服务 ==========

// CreateRule 创建运费规则
func (s *Service) CreateRule(templateID uint, req *CreateTemplateRuleRequest) (*shippingRepo.ShippingTemplateRule, error) {
	ctx := context.Background()

	// 检查模板是否存在
	if _, err := s.templateRepo.FindByID(ctx, templateID); err != nil {
		return nil, errors.New("模板不存在")
	}

	additionalUnit := req.AdditionalUnit
	if additionalUnit == 0 {
		additionalUnit = 100
	}
	currency := req.Currency
	if currency == "" {
		currency = "CNY"
	}

	rule := &shippingRepo.ShippingTemplateRule{
		TemplateID:      templateID,
		ToRegion:        req.ToRegion,
		MinWeight:       req.MinWeight,
		MaxWeight:       req.MaxWeight,
		FirstWeight:     req.FirstWeight,
		FirstPrice:      req.FirstPrice,
		AdditionalUnit:  additionalUnit,
		AdditionalPrice: req.AdditionalPrice,
		Currency:        currency,
		EstimatedDays:   req.EstimatedDays,
	}

	if err := s.templateRuleRepo.Create(ctx, rule); err != nil {
		return nil, err
	}

	return rule, nil
}

// UpdateRule 更新运费规则
func (s *Service) UpdateRule(id uint, req *UpdateTemplateRuleRequest) error {
	ctx := context.Background()

	rule, err := s.templateRuleRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if req.ToRegion != "" {
		rule.ToRegion = req.ToRegion
	}
	rule.MinWeight = req.MinWeight
	rule.MaxWeight = req.MaxWeight
	rule.FirstWeight = req.FirstWeight
	rule.FirstPrice = req.FirstPrice
	if req.AdditionalUnit > 0 {
		rule.AdditionalUnit = req.AdditionalUnit
	}
	rule.AdditionalPrice = req.AdditionalPrice
	if req.Currency != "" {
		rule.Currency = req.Currency
	}
	rule.EstimatedDays = req.EstimatedDays

	return s.templateRuleRepo.Update(ctx, rule)
}

// DeleteRule 删除运费规则
func (s *Service) DeleteRule(id uint) error {
	ctx := context.Background()
	return s.templateRuleRepo.Delete(ctx, id)
}

// GetRulesByTemplateID 获取模板的所有规则
func (s *Service) GetRulesByTemplateID(templateID uint) ([]shippingRepo.ShippingTemplateRule, error) {
	ctx := context.Background()
	return s.templateRuleRepo.FindByTemplateID(ctx, templateID)
}

// ========== 运费计算服务 ==========

// CalculateShipping 计算运费
// 计算公式: 首重费用 + ceil((重量 - 首重) / 续重单位) * 续重单价
func (s *Service) CalculateShipping(templateID uint, toRegion string, weight int) (*CalculateShippingResponse, error) {
	ctx := context.Background()

	// 获取模板
	template, err := s.templateRepo.FindByID(ctx, templateID)
	if err != nil {
		return nil, errors.New("模板不存在")
	}

	// 查找匹配的规则
	rule, err := s.templateRuleRepo.FindMatchingRule(ctx, templateID, toRegion, weight)
	if err != nil {
		return nil, errors.New("未找到匹配的运费规则")
	}

	// 计算运费
	var shippingFee float64
	if weight <= rule.FirstWeight {
		shippingFee = rule.FirstPrice
	} else {
		additionalWeight := weight - rule.FirstWeight
		additionalUnits := math.Ceil(float64(additionalWeight) / float64(rule.AdditionalUnit))
		shippingFee = rule.FirstPrice + additionalUnits*rule.AdditionalPrice
	}

	return &CalculateShippingResponse{
		TemplateID:    templateID,
		TemplateName:  template.Name,
		ToRegion:      toRegion,
		Weight:        weight,
		ShippingFee:   shippingFee,
		Currency:      rule.Currency,
		EstimatedDays: rule.EstimatedDays,
	}, nil
}

// BatchCalculateShipping 批量计算运费
func (s *Service) BatchCalculateShipping(items []CalculateShippingRequest) ([]CalculateShippingResponse, error) {
	var results []CalculateShippingResponse
	for _, item := range items {
		result, err := s.CalculateShipping(item.TemplateID, item.ToRegion, item.Weight)
		if err != nil {
			// 单个失败不影响其他计算，返回零值
			results = append(results, CalculateShippingResponse{
				TemplateID:  item.TemplateID,
				ToRegion:    item.ToRegion,
				Weight:      item.Weight,
				ShippingFee: 0,
				Currency:    "CNY",
			})
		} else {
			results = append(results, *result)
		}
	}
	return results, nil
}
