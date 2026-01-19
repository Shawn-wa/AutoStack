<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api, { type OrderSummaryItem } from '../api'
import { getAuths, type AuthResponse } from '@/modules/order/api'

defineOptions({ name: 'OrderSummary' })

const loading = ref(false)
const tableData = ref<OrderSummaryItem[]>([])
const authId = ref<number | undefined>(undefined)
const authOptions = ref<AuthResponse[]>([])
const dateRange = ref<[Date, Date] | null>(null)

// 默认本月
const defaultDateRange = () => {
  const end = new Date()
  const start = new Date()
  start.setDate(1) // 本月1号
  dateRange.value = [start, end]
}

// 获取授权列表
const fetchAuths = async () => {
  try {
    const res = await getAuths()
    authOptions.value = res.data.list || []
    if (authOptions.value.length > 0) {
      // 默认选中第一个
      // authId.value = authOptions.value[0].id
    }
    fetchSummary()
  } catch (error) {
    console.error('获取授权列表失败', error)
  }
}

// 获取汇总数据
const fetchSummary = async () => {
  loading.value = true
  try {
    const params: any = {}
    if (authId.value) {
      params.auth_id = authId.value
    }
    if (dateRange.value) {
      params.start_time = formatDate(dateRange.value[0])
      params.end_time = formatDate(dateRange.value[1])
    }

    const res = await api.getOrderSummary(params)
    tableData.value = res.data || []
  } catch (error) {
    console.error('获取订单汇总失败', error)
  } finally {
    loading.value = false
  }
}

const formatDate = (date: Date) => {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

// 获取状态标签类型
const getStatusType = (status: string) => {
  const typeMap: Record<string, string> = {
    'shipped': 'success',
    'delivered': 'success',
    'cancelled': 'danger',
    'pending': 'warning',
    'ready_to_ship': 'primary',
  }
  return typeMap[status] || 'info'
}

onMounted(() => {
  defaultDateRange()
  fetchAuths()
})
</script>

<template>
  <div class="order-summary">
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">订单汇总</h1>
        <p class="page-desc">按SKU统计销售数据</p>
      </div>
    </div>

    <div class="filter-card">
      <el-form inline>
        <el-form-item label="店铺">
          <el-select v-model="authId" placeholder="全部店铺" clearable style="width: 200px" @change="fetchSummary">
            <el-option
              v-for="item in authOptions"
              :key="item.id"
              :label="`${item.platform} - ${item.shop_name}`"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="日期范围">
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            @change="fetchSummary"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchSummary">查询</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="content-card">
      <el-table
        v-loading="loading"
        :data="tableData"
        style="width: 100%"
        stripe
        row-key="local_sku"
      >
        <el-table-column type="expand">
          <template #default="{ row }">
            <div class="expand-content">
              <div class="expand-section">
                <h4>平台SKU</h4>
                <div class="platform-skus">
                  <el-tag v-for="sku in row.platform_skus" :key="sku" type="info" class="sku-tag">
                    {{ sku }}
                  </el-tag>
                </div>
              </div>
              <div class="expand-section">
                <h4>各状态明细</h4>
                <el-table :data="row.status_details" size="small" border>
                  <el-table-column prop="status" label="订单状态" width="120">
                    <template #default="{ row: detail }">
                      <el-tag :type="getStatusType(detail.status)" size="small">
                        {{ detail.status }}
                      </el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column prop="quantity" label="销量" width="100" />
                  <el-table-column prop="amount" label="销售额" width="150">
                    <template #default="{ row: detail }">
                      {{ detail.amount.toFixed(2) }} {{ row.currency }}
                    </template>
                  </el-table-column>
                </el-table>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="local_sku" label="本地SKU" width="150">
          <template #default="{ row }">
            <el-tag v-if="row.local_sku" type="success">{{ row.local_sku }}</el-tag>
            <span v-else class="text-secondary">未关联</span>
          </template>
        </el-table-column>
        <el-table-column prop="product_name" label="本地名称" min-width="150" show-overflow-tooltip />
        <el-table-column prop="platform_skus" label="平台SKU" width="200">
          <template #default="{ row }">
            <span v-if="row.platform_skus?.length === 1">{{ row.platform_skus[0] }}</span>
            <el-tooltip v-else-if="row.platform_skus?.length > 1" :content="row.platform_skus.join(', ')" placement="top">
              <span>{{ row.platform_skus[0] }} <el-tag size="small" type="info">+{{ row.platform_skus.length - 1 }}</el-tag></span>
            </el-tooltip>
            <span v-else class="text-secondary">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="quantity" label="总销量" width="100" sortable />
        <el-table-column prop="amount" label="总销售额" width="150" sortable>
          <template #default="{ row }">
            {{ row.amount.toFixed(2) }} {{ row.currency }}
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<style scoped lang="scss">
.order-summary {
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

.filter-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: 24px 24px 0;
  margin-bottom: 24px;
}

.content-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: 24px;
}

.text-secondary {
  color: var(--text-secondary);
}

.expand-content {
  padding: 16px 24px;
  background: var(--bg-page);
  border-radius: var(--radius-md);
}

.expand-section {
  margin-bottom: 16px;

  &:last-child {
    margin-bottom: 0;
  }

  h4 {
    margin: 0 0 12px 0;
    font-size: 14px;
    font-weight: 500;
    color: var(--text-primary);
  }
}

.platform-skus {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.sku-tag {
  font-family: monospace;
}
</style>
