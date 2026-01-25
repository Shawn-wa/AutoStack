package user

import (
	"encoding/json"
	"time"
)

// 用户角色常量
const (
	RoleSuperAdmin = "super_admin"
	RoleAdmin      = "admin"
	RoleUser       = "user"
)

// 用户状态常量
const (
	StatusActive   = 1
	StatusDisabled = 0
)

// 权限常量
const (
	// 用户管理权限
	PermUserCreate = "user:create"
	PermUserUpdate = "user:update"
	PermUserDelete = "user:delete"
	PermUserList   = "user:list"

	// 项目管理权限
	PermProjectCreate = "project:create"
	PermProjectUpdate = "project:update"
	PermProjectDelete = "project:delete"
	PermProjectList   = "project:list"

	// 部署管理权限
	PermDeploymentCreate = "deployment:create"
	PermDeploymentUpdate = "deployment:update"
	PermDeploymentDelete = "deployment:delete"
	PermDeploymentList   = "deployment:list"
	PermDeploymentStart  = "deployment:start"
	PermDeploymentStop   = "deployment:stop"

	// 模板管理权限
	PermTemplateCreate = "template:create"
	PermTemplateUpdate = "template:update"
	PermTemplateDelete = "template:delete"
	PermTemplateList   = "template:list"

	// 平台授权权限
	PermPlatformAuthCreate = "platform_auth:create"
	PermPlatformAuthUpdate = "platform_auth:update"
	PermPlatformAuthDelete = "platform_auth:delete"
	PermPlatformAuthList   = "platform_auth:list"
	PermPlatformAuthSync   = "platform_auth:sync"

	// 订单管理权限
	PermOrderList   = "order:list"
	PermOrderDetail = "order:detail"
)

// PermissionDef 权限定义
type PermissionDef struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Module string `json:"module"`
}

// AllPermissions 所有权限定义
var AllPermissions = []PermissionDef{
	// 用户管理
	{Code: PermUserCreate, Name: "创建用户", Module: "user"},
	{Code: PermUserUpdate, Name: "更新用户", Module: "user"},
	{Code: PermUserDelete, Name: "删除用户", Module: "user"},
	{Code: PermUserList, Name: "查看用户列表", Module: "user"},
	// 项目管理
	{Code: PermProjectCreate, Name: "创建项目", Module: "project"},
	{Code: PermProjectUpdate, Name: "更新项目", Module: "project"},
	{Code: PermProjectDelete, Name: "删除项目", Module: "project"},
	{Code: PermProjectList, Name: "查看项目列表", Module: "project"},
	// 部署管理
	{Code: PermDeploymentCreate, Name: "创建部署", Module: "deployment"},
	{Code: PermDeploymentUpdate, Name: "更新部署", Module: "deployment"},
	{Code: PermDeploymentDelete, Name: "删除部署", Module: "deployment"},
	{Code: PermDeploymentList, Name: "查看部署列表", Module: "deployment"},
	{Code: PermDeploymentStart, Name: "启动部署", Module: "deployment"},
	{Code: PermDeploymentStop, Name: "停止部署", Module: "deployment"},
	// 模板管理
	{Code: PermTemplateCreate, Name: "创建模板", Module: "template"},
	{Code: PermTemplateUpdate, Name: "更新模板", Module: "template"},
	{Code: PermTemplateDelete, Name: "删除模板", Module: "template"},
	{Code: PermTemplateList, Name: "查看模板列表", Module: "template"},
	// 平台授权
	{Code: PermPlatformAuthCreate, Name: "添加平台授权", Module: "platform_auth"},
	{Code: PermPlatformAuthUpdate, Name: "更新平台授权", Module: "platform_auth"},
	{Code: PermPlatformAuthDelete, Name: "删除平台授权", Module: "platform_auth"},
	{Code: PermPlatformAuthList, Name: "查看平台授权", Module: "platform_auth"},
	{Code: PermPlatformAuthSync, Name: "同步订单", Module: "platform_auth"},
	// 订单管理
	{Code: PermOrderList, Name: "查看订单列表", Module: "order"},
	{Code: PermOrderDetail, Name: "查看订单详情", Module: "order"},
}

// UserManagePermissions 用户管理权限列表（仅super_admin可授予）
var UserManagePermissions = []string{
	PermUserCreate, PermUserUpdate, PermUserDelete, PermUserList,
}

// User 用户模型
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:50;not null" json:"username"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	Email        string    `gorm:"uniqueIndex;size:100;not null" json:"email"`
	Role         string    `gorm:"size:20;default:user" json:"role"`
	Status       int       `gorm:"default:1" json:"status"`
	Permissions  string    `gorm:"type:text" json:"-"`
	CreatedBy    *uint     `gorm:"index" json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// IsSuperAdmin 判断是否为超级管理员
func (u *User) IsSuperAdmin() bool {
	return u.Role == RoleSuperAdmin
}

// IsAdmin 判断是否为管理员（包含超级管理员）
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin || u.Role == RoleSuperAdmin
}

// IsActive 判断用户是否激活
func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

// GetPermissions 获取权限列表
func (u *User) GetPermissions() []string {
	// 超级管理员拥有所有权限
	if u.IsSuperAdmin() {
		perms := make([]string, len(AllPermissions))
		for i, p := range AllPermissions {
			perms[i] = p.Code
		}
		return perms
	}

	if u.Permissions == "" {
		return []string{}
	}

	var perms []string
	if err := json.Unmarshal([]byte(u.Permissions), &perms); err != nil {
		return []string{}
	}
	return perms
}

// SetPermissions 设置权限列表
func (u *User) SetPermissions(perms []string) error {
	data, err := json.Marshal(perms)
	if err != nil {
		return err
	}
	u.Permissions = string(data)
	return nil
}

// HasPermission 检查是否有某个权限
func (u *User) HasPermission(perm string) bool {
	if u.IsSuperAdmin() {
		return true
	}
	for _, p := range u.GetPermissions() {
		if p == perm {
			return true
		}
	}
	return false
}

// CanManageRole 检查是否可以管理某个角色
func (u *User) CanManageRole(targetRole string) bool {
	switch u.Role {
	case RoleSuperAdmin:
		return true
	case RoleAdmin:
		return targetRole == RoleUser
	default:
		return false
	}
}
