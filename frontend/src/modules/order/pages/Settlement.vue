<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElDatePicker } from 'element-plus'
import { getAuths, type PlatformAuth } from '@/modules/order/api'
import request from '@/commonBase/api/request'

defineOptions({ name: 'Settlement' })

const loading = ref(false)
const auths = ref<PlatformAuth[]>([])
const selectedAuthId = ref<number | undefined>(undefined)
const selectedMonth = ref<string | null>(null)
const settlementData = ref<any>(null)

// 获取授权列表
const fetchAuths = async () => {
  try {
    const res = await getAuths(1, 100)
    auths.value = res.data.list.filter((a: PlatformAuth) => a.platform === 'ozon')
    if (auths.value.length > 0) {
      selectedAuthId.value = auths.value[0].id
    }
  } catch (error) {
    console.error('获取授权列表失败', error)
  }
}

// 获取结算报告
const fetchSettlement = async () => {
  if (!selectedAuthId.value) {
    ElMessage.warning('请选择店铺')
    return
  }

  loading.value = true
  settlementData.value = null

  try {
    const params: any = {}
    if (selectedMonth.value) {
      // selectedMonth 格式为 "YYYY-MM"
      const [year, month] = selectedMonth.value.split('-').map(Number)
      // 使用选中月份的最后一天作为查询日期
      const lastDayOfMonth = new Date(year, month, 0)
      params.to = lastDayOfMonth.toISOString()
    }

    const res = await request.post(`/order/auths/${selectedAuthId.value}/mutual-settlement`, params)
    settlementData.value = res.data
    ElMessage.success('获取成功')
  } catch (error: any) {
    console.error('获取结算报告失败', error)
    const errorMsg = error?.message || error?.response?.data?.message || '获取失败'
    if (errorMsg.includes('NotFound') || errorMsg.includes('not found')) {
      ElMessage.warning('该月份暂无结算报告数据，请尝试其他月份')
    } else {
      ElMessage.error(errorMsg)
    }
  } finally {
    loading.value = false
  }
}

// 格式化金额
const formatAmount = (value: number | undefined, currency: string = 'RUB') => {
  if (value === undefined || value === null) return '-'
  const symbols: Record<string, string> = { RUB: '₽', USD: '$', CNY: '¥', EUR: '€' }
  const symbol = symbols[currency] || currency
  return `${value.toLocaleString('ru-RU', { minimumFractionDigits: 2 })} ${symbol}`
}

onMounted(() => {
  fetchAuths()
  // 默认选择当月
  const now = new Date()
  selectedMonth.value = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
})
</script>

<template>
  <div class="settlement-page">
    <div class="page-header">
      <h1 class="page-title">结算报告</h1>
      <p class="page-desc">查看 Ozon 平台的结算余额信息</p>
    </div>

    <div class="content-card">
      <!-- 筛选区域 -->
      <div class="filter-area">
        <div class="filter-row">
          <div class="filter-item">
            <label>店铺</label>
            <el-select v-model="selectedAuthId" placeholder="选择店铺" style="width: 200px">
              <el-option
                v-for="auth in auths"
                :key="auth.id"
                :label="auth.shop_name"
                :value="auth.id"
              />
            </el-select>
          </div>
          <div class="filter-item">
            <label>月份</label>
            <el-date-picker
              v-model="selectedMonth"
              type="month"
              placeholder="选择月份"
              format="YYYY年MM月"
              value-format="YYYY-MM"
              style="width: 160px"
            />
          </div>
          <el-button type="primary" :loading="loading" @click="fetchSettlement">
            查询
          </el-button>
        </div>
      </div>

      <!-- 结算数据展示 -->
      <div class="settlement-content" v-loading="loading">
        <template v-if="settlementData">
          <!-- 原始响应 -->
          <div class="data-section">
            <h3>API 响应数据</h3>
            <pre class="json-display">{{ JSON.stringify(settlementData, null, 2) }}</pre>
          </div>

          <!-- 如果有 details 字段，展示余额卡片 -->
          <template v-if="settlementData.result?.details?.length > 0">
            <div class="balance-cards">
              <div 
                v-for="(detail, index) in settlementData.result.details" 
                :key="index"
                class="balance-card"
              >
                <div class="balance-main">
                  <div class="balance-value">{{ formatAmount(detail.balance, detail.currency_code) }}</div>
                  <div class="balance-label">当前余额</div>
                </div>
                <div class="balance-details">
                  <div class="detail-item">
                    <span class="detail-label">期初余额</span>
                    <span class="detail-value">{{ formatAmount(detail.balance_at_beginning, detail.currency_code) }}</span>
                  </div>
                  <div class="detail-item">
                    <span class="detail-label">应计金额</span>
                    <span class="detail-value positive">{{ formatAmount(detail.accrued_amount, detail.currency_code) }}</span>
                  </div>
                  <div class="detail-item">
                    <span class="detail-label">已支付</span>
                    <span class="detail-value">{{ formatAmount(detail.paid_amount, detail.currency_code) }}</span>
                  </div>
                </div>
              </div>
            </div>
          </template>
        </template>

        <div v-else class="empty-state">
          <span class="empty-icon">📊</span>
          <span>请选择店铺和时间范围后查询</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.settlement-page {
  padding: 0;
}

.page-header {
  margin-bottom: 24px;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  margin: 0 0 8px 0;
}

.page-desc {
  color: var(--text-secondary);
  margin: 0;
  font-size: 14px;
}

.content-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: 24px;
}

.filter-area {
  margin-bottom: 24px;
  padding-bottom: 24px;
  border-bottom: 1px solid var(--border-color);
}

.filter-row {
  display: flex;
  align-items: flex-end;
  gap: 16px;
  flex-wrap: wrap;
}

.filter-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
  
  label {
    font-size: 14px;
    color: var(--text-secondary);
  }
}

.settlement-content {
  min-height: 300px;
}

.data-section {
  margin-bottom: 24px;
  
  h3 {
    font-size: 16px;
    font-weight: 600;
    margin-bottom: 12px;
    color: var(--text-primary);
  }
}

.json-display {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 16px;
  font-family: var(--font-mono);
  font-size: 13px;
  color: var(--text-primary);
  overflow-x: auto;
  max-height: 400px;
  overflow-y: auto;
  white-space: pre-wrap;
  word-break: break-all;
}

.balance-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 20px;
}

.balance-card {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: 24px;
}

.balance-main {
  text-align: center;
  margin-bottom: 20px;
  padding-bottom: 20px;
  border-bottom: 1px solid var(--border-color);
}

.balance-value {
  font-size: 32px;
  font-weight: 700;
  font-family: var(--font-mono);
  color: var(--color-primary);
}

.balance-label {
  font-size: 14px;
  color: var(--text-secondary);
  margin-top: 4px;
}

.balance-details {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.detail-label {
  font-size: 14px;
  color: var(--text-secondary);
}

.detail-value {
  font-size: 16px;
  font-weight: 600;
  font-family: var(--font-mono);
  
  &.positive {
    color: var(--color-success);
  }
  
  &.negative {
    color: var(--color-danger);
  }
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 24px;
  color: var(--text-muted);
  font-size: 14px;
  gap: 12px;
  
  .empty-icon {
    font-size: 48px;
  }
}
</style>

