import request from '@/commonBase/api/request'

export interface LoginParams {
  username: string
  password: string
}

export interface RegisterParams {
  username: string
  password: string
  email: string
}

export interface UserInfo {
  id: number
  username: string
  email: string
  role: string
  status?: number
  permissions?: string[]
}

export interface LoginResult {
  token: string
  user: UserInfo
}

// 登录
export function login(data: LoginParams) {
  return request.post<any, { data: LoginResult }>('/auth/login', data)
}

// 注册
export function register(data: RegisterParams) {
  return request.post<any, { data: UserInfo }>('/auth/register', data)
}
