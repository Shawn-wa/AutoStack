<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, Refresh } from '@element-plus/icons-vue'
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

// 佣金详情展开状态
const commissionExpanded = ref<string[]>(['commission'])

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

// 返回列表
const handleBack = () => {
  router.push({ name: 'Orders' })
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
            <el-descriptions-item label="订单号">{{ order.platform_order_no }}</el-descriptions-item>
            <el-descriptions-item label="平台">
              <el-tag type="primary" size="small">{{ getPlatformLabel(order.platform) }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="平台状态">{{ order.platform_status }}</el-descriptions-item>
            <el-descriptions-item label="订单金额">
              <span class="amount">{{ order.currency }} {{ order.total_amount?.toFixed(2) }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="下单时间">
              {{ order.order_time ? formatDateTime(order.order_time) : '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="发货时间">
              {{ order.ship_time ? formatDateTime(order.ship_time) : '-' }}
            </el-descriptions-item>
          </el-descriptions>
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
            <div class="commission-item main">
              <span class="label">销售佣金</span>
              <span class="value warning">{{ commissionCurrency }} {{ order.sale_commission?.toFixed(2) || '0.00' }}</span>
            </div>
            <div class="commission-item">
              <span class="label">总扣款金额</span>
              <span class="value danger">{{ commissionCurrency }} {{ order.commission_amount?.toFixed(2) || '0.00' }}</span>
            </div>
          </div>
          
          <el-collapse v-model="commissionExpanded">
            <el-collapse-item title="查看详细佣金明细" name="commission">
              <el-descriptions :column="2" border size="small">
                <el-descriptions-item label="销售佣金">
                  {{ commissionCurrency }} {{ order.sale_commission?.toFixed(2) || '0.00' }}
                </el-descriptions-item>
                <el-descriptions-item label="销售收入">
                  {{ commissionCurrency }} {{ order.accruals_for_sale?.toFixed(2) || '0.00' }}
                </el-descriptions-item>
                <el-descriptions-item label="配送费">
                  {{ commissionCurrency }} {{ order.delivery_charge?.toFixed(2) || '0.00' }}
                </el-descriptions-item>
                <el-descriptions-item label="退货配送费">
                  {{ commissionCurrency }} {{ order.return_delivery_charge?.toFixed(2) || '0.00' }}
                </el-descriptions-item>
              </el-descriptions>
            </el-collapse-item>
          </el-collapse>
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

        <!-- 商品信息卡片 -->
        <div class="info-card">
          <div class="card-header">
            <h2 class="card-title">商品信息</h2>
            <span class="item-count">共 {{ order.items?.length || 0 }} 件商品</span>
          </div>
          <el-table :data="order.items" border stripe>
            <el-table-column prop="name" label="商品名称" min-width="250" show-overflow-tooltip />
            <el-table-column prop="platform_sku" label="平台SKU" width="150" />
            <el-table-column prop="sku" label="SKU" width="150" />
            <el-table-column prop="quantity" label="数量" width="80" align="center" />
            <el-table-column label="单价" width="120" align="right">
              <template #default="{ row }">
                {{ row.currency }} {{ row.price?.toFixed(2) }}
              </template>
            </el-table-column>
            <el-table-column label="小计" width="120" align="right">
              <template #default="{ row }">
                {{ row.currency }} {{ (row.price * row.quantity)?.toFixed(2) }}
              </template>
            </el-table-column>
          </el-table>
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
  }

  .value {
    font-size: 20px;
    font-weight: 600;

    &.warning {
      color: var(--el-color-warning);
    }

    &.danger {
      color: var(--el-color-danger);
    }
  }

  &.main {
    padding-right: 40px;
    border-right: 1px solid var(--border-color);
  }
}

:deep(.el-descriptions) {
  margin-bottom: 0;
}

:deep(.el-collapse-item__header) {
  font-size: 13px;
  color: var(--el-color-primary);
}
</style>

