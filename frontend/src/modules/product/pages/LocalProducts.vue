<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Picture, CopyDocument, Upload, Download } from '@element-plus/icons-vue'
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
const filterWarehouseId = ref<number | ''>('')

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
  shipping_fee: 0,
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

// 格式化URL，确保有协议前缀
const formatUrl = (url: string) => {
  if (!url) return ''
  if (url.startsWith('http://') || url.startsWith('https://')) {
    return url
  }
  return 'https://' + url
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
    if (filterWarehouseId.value) {
      params.wid = filterWarehouseId.value
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
  filterWarehouseId.value = ''
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
    shipping_fee: 0,
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
    shipping_fee: row.shipping_fee,
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

// ========== 批量导入相关 ==========
const importDialogVisible = ref(false)
const importLoading = ref(false)
const importFile = ref<File | null>(null)
const importResult = ref<{ total_count: number; success_count: number; fail_count: number; fail_reasons?: string[] } | null>(null)

// 打开导入对话框
const handleOpenImport = () => {
  importFile.value = null
  importResult.value = null
  importDialogVisible.value = true
}

// 下载导入模板
const handleDownloadTemplate = async () => {
  try {
    const res = await api.exportSupplierTemplate()
    const blob = new Blob([res.data], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = 'supplier_import_template.xlsx'
    link.click()
    window.URL.revokeObjectURL(url)
  } catch (error) {
    console.error('下载模板失败', error)
    ElMessage.error('下载模板失败')
  }
}

// 处理文件选择
const handleFileChange = (file: any) => {
  importFile.value = file.raw
}

// 执行导入
const handleImport = async () => {
  if (!importFile.value) {
    ElMessage.warning('请选择要导入的文件')
    return
  }

  importLoading.value = true
  try {
    const res = await api.importSuppliers(importFile.value)
    importResult.value = res.data
    ElMessage.success(`导入完成：成功 ${res.data.success_count} 条，失败 ${res.data.fail_count} 条`)
    if (res.data.success_count > 0) {
      fetchProducts()
    }
  } catch (error: any) {
    console.error('导入失败', error)
    ElMessage.error('导入失败')
  } finally {
    importLoading.value = false
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
        <el-button :icon="Upload" @click="handleOpenImport">
          批量导入采购信息
        </el-button>
        <el-button type="primary" :icon="Plus" @click="handleCreate">
          新增产品
        </el-button>
      </div>
    </div>

    <div class="filter-card">
      <el-form inline>
        <el-form-item label="仓库">
          <el-select
            v-model="filterWarehouseId"
            placeholder="全部仓库"
            clearable
            style="width: 150px"
            @change="handleSearch"
          >
            <el-option
              v-for="wh in warehouseList"
              :key="wh.id"
              :label="wh.name"
              :value="wh.id"
            />
          </el-select>
        </el-form-item>
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
      :title="`采购信息 - ${currentProduct?.sku}（${currentProduct?.name || ''}）`"
      width="1300px"
      destroy-on-close
      class="supplier-dialog"
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
      >
        <el-table-column label="供应商/店铺" prop="supplier_name" min-width="180" show-overflow-tooltip />
        <el-table-column label="采购链接" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <a
              v-if="row.purchase_link"
              :href="formatUrl(row.purchase_link)"
              target="_blank"
              rel="noopener noreferrer"
              class="purchase-link"
            >{{ row.purchase_link }}</a>
            <span v-else class="text-secondary">-</span>
          </template>
        </el-table-column>
        <el-table-column label="采购单价" width="130" align="right">
          <template #default="{ row }">
            {{ row.unit_price?.toFixed(2) }} {{ row.currency }}
          </template>
        </el-table-column>
        <el-table-column label="物流费" width="120" align="right">
          <template #default="{ row }">
            {{ row.shipping_fee?.toFixed(2) }} {{ row.currency }}
          </template>
        </el-table-column>
        <el-table-column label="起订量" prop="min_order_qty" width="90" align="center" />
        <el-table-column label="交货周期" width="100" align="center">
          <template #default="{ row }">
            {{ row.lead_time ? `${row.lead_time}天` : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="预估时效" width="100" align="center">
          <template #default="{ row }">
            {{ row.estimated_days ? `${row.estimated_days}天` : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="默认" width="70" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.is_default" type="success" size="small">是</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="备注" prop="remark" min-width="180" show-overflow-tooltip />
        <el-table-column label="操作" width="130" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="handleEditSupplier(row)">编辑</el-button>
            <el-button type="danger" link @click="handleDeleteSupplier(row)">删除</el-button>
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
        <el-form-item label="物流费">
          <el-input-number
            v-model="supplierFormData.shipping_fee"
            :precision="2"
            :step="0.1"
            :min="0"
            controls-position="right"
            style="width: 140px"
          />
          <span class="form-unit">元/件</span>
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

    <!-- 批量导入对话框 -->
    <el-dialog
      v-model="importDialogVisible"
      title="批量导入采购信息"
      width="520px"
      destroy-on-close
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
            <p class="step-tip">请按照模板格式填写数据，SKU 必须与系统中已有产品匹配</p>
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
:deep(.supplier-dialog) {
  .el-dialog__body {
    padding: 16px 24px 24px;
  }
}

.supplier-header {
  margin-bottom: 16px;
  display: flex;
  justify-content: flex-end;
}

.text-secondary {
  color: var(--text-secondary);
}

.purchase-link {
  color: var(--el-color-primary);
  text-decoration: none;
  &:hover {
    text-decoration: underline;
  }
}

// 导入对话框样式
.import-content {
  padding: 0 8px;
}

.import-step {
  margin-bottom: 24px;
  
  &:last-child {
    margin-bottom: 0;
  }
}

.step-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.step-num {
  width: 24px;
  height: 24px;
  background: var(--color-primary);
  color: #fff;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 600;
}

.step-title {
  font-weight: 500;
  font-size: 14px;
}

.step-body {
  padding-left: 36px;
}

.step-tip {
  margin-top: 8px;
  font-size: 12px;
  color: var(--text-secondary);
}

.upload-area {
  width: 100%;
}

.import-result {
  margin-top: 20px;
  padding: 16px;
  background: var(--bg-page);
  border-radius: var(--radius-md);
}

.result-header {
  font-weight: 500;
  margin-bottom: 12px;
}

.result-stats {
  display: flex;
  gap: 24px;
  margin-bottom: 12px;
}

.stat-item {
  font-size: 14px;
  
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
  padding-top: 12px;
  border-top: 1px solid var(--border-color);
}

.error-title {
  font-size: 12px;
  color: var(--text-secondary);
  margin-bottom: 8px;
}

.error-item {
  font-size: 12px;
  color: var(--el-color-danger);
  padding: 4px 0;
}

.error-more {
  font-size: 12px;
  color: var(--text-secondary);
  margin-top: 4px;
}
</style>
