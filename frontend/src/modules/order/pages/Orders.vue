<script setup lang="ts">
import { ref, onMounted, watch, onActivated } from 'vue'
import { useRouter, useRoute } from 'vue-router'

defineOptions({ name: 'Orders' })
import { Search, View, CopyDocument } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import {
  getOrders,
  getPlatforms,
  getAuths,
  syncOrders,
  type Order,
  type PlatformInfo,
  type PlatformAuth
} from '@/modules/order/api'
import { formatDateTime } from '@/utils/format'

const router = useRouter()
const route = useRoute()

const loading = ref(false)
const tableData = ref<Order[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)

// 筛选条件
const filters = ref({
  platform: '',
  auth_id: undefined as number | undefined,
  status: '',
  keyword: '',
  start_time: '',
  end_time: ''
})

// 日期范围选择器值
const dateRange = ref<[Date, Date] | null>(null)

// 重置筛选条件
const resetFilters = () => {
  filters.value = {
    platform: '',
    auth_id: undefined,
    status: '',
    keyword: '',
    start_time: '',
    end_time: ''
  }
  dateRange.value = null
  currentPage.value = 1
  pageSize.value = 20
}

// 格式化日期时间
const formatDateTimeStr = (date: Date) => {
  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, '0')
  const d = String(date.getDate()).padStart(2, '0')
  const h = String(date.getHours()).padStart(2, '0')
  const min = String(date.getMinutes()).padStart(2, '0')
  const s = String(date.getSeconds()).padStart(2, '0')
  return `${y}-${m}-${d} ${h}:${min}:${s}`
}

// 日期范围变化处理
const handleDateRangeChange = (val: [Date, Date] | null) => {
  if (!val) {
    filters.value.start_time = ''
    filters.value.end_time = ''
    return
  }
  
  const [start, end] = val
  const startDate = new Date(start)
  const endDate = new Date(end)
  
  // 判断是否选择了同一天且时间都是00:00:00（即只选了日期没选时间）
  const isSameDay = startDate.getFullYear() === endDate.getFullYear() &&
                    startDate.getMonth() === endDate.getMonth() &&
                    startDate.getDate() === endDate.getDate()
  
  const isDefaultTime = startDate.getHours() === 0 && startDate.getMinutes() === 0 && startDate.getSeconds() === 0 &&
                        endDate.getHours() === 0 && endDate.getMinutes() === 0 && endDate.getSeconds() === 0
  
  if (isSameDay && isDefaultTime) {
    // 同一天且未选时间：设置为 00:00:00 到 23:59:59
    const dateStr = `${startDate.getFullYear()}-${String(startDate.getMonth() + 1).padStart(2, '0')}-${String(startDate.getDate()).padStart(2, '0')}`
    filters.value.start_time = `${dateStr} 00:00:00`
    filters.value.end_time = `${dateStr} 23:59:59`
  } else {
    // 使用用户选择的具体时间
    filters.value.start_time = formatDateTimeStr(startDate)
    filters.value.end_time = formatDateTimeStr(endDate)
  }
}

// 从 URL query 初始化筛选条件（先清空再设置，确保只使用URL中的参数）
const initFiltersFromQuery = () => {
  // 先重置所有过滤条件
  resetFilters()
  
  // 再根据URL参数设置
  const query = route.query
  if (query.platform) filters.value.platform = query.platform as string
  if (query.auth_id) filters.value.auth_id = Number(query.auth_id)
  if (query.status) filters.value.status = query.status as string
  if (query.keyword) filters.value.keyword = query.keyword as string
  if (query.start_time) filters.value.start_time = query.start_time as string
  if (query.end_time) filters.value.end_time = query.end_time as string
  if (query.page) currentPage.value = Number(query.page)
  if (query.page_size) pageSize.value = Number(query.page_size)
  
  // 恢复日期范围选择器的值（包含完整时间）
  if (filters.value.start_time && filters.value.end_time) {
    dateRange.value = [
      new Date(filters.value.start_time.replace(' ', 'T')),
      new Date(filters.value.end_time.replace(' ', 'T'))
    ]
  }
}

// 更新 URL query（不触发导航）
const updateQueryParams = () => {
  const query: Record<string, string> = {}
  if (filters.value.platform) query.platform = filters.value.platform
  if (filters.value.auth_id) query.auth_id = String(filters.value.auth_id)
  if (filters.value.status) query.status = filters.value.status
  if (filters.value.keyword) query.keyword = filters.value.keyword
  if (filters.value.start_time) query.start_time = filters.value.start_time
  if (filters.value.end_time) query.end_time = filters.value.end_time
  if (currentPage.value > 1) query.page = String(currentPage.value)
  if (pageSize.value !== 20) query.page_size = String(pageSize.value)
  
  router.replace({ query })
}

// 平台列表
const platforms = ref<PlatformInfo[]>([])
// 授权列表
const auths = ref<PlatformAuth[]>([])


// 状态选项
const statusOptions = [
  { label: '全部状态', value: '' },
  { label: '待处理（含待发货）', value: 'pending,ready_to_ship' },
  { label: '待处理', value: 'pending' },
  { label: '待发货', value: 'ready_to_ship' },
  { label: '已发货', value: 'shipped' },
  { label: '已签收', value: 'delivered' },
  { label: '已取消', value: 'cancelled' }
]

// 获取平台列表
const fetchPlatforms = async () => {
  try {
    const res = await getPlatforms()
    platforms.value = res.data
  } catch (error) {
    console.error('获取平台列表失败', error)
  }
}

// 获取授权列表
const fetchAuths = async () => {
  try {
    const res = await getAuths(1, 100)
    auths.value = res.data.list
  } catch (error) {
    console.error('获取授权列表失败', error)
  }
}

// 获取订单列表
const fetchOrders = async () => {
  loading.value = true
  try {
    const res = await getOrders({
      page: currentPage.value,
      page_size: pageSize.value,
      platform: filters.value.platform,
      auth_id: filters.value.auth_id,
      status: filters.value.status,
      keyword: filters.value.keyword,
      start_time: filters.value.start_time,
      end_time: filters.value.end_time
    })
    tableData.value = res.data.list
    total.value = res.data.total
  } catch (error) {
    console.error('获取订单列表失败', error)
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  currentPage.value = 1
  updateQueryParams()
  fetchOrders()
}

// 重置筛选
const handleReset = () => {
  resetFilters()
  router.replace({ query: {} })
  fetchOrders()
}

// 分页变化
const handlePageChange = (page: number) => {
  currentPage.value = page
  updateQueryParams()
  fetchOrders()
}

// 每页数量变化
const handleSizeChange = (size: number) => {
  pageSize.value = size
  currentPage.value = 1
  updateQueryParams()
  fetchOrders()
}

// 复制订单号
const handleCopyOrderNo = async (orderNo: string, event: Event) => {
  event.stopPropagation()
  try {
    await navigator.clipboard.writeText(orderNo)
    ElMessage.success('订单号已复制')
  } catch (error) {
    ElMessage.error('复制失败')
  }
}

// 查看详情
const handleViewDetail = (row: Order) => {
  router.push({ name: 'OrderDetail', params: { id: row.id } })
}

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

// 监听路由query变化（处理从其他页面跳转带参数的情况）
watch(
  () => route.query,
  (newQuery, oldQuery) => {
    // 只有当query真正变化时才重新加载
    if (JSON.stringify(newQuery) !== JSON.stringify(oldQuery)) {
      initFiltersFromQuery()
      fetchOrders()
    }
  }
)

// keep-alive 激活时重新检查URL参数
onActivated(() => {
  initFiltersFromQuery()
  fetchOrders()
})

// ========== 同步订单相关 ==========
const syncDialogVisible = ref(false)
const syncAuthId = ref<number | undefined>(undefined)
const syncLoading = ref(false)
const syncDays = ref(7)

// 打开同步对话框
const openSyncDialog = () => {
  syncAuthId.value = filters.value.auth_id || (auths.value.length > 0 ? auths.value[0].id : undefined)
  syncDays.value = 7
  syncDialogVisible.value = true
}

// 执行同步
const handleSync = async () => {
  if (!syncAuthId.value) {
    ElMessage.warning('请选择要同步的店铺')
    return
  }
  
  syncLoading.value = true
  try {
    const now = new Date()
    let since: Date
    if (syncDays.value === -1) {
      since = new Date('2020-01-01')
    } else {
      since = new Date(now.getTime() - syncDays.value * 24 * 60 * 60 * 1000)
    }
    
    const res = await syncOrders(syncAuthId.value, {
      since: since.toISOString(),
      to: now.toISOString()
    })
    ElMessage.success(`同步完成：共 ${res.data.total} 条，新增 ${res.data.created} 条，更新 ${res.data.updated} 条`)
    syncDialogVisible.value = false
    fetchOrders() // 刷新列表
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.message || '同步订单失败')
  } finally {
    syncLoading.value = false
  }
}

onMounted(() => {
  // 从 URL query 恢复筛选条件
  initFiltersFromQuery()
  fetchPlatforms()
  fetchAuths()
  fetchOrders()
})
</script>

<template>
  <div class="orders-page">
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">订单列表</h1>
        <p class="page-desc">查看和管理从各电商平台同步的订单</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="openSyncDialog">同步订单</el-button>
      </div>
    </div>

    <div class="content-card">
      <!-- 筛选区域 -->
      <div class="filter-area">
        <el-form :inline="true" :model="filters" class="filter-form">
          <el-form-item label="平台">
            <el-select v-model="filters.platform" placeholder="全部平台" clearable style="width: 120px">
              <el-option
                v-for="platform in platforms"
                :key="platform.name"
                :label="platform.label"
                :value="platform.name"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="店铺">
            <el-select v-model="filters.auth_id" placeholder="全部店铺" clearable style="width: 150px">
              <el-option
                v-for="auth in auths"
                :key="auth.id"
                :label="auth.shop_name"
                :value="auth.id"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="filters.status" placeholder="全部状态" clearable style="width: 120px">
              <el-option
                v-for="option in statusOptions"
                :key="option.value"
                :label="option.label"
                :value="option.value"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="关键词">
            <el-input v-model="filters.keyword" placeholder="订单号/收件人" clearable style="width: 160px" />
          </el-form-item>
          <el-form-item label="下单时间">
            <el-date-picker
              v-model="dateRange"
              type="datetimerange"
              range-separator="至"
              start-placeholder="开始时间"
              end-placeholder="结束时间"
              format="YYYY-MM-DD HH:mm:ss"
              date-format="YYYY-MM-DD"
              time-format="HH:mm:ss"
              :unlink-panels="false"
              style="width: 380px"
              @change="handleDateRangeChange"
            />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :icon="Search" @click="handleSearch">搜索</el-button>
            <el-button @click="handleReset">重置</el-button>
          </el-form-item>
        </el-form>
      </div>

      <!-- 表格 -->
      <el-table
        v-loading="loading"
        :data="tableData"
        style="width: 100%"
        stripe
        border
      >
        <el-table-column prop="platform_order_no" label="订单号" min-width="180">
          <template #default="{ row }">
            <span class="order-no-wrapper">
              <span class="order-no-link" @click="handleViewDetail(row)">{{ row.platform_order_no }}</span>
              <el-icon class="copy-icon" @click="handleCopyOrderNo(row.platform_order_no, $event)" title="复制订单号">
                <CopyDocument />
              </el-icon>
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="platform" label="平台" min-width="80">
          <template #default="{ row }">
            <el-tag type="primary" size="small">
              {{ getPlatformLabel(row.platform) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" min-width="90">
          <template #default="{ row }">
            <el-tag :type="getStatusTagType(row.status)" size="small">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="total_amount" label="金额" min-width="110">
          <template #default="{ row }">
            {{ row.currency }} {{ row.total_amount?.toFixed(2) }}
          </template>
        </el-table-column>
        <el-table-column prop="sale_commission" label="佣金" min-width="110">
          <template #default="{ row }">
            <span v-if="row.sale_commission" style="color: var(--el-color-warning)">
              {{ row.commission_currency || row.currency }} {{ row.sale_commission?.toFixed(2) }}
            </span>
            <span v-else style="color: var(--text-secondary)">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="recipient_name" label="收件人" min-width="100" show-overflow-tooltip />
        <el-table-column prop="country" label="国家/地区" min-width="100" show-overflow-tooltip />
        <el-table-column label="商品数" min-width="70" align="center">
          <template #default="{ row }">
            {{ row.items?.length || 0 }}
          </template>
        </el-table-column>
        <el-table-column prop="order_time" label="下单时间" min-width="160">
          <template #default="{ row }">
            {{ row.order_time ? formatDateTime(row.order_time) : '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="updated_at" label="更新时间" min-width="160">
          <template #default="{ row }">
            {{ row.updated_at ? formatDateTime(row.updated_at) : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="80" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" :icon="View" @click="handleViewDetail(row)">
              详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="handlePageChange"
          @size-change="handleSizeChange"
        />
      </div>
    </div>

    <!-- 同步对话框 -->
    <el-dialog
      v-model="syncDialogVisible"
      title="同步订单"
      width="400px"
    >
      <el-form label-width="80px">
        <el-form-item label="选择店铺">
          <el-select v-model="syncAuthId" placeholder="选择店铺" style="width: 100%">
            <el-option
              v-for="auth in auths"
              :key="auth.id"
              :label="auth.shop_name"
              :value="auth.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="同步范围">
          <el-select v-model="syncDays" style="width: 100%">
            <el-option :value="1" label="最近1天" />
            <el-option :value="3" label="最近3天" />
            <el-option :value="7" label="最近7天" />
            <el-option :value="14" label="最近14天" />
            <el-option :value="30" label="最近30天" />
            <el-option :value="90" label="最近90天" />
            <el-option :value="-1" label="全部订单" />
          </el-select>
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
.orders-page {
  padding: 0;
}

.page-header {
  margin-bottom: 24px;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.header-left {
  flex: 1;
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
  margin-bottom: 20px;
  padding-bottom: 20px;
  border-bottom: 1px solid var(--border-color);
}

.filter-form {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 0;
}

.pagination-wrapper {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

:deep(.el-descriptions) {
  margin-bottom: 0;
}

.order-no-wrapper {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.order-no-link {
  color: var(--el-color-primary);
  cursor: pointer;
  transition: color 0.2s;

  &:hover {
    color: var(--el-color-primary-light-3);
    text-decoration: underline;
  }
}

.copy-icon {
  cursor: pointer;
  color: var(--text-secondary);
  font-size: 14px;
  transition: color 0.2s;

  &:hover {
    color: var(--el-color-primary);
  }
}
</style>
