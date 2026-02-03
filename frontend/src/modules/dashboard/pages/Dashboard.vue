<script setup lang="ts">
import { ref, onMounted, onActivated, computed, watch, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getDashboardStats, getRecentOrders, getOrderTrend, initDashboardStats, refreshDashboardStats, type DashboardStats, type RecentOrder, type OrderTrendItem } from '@/modules/order/api'
import { formatCurrency, formatDateTime } from '@/utils/format'
import { cacheStore } from '@/utils/storage'
import * as echarts from 'echarts'

const STATS_CACHE_KEY = 'dashboard_stats'
const STATS_CACHE_MINUTES = 30 // 缓存30分钟

defineOptions({ name: 'Dashboard' })

const router = useRouter()
const loading = ref(false)
const stats = ref<DashboardStats | null>(null)
const recentOrders = ref<RecentOrder[]>([])
const trendData = ref<OrderTrendItem[]>([])
const chartRef = ref<HTMLElement | null>(null)
const trendRefreshing = ref(false)
const currentTrendCurrency = ref('RUB') // 默认币种

// 可选币种（排除CNY）
const availableCurrencies = computed(() => {
  if (!stats.value || !stats.value.total_amounts || stats.value.total_amounts.length === 0) {
    return ['RUB']
  }
  return stats.value.total_amounts.map(a => a.currency).filter(c => c !== 'CNY')
})

let chartInstance: echarts.ECharts | null = null

// 状态映射
const statusMap: Record<string, { label: string; color: string }> = {
  pending: { label: '待处理', color: 'warning' },
  ready_to_ship: { label: '待发货', color: 'primary' },
  shipped: { label: '已发货', color: 'cyan' },
  delivered: { label: '已签收', color: 'success' },
  cancelled: { label: '已取消', color: 'danger' },
}

// 格式化金额
const formatAmount = (value: number, currency: string = 'RUB') => {
  return formatCurrency(value, currency)
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
      suffix: '单',
      filter: {} // 无过滤条件，显示全部
    },
    { 
      label: '已签收订单', 
      value: stats.value.delivered_orders, 
      icon: '✅', 
      color: 'success',
      suffix: '单',
      filter: { status: 'delivered' }
    },
    { 
      label: '已发货', 
      value: stats.value.shipped_orders, 
      icon: '🚚', 
      color: 'cyan',
      suffix: '单',
      filter: { status: 'shipped' }
    },
    { 
      label: '今日新增', 
      value: stats.value.today_orders, 
      icon: '📈', 
      color: 'accent',
      suffix: '单',
      filter: { start_time: getTodayStart(), end_time: getTodayEnd() }
    },
    { 
      label: '待处理订单', 
      value: stats.value.pending_orders, 
      icon: '⏳', 
      color: 'warning',
      suffix: '单',
      filter: { status: 'pending,ready_to_ship' } // 待处理+待发货
    },
    { 
      label: '即将超时', 
      value: stats.value.timeout_orders, 
      icon: '⚠️', 
      color: 'danger',
      suffix: '单',
      filter: { status: 'pending,ready_to_ship' } // 跳转到待处理订单列表
    },
  ]
})

// 金额统计卡片
const amountCards = computed(() => {
  if (!stats.value) return []
  
  // 处理订单总金额（多币种）
  let totalAmountValue: any = 0
  let isMulti = false
  
  if (stats.value.total_amounts && stats.value.total_amounts.length > 0) {
    const validAmounts = stats.value.total_amounts.filter(a => a.amount > 0)
    if (validAmounts.length > 0) {
      totalAmountValue = validAmounts
      isMulti = true
    }
  }

  return [
    { 
      label: '订单总金额', 
      value: totalAmountValue,
      currency: stats.value.currency,
      icon: '💰', 
      color: 'primary',
      isMultiCurrency: isMulti
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
    // 先触发初始化统计（异步，不阻塞）
    initDashboardStats().catch(() => {})
    
    // 尝试从缓存获取统计数据
    const cachedStats = cacheStore.get<DashboardStats>(STATS_CACHE_KEY)
    
    let statsData: DashboardStats | null = null
    
    if (cachedStats) {
      // 使用缓存数据
      statsData = cachedStats
    } else {
      // 缓存不存在或已过期，请求接口
      const statsRes = await getDashboardStats()
      statsData = statsRes.data
      // 存入缓存（30分钟过期，关闭浏览器也会失效）
      cacheStore.set(STATS_CACHE_KEY, statsData, STATS_CACHE_MINUTES)
    }
    
    stats.value = statsData
    
    // 其他数据不缓存，每次请求
    const [ordersRes, trendRes] = await Promise.all([
      getRecentOrders(8),
      getOrderTrend(30, currentTrendCurrency.value)
    ])
    recentOrders.value = ordersRes.data || []
    trendData.value = trendRes.data?.items || []
    
    // 如果返回的统计中有币种，且当前默认的RUB不在其中，则切换到第一个币种
    if (stats.value.total_amounts && stats.value.total_amounts.length > 0) {
      const hasCurrent = stats.value.total_amounts.some(a => a.currency === currentTrendCurrency.value)
      if (!hasCurrent) {
        currentTrendCurrency.value = stats.value.total_amounts[0].currency
        // 重新加载走势
        const newTrendRes = await getOrderTrend(30, currentTrendCurrency.value)
        trendData.value = newTrendRes.data?.items || []
      }
    }

    // 初始化图表
    initChart()
  } catch (error) {
    console.error('加载仪表盘数据失败:', error)
  } finally {
    loading.value = false
  }
}

// 切换走势图币种
const changeTrendCurrency = async (currency: string) => {
  if (currentTrendCurrency.value === currency) return
  currentTrendCurrency.value = currency
  trendRefreshing.value = true
  try {
    const trendRes = await getOrderTrend(30, currency)
    trendData.value = trendRes.data?.items || []
    initChart()
  } catch (error) {
    console.error('切换币种失败:', error)
  } finally {
    trendRefreshing.value = false
  }
}

// 刷新走势统计
const refreshTrendStats = async () => {
  trendRefreshing.value = true
  try {
    // 清除统计缓存
    cacheStore.remove(STATS_CACHE_KEY)
    
    // 强制刷新统计数据
    await refreshDashboardStats()
    
    // 重新请求统计数据并缓存
    const statsRes = await getDashboardStats()
    stats.value = statsRes.data
    cacheStore.set(STATS_CACHE_KEY, statsRes.data, STATS_CACHE_MINUTES)
    
    // 重新加载走势数据
    const trendRes = await getOrderTrend(30, currentTrendCurrency.value)
    trendData.value = trendRes.data?.items || []
    initChart()
    ElMessage.success('数据已更新')
  } catch (error) {
    console.error('刷新数据失败:', error)
    ElMessage.error('刷新失败，请稍后重试')
  } finally {
    trendRefreshing.value = false
  }
}

// 初始化图表
const initChart = () => {
  if (!chartRef.value || trendData.value.length === 0) return
  
  if (chartInstance) {
    chartInstance.dispose()
  }
  
  chartInstance = echarts.init(chartRef.value)
  
  const dates = trendData.value.map(item => {
    const d = new Date(item.date)
    return `${d.getMonth() + 1}/${d.getDate()}`
  })
  const counts = trendData.value.map(item => item.count)
  const amounts = trendData.value.map(item => item.amount)
  
  const option: echarts.EChartsOption = {
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(30, 30, 30, 0.95)',
      borderColor: '#333',
      borderWidth: 1,
      textStyle: {
        color: '#fff'
      },
      formatter: (params: any) => {
        const index = params[0].dataIndex
        const item = trendData.value[index]
        return `<div style="padding: 4px;">
          <div style="color: #999; margin-bottom: 6px; font-size: 12px;">${item.date}</div>
          <div style="display: flex; justify-content: space-between; gap: 16px; margin-bottom: 4px;">
            <span style="color: #00d4ff;">● 订单量</span>
            <span style="font-weight: bold;">${item.count}</span>
          </div>
          <div style="display: flex; justify-content: space-between; gap: 16px;">
            <span style="color: #ffd700;">● 销售额</span>
            <span style="font-weight: bold;">₽${item.amount.toLocaleString()}</span>
          </div>
        </div>`
      }
    },
    legend: {
      data: ['订单量', '销售额'],
      bottom: 0,
      textStyle: { color: 'rgba(255, 255, 255, 0.6)', fontSize: 11 },
      itemWidth: 12,
      itemHeight: 12
    },
    grid: {
      left: 40,
      right: 60,
      top: 30,
      bottom: 30
    },
    xAxis: {
      type: 'category',
      data: dates,
      axisLine: {
        show: true,
        lineStyle: { color: 'rgba(255,255,255,0.2)' }
      },
      axisLabel: {
        color: 'rgba(255,255,255,0.5)',
        fontSize: 10,
        interval: Math.floor(dates.length / 10)
      },
      axisTick: { show: false },
      splitLine: {
        show: true,
        lineStyle: { 
          color: 'rgba(255,255,255,0.06)',
          type: 'dashed'
        }
      }
    },
    yAxis: [
      {
        type: 'value',
        name: '订单量',
        nameTextStyle: { color: 'rgba(255,255,255,0.4)', padding: [0, 20, 0, 0] },
        splitNumber: 5,
        splitLine: {
          show: true,
          lineStyle: { 
            color: 'rgba(255,255,255,0.08)',
            type: 'dashed'
          }
        },
        axisLine: { 
          show: true,
          lineStyle: { color: 'rgba(255,255,255,0.1)' }
        },
        axisLabel: {
          color: 'rgba(255,255,255,0.5)',
          fontSize: 11
        }
      },
      {
        type: 'value',
        name: '销售额',
        nameTextStyle: { color: 'rgba(255,255,255,0.4)', padding: [0, 0, 0, 20] },
        splitLine: { show: false },
        axisLine: { 
          show: true,
          lineStyle: { color: 'rgba(255,255,255,0.1)' }
        },
        axisLabel: {
          color: 'rgba(255,255,255,0.5)',
          fontSize: 10,
          formatter: (value: number) => value >= 1000 ? `${(value/1000).toFixed(0)}k` : `${value}`
        }
      }
    ],
    series: [
      {
        name: '订单量',
        data: counts,
        type: 'line',
        smooth: true,
        symbol: 'none',
        yAxisIndex: 0,
        lineStyle: {
          color: '#00d4ff',
          width: 2
        },
        itemStyle: {
          color: '#00d4ff'
        },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(0, 212, 255, 0.25)' },
            { offset: 1, color: 'rgba(0, 212, 255, 0.01)' }
          ])
        }
      },
      {
        name: '销售额',
        data: amounts,
        type: 'line',
        smooth: true,
        symbol: 'none',
        yAxisIndex: 1,
        lineStyle: {
          color: '#ffd700',
          width: 2
        },
        itemStyle: {
          color: '#ffd700'
        },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(255, 215, 0, 0.25)' },
            { offset: 1, color: 'rgba(255, 215, 0, 0.01)' }
          ])
        }
      }
    ]
  }
  
  chartInstance.setOption(option)
}

// 窗口resize时重绘图表
const handleResize = () => {
  chartInstance?.resize()
}

// 监听主题变化
watch(() => trendData.value, () => {
  if (chartRef.value) {
    initChart()
  }
})

// 跳转到订单详情
const goToOrderDetail = (id: number) => {
  router.push({ name: 'OrderDetail', params: { id } })
}

// 获取今日开始时间 YYYY-MM-DD 00:00:00
const getTodayStart = () => {
  const today = new Date()
  return today.toISOString().split('T')[0] + ' 00:00:00'
}

// 获取今日结束时间 YYYY-MM-DD 23:59:59
const getTodayEnd = () => {
  const today = new Date()
  return today.toISOString().split('T')[0] + ' 23:59:59'
}

// 跳转到订单列表
const goToOrders = (filter?: { status?: string; start_time?: string; end_time?: string }) => {
  if (filter && Object.keys(filter).length > 0) {
    router.push({ name: 'Orders', query: filter })
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
  window.addEventListener('resize', handleResize)
})

// keep-alive 激活时重新加载数据
onActivated(() => {
  loadData()
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  if (chartInstance) {
    chartInstance.dispose()
    chartInstance = null
  }
})
</script>

<template>
  <div class="dashboard" v-loading="loading">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1>数据概览</h1>
      <p class="page-desc">查看您的订单和经营数据统计</p>
    </div>

    <!-- 订单统计 + 经营数据 并排 -->
    <div class="stats-row">
      <!-- 订单统计卡片 -->
      <section class="stats-section stats-section-half">
        <h2 class="section-title">
          <span class="title-icon">📊</span>
          订单统计
        </h2>
        <div class="stats-grid stats-grid-3">
          <div 
            v-for="stat in statCards" 
            :key="stat.label" 
            class="stat-card stat-card-compact"
            :class="`stat-${stat.color}`"
            @click="goToOrders(stat.filter)"
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
      <section class="stats-section stats-section-half">
        <h2 class="section-title">
          <span class="title-icon">💰</span>
          经营数据
        </h2>
        <div class="stats-grid stats-grid-2">
          <div 
            v-for="stat in amountCards" 
            :key="stat.label" 
            class="stat-card stat-card-compact amount-card"
            :class="`stat-${stat.color}`"
          >
            <div class="stat-icon">{{ stat.icon }}</div>
            <div class="stat-content">
              <div class="stat-value amount-value" :class="{ 'multi-value': stat.isMultiCurrency }">
                <template v-if="stat.isMultiCurrency">
                  <div v-for="item in stat.value" :key="item.currency" class="multi-currency-item">
                    {{ formatAmount(item.amount, item.currency) }}
                  </div>
                </template>
                <template v-else>
                  {{ formatAmount(stat.value || 0, stat.currency) }}
                </template>
              </div>
              <div class="stat-label">{{ stat.label }}</div>
            </div>
          </div>
        </div>
      </section>
    </div>

    <!-- 底部区域：订单走势 + 最近订单 + 授权统计 -->
    <div class="bottom-section">
      <!-- 订单走势图 -->
      <section class="trend-section">
        <div class="section-header">
          <h2 class="section-title">
            <span class="title-icon">📈</span>
            订单走势
          </h2>
          <div class="trend-actions">
            <div class="trend-tabs">
              <button 
                v-for="curr in availableCurrencies" 
                :key="curr"
                class="trend-tab"
                :class="{ active: currentTrendCurrency === curr }"
                @click="changeTrendCurrency(curr)"
              >
                {{ curr }}
              </button>
            </div>
            <span class="trend-period">近30天</span>
            <button class="refresh-btn" @click="refreshTrendStats" :disabled="trendRefreshing">
              <span class="refresh-icon" :class="{ 'spinning': trendRefreshing }">🔄</span>
            </button>
          </div>
        </div>
        <div class="trend-chart" ref="chartRef"></div>
      </section>

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
          >
            <div class="order-info">
              <div class="order-no" @click="goToOrderDetail(order.id)">{{ order.platform_order_no }}</div>
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

.stats-row {
  display: flex;
  gap: 24px;
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
  
  &.stats-section-half {
    flex: 1;
    min-width: 0;
    
    &:first-child {
      flex: 1.5; // 订单统计区域占更大比例
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
  
  &.stats-grid-2 {
    grid-template-columns: repeat(2, 1fr);
    gap: 12px;
  }
  
  &.stats-grid-3 {
    grid-template-columns: repeat(3, 1fr);
    gap: 12px;
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
  
  &.stat-card-compact {
    padding: 16px;
    gap: 12px;
    
    &:hover {
      transform: translateY(-2px);
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
  
  .stat-card-compact & {
    width: 44px;
    height: 44px;
    font-size: 20px;
  }
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

  &.multi-value {
    flex-direction: column;
    align-items: flex-start;
    gap: 2px;
  }
  
  .multi-currency-item {
    font-size: 18px;
    line-height: 1.4;
  }
  
  &.amount-value {
    font-size: 20px;
  }
  
  .stat-card-compact & {
    font-size: 22px;
    
    &.amount-value {
      font-size: 16px;
    }
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
  grid-template-columns: 1fr 1fr 1fr;
  gap: 24px;
}

.trend-section, .recent-section, .auth-section {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: 24px;
}

.trend-section {
  .trend-actions {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  
  .trend-tabs {
    display: flex;
    background: var(--bg-secondary);
    border-radius: var(--radius-sm);
    padding: 2px;
    margin-right: 8px;
    
    .trend-tab {
      background: none;
      border: none;
      padding: 2px 8px;
      font-size: 12px;
      color: var(--text-secondary);
      cursor: pointer;
      border-radius: var(--radius-xs);
      transition: all 0.2s;
      
      &:hover {
        color: var(--text-primary);
      }
      
      &.active {
        background: var(--bg-card);
        color: var(--color-primary);
        font-weight: 500;
        box-shadow: 0 1px 2px rgba(0,0,0,0.1);
      }
    }
  }
  
  .trend-period {
    font-size: 12px;
    color: var(--text-muted);
    background: var(--bg-secondary);
    padding: 4px 8px;
    border-radius: var(--radius-sm);
  }
  
  .refresh-btn {
    background: var(--bg-secondary);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-sm);
    padding: 4px 8px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.2s;
    
    &:hover:not(:disabled) {
      background: var(--color-primary);
      border-color: var(--color-primary);
      .refresh-icon {
        filter: brightness(10);
      }
    }
    
    &:disabled {
      cursor: not-allowed;
      opacity: 0.6;
    }
    
    .refresh-icon {
      font-size: 14px;
      display: inline-block;
      
      &.spinning {
        animation: spin 1s linear infinite;
      }
    }
  }
  
  @keyframes spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
  }
  
  .trend-chart {
    height: 220px;
    margin-top: 8px;
  }
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
  transition: all var(--transition-fast);
}

.order-info {
  min-width: 0;
}

.order-no {
  display: inline-block;
  font-family: var(--font-mono);
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
  cursor: pointer;
  transition: color var(--transition-fast);
  
  &:hover {
    color: var(--accent);
  }
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
    
    &.status-cyan {
      background: rgba(0, 206, 209, 0.1);
      color: #00ced1;
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

@media (max-width: 1400px) {
  .bottom-section {
    grid-template-columns: 1fr 1fr;
    
    .auth-section {
      grid-column: span 2;
    }
  }
}

@media (max-width: 1200px) {
  .stats-row {
    flex-direction: column;
  }
  
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
    
    &.stats-grid-2 {
      grid-template-columns: repeat(2, 1fr);
    }
    
    &.stats-grid-3 {
      grid-template-columns: repeat(3, 1fr);
    }
  }
  
  .bottom-section {
    grid-template-columns: 1fr;
    
    .auth-section {
      grid-column: span 1;
    }
  }
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr;
    
    &.stats-grid-2 {
      grid-template-columns: 1fr;
    }
    
    &.stats-grid-3 {
      grid-template-columns: repeat(2, 1fr);
    }
  }
}

@media (max-width: 480px) {
  .stats-grid {
    &.stats-grid-3 {
      grid-template-columns: 1fr;
    }
  }
}
</style>
