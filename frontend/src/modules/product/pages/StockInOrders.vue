<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Plus, Upload, Download } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import api, { type StockInOrderResponse, type ImportStockInResponse, type WarehouseResponse } from '../api'

defineOptions({ name: 'StockInOrders' })

const router = useRouter()
const loading = ref(false)
const tableData = ref<StockInOrderResponse[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)
const statusFilter = ref('')

const statusOptions = [
  { value: 'pending', label: '待入库' },
  { value: 'completed', label: '已完成' },
  { value: 'cancelled', label: '已取消' },
]

// Excel 导入相关
const importDialogVisible = ref(false)
const importLoading = ref(false)
const importWarehouseId = ref<number | undefined>(undefined)
const importRemark = ref('')
const importFile = ref<File | null>(null)
const importFileList = ref<any[]>([])
const importResult = ref<ImportStockInResponse | null>(null)
const warehouseList = ref<WarehouseResponse[]>([])

const getStatusType = (status: string) => {
  const typeMap: Record<string, string> = {
    'pending': 'warning',
    'completed': 'success',
    'cancelled': 'danger',
  }
  return typeMap[status] || 'info'
}

const getStatusLabel = (status: string) => {
  const labelMap: Record<string, string> = {
    'pending': '待入库',
    'completed': '已完成',
    'cancelled': '已取消',
  }
  return labelMap[status] || status
}

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

const handleReset = () => {
  statusFilter.value = ''
  currentPage.value = 1
  fetchList()
}

const handlePageChange = (page: number) => {
  currentPage.value = page
  fetchList()
}

const getTotalQuantity = (items: StockInOrderResponse['items']) => {
  return items?.reduce((sum, item) => sum + item.quantity, 0) || 0
}

const handleDownloadTemplate = async () => {
  try {
    const res = await api.exportStockInTemplate()
    const blob = new Blob([res.data], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' })
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'stock_in_import_template.xlsx'
    a.click()
    window.URL.revokeObjectURL(url)
  } catch {
    ElMessage.error('下载模板失败')
  }
}

const handleOpenImport = async () => {
  importFile.value = null
  importFileList.value = []
  importWarehouseId.value = undefined
  importRemark.value = ''
  importResult.value = null
  importDialogVisible.value = true

  if (warehouseList.value.length === 0) {
    try {
      const res = await api.listAvailableWarehouses()
      warehouseList.value = res.data.list || []
    } catch {
      console.error('获取仓库列表失败')
    }
  }
}

const handleFileChange = (uploadFile: any) => {
  importFile.value = uploadFile.raw
}

const handleImportSubmit = async () => {
  if (!importWarehouseId.value) {
    ElMessage.warning('请选择入库仓库')
    return
  }
  if (!importFile.value) {
    ElMessage.warning('请上传 Excel 文件')
    return
  }

  importLoading.value = true
  try {
    const res = await api.importStockInOrder(importFile.value, importWarehouseId.value, importRemark.value || undefined)
    importResult.value = res.data
    if (res.data.success_count > 0) {
      ElMessage.success(`导入成功：${res.data.success_count} 个SKU，共 ${res.data.total_qty} 件`)
      fetchList()
    } else {
      ElMessage.warning('没有成功导入的数据')
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '导入失败')
  } finally {
    importLoading.value = false
  }
}

onMounted(() => {
  fetchList()
})
</script>

<template>
  <div class="stock-in-orders">
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">入库单</h1>
        <p class="page-desc">管理系统入库记录</p>
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
        <el-form-item>
          <el-button type="primary" @click="router.push('/warehouse/stock-in/create')">
            <el-icon><Plus /></el-icon>
            新建入库单
          </el-button>
          <el-button type="success" @click="handleOpenImport">
            <el-icon><Upload /></el-icon>
            Excel 导入
          </el-button>
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
        <el-table-column label="入库明细" width="900">
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
        <el-table-column prop="remark" label="备注" width="300">
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

    <!-- Excel 导入对话框 -->
    <el-dialog v-model="importDialogVisible" title="Excel 批量导入入库" width="520px" :close-on-click-modal="false">
      <div class="import-dialog-content">
        <div class="import-step">
          <div class="step-label">1. 下载模板</div>
          <el-button type="primary" link @click="handleDownloadTemplate">
            <el-icon><Download /></el-icon>
            下载导入模板 (.xlsx)
          </el-button>
          <p class="step-hint">模板包含 SKU 和数量两列，请按格式填写</p>
        </div>

        <div class="import-step">
          <div class="step-label">2. 选择入库仓库</div>
          <el-select v-model="importWarehouseId" placeholder="请选择仓库" style="width: 100%">
            <el-option
              v-for="wh in warehouseList"
              :key="wh.id"
              :label="wh.name"
              :value="wh.id"
            />
          </el-select>
        </div>

        <div class="import-step">
          <div class="step-label">3. 上传 Excel 文件</div>
          <el-upload
            v-model:file-list="importFileList"
            :auto-upload="false"
            :limit="1"
            accept=".xlsx,.xls"
            :on-change="handleFileChange"
            :on-remove="() => importFile = null"
            drag
          >
            <el-icon class="el-icon--upload"><Upload /></el-icon>
            <div class="el-upload__text">拖拽文件到此处或 <em>点击上传</em></div>
            <template #tip>
              <div class="el-upload__tip">仅支持 .xlsx / .xls 格式</div>
            </template>
          </el-upload>
        </div>

        <div class="import-step">
          <div class="step-label">4. 备注（选填）</div>
          <el-input v-model="importRemark" placeholder="选填" />
        </div>

        <!-- 导入结果 -->
        <div v-if="importResult" class="import-result">
          <el-divider />
          <div class="result-title">导入结果</div>
          <div class="result-stats">
            <el-tag type="info" size="small">入库单号: {{ importResult.order_no }}</el-tag>
            <el-tag type="success" size="small">成功: {{ importResult.success_count }} 个SKU</el-tag>
            <el-tag v-if="importResult.fail_count > 0" type="danger" size="small">失败: {{ importResult.fail_count }}</el-tag>
            <el-tag type="primary" size="small">总数量: {{ importResult.total_qty }} 件</el-tag>
          </div>
          <div v-if="importResult.fail_reasons?.length" class="fail-reasons">
            <div v-for="(reason, idx) in importResult.fail_reasons" :key="idx" class="fail-reason-item">
              {{ reason }}
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <el-button @click="importDialogVisible = false">关闭</el-button>
        <el-button type="primary" :loading="importLoading" @click="handleImportSubmit" :disabled="!!importResult">
          确认导入
        </el-button>
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

.import-dialog-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.import-step {
  .step-label {
    font-size: 14px;
    font-weight: 500;
    margin-bottom: 8px;
    color: var(--text-primary);
  }

  .step-hint {
    margin: 6px 0 0;
    font-size: 12px;
    color: var(--text-muted);
  }
}

.import-result {
  .result-title {
    font-size: 14px;
    font-weight: 600;
    margin-bottom: 10px;
  }

  .result-stats {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }

  .fail-reasons {
    margin-top: 10px;
    max-height: 120px;
    overflow-y: auto;
    background: var(--el-color-danger-light-9);
    border-radius: var(--radius-sm);
    padding: 8px 12px;
  }

  .fail-reason-item {
    font-size: 12px;
    color: var(--el-color-danger);
    line-height: 1.8;
  }
}
</style>
