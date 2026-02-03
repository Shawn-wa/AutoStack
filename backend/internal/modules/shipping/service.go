package shipping

import (
	"context"
	"errors"
	"math"
	"time"

	shippingRepo "autostack/internal/repository/shipping"
)

var service *Service

// Service 运费模块服务
type Service struct {
	templateRepo                shippingRepo.ShippingTemplateRepository
	templateRuleRepo            shippingRepo.ShippingTemplateRuleRepository
	productShippingRepo         shippingRepo.ProductShippingTemplateRepository
	platformProductShippingRepo shippingRepo.PlatformProductShippingTemplateRepository
}

// InitService 初始化服务
func InitService(
	templateRepo shippingRepo.ShippingTemplateRepository,
	templateRuleRepo shippingRepo.ShippingTemplateRuleRepository,
	productShippingRepo shippingRepo.ProductShippingTemplateRepository,
	platformProductShippingRepo shippingRepo.PlatformProductShippingTemplateRepository,
) {
	service = &Service{
		templateRepo:                templateRepo,
		templateRuleRepo:            templateRuleRepo,
		productShippingRepo:         productShippingRepo,
		platformProductShippingRepo: platformProductShippingRepo,
	}
}

// GetService 获取服务实例
func GetService() *Service {
	return service
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

// ========== 本地产品运费模版关联服务 ==========

// BindProductShippingTemplate 绑定本地产品运费模版
func (s *Service) BindProductShippingTemplate(req *BindProductShippingTemplateRequest) (*shippingRepo.ProductShippingTemplate, error) {
	ctx := context.Background()

	// 检查模板是否存在
	if _, err := s.templateRepo.FindByID(ctx, req.ShippingTemplateID); err != nil {
		return nil, errors.New("运费模版不存在")
	}

	pst := &shippingRepo.ProductShippingTemplate{
		ProductID:          req.ProductID,
		ShippingTemplateID: req.ShippingTemplateID,
		IsDefault:          req.IsDefault,
		SortOrder:          req.SortOrder,
		Status:             shippingRepo.TemplateStatusActive,
	}

	if err := s.productShippingRepo.Create(ctx, pst); err != nil {
		return nil, err
	}

	// 如果设置为默认，更新其他记录
	if req.IsDefault {
		if err := s.productShippingRepo.SetDefault(ctx, req.ProductID, req.ShippingTemplateID); err != nil {
			return nil, err
		}
	}

	return pst, nil
}

// UnbindProductShippingTemplate 解绑本地产品运费模版
func (s *Service) UnbindProductShippingTemplate(id uint) error {
	ctx := context.Background()
	return s.productShippingRepo.Delete(ctx, id)
}

// GetProductShippingTemplates 获取本地产品的运费模版列表
func (s *Service) GetProductShippingTemplates(productID uint) ([]shippingRepo.ProductShippingTemplate, error) {
	ctx := context.Background()
	return s.productShippingRepo.FindByProductID(ctx, productID)
}

// GetDefaultProductShippingTemplate 获取本地产品的默认运费模版
func (s *Service) GetDefaultProductShippingTemplate(productID uint) (*shippingRepo.ProductShippingTemplate, error) {
	ctx := context.Background()
	return s.productShippingRepo.FindDefaultByProductID(ctx, productID)
}

// SetProductDefaultShippingTemplate 设置本地产品的默认运费模版
func (s *Service) SetProductDefaultShippingTemplate(productID uint, shippingTemplateID uint) error {
	ctx := context.Background()
	return s.productShippingRepo.SetDefault(ctx, productID, shippingTemplateID)
}

// ========== 平台产品运费模版关联服务 ==========

// BindPlatformProductShippingTemplate 绑定平台产品运费模版
func (s *Service) BindPlatformProductShippingTemplate(req *BindPlatformProductShippingTemplateRequest) (*shippingRepo.PlatformProductShippingTemplate, error) {
	ctx := context.Background()

	// 检查模板是否存在
	if _, err := s.templateRepo.FindByID(ctx, req.ShippingTemplateID); err != nil {
		return nil, errors.New("运费模版不存在")
	}

	ppst := &shippingRepo.PlatformProductShippingTemplate{
		PlatformProductID:  req.PlatformProductID,
		ShippingTemplateID: req.ShippingTemplateID,
		IsDefault:          req.IsDefault,
		SortOrder:          req.SortOrder,
		Status:             shippingRepo.TemplateStatusActive,
	}

	if err := s.platformProductShippingRepo.Create(ctx, ppst); err != nil {
		return nil, err
	}

	// 如果设置为默认，更新其他记录
	if req.IsDefault {
		if err := s.platformProductShippingRepo.SetDefault(ctx, req.PlatformProductID, req.ShippingTemplateID); err != nil {
			return nil, err
		}
	}

	return ppst, nil
}

// UnbindPlatformProductShippingTemplate 解绑平台产品运费模版
func (s *Service) UnbindPlatformProductShippingTemplate(id uint) error {
	ctx := context.Background()
	return s.platformProductShippingRepo.Delete(ctx, id)
}

// GetPlatformProductShippingTemplates 获取平台产品的运费模版列表
func (s *Service) GetPlatformProductShippingTemplates(platformProductID uint) ([]shippingRepo.PlatformProductShippingTemplate, error) {
	ctx := context.Background()
	return s.platformProductShippingRepo.FindByPlatformProductID(ctx, platformProductID)
}

// GetDefaultPlatformProductShippingTemplate 获取平台产品的默认运费模版
func (s *Service) GetDefaultPlatformProductShippingTemplate(platformProductID uint) (*shippingRepo.PlatformProductShippingTemplate, error) {
	ctx := context.Background()
	return s.platformProductShippingRepo.FindDefaultByPlatformProductID(ctx, platformProductID)
}

// SetPlatformProductDefaultShippingTemplate 设置平台产品的默认运费模版
func (s *Service) SetPlatformProductDefaultShippingTemplate(platformProductID uint, shippingTemplateID uint) error {
	ctx := context.Background()
	return s.platformProductShippingRepo.SetDefault(ctx, platformProductID, shippingTemplateID)
}

// ========== 订单运费估算服务 ==========

// ShippingEstimateInput 运费估算输入（供订单模块调用）
type ShippingEstimateInput struct {
	PlatformProductID uint    // 平台产品ID
	ProductID         uint    // 本地产品ID
	Weight            float64 // 产品重量(kg)
	Quantity          int     // 数量
	ToRegion          string  // 收货区域/国家
}

// ShippingEstimateResult 运费估算结果
type ShippingEstimateResult struct {
	TemplateID       uint    // 使用的模版ID
	TemplateName     string  // 模版名称
	ShippingFee      float64 // 单件运费
	TotalShippingFee float64 // 总运费
	Currency         string  // 货币
	EstimatedDays    int     // 预估时效
	Source           string  // 来源: platform_product/product/none
}

// EstimateShippingForItem 估算单个商品的运费
// 优先使用平台产品的运费模版，如果没有则使用本地产品的
func (s *Service) EstimateShippingForItem(input *ShippingEstimateInput) (*ShippingEstimateResult, error) {
	ctx := context.Background()

	var templateID uint
	var source string

	// 1. 优先查找平台产品的运费模版
	if input.PlatformProductID > 0 {
		ppst, err := s.platformProductShippingRepo.FindDefaultByPlatformProductID(ctx, input.PlatformProductID)
		if err == nil && ppst != nil && ppst.ShippingTemplate != nil {
			templateID = ppst.ShippingTemplateID
			source = "platform_product"
		}
	}

	// 2. 如果平台产品没有配置，使用本地产品的运费模版
	if templateID == 0 && input.ProductID > 0 {
		pst, err := s.productShippingRepo.FindDefaultByProductID(ctx, input.ProductID)
		if err == nil && pst != nil && pst.ShippingTemplate != nil {
			templateID = pst.ShippingTemplateID
			source = "product"
		}
	}

	// 3. 如果都没有配置，返回空结果
	if templateID == 0 {
		return &ShippingEstimateResult{
			Source: "none",
		}, nil
	}

	// 4. 计算运费（重量转换：kg -> g）
	weightInGrams := int(input.Weight * 1000)
	if weightInGrams <= 0 {
		weightInGrams = 1 // 最小1g
	}

	result, err := s.CalculateShipping(templateID, input.ToRegion, weightInGrams)
	if err != nil {
		return &ShippingEstimateResult{
			TemplateID: templateID,
			Source:     source,
		}, nil
	}

	return &ShippingEstimateResult{
		TemplateID:       templateID,
		TemplateName:     result.TemplateName,
		ShippingFee:      result.ShippingFee,
		TotalShippingFee: result.ShippingFee * float64(input.Quantity),
		Currency:         result.Currency,
		EstimatedDays:    result.EstimatedDays,
		Source:           source,
	}, nil
}

// ConvertProductShippingTemplateToResponse 转换ProductShippingTemplate为响应DTO
func ConvertProductShippingTemplateToResponse(pst *shippingRepo.ProductShippingTemplate) *ProductShippingTemplateResponse {
	resp := &ProductShippingTemplateResponse{
		ID:                 pst.ID,
		ProductID:          pst.ProductID,
		ShippingTemplateID: pst.ShippingTemplateID,
		IsDefault:          pst.IsDefault,
		SortOrder:          pst.SortOrder,
		Status:             pst.Status,
		CreatedAt:          pst.CreatedAt.Format(time.RFC3339),
	}
	if pst.ShippingTemplate != nil {
		resp.TemplateName = pst.ShippingTemplate.Name
		resp.Carrier = pst.ShippingTemplate.Carrier
	}
	return resp
}

// ConvertPlatformProductShippingTemplateToResponse 转换PlatformProductShippingTemplate为响应DTO
func ConvertPlatformProductShippingTemplateToResponse(ppst *shippingRepo.PlatformProductShippingTemplate) *PlatformProductShippingTemplateResponse {
	resp := &PlatformProductShippingTemplateResponse{
		ID:                 ppst.ID,
		PlatformProductID:  ppst.PlatformProductID,
		ShippingTemplateID: ppst.ShippingTemplateID,
		IsDefault:          ppst.IsDefault,
		SortOrder:          ppst.SortOrder,
		Status:             ppst.Status,
		CreatedAt:          ppst.CreatedAt.Format(time.RFC3339),
	}
	if ppst.ShippingTemplate != nil {
		resp.TemplateName = ppst.ShippingTemplate.Name
		resp.Carrier = ppst.ShippingTemplate.Carrier
	}
	return resp
}
