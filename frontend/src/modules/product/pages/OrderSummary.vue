<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Picture } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import api, { type OrderSummaryItem, type StockInOrderItemRequest } from '../api'
import { getAuths, type AuthResponse } from '@/modules/order/api'
import ImagePreview from '@/components/ImagePreview.vue'

defineOptions({ name: 'OrderSummary' })

const imagePreviewRef = ref<InstanceType<typeof ImagePreview>>()
const showImagePreview = (src: string) => {
  imagePreviewRef.value?.show(src)
}

const loading = ref(false)
const tableData = ref<OrderSummaryItem[]>([])
const authId = ref<number | undefined>(undefined)
const authOptions = ref<AuthResponse[]>([])
const dateRange = ref<[Date, Date] | null>(null)
const keyword = ref('')
const status = ref('')

// 订单状态选项
const statusOptions = [
  { value: 'pending', label: '待处理' },
  { value: 'ready_to_ship', label: '待发货' },
  { value: 'shipped', label: '已发货' },
  { value: 'delivered', label: '已签收' },
  { value: 'cancelled', label: '已取消' },
]

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
    if (keyword.value) {
      params.keyword = keyword.value
    }
    if (status.value) {
      params.status = status.value
    }

    const res = await api.getOrderSummary(params)
    tableData.value = res.data || []
  } catch (error) {
    console.error('获取订单汇总失败', error)
  } finally {
    loading.value = false
  }
}

// 重置搜索
const handleReset = () => {
  keyword.value = ''
  status.value = ''
  fetchSummary()
}

const formatDate = (date: Date) => {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

// 获取状态标签类型（根据中文状态名称）
const getStatusType = (status: string) => {
  const typeMap: Record<string, string> = {
    '已发货': 'success',
    '已签收': 'success',
    '配送中': 'primary',
    '已取消': 'danger',
    '已退货': 'danger',
    '待处理': 'warning',
    '待打包': 'warning',
    '待发货': 'warning',
    '待揽收': 'primary',
    '仲裁中': 'danger',
  }
  return typeMap[status] || 'info'
}

// 获取指定状态的销量
const getStatusQuantity = (row: OrderSummaryItem, statusName: string) => {
  const detail = row.status_details?.find(d => d.status === statusName)
  return detail?.quantity || 0
}

// 获取需求量（待处理 + 待发货）
const getNeedQuantity = (row: OrderSummaryItem) => {
  return getStatusQuantity(row, '待处理') + getStatusQuantity(row, '待发货')
}

// 获取库存状态样式类
const getStockStatusClass = (row: OrderSummaryItem) => {
  const need = getNeedQuantity(row)
  if (need === 0) return ''
  return row.available_stock >= need ? 'stock-sufficient' : 'stock-insufficient'
}

// ========== 备货相关 ==========
const stockInDialogVisible = ref(false)
const stockInLoading = ref(false)
const stockInForm = ref<{
  product_id: number
  sku: string
  product_name: string
  quantity: number
  suggested_quantity: number
}>({
  product_id: 0,
  sku: '',
  product_name: '',
  quantity: 0,
  suggested_quantity: 0
})

// 打开备货弹窗
const handleStockIn = (row: OrderSummaryItem) => {
  if (!row.local_sku) {
    ElMessage.warning('该产品未关联本地SKU，无法备货')
    return
  }
  
  // 计算建议备货数量 = 待处理 + 待发货
  const pending = getStatusQuantity(row, '待处理')
  const readyToShip = getStatusQuantity(row, '待发货')
  const suggested = pending + readyToShip
  
  // 需要从后端获取 product_id，这里通过本地产品列表查找
  // 简化处理：使用 API 查询
  stockInForm.value = {
    product_id: 0, // 需要通过 SKU 查询
    sku: row.local_sku,
    product_name: row.product_name || row.local_sku,
    quantity: suggested,
    suggested_quantity: suggested
  }
  
  // 查询本地产品获取 product_id
  fetchProductIdBySku(row.local_sku)
  stockInDialogVisible.value = true
}

// 通过 SKU 查询产品ID
const fetchProductIdBySku = async (sku: string) => {
  try {
    const res = await api.listProducts({ page: 1, page_size: 1000 })
    const product = res.data.list?.find((p: any) => p.sku === sku)
    if (product) {
      stockInForm.value.product_id = product.id
    } else {
      ElMessage.error('未找到对应的本地产品')
    }
  } catch (error) {
    console.error('查询产品失败', error)
  }
}

// 提交备货
const submitStockIn = async () => {
  if (!stockInForm.value.product_id) {
    ElMessage.error('产品信息不完整')
    return
  }
  if (stockInForm.value.quantity <= 0) {
    ElMessage.error('入库数量必须大于0')
    return
  }
  
  stockInLoading.value = true
  try {
    const items: StockInOrderItemRequest[] = [{
      product_id: stockInForm.value.product_id,
      quantity: stockInForm.value.quantity
    }]
    
    await api.createStockInOrder({
      items,
      remark: `订单汇总备货 - ${stockInForm.value.sku}`
    })
    
    ElMessage.success('入库单创建成功，库存已更新')
    stockInDialogVisible.value = false
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '创建入库单失败')
  } finally {
    stockInLoading.value = false
  }
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
        <el-form-item label="订单状态">
          <el-select v-model="status" placeholder="全部状态" clearable style="width: 140px" @change="fetchSummary">
            <el-option
              v-for="item in statusOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词">
          <el-input
            v-model="keyword"
            placeholder="SKU/标题"
            clearable
            style="width: 180px"
            @keyup.enter="fetchSummary"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchSummary">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
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
        <el-table-column prop="local_sku" label="系统SKU" width="200">
          <template #default="{ row }">
            <el-tag v-if="row.local_sku" type="success">{{ row.local_sku }}</el-tag>
            <span v-else class="text-secondary">未关联</span>
          </template>
        </el-table-column>
        <el-table-column prop="product_name" label="系统产品名" width="400">
          <template #default="{ row }">
            <span class="wrap-text">{{ row.product_name || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="platform_skus" label="平台SKU" width="180">
          <template #default="{ row }">
            <span v-if="row.platform_skus?.length === 1">{{ row.platform_skus[0] }}</span>
            <el-tooltip v-else-if="row.platform_skus?.length > 1" :content="row.platform_skus.join(', ')" placement="top">
              <span>{{ row.platform_skus[0] }} <el-tag size="small" type="info">+{{ row.platform_skus.length - 1 }}</el-tag></span>
            </el-tooltip>
            <span v-else class="text-secondary">-</span>
          </template>
        </el-table-column>
        <el-table-column label="平台产品" min-width="280">
          <template #default="{ row }">
            <div v-if="row.platform_products?.length > 0" class="platform-product-cell">
              <el-image
                v-if="row.platform_products[0].image"
                :src="row.platform_products[0].image"
                fit="cover"
                class="product-image"
                @click="showImagePreview(row.platform_products[0].image)"
              />
              <div v-else class="product-image-placeholder">
                <el-icon><Picture /></el-icon>
              </div>
              <div class="product-info">
                <span class="product-name" :title="row.platform_products[0].name">{{ row.platform_products[0].name || '-' }}</span>
                <span v-if="row.platform_products.length > 1" class="more-products">
                  +{{ row.platform_products.length - 1 }} 个产品
                </span>
              </div>
            </div>
            <span v-else class="text-secondary">-</span>
          </template>
        </el-table-column>
        <el-table-column label="待处理" width="90" align="center">
          <template #default="{ row }">
            <span :class="{ 'status-value': getStatusQuantity(row, '待处理') > 0 }">
              {{ getStatusQuantity(row, '待处理') || '-' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="待发货" width="80" align="center">
          <template #default="{ row }">
            <span :class="{ 'status-value': getStatusQuantity(row, '待发货') > 0 }">
              {{ getStatusQuantity(row, '待发货') || '-' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="可用库存" width="100" align="center">
          <template #default="{ row }">
            <span class="stock-display" :class="getStockStatusClass(row)">
              {{ row.available_stock || 0 }}
              <el-tag 
                v-if="getNeedQuantity(row) > 0"
                :type="row.available_stock >= getNeedQuantity(row) ? 'success' : 'danger'" 
                size="small"
                class="stock-tag"
              >
                {{ row.available_stock >= getNeedQuantity(row) ? '充足' : '不足' }}
              </el-tag>
            </span>
          </template>
        </el-table-column>
        <el-table-column label="已发货" width="90" align="center">
          <template #default="{ row }">
            <span :class="{ 'status-value-success': getStatusQuantity(row, '已发货') > 0 }">
              {{ getStatusQuantity(row, '已发货') || '-' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="quantity" label="总销量" width="100" sortable />
        <el-table-column label="操作" width="100" fixed="right" align="center">
          <template #default="{ row }">
            <el-button
              v-if="row.local_sku"
              type="primary"
              link
              size="small"
              @click="handleStockIn(row)"
            >
              去备货
            </el-button>
            <span v-else class="text-secondary">-</span>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 备货弹窗 -->
    <el-dialog
      v-model="stockInDialogVisible"
      title="创建入库单"
      width="450px"
      :close-on-click-modal="false"
    >
      <el-form label-width="100px">
        <el-form-item label="系统SKU">
          <el-tag type="success">{{ stockInForm.sku }}</el-tag>
        </el-form-item>
        <el-form-item label="产品名称">
          <span>{{ stockInForm.product_name }}</span>
        </el-form-item>
        <el-form-item label="建议数量">
          <span class="suggested-quantity">{{ stockInForm.suggested_quantity }}</span>
          <span class="quantity-hint">（待处理 + 待发货）</span>
        </el-form-item>
        <el-form-item label="入库数量" required>
          <el-input-number
            v-model="stockInForm.quantity"
            :min="1"
            :max="99999"
            controls-position="right"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="stockInDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="stockInLoading" @click="submitStockIn">
          确认入库
        </el-button>
      </template>
    </el-dialog>

    <!-- 图片预览 -->
    <ImagePreview ref="imagePreviewRef" />
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

.status-value {
  color: #e6a23c;
  font-weight: 500;
}

.status-value-success {
  color: #67c23a;
  font-weight: 500;
}

.platform-product-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.product-image {
  width: 50px;
  height: 50px;
  min-width: 50px;
  border-radius: 4px;
  cursor: pointer;
}

.product-image-placeholder {
  width: 50px;
  height: 50px;
  min-width: 50px;
  border-radius: 4px;
  background: var(--bg-page);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-secondary);
  font-size: 20px;
}

.product-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.product-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 13px;
}

.more-products {
  font-size: 12px;
  color: var(--text-secondary);
}

.wrap-text {
  word-break: break-word;
  white-space: normal;
  line-height: 1.5;
}

.suggested-quantity {
  color: #409eff;
  font-weight: 600;
  font-size: 16px;
}

.quantity-hint {
  color: var(--text-secondary);
  font-size: 12px;
  margin-left: 8px;
}

.stock-display {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-weight: 500;
}

.stock-tag {
  transform: scale(0.9);
}

.stock-sufficient {
  color: var(--el-color-success);
}

.stock-insufficient {
  color: var(--el-color-danger);
}
</style>
