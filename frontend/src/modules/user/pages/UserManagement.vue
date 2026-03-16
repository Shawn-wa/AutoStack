<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'

defineOptions({ name: 'UserManagement' })

import { 
  getUsers, 
  createUser, 
  updateUser, 
  deleteUser, 
  type UserInfo, 
  type UpdateUserParams,
  type CreateUserParams
} from '@/modules/user/api'
import { useUserStore } from '@/modules/auth/stores'
import { formatDateTime } from '@/utils/format'

const userStore = useUserStore()

const loading = ref(false)
const tableData = ref<UserInfo[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)

// 筛选条件
const filterKeyword = ref('')
const filterRole = ref('')

// 创建对话框
const createDialogVisible = ref(false)
const createFormId = ref('')  // 用于生成唯一的 name 属性，防止浏览器自动填充
const createFormReadonly = ref(true)  // 初始为只读，防止浏览器自动填充
const createForm = ref<CreateUserParams>({
  username: '',
  password: '',
  email: '',
  role: 'user'
})
const createLoading = ref(false)

// 编辑对话框
const editDialogVisible = ref(false)
const editForm = ref<UpdateUserParams & { id: number; username: string }>({
  id: 0,
  username: '',
  email: '',
  role: '',
  status: 1
})
const editLoading = ref(false)

// 角色选项（根据当前用户动态计算）
const roleOptions = computed(() => {
  const options = [{ label: '普通用户', value: 'user' }]
  // 只有超级管理员可以创建/编辑管理员
  if (userStore.isSuperAdmin) {
    options.unshift({ label: '管理员', value: 'admin' })
  }
  return options
})

// 状态选项
const statusOptions = [
  { label: '正常', value: 1 },
  { label: '禁用', value: 0 }
]

// 获取用户列表
const fetchUsers = async () => {
  loading.value = true
  try {
    const res = await getUsers({
      page: currentPage.value,
      page_size: pageSize.value,
      keyword: filterKeyword.value || undefined,
      role: filterRole.value || undefined
    })
    tableData.value = res.data.list
    total.value = res.data.total
  } catch (error) {
    console.error('获取用户列表失败', error)
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  currentPage.value = 1
  fetchUsers()
}

// 重置筛选
const handleReset = () => {
  filterKeyword.value = ''
  filterRole.value = ''
  currentPage.value = 1
  fetchUsers()
}

// 分页变化
const handlePageChange = (page: number) => {
  currentPage.value = page
  fetchUsers()
}

// 每页数量变化
const handleSizeChange = (size: number) => {
  pageSize.value = size
  currentPage.value = 1
  fetchUsers()
}

// 打开创建对话框
// 重置创建表单
const resetCreateForm = () => {
  createForm.value = {
    username: '',
    password: '',
    email: '',
    role: 'user'
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

// 保存创建
const handleSaveCreate = async () => {
  if (!createForm.value.username || !createForm.value.password || !createForm.value.email) {
    ElMessage.warning('请填写完整信息')
    return
  }

  createLoading.value = true
  try {
    await createUser(createForm.value)
    ElMessage.success('创建成功')
    createDialogVisible.value = false
    fetchUsers()
  } catch (error: any) {
    console.error('创建用户失败', error)
  } finally {
    createLoading.value = false
  }
}

// 编辑用户
const handleEdit = (row: UserInfo) => {
  // 检查是否有权限编辑
  if (!canManageUser(row)) {
    ElMessage.warning('无权管理该用户')
    return
  }

  editForm.value = {
    id: row.id,
    username: row.username,
    email: row.email,
    role: row.role,
    status: row.status
  }
  editDialogVisible.value = true
}

// 保存编辑
const handleSaveEdit = async () => {
  editLoading.value = true
  try {
    await updateUser(editForm.value.id, {
      email: editForm.value.email,
      role: editForm.value.role,
      status: editForm.value.status
    })
    ElMessage.success('更新成功')
    editDialogVisible.value = false
    fetchUsers()
  } catch (error) {
    console.error('更新用户失败', error)
  } finally {
    editLoading.value = false
  }
}

// 删除用户
const handleDelete = async (row: UserInfo) => {
  if (row.id === userStore.user?.id) {
    ElMessage.warning('不能删除自己')
    return
  }

  if (row.role === 'super_admin') {
    ElMessage.warning('不能删除超级管理员')
    return
  }

  if (!canManageUser(row)) {
    ElMessage.warning('无权删除该用户')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要删除用户 "${row.username}" 吗？此操作不可恢复。`,
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await deleteUser(row.id)
    ElMessage.success('删除成功')
    fetchUsers()
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('删除用户失败', error)
    }
  }
}

// 检查是否可以管理该用户
const canManageUser = (row: UserInfo): boolean => {
  // 不能管理自己
  if (row.id === userStore.user?.id) return false
  // 不能管理超级管理员
  if (row.role === 'super_admin') return false
  // 超级管理员可以管理所有人
  if (userStore.isSuperAdmin) return true
  // 管理员只能管理普通用户
  if (userStore.isAdmin && row.role === 'user') return true
  return false
}

// 获取角色标签类型
const getRoleTagType = (role: string) => {
  switch (role) {
    case 'super_admin': return 'danger'
    case 'admin': return 'warning'
    default: return 'info'
  }
}

// 获取角色显示名称
const getRoleName = (role: string) => {
  switch (role) {
    case 'super_admin': return '超级管理员'
    case 'admin': return '管理员'
    default: return '普通用户'
  }
}

// 获取状态标签类型
const getStatusTagType = (status: number) => {
  return status === 1 ? 'success' : 'danger'
}

onMounted(() => {
  fetchUsers()
})
</script>

<template>
  <div class="user-management">
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">用户管理</h1>
        <p class="page-desc">管理系统中的所有用户</p>
      </div>
    </div>

    <div class="filter-card">
      <el-form inline>
        <el-form-item label="关键词">
          <el-input v-model="filterKeyword" placeholder="用户名/邮箱" clearable style="width: 180px" @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="filterRole" placeholder="全部角色" clearable style="width: 140px">
            <el-option label="超级管理员" value="super_admin" />
            <el-option label="管理员" value="admin" />
            <el-option label="普通用户" value="user" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
          <el-button type="primary" :icon="Plus" @click="handleCreate">创建用户</el-button>
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
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="username" label="用户名" width="150" />
        <el-table-column prop="email" label="邮箱" min-width="200" />
        <el-table-column prop="role" label="角色" width="120">
          <template #default="{ row }">
            <el-tag :type="getRoleTagType(row.role)" size="small">
              {{ getRoleName(row.role) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusTagType(row.status)" size="small">
              {{ row.status === 1 ? '正常' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="permissions" label="权限数" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.role === 'super_admin'" type="danger" size="small">全部</el-tag>
            <span v-else>{{ row.permissions?.length || 0 }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button 
              type="primary" 
              link 
              size="small" 
              @click="handleEdit(row)"
              :disabled="!canManageUser(row)"
            >
              编辑
            </el-button>
            <el-button 
              type="danger" 
              link 
              size="small" 
              @click="handleDelete(row)"
              :disabled="!canManageUser(row)"
            >
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
      title="创建用户"
      width="600px"
      :close-on-click-modal="false"
      destroy-on-close
      @opened="onCreateDialogOpened"
    >
      <el-form :model="createForm" label-width="80px" autocomplete="off">
        <!-- 隐藏的诱饵输入框，用于欺骗浏览器自动填充 -->
        <input type="text" style="display:none" autocomplete="off" />
        <input type="password" style="display:none" autocomplete="new-password" />
        <el-form-item label="用户名" required>
          <el-input 
            v-model="createForm.username" 
            placeholder="请输入用户名（3-20位）" 
            autocomplete="off"
            :name="`create_username_${createFormId}`"
            :readonly="createFormReadonly"
            @focus="createFormReadonly = false"
          />
        </el-form-item>
        <el-form-item label="密码" required>
          <el-input 
            v-model="createForm.password" 
            type="password" 
            placeholder="请输入密码（至少6位）" 
            show-password 
            autocomplete="new-password"
            :name="`create_password_${createFormId}`"
            :readonly="createFormReadonly"
            @focus="createFormReadonly = false"
          />
        </el-form-item>
        <el-form-item label="邮箱" required>
          <el-input v-model="createForm.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="createForm.role" placeholder="请选择角色" style="width: 100%">
            <el-option
              v-for="item in roleOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
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
      title="编辑用户"
      width="600px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form :model="editForm" label-width="80px" autocomplete="off">
        <el-form-item label="用户名">
          <el-input v-model="editForm.username" disabled />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="editForm.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="editForm.role" placeholder="请选择角色" style="width: 100%">
            <el-option
              v-for="item in roleOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="editForm.status" placeholder="请选择状态" style="width: 100%">
            <el-option
              v-for="item in statusOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="editLoading" @click="handleSaveEdit">
          保存
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped lang="scss">
.user-management {
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

</style>
