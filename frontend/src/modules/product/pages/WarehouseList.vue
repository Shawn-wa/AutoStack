<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import api, { type WarehouseResponse } from '../api'

defineOptions({ name: 'WarehouseList' })

const loading = ref(false)
const tableData = ref<WarehouseResponse[]>([])
const activeTab = ref('all')

// 仓库类型TAB选项
const tabOptions = [
  { name: 'all', label: '全部' },
  { name: 'local', label: '本地仓' },
  { name: 'overseas', label: '海外仓' },
  { name: 'fba', label: 'FBA仓' },
  { name: 'third', label: '第三方仓' },
  { name: 'virtual', label: '虚拟仓' },
]

// 仓库类型选项（用于新建仓库）
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

// 获取仓库类型显示文本
const getWarehouseTypeLabel = (type: string) => {
  const item = warehouseTypeOptions.find(o => o.value === type)
  return item?.label || type
}

// 获取状态标签类型
const getStatusType = (status: string) => {
  return status === 'active' ? 'success' : 'danger'
}

// 获取状态显示文本
const getStatusLabel = (status: string) => {
  return status === 'active' ? '启用' : '停用'
}

// 获取仓库列表
const fetchList = async () => {
  loading.value = true
  try {
    const res = await api.listAllWarehouses(activeTab.value)
    tableData.value = res.data || []
  } catch (error) {
    console.error('获取仓库列表失败', error)
  } finally {
    loading.value = false
  }
}

// TAB切换
const handleTabChange = () => {
  fetchList()
}

// 打开新建仓库弹窗
const openWarehouseDialog = () => {
  // 根据当前TAB预设仓库类型
  const presetType = activeTab.value !== 'all' ? activeTab.value : 'local'
  warehouseForm.value = {
    code: '',
    name: '',
    type: presetType,
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
    fetchList()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '创建失败')
  } finally {
    warehouseLoading.value = false
  }
}

// 统计各类型仓库数量
const allWarehouses = ref<WarehouseResponse[]>([])
const tabCounts = computed(() => {
  const counts: Record<string, number> = { all: allWarehouses.value.length }
  warehouseTypeOptions.forEach(opt => {
    counts[opt.value] = allWarehouses.value.filter(w => w.type === opt.value).length
  })
  return counts
})

// 初始加载所有仓库用于统计
const fetchAllForCounts = async () => {
  try {
    const res = await api.listAllWarehouses()
    allWarehouses.value = res.data || []
  } catch (error) {
    console.error('获取仓库统计失败', error)
  }
}

onMounted(() => {
  fetchAllForCounts()
  fetchList()
})
</script>

<template>
  <div class="warehouse-list">
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">仓库列表</h1>
        <p class="page-desc">管理系统仓库信息</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="openWarehouseDialog">
          <el-icon><Plus /></el-icon>
          新建仓库
        </el-button>
      </div>
    </div>

    <div class="content-card">
      <!-- TAB切换 -->
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane
          v-for="tab in tabOptions"
          :key="tab.name"
          :name="tab.name"
        >
          <template #label>
            <span class="tab-label">
              {{ tab.label }}
              <el-badge
                v-if="tabCounts[tab.name] > 0"
                :value="tabCounts[tab.name]"
                class="tab-badge"
                type="primary"
              />
            </span>
          </template>
        </el-tab-pane>
      </el-tabs>

      <!-- 仓库表格 -->
      <el-table
        v-loading="loading"
        :data="tableData"
        style="width: 100%"
        stripe
      >
        <el-table-column prop="code" label="仓库编码" width="150">
          <template #default="{ row }">
            <span class="warehouse-code">{{ row.code }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="仓库名称" width="200" />
        <el-table-column prop="type" label="仓库类型" width="120">
          <template #default="{ row }">
            <el-tag size="small">{{ getWarehouseTypeLabel(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="address" label="地址" min-width="250">
          <template #default="{ row }">
            <span>{{ row.address || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)" size="small">
              {{ getStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180" />
      </el-table>
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
  </div>
</template>

<style scoped lang="scss">
.warehouse-list {
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

.content-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: 24px;
}

.warehouse-code {
  font-family: monospace;
  font-weight: 500;
  color: var(--color-primary);
}

.tab-label {
  display: flex;
  align-items: center;
  gap: 8px;
}

.tab-badge {
  :deep(.el-badge__content) {
    font-size: 11px;
    height: 16px;
    line-height: 16px;
    padding: 0 5px;
  }
}

:deep(.el-tabs__header) {
  margin-bottom: 20px;
}

:deep(.el-tabs__nav-wrap::after) {
  height: 1px;
}
</style>
