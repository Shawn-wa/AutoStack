import request from '@/commonBase/api/request'

// 产品接口
export interface Product {
  id: number
  wid: number              // 仓库ID
  warehouse_name: string   // 仓库名称
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
  image: string
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

// 平台产品SKU详情
export interface OrderSummaryPlatformSKU {
  sku: string
  name: string
  image: string
}

// 订单汇总项接口（按本地SKU合并）
export interface OrderSummaryItem {
  local_sku: string
  product_name: string
  platform_skus: string[]
  platform_products: OrderSummaryPlatformSKU[]
  quantity: number
  amount: number
  currency: string
  status_details: OrderSummaryStatusDetail[]
  available_stock: number  // 系统可用库存
}

// 创建产品请求
export interface CreateProductRequest {
  wid?: number             // 仓库ID
  sku: string
  name: string
  image?: string
  cost_price?: number
  weight?: number
  dimensions?: string
}

// 更新产品请求
export interface UpdateProductRequest {
  wid?: number             // 仓库ID
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
  keyword?: string
}

// 平台产品列表查询请求
export interface ListPlatformProductRequest extends ListRequest {
  platform_auth_id?: number
  keyword?: string
}

// 订单汇总查询请求
export interface OrderSummaryRequest {
  start_time?: string
  end_time?: string
  auth_id?: number
  platform?: string
  keyword?: string // 搜索关键词（本地SKU/标题/平台SKU）
  status?: string  // 订单状态筛选
}

// 初始化产品响应
export interface InitProductsResponse {
  total_platform_products: number // 平台产品总数
  skipped_mapped: number          // 跳过（已有映射）
  skipped_existing: number        // 跳过（SKU已存在但已关联）
  created_products: number        // 新创建的本地产品数
  created_mappings: number        // 新创建的映射数
}

// ========== 入库单相关 ==========

// 入库单明细请求
export interface StockInOrderItemRequest {
  product_id: number
  quantity: number
}

// 创建入库单请求
export interface CreateStockInOrderRequest {
  warehouse_id: number
  items: StockInOrderItemRequest[]
  remark?: string
}

// 入库单明细响应
export interface StockInOrderItemResponse {
  id: number
  product_id: number
  sku: string
  product_name: string
  quantity: number
}

// 入库单响应
export interface StockInOrderResponse {
  id: number
  order_no: string
  warehouse_id: number
  warehouse_name: string
  status: string
  remark: string
  items: StockInOrderItemResponse[]
  created_at: string
}

// ========== 仓库相关 ==========

// 仓库响应
export interface WarehouseResponse {
  id: number
  code: string
  name: string
  type: string      // 仓库类型：local/overseas/fba/third/virtual
  address: string
  status: string
  created_at: string
}

// ========== 库存相关 ==========

// 库存明细响应
export interface InventoryResponse {
  id: number
  product_id: number
  warehouse_id: number
  sku: string
  product_name: string
  product_image: string
  warehouse_code: string
  warehouse_name: string
  available_stock: number
  locked_stock: number
  in_transit_stock: number
  total_stock: number
  updated_at: string
}

// 库存列表查询请求
export interface ListInventoryRequest extends ListRequest {
  warehouse_id?: number
  keyword?: string
}

// ========== 供应商/采购信息相关 ==========

// 供应商接口
export interface Supplier {
  id: number
  product_id: number
  product_sku?: string
  product_name?: string
  supplier_name: string
  purchase_link: string
  unit_price: number
  currency: string
  min_order_qty: number
  lead_time: number
  estimated_days: number
  remark: string
  is_default: boolean
  status: string
  created_at: string
  updated_at: string
}

// 创建供应商请求
export interface CreateSupplierRequest {
  product_id: number
  supplier_name: string
  purchase_link?: string
  unit_price?: number
  currency?: string
  min_order_qty?: number
  lead_time?: number
  estimated_days?: number
  remark?: string
  is_default?: boolean
}

// 更新供应商请求
export interface UpdateSupplierRequest {
  supplier_name?: string
  purchase_link?: string
  unit_price?: number
  currency?: string
  min_order_qty?: number
  lead_time?: number
  estimated_days?: number
  remark?: string
  is_default?: boolean
  status?: string
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
  },

  // 初始化本地产品（根据平台SKU生成）
  initProducts: (platformAuthId?: number) => {
    return request.post<InitProductsResponse>('/product/init', { platform_auth_id: platformAuthId || 0 })
  },

  // 入库单
  listStockInOrders: (params: ListRequest & { status?: string }) => {
    return request.get('/product/stock-in-orders', { params })
  },
  createStockInOrder: (data: CreateStockInOrderRequest) => {
    return request.post<StockInOrderResponse>('/product/stock-in-orders', data)
  },
  getStockInOrder: (id: number) => {
    return request.get<StockInOrderResponse>(`/product/stock-in-orders/${id}`)
  },

  // 仓库
  listWarehouses: () => {
    return request.get<WarehouseResponse[]>('/product/warehouses')
  },
  // 获取当前用户可用仓库（用于入库单等业务场景）
  listAvailableWarehouses: () => {
    return request.get<{ list: WarehouseResponse[]; total: number }>('/product/warehouses/available')
  },
  listAllWarehouses: (type?: string) => {
    const params = type && type !== 'all' ? { type } : {}
    return request.get<WarehouseResponse[]>('/product/warehouses/all', { params })
  },
  createWarehouse: (data: { code: string; name: string; type?: string; address?: string }) => {
    return request.post<WarehouseResponse>('/product/warehouses', data)
  },

  // 库存
  listInventory: (params: ListInventoryRequest) => {
    return request.get('/product/inventory', { params })
  },
  updateInventory: (data: { product_id: number; warehouse_id: number; available_stock?: number; locked_stock?: number; in_transit_stock?: number }) => {
    return request.put('/product/inventory', data)
  },
  initInventory: (warehouseId: number) => {
    return request.post('/product/inventory/init', null, { params: { warehouse_id: warehouseId } })
  },

  // 供应商/采购信息
  listSuppliers: (params: ListRequest & { product_id?: number; status?: string }) => {
    return request.get('/product/suppliers', { params })
  },
  getProductSuppliers: (productId: number) => {
    return request.get<Supplier[]>(`/product/products/${productId}/suppliers`)
  },
  createSupplier: (data: CreateSupplierRequest) => {
    return request.post<Supplier>('/product/suppliers', data)
  },
  updateSupplier: (id: number, data: UpdateSupplierRequest) => {
    return request.put<Supplier>(`/product/suppliers/${id}`, data)
  },
  deleteSupplier: (id: number) => {
    return request.delete(`/product/suppliers/${id}`)
  }
}

export default api
