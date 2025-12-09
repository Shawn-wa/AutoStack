package user

import (
	"errors"

	"gorm.io/gorm"

	"autostack/internal/commonBase/database"
	"autostack/internal/utils"
)

var (
	ErrUserNotFound       = errors.New("用户不存在")
	ErrUserExists         = errors.New("用户名或邮箱已存在")
	ErrInvalidPassword    = errors.New("密码错误")
	ErrUserDisabled       = errors.New("用户已被禁用")
	ErrInvalidOldPassword = errors.New("原密码错误")
	ErrPermissionDenied   = errors.New("权限不足")
	ErrCannotModifySelf   = errors.New("不能修改自己")
)

// Service 用户服务
type Service struct{}

// NewService 创建用户服务实例
func NewService() *Service {
	return &Service{}
}

// CreateUser 创建用户（公开注册，只能创建普通用户）
func (s *Service) CreateUser(username, password, email, role string) (*User, error) {
	return s.CreateUserWithPermissions(username, password, email, role, nil, nil)
}

// CreateUserWithPermissions 创建用户（管理员创建，可指定角色和权限）
func (s *Service) CreateUserWithPermissions(username, password, email, role string, permissions []string, createdBy *uint) (*User, error) {
	db := database.GetDB()

	// 检查用户名或邮箱是否已存在
	var count int64
	db.Model(&User{}).Where("username = ? OR email = ?", username, email).Count(&count)
	if count > 0 {
		return nil, ErrUserExists
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &User{
		Username:     username,
		PasswordHash: hashedPassword,
		Email:        email,
		Role:         role,
		Status:       StatusActive,
		CreatedBy:    createdBy,
	}

	// 设置权限
	if permissions != nil && len(permissions) > 0 {
		if err := user.SetPermissions(permissions); err != nil {
			return nil, err
		}
	}

	if err := db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByID 根据ID获取用户
func (s *Service) GetUserByID(id uint) (*User, error) {
	db := database.GetDB()

	var user User
	if err := db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// GetUserByUsername 根据用户名获取用户
func (s *Service) GetUserByUsername(username string) (*User, error) {
	db := database.GetDB()

	var user User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// UpdateUser 更新用户信息
func (s *Service) UpdateUser(id uint, updates map[string]interface{}) (*User, error) {
	db := database.GetDB()

	var user User
	if err := db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if err := db.Model(&user).Updates(updates).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// ChangePassword 修改密码
func (s *Service) ChangePassword(id uint, oldPassword, newPassword string) error {
	db := database.GetDB()

	var user User
	if err := db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	// 验证旧密码
	if !utils.CheckPassword(oldPassword, user.PasswordHash) {
		return ErrInvalidOldPassword
	}

	// 加密新密码
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return db.Model(&user).Update("password_hash", hashedPassword).Error
}

// ListUsers 获取用户列表
func (s *Service) ListUsers(page, pageSize int) ([]User, int64, error) {
	db := database.GetDB()

	var users []User
	var total int64

	db.Model(&User{}).Count(&total)

	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Order("id DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// DeleteUser 删除用户
func (s *Service) DeleteUser(id uint) error {
	db := database.GetDB()

	result := db.Delete(&User{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// InitDefaultSuperAdmin 初始化默认超级管理员
func InitDefaultSuperAdmin() error {
	db := database.GetDB()

	// 检查是否已存在超级管理员
	var superAdmin User
	if err := db.Where("role = ?", RoleSuperAdmin).First(&superAdmin).Error; err == nil {
		// 已存在超级管理员，无需操作
		return nil
	}

	// 检查是否存在旧版本的 admin 用户，升级为 super_admin
	var oldAdmin User
	if err := db.Where("username = ?", "admin").First(&oldAdmin).Error; err == nil {
		// 升级为超级管理员
		return db.Model(&oldAdmin).Update("role", RoleSuperAdmin).Error
	}

	// 都不存在，创建新的超级管理员
	hashedPassword, err := utils.HashPassword("autoStack123")
	if err != nil {
		return err
	}

	superAdmin = User{
		Username:     "admin",
		PasswordHash: hashedPassword,
		Email:        "admin@autostack.local",
		Role:         RoleSuperAdmin,
		Status:       StatusActive,
	}
	return db.Create(&superAdmin).Error
}

// GetAllPermissions 获取所有权限定义
func (s *Service) GetAllPermissions() PermissionsResponse {
	modules := make(map[string][]PermissionDef)
	for _, p := range AllPermissions {
		modules[p.Module] = append(modules[p.Module], p)
	}
	return PermissionsResponse{
		Permissions: AllPermissions,
		Modules:     modules,
	}
}

// GetAssignablePermissions 获取当前用户可分配的权限
func (s *Service) GetAssignablePermissions(currentUser *User, targetRole string) []PermissionDef {
	var assignable []PermissionDef

	for _, p := range AllPermissions {
		// 用户管理权限只有超级管理员可以分配
		if isUserManagePermission(p.Code) {
			if currentUser.IsSuperAdmin() && targetRole == RoleAdmin {
				assignable = append(assignable, p)
			}
			continue
		}
		// 其他权限都可以分配
		assignable = append(assignable, p)
	}

	return assignable
}

// isUserManagePermission 判断是否为用户管理权限
func isUserManagePermission(perm string) bool {
	for _, p := range UserManagePermissions {
		if p == perm {
			return true
		}
	}
	return false
}

// ValidatePermissions 验证权限是否可被授予
func (s *Service) ValidatePermissions(currentUser *User, targetRole string, permissions []string) error {
	assignable := s.GetAssignablePermissions(currentUser, targetRole)
	assignableMap := make(map[string]bool)
	for _, p := range assignable {
		assignableMap[p.Code] = true
	}

	for _, perm := range permissions {
		if !assignableMap[perm] {
			return ErrPermissionDenied
		}
	}
	return nil
}

// CanManageUser 检查当前用户是否可以管理目标用户
func (s *Service) CanManageUser(currentUser *User, targetUser *User) bool {
	if currentUser.ID == targetUser.ID {
		return false // 不能管理自己
	}
	return currentUser.CanManageRole(targetUser.Role)
}
