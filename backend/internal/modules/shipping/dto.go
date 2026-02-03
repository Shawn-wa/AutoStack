package shipping

// ========== 运费模板 DTO ==========

// CreateTemplateRequest 创建运费模板请求
type CreateTemplateRequest struct {
	Name        string                      `json:"name" binding:"required"`
	Carrier     string                      `json:"carrier"`
	FromRegion  string                      `json:"from_region"`
	Description string                      `json:"description"`
	Rules       []CreateTemplateRuleRequest `json:"rules"`
}

// UpdateTemplateRequest 更新运费模板请求
type UpdateTemplateRequest struct {
	Name        string `json:"name"`
	Carrier     string `json:"carrier"`
	FromRegion  string `json:"from_region"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

// TemplateResponse 运费模板响应
type TemplateResponse struct {
	ID          uint                   `json:"id"`
	Name        string                 `json:"name"`
	Carrier     string                 `json:"carrier"`
	FromRegion  string                 `json:"from_region"`
	Description string                 `json:"description"`
	Status      string                 `json:"status"`
	RuleCount   int                    `json:"rule_count"`
	Rules       []TemplateRuleResponse `json:"rules,omitempty"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
}

// TemplateListResponse 运费模板列表响应
type TemplateListResponse struct {
	List     []TemplateResponse `json:"list"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
}

// TemplateOptionResponse 模板选项响应（用于下拉选择）
type TemplateOptionResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// ========== 运费规则 DTO ==========

// CreateTemplateRuleRequest 创建运费规则请求
type CreateTemplateRuleRequest struct {
	ToRegion        string  `json:"to_region" binding:"required"`
	MinWeight       int     `json:"min_weight"`
	MaxWeight       int     `json:"max_weight"`
	FirstWeight     int     `json:"first_weight"`
	FirstPrice      float64 `json:"first_price"`
	AdditionalUnit  int     `json:"additional_unit"`
	AdditionalPrice float64 `json:"additional_price"`
	Currency        string  `json:"currency"`
	EstimatedDays   int     `json:"estimated_days"`
}

// UpdateTemplateRuleRequest 更新运费规则请求
type UpdateTemplateRuleRequest struct {
	ToRegion        string  `json:"to_region"`
	MinWeight       int     `json:"min_weight"`
	MaxWeight       int     `json:"max_weight"`
	FirstWeight     int     `json:"first_weight"`
	FirstPrice      float64 `json:"first_price"`
	AdditionalUnit  int     `json:"additional_unit"`
	AdditionalPrice float64 `json:"additional_price"`
	Currency        string  `json:"currency"`
	EstimatedDays   int     `json:"estimated_days"`
}

// TemplateRuleResponse 运费规则响应
type TemplateRuleResponse struct {
	ID              uint    `json:"id"`
	TemplateID      uint    `json:"template_id"`
	ToRegion        string  `json:"to_region"`
	MinWeight       int     `json:"min_weight"`
	MaxWeight       int     `json:"max_weight"`
	FirstWeight     int     `json:"first_weight"`
	FirstPrice      float64 `json:"first_price"`
	AdditionalUnit  int     `json:"additional_unit"`
	AdditionalPrice float64 `json:"additional_price"`
	Currency        string  `json:"currency"`
	EstimatedDays   int     `json:"estimated_days"`
	CreatedAt       string  `json:"created_at"`
}

// ========== 运费计算 DTO ==========

// CalculateShippingRequest 计算运费请求
type CalculateShippingRequest struct {
	TemplateID uint   `json:"template_id" binding:"required"`
	ToRegion   string `json:"to_region" binding:"required"`
	Weight     int    `json:"weight" binding:"required"` // 重量(g)
}

// CalculateShippingResponse 计算运费响应
type CalculateShippingResponse struct {
	TemplateID    uint    `json:"template_id"`
	TemplateName  string  `json:"template_name"`
	ToRegion      string  `json:"to_region"`
	Weight        int     `json:"weight"`
	ShippingFee   float64 `json:"shipping_fee"`
	Currency      string  `json:"currency"`
	EstimatedDays int     `json:"estimated_days"`
}

// BatchCalculateRequest 批量计算运费请求
type BatchCalculateRequest struct {
	Items []CalculateShippingRequest `json:"items" binding:"required"`
}

// BatchCalculateResponse 批量计算运费响应
type BatchCalculateResponse struct {
	Results []CalculateShippingResponse `json:"results"`
}

// ========== 产品运费模版关联 DTO ==========

// BindProductShippingTemplateRequest 绑定本地产品运费模版请求
type BindProductShippingTemplateRequest struct {
	ProductID          uint `json:"product_id" binding:"required"`
	ShippingTemplateID uint `json:"shipping_template_id" binding:"required"`
	IsDefault          bool `json:"is_default"`
	SortOrder          int  `json:"sort_order"`
}

// BindPlatformProductShippingTemplateRequest 绑定平台产品运费模版请求
type BindPlatformProductShippingTemplateRequest struct {
	PlatformProductID  uint `json:"platform_product_id" binding:"required"`
	ShippingTemplateID uint `json:"shipping_template_id" binding:"required"`
	IsDefault          bool `json:"is_default"`
	SortOrder          int  `json:"sort_order"`
}

// SetDefaultShippingTemplateRequest 设置默认运费模版请求
type SetDefaultShippingTemplateRequest struct {
	ShippingTemplateID uint `json:"shipping_template_id" binding:"required"`
}

// ProductShippingTemplateResponse 产品运费模版关联响应
type ProductShippingTemplateResponse struct {
	ID                 uint   `json:"id"`
	ProductID          uint   `json:"product_id"`
	ShippingTemplateID uint   `json:"shipping_template_id"`
	TemplateName       string `json:"template_name"`
	Carrier            string `json:"carrier"`
	IsDefault          bool   `json:"is_default"`
	SortOrder          int    `json:"sort_order"`
	Status             string `json:"status"`
	CreatedAt          string `json:"created_at"`
}

// PlatformProductShippingTemplateResponse 平台产品运费模版关联响应
type PlatformProductShippingTemplateResponse struct {
	ID                 uint   `json:"id"`
	PlatformProductID  uint   `json:"platform_product_id"`
	ShippingTemplateID uint   `json:"shipping_template_id"`
	TemplateName       string `json:"template_name"`
	Carrier            string `json:"carrier"`
	IsDefault          bool   `json:"is_default"`
	SortOrder          int    `json:"sort_order"`
	Status             string `json:"status"`
	CreatedAt          string `json:"created_at"`
}

// ========== 订单运费估算 DTO ==========

// EstimateOrderShippingRequest 估算订单运费请求
type EstimateOrderShippingRequest struct {
	OrderID uint `json:"order_id" binding:"required"`
}

// EstimateOrderShippingResponse 估算订单运费响应
type EstimateOrderShippingResponse struct {
	OrderID          uint                        `json:"order_id"`
	TotalShippingFee float64                     `json:"total_shipping_fee"`
	Currency         string                      `json:"currency"`
	Items            []OrderItemShippingEstimate `json:"items"`
	EstimatedAt      string                      `json:"estimated_at"`
}

// OrderItemShippingEstimate 订单项运费估算
type OrderItemShippingEstimate struct {
	OrderItemID      uint    `json:"order_item_id"`
	PlatformSku      string  `json:"platform_sku"`
	LocalSku         string  `json:"local_sku"`
	Quantity         int     `json:"quantity"`
	Weight           int     `json:"weight"`             // 单件重量(g)
	TotalWeight      int     `json:"total_weight"`       // 总重量(g)
	ShippingFee      float64 `json:"shipping_fee"`       // 单件运费
	TotalShippingFee float64 `json:"total_shipping_fee"` // 总运费
	Currency         string  `json:"currency"`
	TemplateID       uint    `json:"template_id"`
	TemplateName     string  `json:"template_name"`
	EstimatedDays    int     `json:"estimated_days"`
	Source           string  `json:"source"` // 运费模版来源: platform_product/product/none
}
