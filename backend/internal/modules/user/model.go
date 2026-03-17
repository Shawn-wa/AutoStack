package user

// 本文件为类型别名定义，实际实体已迁移至 repository 层
// 保持向后兼容，避免修改现有代码的导入路径

import (
	companyRepo "autostack/internal/repository/company"
	userRepo "autostack/internal/repository/user"
)

// ========== 企业域类型别名 ==========

// Company 企业模型
type Company = companyRepo.Company

// ========== 用户域类型别名 ==========

// User 用户模型
type User = userRepo.User

// RolePermission 角色权限模型
type RolePermission = userRepo.RolePermission

// Permission 权限节点模型
type Permission = userRepo.Permission

// RolePermissionBinding 角色权限绑定模型
type RolePermissionBinding = userRepo.RolePermissionBinding

// RoleDefinition 角色定义模型
type RoleDefinition = userRepo.RoleDefinition

// PermissionDef 权限定义
type PermissionDef = userRepo.PermissionDef

// PermissionRouteNode 路由权限树节点
type PermissionRouteNode = userRepo.PermissionRouteNode

// 注意：PermissionsResponse 在 dto.go 中定义，不在此处别名

// ========== 用户角色常量别名 ==========

const (
	RoleSuperAdmin = userRepo.RoleSuperAdmin
	RoleAdmin      = userRepo.RoleAdmin
	RoleUser       = userRepo.RoleUser
)

// ========== 用户状态常量别名 ==========

const (
	StatusActive   = userRepo.StatusActive
	StatusDisabled = userRepo.StatusDisabled
)

// ========== 模块权限常量别名 ==========

const (
	PermDashboardView    = userRepo.PermDashboardView
	PermProductView      = userRepo.PermProductView
	PermPlatformAuthView = userRepo.PermPlatformAuthView
	PermOrderView        = userRepo.PermOrderView
	PermReportView       = userRepo.PermReportView
	PermWarehouseView    = userRepo.PermWarehouseView
	PermShippingView     = userRepo.PermShippingView
	PermUserView         = userRepo.PermUserView
)

// ========== 权限列表别名 ==========

var AllPermissions = userRepo.AllPermissions
var UserManagePermissions = userRepo.UserManagePermissions
var PermissionRouteTree = userRepo.PermissionRouteTree

func DefaultRolePermissionsByRole(role string) []string {
	return userRepo.DefaultRolePermissionsByRole(role)
}
