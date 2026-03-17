package user

// ProfileResponse 用户信息响应
type ProfileResponse struct {
	ID          uint     `json:"id"`
	CompanyID   uint     `json:"company_id"`
	CompanyName string   `json:"company_name"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Role        string   `json:"role"`
	Status      int      `json:"status"`
	Permissions []string `json:"permissions"`
}

// UpdateProfileRequest 更新个人信息请求
type UpdateProfileRequest struct {
	Email string `json:"email" binding:"omitempty,email"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// CreateUserRequest 创建用户请求（管理员）
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"required,email"`
	Role     string `json:"role" binding:"required,min=2,max=50"`
}

// UpdateUserRequest 更新用户请求（管理员）
type UpdateUserRequest struct {
	Email  string `json:"email" binding:"omitempty,email"`
	Role   string `json:"role" binding:"omitempty,min=2,max=50"`
	Status *int   `json:"status" binding:"omitempty,oneof=0 1"`
}

// UserDetailResponse 用户详情响应
type UserDetailResponse struct {
	ID          uint     `json:"id"`
	CompanyID   uint     `json:"company_id"`
	CompanyName string   `json:"company_name"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Role        string   `json:"role"`
	Status      int      `json:"status"`
	Permissions []string `json:"permissions"`
	CreatedBy   *uint    `json:"created_by"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

// UserListItem 用户列表项
type UserListItem struct {
	ID          uint     `json:"id"`
	CompanyID   uint     `json:"company_id"`
	CompanyName string   `json:"company_name"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Role        string   `json:"role"`
	Status      int      `json:"status"`
	Permissions []string `json:"permissions"`
	CreatedBy   *uint    `json:"created_by"`
	CreatedAt   string   `json:"created_at"`
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	List     []UserListItem `json:"list"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

// PermissionsResponse 权限列表响应
type PermissionsResponse struct {
	Permissions []PermissionDef            `json:"permissions"`
	Modules     map[string][]PermissionDef `json:"modules"`
}

// RolePermissionsResponse 角色权限配置响应
type RolePermissionsResponse struct {
	Permissions     []PermissionDef            `json:"permissions"`
	Modules         map[string][]PermissionDef `json:"modules"`
	RouteTree       []PermissionRouteNode      `json:"route_tree"`
	RolePermissions map[string][]string        `json:"role_permissions"`
}

// PermissionRoutesResponse 当前用户权限路由响应
type PermissionRoutesResponse struct {
	Permissions []string              `json:"permissions"`
	RouteTree   []PermissionRouteNode `json:"route_tree"`
}

// UpdateRolePermissionsRequest 更新角色权限请求
type UpdateRolePermissionsRequest struct {
	Permissions []string `json:"permissions" binding:"required"`
}

// RoleItem 角色列表项
type RoleItem struct {
	ID              uint   `json:"id"`
	Role            string `json:"role"` // 角色编码
	RoleName        string `json:"role_name"`
	Description     string `json:"description"`
	Enabled         bool   `json:"enabled"`
	IsSystem        bool   `json:"is_system"`
	PermissionCount int    `json:"permission_count"`
}

// RoleListResponse 角色列表响应
type RoleListResponse struct {
	List []RoleItem `json:"list"`
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Role        string `json:"role" binding:"required,min=2,max=50"`
	RoleName    string `json:"role_name" binding:"required,min=2,max=60"`
	Description string `json:"description" binding:"omitempty,max=255"`
	Enabled     *int   `json:"enabled" binding:"omitempty,oneof=0 1"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Role        string `json:"role" binding:"required,min=2,max=50"`
	RoleName    string `json:"role_name" binding:"required,min=2,max=60"`
	Description string `json:"description" binding:"omitempty,max=255"`
	Enabled     *int   `json:"enabled" binding:"omitempty,oneof=0 1"`
}

// PermissionMigrationRequest 权限初始化迁移请求
type PermissionMigrationRequest struct {
	RebuildRoleBindings bool `json:"rebuild_role_bindings"`
}

// PermissionMigrationResponse 权限初始化迁移结果
type PermissionMigrationResponse struct {
	CompaniesTotal     int `json:"companies_total"`
	CompaniesProcessed int `json:"companies_processed"`
	RebuildRoleBinding int `json:"rebuild_role_binding"` // 1=执行角色绑定补齐，0=仅目录迁移
}
