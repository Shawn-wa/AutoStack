import request from '@/commonBase/api/request'

export interface UserInfo {
  id: number
  username: string
  email: string
  role: string
  status: number
  permissions: string[]
  created_by?: number
  created_at: string
  updated_at?: string
}

export interface UpdateProfileParams {
  email?: string
}

export interface ChangePasswordParams {
  old_password: string
  new_password: string
}

export interface CreateUserParams {
  username: string
  password: string
  email: string
  role: string
  permissions: string[]
}

export interface UpdateUserParams {
  email?: string
  role?: string
  status?: number
  permissions?: string[]
}

export interface UserListResult {
  list: UserInfo[]
  total: number
  page: number
  page_size: number
}

export interface PermissionDef {
  code: string
  name: string
  module: string
}

export interface PermissionsResult {
  permissions: PermissionDef[]
  modules: Record<string, PermissionDef[]>
}

// 获取当前用户信息
export function getProfile() {
  return request.get<any, { data: UserInfo }>('/user/profile')
}

// 更新个人信息
export function updateProfile(data: UpdateProfileParams) {
  return request.put<any, { data: UserInfo }>('/user/profile', data)
}

// 修改密码
export function changePassword(data: ChangePasswordParams) {
  return request.put<any, { data: null }>('/user/password', data)
}

// 获取权限列表（管理员）
export function getPermissions() {
  return request.get<any, { data: PermissionsResult }>('/admin/permissions')
}

// 获取用户列表（管理员）
export function getUsers(page: number = 1, pageSize: number = 10) {
  return request.get<any, { data: UserListResult }>('/admin/users', {
    params: { page, page_size: pageSize }
  })
}

// 获取单个用户详情（管理员）
export function getUser(id: number) {
  return request.get<any, { data: UserInfo }>(`/admin/users/${id}`)
}

// 创建用户（管理员）
export function createUser(data: CreateUserParams) {
  return request.post<any, { data: UserInfo }>('/admin/users', data)
}

// 更新用户（管理员）
export function updateUser(id: number, data: UpdateUserParams) {
  return request.put<any, { data: UserInfo }>(`/admin/users/${id}`, data)
}

// 删除用户（管理员）
export function deleteUser(id: number) {
  return request.delete<any, { data: null }>(`/admin/users/${id}`)
}
