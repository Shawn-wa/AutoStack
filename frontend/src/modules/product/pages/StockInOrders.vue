<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { Plus, Delete, Search } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api, { type StockInOrderResponse, type Product, type StockInOrderItemRequest, type WarehouseResponse } from '../api'

defineOptions({ name: 'StockInOrders' })

const loading = ref(false)
const tableData = ref<StockInOrderResponse[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)
const statusFilter = ref('')

// 仓库列表
const warehouseList = ref<WarehouseResponse[]>([])

// 新建入库单相关
const dialogVisible = ref(false)
const dialogLoading = ref(false)
const selectedWarehouseId = ref<number | undefined>(undefined)
const searchKeyword = ref('')
const searchLoading = ref(false)
const searchResults = ref<Product[]>([])
const selectedProducts = ref<Array<{ product: Product; quantity: number }>>([])
const remark = ref('')

// 计算入库单总数量
const dialogTotalQuantity = computed(() => {
  return selectedProducts.value.reduce((sum, item) => sum + item.quantity, 0)
})

// 获取可用仓库列表
const fetchWarehouses = async () => {
  try {
    const res = await api.listAvailableWarehouses()
    warehouseList.value = res.data.list || []
    // 设置默认仓库
    if (warehouseList.value.length > 0 && !selectedWarehouseId.value) {
      selectedWarehouseId.value = warehouseList.value[0].id
    }
  } catch (error) {
    console.error('获取仓库列表失败', error)
  }
}

// 状态选项
const statusOptions = [
  { value: 'pending', label: '待入库' },
  { value: 'completed', label: '已完成' },
  { value: 'cancelled', label: '已取消' },
]

// 获取状态标签类型
const getStatusType = (status: string) => {
  const typeMap: Record<string, string> = {
    'pending': 'warning',
    'completed': 'success',
    'cancelled': 'danger',
  }
  return typeMap[status] || 'info'
}

// 获取状态显示文本
const getStatusLabel = (status: string) => {
  const labelMap: Record<string, string> = {
    'pending': '待入库',
    'completed': '已完成',
    'cancelled': '已取消',
  }
  return labelMap[status] || status
}

// 获取入库单列表
const fetchList = async () => {
  loading.value = true
  try {
    const params: any = {
      page: currentPage.value,
      page_size: pageSize.value,
    }
    if (statusFilter.value) {
      params.status = statusFilter.value
    }
    
    const res = await api.listStockInOrders(params)
    tableData.value = res.data.list || []
    total.value = res.data.total || 0
  } catch (error) {
    console.error('获取入库单列表失败', error)
  } finally {
    loading.value = false
  }
}

// 重置筛选
const handleReset = () => {
  statusFilter.value = ''
  currentPage.value = 1
  fetchList()
}

// 分页改变
const handlePageChange = (page: number) => {
  currentPage.value = page
  fetchList()
}

// 计算入库单总数量
const getTotalQuantity = (items: StockInOrderResponse['items']) => {
  return items?.reduce((sum, item) => sum + item.quantity, 0) || 0
}

// ========== 新建入库单相关 ==========

// 打开新建弹窗
const openDialog = () => {
  searchKeyword.value = ''
  searchResults.value = []
  selectedProducts.value = []
  remark.value = ''
  selectedWarehouseId.value = undefined // 不设置默认仓库，让用户主动选择
  dialogVisible.value = true
}

// 搜索本地产品
const handleSearch = async (query: string) => {
  if (!query || query.length < 1) {
    searchResults.value = []
    return
  }
  
  searchLoading.value = true
  try {
    // 使用产品列表API搜索（后端需要支持keyword参数）
    const res = await api.listProducts({ page: 1, page_size: 20, keyword: query } as any)
    searchResults.value = res.data.list || []
  } catch (error) {
    console.error('搜索产品失败', error)
    searchResults.value = []
  } finally {
    searchLoading.value = false
  }
}

// 选择产品
const handleSelectProduct = (product: Product) => {
  // 检查是否已添加
  const exists = selectedProducts.value.find(item => item.product.id === product.id)
  if (exists) {
    ElMessage.warning('该产品已添加')
    return
  }
  
  selectedProducts.value.push({
    product,
    quantity: 1
  })
  
  // 清空搜索
  searchKeyword.value = ''
  searchResults.value = []
}

// 移除产品
const handleRemoveProduct = (index: number) => {
  selectedProducts.value.splice(index, 1)
}

// 提交入库单
const handleSubmit = async () => {
  if (!selectedWarehouseId.value) {
    ElMessage.warning('请选择入库仓库')
    return
  }
  
  if (selectedProducts.value.length === 0) {
    ElMessage.warning('请至少添加一个产品')
    return
  }
  
  // 检查数量
  for (const item of selectedProducts.value) {
    if (item.quantity <= 0) {
      ElMessage.warning(`产品 ${item.product.sku} 的入库数量必须大于0`)
      return
    }
  }
  
  // 获取仓库名称
  const warehouse = warehouseList.value.find(w => w.id === selectedWarehouseId.value)
  const warehouseName = warehouse?.name || '未知仓库'
  
  // 二次确认
  try {
    await ElMessageBox.confirm(
      `确认将 ${selectedProducts.value.length} 个产品（共 ${dialogTotalQuantity.value} 件）入库到「${warehouseName}」？`,
      '确认入库',
      {
        confirmButtonText: '确认',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
  } catch {
    return // 用户取消
  }
  
  dialogLoading.value = true
  try {
    const items: StockInOrderItemRequest[] = selectedProducts.value.map(item => ({
      product_id: item.product.id,
      quantity: item.quantity
    }))
    
    await api.createStockInOrder({
      warehouse_id: selectedWarehouseId.value,
      items,
      remark: remark.value || undefined
    })
    
    ElMessage.success('入库单创建成功')
    dialogVisible.value = false
    fetchList()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '创建失败')
  } finally {
    dialogLoading.value = false
  }
}

onMounted(() => {
  fetchList()
  fetchWarehouses()
})
</script>

<template>
  <div class="stock-in-orders">
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">入库单</h1>
        <p class="page-desc">管理系统入库记录</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="openDialog">
          <el-icon><Plus /></el-icon>
          新建入库单
        </el-button>
      </div>
    </div>

    <div class="filter-card">
      <el-form inline>
        <el-form-item label="状态">
          <el-select v-model="statusFilter" placeholder="全部状态" clearable style="width: 140px" @change="fetchList">
            <el-option
              v-for="item in statusOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchList">查询</el-button>
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
      >
        <el-table-column prop="order_no" label="入库单号" width="200">
          <template #default="{ row }">
            <span class="order-no">{{ row.order_no }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="warehouse_name" label="入库仓库" width="120">
          <template #default="{ row }">
            <span class="warehouse-name">{{ row.warehouse_name || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)" size="small">
              {{ getStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="入库明细" min-width="300">
          <template #default="{ row }">
            <div class="items-cell">
              <div v-for="item in row.items?.slice(0, 3)" :key="item.id" class="item-row">
                <el-tag type="success" size="small">{{ item.sku }}</el-tag>
                <span class="item-name">{{ item.product_name }}</span>
                <span class="item-qty">x {{ item.quantity }}</span>
              </div>
              <div v-if="row.items?.length > 3" class="more-items">
                +{{ row.items.length - 3 }} 个SKU
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="总数量" width="100" align="center">
          <template #default="{ row }">
            <span class="total-qty">{{ getTotalQuantity(row.items) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" width="200">
          <template #default="{ row }">
            <span class="remark-text">{{ row.remark || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180" />
      </el-table>

      <div class="pagination-wrapper" v-if="total > pageSize">
        <el-pagination
          background
          layout="prev, pager, next"
          :total="total"
          :page-size="pageSize"
          :current-page="currentPage"
          @current-change="handlePageChange"
        />
      </div>
    </div>

    <!-- 新建入库单弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      title="新建入库单"
      width="720px"
      :close-on-click-modal="false"
      class="stock-in-dialog"
    >
      <div class="dialog-content">
        <!-- 基本信息 -->
        <div class="form-section">
          <div class="form-row">
            <div class="form-item" style="max-width: 280px">
              <label class="form-label required">入库仓库</label>
              <el-select
                v-model="selectedWarehouseId"
                placeholder="请选择仓库"
                style="width: 100%"
              >
                <el-option
                  v-for="wh in warehouseList"
                  :key="wh.id"
                  :label="wh.name"
                  :value="wh.id"
                />
              </el-select>
            </div>
          </div>
        </div>

        <!-- 搜索产品 -->
        <div class="form-section">
          <label class="form-label">添加产品</label>
          <div class="search-wrapper" style="max-width: 400px">
            <el-input
              v-model="searchKeyword"
              placeholder="输入SKU或产品名称搜索"
              clearable
              @input="handleSearch"
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
            
            <!-- 搜索结果 -->
            <div v-if="searchResults.length > 0" class="search-results">
              <div
                v-for="product in searchResults"
                :key="product.id"
                class="search-result-item"
                @click="handleSelectProduct(product)"
              >
                <el-tag type="success" size="small">{{ product.sku }}</el-tag>
                <span class="result-name">{{ product.name }}</span>
                <el-button type="primary" link size="small">添加</el-button>
              </div>
            </div>
            <div v-else-if="searchKeyword && !searchLoading" class="no-results">
              未找到匹配的产品
            </div>
          </div>
        </div>

        <!-- 已选产品列表 -->
        <div class="selected-section">
          <div class="section-header">
            <span class="section-title">入库明细</span>
            <span class="section-count">共 {{ selectedProducts.length }} 项，合计 {{ dialogTotalQuantity }} 件</span>
          </div>
          
          <div v-if="selectedProducts.length === 0" class="empty-list">
            请搜索并添加要入库的产品
          </div>
          
          <div v-else class="selected-list">
            <div v-for="(item, index) in selectedProducts" :key="item.product.id" class="selected-item">
              <div class="item-product">
                <div class="product-image">
                  <el-image
                    v-if="item.product.image"
                    :src="item.product.image"
                    fit="cover"
                  />
                  <div v-else class="image-placeholder">暂无图片</div>
                </div>
                <div class="product-detail">
                  <div class="product-sku">{{ item.product.sku }}</div>
                  <div class="product-name">{{ item.product.name }}</div>
                </div>
              </div>
              <div class="item-actions">
                <span class="qty-label">数量</span>
                <el-input-number
                  v-model="item.quantity"
                  :min="1"
                  :max="99999"
                  controls-position="right"
                  style="width: 140px"
                />
                <el-button
                  type="danger"
                  link
                  @click="handleRemoveProduct(index)"
                >
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
            </div>
          </div>
        </div>

        <!-- 备注 -->
        <div class="form-section">
          <label class="form-label">备注</label>
          <el-input
            v-model="remark"
            type="textarea"
            :rows="2"
            placeholder="选填"
          />
        </div>
      </div>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="dialogLoading" @click="handleSubmit">
            确认入库
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped lang="scss">
.stock-in-orders {
  padding: 0;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
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

.order-no {
  font-family: monospace;
  font-weight: 500;
  color: var(--color-primary);
}

.items-cell {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.item-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.item-name {
  flex: 1;
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item-qty {
  font-size: 13px;
  color: var(--text-secondary);
  font-weight: 500;
}

.more-items {
  font-size: 12px;
  color: var(--text-secondary);
}

.total-qty {
  font-size: 16px;
  font-weight: 600;
  color: var(--color-primary);
}

.remark-text {
  font-size: 13px;
  color: var(--text-secondary);
}

.warehouse-name {
  font-size: 13px;
}

.pagination-wrapper {
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
}

// 弹窗样式
.dialog-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-section {
  // 表单区块
}

.form-row {
  display: flex;
  gap: 16px;
}

.form-item {
  flex: 1;
}

.form-label {
  display: block;
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 8px;
  color: var(--text-primary);

  &.required::before {
    content: '*';
    color: var(--el-color-danger);
    margin-right: 4px;
  }
}

.search-wrapper {
  position: relative;
}

.search-results {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  margin-top: 4px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  z-index: 100;
  max-height: 240px;
  overflow-y: auto;
}

.search-result-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 16px;
  cursor: pointer;
  transition: background 0.2s;

  &:hover {
    background: var(--bg-hover);
  }

  .result-name {
    flex: 1;
    font-size: 13px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.no-results {
  padding: 12px;
  text-align: center;
  color: var(--text-secondary);
  font-size: 13px;
}

.selected-section {
  background: var(--bg-page);
  border-radius: var(--radius-md);
  padding: 16px;
  border: 1px solid var(--border-color);
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--border-color);
}

.section-title {
  font-weight: 600;
  font-size: 14px;
}

.section-count {
  font-size: 13px;
  color: var(--color-primary);
  font-weight: 500;
}

.empty-list {
  padding: 32px;
  text-align: center;
  color: var(--text-muted);
  font-size: 13px;
}

.selected-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  max-height: 280px;
  overflow-y: auto;
}

.selected-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 16px;
  background: var(--bg-card);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-color);
}

.item-product {
  display: flex;
  align-items: center;
  gap: 14px;
  flex: 1;
  min-width: 0;
}

.product-image {
  width: 56px;
  height: 56px;
  flex-shrink: 0;
  border-radius: var(--radius-sm);
  overflow: hidden;
  border: 1px solid var(--border-color);
  background: var(--bg-page);

  .el-image {
    width: 100%;
    height: 100%;
  }
}

.image-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  color: var(--text-muted);
  background: var(--bg-page);
}

.product-detail {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.product-sku {
  font-size: 14px;
  font-weight: 500;
  color: var(--color-primary);
}

.product-name {
  font-size: 13px;
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
}

.qty-label {
  font-size: 14px;
  color: var(--text-secondary);
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>
