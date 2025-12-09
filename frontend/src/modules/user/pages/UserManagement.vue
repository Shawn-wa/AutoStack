<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { 
  getUsers, 
  getPermissions,
  createUser, 
  updateUser, 
  deleteUser, 
  type UserInfo, 
  type UpdateUserParams,
  type CreateUserParams,
  type PermissionDef,
  type PermissionsResult
} from '@/modules/user/api'
import { useUserStore } from '@/modules/auth/stores'
import { formatDateTime } from '@/utils/format'

const userStore = useUserStore()

const loading = ref(false)
const tableData = ref<UserInfo[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)

// 权限数据
const permissionsData = ref<PermissionsResult | null>(null)

// 创建对话框
const createDialogVisible = ref(false)
const createForm = ref<CreateUserParams>({
  username: '',
  password: '',
  email: '',
  role: 'user',
  permissions: []
})
const createLoading = ref(false)

// 编辑对话框
const editDialogVisible = ref(false)
const editForm = ref<UpdateUserParams & { id: number; username: string }>({
  id: 0,
  username: '',
  email: '',
  role: '',
  status: 1,
  permissions: []
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

// 模块名称映射
const moduleNames: Record<string, string> = {
  user: '用户管理',
  project: '项目管理',
  deployment: '部署管理',
  template: '模板管理'
}

// 获取权限列表
const fetchPermissions = async () => {
  try {
    const res = await getPermissions()
    permissionsData.value = res.data
  } catch (error) {
    console.error('获取权限列表失败', error)
  }
}

// 获取可分配的权限（根据目标角色过滤）
const getAssignablePermissions = (targetRole: string): Record<string, PermissionDef[]> => {
  if (!permissionsData.value) return {}
  
  const result: Record<string, PermissionDef[]> = {}
  
  for (const [module, perms] of Object.entries(permissionsData.value.modules)) {
    // 用户管理权限只有超级管理员可以分配给管理员
    if (module === 'user') {
      if (userStore.isSuperAdmin && targetRole === 'admin') {
        result[module] = perms
      }
      continue
    }
    // 其他权限都可以分配
    result[module] = perms
  }
  
  return result
}

// 获取用户列表
const fetchUsers = async () => {
  loading.value = true
  try {
    const res = await getUsers(currentPage.value, pageSize.value)
    tableData.value = res.data.list
    total.value = res.data.total
  } catch (error) {
    console.error('获取用户列表失败', error)
  } finally {
    loading.value = false
  }
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
const handleCreate = () => {
  createForm.value = {
    username: '',
    password: '',
    email: '',
    role: 'user',
    permissions: []
  }
  createDialogVisible.value = true
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
    status: row.status,
    permissions: row.permissions || []
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
      status: editForm.value.status,
      permissions: editForm.value.permissions
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

// 当创建表单角色变化时，清空不适用的权限
const onCreateRoleChange = () => {
  createForm.value.permissions = []
}

// 当编辑表单角色变化时，清空不适用的权限
const onEditRoleChange = () => {
  editForm.value.permissions = []
}

onMounted(() => {
  fetchUsers()
  fetchPermissions()
})
</script>

<template>
  <div class="user-management">
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">用户管理</h1>
        <p class="page-desc">管理系统中的所有用户</p>
      </div>
      <div class="header-right">
        <el-button type="primary" :icon="Plus" @click="handleCreate">
          创建用户
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
    >
      <el-form :model="createForm" label-width="80px">
        <el-form-item label="用户名" required>
          <el-input v-model="createForm.username" placeholder="请输入用户名（3-20位）" />
        </el-form-item>
        <el-form-item label="密码" required>
          <el-input v-model="createForm.password" type="password" placeholder="请输入密码（至少6位）" show-password />
        </el-form-item>
        <el-form-item label="邮箱" required>
          <el-input v-model="createForm.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="createForm.role" placeholder="请选择角色" style="width: 100%" @change="onCreateRoleChange">
            <el-option
              v-for="item in roleOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="权限">
          <div class="permissions-container">
            <template v-for="(perms, module) in getAssignablePermissions(createForm.role)" :key="module">
              <div class="permission-module">
                <div class="module-title">{{ moduleNames[module] || module }}</div>
                <el-checkbox-group v-model="createForm.permissions">
                  <el-checkbox
                    v-for="perm in perms"
                    :key="perm.code"
                    :label="perm.code"
                  >
                    {{ perm.name }}
                  </el-checkbox>
                </el-checkbox-group>
              </div>
            </template>
            <el-empty v-if="Object.keys(getAssignablePermissions(createForm.role)).length === 0" description="无可分配权限" :image-size="60" />
          </div>
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
    >
      <el-form :model="editForm" label-width="80px">
        <el-form-item label="用户名">
          <el-input v-model="editForm.username" disabled />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="editForm.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="editForm.role" placeholder="请选择角色" style="width: 100%" @change="onEditRoleChange">
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
        <el-form-item label="权限">
          <div class="permissions-container">
            <template v-for="(perms, module) in getAssignablePermissions(editForm.role || '')" :key="module">
              <div class="permission-module">
                <div class="module-title">{{ moduleNames[module] || module }}</div>
                <el-checkbox-group v-model="editForm.permissions">
                  <el-checkbox
                    v-for="perm in perms"
                    :key="perm.code"
                    :label="perm.code"
                  >
                    {{ perm.name }}
                  </el-checkbox>
                </el-checkbox-group>
              </div>
            </template>
            <el-empty v-if="Object.keys(getAssignablePermissions(editForm.role || '')).length === 0" description="无可分配权限" :image-size="60" />
          </div>
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

.permissions-container {
  width: 100%;
  max-height: 300px;
  overflow-y: auto;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 12px;
}

.permission-module {
  margin-bottom: 16px;

  &:last-child {
    margin-bottom: 0;
  }
}

.module-title {
  font-weight: 600;
  font-size: 14px;
  margin-bottom: 8px;
  color: var(--text-primary);
  padding-bottom: 8px;
  border-bottom: 1px solid var(--border-color);
}

:deep(.el-checkbox-group) {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 16px;
}

:deep(.el-checkbox) {
  margin-right: 0;
}
</style>
