import request from '@/commonBase/api/request'

export interface UserInfo {
  id: number
  company_id: number
  company_name: string
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
}

export interface UpdateUserParams {
  email?: string
  role?: string
  status?: number
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
  route_key: string
  parent_route: string
  action: 'create' | 'read' | 'update' | 'delete'
}

export interface PermissionsResult {
  permissions: PermissionDef[]
  modules: Record<string, PermissionDef[]>
}

export interface PermissionRouteNode {
  key: string
  name: string
  level: number
  permissions: PermissionDef[]
  children?: PermissionRouteNode[]
}

export interface RolePermissionsResult {
  permissions: PermissionDef[]
  modules: Record<string, PermissionDef[]>
  route_tree: PermissionRouteNode[]
  role_permissions: Record<string, string[]>
}

export interface PermissionRoutesResult {
  permissions: string[]
  route_tree: PermissionRouteNode[]
}

export interface RoleItem {
  id: number
  role: string
  role_name: string
  description: string
  enabled: boolean
  is_system: boolean
  permission_count: number
}

export interface RoleListResult {
  list: RoleItem[]
}

export interface CreateRoleParams {
  role: string
  role_name: string
  description?: string
  enabled?: number
}

export interface UpdateRoleParams {
  role: string
  role_name: string
  description?: string
  enabled?: number
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

// 获取角色权限配置（管理员）
export function getRolePermissions() {
  return request.get<any, { data: RolePermissionsResult }>('/admin/role-permissions')
}

// 获取当前用户可见权限路由树
export function getPermissionRoutes() {
  return request.get<any, { data: PermissionRoutesResult }>('/user/permission-routes')
}

// 获取角色列表（管理员）
export function getRoles() {
  return request.get<any, { data: RoleListResult }>('/admin/roles')
}

// 创建自定义角色（管理员）
export function createRole(data: CreateRoleParams) {
  return request.post<any, { data: null }>('/admin/roles', data)
}

export function updateRole(id: number, data: UpdateRoleParams) {
  return request.put<any, { data: null }>(`/admin/roles/${id}`, data)
}

export function deleteRole(id: number) {
  return request.delete<any, { data: null }>(`/admin/roles/${id}`)
}

// 更新角色权限配置（管理员）
export function updateRolePermissions(role: string, permissions: string[]) {
  return request.put<any, { data: null }>(`/admin/role-permissions/${role}`, { permissions })
}

// 获取用户列表（管理员）
export function getUsers(params: { page?: number; page_size?: number; keyword?: string; role?: string } = {}) {
  return request.get<any, { data: UserListResult }>('/admin/users', {
    params: { page: params.page || 1, page_size: params.page_size || 10, ...params.keyword && { keyword: params.keyword }, ...params.role && { role: params.role } }
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
