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

// 权限常量 — 模块级查看权限（控制菜单/路由可见性）
const (
	PermDashboardView    = "dashboard:view"
	PermProductView      = "product:view"
	PermPlatformAuthView = "platform_auth:view"
	PermOrderView        = "order:view"
	PermReportView       = "report:view"
	PermWarehouseView    = "warehouse:view"
	PermShippingView     = "shipping:view"
	PermUserView         = "user:view"
)

// PermissionDef 权限定义
type PermissionDef struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Module string `json:"module"`
}

// AllPermissions 所有模块权限定义
var AllPermissions = []PermissionDef{
	{Code: PermDashboardView, Name: "控制台", Module: "dashboard"},
	{Code: PermProductView, Name: "产品管理", Module: "product"},
	{Code: PermPlatformAuthView, Name: "平台授权", Module: "platform_auth"},
	{Code: PermOrderView, Name: "订单管理", Module: "order"},
	{Code: PermReportView, Name: "报表", Module: "report"},
	{Code: PermWarehouseView, Name: "仓库管理", Module: "warehouse"},
	{Code: PermShippingView, Name: "物流管理", Module: "shipping"},
	{Code: PermUserView, Name: "用户管理", Module: "user"},
}

// UserManagePermissions 用户管理权限（仅super_admin可授予admin）
var UserManagePermissions = []string{
	PermUserView,
}

// User 用户模型
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	CompanyID    uint      `gorm:"index;not null;default:0" json:"company_id"`
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
