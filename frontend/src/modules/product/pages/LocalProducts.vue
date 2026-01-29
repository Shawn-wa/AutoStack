<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Picture, CopyDocument, Link } from '@element-plus/icons-vue'
import api, { type Product, type CreateProductRequest, type UpdateProductRequest, type WarehouseResponse, type Supplier, type CreateSupplierRequest, type UpdateSupplierRequest } from '../api'
import { formatDateTime } from '@/utils/format'
import ImagePreview from '@/components/ImagePreview.vue'

defineOptions({ name: 'LocalProducts' })

const imagePreviewRef = ref<InstanceType<typeof ImagePreview>>()
const showImagePreview = (src: string, event: MouseEvent) => {
  imagePreviewRef.value?.show(src, event)
}
const hideImagePreview = () => {
  imagePreviewRef.value?.hide()
}

const loading = ref(false)
const tableData = ref<Product[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)

// 筛选条件
const keyword = ref('')

// 仓库列表
const warehouseList = ref<WarehouseResponse[]>([])

// 对话框控制
const dialogVisible = ref(false)
const isEdit = ref(false)
const formLoading = ref(false)
const formData = ref<CreateProductRequest & { id?: number }>({
  wid: 0,
  sku: '',
  name: '',
  image: '',
  cost_price: 0,
  weight: 0,
  dimensions: ''
})

// 供应商对话框控制
const supplierDialogVisible = ref(false)
const supplierLoading = ref(false)
const supplierList = ref<Supplier[]>([])
const currentProduct = ref<Product | null>(null)
const supplierFormVisible = ref(false)
const supplierFormLoading = ref(false)
const isEditSupplier = ref(false)
const supplierFormData = ref<CreateSupplierRequest & { id?: number }>({
  product_id: 0,
  supplier_name: '',
  purchase_link: '',
  unit_price: 0,
  currency: 'CNY',
  min_order_qty: 1,
  lead_time: 0,
  estimated_days: 0,
  remark: '',
  is_default: false
})

// 获取仓库列表
const fetchWarehouses = async () => {
  try {
    const res = await api.listAvailableWarehouses()
    warehouseList.value = res.data.list || []
  } catch (error) {
    console.error('获取仓库列表失败', error)
  }
}

// 获取仓库名称
const getWarehouseName = (wid: number) => {
  const warehouse = warehouseList.value.find(w => w.id === wid)
  return warehouse?.name || '-'
}

// 复制 SKU
const copySku = async (sku: string) => {
  try {
    await navigator.clipboard.writeText(sku)
    ElMessage.success('已复制')
  } catch {
    ElMessage.error('复制失败')
  }
}

// 获取产品列表
const fetchProducts = async () => {
  loading.value = true
  try {
    const params: any = {
      page: currentPage.value,
      page_size: pageSize.value
    }
    if (keyword.value) {
      params.keyword = keyword.value
    }
    const res = await api.listProducts(params)
    tableData.value = res.data.list || []
    total.value = res.data.total || 0
  } catch (error) {
    console.error('获取产品列表失败', error)
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  currentPage.value = 1
  fetchProducts()
}

// 重置筛选
const handleReset = () => {
  keyword.value = ''
  currentPage.value = 1
  fetchProducts()
}

// 分页变化
const handlePageChange = (page: number) => {
  currentPage.value = page
  fetchProducts()
}

// 每页数量变化
const handleSizeChange = (size: number) => {
  pageSize.value = size
  currentPage.value = 1
  fetchProducts()
}

// 打开创建对话框
const handleCreate = () => {
  isEdit.value = false
  formData.value = {
    wid: warehouseList.value.length > 0 ? warehouseList.value[0].id : 0,
    sku: '',
    name: '',
    image: '',
    cost_price: 0,
    weight: 0,
    dimensions: ''
  }
  dialogVisible.value = true
}

// 打开编辑对话框
const handleEdit = (row: Product) => {
  isEdit.value = true
  formData.value = {
    id: row.id,
    wid: row.wid,
    sku: row.sku,
    name: row.name,
    image: row.image,
    cost_price: row.cost_price,
    weight: row.weight,
    dimensions: row.dimensions
  }
  dialogVisible.value = true
}

// 保存表单
const handleSave = async () => {
  if (!formData.value.sku || !formData.value.name) {
    ElMessage.warning('请填写SKU和名称')
    return
  }

  formLoading.value = true
  try {
    if (isEdit.value && formData.value.id) {
      await api.updateProduct(formData.value.id, formData.value)
      ElMessage.success('更新成功')
    } else {
      await api.createProduct(formData.value)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchProducts()
  } catch (error: any) {
    console.error('保存失败', error)
  } finally {
    formLoading.value = false
  }
}

// 删除产品
const handleDelete = async (row: Product) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除产品 "${row.sku}" 吗？`,
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await api.deleteProduct(row.id)
    ElMessage.success('删除成功')
    fetchProducts()
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('删除产品失败', error)
    }
  }
}

// ========== 供应商/采购信息相关 ==========

// 打开供应商管理对话框
const handleSuppliers = async (row: Product) => {
  currentProduct.value = row
  supplierDialogVisible.value = true
  await fetchSuppliers()
}

// 获取供应商列表
const fetchSuppliers = async () => {
  if (!currentProduct.value) return
  supplierLoading.value = true
  try {
    const res = await api.getProductSuppliers(currentProduct.value.id)
    supplierList.value = res.data || []
  } catch (error) {
    console.error('获取供应商列表失败', error)
  } finally {
    supplierLoading.value = false
  }
}

// 打开添加供应商表单
const handleAddSupplier = () => {
  if (!currentProduct.value) return
  isEditSupplier.value = false
  supplierFormData.value = {
    product_id: currentProduct.value.id,
    supplier_name: '',
    purchase_link: '',
    unit_price: 0,
    currency: 'CNY',
    min_order_qty: 1,
    lead_time: 0,
    estimated_days: 0,
    remark: '',
    is_default: supplierList.value.length === 0
  }
  supplierFormVisible.value = true
}

// 编辑供应商
const handleEditSupplier = (row: Supplier) => {
  isEditSupplier.value = true
  supplierFormData.value = {
    id: row.id,
    product_id: row.product_id,
    supplier_name: row.supplier_name,
    purchase_link: row.purchase_link,
    unit_price: row.unit_price,
    currency: row.currency,
    min_order_qty: row.min_order_qty,
    lead_time: row.lead_time,
    estimated_days: row.estimated_days,
    remark: row.remark,
    is_default: row.is_default
  }
  supplierFormVisible.value = true
}

// 保存供应商
const handleSaveSupplier = async () => {
  if (!supplierFormData.value.supplier_name) {
    ElMessage.warning('请填写供应商名称')
    return
  }

  supplierFormLoading.value = true
  try {
    if (isEditSupplier.value && supplierFormData.value.id) {
      await api.updateSupplier(supplierFormData.value.id, supplierFormData.value as UpdateSupplierRequest)
      ElMessage.success('更新成功')
    } else {
      await api.createSupplier(supplierFormData.value)
      ElMessage.success('添加成功')
    }
    supplierFormVisible.value = false
    await fetchSuppliers()
  } catch (error: any) {
    console.error('保存供应商失败', error)
  } finally {
    supplierFormLoading.value = false
  }
}

// 删除供应商
const handleDeleteSupplier = async (row: Supplier) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除供应商 "${row.supplier_name}" 吗？`,
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await api.deleteSupplier(row.id)
    ElMessage.success('删除成功')
    await fetchSuppliers()
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('删除供应商失败', error)
    }
  }
}

// 打开采购链接
const openPurchaseLink = (url: string) => {
  if (url) {
    window.open(url, '_blank')
  }
}

onMounted(() => {
  fetchProducts()
  fetchWarehouses()
})
</script>

<template>
  <div class="local-products">
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">系统产品管理</h1>
        <p class="page-desc">管理系统产品基础信息</p>
      </div>
      <div class="header-right">
        <el-button type="primary" :icon="Plus" @click="handleCreate">
          新增产品
        </el-button>
      </div>
    </div>

    <div class="filter-card">
      <el-form inline>
        <el-form-item label="关键词">
          <el-input
            v-model="keyword"
            placeholder="SKU/名称"
            clearable
            style="width: 200px"
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
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
        <el-table-column label="仓库" width="120">
          <template #default="{ row }">
            <span>{{ row.warehouse_name || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="SKU" width="300">
          <template #default="{ row }">
            <div class="sku-cell">
              <el-tooltip :content="row.sku" placement="top" :show-after="300">
                <span class="sku-text">{{ row.sku }}</span>
              </el-tooltip>
              <el-icon class="copy-icon" @click.stop="copySku(row.sku)"><CopyDocument /></el-icon>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="image" label="图片" width="100">
          <template #default="{ row }">
            <el-image
              v-if="row.image"
              style="width: 50px; height: 50px; cursor: pointer"
              :src="row.image"
              fit="cover"
              @mouseenter="showImagePreview(row.image, $event)"
              @mouseleave="hideImagePreview"
            />
            <div v-else class="image-placeholder">
              <el-icon><Picture /></el-icon>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="名称" min-width="200" show-overflow-tooltip />
        <el-table-column prop="cost_price" label="成本价" width="120">
          <template #default="{ row }">
            {{ row.cost_price }}
          </template>
        </el-table-column>
        <el-table-column prop="weight" label="重量(kg)" width="100" />
        <el-table-column prop="dimensions" label="尺寸(cm)" width="150" />
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEdit(row)">
              编辑
            </el-button>
            <el-button type="success" link size="small" @click="handleSuppliers(row)">
              采购
            </el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">
              删除
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

    <!-- 编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑产品' : '新增产品'"
      width="520px"
      destroy-on-close
      :close-on-click-modal="false"
    >
      <el-form :model="formData" label-width="90px">
        <el-form-item label="所属仓库" required>
          <el-select v-model="formData.wid" placeholder="请选择仓库" style="width: 200px">
            <el-option
              v-for="wh in warehouseList"
              :key="wh.id"
              :label="wh.name"
              :value="wh.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="SKU" required>
          <el-input v-model="formData.sku" :disabled="isEdit" placeholder="请输入SKU" />
        </el-form-item>
        <el-form-item label="产品名称" required>
          <el-input v-model="formData.name" placeholder="请输入产品名称" />
        </el-form-item>
        <el-form-item label="产品图片">
          <div class="image-form-item">
            <div 
              class="image-preview-small" 
              :class="{ 'has-image': formData.image }"
              @mouseenter="formData.image && showImagePreview(formData.image, $event)"
              @mouseleave="hideImagePreview"
            >
              <el-image
                v-if="formData.image"
                :src="formData.image"
                fit="contain"
              />
              <el-icon v-else :size="24"><Picture /></el-icon>
            </div>
            <el-input 
              v-model="formData.image" 
              placeholder="请输入图片URL"
              class="image-input"
            />
          </div>
        </el-form-item>
        <el-form-item label="成本价">
          <el-input-number
            v-model="formData.cost_price"
            :precision="2"
            :step="0.1"
            :min="0"
            controls-position="right"
            style="width: 160px"
          />
          <span class="form-unit">元</span>
        </el-form-item>
        <el-form-item label="重量">
          <el-input-number
            v-model="formData.weight"
            :precision="3"
            :step="0.01"
            :min="0"
            controls-position="right"
            style="width: 160px"
          />
          <span class="form-unit">kg</span>
        </el-form-item>
        <el-form-item label="尺寸">
          <el-input v-model="formData.dimensions" placeholder="例如: 10*10*10 cm" style="width: 200px" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="formLoading" @click="handleSave">
          保存
        </el-button>
      </template>
    </el-dialog>

    <!-- 供应商管理对话框 -->
    <el-dialog
      v-model="supplierDialogVisible"
      :title="`采购信息 - ${currentProduct?.sku || ''}`"
      width="1100px"
      destroy-on-close
    >
      <div class="supplier-header">
        <el-button type="primary" size="small" :icon="Plus" @click="handleAddSupplier">
          添加采购渠道
        </el-button>
      </div>
      
      <el-table
        v-loading="supplierLoading"
        :data="supplierList"
        style="width: 100%"
        stripe
        size="small"
      >
        <el-table-column label="供应商/店铺" prop="supplier_name" width="180" show-overflow-tooltip />
        <el-table-column label="采购链接" width="100">
          <template #default="{ row }">
            <el-button
              v-if="row.purchase_link"
              type="primary"
              link
              size="small"
              :icon="Link"
              @click="openPurchaseLink(row.purchase_link)"
            >
              打开链接
            </el-button>
            <span v-else class="text-secondary">-</span>
          </template>
        </el-table-column>
        <el-table-column label="采购单价" width="130" align="right">
          <template #default="{ row }">
            {{ row.unit_price?.toFixed(2) }} {{ row.currency }}
          </template>
        </el-table-column>
        <el-table-column label="起订量" prop="min_order_qty" width="80" align="center" />
        <el-table-column label="交货周期" width="90" align="center">
          <template #default="{ row }">
            {{ row.lead_time ? `${row.lead_time}天` : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="预估时效" width="90" align="center">
          <template #default="{ row }">
            {{ row.estimated_days ? `${row.estimated_days}天` : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="默认" width="60" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.is_default" type="success" size="small">是</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="备注" prop="remark" min-width="150" show-overflow-tooltip />
        <el-table-column label="操作" width="120">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEditSupplier(row)">编辑</el-button>
            <el-button type="danger" link size="small" @click="handleDeleteSupplier(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <template v-if="supplierList.length === 0 && !supplierLoading">
        <el-empty description="暂无采购信息" :image-size="80" />
      </template>
    </el-dialog>

    <!-- 添加/编辑供应商对话框 -->
    <el-dialog
      v-model="supplierFormVisible"
      :title="isEditSupplier ? '编辑采购渠道' : '添加采购渠道'"
      width="520px"
      destroy-on-close
      :close-on-click-modal="false"
    >
      <el-form :model="supplierFormData" label-width="100px">
        <el-form-item label="供应商名称" required>
          <el-input v-model="supplierFormData.supplier_name" placeholder="请输入供应商/店铺名称" />
        </el-form-item>
        <el-form-item label="采购链接">
          <el-input v-model="supplierFormData.purchase_link" placeholder="请输入采购链接URL" />
        </el-form-item>
        <el-form-item label="采购单价">
          <el-input-number
            v-model="supplierFormData.unit_price"
            :precision="2"
            :step="0.1"
            :min="0"
            controls-position="right"
            style="width: 140px"
          />
          <el-select v-model="supplierFormData.currency" style="width: 80px; margin-left: 8px">
            <el-option label="CNY" value="CNY" />
            <el-option label="USD" value="USD" />
            <el-option label="RUB" value="RUB" />
          </el-select>
        </el-form-item>
        <el-form-item label="最小起订量">
          <el-input-number
            v-model="supplierFormData.min_order_qty"
            :min="1"
            :step="1"
            controls-position="right"
            style="width: 140px"
          />
          <span class="form-unit">件</span>
        </el-form-item>
        <el-form-item label="交货周期">
          <el-input-number
            v-model="supplierFormData.lead_time"
            :min="0"
            :step="1"
            controls-position="right"
            style="width: 140px"
          />
          <span class="form-unit">天</span>
        </el-form-item>
        <el-form-item label="预估时效">
          <el-input-number
            v-model="supplierFormData.estimated_days"
            :min="0"
            :step="1"
            controls-position="right"
            style="width: 140px"
          />
          <span class="form-unit">天</span>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="supplierFormData.remark" type="textarea" :rows="2" placeholder="备注信息" />
        </el-form-item>
        <el-form-item label="设为默认">
          <el-switch v-model="supplierFormData.is_default" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="supplierFormVisible = false">取消</el-button>
        <el-button type="primary" :loading="supplierFormLoading" @click="handleSaveSupplier">
          保存
        </el-button>
      </template>
    </el-dialog>

    <!-- 图片预览 -->
    <ImagePreview ref="imagePreviewRef" />
  </div>
</template>

<style scoped lang="scss">
.local-products {
  padding: 0;
}

.page-header {
  margin-bottom: 24px;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
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

.pagination-wrapper {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.sku-cell {
  display: flex;
  align-items: center;
  gap: 6px;
  width: 100%;
}

.sku-text {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.copy-icon {
  flex-shrink: 0;
  cursor: pointer;
  color: var(--text-secondary);
  font-size: 14px;
  opacity: 0;
  transition: opacity 0.2s, color 0.2s;
  
  &:hover {
    color: var(--color-primary);
  }
}

.el-table__row:hover .copy-icon {
  opacity: 1;
}

.image-placeholder {
  width: 50px;
  height: 50px;
  border-radius: 4px;
  background: var(--bg-page);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-secondary);
  font-size: 20px;
}

// 表单样式
.image-form-item {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
}

.image-preview-small {
  width: 60px;
  height: 60px;
  flex-shrink: 0;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  background: var(--bg-page);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  color: var(--text-placeholder);
  transition: border-color 0.2s;
  
  .el-image {
    width: 100%;
    height: 100%;
  }
  
  &.has-image {
    cursor: pointer;
    
    &:hover {
      border-color: var(--color-primary);
    }
  }
}

.image-input {
  flex: 1;
}

.form-unit {
  margin-left: 8px;
  color: var(--text-secondary);
  font-size: 14px;
}

// 供应商对话框样式
.supplier-header {
  margin-bottom: 16px;
  display: flex;
  justify-content: flex-end;
}

.text-secondary {
  color: var(--text-secondary);
}
</style>
