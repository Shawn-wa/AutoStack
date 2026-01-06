<script setup lang="ts" name="CashFlow">
import { ref, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import { 
  getCashFlowList, 
  syncCashFlow, 
  getAuths,
  type CashFlowStatement,
  type PlatformAuth
} from '../api'
import { formatCurrency } from '@/utils/format'

// 状态
const loading = ref(false)
const cashFlowList = ref<CashFlowStatement[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)

// 筛选条件
const authId = ref<number | undefined>(undefined)
const authList = ref<PlatformAuth[]>([])

// 同步对话框
const syncDialogVisible = ref(false)
const syncLoading = ref(false)
const syncForm = ref({
  authId: undefined as number | undefined,
  since: '',
  to: ''
})

// 授权选项
const authOptions = computed(() => {
  return authList.value.map(auth => ({
    value: auth.id,
    label: `${auth.shop_name} (${auth.platform})`
  }))
})

// 获取数据
const fetchData = async () => {
  loading.value = true
  try {
    const { data } = await getCashFlowList({
      page: currentPage.value,
      page_size: pageSize.value,
      auth_id: authId.value
    })
    cashFlowList.value = data.list || []
    total.value = data.total
  } catch {
    ElMessage.error('获取现金流报表失败')
  } finally {
    loading.value = false
  }
}

// 获取授权列表
const fetchAuths = async () => {
  try {
    const { data } = await getAuths(1, 100)
    authList.value = data.list || []
  } catch {
    // ignore
  }
}

// 打开同步对话框
const handleOpenSync = () => {
  // 默认同步最近90天
  const now = new Date()
  const since = new Date(now.getTime() - 90 * 24 * 60 * 60 * 1000)
  syncForm.value = {
    authId: authList.value[0]?.id,
    since: since.toISOString().slice(0, 10),
    to: now.toISOString().slice(0, 10)
  }
  syncDialogVisible.value = true
}

// 执行同步
const handleSync = async () => {
  if (!syncForm.value.authId) {
    ElMessage.warning('请选择授权账号')
    return
  }

  syncLoading.value = true
  try {
    const { data } = await syncCashFlow(syncForm.value.authId, {
      since: syncForm.value.since ? `${syncForm.value.since}T00:00:00Z` : undefined,
      to: syncForm.value.to ? `${syncForm.value.to}T23:59:59Z` : undefined
    })
    ElMessage.success(`同步成功：共${data.total}条，新增${data.created}条，更新${data.updated}条`)
    syncDialogVisible.value = false
    fetchData()
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.message || '同步失败')
  } finally {
    syncLoading.value = false
  }
}

// 分页
const handlePageChange = (page: number) => {
  currentPage.value = page
  fetchData()
}

const handleSizeChange = (size: number) => {
  pageSize.value = size
  currentPage.value = 1
  fetchData()
}

// 筛选变化
const handleFilterChange = () => {
  currentPage.value = 1
  fetchData()
}

// 获取店铺名称
const getShopName = (authId: number) => {
  const auth = authList.value.find(a => a.id === authId)
  return auth ? auth.shop_name : '-'
}

// 格式化周期
const formatPeriod = (begin: string | null, end: string | null) => {
  if (!begin || !end) return '-'
  return `${begin.slice(0, 10)} ~ ${end.slice(0, 10)}`
}

// 计算净利润
const calcNetProfit = (row: CashFlowStatement) => {
  return row.orders_amount + row.returns_amount + row.commission_amount + 
         row.services_amount + row.item_delivery_and_return_amount
}

onMounted(() => {
  fetchAuths()
  fetchData()
})
</script>

<template>
  <div class="cash-flow-page">
    <div class="page-header">
      <h2>报表</h2>
      <div class="header-actions">
        <el-button type="primary" :icon="Refresh" @click="handleOpenSync">
          同步报表
        </el-button>
      </div>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar">
      <el-select 
        v-model="authId" 
        placeholder="选择店铺" 
        clearable 
        @change="handleFilterChange"
        style="width: 200px"
      >
        <el-option
          v-for="opt in authOptions"
          :key="opt.value"
          :label="opt.label"
          :value="opt.value"
        />
      </el-select>
    </div>

    <!-- 数据表格 -->
    <el-table :data="cashFlowList" v-loading="loading" border stripe>
      <el-table-column label="店铺" min-width="120">
        <template #default="{ row }">
          {{ getShopName(row.platform_auth_id) }}
        </template>
      </el-table-column>
      <el-table-column label="结算周期" min-width="180">
        <template #default="{ row }">
          {{ formatPeriod(row.period_begin, row.period_end) }}
        </template>
      </el-table-column>
      <el-table-column label="货币" prop="currency_code" width="80" align="center" />
      <el-table-column label="订单金额" min-width="120" align="right">
        <template #default="{ row }">
          {{ formatCurrency(row.orders_amount, row.currency_code) }}
        </template>
      </el-table-column>
      <el-table-column label="退货金额" min-width="100" align="right">
        <template #default="{ row }">
          {{ formatCurrency(row.returns_amount, row.currency_code) }}
        </template>
      </el-table-column>
      <el-table-column label="佣金" min-width="100" align="right">
        <template #default="{ row }">
          <span :class="{ 'text-danger': row.commission_amount < 0 }">
            {{ formatCurrency(row.commission_amount, row.currency_code) }}
          </span>
        </template>
      </el-table-column>
      <el-table-column label="服务费" min-width="100" align="right">
        <template #default="{ row }">
          <span :class="{ 'text-danger': row.services_amount < 0 }">
            {{ formatCurrency(row.services_amount, row.currency_code) }}
          </span>
        </template>
      </el-table-column>
      <el-table-column label="配送/退货费" min-width="110" align="right">
        <template #default="{ row }">
          <span :class="{ 'text-danger': row.item_delivery_and_return_amount < 0 }">
            {{ formatCurrency(row.item_delivery_and_return_amount, row.currency_code) }}
          </span>
        </template>
      </el-table-column>
      <el-table-column label="净利润" min-width="120" align="right">
        <template #default="{ row }">
          <span :class="{ 'text-success': calcNetProfit(row) > 0, 'text-danger': calcNetProfit(row) < 0 }">
            <strong>{{ formatCurrency(calcNetProfit(row), row.currency_code) }}</strong>
          </span>
        </template>
      </el-table-column>
      <el-table-column label="同步时间" min-width="160">
        <template #default="{ row }">
          {{ row.synced_at || '-' }}
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div class="pagination-wrapper">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="total"
        layout="total, sizes, prev, pager, next, jumper"
        @current-change="handlePageChange"
        @size-change="handleSizeChange"
        background
      />
    </div>

    <!-- 同步对话框 -->
    <el-dialog v-model="syncDialogVisible" title="同步报表" width="400px">
      <el-form :model="syncForm" label-width="80px">
        <el-form-item label="授权账号" required>
          <el-select v-model="syncForm.authId" placeholder="选择账号" style="width: 100%">
            <el-option
              v-for="opt in authOptions"
              :key="opt.value"
              :label="opt.label"
              :value="opt.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="开始日期">
          <el-date-picker
            v-model="syncForm.since"
            type="date"
            placeholder="选择开始日期"
            value-format="YYYY-MM-DD"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="结束日期">
          <el-date-picker
            v-model="syncForm.to"
            type="date"
            placeholder="选择结束日期"
            value-format="YYYY-MM-DD"
            style="width: 100%"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="syncDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="syncLoading" @click="handleSync">
          开始同步
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped lang="scss">
.cash-flow-page {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;

  h2 {
    margin: 0;
    font-size: 20px;
    font-weight: 600;
  }
}

.filter-bar {
  display: flex;
  gap: 16px;
  margin-bottom: 16px;
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

.text-success {
  color: #67c23a;
}

.text-danger {
  color: #f56c6c;
}
</style>
