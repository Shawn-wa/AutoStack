package user

// ProfileResponse 用户信息响应
type ProfileResponse struct {
	ID          uint     `json:"id"`
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
	Username    string   `json:"username" binding:"required,min=3,max=20"`
	Password    string   `json:"password" binding:"required,min=6"`
	Email       string   `json:"email" binding:"required,email"`
	Role        string   `json:"role" binding:"required,oneof=admin user"`
	Permissions []string `json:"permissions"`
}

// UpdateUserRequest 更新用户请求（管理员）
type UpdateUserRequest struct {
	Email       string   `json:"email" binding:"omitempty,email"`
	Role        string   `json:"role" binding:"omitempty,oneof=super_admin admin user"`
	Status      *int     `json:"status" binding:"omitempty,oneof=0 1"`
	Permissions []string `json:"permissions" binding:"omitempty"`
}

// UserDetailResponse 用户详情响应
type UserDetailResponse struct {
	ID          uint     `json:"id"`
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
	Permissions []PermissionDef         `json:"permissions"`
	Modules     map[string][]PermissionDef `json:"modules"`
}
