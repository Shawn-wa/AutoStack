package user

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"autostack/internal/repository"
	userRepo "autostack/internal/repository/user"
	"autostack/pkg/response"
)

// userService 用户服务实例
var userService *Service

// InitHandler 初始化 Handler，注入 Service 依赖
// 应在服务器启动时调用
func InitHandler(db *gorm.DB) {
	txManager := repository.NewTxManager(db)
	userService = NewService(
		txManager,
		userRepo.NewUserRepository(db),
	)
}

// GetService 获取服务实例（用于外部调用）
func GetService() *Service {
	return userService
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
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
		Status:      user.Status,
		Permissions: user.GetPermissions(),
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
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
		Status:      user.Status,
		Permissions: user.GetPermissions(),
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

// CreateUser 创建用户（管理员）
func CreateUser(c *gin.Context) {
	// 获取当前用户
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

	// 检查是否有权限创建该角色
	if !currentUser.CanManageRole(req.Role) {
		response.Error(c, http.StatusForbidden, "无权创建该角色用户")
		return
	}

	// 只有超级管理员可以创建管理员
	if req.Role == RoleAdmin && !currentUser.IsSuperAdmin() {
		response.Error(c, http.StatusForbidden, "只有超级管理员可以创建管理员")
		return
	}

	// 验证权限是否可被授予
	if len(req.Permissions) > 0 {
		if err := userService.ValidatePermissions(currentUser, req.Role, req.Permissions); err != nil {
			response.Error(c, http.StatusForbidden, "包含无法授予的权限")
			return
		}
	}

	// 创建用户
	createdBy := currentUser.ID
	user, err := userService.CreateUserWithPermissions(
		req.Username,
		req.Password,
		req.Email,
		req.Role,
		req.Permissions,
		&createdBy,
	)
	if err != nil {
		if err == ErrUserExists {
			response.Error(c, http.StatusConflict, "用户名或邮箱已存在")
			return
		}
		response.Error(c, http.StatusInternalServerError, "创建用户失败")
		return
	}

	response.Success(c, http.StatusCreated, "创建成功", UserDetailResponse{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
		Status:      user.Status,
		Permissions: user.GetPermissions(),
		CreatedBy:   user.CreatedBy,
		CreatedAt:   user.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   user.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

// ListUsers 获取用户列表（管理员）
func ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	users, total, err := userService.ListUsers(page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取用户列表失败")
		return
	}

	// 转换为列表项
	list := make([]UserListItem, len(users))
	for i, u := range users {
		list[i] = UserListItem{
			ID:          u.ID,
			Username:    u.Username,
			Email:       u.Email,
			Role:        u.Role,
			Status:      u.Status,
			Permissions: u.GetPermissions(),
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
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
		Status:      user.Status,
		Permissions: user.GetPermissions(),
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

	// 获取当前用户
	currentUser, err := getCurrentUser(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	// 获取目标用户
	targetUser, err := userService.GetUserByID(uint(id))
	if err != nil {
		if err == ErrUserNotFound {
			response.Error(c, http.StatusNotFound, "用户不存在")
			return
		}
		response.Error(c, http.StatusInternalServerError, "获取用户失败")
		return
	}

	// 不能修改自己（通过此接口）
	if currentUser.ID == targetUser.ID {
		response.Error(c, http.StatusBadRequest, "不能通过此接口修改自己")
		return
	}

	// 检查是否有权限管理目标用户
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

	// 处理角色变更
	newRole := targetUser.Role
	if req.Role != "" && req.Role != targetUser.Role {
		// 检查是否有权限变更角色
		if !currentUser.CanManageRole(req.Role) {
			response.Error(c, http.StatusForbidden, "无权设置该角色")
			return
		}
		// 只有超级管理员可以设置管理员角色
		if req.Role == RoleAdmin && !currentUser.IsSuperAdmin() {
			response.Error(c, http.StatusForbidden, "只有超级管理员可以设置管理员角色")
			return
		}
		// 不能修改超级管理员的角色
		if targetUser.IsSuperAdmin() {
			response.Error(c, http.StatusForbidden, "不能修改超级管理员的角色")
			return
		}
		updates["role"] = req.Role
		newRole = req.Role
	}

	// 处理权限变更
	if req.Permissions != nil {
		if err := userService.ValidatePermissions(currentUser, newRole, req.Permissions); err != nil {
			response.Error(c, http.StatusForbidden, "包含无法授予的权限")
			return
		}
		// 将权限序列化为JSON
		if err := targetUser.SetPermissions(req.Permissions); err != nil {
			response.Error(c, http.StatusInternalServerError, "设置权限失败")
			return
		}
		updates["permissions"] = targetUser.Permissions
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
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
		Status:      user.Status,
		Permissions: user.GetPermissions(),
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

	// 获取当前用户
	currentUser, err := getCurrentUser(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	// 不能删除自己
	if currentUser.ID == uint(id) {
		response.Error(c, http.StatusBadRequest, "不能删除自己")
		return
	}

	// 获取目标用户
	targetUser, err := userService.GetUserByID(uint(id))
	if err != nil {
		if err == ErrUserNotFound {
			response.Error(c, http.StatusNotFound, "用户不存在")
			return
		}
		response.Error(c, http.StatusInternalServerError, "获取用户失败")
		return
	}

	// 不能删除超级管理员
	if targetUser.IsSuperAdmin() {
		response.Error(c, http.StatusForbidden, "不能删除超级管理员")
		return
	}

	// 检查是否有权限删除目标用户
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
