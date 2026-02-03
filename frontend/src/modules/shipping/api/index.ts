import request from '@/commonBase/api/request'

// ========== 类型定义 ==========

// 运费模板
export interface ShippingTemplate {
  id: number
  name: string
  carrier: string
  from_region: string
  description: string
  status: string
  rule_count: number
  rules?: ShippingTemplateRule[]
  created_at: string
  updated_at: string
}

// 运费规则
export interface ShippingTemplateRule {
  id: number
  template_id: number
  to_region: string
  min_weight: number
  max_weight: number
  first_weight: number
  first_price: number
  additional_unit: number
  additional_price: number
  currency: string
  estimated_days: number
  created_at: string
}

// 模板选项（下拉框）
export interface TemplateOption {
  id: number
  name: string
}

// 创建模板请求
export interface CreateTemplateRequest {
  name: string
  carrier?: string
  from_region?: string
  description?: string
  rules?: CreateRuleRequest[]
}

// 更新模板请求
export interface UpdateTemplateRequest {
  name?: string
  carrier?: string
  from_region?: string
  description?: string
  status?: string
}

// 创建规则请求
export interface CreateRuleRequest {
  to_region: string
  min_weight?: number
  max_weight?: number
  first_weight?: number
  first_price?: number
  additional_unit?: number
  additional_price?: number
  currency?: string
  estimated_days?: number
}

// 更新规则请求
export interface UpdateRuleRequest {
  to_region?: string
  min_weight?: number
  max_weight?: number
  first_weight?: number
  first_price?: number
  additional_unit?: number
  additional_price?: number
  currency?: string
  estimated_days?: number
}

// 计算运费请求
export interface CalculateShippingRequest {
  template_id: number
  to_region: string
  weight: number
}

// 计算运费响应
export interface CalculateShippingResponse {
  template_id: number
  template_name: string
  to_region: string
  weight: number
  shipping_fee: number
  currency: string
  estimated_days: number
}

// 列表响应
export interface TemplateListResponse {
  list: ShippingTemplate[]
  total: number
  page: number
  page_size: number
}

// 列表查询参数
export interface ListTemplateParams {
  page?: number
  page_size?: number
  keyword?: string
  status?: string
}

// ========== 产品运费模版绑定 ==========

// 本地产品运费模版关联
export interface ProductShippingTemplate {
  id: number
  product_id: number
  shipping_template_id: number
  template_name: string
  carrier: string
  is_default: boolean
  sort_order: number
  status: string
  created_at: string
}

// 平台产品运费模版关联
export interface PlatformProductShippingTemplate {
  id: number
  platform_product_id: number
  shipping_template_id: number
  template_name: string
  carrier: string
  is_default: boolean
  sort_order: number
  status: string
  created_at: string
}

// 绑定本地产品运费模版请求
export interface BindProductShippingTemplateRequest {
  product_id: number
  shipping_template_id: number
  is_default?: boolean
  sort_order?: number
}

// 绑定平台产品运费模版请求
export interface BindPlatformProductShippingTemplateRequest {
  platform_product_id: number
  shipping_template_id: number
  is_default?: boolean
  sort_order?: number
}

// 设置默认运费模版请求
export interface SetDefaultShippingTemplateRequest {
  shipping_template_id: number
}

// ========== API 接口 ==========

export default {
  // 获取运费模板列表
  listTemplates: (params?: ListTemplateParams) => {
    return request.get<TemplateListResponse>('/shipping/templates', { params })
  },

  // 获取所有启用的模板（下拉选择）
  listAllTemplates: () => {
    return request.get<TemplateOption[]>('/shipping/templates/all')
  },

  // 获取模板详情
  getTemplate: (id: number) => {
    return request.get<ShippingTemplate>(`/shipping/templates/${id}`)
  },

  // 创建模板
  createTemplate: (data: CreateTemplateRequest) => {
    return request.post<ShippingTemplate>('/shipping/templates', data)
  },

  // 更新模板
  updateTemplate: (id: number, data: UpdateTemplateRequest) => {
    return request.put(`/shipping/templates/${id}`, data)
  },

  // 删除模板
  deleteTemplate: (id: number) => {
    return request.delete(`/shipping/templates/${id}`)
  },

  // 获取模板规则列表
  getTemplateRules: (templateId: number) => {
    return request.get<ShippingTemplateRule[]>(`/shipping/templates/${templateId}/rules`)
  },

  // 创建规则
  createRule: (templateId: number, data: CreateRuleRequest) => {
    return request.post<ShippingTemplateRule>(`/shipping/templates/${templateId}/rules`, data)
  },

  // 更新规则
  updateRule: (templateId: number, ruleId: number, data: UpdateRuleRequest) => {
    return request.put(`/shipping/templates/${templateId}/rules/${ruleId}`, data)
  },

  // 删除规则
  deleteRule: (templateId: number, ruleId: number) => {
    return request.delete(`/shipping/templates/${templateId}/rules/${ruleId}`)
  },

  // 计算运费
  calculateShipping: (data: CalculateShippingRequest) => {
    return request.post<CalculateShippingResponse>('/shipping/calculate', data)
  },

  // 批量计算运费
  batchCalculateShipping: (items: CalculateShippingRequest[]) => {
    return request.post<{ results: CalculateShippingResponse[] }>('/shipping/calculate/batch', { items })
  },

  // ========== 本地产品运费模版绑定 ==========

  // 绑定本地产品运费模版
  bindProductShippingTemplate: (data: BindProductShippingTemplateRequest) => {
    return request.post<ProductShippingTemplate>('/shipping/product-templates', data)
  },

  // 解绑本地产品运费模版
  unbindProductShippingTemplate: (id: number) => {
    return request.delete(`/shipping/product-templates/${id}`)
  },

  // 获取本地产品的运费模版列表
  getProductShippingTemplates: (productId: number) => {
    return request.get<ProductShippingTemplate[]>(`/shipping/products/${productId}/templates`)
  },

  // 设置本地产品的默认运费模版
  setProductDefaultShippingTemplate: (productId: number, shippingTemplateId: number) => {
    return request.put(`/shipping/products/${productId}/default-template`, { shipping_template_id: shippingTemplateId })
  },

  // ========== 平台产品运费模版绑定 ==========

  // 绑定平台产品运费模版
  bindPlatformProductShippingTemplate: (data: BindPlatformProductShippingTemplateRequest) => {
    return request.post<PlatformProductShippingTemplate>('/shipping/platform-product-templates', data)
  },

  // 解绑平台产品运费模版
  unbindPlatformProductShippingTemplate: (id: number) => {
    return request.delete(`/shipping/platform-product-templates/${id}`)
  },

  // 获取平台产品的运费模版列表
  getPlatformProductShippingTemplates: (platformProductId: number) => {
    return request.get<PlatformProductShippingTemplate[]>(`/shipping/platform-products/${platformProductId}/templates`)
  },

  // 设置平台产品的默认运费模版
  setPlatformProductDefaultShippingTemplate: (platformProductId: number, shippingTemplateId: number) => {
    return request.put(`/shipping/platform-products/${platformProductId}/default-template`, { shipping_template_id: shippingTemplateId })
  }
}
