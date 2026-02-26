<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { Delete, Search, Upload, Download } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouter } from 'vue-router'
import api, { type Product, type StockInOrderItemRequest, type WarehouseResponse } from '../api'

defineOptions({ name: 'StockInCreate' })

const router = useRouter()

const dialogLoading = ref(false)
const selectedWarehouseId = ref<number | undefined>(undefined)
const searchKeyword = ref('')
const searchLoading = ref(false)
const searchResults = ref<Product[]>([])
const selectedProducts = ref<Array<{ product: Product; quantity: number }>>([])
const remark = ref('')

const warehouseList = ref<WarehouseResponse[]>([])

const dialogTotalQuantity = computed(() => {
  return selectedProducts.value.reduce((sum, item) => sum + item.quantity, 0)
})

const fetchWarehouses = async () => {
  try {
    const res = await api.listAvailableWarehouses()
    warehouseList.value = res.data.list || []
  } catch (error) {
    console.error('获取仓库列表失败', error)
  }
}

const handleSearch = async (query: string) => {
  if (!query || query.length < 1) {
    searchResults.value = []
    return
  }

  searchLoading.value = true
  try {
    const res = await api.listProducts({ page: 1, page_size: 20, keyword: query } as any)
    searchResults.value = res.data.list || []
  } catch (error) {
    console.error('搜索产品失败', error)
    searchResults.value = []
  } finally {
    searchLoading.value = false
  }
}

const handleSelectProduct = (product: Product) => {
  const exists = selectedProducts.value.find(item => item.product.id === product.id)
  if (exists) {
    ElMessage.warning('该产品已添加')
    return
  }

  selectedProducts.value.push({
    product,
    quantity: 1
  })

  searchKeyword.value = ''
  searchResults.value = []
}

const handleRemoveProduct = (index: number) => {
  selectedProducts.value.splice(index, 1)
}

const handleSubmit = async () => {
  if (!selectedWarehouseId.value) {
    ElMessage.warning('请选择入库仓库')
    return
  }

  if (selectedProducts.value.length === 0) {
    ElMessage.warning('请至少添加一个产品')
    return
  }

  for (const item of selectedProducts.value) {
    if (item.quantity <= 0) {
      ElMessage.warning(`产品 ${item.product.sku} 的入库数量必须大于0`)
      return
    }
  }

  const warehouse = warehouseList.value.find(w => w.id === selectedWarehouseId.value)
  const warehouseName = warehouse?.name || '未知仓库'

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
    return
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
    router.push('/warehouse/stock-in')
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '创建失败')
  } finally {
    dialogLoading.value = false
  }
}

// ========== 批量导入相关 ==========
const importDialogVisible = ref(false)
const importLoading = ref(false)
const importFile = ref<File | null>(null)
const importResult = ref<{ total_count: number; success_count: number; fail_count: number; fail_reasons?: string[] } | null>(null)

const handleOpenImport = () => {
  importFile.value = null
  importResult.value = null
  importDialogVisible.value = true
}

const handleDownloadTemplate = async () => {
  try {
    const res = await api.exportStockInTemplate()
    const blob = new Blob([res.data], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = 'stock_in_import_template.xlsx'
    link.click()
    window.URL.revokeObjectURL(url)
  } catch (error) {
    console.error('下载模板失败', error)
    ElMessage.error('下载模板失败')
  }
}

const handleFileChange = (file: any) => {
  importFile.value = file.raw
}

const handleImport = async () => {
  if (!importFile.value) {
    ElMessage.warning('请选择要导入的文件')
    return
  }
  if (!selectedWarehouseId.value) {
    ElMessage.warning('请先选择入库仓库')
    return
  }

  importLoading.value = true
  try {
    const res = await api.importStockInOrders(importFile.value, selectedWarehouseId.value, remark.value || undefined)
    importResult.value = res.data
    ElMessage.success(`导入完成：成功 ${res.data.success_count} 条，失败 ${res.data.fail_count} 条`)
  } catch (error: any) {
    console.error('导入失败', error)
    ElMessage.error(error.response?.data?.message || '导入失败')
  } finally {
    importLoading.value = false
  }
}

onMounted(() => {
  fetchWarehouses()
})
</script>

<template>
  <div class="stock-in-create">
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">新建入库单</h1>
        <p class="page-desc">创建新的入库记录</p>
      </div>
      <div class="header-actions">
        <el-button :icon="Upload" @click="handleOpenImport">
          Excel批量导入
        </el-button>
      </div>
    </div>

    <div class="create-card">
      <div class="create-form">
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

        <div class="form-section">
          <label class="form-label">备注</label>
          <el-input
            v-model="remark"
            type="textarea"
            :rows="2"
            placeholder="选填"
          />
        </div>

        <div class="form-actions">
          <el-button @click="router.push('/warehouse/stock-in')">取消</el-button>
          <el-button type="primary" :loading="dialogLoading" @click="handleSubmit">
            确认入库
          </el-button>
        </div>
      </div>
    </div>

    <!-- 批量导入弹窗 -->
    <el-dialog
      v-model="importDialogVisible"
      title="Excel批量导入入库"
      width="560px"
      :close-on-click-modal="false"
    >
      <div class="import-content">
        <div class="import-step">
          <div class="step-header">
            <span class="step-num">1</span>
            <span class="step-title">下载导入模板</span>
          </div>
          <div class="step-body">
            <el-button :icon="Download" @click="handleDownloadTemplate">
              下载模板
            </el-button>
            <p class="step-tip">模板包含 SKU 和数量两列，SKU 必须与系统中已有产品匹配</p>
          </div>
        </div>

        <div class="import-step">
          <div class="step-header">
            <span class="step-num">2</span>
            <span class="step-title">上传文件</span>
          </div>
          <div class="step-body">
            <el-upload
              class="upload-area"
              drag
              :auto-upload="false"
              :limit="1"
              accept=".xlsx,.xls"
              :on-change="handleFileChange"
            >
              <el-icon class="el-icon--upload"><Upload /></el-icon>
              <div class="el-upload__text">
                拖拽文件到此处，或 <em>点击上传</em>
              </div>
              <template #tip>
                <div class="el-upload__tip">
                  仅支持 .xlsx / .xls 格式
                </div>
              </template>
            </el-upload>
            <p class="step-tip">请确保已在上方选择入库仓库后再导入</p>
          </div>
        </div>

        <div v-if="importResult" class="import-result">
          <div class="result-header">导入结果</div>
          <div class="result-stats">
            <span class="stat-item success">成功：{{ importResult.success_count }}</span>
            <span class="stat-item fail">失败：{{ importResult.fail_count }}</span>
            <span class="stat-item total">总计：{{ importResult.total_count }}</span>
          </div>
          <div v-if="importResult.fail_reasons && importResult.fail_reasons.length > 0" class="result-errors">
            <div class="error-title">失败原因：</div>
            <div v-for="(reason, index) in importResult.fail_reasons.slice(0, 10)" :key="index" class="error-item">
              {{ reason }}
            </div>
            <div v-if="importResult.fail_reasons.length > 10" class="error-more">
              ... 还有 {{ importResult.fail_reasons.length - 10 }} 条错误
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <el-button @click="importDialogVisible = false">关闭</el-button>
        <el-button type="primary" :loading="importLoading" @click="handleImport">
          开始导入
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped lang="scss">
.stock-in-create {
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

.create-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: 24px;
}

.create-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
  max-width: 720px;
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
  max-height: 400px;
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

.form-actions {
  display: flex;
  gap: 12px;
  padding-top: 12px;
  border-top: 1px solid var(--border-color);
}

.import-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.import-step {
  .step-header {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: 12px;
  }

  .step-num {
    width: 24px;
    height: 24px;
    border-radius: 50%;
    background: var(--color-primary);
    color: #fff;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 13px;
    font-weight: 600;
    flex-shrink: 0;
  }

  .step-title {
    font-size: 14px;
    font-weight: 500;
  }

  .step-body {
    padding-left: 34px;
  }

  .step-tip {
    font-size: 12px;
    color: var(--text-secondary);
    margin-top: 8px;
    margin-bottom: 0;
  }
}

.import-result {
  background: var(--bg-page);
  border-radius: var(--radius-md);
  padding: 16px;
  border: 1px solid var(--border-color);

  .result-header {
    font-weight: 600;
    font-size: 14px;
    margin-bottom: 12px;
  }

  .result-stats {
    display: flex;
    gap: 24px;
    margin-bottom: 12px;
  }

  .stat-item {
    font-size: 14px;
    font-weight: 500;

    &.success {
      color: var(--el-color-success);
    }

    &.fail {
      color: var(--el-color-danger);
    }

    &.total {
      color: var(--text-secondary);
    }
  }

  .result-errors {
    margin-top: 8px;
    padding-top: 8px;
    border-top: 1px solid var(--border-color);
  }

  .error-title {
    font-size: 13px;
    font-weight: 500;
    margin-bottom: 6px;
    color: var(--el-color-danger);
  }

  .error-item {
    font-size: 12px;
    color: var(--text-secondary);
    padding: 2px 0;
  }

  .error-more {
    font-size: 12px;
    color: var(--text-muted);
    margin-top: 4px;
  }
}
</style>
