package user

import (
	"encoding/json"
	"fmt"
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

// 权限动作常量（增删改查）
const (
	PermActionCreate = "create"
	PermActionRead   = "read"
	PermActionUpdate = "update"
	PermActionDelete = "delete"
)

// 一级路由键
const (
	RouteDashboard = "dashboard"
	RouteProduct   = "product"
	RouteOrder     = "order"
	RouteWarehouse = "warehouse"
	RouteShipping  = "shipping"
	RouteSystem    = "system"
)

// 二级页面路由键
const (
	RouteDashboardHome      = "dashboard.home"
	RouteProductLocal       = "product.local_products"
	RouteProductPlatform    = "product.platform_products"
	RouteProductSummary     = "product.order_summary"
	RouteOrderAuths         = "order.platform_auths"
	RouteOrderOrders        = "order.orders"
	RouteOrderCashFlow      = "order.cashflow"
	RouteOrderSettlement    = "order.settlement"
	RouteWarehouseList      = "warehouse.list"
	RouteWarehouseInventory = "warehouse.inventory"
	RouteWarehouseStockIn   = "warehouse.stock_in_orders"
	RouteShippingTemplates  = "shipping.templates"
	RouteSystemUsers        = "system.users"
	RouteSystemRoles        = "system.roles"
	RouteSystemProjects     = "system.projects"
	RouteSystemDeployments  = "system.deployments"
	RouteSystemTemplates    = "system.templates"
)

// 兼容旧逻辑：用户管理可见权限（read）
const PermUserView = "route:system.users:read"
const PermRoleView = "route:system.roles:read"

// 兼容旧常量命名（模块 view -> 页面 read）
const (
	PermDashboardView    = "route:dashboard.home:read"
	PermProductView      = "route:product.local_products:read"
	PermPlatformAuthView = "route:order.platform_auths:read"
	PermOrderView        = "route:order.orders:read"
	PermReportView       = "route:order.cashflow:read"
	PermWarehouseView    = "route:warehouse.list:read"
	PermShippingView     = "route:shipping.templates:read"
)

// PermissionDef 权限定义
type PermissionDef struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Module      string `json:"module"`       // 一级路由键
	RouteKey    string `json:"route_key"`    // 当前节点键
	ParentRoute string `json:"parent_route"` // 一级节点键
	Action      string `json:"action"`       // create/read/update/delete
}

// PermissionRouteNode 路由权限树节点
type PermissionRouteNode struct {
	Key         string                `json:"key"`
	Name        string                `json:"name"`
	Level       int                   `json:"level"` // 1=一级路由 2=二级页面
	Permissions []PermissionDef       `json:"permissions"`
	Children    []PermissionRouteNode `json:"children,omitempty"`
}

func permCode(routeKey, action string) string {
	return fmt.Sprintf("route:%s:%s", routeKey, action)
}

func permDefs(parent, routeKey, routeName string, actions ...string) []PermissionDef {
	defs := make([]PermissionDef, 0, len(actions))
	actionNameMap := map[string]string{
		PermActionCreate: "添加",
		PermActionRead:   "查看",
		PermActionUpdate: "更新",
		PermActionDelete: "删除",
	}
	for _, action := range actions {
		defs = append(defs, PermissionDef{
			Code:        permCode(routeKey, action),
			Name:        routeName + "-" + actionNameMap[action],
			Module:      parent,
			RouteKey:    routeKey,
			ParentRoute: parent,
			Action:      action,
		})
	}
	return defs
}

// PermissionRouteTree 路由权限树（支持按一级/二级路由批量勾选）
var PermissionRouteTree = []PermissionRouteNode{
	{
		Key:   RouteDashboard,
		Name:  "控制台",
		Level: 1,
		Children: []PermissionRouteNode{
			{
				Key:         RouteDashboardHome,
				Name:        "首页",
				Level:       2,
				Permissions: permDefs(RouteDashboard, RouteDashboardHome, "首页", PermActionRead, PermActionUpdate),
			},
		},
	},
	{
		Key:   RouteProduct,
		Name:  "产品管理",
		Level: 1,
		Children: []PermissionRouteNode{
			{Key: RouteProductLocal, Name: "系统产品", Level: 2, Permissions: permDefs(RouteProduct, RouteProductLocal, "系统产品", PermActionCreate, PermActionRead, PermActionUpdate, PermActionDelete)},
			{Key: RouteProductPlatform, Name: "平台产品", Level: 2, Permissions: permDefs(RouteProduct, RouteProductPlatform, "平台产品", PermActionCreate, PermActionRead, PermActionUpdate, PermActionDelete)},
			{Key: RouteProductSummary, Name: "订单汇总", Level: 2, Permissions: permDefs(RouteProduct, RouteProductSummary, "订单汇总", PermActionRead)},
		},
	},
	{
		Key:   RouteOrder,
		Name:  "订单管理",
		Level: 1,
		Children: []PermissionRouteNode{
			{Key: RouteOrderOrders, Name: "订单列表", Level: 2, Permissions: permDefs(RouteOrder, RouteOrderOrders, "订单列表", PermActionRead, PermActionUpdate)},
			{Key: RouteOrderCashFlow, Name: "财务报告", Level: 2, Permissions: permDefs(RouteOrder, RouteOrderCashFlow, "财务报告", PermActionRead)},
			{Key: RouteOrderSettlement, Name: "结算报告", Level: 2, Permissions: permDefs(RouteOrder, RouteOrderSettlement, "结算报告", PermActionRead)},
		},
	},
	{
		Key:   RouteWarehouse,
		Name:  "仓库管理",
		Level: 1,
		Children: []PermissionRouteNode{
			{Key: RouteWarehouseList, Name: "仓库列表", Level: 2, Permissions: permDefs(RouteWarehouse, RouteWarehouseList, "仓库列表", PermActionCreate, PermActionRead, PermActionUpdate, PermActionDelete)},
			{Key: RouteWarehouseInventory, Name: "库存明细", Level: 2, Permissions: permDefs(RouteWarehouse, RouteWarehouseInventory, "库存明细", PermActionRead, PermActionUpdate)},
			{Key: RouteWarehouseStockIn, Name: "入库单", Level: 2, Permissions: permDefs(RouteWarehouse, RouteWarehouseStockIn, "入库单", PermActionCreate, PermActionRead, PermActionUpdate, PermActionDelete)},
		},
	},
	{
		Key:   RouteShipping,
		Name:  "物流管理",
		Level: 1,
		Children: []PermissionRouteNode{
			{Key: RouteShippingTemplates, Name: "运费模板", Level: 2, Permissions: permDefs(RouteShipping, RouteShippingTemplates, "运费模板", PermActionCreate, PermActionRead, PermActionUpdate, PermActionDelete)},
		},
	},
	{
		Key:   RouteSystem,
		Name:  "系统",
		Level: 1,
		Children: []PermissionRouteNode{
			{Key: RouteOrderAuths, Name: "平台授权", Level: 2, Permissions: permDefs(RouteSystem, RouteOrderAuths, "平台授权", PermActionCreate, PermActionRead, PermActionUpdate, PermActionDelete)},
			{Key: RouteSystemUsers, Name: "用户管理", Level: 2, Permissions: permDefs(RouteSystem, RouteSystemUsers, "用户管理", PermActionCreate, PermActionRead, PermActionUpdate, PermActionDelete)},
			{Key: RouteSystemRoles, Name: "角色管理", Level: 2, Permissions: permDefs(RouteSystem, RouteSystemRoles, "角色管理", PermActionCreate, PermActionRead, PermActionUpdate, PermActionDelete)},
			{Key: RouteSystemProjects, Name: "项目管理", Level: 2, Permissions: permDefs(RouteSystem, RouteSystemProjects, "项目管理", PermActionCreate, PermActionRead, PermActionUpdate, PermActionDelete)},
			{Key: RouteSystemDeployments, Name: "部署管理", Level: 2, Permissions: permDefs(RouteSystem, RouteSystemDeployments, "部署管理", PermActionCreate, PermActionRead, PermActionUpdate, PermActionDelete)},
			{Key: RouteSystemTemplates, Name: "模板市场", Level: 2, Permissions: permDefs(RouteSystem, RouteSystemTemplates, "模板市场", PermActionCreate, PermActionRead, PermActionUpdate, PermActionDelete)},
		},
	},
}

func flattenPermissions(tree []PermissionRouteNode) []PermissionDef {
	var defs []PermissionDef
	var walk func(nodes []PermissionRouteNode)
	walk = func(nodes []PermissionRouteNode) {
		for _, n := range nodes {
			defs = append(defs, n.Permissions...)
			if len(n.Children) > 0 {
				walk(n.Children)
			}
		}
	}
	walk(tree)
	return defs
}

// AllPermissions 所有权限定义
var AllPermissions = flattenPermissions(PermissionRouteTree)

// UserManagePermissions 用户管理权限
var UserManagePermissions = []string{PermUserView}

// AllPermissionCodes 返回全部权限编码
func AllPermissionCodes() []string {
	codes := make([]string, 0, len(AllPermissions))
	for _, p := range AllPermissions {
		codes = append(codes, p.Code)
	}
	return codes
}

// DefaultRolePermissionsByRole 默认角色权限（super_admin 固定全量）
func DefaultRolePermissionsByRole(role string) []string {
	switch role {
	case RoleSuperAdmin:
		return AllPermissionCodes()
	case RoleAdmin:
		// 默认管理员拥有全部页面权限（可在角色权限管理中调整）
		return AllPermissionCodes()
	case RoleUser:
		// 默认普通用户只给只读权限
		var perms []string
		for _, p := range AllPermissions {
			if p.Action == PermActionRead {
				perms = append(perms, p.Code)
			}
		}
		return perms
	default:
		return []string{}
	}
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

// RolePermission 角色权限（旧版：JSON 存储，保留用于兼容迁移）
type RolePermission struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CompanyID   uint      `gorm:"not null;index:idx_company_role,unique" json:"company_id"`
	Role        string    `gorm:"size:20;not null;index:idx_company_role,unique" json:"role"`
	Permissions string    `gorm:"type:text" json:"-"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (RolePermission) TableName() string {
	return "role_permissions"
}

func (rp *RolePermission) GetPermissions() []string {
	if rp == nil || rp.Permissions == "" {
		return []string{}
	}
	var perms []string
	if err := json.Unmarshal([]byte(rp.Permissions), &perms); err != nil {
		return []string{}
	}
	return perms
}

func (rp *RolePermission) SetPermissions(perms []string) error {
	data, err := json.Marshal(perms)
	if err != nil {
		return err
	}
	rp.Permissions = string(data)
	return nil
}

// Permission 权限节点（支持无限级父子关系）
type Permission struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ParentID  *uint     `gorm:"index" json:"parent_id"`
	NodeKey   string    `gorm:"size:120;not null;uniqueIndex" json:"node_key"` // 唯一节点键
	Code      string    `gorm:"size:120;index" json:"code"`                    // 动作权限编码（仅叶子动作节点有值）
	Name      string    `gorm:"size:120;not null" json:"name"`
	NodeType  string    `gorm:"size:20;not null;index" json:"node_type"` // module/route/action/custom
	Action    string    `gorm:"size:20;default:''" json:"action"`        // create/read/update/delete
	Sort      int       `gorm:"default:0" json:"sort"`
	Enabled   int       `gorm:"default:1" json:"enabled"` // 1 启用，0 禁用
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Permission) TableName() string {
	return "permissions"
}

// RolePermissionBinding 角色-权限绑定（按企业隔离）
type RolePermissionBinding struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	CompanyID    uint      `gorm:"not null;index:idx_company_role_permission,unique" json:"company_id"`
	Role         string    `gorm:"size:20;not null;index:idx_company_role_permission,unique;index" json:"role"`
	PermissionID uint      `gorm:"not null;index:idx_company_role_permission,unique;index" json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (RolePermissionBinding) TableName() string {
	return "role_permission_bindings"
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
