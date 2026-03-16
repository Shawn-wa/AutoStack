import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login as loginApi, register as registerApi, type LoginParams, type RegisterParams, type UserInfo } from '@/modules/auth/api'
import { getProfile, getPermissionRoutes, type PermissionRouteNode } from '@/modules/user/api'
import { storage } from '@/utils/storage'
import router from '@/router'

export const useUserStore = defineStore('user', () => {
  // 状态
  const token = ref<string | null>(storage.get('token'))
  const user = ref<UserInfo | null>(storage.get<UserInfo>('user'))
  const permissionRoutes = ref<PermissionRouteNode[]>(storage.get<PermissionRouteNode[]>('permission_routes') || [])

  // 计算属性
  const isLoggedIn = computed(() => !!token.value)
  const isSuperAdmin = computed(() => user.value?.role === 'super_admin')
  const isAdmin = computed(() => user.value?.role === 'admin' || user.value?.role === 'super_admin')
  const username = computed(() => user.value?.username || '')
  const companyName = computed(() => user.value?.company_name || '')
  const companyId = computed(() => user.value?.company_id || 0)
  const permissions = computed(() => user.value?.permissions || [])

  // 登录
  async function login(params: LoginParams) {
    const res = await loginApi(params)
    const { token: newToken, user: userInfo } = res.data

    // 保存到状态
    token.value = newToken
    user.value = userInfo

    // 持久化到 localStorage
    storage.set('token', newToken)
    storage.set('user', userInfo)

    // 登录后立即获取完整 profile（确保权限最新）
    await fetchProfile()
    await fetchPermissionRoutes()

    return res
  }

  // 注册
  async function register(params: RegisterParams) {
    return registerApi(params)
  }

  // 登出
  function logout() {
    token.value = null
    user.value = null
    permissionRoutes.value = []
    storage.remove('token')
    storage.remove('user')
    storage.remove('permission_routes')
    router.push('/login')
  }

  // 获取用户信息
  async function fetchProfile() {
    if (!token.value) return

    try {
      const res = await getProfile()
      user.value = res.data
      storage.set('user', res.data)
      await fetchPermissionRoutes()
    } catch (error) {
      // Token 无效时会被响应拦截器处理
      console.error('获取用户信息失败', error)
    }
  }

  // 获取当前用户可见权限路由树
  async function fetchPermissionRoutes() {
    if (!token.value) return
    try {
      const res = await getPermissionRoutes()
      permissionRoutes.value = res.data.route_tree || []
      storage.set('permission_routes', permissionRoutes.value)
      // 确保 permissions 与后端返回同步
      if (user.value && res.data.permissions) {
        user.value.permissions = res.data.permissions
        storage.set('user', user.value)
      }
    } catch (error) {
      console.error('获取权限路由失败', error)
    }
  }

  // 更新用户信息（本地）
  function updateUserInfo(info: Partial<UserInfo>) {
    if (user.value) {
      user.value = { ...user.value, ...info }
      storage.set('user', user.value)
    }
  }

  // 检查是否有某个权限
  function hasPermission(permission: string): boolean {
    // 超级管理员拥有所有权限
    if (isSuperAdmin.value) return true
    return permissions.value.includes(permission)
  }

  // 检查是否有任一权限
  function hasAnyPermission(...perms: string[]): boolean {
    if (isSuperAdmin.value) return true
    return perms.some(p => permissions.value.includes(p))
  }

  // 检查是否具备某路由节点下的动作权限（用于功能节点显隐）
  function hasRouteAction(routeKey: string, action: string): boolean {
    if (isSuperAdmin.value) return true
    const permissionCode = `route:${routeKey}:${action}`
    return permissions.value.includes(permissionCode)
  }

  // 检查是否可以管理某个角色
  function canManageRole(role: string): boolean {
    if (isSuperAdmin.value) return true
    if (isAdmin.value && role === 'user') return true
    return false
  }

  return {
    // 状态
    token,
    user,
    permissionRoutes,
    // 计算属性
    isLoggedIn,
    isSuperAdmin,
    isAdmin,
    username,
    companyName,
    companyId,
    permissions,
    // 方法
    login,
    register,
    logout,
    fetchProfile,
    fetchPermissionRoutes,
    updateUserInfo,
    hasPermission,
    hasAnyPermission,
    hasRouteAction,
    canManageRole
  }
})
