<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Picture, Plus } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import api, { type InventoryResponse, type WarehouseResponse } from '../api'
import ImagePreview from '@/components/ImagePreview.vue'

defineOptions({ name: 'InventoryList' })

const imagePreviewRef = ref<InstanceType<typeof ImagePreview>>()
const showImagePreview = (src: string, event: MouseEvent) => {
  imagePreviewRef.value?.show(src, event)
}
const hideImagePreview = () => {
  imagePreviewRef.value?.hide()
}

const loading = ref(false)
const tableData = ref<InventoryResponse[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)
const warehouseId = ref<number | undefined>(undefined)
const keyword = ref('')
const warehouses = ref<WarehouseResponse[]>([])

// 仓库类型选项
const warehouseTypeOptions = [
  { value: 'local', label: '本地仓' },
  { value: 'overseas', label: '海外仓' },
  { value: 'fba', label: 'FBA仓' },
  { value: 'third', label: '第三方仓' },
  { value: 'virtual', label: '虚拟仓' },
]

// 新建仓库弹窗
const warehouseDialogVisible = ref(false)
const warehouseLoading = ref(false)
const warehouseForm = ref({
  code: '',
  name: '',
  type: 'local',
  address: ''
})

// 获取仓库列表
const fetchWarehouses = async () => {
  try {
    const res = await api.listWarehouses()
    warehouses.value = res.data || []
    // 默认选中第一个仓库
    if (warehouses.value.length > 0 && !warehouseId.value) {
      warehouseId.value = warehouses.value[0].id
    }
  } catch (error) {
    console.error('获取仓库列表失败', error)
  }
}

// 获取库存列表
const fetchList = async () => {
  loading.value = true
  try {
    const params: any = {
      page: currentPage.value,
      page_size: pageSize.value,
    }
    if (warehouseId.value) {
      params.warehouse_id = warehouseId.value
    }
    if (keyword.value) {
      params.keyword = keyword.value
    }
    
    const res = await api.listInventory(params)
    tableData.value = res.data.list || []
    total.value = res.data.total || 0
  } catch (error) {
    console.error('获取库存列表失败', error)
  } finally {
    loading.value = false
  }
}

// 初始化库存
const handleInitInventory = async () => {
  if (!warehouseId.value) {
    ElMessage.warning('请先选择仓库')
    return
  }
  
  try {
    const res = await api.initInventory(warehouseId.value)
    ElMessage.success(`初始化完成，新建 ${res.data.created} 条库存记录`)
    fetchList()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '初始化失败')
  }
}

// 重置筛选
const handleReset = () => {
  keyword.value = ''
  currentPage.value = 1
  fetchList()
}

// 分页改变
const handlePageChange = (page: number) => {
  currentPage.value = page
  fetchList()
}

// 仓库变化
const handleWarehouseChange = () => {
  currentPage.value = 1
  fetchList()
}

// 获取库存状态样式
const getStockClass = (stock: number) => {
  if (stock <= 0) return 'stock-zero'
  if (stock < 10) return 'stock-low'
  return 'stock-normal'
}

// 获取仓库类型显示文本
const getWarehouseTypeLabel = (type: string) => {
  const item = warehouseTypeOptions.find(o => o.value === type)
  return item?.label || type
}

// 打开新建仓库弹窗
const openWarehouseDialog = () => {
  warehouseForm.value = {
    code: '',
    name: '',
    type: 'local',
    address: ''
  }
  warehouseDialogVisible.value = true
}

// 提交新建仓库
const submitWarehouse = async () => {
  if (!warehouseForm.value.code || !warehouseForm.value.name) {
    ElMessage.warning('请填写仓库编码和名称')
    return
  }
  
  warehouseLoading.value = true
  try {
    await api.createWarehouse(warehouseForm.value)
    ElMessage.success('仓库创建成功')
    warehouseDialogVisible.value = false
    await fetchWarehouses()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '创建失败')
  } finally {
    warehouseLoading.value = false
  }
}

onMounted(async () => {
  await fetchWarehouses()
  fetchList()
})
</script>

<template>
  <div class="inventory-list">
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">库存明细</h1>
        <p class="page-desc">查看系统库存信息</p>
      </div>
      <div class="header-actions">
        <el-button @click="openWarehouseDialog">
          <el-icon><Plus /></el-icon>
          新建仓库
        </el-button>
        <el-button type="primary" @click="handleInitInventory">初始化库存</el-button>
      </div>
    </div>

    <div class="filter-card">
      <el-form inline>
        <el-form-item label="仓库">
          <el-select v-model="warehouseId" placeholder="选择仓库" style="width: 220px" @change="handleWarehouseChange">
            <el-option
              v-for="item in warehouses"
              :key="item.id"
              :label="`${item.name}（${getWarehouseTypeLabel(item.type)}）`"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词">
          <el-input
            v-model="keyword"
            placeholder="SKU"
            clearable
            style="width: 180px"
            @keyup.enter="fetchList"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchList">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="content-card">
      <div class="table-wrapper">
        <el-table
          v-loading="loading"
          :data="tableData"
          style="width: 100%"
          stripe
          height="100%"
        >
        <el-table-column label="产品信息" width="580">
          <template #default="{ row }">
            <div class="product-cell">
              <el-image
                v-if="row.product_image"
                :src="row.product_image"
                fit="cover"
                class="product-image"
                @mouseenter="showImagePreview(row.product_image, $event)"
                @mouseleave="hideImagePreview"
              />
              <div v-else class="product-image-placeholder">
                <el-icon><Picture /></el-icon>
              </div>
              <div class="product-info">
                <div class="product-sku">
                  <el-tag type="success" size="small">{{ row.sku }}</el-tag>
                </div>
                <div class="product-name" :title="row.product_name">{{ row.product_name || '-' }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="warehouse_name" label="仓库" width="100" />
        <el-table-column label="可用" width="90" align="center">
          <template #default="{ row }">
            <span :class="getStockClass(row.available_stock)" class="stock-value">
              {{ row.available_stock }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="锁定" width="90" align="center">
          <template #default="{ row }">
            <span class="stock-locked">{{ row.locked_stock }}</span>
          </template>
        </el-table-column>
        <el-table-column label="在途" width="90" align="center">
          <template #default="{ row }">
            <span class="stock-transit">{{ row.in_transit_stock }}</span>
          </template>
        </el-table-column>
        <el-table-column label="总计" width="90" align="center">
          <template #default="{ row }">
            <span class="stock-total">{{ row.total_stock }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="updated_at" label="更新时间" min-width="160" />
        </el-table>
      </div>

      <div class="pagination-footer">
        <div class="total-info">
          共 <span class="total-count">{{ total }}</span> 条记录
        </div>
        <el-pagination
          v-if="total > 0"
          background
          layout="prev, pager, next, jumper"
          :total="total"
          :page-size="pageSize"
          :current-page="currentPage"
          @current-change="handlePageChange"
        />
      </div>
    </div>

    <!-- 新建仓库弹窗 -->
    <el-dialog
      v-model="warehouseDialogVisible"
      title="新建仓库"
      width="500px"
      :close-on-click-modal="false"
    >
      <el-form :model="warehouseForm" label-width="100px">
        <el-form-item label="仓库编码" required>
          <el-input v-model="warehouseForm.code" placeholder="如：WH001" />
        </el-form-item>
        <el-form-item label="仓库名称" required>
          <el-input v-model="warehouseForm.name" placeholder="请输入仓库名称" />
        </el-form-item>
        <el-form-item label="仓库类型">
          <el-select v-model="warehouseForm.type" style="width: 100%">
            <el-option
              v-for="item in warehouseTypeOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="仓库地址">
          <el-input v-model="warehouseForm.address" placeholder="请输入仓库地址（选填）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="warehouseDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="warehouseLoading" @click="submitWarehouse">
          确认创建
        </el-button>
      </template>
    </el-dialog>

    <!-- 图片预览 -->
    <ImagePreview ref="imagePreviewRef" />
  </div>
</template>

<style scoped lang="scss">
.inventory-list {
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
  display: flex;
  flex-direction: column;
  height: calc(100vh - 320px);
  min-height: 400px;
}

.table-wrapper {
  flex: 1;
  overflow: hidden;
}

.product-cell {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 4px 0;
}

.product-image {
  width: 56px;
  height: 56px;
  min-width: 56px;
  border-radius: 6px;
  cursor: pointer;
  border: 1px solid var(--border-color);
  transition: transform 0.2s ease;

  &:hover {
    transform: scale(1.05);
  }
}

.product-image-placeholder {
  width: 56px;
  height: 56px;
  min-width: 56px;
  border-radius: 6px;
  background: var(--bg-page);
  border: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-muted);
  font-size: 22px;
}

.product-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.product-sku {
  display: flex;
  align-items: center;
}

.product-name {
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  font-size: 13px;
  line-height: 1.5;
  color: var(--text-primary);
}

.stock-value {
  font-weight: 600;
  font-size: 16px;
}

.stock-normal {
  color: #67c23a;
}

.stock-low {
  color: #e6a23c;
}

.stock-zero {
  color: #f56c6c;
}

.stock-locked {
  color: #909399;
  font-weight: 500;
}

.stock-transit {
  color: #409eff;
  font-weight: 500;
}

.stock-total {
  font-weight: 600;
  font-size: 16px;
  color: var(--text-primary);
}

.pagination-footer {
  padding-top: 16px;
  border-top: 1px solid var(--border-color);
  margin-top: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-shrink: 0;
}

.total-info {
  font-size: 14px;
  color: var(--text-secondary);
}

.total-count {
  font-weight: 600;
  color: var(--color-primary);
}
</style>
