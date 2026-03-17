<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { getRolePermissions, getRoles, createRole, updateRole, deleteRole, type RolePermissionsResult, type RoleItem } from '@/modules/user/api'
import { useUserStore } from '@/modules/auth/stores'

defineOptions({ name: 'RoleManagement' })

const userStore = useUserStore()
const router = useRouter()

const rolePermissionsData = ref<RolePermissionsResult | null>(null)
const roleList = ref<RoleItem[]>([])
const loading = ref(false)
const toggleLoadingMap = ref<Record<number, boolean>>({})

const createDialogVisible = ref(false)
const editDialogVisible = ref(false)
const submitting = ref(false)

const createRoleForm = ref({
  role: '',
  role_name: '',
  description: '',
  enabled: true
})

const editRoleForm = ref({
  id: 0,
  role: '',
  role_name: '',
  description: '',
  enabled: true
})
type RoleRow = RoleItem & { permissionCount: number; canEdit: boolean }

const canCreateRole = computed(() => userStore.isAdmin)

const canEditRole = (role: string) => {
  if (userStore.isSuperAdmin) return role !== 'super_admin'
  if (userStore.isAdmin) return role !== 'super_admin' && role !== 'admin'
  return false
}

const normalizeRoleCode = (v: string) => v.trim().toLowerCase()
const normalizeRoleName = (v: string) => v.trim()

const validateRoleForm = (roleCode: string, roleName: string, excludeId?: number): boolean => {
  if (!roleCode || !roleName) {
    ElMessage.warning('请填写角色编码和角色名')
    return false
  }
  if (!/^[a-z0-9_]+$/.test(roleCode)) {
    ElMessage.warning('角色编码仅支持小写字母、数字、下划线')
    return false
  }
  const codeExists = roleList.value.some(r => normalizeRoleCode(r.role) === roleCode && r.id !== excludeId)
  if (codeExists) {
    ElMessage.warning('角色编码已存在')
    return false
  }
  const nameExists = roleList.value.some(r => normalizeRoleName(r.role_name) === roleName && r.id !== excludeId)
  if (nameExists) {
    ElMessage.warning('角色名已存在')
    return false
  }
  return true
}

const fetchData = async () => {
  loading.value = true
  try {
    const [rolePermRes, roleListRes] = await Promise.allSettled([getRolePermissions(), getRoles()])

    if (rolePermRes.status === 'fulfilled') {
      rolePermissionsData.value = rolePermRes.value.data
    } else {
      rolePermissionsData.value = null
      console.error('获取角色权限配置失败', rolePermRes.reason)
    }

    if (roleListRes.status === 'fulfilled') {
      roleList.value = roleListRes.value.data.list || []
    } else {
      roleList.value = []
      console.error('获取角色列表失败', roleListRes.reason)
    }
  } catch (error) {
    console.error('获取角色数据失败', error)
  } finally {
    loading.value = false
  }
}

const roleRows = computed<RoleRow[]>(() => {
  const rolePermissions = rolePermissionsData.value?.role_permissions || {}
  return roleList.value.map(item => ({
    ...item,
    permissionCount: rolePermissions[item.role]?.length ?? item.permission_count ?? 0,
    canEdit: canEditRole(item.role)
  }))
})

const handleEditPermissions = (role: string) => {
  router.push(`/users/roles/${role}`)
}

const openCreateRole = () => {
  createRoleForm.value = { role: '', role_name: '', description: '', enabled: true }
  createDialogVisible.value = true
}

const submitCreateRole = async () => {
  const role = normalizeRoleCode(createRoleForm.value.role)
  const roleName = normalizeRoleName(createRoleForm.value.role_name)
  if (!validateRoleForm(role, roleName)) return
  submitting.value = true
  try {
    await createRole({
      role,
      role_name: roleName,
      description: createRoleForm.value.description.trim(),
      enabled: createRoleForm.value.enabled ? 1 : 0
    })
    ElMessage.success('角色创建成功')
    createDialogVisible.value = false
    fetchData()
  } catch (error) {
    console.error('创建角色失败', error)
  } finally {
    submitting.value = false
  }
}

const openEditRole = (row: RoleRow) => {
  editRoleForm.value = {
    id: row.id,
    role: row.role,
    role_name: row.role_name,
    description: row.description || '',
    enabled: row.enabled
  }
  editDialogVisible.value = true
}

const submitEditRole = async () => {
  const role = normalizeRoleCode(editRoleForm.value.role)
  const roleName = normalizeRoleName(editRoleForm.value.role_name)
  if (!validateRoleForm(role, roleName, editRoleForm.value.id)) return
  submitting.value = true
  try {
    await updateRole(editRoleForm.value.id, {
      role,
      role_name: roleName,
      description: editRoleForm.value.description.trim(),
      enabled: editRoleForm.value.enabled ? 1 : 0
    })
    ElMessage.success('角色更新成功')
    editDialogVisible.value = false
    fetchData()
  } catch (error) {
    console.error('更新角色失败', error)
  } finally {
    submitting.value = false
  }
}

const handleDeleteRole = async (row: RoleRow) => {
  try {
    await ElMessageBox.confirm(`确认删除角色「${row.role_name}」？`, '删除确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await deleteRole(row.id)
    ElMessage.success('角色删除成功')
    fetchData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除角色失败', error)
    }
  }
}

const handleToggleEnabled = async (row: RoleRow, enabled: boolean) => {
  if (!row.canEdit) return
  const current = roleList.value.find(item => item.id === row.id)
  if (!current) return
  const previous = current.enabled
  current.enabled = enabled
  toggleLoadingMap.value[row.id] = true
  try {
    await updateRole(row.id, {
      role: row.role,
      role_name: row.role_name,
      description: row.description || '',
      enabled: enabled ? 1 : 0
    })
    ElMessage.success(`已${enabled ? '启用' : '停用'}角色`)
  } catch (error) {
    current.enabled = previous
    console.error('切换角色启用状态失败', error)
  } finally {
    toggleLoadingMap.value[row.id] = false
  }
}

onMounted(() => {
  fetchData()
})
</script>

<template>
  <div class="role-management">
    <div class="page-header">
      <div>
        <h1 class="page-title">角色管理</h1>
      </div>
      <el-button v-if="canCreateRole" type="primary" :icon="Plus" @click="openCreateRole">新增角色</el-button>
    </div>

    <div class="content-card">
      <el-table v-loading="loading" :data="roleRows" stripe>
        <el-table-column prop="role_name" label="角色名" min-width="160" />
        <el-table-column prop="role" label="角色编码" min-width="180" />
        <el-table-column prop="description" label="说明" min-width="220" />
        <el-table-column label="是否启用" width="120">
          <template #default="{ row }">
            <el-switch
              :model-value="row.enabled"
              :loading="!!toggleLoadingMap[row.id]"
              :disabled="!row.canEdit"
              inline-prompt
              active-text="启用"
              inactive-text="停用"
              @change="(val: string | number | boolean) => handleToggleEnabled(row, !!val)"
            />
          </template>
        </el-table-column>
        <el-table-column prop="permissionCount" label="权限数量" width="100" />
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="handleEditPermissions(row.role)" :disabled="!row.canEdit">
              编辑权限
            </el-button>
            <el-button type="primary" link @click="openEditRole(row)" :disabled="!row.canEdit">
              编辑角色
            </el-button>
            <el-button type="danger" link @click="handleDeleteRole(row)" :disabled="!row.canEdit">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="createDialogVisible" title="新增角色" width="560px" :close-on-click-modal="false" destroy-on-close>
      <el-form :model="createRoleForm" label-width="90px">
        <el-form-item label="角色名" required>
          <el-input v-model="createRoleForm.role_name" placeholder="例如：运营专员" maxlength="60" />
        </el-form-item>
        <el-form-item label="角色编码" required>
          <el-input v-model="createRoleForm.role" placeholder="例如：operator" maxlength="50" />
        </el-form-item>
        <el-form-item label="说明">
          <el-input v-model="createRoleForm.description" type="textarea" :rows="3" maxlength="255" show-word-limit />
        </el-form-item>
        <el-form-item label="是否启用">
          <el-switch v-model="createRoleForm.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitCreateRole">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="editDialogVisible" title="编辑角色" width="560px" :close-on-click-modal="false" destroy-on-close>
      <el-form :model="editRoleForm" label-width="90px">
        <el-form-item label="角色名" required>
          <el-input v-model="editRoleForm.role_name" placeholder="例如：运营专员" maxlength="60" />
        </el-form-item>
        <el-form-item label="角色编码" required>
          <el-input v-model="editRoleForm.role" placeholder="例如：operator" maxlength="50" />
        </el-form-item>
        <el-form-item label="说明">
          <el-input v-model="editRoleForm.description" type="textarea" :rows="3" maxlength="255" show-word-limit />
        </el-form-item>
        <el-form-item label="是否启用">
          <el-switch v-model="editRoleForm.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitEditRole">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped lang="scss">
.role-management { padding: 0; }
.page-header { margin-bottom: 24px; display: flex; justify-content: space-between; align-items: flex-start; }
.page-title { font-size: 24px; font-weight: 600; margin: 0 0 8px 0; }
.page-desc { color: var(--text-secondary); margin: 0; font-size: 14px; }
.content-card { background: var(--bg-card); border: 1px solid var(--border-color); border-radius: var(--radius-lg); padding: 24px; }
</style>
