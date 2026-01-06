<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { getDashboardStats, getRecentOrders, type DashboardStats, type RecentOrder } from '@/modules/order/api'
import { formatCurrency, formatDateTime } from '@/utils/format'

defineOptions({ name: 'Dashboard' })

const router = useRouter()
const loading = ref(false)
const stats = ref<DashboardStats | null>(null)
const recentOrders = ref<RecentOrder[]>([])

// 状态映射
const statusMap: Record<string, { label: string; color: string }> = {
  pending: { label: '待处理', color: 'warning' },
  accepted: { label: '已接单', color: 'primary' },
  shipped: { label: '已发货', color: 'info' },
  delivered: { label: '已签收', color: 'success' },
  cancelled: { label: '已取消', color: 'danger' },
}

// 格式化金额
const formatAmount = (value: number, currency: string = 'RUB') => {
  return formatCurrency(value, currency, 'ru-RU')
}

// 统计卡片数据
const statCards = computed(() => {
  if (!stats.value) return []
  return [
    { 
      label: '总订单数', 
      value: stats.value.total_orders, 
      icon: '📦', 
      color: 'primary',
      suffix: '单'
    },
    { 
      label: '已签收订单', 
      value: stats.value.delivered_orders, 
      icon: '✅', 
      color: 'success',
      suffix: '单'
    },
    { 
      label: '今日新增', 
      value: stats.value.today_orders, 
      icon: '📈', 
      color: 'accent',
      suffix: '单'
    },
    { 
      label: '待处理订单', 
      value: stats.value.pending_orders, 
      icon: '⏳', 
      color: 'warning',
      suffix: '单'
    },
  ]
})

// 金额统计卡片
const amountCards = computed(() => {
  if (!stats.value) return []
  return [
    { 
      label: '订单总金额', 
      value: stats.value.total_order_amount,
      currency: stats.value.currency,
      icon: '💰', 
      color: 'primary'
    },
    { 
      label: '累计利润', 
      value: stats.value.total_profit,
      currency: stats.value.currency,
      icon: '📊', 
      color: stats.value.total_profit >= 0 ? 'success' : 'danger'
    },
    { 
      label: '销售佣金', 
      value: stats.value.total_commission,
      currency: stats.value.currency,
      icon: '💸', 
      color: 'accent'
    },
    { 
      label: '服务费', 
      value: stats.value.total_service_fee,
      currency: stats.value.currency,
      icon: '🏷️', 
      color: 'warning'
    },
  ]
})

// 授权统计
const authStats = computed(() => {
  if (!stats.value) return null
  return {
    total: stats.value.total_auths,
    active: stats.value.active_auths
  }
})

// 加载数据
const loadData = async () => {
  loading.value = true
  try {
    const [statsRes, ordersRes] = await Promise.all([
      getDashboardStats(),
      getRecentOrders(8)
    ])
    stats.value = statsRes.data
    recentOrders.value = ordersRes.data || []
  } catch (error) {
    console.error('加载仪表盘数据失败:', error)
  } finally {
    loading.value = false
  }
}

// 跳转到订单详情
const goToOrderDetail = (id: number) => {
  router.push({ name: 'OrderDetail', params: { id } })
}

// 跳转到订单列表
const goToOrders = (status?: string) => {
  if (status) {
    router.push({ name: 'Orders', query: { status } })
  } else {
    router.push({ name: 'Orders' })
  }
}

// 跳转到平台授权
const goToAuths = () => {
  router.push({ name: 'PlatformAuths' })
}

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="dashboard" v-loading="loading">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1>数据概览</h1>
      <p class="page-desc">查看您的订单和经营数据统计</p>
    </div>

    <!-- 订单统计卡片 -->
    <section class="stats-section">
      <h2 class="section-title">
        <span class="title-icon">📊</span>
        订单统计
      </h2>
      <div class="stats-grid">
        <div 
          v-for="stat in statCards" 
          :key="stat.label" 
          class="stat-card"
          :class="`stat-${stat.color}`"
          @click="goToOrders(stat.label === '已签收订单' ? 'delivered' : stat.label === '待处理订单' ? 'accepted' : undefined)"
        >
          <div class="stat-icon">{{ stat.icon }}</div>
          <div class="stat-content">
            <div class="stat-value">
              {{ stat.value?.toLocaleString() || 0 }}
              <span class="stat-suffix">{{ stat.suffix }}</span>
            </div>
            <div class="stat-label">{{ stat.label }}</div>
          </div>
        </div>
      </div>
    </section>

    <!-- 金额统计卡片 -->
    <section class="stats-section">
      <h2 class="section-title">
        <span class="title-icon">💰</span>
        经营数据
      </h2>
      <div class="stats-grid stats-grid-4">
        <div 
          v-for="stat in amountCards" 
          :key="stat.label" 
          class="stat-card amount-card"
          :class="`stat-${stat.color}`"
        >
          <div class="stat-icon">{{ stat.icon }}</div>
          <div class="stat-content">
            <div class="stat-value amount-value">
              {{ formatAmount(stat.value || 0, stat.currency) }}
            </div>
            <div class="stat-label">{{ stat.label }}</div>
          </div>
        </div>
      </div>
    </section>

    <!-- 底部区域：最近订单 + 授权统计 -->
    <div class="bottom-section">
      <!-- 最近订单 -->
      <section class="recent-section">
        <div class="section-header">
          <h2 class="section-title">
            <span class="title-icon">📋</span>
            最近订单
          </h2>
          <button class="view-all-btn" @click="goToOrders()">查看全部 →</button>
        </div>
        <div class="orders-list" v-if="recentOrders.length > 0">
          <div 
            v-for="order in recentOrders" 
            :key="order.id" 
            class="order-item"
            @click="goToOrderDetail(order.id)"
          >
            <div class="order-info">
              <div class="order-no">{{ order.platform_order_no }}</div>
              <div class="order-time">{{ order.order_time ? formatDateTime(order.order_time) : '-' }}</div>
            </div>
            <div class="order-amount">
              {{ formatAmount(order.total_amount, order.currency) }}
            </div>
            <div class="order-status">
              <span 
                class="status-tag" 
                :class="`status-${statusMap[order.status]?.color || 'info'}`"
              >
                {{ statusMap[order.status]?.label || order.status }}
              </span>
            </div>
          </div>
        </div>
        <div class="empty-state" v-else>
          <span class="empty-icon">📭</span>
          <span>暂无订单数据</span>
        </div>
      </section>

      <!-- 授权统计 -->
      <section class="auth-section">
        <div class="section-header">
          <h2 class="section-title">
            <span class="title-icon">🔑</span>
            平台授权
          </h2>
          <button class="view-all-btn" @click="goToAuths()">管理 →</button>
        </div>
        <div class="auth-stats" v-if="authStats">
          <div class="auth-stat-item">
            <div class="auth-stat-value">{{ authStats.total }}</div>
            <div class="auth-stat-label">总授权数</div>
          </div>
          <div class="auth-stat-divider"></div>
          <div class="auth-stat-item">
            <div class="auth-stat-value active">{{ authStats.active }}</div>
            <div class="auth-stat-label">活跃中</div>
          </div>
        </div>
        <div class="auth-tip">
          <span class="tip-icon">💡</span>
          添加平台授权后，可自动同步订单和佣金数据
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped lang="scss">
.dashboard {
  display: flex;
  flex-direction: column;
  gap: 28px;
}

.page-header {
  margin-bottom: 8px;
  
  h1 {
    font-size: 24px;
    font-weight: 600;
    margin-bottom: 8px;
    color: var(--text-primary);
  }
  
  .page-desc {
    font-size: 14px;
    color: var(--text-secondary);
  }
}

.stats-section {
  .section-title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 16px;
    font-weight: 600;
    margin-bottom: 16px;
    color: var(--text-primary);
    
    .title-icon {
      font-size: 18px;
    }
  }
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  
  &.stats-grid-4 {
    grid-template-columns: repeat(4, 1fr);
  }
}

.stat-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: 24px;
  display: flex;
  align-items: center;
  gap: 20px;
  transition: all var(--transition-normal);
  cursor: pointer;
  
  &:hover {
    transform: translateY(-4px);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
    border-color: transparent;
  }
  
  &.amount-card {
    cursor: default;
    
    &:hover {
      transform: none;
    }
  }
}

.stat-icon {
  width: 56px;
  height: 56px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 26px;
  flex-shrink: 0;
}

.stat-success .stat-icon {
  background: rgba(0, 255, 136, 0.1);
}

.stat-primary .stat-icon {
  background: rgba(0, 212, 255, 0.1);
}

.stat-accent .stat-icon {
  background: rgba(255, 215, 0, 0.1);
}

.stat-warning .stat-icon {
  background: rgba(255, 170, 0, 0.1);
}

.stat-danger .stat-icon {
  background: rgba(255, 77, 79, 0.1);
}

.stat-content {
  flex: 1;
  min-width: 0;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  font-family: var(--font-mono);
  color: var(--text-primary);
  display: flex;
  align-items: baseline;
  gap: 4px;
  
  &.amount-value {
    font-size: 20px;
  }
}

.stat-suffix {
  font-size: 14px;
  font-weight: 400;
  color: var(--text-secondary);
}

.stat-label {
  font-size: 14px;
  color: var(--text-secondary);
  margin-top: 4px;
}

.bottom-section {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 24px;
}

.recent-section, .auth-section {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: 24px;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  
  .section-title {
    margin-bottom: 0;
  }
}

.view-all-btn {
  background: none;
  border: none;
  color: var(--color-primary);
  font-size: 14px;
  cursor: pointer;
  transition: all var(--transition-fast);
  
  &:hover {
    opacity: 0.8;
  }
}

.orders-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.order-item {
  display: grid;
  grid-template-columns: 1fr auto auto;
  align-items: center;
  gap: 16px;
  padding: 16px;
  background: var(--bg-secondary);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--transition-fast);
  
  &:hover {
    background: var(--bg-hover);
  }
}

.order-info {
  min-width: 0;
}

.order-no {
  font-family: var(--font-mono);
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.order-time {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 4px;
}

.order-amount {
  font-family: var(--font-mono);
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  white-space: nowrap;
}

.order-status {
  .status-tag {
    display: inline-block;
    padding: 4px 10px;
    border-radius: var(--radius-sm);
    font-size: 12px;
    font-weight: 500;
    
    &.status-success {
      background: rgba(0, 255, 136, 0.1);
      color: var(--color-success);
    }
    
    &.status-primary {
      background: rgba(0, 212, 255, 0.1);
      color: var(--color-primary);
    }
    
    &.status-warning {
      background: rgba(255, 170, 0, 0.1);
      color: var(--color-warning);
    }
    
    &.status-info {
      background: rgba(144, 147, 153, 0.1);
      color: var(--text-secondary);
    }
    
    &.status-danger {
      background: rgba(255, 77, 79, 0.1);
      color: #ff4d4f;
    }
  }
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 24px;
  color: var(--text-muted);
  font-size: 14px;
  gap: 12px;
  
  .empty-icon {
    font-size: 36px;
  }
}

.auth-stats {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 32px;
  padding: 32px 24px;
  background: var(--bg-secondary);
  border-radius: var(--radius-md);
  margin-bottom: 16px;
}

.auth-stat-item {
  text-align: center;
}

.auth-stat-value {
  font-size: 36px;
  font-weight: 700;
  font-family: var(--font-mono);
  color: var(--text-primary);
  
  &.active {
    color: var(--color-success);
  }
}

.auth-stat-label {
  font-size: 13px;
  color: var(--text-secondary);
  margin-top: 4px;
}

.auth-stat-divider {
  width: 1px;
  height: 48px;
  background: var(--border-color);
}

.auth-tip {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 12px 16px;
  background: rgba(0, 212, 255, 0.05);
  border-radius: var(--radius-md);
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.5;
  
  .tip-icon {
    flex-shrink: 0;
  }
}

@media (max-width: 1200px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  
  .bottom-section {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
}
</style>
