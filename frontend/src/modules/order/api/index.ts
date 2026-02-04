import request from '@/commonBase/api/request'

// 凭证字段定义
export interface CredentialField {
  key: string
  label: string
  type: string
  required: boolean
}

// 平台信息
export interface PlatformInfo {
  name: string
  label: string
  fields: CredentialField[]
}

// 平台授权
export interface PlatformAuth {
  id: number
  platform: string
  shop_name: string
  status: number
  masked_credentials?: Record<string, string>
  last_sync_at: string | null
  created_at: string
  updated_at: string
}

// 创建授权参数
export interface CreateAuthParams {
  platform: string
  shop_name: string
  credentials: Record<string, string>
}

// 更新授权参数
export interface UpdateAuthParams {
  shop_name?: string
  credentials?: Record<string, string>
  status?: number
}

// 授权列表响应
export interface AuthListResult {
  list: PlatformAuth[]
  total: number
  page: number
  page_size: number
}

// 同步订单参数
export interface SyncOrdersParams {
  since?: string
  to?: string
}

// 同步订单结果
export interface SyncOrdersResult {
  total: number
  created: number
  updated: number
}

// 订单商品
export interface OrderItem {
  id: number
  platform_sku: string
  sku: string
  name: string
  quantity: number
  price: number
  currency: string
  estimated_shipping_fee: number
  estimated_shipping_currency: string
}

// 订单
export interface Order {
  id: number
  platform: string
  platform_order_no: string
  status: string
  platform_status: string
  total_amount: number
  currency: string
  recipient_name: string
  recipient_phone: string
  country: string
  province: string
  city: string
  zip_code: string
  address: string
  order_time: string | null
  ship_time: string | null
  ship_deadline: string | null  // 发货截止时间
  // 佣金信息
  sale_commission: number
  accruals_for_sale: number
  processing_and_delivery: number      // 加工和配送费
  refunds_and_cancellations: number    // 退款和取消
  services_amount: number              // 服务费
  compensation_amount: number          // 补偿金额
  money_transfer: number               // 资金转账
  others_amount: number                // 其他费用
  profit_amount: number                // 订单利润额
  commission_currency: string
  commission_synced_at: string | null
  // 物流费用估算
  estimated_shipping_fee: number
  estimated_shipping_currency: string
  shipping_template_id: number
  shipping_estimated_at: string | null
  items: OrderItem[]
  created_at: string
  updated_at: string
}

// 订单列表参数
export interface OrderListParams {
  page?: number
  page_size?: number
  platform?: string
  auth_id?: number
  status?: string
  keyword?: string
  start_time?: string
  end_time?: string
  deadline_filter?: string  // 发货截止时间筛选: overdue, within_1d, within_3d
}

// 订单列表响应
export interface OrderListResult {
  list: Order[]
  total: number
  page: number
  page_size: number
}

// 获取支持的平台列表
export function getPlatforms() {
  return request.get<any, { data: PlatformInfo[] }>('/order/platforms')
}

// 获取授权列表
export function getAuths(page: number = 1, pageSize: number = 10) {
  return request.get<any, { data: AuthListResult }>('/order/auths', {
    params: { page, page_size: pageSize }
  })
}

// 创建授权
export function createAuth(data: CreateAuthParams) {
  return request.post<any, { data: PlatformAuth }>('/order/auths', data)
}

// 更新授权
export function updateAuth(id: number, data: UpdateAuthParams) {
  return request.put<any, { data: PlatformAuth }>(`/order/auths/${id}`, data)
}

// 删除授权
export function deleteAuth(id: number) {
  return request.delete<any, { data: null }>(`/order/auths/${id}`)
}

// 测试授权连接
export function testAuth(id: number) {
  return request.post<any, { data: null }>(`/order/auths/${id}/test`)
}

// 同步订单（超时时间延长至5分钟）
export function syncOrders(id: number, data?: SyncOrdersParams) {
  return request.post<any, { data: SyncOrdersResult }>(`/order/auths/${id}/sync`, data || {}, {
    timeout: 300000 // 5分钟
  })
}

// 获取订单列表
export function getOrders(params: OrderListParams = {}) {
  return request.get<any, { data: OrderListResult }>('/order/orders', { params })
}

// 获取订单详情
export function getOrder(id: number) {
  return request.get<any, { data: Order }>(`/order/orders/${id}`)
}

// 同步佣金参数
export interface SyncCommissionParams {
  since?: string
  to?: string
}

// 同步佣金结果
export interface SyncCommissionResult {
  total: number
  updated: number
}

// 同步佣金
export function syncCommission(id: number, data?: SyncCommissionParams) {
  return request.post<any, { data: SyncCommissionResult }>(`/order/auths/${id}/sync-commission`, data || {})
}

// 同步单个订单的佣金
export function syncOrderCommission(orderId: number) {
  return request.post<any, { data: Order }>(`/order/orders/${orderId}/sync-commission`)
}

// 同步单个订单信息（从平台获取最新状态）
export function syncSingleOrder(orderId: number) {
  return request.post<any, { data: Order }>(`/order/orders/${orderId}/sync`)
}

// ========== 现金流报表相关 ==========

// 现金流报表
export interface CashFlowStatement {
  id: number
  platform_auth_id: number
  platform: string
  period_begin: string | null
  period_end: string | null
  currency_code: string
  orders_amount: number
  returns_amount: number
  commission_amount: number
  services_amount: number
  item_delivery_and_return_amount: number
  synced_at: string
}

// 现金流报表列表参数
export interface CashFlowListParams {
  page?: number
  page_size?: number
  auth_id?: number
}

// 现金流报表列表响应
export interface CashFlowListResult {
  list: CashFlowStatement[]
  total: number
  page: number
  page_size: number
}

// 同步现金流参数
export interface SyncCashFlowParams {
  since?: string
  to?: string
}

// 同步现金流结果
export interface SyncCashFlowResult {
  total: number
  created: number
  updated: number
  skipped: number
}

// 获取现金流报表列表
export function getCashFlowList(params: CashFlowListParams = {}) {
  return request.get<any, { data: CashFlowListResult }>('/order/cashflow', { params })
}

// 同步现金流报表
export function syncCashFlow(authId: number, data?: SyncCashFlowParams) {
  return request.post<any, { data: SyncCashFlowResult }>(`/order/auths/${authId}/sync-cashflow`, data || {})
}

// ========== 仪表盘统计相关 ==========

// 仪表盘统计数据
export interface DashboardStats {
  total_orders: number
  delivered_orders: number
  pending_orders: number
  today_orders: number
  shipped_orders: number   // 已发货订单数
  timeout_orders: number   // 即将超时订单数
  total_amounts: Array<{ currency: string; amount: number }>
  total_profit: number
  total_commission: number
  total_service_fee: number
  total_auths: number
  active_auths: number
  currency: string
}

// 最近订单
export interface RecentOrder {
  id: number
  platform_order_no: string
  status: string
  total_amount: number
  currency: string
  order_time: string | null
}

// 获取仪表盘统计数据
export function getDashboardStats() {
  return request.get<any, { data: DashboardStats }>('/order/dashboard/stats')
}

// 获取最近订单
export function getRecentOrders(limit: number = 10) {
  return request.get<any, { data: RecentOrder[] }>('/order/dashboard/recent-orders', {
    params: { limit }
  })
}

// 订单趋势数据项
export interface OrderTrendItem {
  date: string
  count: number
  amount: number
}

// 订单趋势响应
export interface OrderTrendResponse {
  items: OrderTrendItem[]
}

// 获取订单趋势数据
export function getOrderTrend(days: number = 7, currency?: string) {
  return request.get<any, { data: OrderTrendResponse }>('/order/dashboard/trend', {
    params: { days, currency }
  })
}

// 初始化仪表盘统计数据（首次访问时调用）
export function initDashboardStats() {
  return request.post<any, { data: null }>('/order/dashboard/init')
}

// 刷新仪表盘统计数据（强制重新计算）
export function refreshDashboardStats() {
  return request.post<any, { data: null }>('/order/dashboard/refresh')
}
