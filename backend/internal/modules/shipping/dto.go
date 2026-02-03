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
