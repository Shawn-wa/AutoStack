package user

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"autostack/internal/commonBase/database"
	"autostack/internal/migration/companyid"
	"autostack/internal/repository"
	companyRepo "autostack/internal/repository/company"
	userRepo "autostack/internal/repository/user"
	"autostack/pkg/response"
)

// userService 用户服务实例
var userService *Service

// InitHandler 初始化 Handler，注入 Service 依赖
func InitHandler(db *gorm.DB) {
	txManager := repository.NewTxManager(db)
	userService = NewService(
		txManager,
		userRepo.NewUserRepository(db),
		companyRepo.NewCompanyRepository(db),
	)
}

// GetService 获取服务实例（用于外部调用）
func GetService() *Service {
	return userService
}

// getCompanyName 获取企业名称
func getCompanyName(companyID uint) string {
	if companyID == 0 {
		return ""
	}
	company, err := userService.GetCompanyByID(companyID)
	if err != nil {
		return ""
	}
	return company.Name
}

// GetProfile 获取当前用户信息
func GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	uid := parseUserID(userID)
	if uid == 0 {
		response.Error(c, http.StatusInternalServerError, "用户ID类型错误")
		return
	}

	user, err := userService.GetUserByID(uid)
	if err != nil {
		if err == ErrUserNotFound {
			response.Error(c, http.StatusNotFound, "用户不存在")
			return
		}
		response.Error(c, http.StatusInternalServerError, "获取用户信息失败")
		return
	}

	response.Success(c, http.StatusOK, "获取成功", ProfileResponse{
		ID:          user.ID,
		CompanyID:   user.CompanyID,
		CompanyName: getCompanyName(user.CompanyID),
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
		Status:      user.Status,
		Permissions: userService.GetEffectivePermissions(user),
	})
}

// UpdateProfile 更新个人信息
func UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	uid := parseUserID(userID)
	if uid == 0 {
		response.Error(c, http.StatusInternalServerError, "用户ID类型错误")
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	updates := make(map[string]interface{})
	if req.Email != "" {
		updates["email"] = req.Email
	}

	if len(updates) == 0 {
		response.Error(c, http.StatusBadRequest, "没有要更新的内容")
		return
	}

	user, err := userService.UpdateUser(uid, updates)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "更新失败")
		return
	}

	response.Success(c, http.StatusOK, "更新成功", ProfileResponse{
		ID:          user.ID,
		CompanyID:   user.CompanyID,
		CompanyName: getCompanyName(user.CompanyID),
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
		Status:      user.Status,
		Permissions: userService.GetEffectivePermissions(user),
	})
}

// ChangePassword 修改密码
func ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	uid := parseUserID(userID)
	if uid == 0 {
		response.Error(c, http.StatusInternalServerError, "用户ID类型错误")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	err := userService.ChangePassword(uid, req.OldPassword, req.NewPassword)
	if err != nil {
		if err == ErrInvalidOldPassword {
			response.Error(c, http.StatusBadRequest, "原密码错误")
			return
		}
		response.Error(c, http.StatusInternalServerError, "修改密码失败")
		return
	}

	response.Success(c, http.StatusOK, "密码修改成功", nil)
}

// GetPermissions 获取权限列表（管理员）
func GetPermissions(c *gin.Context) {
	perms := userService.GetAllPermissions()
	response.Success(c, http.StatusOK, "获取成功", perms)
}

// GetPermissionRoutes 获取当前用户可见权限路由树
func GetPermissionRoutes(c *gin.Context) {
	currentUser, err := getCurrentUser(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	permissionCodes := userService.GetEffectivePermissions(currentUser)
	routeTree, err := userService.GetPermissionRouteTree()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取权限路由失败")
		return
	}
	filteredTree := userService.FilterPermissionRouteTree(routeTree, permissionCodes)

	response.Success(c, http.StatusOK, "获取成功", PermissionRoutesResponse{
		Permissions: permissionCodes,
		RouteTree:   filteredTree,
	})
}

// GetRolePermissions 获取角色权限配置（管理员）
func GetRolePermissions(c *gin.Context) {
	currentUser, err := getCurrentUser(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}
	rolePerms, err := userService.GetRolePermissions(currentUser.CompanyID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取角色权限失败")
		return
	}
	perms := userService.GetAllPermissions()
	routeTree, treeErr := userService.GetPermissionRouteTree()
	if treeErr != nil {
		response.Error(c, http.StatusInternalServerError, "获取权限路由树失败")
		return
	}
	response.Success(c, http.StatusOK, "获取成功", RolePermissionsResponse{
		Permissions:     perms.Permissions,
		Modules:         perms.Modules,
		RouteTree:       routeTree,
		RolePermissions: rolePerms,
	})
}

// ListRoles 获取角色列表
func ListRoles(c *gin.Context) {
	currentUser, err := getCurrentUser(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}
	roles, err := userService.ListRoles(currentUser.CompanyID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取角色列表失败")
		return
	}
	response.Success(c, http.StatusOK, "获取成功", RoleListResponse{List: roles})
}

// CreateRole 创建自定义角色
func CreateRole(c *gin.Context) {
	currentUser, err := getCurrentUser(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}
	enabled := 1
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	if err := userService.CreateRole(currentUser, req.Role, req.RoleName, req.Description, enabled); err != nil {
		switch err {
		case ErrPermissionDenied:
			response.Error(c, http.StatusForbidden, "无权创建该角色")
		case ErrRoleExists:
			response.Error(c, http.StatusConflict, "角色编码已存在")
		case ErrRoleNameExists:
			response.Error(c, http.StatusConflict, "角色名已存在")
		default:
			response.Error(c, http.StatusInternalServerError, "创建角色失败")
		}
		return
	}
	response.Success(c, http.StatusCreated, "创建成功", nil)
}

// UpdateRole 更新角色
func UpdateRole(c *gin.Context) {
	currentUser, err := getCurrentUser(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的角色ID")
		return
	}
	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}
	enabled := 1
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	if err := userService.UpdateRole(currentUser, uint(id), req.Role, req.RoleName, req.Description, enabled); err != nil {
		switch err {
		case ErrPermissionDenied:
			response.Error(c, http.StatusForbidden, "无权更新该角色")
		case ErrRoleNotFound:
			response.Error(c, http.StatusNotFound, "角色不存在")
		case ErrRoleExists:
			response.Error(c, http.StatusConflict, "角色编码已存在")
		case ErrRoleNameExists:
			response.Error(c, http.StatusConflict, "角色名已存在")
		default:
			response.Error(c, http.StatusInternalServerError, "更新角色失败")
		}
		return
	}
	response.Success(c, http.StatusOK, "更新成功", nil)
}

// DeleteRole 删除角色
func DeleteRole(c *gin.Context) {
	currentUser, err := getCurrentUser(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的角色ID")
		return
	}
	if err := userService.DeleteRole(currentUser, uint(id)); err != nil {
		switch err {
		case ErrPermissionDenied:
			response.Error(c, http.StatusForbidden, "无权删除该角色")
		case ErrRoleNotFound:
			response.Error(c, http.StatusNotFound, "角色不存在")
		case ErrRoleInUse:
			response.Error(c, http.StatusConflict, "角色已被用户使用，无法删除")
		default:
			response.Error(c, http.StatusInternalServerError, "删除角色失败")
		}
		return
	}
	response.Success(c, http.StatusOK, "删除成功", nil)
}

// UpdateRolePermissions 更新角色权限（管理员）
func UpdateRolePermissions(c *gin.Context) {
	currentUser, err := getCurrentUser(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}
	role := c.Param("role")
	if role == RoleSuperAdmin {
		response.Error(c, http.StatusForbidden, "超级管理员权限不可编辑")
		return
	}
	var req UpdateRolePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}
	if err := userService.UpdateRolePermissions(currentUser, role, req.Permissions); err != nil {
		if err == ErrPermissionDenied {
			response.Error(c, http.StatusForbidden, "无权修改该角色权限")
			return
		}
		if err == ErrRoleNotFound {
			response.Error(c, http.StatusNotFound, "角色不存在")
			return
		}
		response.Error(c, http.StatusInternalServerError, "更新角色权限失败")
		return
	}
	response.Success(c, http.StatusOK, "更新成功", nil)
}

// CreateUser 创建用户（管理员）
func CreateUser(c *gin.Context) {
	currentUser, err := getCurrentUser(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	if !userService.CanManageRole(currentUser, req.Role) {
		response.Error(c, http.StatusForbidden, "无权创建该角色用户")
		return
	}

	createdBy := currentUser.ID
	user, err := userService.CreateUserWithPermissions(
		req.Username,
		req.Password,
		req.Email,
		req.Role,
		nil,
		&createdBy,
		currentUser.CompanyID,
	)
	if err != nil {
		if err == ErrUserExists {
			response.Error(c, http.StatusConflict, "用户名或邮箱已存在")
			return
		}
		if err == ErrRoleNotFound {
			response.Error(c, http.StatusBadRequest, "角色不存在，请先在角色管理中创建")
			return
		}
		if err == ErrRoleDisabled {
			response.Error(c, http.StatusBadRequest, "角色已禁用，不能分配给用户")
			return
		}
		response.Error(c, http.StatusInternalServerError, "创建用户失败")
		return
	}

	response.Success(c, http.StatusCreated, "创建成功", UserDetailResponse{
		ID:          user.ID,
		CompanyID:   user.CompanyID,
		CompanyName: getCompanyName(user.CompanyID),
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
		Status:      user.Status,
		Permissions: userService.GetEffectivePermissions(user),
		CreatedBy:   user.CreatedBy,
		CreatedAt:   user.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   user.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

// ListUsers 获取用户列表（管理员，同企业内）
func ListUsers(c *gin.Context) {
	currentUser, err := getCurrentUser(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.Query("keyword")
	role := c.Query("role")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	users, total, err := userService.ListUsers(currentUser.CompanyID, keyword, role, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取用户列表失败")
		return
	}

	companyName := getCompanyName(currentUser.CompanyID)

	list := make([]UserListItem, len(users))
	for i, u := range users {
		list[i] = UserListItem{
			ID:          u.ID,
			CompanyID:   u.CompanyID,
			CompanyName: companyName,
			Username:    u.Username,
			Email:       u.Email,
			Role:        u.Role,
			Status:      u.Status,
			Permissions: userService.GetEffectivePermissions(&u),
			CreatedBy:   u.CreatedBy,
			CreatedAt:   u.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	response.Success(c, http.StatusOK, "获取成功", UserListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// GetUser 获取单个用户详情（管理员）
func GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的用户ID")
		return
	}

	user, err := userService.GetUserByID(uint(id))
	if err != nil {
		if err == ErrUserNotFound {
			response.Error(c, http.StatusNotFound, "用户不存在")
			return
		}
		response.Error(c, http.StatusInternalServerError, "获取用户失败")
		return
	}

	response.Success(c, http.StatusOK, "获取成功", UserDetailResponse{
		ID:          user.ID,
		CompanyID:   user.CompanyID,
		CompanyName: getCompanyName(user.CompanyID),
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
		Status:      user.Status,
		Permissions: userService.GetEffectivePermissions(user),
		CreatedBy:   user.CreatedBy,
		CreatedAt:   user.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   user.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

// UpdateUser 更新用户（管理员）
func UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的用户ID")
		return
	}

	currentUser, err := getCurrentUser(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	targetUser, err := userService.GetUserByID(uint(id))
	if err != nil {
		if err == ErrUserNotFound {
			response.Error(c, http.StatusNotFound, "用户不存在")
			return
		}
		response.Error(c, http.StatusInternalServerError, "获取用户失败")
		return
	}

	if currentUser.ID == targetUser.ID {
		response.Error(c, http.StatusBadRequest, "不能通过此接口修改自己")
		return
	}

	if !userService.CanManageUser(currentUser, targetUser) {
		response.Error(c, http.StatusForbidden, "无权管理该用户")
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	updates := make(map[string]interface{})
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if req.Role != "" && req.Role != targetUser.Role {
		if !userService.CanManageRole(currentUser, req.Role) {
			response.Error(c, http.StatusForbidden, "无权设置该角色")
			return
		}
		exists, err := userService.RoleExists(currentUser.CompanyID, req.Role)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "校验角色失败")
			return
		}
		if !exists {
			response.Error(c, http.StatusBadRequest, "角色不存在，请先在角色管理中创建")
			return
		}
		roleDef, err := userService.GetRoleDefinitionByCode(currentUser.CompanyID, req.Role)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "校验角色失败")
			return
		}
		if roleDef != nil && roleDef.Enabled != 1 {
			response.Error(c, http.StatusBadRequest, "角色已禁用，不能分配给用户")
			return
		}
		if targetUser.IsSuperAdmin() {
			response.Error(c, http.StatusForbidden, "不能修改超级管理员的角色")
			return
		}
		updates["role"] = req.Role
	}

	if len(updates) == 0 {
		response.Error(c, http.StatusBadRequest, "没有要更新的内容")
		return
	}

	user, err := userService.UpdateUser(uint(id), updates)
	if err != nil {
		if err == ErrUserNotFound {
			response.Error(c, http.StatusNotFound, "用户不存在")
			return
		}
		response.Error(c, http.StatusInternalServerError, "更新失败")
		return
	}

	response.Success(c, http.StatusOK, "更新成功", UserDetailResponse{
		ID:          user.ID,
		CompanyID:   user.CompanyID,
		CompanyName: getCompanyName(user.CompanyID),
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
		Status:      user.Status,
		Permissions: userService.GetEffectivePermissions(user),
		CreatedBy:   user.CreatedBy,
		CreatedAt:   user.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   user.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

// DeleteUser 删除用户（管理员）
func DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的用户ID")
		return
	}

	currentUser, err := getCurrentUser(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	if currentUser.ID == uint(id) {
		response.Error(c, http.StatusBadRequest, "不能删除自己")
		return
	}

	targetUser, err := userService.GetUserByID(uint(id))
	if err != nil {
		if err == ErrUserNotFound {
			response.Error(c, http.StatusNotFound, "用户不存在")
			return
		}
		response.Error(c, http.StatusInternalServerError, "获取用户失败")
		return
	}

	if targetUser.IsSuperAdmin() {
		response.Error(c, http.StatusForbidden, "不能删除超级管理员")
		return
	}

	if !userService.CanManageUser(currentUser, targetUser) {
		response.Error(c, http.StatusForbidden, "无权删除该用户")
		return
	}

	if err := userService.DeleteUser(uint(id)); err != nil {
		if err == ErrUserNotFound {
			response.Error(c, http.StatusNotFound, "用户不存在")
			return
		}
		response.Error(c, http.StatusInternalServerError, "删除失败")
		return
	}

	response.Success(c, http.StatusOK, "删除成功", nil)
}

// RunCompanyIDMigration 手动触发 company_id 全量迁移（可重复调用，幂等）
func RunCompanyIDMigration(c *gin.Context) {
	currentUser, err := getCurrentUser(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}
	if !currentUser.IsSuperAdmin() {
		response.Error(c, http.StatusForbidden, "仅超级管理员可执行迁移")
		return
	}

	if err := companyid.Run(database.GetDB()); err != nil {
		response.Error(c, http.StatusInternalServerError, "迁移失败: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "company_id 迁移执行完成", gin.H{
		"repeatable": true,
		"idempotent": true,
	})
}

// getCurrentUser 获取当前登录用户
func getCurrentUser(c *gin.Context) (*User, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return nil, ErrUserNotFound
	}

	uid := parseUserID(userID)
	if uid == 0 {
		return nil, ErrUserNotFound
	}

	return userService.GetUserByID(uid)
}

// parseUserID 解析用户ID
func parseUserID(userID interface{}) uint {
	switch v := userID.(type) {
	case float64:
		return uint(v)
	case uint:
		return v
	case int:
		return uint(v)
	default:
		return 0
	}
}
