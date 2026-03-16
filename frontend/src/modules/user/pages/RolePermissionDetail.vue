<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import { getRolePermissions, updateRolePermissions, type PermissionRouteNode, type RolePermissionsResult } from '@/modules/user/api'
import { useUserStore } from '@/modules/auth/stores'

defineOptions({ name: 'RolePermissionDetail' })

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const rolePermissionsData = ref<RolePermissionsResult | null>(null)
const savingRolePermissions = ref(false)

const selectedRole = computed<'super_admin' | 'admin' | 'user'>(() => {
  const raw = String(route.params.role || 'user')
  if (raw === 'super_admin' || raw === 'admin' || raw === 'user') return raw
  return 'user'
})

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

const rolePermissionTree = computed<PermissionRouteNode[]>(() => rolePermissionsData.value?.route_tree || [])

const selectedRolePermissions = computed<string[]>(() => {
  if (!rolePermissionsData.value) return []
  return rolePermissionsData.value.role_permissions[selectedRole.value] || []
})

const selectedPermissionSet = computed(() => new Set(selectedRolePermissions.value))

const isSuperAdminRoleSelected = computed(() => selectedRole.value === 'super_admin')
const canEditSelectedRolePermissions = computed(() => !isSuperAdminRoleSelected.value && editableRoles.value.includes(selectedRole.value))

const actionLabelMap: Record<string, string> = {
  create: '添加',
  update: '更新',
  read: '查看',
  delete: '删除'
}

const getRouteCodes = (node: PermissionRouteNode): string[] => {
  const codes: string[] = []
  const walk = (n: PermissionRouteNode) => {
    for (const p of n.permissions || []) codes.push(p.code)
    for (const c of n.children || []) walk(c)
  }
  walk(node)
  return codes
}

const hasAllCodes = (codes: string[]): boolean => {
  if (codes.length === 0) return false
  return codes.every(code => selectedPermissionSet.value.has(code))
}

const setCodesChecked = (codes: string[], checked: boolean) => {
  const next = new Set(selectedRolePermissions.value)
  if (checked) {
    codes.forEach(code => next.add(code))
  } else {
    codes.forEach(code => next.delete(code))
  }
  if (rolePermissionsData.value) {
    rolePermissionsData.value.role_permissions[selectedRole.value] = Array.from(next)
  }
}

const toggleRouteAll = (node: PermissionRouteNode, checked: boolean) => {
  setCodesChecked(getRouteCodes(node), checked)
}

const fetchRolePermissions = async () => {
  try {
    const res = await getRolePermissions()
    rolePermissionsData.value = res.data
  } catch (error) {
    console.error('获取角色权限配置失败', error)
  }
}

const handleSaveRolePermissions = async () => {
  if (!canEditSelectedRolePermissions.value || isSuperAdminRoleSelected.value) return
  savingRolePermissions.value = true
  try {
    await updateRolePermissions(selectedRole.value, selectedRolePermissions.value)
    ElMessage.success('角色权限已更新')
  } catch (error) {
    console.error('更新角色权限失败', error)
  } finally {
    savingRolePermissions.value = false
  }
}

const handleBack = () => {
  router.push('/users/roles')
}

onMounted(() => {
  if (selectedRole.value === 'super_admin') {
    ElMessage.warning('超级管理员默认拥有全部权限，无需在页面管理')
    handleBack()
    return
  }
  fetchRolePermissions()
})
</script>

<template>
  <div class="role-permission-detail">
    <div class="page-header">
      <div>
        <h1 class="page-title">角色权限详情</h1>
        <p class="page-desc">{{ roleNameMap[selectedRole] }} - 权限配置</p>
      </div>
      <el-button :icon="ArrowLeft" @click="handleBack">返回角色列表</el-button>
    </div>

    <div class="content-card">
      <div class="role-permission-tip">
        <span v-if="isSuperAdminRoleSelected">`super_admin` 默认拥有全部权限，且不可编辑。</span>
        <span v-else-if="!canEditSelectedRolePermissions">当前账号无权编辑 {{ roleNameMap[selectedRole] }} 的权限。</span>
        <span v-else>仅支持按一级/二级目录做批量勾选，动作维度请单独勾选。</span>
      </div>

      <div class="route-permissions-container">
        <div class="perm-grid-header">
          <div class="col-left">一级目录</div>
          <div class="col-mid">二级目录</div>
          <div class="col-right">权限项</div>
        </div>
        <div v-for="moduleNode in rolePermissionTree" :key="moduleNode.key" class="module-block">
          <div v-for="(pageNode, pageIndex) in (moduleNode.children || [])" :key="pageNode.key" class="perm-grid-row">
            <div class="col-left">
              <el-checkbox
                v-if="pageIndex === 0"
                :model-value="hasAllCodes(getRouteCodes(moduleNode))"
                :disabled="!canEditSelectedRolePermissions"
                @change="(val: any) => toggleRouteAll(moduleNode, !!val)"
              >
                {{ moduleNode.name }}
              </el-checkbox>
            </div>
            <div class="col-mid">
              <el-checkbox
                :model-value="hasAllCodes(getRouteCodes(pageNode))"
                :disabled="!canEditSelectedRolePermissions"
                @change="(val: any) => toggleRouteAll(pageNode, !!val)"
              >
                {{ pageNode.name }}
              </el-checkbox>
            </div>
            <div class="col-right">
              <el-checkbox-group
                :model-value="selectedRolePermissions"
                @change="(val: any) => rolePermissionsData && rolePermissionsData.role_permissions && (rolePermissionsData.role_permissions[selectedRole] = (val as string[]))"
              >
                <el-checkbox
                  v-for="perm in pageNode.permissions"
                  :key="perm.code"
                  :label="perm.code"
                  :disabled="!canEditSelectedRolePermissions"
                >
                  {{ actionLabelMap[perm.action] || perm.action }}
                </el-checkbox>
              </el-checkbox-group>
            </div>
          </div>
        </div>
      </div>

      <div class="footer-actions">
        <el-button @click="handleBack">取消</el-button>
        <el-button type="primary" :loading="savingRolePermissions" :disabled="!canEditSelectedRolePermissions" @click="handleSaveRolePermissions">
          保存
        </el-button>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.role-permission-detail {
  padding: 0;
  min-height: calc(100vh - 170px);
  display: flex;
  flex-direction: column;
}
.page-header { margin-bottom: 24px; display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 24px; font-weight: 600; margin: 0 0 8px 0; }
.page-desc { color: var(--text-secondary); margin: 0; font-size: 14px; }
.content-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: 24px;
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.role-permission-tip { font-size: 13px; color: var(--text-secondary); margin-bottom: 12px; }
.route-permissions-container {
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 12px;
  overflow-y: auto;
  flex: 1;
  min-height: 0;
}
.module-block { padding: 10px 0; border-bottom: 1px dashed var(--border-color); }
.module-block:last-child { border-bottom: 0; }
.perm-grid-header {
  display: grid;
  grid-template-columns: 220px 260px 1fr;
  gap: 12px;
  margin-bottom: 8px;
  padding: 0 0 8px 0;
  border-bottom: 1px solid var(--border-color);
  font-size: 13px;
  color: var(--text-secondary);
}
.perm-grid-row {
  display: grid;
  grid-template-columns: 220px 260px 1fr;
  gap: 12px;
  align-items: flex-start;
  margin: 8px 0;
}
.col-left, .col-mid, .col-right { text-align: left; }
.perm-grid-header .col-mid,
.perm-grid-header .col-right,
.perm-grid-row .col-mid,
.perm-grid-row .col-right {
  border-left: 1px solid var(--border-color);
  padding-left: 12px;
}
:deep(.col-right .el-checkbox-group) {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 16px;
}
:deep(.col-right .el-checkbox) { margin-right: 0; }
.footer-actions {
  margin-top: 16px;
  display: flex;
  justify-content: center;
  position: sticky;
  bottom: 0;
  z-index: 10;
  padding: 12px 0 8px;
  background: var(--bg-card);
  border-top: 1px solid var(--border-color);
  gap: 12px;
}
</style>
