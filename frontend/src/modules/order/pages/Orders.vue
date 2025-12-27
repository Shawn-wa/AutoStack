<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Search, View } from '@element-plus/icons-vue'
import {
  getOrders,
  getPlatforms,
  getAuths,
  type Order,
  type PlatformInfo,
  type PlatformAuth
} from '@/modules/order/api'
import { formatDateTime } from '@/utils/format'

const router = useRouter()

const loading = ref(false)
const tableData = ref<Order[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)

// 筛选条件
const filters = ref({
  platform: '',
  auth_id: undefined as number | undefined,
  status: '',
  keyword: '',
  start_time: '',
  end_time: ''
})

// 平台列表
const platforms = ref<PlatformInfo[]>([])
// 授权列表
const auths = ref<PlatformAuth[]>([])


// 状态选项
const statusOptions = [
  { label: '全部状态', value: '' },
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
  fetchOrders()
}

// 重置筛选
const handleReset = () => {
  filters.value = {
    platform: '',
    auth_id: undefined,
    status: '',
    keyword: '',
    start_time: '',
    end_time: ''
  }
  currentPage.value = 1
  fetchOrders()
}

// 分页变化
const handlePageChange = (page: number) => {
  currentPage.value = page
  fetchOrders()
}

// 每页数量变化
const handleSizeChange = (size: number) => {
  pageSize.value = size
  currentPage.value = 1
  fetchOrders()
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

onMounted(() => {
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
              v-model="filters.start_time"
              type="date"
              placeholder="开始日期"
              format="YYYY-MM-DD"
              value-format="YYYY-MM-DD"
              style="width: 140px"
            />
            <span style="margin: 0 8px; color: var(--text-secondary)">至</span>
            <el-date-picker
              v-model="filters.end_time"
              type="date"
              placeholder="结束日期"
              format="YYYY-MM-DD"
              value-format="YYYY-MM-DD"
              style="width: 140px"
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
      >
        <el-table-column prop="platform_order_no" label="订单号" width="180" />
        <el-table-column prop="platform" label="平台" width="100">
          <template #default="{ row }">
            <el-tag type="primary" size="small">
              {{ getPlatformLabel(row.platform) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusTagType(row.status)" size="small">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="total_amount" label="金额" width="120">
          <template #default="{ row }">
            {{ row.currency }} {{ row.total_amount?.toFixed(2) }}
          </template>
        </el-table-column>
        <el-table-column prop="sale_commission" label="佣金" width="120">
          <template #default="{ row }">
            <span v-if="row.sale_commission" style="color: var(--el-color-warning)">
              {{ row.commission_currency || row.currency }} {{ row.sale_commission?.toFixed(2) }}
            </span>
            <span v-else style="color: var(--text-secondary)">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="recipient_name" label="收件人" width="120" />
        <el-table-column prop="country" label="国家/地区" width="120" />
        <el-table-column label="商品数" width="80">
          <template #default="{ row }">
            {{ row.items?.length || 0 }}
          </template>
        </el-table-column>
        <el-table-column prop="order_time" label="下单时间" width="180">
          <template #default="{ row }">
            {{ row.order_time ? formatDateTime(row.order_time) : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
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
</style>
