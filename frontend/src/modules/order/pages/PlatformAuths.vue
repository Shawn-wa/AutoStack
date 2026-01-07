<script setup lang="ts">
import { ref, computed, onMounted, nextTick } from 'vue'

defineOptions({ name: 'PlatformAuths' })
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Connection, Coin } from '@element-plus/icons-vue'
import {
  getPlatforms,
  getAuths,
  createAuth,
  updateAuth,
  deleteAuth,
  testAuth,
  syncOrders,
  syncCommission,
  type PlatformInfo,
  type PlatformAuth,
  type CreateAuthParams
} from '@/modules/order/api'
import { formatDateTime } from '@/utils/format'

const loading = ref(false)
const tableData = ref<PlatformAuth[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)

// 平台列表
const platforms = ref<PlatformInfo[]>([])

// 创建对话框
const createDialogVisible = ref(false)
const createFormId = ref('')  // 用于生成唯一的 name 属性，防止浏览器自动填充
const createFormReadonly = ref(true)  // 初始为只读，防止浏览器自动填充
const createForm = ref<CreateAuthParams>({
  platform: '',
  shop_name: '',
  credentials: {}
})
const createLoading = ref(false)

// 编辑对话框
const editDialogVisible = ref(false)
const editForm = ref<{
  id: number
  shop_name: string
  credentials: Record<string, string>
  status: number
}>({
  id: 0,
  shop_name: '',
  credentials: {},
  status: 1
})
const editLoading = ref(false)
const editMaskedCredentials = ref<Record<string, string>>({})

// 同步对话框
const syncDialogVisible = ref(false)
const syncAuthId = ref(0)
const syncLoading = ref(false)
const syncDays = ref(7)

// 佣金同步对话框
const commissionDialogVisible = ref(false)
const commissionAuthId = ref(0)
const commissionLoading = ref(false)
const commissionDays = ref(30)

// 获取当前选中平台的凭证字段
const currentPlatformFields = computed(() => {
  const platform = platforms.value.find(p => p.name === createForm.value.platform)
  return platform?.fields || []
})

// 获取编辑时的平台信息
const editPlatformFields = computed(() => {
  const auth = tableData.value.find(a => a.id === editForm.value.id)
  if (!auth) return []
  const platform = platforms.value.find(p => p.name === auth.platform)
  return platform?.fields || []
})

// 获取平台列表
const fetchPlatforms = async () => {
  try {
    const res = await getPlatforms()
    platforms.value = res.data
  } catch (error) {
    console.error('获取平台列表失败', error)
  }
}

// 获取授权列表
const fetchAuths = async () => {
  loading.value = true
  try {
    const res = await getAuths(currentPage.value, pageSize.value)
    tableData.value = res.data.list
    total.value = res.data.total
  } catch (error) {
    console.error('获取授权列表失败', error)
  } finally {
    loading.value = false
  }
}

// 分页变化
const handlePageChange = (page: number) => {
  currentPage.value = page
  fetchAuths()
}

// 每页数量变化
const handleSizeChange = (size: number) => {
  pageSize.value = size
  currentPage.value = 1
  fetchAuths()
}

// 重置创建表单
const resetCreateForm = () => {
  createForm.value = {
    platform: platforms.value[0]?.name || '',
    shop_name: '',
    credentials: {}
  }
  // 初始化凭证字段
  if (platforms.value[0]) {
    platforms.value[0].fields.forEach(field => {
      createForm.value.credentials[field.key] = ''
    })
  }
}

// 生成唯一 ID
const generateUniqueId = () => {
  return `${Date.now()}_${Math.random().toString(36).substring(2, 11)}`
}

// 打开创建对话框
const handleCreate = () => {
  createFormId.value = generateUniqueId()
  createFormReadonly.value = true  // 初始只读，防止自动填充
  resetCreateForm()
  createDialogVisible.value = true
}

// 创建对话框完全打开后的回调（动画完成后，浏览器自动填充已执行）
const onCreateDialogOpened = () => {
  // 此时浏览器自动填充已经完成，直接清空表单
  resetCreateForm()
  createFormReadonly.value = false
}

// 平台变化时更新凭证字段
const onPlatformChange = () => {
  createForm.value.credentials = {}
  currentPlatformFields.value.forEach(field => {
    createForm.value.credentials[field.key] = ''
  })
}

// 保存创建
const handleSaveCreate = async () => {
  if (!createForm.value.shop_name) {
    ElMessage.warning('请填写店铺名称')
    return
  }

  // 检查必填字段
  for (const field of currentPlatformFields.value) {
    if (field.required && !createForm.value.credentials[field.key]) {
      ElMessage.warning(`请填写 ${field.label}`)
      return
    }
  }

  createLoading.value = true
  try {
    await createAuth(createForm.value)
    ElMessage.success('创建成功')
    createDialogVisible.value = false
    fetchAuths()
  } catch (error: any) {
    console.error('创建授权失败', error)
  } finally {
    createLoading.value = false
  }
}

// 编辑授权
const handleEdit = (row: PlatformAuth) => {
  editForm.value = {
    id: row.id,
    shop_name: row.shop_name,
    credentials: {},
    status: row.status
  }
  // 保存脱敏凭证用于 placeholder 显示
  editMaskedCredentials.value = row.masked_credentials || {}
  // 初始化凭证字段为空（只有用户填写时才更新）
  const platform = platforms.value.find(p => p.name === row.platform)
  if (platform) {
    platform.fields.forEach(field => {
      editForm.value.credentials[field.key] = ''
    })
  }
  editDialogVisible.value = true
}

// 保存编辑
const handleSaveEdit = async () => {
  editLoading.value = true
  try {
    const updateData: any = {
      shop_name: editForm.value.shop_name,
      status: editForm.value.status
    }
    // 只有填写了新凭证才更新
    const hasCredentials = Object.values(editForm.value.credentials).some(v => v)
    if (hasCredentials) {
      updateData.credentials = editForm.value.credentials
    }
    await updateAuth(editForm.value.id, updateData)
    ElMessage.success('更新成功')
    editDialogVisible.value = false
    fetchAuths()
  } catch (error) {
    console.error('更新授权失败', error)
  } finally {
    editLoading.value = false
  }
}

// 删除授权
const handleDelete = async (row: PlatformAuth) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除店铺 "${row.shop_name}" 的授权吗？删除后将无法同步该店铺的订单。`,
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await deleteAuth(row.id)
    ElMessage.success('删除成功')
    fetchAuths()
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('删除授权失败', error)
    }
  }
}

// 测试连接
const handleTest = async (row: PlatformAuth) => {
  try {
    await testAuth(row.id)
    ElMessage.success('连接成功')
  } catch (error: any) {
    console.error('测试连接失败', error)
  }
}

// 打开同步对话框
const handleOpenSync = (row: PlatformAuth) => {
  syncAuthId.value = row.id
  syncDays.value = 7
  syncDialogVisible.value = true
}

// 同步订单
const handleSync = async () => {
  syncLoading.value = true
  try {
    const now = new Date()
    let since: Date
    
    if (syncDays.value === -1) {
      // 历史订单：从2025-01-01开始
      since = new Date('2025-01-01T00:00:00Z')
    } else {
      since = new Date(now.getTime() - syncDays.value * 24 * 60 * 60 * 1000)
    }
    
    const res = await syncOrders(syncAuthId.value, {
      since: since.toISOString(),
      to: now.toISOString()
    })
    ElMessage.success(`同步完成：共 ${res.data.total} 条，新增 ${res.data.created} 条，更新 ${res.data.updated} 条`)
    syncDialogVisible.value = false
    fetchAuths()
  } catch (error: any) {
    console.error('同步订单失败', error)
  } finally {
    syncLoading.value = false
  }
}

// 打开佣金同步对话框
const handleOpenCommissionSync = (row: PlatformAuth) => {
  commissionAuthId.value = row.id
  commissionDays.value = 30
  commissionDialogVisible.value = true
}

// 同步佣金
const handleSyncCommission = async () => {
  commissionLoading.value = true
  try {
    const now = new Date()
    const since = new Date(now.getTime() - commissionDays.value * 24 * 60 * 60 * 1000)
    const res = await syncCommission(commissionAuthId.value, {
      since: since.toISOString(),
      to: now.toISOString()
    })
    ElMessage.success(`佣金同步完成：共处理 ${res.data.total} 条交易，更新 ${res.data.updated} 个订单`)
    commissionDialogVisible.value = false
  } catch (error: any) {
    console.error('同步佣金失败', error)
  } finally {
    commissionLoading.value = false
  }
}

// 获取平台显示名称
const getPlatformLabel = (name: string) => {
  const platform = platforms.value.find(p => p.name === name)
  return platform?.label || name
}

// 获取状态标签类型
const getStatusTagType = (status: number) => {
  switch (status) {
    case 1: return 'success'
    case 0: return 'info'
    case 2: return 'danger'
    default: return 'info'
  }
}

// 获取状态显示文字
const getStatusText = (status: number) => {
  switch (status) {
    case 1: return '正常'
    case 0: return '禁用'
    case 2: return '授权失效'
    default: return '未知'
  }
}

onMounted(() => {
  fetchPlatforms()
  fetchAuths()
})
</script>

<template>
  <div class="platform-auths">
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">平台授权</h1>
        <p class="page-desc">管理电商平台的API授权，用于同步订单数据</p>
      </div>
      <div class="header-right">
        <el-button type="primary" :icon="Plus" @click="handleCreate">
          添加授权
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
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="platform" label="平台" width="120">
          <template #default="{ row }">
            <el-tag type="primary" size="small">
              {{ getPlatformLabel(row.platform) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="shop_name" label="店铺名称" min-width="150" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusTagType(row.status)" size="small">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="last_sync_at" label="最后同步" width="180">
          <template #default="{ row }">
            {{ row.last_sync_at ? formatDateTime(row.last_sync_at) : '从未同步' }}
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="350" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" :icon="Connection" @click="handleTest(row)">
              测试
            </el-button>
            <el-button type="success" link size="small" :icon="Refresh" @click="handleOpenSync(row)">
              同步订单
            </el-button>
            <el-button type="warning" link size="small" :icon="Coin" @click="handleOpenCommissionSync(row)">
              同步佣金
            </el-button>
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

    <!-- 创建对话框 -->
    <el-dialog
      v-model="createDialogVisible"
      title="添加平台授权"
      width="500px"
      :close-on-click-modal="false"
      destroy-on-close
      @opened="onCreateDialogOpened"
    >
      <el-form :model="createForm" label-width="100px" autocomplete="off">
        <!-- 隐藏的诱饵输入框，用于欺骗浏览器自动填充 -->
        <input type="text" style="display:none" autocomplete="off" />
        <input type="password" style="display:none" autocomplete="new-password" />
        <el-form-item label="选择平台" required>
          <el-select v-model="createForm.platform" placeholder="请选择平台" style="width: 100%" @change="onPlatformChange">
            <el-option
              v-for="platform in platforms"
              :key="platform.name"
              :label="platform.label"
              :value="platform.name"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="店铺名称" required>
          <el-input 
            v-model="createForm.shop_name" 
            placeholder="请输入店铺名称（自定义）" 
            autocomplete="off"
            :name="`shop_name_${createFormId}`"
            :readonly="createFormReadonly"
            @focus="createFormReadonly = false"
          />
        </el-form-item>
        <el-divider>API 凭证</el-divider>
        <el-form-item 
          v-for="(field, index) in currentPlatformFields" 
          :key="field.key" 
          :label="field.label"
          :required="field.required"
        >
          <el-input
            v-model="createForm.credentials[field.key]"
            :type="field.type === 'password' ? 'password' : 'text'"
            :placeholder="`请输入 ${field.label}`"
            :show-password="field.type === 'password'"
            :autocomplete="field.type === 'password' ? 'new-password' : 'off'"
            :name="`${field.key}_${index}_${createFormId}`"
            :readonly="createFormReadonly"
            @focus="createFormReadonly = false"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="createLoading" @click="handleSaveCreate">
          创建
        </el-button>
      </template>
    </el-dialog>

    <!-- 编辑对话框 -->
    <el-dialog
      v-model="editDialogVisible"
      title="编辑授权"
      width="500px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form :model="editForm" label-width="100px" autocomplete="off">
        <el-form-item label="店铺名称">
          <el-input v-model="editForm.shop_name" placeholder="请输入店铺名称" autocomplete="off" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="editForm.status" placeholder="请选择状态" style="width: 100%">
            <el-option label="正常" :value="1" />
            <el-option label="禁用" :value="0" />
          </el-select>
        </el-form-item>
        <el-divider>更新 API 凭证（留空则不更新）</el-divider>
        <el-form-item 
          v-for="field in editPlatformFields" 
          :key="field.key" 
          :label="field.label"
        >
          <el-input
            v-model="editForm.credentials[field.key]"
            :type="field.type === 'password' ? 'password' : 'text'"
            :placeholder="editMaskedCredentials[field.key] || '留空则不更新'"
            :show-password="field.type === 'password'"
            :autocomplete="field.type === 'password' ? 'new-password' : 'off'"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="editLoading" @click="handleSaveEdit">
          保存
        </el-button>
      </template>
    </el-dialog>

    <!-- 同步对话框 -->
    <el-dialog
      v-model="syncDialogVisible"
      title="同步订单"
      width="400px"
      :close-on-click-modal="false"
    >
      <el-form label-width="100px">
        <el-form-item label="同步范围">
          <el-select v-model="syncDays" style="width: 100%">
            <el-option label="最近 1 天" :value="1" />
            <el-option label="最近 3 天" :value="3" />
            <el-option label="最近 7 天" :value="7" />
            <el-option label="最近 15 天" :value="15" />
            <el-option label="最近 30 天" :value="30" />
            <el-option label="历史订单（从2025-01-01）" :value="-1" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="syncDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="syncLoading" @click="handleSync">
          开始同步
        </el-button>
      </template>
    </el-dialog>

    <!-- 佣金同步对话框 -->
    <el-dialog
      v-model="commissionDialogVisible"
      title="同步佣金"
      width="400px"
      :close-on-click-modal="false"
    >
      <el-alert
        title="佣金同步说明"
        type="info"
        :closable="false"
        show-icon
        style="margin-bottom: 16px"
      >
        将从Ozon平台获取财务交易数据，更新订单的佣金信息。
      </el-alert>
      <el-form label-width="100px">
        <el-form-item label="同步范围">
          <el-select v-model="commissionDays" style="width: 100%">
            <el-option label="最近 7 天" :value="7" />
            <el-option label="最近 15 天" :value="15" />
            <el-option label="最近 30 天" :value="30" />
            <el-option label="最近 60 天" :value="60" />
            <el-option label="最近 90 天" :value="90" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="commissionDialogVisible = false">取消</el-button>
        <el-button type="warning" :loading="commissionLoading" @click="handleSyncCommission">
          开始同步佣金
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped lang="scss">
.platform-auths {
  padding: 0;
}

.page-header {
  margin-bottom: 24px;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.header-left {
  flex: 1;
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
