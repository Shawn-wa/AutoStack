package user

// 本文件为类型别名定义，实际实体已迁移至 repository 层
// 保持向后兼容，避免修改现有代码的导入路径

import (
	userRepo "autostack/internal/repository/user"
)

// ========== 用户域类型别名 ==========

// User 用户模型
type User = userRepo.User

// PermissionDef 权限定义
type PermissionDef = userRepo.PermissionDef

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

// ========== 权限常量别名 ==========

const (
	// 用户管理权限
	PermUserCreate = userRepo.PermUserCreate
	PermUserUpdate = userRepo.PermUserUpdate
	PermUserDelete = userRepo.PermUserDelete
	PermUserList   = userRepo.PermUserList

	// 项目管理权限
	PermProjectCreate = userRepo.PermProjectCreate
	PermProjectUpdate = userRepo.PermProjectUpdate
	PermProjectDelete = userRepo.PermProjectDelete
	PermProjectList   = userRepo.PermProjectList

	// 部署管理权限
	PermDeploymentCreate = userRepo.PermDeploymentCreate
	PermDeploymentUpdate = userRepo.PermDeploymentUpdate
	PermDeploymentDelete = userRepo.PermDeploymentDelete
	PermDeploymentList   = userRepo.PermDeploymentList
	PermDeploymentStart  = userRepo.PermDeploymentStart
	PermDeploymentStop   = userRepo.PermDeploymentStop

	// 模板管理权限
	PermTemplateCreate = userRepo.PermTemplateCreate
	PermTemplateUpdate = userRepo.PermTemplateUpdate
	PermTemplateDelete = userRepo.PermTemplateDelete
	PermTemplateList   = userRepo.PermTemplateList

	// 平台授权权限
	PermPlatformAuthCreate = userRepo.PermPlatformAuthCreate
	PermPlatformAuthUpdate = userRepo.PermPlatformAuthUpdate
	PermPlatformAuthDelete = userRepo.PermPlatformAuthDelete
	PermPlatformAuthList   = userRepo.PermPlatformAuthList
	PermPlatformAuthSync   = userRepo.PermPlatformAuthSync

	// 订单管理权限
	PermOrderList   = userRepo.PermOrderList
	PermOrderDetail = userRepo.PermOrderDetail
)

// ========== 权限列表别名 ==========

var AllPermissions = userRepo.AllPermissions
var UserManagePermissions = userRepo.UserManagePermissions
