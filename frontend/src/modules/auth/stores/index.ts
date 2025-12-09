import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login as loginApi, register as registerApi, type LoginParams, type RegisterParams, type UserInfo } from '@/modules/auth/api'
import { getProfile } from '@/modules/user/api'
import { storage } from '@/utils/storage'
import router from '@/router'

export const useUserStore = defineStore('user', () => {
  // 状态
  const token = ref<string | null>(storage.get('token'))
  const user = ref<UserInfo | null>(storage.get<UserInfo>('user'))

  // 计算属性
  const isLoggedIn = computed(() => !!token.value)
  const isSuperAdmin = computed(() => user.value?.role === 'super_admin')
  const isAdmin = computed(() => user.value?.role === 'admin' || user.value?.role === 'super_admin')
  const username = computed(() => user.value?.username || '')
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
    storage.remove('token')
    storage.remove('user')
    router.push('/login')
  }

  // 获取用户信息
  async function fetchProfile() {
    if (!token.value) return

    try {
      const res = await getProfile()
      user.value = res.data
      storage.set('user', res.data)
    } catch (error) {
      // Token 无效时会被响应拦截器处理
      console.error('获取用户信息失败', error)
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
    // 计算属性
    isLoggedIn,
    isSuperAdmin,
    isAdmin,
    username,
    permissions,
    // 方法
    login,
    register,
    logout,
    fetchProfile,
    updateUserInfo,
    hasPermission,
    hasAnyPermission,
    canManageRole
  }
})
