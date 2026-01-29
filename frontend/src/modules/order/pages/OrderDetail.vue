<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

defineOptions({ name: 'OrderDetail' })
import { ArrowLeft, Refresh, CopyDocument, QuestionFilled } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import {
  getOrder,
  getPlatforms,
  syncOrderCommission,
  type Order,
  type PlatformInfo
} from '@/modules/order/api'
import { formatDateTime } from '@/utils/format'

const route = useRoute()
const router = useRouter()

const loading = ref(true)
const syncingCommission = ref(false)
const order = ref<Order | null>(null)
const platforms = ref<PlatformInfo[]>([])

// 获取订单ID
const orderId = computed(() => Number(route.params.id))

// 获取平台显示名称
const getPlatformLabel = (name: string) => {
  const platform = platforms.value.find(p => p.name === name)
  return platform?.label || name
}

// 获取状态标签类型
const getStatusTagType = (status: string) => {
  switch (status) {
    case 'pending': return 'warning'
    case 'ready_to_ship': return 'primary'
    case 'shipped': return 'info'
    case 'delivered': return 'success'
    case 'cancelled': return 'danger'
    default: return 'info'
  }
}

// 获取状态显示文字
const getStatusText = (status: string) => {
  switch (status) {
    case 'pending': return '待处理'
    case 'ready_to_ship': return '待发货'
    case 'shipped': return '已发货'
    case 'delivered': return '已签收'
    case 'cancelled': return '已取消'
    default: return status
  }
}

// 佣金货币（优先使用佣金货币，否则使用订单货币）
const commissionCurrency = computed(() => {
  return order.value?.commission_currency || order.value?.currency || ''
})

// 订单利润额（直接使用后端计算的值）
const profitAmount = computed(() => {
  return order.value?.profit_amount || 0
})

// 获取平台列表
const fetchPlatforms = async () => {
  try {
    const res = await getPlatforms()
    platforms.value = res.data
  } catch (error) {
    console.error('获取平台列表失败', error)
  }
}

// 获取订单详情
const fetchOrder = async () => {
  loading.value = true
  try {
    const res = await getOrder(orderId.value)
    order.value = res.data
  } catch (error) {
    console.error('获取订单详情失败', error)
    ElMessage.error('获取订单详情失败')
  } finally {
    loading.value = false
  }
}

// 复制订单号
const handleCopyOrderNo = async () => {
  if (!order.value) return
  try {
    await navigator.clipboard.writeText(order.value.platform_order_no)
    ElMessage.success('订单号已复制')
  } catch (error) {
    ElMessage.error('复制失败')
  }
}

// 返回列表（使用浏览器历史返回，保留筛选条件）
const handleBack = () => {
  // 如果有历史记录则返回，否则跳转到订单列表
  if (window.history.length > 1) {
    router.back()
  } else {
    router.push({ name: 'Orders' })
  }
}

// 刷新
const handleRefresh = () => {
  fetchOrder()
}

// 同步佣金
const handleSyncCommission = async () => {
  if (!order.value) return
  
  syncingCommission.value = true
  try {
    const res = await syncOrderCommission(order.value.id)
    order.value = res.data
    ElMessage.success('佣金同步成功')
  } catch (error) {
    console.error('同步佣金失败', error)
  } finally {
    syncingCommission.value = false
  }
}

// 监听订单ID变化，重新获取数据（解决组件复用时数据不刷新的问题）
watch(orderId, (newId, oldId) => {
  if (newId !== oldId && newId) {
    // 重置状态，避免显示旧数据
    order.value = null
    fetchOrder()
  }
})

onMounted(() => {
  fetchPlatforms()
  fetchOrder()
})
</script>

<template>
  <div class="order-detail-page">
    <div class="page-header">
      <div class="header-left">
        <el-button :icon="ArrowLeft" @click="handleBack">返回列表</el-button>
        <div class="title-section">
          <h1 class="page-title">订单详情</h1>
          <p class="page-desc" v-if="order">订单号：{{ order.platform_order_no }}</p>
        </div>
      </div>
      <div class="header-right">
        <el-button :icon="Refresh" @click="handleRefresh" :loading="loading">刷新</el-button>
      </div>
    </div>

    <div v-loading="loading" class="content-wrapper">
      <template v-if="order">
        <!-- 基本信息卡片 -->
        <div class="info-card">
          <div class="card-header">
            <h2 class="card-title">基本信息</h2>
            <el-tag :type="getStatusTagType(order.status)" size="large">
              {{ getStatusText(order.status) }}
            </el-tag>
          </div>
          <el-descriptions :column="3" border>
            <el-descriptions-item label="订单号">
              <span class="order-no-wrapper">
                {{ order.platform_order_no }}
                <el-icon class="copy-icon" @click="handleCopyOrderNo" title="复制订单号">
                  <CopyDocument />
                </el-icon>
              </span>
            </el-descriptions-item>
            <el-descriptions-item label="平台">
              <el-tag type="primary" size="small">{{ getPlatformLabel(order.platform) }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="平台状态">{{ order.platform_status }}</el-descriptions-item>
            <el-descriptions-item label="订单金额">
              <span class="amount">{{ order.total_amount?.toFixed(2) }} {{ order.currency }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="下单时间">
              {{ order.order_time ? formatDateTime(order.order_time) : '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="发货时间">
              {{ order.ship_time ? formatDateTime(order.ship_time) : '-' }}
            </el-descriptions-item>
          </el-descriptions>
        </div>

        <!-- 商品信息卡片 -->
        <div class="info-card">
          <div class="card-header">
            <h2 class="card-title">商品信息</h2>
            <span class="item-count">共 {{ order.items?.length || 0 }} 件商品</span>
          </div>
          <el-table :data="order.items" border stripe :cell-style="{ whiteSpace: 'nowrap' }">
            <el-table-column prop="name" label="商品名称" min-width="250" show-overflow-tooltip />
            <el-table-column prop="platform_sku" label="平台SKU" min-width="150" />
            <el-table-column prop="sku" label="SKU" min-width="150" />
            <el-table-column prop="quantity" label="数量" width="80" align="center" />
            <el-table-column label="单价" min-width="120" align="right">
              <template #default="{ row }">
                {{ row.price?.toFixed(2) }}&nbsp;{{ row.currency }}
              </template>
            </el-table-column>
            <el-table-column label="小计" min-width="120" align="right">
              <template #default="{ row }">
                {{ (row.price * row.quantity)?.toFixed(2) }}&nbsp;{{ row.currency }}
              </template>
            </el-table-column>
          </el-table>
        </div>

        <!-- 佣金信息卡片 -->
        <div class="info-card">
          <div class="card-header">
            <h2 class="card-title">佣金信息</h2>
            <div class="header-actions">
              <span v-if="order.commission_synced_at" class="sync-time">
                同步时间：{{ formatDateTime(order.commission_synced_at) }}
              </span>
              <el-tag v-else type="info" size="small">未同步</el-tag>
              <el-button 
                type="primary" 
                size="small" 
                :loading="syncingCommission"
                @click="handleSyncCommission"
              >
                更新佣金
              </el-button>
            </div>
          </div>
          <div class="commission-summary">
            <div class="commission-item">
              <span class="label">
                销售收入
                <el-tooltip content="卖家因销售商品获得的收入金额" placement="top">
                  <el-icon class="info-icon"><QuestionFilled /></el-icon>
                </el-tooltip>
              </span>
              <span class="value success">{{ order.accruals_for_sale?.toFixed(2) || '0.00' }} {{ commissionCurrency }}</span>
            </div>
            <div class="commission-item">
              <span class="label">
                销售佣金
                <el-tooltip content="平台从销售中收取的佣金费用" placement="top">
                  <el-icon class="info-icon"><QuestionFilled /></el-icon>
                </el-tooltip>
              </span>
              <span class="value warning">{{ order.sale_commission?.toFixed(2) || '0.00' }} {{ commissionCurrency }}</span>
            </div>
            <div class="commission-item main">
              <span class="label">
                订单利润额
                <el-tooltip content="所有费用项汇总后的最终利润" placement="top">
                  <el-icon class="info-icon"><QuestionFilled /></el-icon>
                </el-tooltip>
              </span>
              <span class="value" :class="profitAmount >= 0 ? 'success' : 'danger'">
                {{ profitAmount.toFixed(2) }} {{ commissionCurrency }}
              </span>
            </div>
          </div>
          
          <el-table :data="[order]" border class="commission-table" :cell-style="{ whiteSpace: 'nowrap' }" :header-cell-style="{ whiteSpace: 'nowrap' }">
            <el-table-column label="销售收入" align="right" min-width="80">
              <template #default>
                <span class="fee-positive">+{{ Math.abs(order.accruals_for_sale || 0).toFixed(2) }}&nbsp;{{ commissionCurrency }}</span>
              </template>
            </el-table-column>
            <el-table-column label="销售佣金" align="right" min-width="80">
              <template #default>
                <span class="fee-negative">{{ (order.sale_commission || 0).toFixed(2) }}&nbsp;{{ commissionCurrency }}</span>
              </template>
            </el-table-column>
            <el-table-column label="加工配送费" align="right" min-width="80">
              <template #default>
                <span class="fee-negative">{{ (order.processing_and_delivery || 0).toFixed(2) }}&nbsp;{{ commissionCurrency }}</span>
              </template>
            </el-table-column>
            <el-table-column label="服务费" align="right" min-width="80">
              <template #default>
                <span class="fee-negative">{{ (order.services_amount || 0).toFixed(2) }}&nbsp;{{ commissionCurrency }}</span>
              </template>
            </el-table-column>
            <el-table-column label="退款取消" align="right" min-width="80">
              <template #default>
                <span class="fee-negative">{{ (order.refunds_and_cancellations || 0).toFixed(2) }}&nbsp;{{ commissionCurrency }}</span>
              </template>
            </el-table-column>
            <el-table-column label="平台补偿" align="right" min-width="80">
              <template #default>
                <span class="fee-positive">+{{ Math.abs(order.compensation_amount || 0).toFixed(2) }}&nbsp;{{ commissionCurrency }}</span>
              </template>
            </el-table-column>
            <el-table-column label="其他" align="right" min-width="80">
              <template #default>
                <span>{{ (order.others_amount || 0).toFixed(2) }}&nbsp;{{ commissionCurrency }}</span>
              </template>
            </el-table-column>
            <el-table-column label="订单利润额" align="right" min-width="80">
              <template #default>
                <span class="fee-profit" :class="profitAmount >= 0 ? 'positive' : 'negative'">
                  {{ profitAmount.toFixed(2) }}&nbsp;{{ commissionCurrency }}
                </span>
              </template>
            </el-table-column>
          </el-table>
        </div>

        <!-- 收件人信息卡片 -->
        <div class="info-card">
          <div class="card-header">
            <h2 class="card-title">收件人信息</h2>
          </div>
          <el-descriptions :column="2" border>
            <el-descriptions-item label="收件人">{{ order.recipient_name || '-' }}</el-descriptions-item>
            <el-descriptions-item label="电话">{{ order.recipient_phone || '-' }}</el-descriptions-item>
            <el-descriptions-item label="国家">{{ order.country || '-' }}</el-descriptions-item>
            <el-descriptions-item label="省/州">{{ order.province || '-' }}</el-descriptions-item>
            <el-descriptions-item label="城市">{{ order.city || '-' }}</el-descriptions-item>
            <el-descriptions-item label="邮编">{{ order.zip_code || '-' }}</el-descriptions-item>
            <el-descriptions-item label="详细地址" :span="2">{{ order.address || '-' }}</el-descriptions-item>
          </el-descriptions>
        </div>
      </template>

      <el-empty v-else-if="!loading" description="订单不存在或已删除">
        <el-button type="primary" @click="handleBack">返回列表</el-button>
      </el-empty>
    </div>
  </div>
</template>

<style scoped lang="scss">
.order-detail-page {
  padding: 0;
}

.page-header {
  margin-bottom: 24px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.title-section {
  display: flex;
  flex-direction: column;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  margin: 0;
}

.page-desc {
  color: var(--text-secondary);
  margin: 4px 0 0 0;
  font-size: 14px;
}

.content-wrapper {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.info-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: 24px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  margin: 0;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.sync-time {
  color: var(--text-secondary);
  font-size: 12px;
}

.item-count {
  color: var(--text-secondary);
  font-size: 14px;
}

.amount {
  font-size: 16px;
  font-weight: 600;
  color: var(--el-color-primary);
}

.order-no-wrapper {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.copy-icon {
  cursor: pointer;
  color: var(--text-secondary);
  transition: color 0.2s;

  &:hover {
    color: var(--el-color-primary);
  }
}

.commission-summary {
  display: flex;
  gap: 40px;
  margin-bottom: 16px;
  padding: 16px;
  background: var(--el-fill-color-light);
  border-radius: var(--radius-md);
}

.commission-item {
  display: flex;
  flex-direction: column;
  gap: 4px;

  .label {
    font-size: 13px;
    color: var(--text-secondary);
    display: flex;
    align-items: center;
  }

  .value {
    font-size: 20px;
    font-weight: 600;

    &.success {
      color: var(--el-color-success);
    }

    &.warning {
      color: var(--el-color-warning);
    }

    &.danger {
      color: var(--el-color-danger);
    }
  }

  &.main {
    padding-left: 40px;
    border-left: 1px solid var(--border-color);
  }
}

:deep(.el-descriptions) {
  margin-bottom: 0;
}

.commission-table {
  margin-top: 8px;
  width: auto;
  max-width: 100%;

  :deep(.el-table__header th) {
    background: var(--el-fill-color-light);
    font-size: 13px;
    font-weight: 600;
    padding: 10px 16px;
    white-space: nowrap;
  }

  :deep(.el-table__body td) {
    font-size: 14px;
    padding: 10px 16px;
    white-space: nowrap;
  }

  :deep(.el-table__cell) {
    white-space: nowrap;
  }
}

.fee-positive {
  color: var(--el-color-success);
  font-weight: 500;
}

.fee-negative {
  color: var(--el-color-warning);
  font-weight: 500;
}

.fee-profit {
  font-weight: 600;
  font-size: 15px;

  &.positive {
    color: var(--el-color-success);
  }

  &.negative {
    color: var(--el-color-danger);
  }
}

.info-icon {
  font-size: 12px;
  color: var(--text-secondary);
  margin-left: 4px;
  cursor: help;
}

.desc-label {
  display: inline-flex;
  align-items: center;
}

.fee-detail-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.fee-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: var(--el-fill-color-lighter);
  border-radius: var(--radius-sm);

  &.total {
    background: var(--el-fill-color);
    font-weight: 600;
    margin-top: 8px;
    padding: 12px;
    border-top: 1px dashed var(--border-color);
  }
}

.fee-label {
  display: flex;
  align-items: center;
  font-size: 14px;
  color: var(--text-primary);
}

.fee-value {
  font-weight: 500;
  font-size: 14px;

  &.positive {
    color: var(--el-color-success);
  }

  &.negative {
    color: var(--el-color-warning);
  }
}

.fee-breakdown {
  padding: 8px 0;
}

.fee-section {
  margin-bottom: 16px;
}

.fee-section-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
  margin-bottom: 8px;
  padding-bottom: 4px;
  border-bottom: 1px solid var(--border-color);
}

.fee-line {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: var(--el-fill-color-lighter);
  border-radius: var(--radius-sm);
  margin-bottom: 4px;
}

.fee-name {
  font-size: 14px;
  color: var(--text-primary);
}

.fee-amount {
  font-size: 14px;
  font-weight: 500;

  &.positive {
    color: var(--el-color-success);
  }

  &.negative {
    color: var(--el-color-warning);
  }
}

.fee-total {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  background: var(--el-fill-color);
  border-radius: var(--radius-sm);
  margin-top: 12px;
  border-top: 2px solid var(--border-color);

  .fee-name {
    font-weight: 600;
  }

  .fee-amount {
    font-size: 18px;
    font-weight: 600;
  }
}
</style>

