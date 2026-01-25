package user

// UserQuery 用户查询条件
type UserQuery struct {
	Page     int
	PageSize int
	Role     string
	Status   *int
	Keyword  string
}

// PermissionsResponse 权限响应
type PermissionsResponse struct {
	Permissions []PermissionDef            `json:"permissions"`
	Modules     map[string][]PermissionDef `json:"modules"`
}
