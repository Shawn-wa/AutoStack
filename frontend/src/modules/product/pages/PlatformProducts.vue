<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Link, Delete, Search, Picture } from '@element-plus/icons-vue'
import api, { type PlatformProduct, type Product } from '../api'
import { getAuths, type AuthResponse } from '@/modules/order/api'
import { formatDateTime } from '@/utils/format'
import ImagePreview from '@/components/ImagePreview.vue'

defineOptions({ name: 'PlatformProducts' })

const imagePreviewRef = ref<InstanceType<typeof ImagePreview>>()
const showImagePreview = (src: string) => {
  imagePreviewRef.value?.show(src)
}

const loading = ref(false)
const tableData = ref<PlatformProduct[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)
const authId = ref<number | undefined>(undefined)
const authOptions = ref<AuthResponse[]>([])
const keyword = ref('')
const syncLoading = ref(false)

// 映射对话框
const mapDialogVisible = ref(false)
const mapLoading = ref(false)
const currentPlatformProduct = ref<PlatformProduct | null>(null)
const selectedProductId = ref<number | undefined>(undefined)
const productOptions = ref<Product[]>([])
const productLoading = ref(false)

// 获取授权列表
const fetchAuths = async () => {
  try {
    const res = await getAuths()
    authOptions.value = res.data.list || []
    if (authOptions.value.length > 0) {
      authId.value = authOptions.value[0].id
      fetchProducts()
    }
  } catch (error) {
    console.error('获取授权列表失败', error)
  }
}

// 获取平台产品列表
const fetchProducts = async () => {
  if (!authId.value) return
  loading.value = true
  try {
    const res = await api.listPlatformProducts({
      page: currentPage.value,
      page_size: pageSize.value,
      platform_auth_id: authId.value,
      keyword: keyword.value || undefined
    })
    tableData.value = res.data.list || []
    total.value = res.data.total || 0
  } catch (error) {
    console.error('获取平台产品列表失败', error)
  } finally {
    loading.value = false
  }
}

// 同步产品
const handleSync = async () => {
  if (!authId.value) return
  syncLoading.value = true
  try {
    await api.syncProducts(authId.value)
    ElMessage.success('同步任务已创建，请稍后刷新查看')
  } catch (error) {
    console.error('同步失败', error)
  } finally {
    syncLoading.value = false
  }
}

// 刷新列表
const handleRefresh = () => {
  fetchProducts()
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

// 搜索本地产品
const searchProducts = async (query: string) => {
  if (!query) {
    productOptions.value = []
    return
  }
  productLoading.value = true
  try {
    // 这里简化处理，实际应该调用带搜索的接口
    // 目前 listProducts 不支持搜索，暂时拉取第一页
    const res = await api.listProducts({ page: 1, page_size: 20 })
    productOptions.value = res.data.list.filter(p => 
      p.sku.toLowerCase().includes(query.toLowerCase()) || 
      p.name.toLowerCase().includes(query.toLowerCase())
    )
  } catch (error) {
    console.error('搜索产品失败', error)
  } finally {
    productLoading.value = false
  }
}

// 打开映射对话框
const handleMap = (row: PlatformProduct) => {
  currentPlatformProduct.value = row
  selectedProductId.value = row.product_mapping?.product_id
  mapDialogVisible.value = true
  // 预加载已关联的产品
  if (row.product_mapping?.product) {
    productOptions.value = [row.product_mapping.product]
  } else {
    productOptions.value = []
    // 自动搜索同名或同SKU产品
    searchProducts(row.platform_sku)
  }
}

// 保存映射
const handleSaveMap = async () => {
  if (!currentPlatformProduct.value || !selectedProductId.value) {
    ElMessage.warning('请选择本地产品')
    return
  }

  mapLoading.value = true
  try {
    await api.mapProduct(currentPlatformProduct.value.id, selectedProductId.value)
    ElMessage.success('关联成功')
    mapDialogVisible.value = false
    fetchProducts()
  } catch (error) {
    console.error('关联失败', error)
  } finally {
    mapLoading.value = false
  }
}

// 解除关联
const handleUnmap = async (row: PlatformProduct) => {
  if (!row.product_mapping) return
  
  try {
    await ElMessageBox.confirm(
      '确定要解除关联吗？',
      '确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await api.unmapProduct(row.product_mapping.platform_product_id)
    ElMessage.success('解除关联成功')
    fetchProducts()
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('解除关联失败', error)
    }
  }
}

onMounted(() => {
  fetchAuths()
})
</script>

<template>
  <div class="platform-products">
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">平台产品管理</h1>
        <p class="page-desc">管理各平台的listing并关联本地产品</p>
      </div>
    </div>

    <div class="filter-card">
      <el-form inline>
        <el-form-item label="店铺">
          <el-select v-model="authId" placeholder="选择店铺" style="width: 200px" @change="handleSearch">
            <el-option
              v-for="item in authOptions"
              :key="item.id"
              :label="`${item.platform} - ${item.shop_name}`"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="keyword" placeholder="SKU/名称" clearable style="width: 180px" @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
          <el-button :icon="Refresh" @click="handleRefresh">刷新</el-button>
          <el-button type="success" :icon="Refresh" :loading="syncLoading" @click="handleSync" :disabled="!authId">
            同步产品
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
        <el-table-column prop="platform_sku" label="平台SKU" width="220" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="sku-cell">{{ row.platform_sku }}</span>
          </template>
        </el-table-column>
        <el-table-column label="平台标题" min-width="280">
          <template #default="{ row }">
            <div class="product-info-cell">
              <el-image
                v-if="row.image"
                :src="row.image"
                fit="cover"
                class="product-image"
                @click="showImagePreview(row.image)"
              />
              <div v-else class="product-image-placeholder">
                <el-icon><Picture /></el-icon>
              </div>
              <span class="product-name" :title="row.name">{{ row.name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="price" label="售价" width="120">
          <template #default="{ row }">
            {{ row.price }} {{ row.currency }}
          </template>
        </el-table-column>
        <el-table-column prop="stock" label="库存" width="100" />
        <el-table-column prop="status" label="状态" width="120" />
        <el-table-column label="关联本地产品" min-width="250">
          <template #default="{ row }">
            <div v-if="row.product_mapping" class="local-product-info">
              <el-tag type="success" size="small">{{ row.product_mapping.product?.sku }}</el-tag>
              <span class="product-name">{{ row.product_mapping.product?.name }}</span>
            </div>
            <span v-else class="text-secondary">未关联</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" :icon="Link" @click="handleMap(row)">
              {{ row.product_mapping ? '修改关联' : '关联' }}
            </el-button>
            <el-button 
              v-if="row.product_mapping" 
              type="danger" 
              link 
              size="small" 
              :icon="Delete" 
              @click="handleUnmap(row)"
            >
              解除
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

    <!-- 关联对话框 -->
    <el-dialog
      v-model="mapDialogVisible"
      title="关联本地产品"
      width="500px"
      destroy-on-close
    >
      <div style="margin-bottom: 16px">
        <p><strong>平台产品：</strong>{{ currentPlatformProduct?.name }}</p>
        <p><strong>平台SKU：</strong>{{ currentPlatformProduct?.platform_sku }}</p>
      </div>
      
      <el-form label-width="80px">
        <el-form-item label="本地产品">
          <el-select
            v-model="selectedProductId"
            filterable
            remote
            reserve-keyword
            placeholder="请输入SKU或名称搜索"
            :remote-method="searchProducts"
            :loading="productLoading"
            style="width: 100%"
          >
            <el-option
              v-for="item in productOptions"
              :key="item.id"
              :label="`${item.sku} - ${item.name}`"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="mapDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="mapLoading" @click="handleSaveMap">
          保存关联
        </el-button>
      </template>
    </el-dialog>

    <!-- 图片预览 -->
    <ImagePreview ref="imagePreviewRef" />
  </div>
</template>

<style scoped lang="scss">
.platform-products {
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

.text-secondary {
  color: var(--text-secondary);
  font-size: 12px;
}

.local-product-info {
  display: flex;
  align-items: center;
  gap: 8px;

  .product-name {
    color: var(--text-secondary);
    font-size: 12px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.sku-cell {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  display: block;
}

.product-info-cell {
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

.product-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  line-height: 1.4;
}
</style>
