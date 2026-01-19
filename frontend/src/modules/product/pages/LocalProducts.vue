<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import api, { type Product, type CreateProductRequest, type UpdateProductRequest } from '../api'
import { formatDateTime } from '@/utils/format'

defineOptions({ name: 'LocalProducts' })

const loading = ref(false)
const tableData = ref<Product[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)

// 对话框控制
const dialogVisible = ref(false)
const isEdit = ref(false)
const formLoading = ref(false)
const formData = ref<CreateProductRequest & { id?: number }>({
  sku: '',
  name: '',
  image: '',
  cost_price: 0,
  weight: 0,
  dimensions: ''
})

// 获取产品列表
const fetchProducts = async () => {
  loading.value = true
  try {
    const res = await api.listProducts({
      page: currentPage.value,
      page_size: pageSize.value
    })
    tableData.value = res.data.list || []
    total.value = res.data.total || 0
  } catch (error) {
    console.error('获取产品列表失败', error)
  } finally {
    loading.value = false
  }
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

onMounted(() => {
  fetchProducts()
})
</script>

<template>
  <div class="local-products">
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">本地产品管理</h1>
        <p class="page-desc">管理本地产品基础信息</p>
      </div>
      <div class="header-right">
        <el-button type="primary" :icon="Plus" @click="handleCreate">
          新增产品
        </el-button>
      </div>
    </div>

    <div class="content-card">
      <el-table
        v-loading="loading"
        :data="tableData"
        style="width: 100%"
        stripe
      >
        <el-table-column prop="sku" label="SKU" width="180" />
        <el-table-column prop="image" label="图片" width="100">
          <template #default="{ row }">
            <el-image
              v-if="row.image"
              style="width: 50px; height: 50px"
              :src="row.image"
              fit="cover"
              :preview-src-list="[row.image]"
              preview-teleported
            />
            <span v-else>-</span>
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
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEdit(row)">
              编辑
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
      width="500px"
      destroy-on-close
    >
      <el-form :model="formData" label-width="80px">
        <el-form-item label="SKU" required>
          <el-input v-model="formData.sku" :disabled="isEdit" placeholder="请输入SKU" />
        </el-form-item>
        <el-form-item label="名称" required>
          <el-input v-model="formData.name" placeholder="请输入产品名称" />
        </el-form-item>
        <el-form-item label="图片URL">
          <el-input v-model="formData.image" placeholder="请输入图片URL" />
        </el-form-item>
        <el-form-item label="成本价">
          <el-input-number v-model="formData.cost_price" :precision="2" :step="0.1" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="重量(kg)">
          <el-input-number v-model="formData.weight" :precision="3" :step="0.01" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="尺寸">
          <el-input v-model="formData.dimensions" placeholder="例如: 10*10*10" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="formLoading" @click="handleSave">
          保存
        </el-button>
      </template>
    </el-dialog>
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
</style>
