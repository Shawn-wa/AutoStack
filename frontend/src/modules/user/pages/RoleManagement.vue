<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getRolePermissions, type RolePermissionsResult } from '@/modules/user/api'
import { useUserStore } from '@/modules/auth/stores'

defineOptions({ name: 'RoleManagement' })

const userStore = useUserStore()
const router = useRouter()

const rolePermissionsData = ref<RolePermissionsResult | null>(null)
const loading = ref(false)

const roleNameMap: Record<string, string> = {
  super_admin: '超级管理员',
  admin: '管理员',
  user: '普通用户'
}

const editableRoles = computed(() => {
  if (userStore.isSuperAdmin) return ['admin', 'user']
  if (userStore.isAdmin) return ['user']
  return []
})

const fetchRolePermissions = async () => {
  loading.value = true
  try {
    const res = await getRolePermissions()
    rolePermissionsData.value = res.data
  } catch (error) {
    console.error('获取角色权限配置失败', error)
  } finally {
    loading.value = false
  }
}

const roleRows = computed(() => {
  if (!rolePermissionsData.value) return []
  const rolePermissions = rolePermissionsData.value.role_permissions || {}
  const allRoles: Array<'admin' | 'user'> = ['admin', 'user']
  return allRoles.map(role => ({
    role,
    roleName: roleNameMap[role],
    permissionCount: rolePermissions[role]?.length || 0,
    canEdit: editableRoles.value.includes(role)
  }))
})

const handleEdit = (role: 'super_admin' | 'admin' | 'user') => {
  router.push(`/users/roles/${role}`)
}

onMounted(() => {
  fetchRolePermissions()
})
</script>

<template>
  <div class="role-management">
    <div class="page-header">
      <h1 class="page-title">角色管理</h1>
      <p class="page-desc">先选择角色，再进入详细权限页面编辑</p>
    </div>

    <div class="content-card">
      <el-table v-loading="loading" :data="roleRows" stripe>
        <el-table-column prop="roleName" label="角色" min-width="180" />
        <el-table-column prop="role" label="角色编码" min-width="180" />
        <el-table-column prop="permissionCount" label="权限数量" width="120" />
        <el-table-column label="说明" min-width="260">
          <template #default="{ row }">
            <span v-if="row.role === 'super_admin'">默认拥有全部权限，不可编辑</span>
            <span v-else>可进入详情页做权限增减</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="handleEdit(row.role)" :disabled="!row.canEdit">
              编辑权限
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<style scoped lang="scss">
.role-management {
  padding: 0;
}

.page-header {
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

</style>
