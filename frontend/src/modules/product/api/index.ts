import request from '@/commonBase/api/request'

// 产品接口
export interface Product {
  id: number
  sku: string
  name: string
  image: string
  cost_price: number
  weight: number
  dimensions: string
  created_at: string
  updated_at: string
}

// 平台产品接口
export interface PlatformProduct {
  id: number
  platform: string
  platform_auth_id: number
  platform_sku: string
  name: string
  stock: number
  price: number
  currency: string
  status: string
  created_at: string
  updated_at: string
  product_mapping?: {
    id: number
    product_id: number
    product?: Product
  }
}

// 订单汇总状态明细
export interface OrderSummaryStatusDetail {
  status: string
  quantity: number
  amount: number
}

// 订单汇总项接口（按本地SKU合并）
export interface OrderSummaryItem {
  local_sku: string
  product_name: string
  platform_skus: string[]
  quantity: number
  amount: number
  currency: string
  status_details: OrderSummaryStatusDetail[]
}

// 创建产品请求
export interface CreateProductRequest {
  sku: string
  name: string
  image?: string
  cost_price?: number
  weight?: number
  dimensions?: string
}

// 更新产品请求
export interface UpdateProductRequest {
  name?: string
  image?: string
  cost_price?: number
  weight?: number
  dimensions?: string
}

// 列表查询请求
export interface ListRequest {
  page?: number
  page_size?: number
}

// 平台产品列表查询请求
export interface ListPlatformProductRequest extends ListRequest {
  platform_auth_id?: number
}

// 订单汇总查询请求
export interface OrderSummaryRequest {
  start_time?: string
  end_time?: string
  auth_id?: number
  platform?: string
}

const api = {
  // 本地产品
  listProducts: (params: ListRequest) => {
    return request.get('/product/products', { params })
  },
  createProduct: (data: CreateProductRequest) => {
    return request.post('/product/products', data)
  },
  updateProduct: (id: number, data: UpdateProductRequest) => {
    return request.put(`/product/products/${id}`, data)
  },
  deleteProduct: (id: number) => {
    return request.delete(`/product/products/${id}`)
  },

  // 平台产品
  listPlatformProducts: (params: ListPlatformProductRequest) => {
    return request.get('/product/platform-products', { params })
  },
  syncProducts: (authId: number) => {
    return request.post('/product/sync', { platform_auth_id: authId })
  },
  mapProduct: (platformProductId: number, productId: number) => {
    return request.post('/product/map', { platform_product_id: platformProductId, product_id: productId })
  },
  unmapProduct: (platformProductId: number) => {
    return request.delete(`/product/map/${platformProductId}`)
  },

  // 订单汇总
  getOrderSummary: (params: OrderSummaryRequest) => {
    return request.get('/order/stats/summary', { params })
  }
}

export default api
