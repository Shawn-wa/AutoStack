package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	companyRepo "autostack/internal/repository/company"
	"autostack/internal/repository"
	userRepo "autostack/internal/repository/user"
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
	ErrCompanyNotFound    = errors.New("企业不存在")
)

// Service 用户服务
type Service struct {
	txManager   repository.TxManager
	userRepo    userRepo.UserRepository
	companyRepo companyRepo.CompanyRepository
}

// NewService 创建用户服务实例
func NewService(txManager repository.TxManager, userRepo userRepo.UserRepository, companyRepo companyRepo.CompanyRepository) *Service {
	return &Service{
		txManager:   txManager,
		userRepo:    userRepo,
		companyRepo: companyRepo,
	}
}

// CreateUser 创建用户（公开注册，只能创建普通用户）
func (s *Service) CreateUser(username, password, email, role string, companyID uint) (*User, error) {
	return s.CreateUserWithPermissions(username, password, email, role, nil, nil, companyID)
}

// CreateUserWithPermissions 创建用户（管理员创建，可指定角色和权限）
func (s *Service) CreateUserWithPermissions(username, password, email, role string, permissions []string, createdBy *uint, companyID uint) (*User, error) {
	ctx := context.Background()

	count, err := s.userRepo.CountByUsernameOrEmail(ctx, username, email)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrUserExists
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &User{
		CompanyID:    companyID,
		Username:     username,
		PasswordHash: hashedPassword,
		Email:        email,
		Role:         role,
		Status:       StatusActive,
		CreatedBy:    createdBy,
	}

	if len(permissions) > 0 {
		if err := user.SetPermissions(permissions); err != nil {
			return nil, err
		}
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// RegisterWithCompany 注册新用户并创建企业（公开注册入口）
func (s *Service) RegisterWithCompany(username, password, email, companyName string) (*User, *Company, error) {
	ctx := context.Background()

	count, err := s.userRepo.CountByUsernameOrEmail(ctx, username, email)
	if err != nil {
		return nil, nil, err
	}
	if count > 0 {
		return nil, nil, ErrUserExists
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, nil, err
	}

	var resultUser *User
	var resultCompany *Company

	err = s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		company := &Company{
			Name:   companyName,
			Status: companyRepo.StatusActive,
		}
		if err := s.companyRepo.Create(txCtx, company); err != nil {
			return err
		}

		user := &User{
			CompanyID:    company.ID,
			Username:     username,
			PasswordHash: hashedPassword,
			Email:        email,
			Role:         RoleSuperAdmin,
			Status:       StatusActive,
		}
		if err := s.userRepo.Create(txCtx, user); err != nil {
			return err
		}

		resultUser = user
		resultCompany = company
		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return resultUser, resultCompany, nil
}

// GetUserByID 根据ID获取用户
func (s *Service) GetUserByID(id uint) (*User, error) {
	ctx := context.Background()

	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

// GetUserByUsername 根据用户名获取用户
func (s *Service) GetUserByUsername(username string) (*User, error) {
	ctx := context.Background()

	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

// GetCompanyByID 根据ID获取企业
func (s *Service) GetCompanyByID(id uint) (*Company, error) {
	ctx := context.Background()

	company, err := s.companyRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCompanyNotFound
		}
		return nil, err
	}
	return company, nil
}

// UpdateUser 更新用户信息
func (s *Service) UpdateUser(id uint, updates map[string]interface{}) (*User, error) {
	ctx := context.Background()

	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if err := s.userRepo.UpdateFields(ctx, id, updates); err != nil {
		return nil, err
	}

	user, _ = s.userRepo.FindByID(ctx, id)
	return user, nil
}

// ChangePassword 修改密码
func (s *Service) ChangePassword(id uint, oldPassword, newPassword string) error {
	ctx := context.Background()

	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	if !utils.CheckPassword(oldPassword, user.PasswordHash) {
		return ErrInvalidOldPassword
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.userRepo.UpdateFields(ctx, id, map[string]interface{}{
		"password_hash": hashedPassword,
	})
}

// ListUsers 获取用户列表（同企业内）
func (s *Service) ListUsers(companyID uint, keyword, role string, page, pageSize int) ([]User, int64, error) {
	ctx := context.Background()

	return s.userRepo.List(ctx, &userRepo.UserQuery{
		Page:      page,
		PageSize:  pageSize,
		Keyword:   keyword,
		Role:      role,
		CompanyID: companyID,
	})
}

// DeleteUser 删除用户
func (s *Service) DeleteUser(id uint) error {
	ctx := context.Background()

	rowsAffected, err := s.userRepo.Delete(ctx, id)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// InitDefaultSuperAdmin 初始化默认超级管理员
func (s *Service) InitDefaultSuperAdmin() error {
	ctx := context.Background()

	// 兼容旧版 MySQL 表结构：历史 user_id 列如果仍是 NOT NULL，会阻塞 company_id 写入
	if err := s.ensureLegacySchemaCompatibility(ctx); err != nil {
		return err
	}
	// 历史 company_id 补齐：将旧短ID迁移为 90 开头的 11 位新ID
	if err := s.migrateLegacyCompanyIDs(ctx); err != nil {
		return err
	}

	// 检查是否已存在超级管理员
	existingAdmin, err := s.userRepo.FindByRole(ctx, RoleSuperAdmin)
	if err == nil && existingAdmin != nil {
		// 已存在超级管理员，确保其有 company
		if existingAdmin.CompanyID == 0 {
			return s.migrateExistingData(ctx, existingAdmin)
		}
		return nil
	}

	// 检查是否存在旧版本的 admin 用户
	oldAdmin, err := s.userRepo.FindByUsername(ctx, "admin")
	if err == nil {
		return s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
			company := &Company{
				Name:   "默认企业",
				Status: companyRepo.StatusActive,
			}
			if err := s.companyRepo.Create(txCtx, company); err != nil {
				return err
			}

			return s.userRepo.UpdateFields(txCtx, oldAdmin.ID, map[string]interface{}{
				"role":       RoleSuperAdmin,
				"company_id": company.ID,
			})
		})
	}

	// 都不存在，创建新的超级管理员和企业
	hashedPassword, err := utils.HashPassword("autoStack123")
	if err != nil {
		return err
	}

	return s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		company := &Company{
			Name:   "默认企业",
			Status: companyRepo.StatusActive,
		}
		if err := s.companyRepo.Create(txCtx, company); err != nil {
			return err
		}

		superAdmin := &User{
			CompanyID:    company.ID,
			Username:     "admin",
			PasswordHash: hashedPassword,
			Email:        "admin@autostack.local",
			Role:         RoleSuperAdmin,
			Status:       StatusActive,
		}
		return s.userRepo.Create(txCtx, superAdmin)
	})
}

// migrateExistingData 为已有的 super_admin 迁移数据（创建 company 并关联）
func (s *Service) migrateExistingData(ctx context.Context, admin *User) error {
	return s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		company := &Company{
			Name:   admin.Username + " 的企业",
			Status: companyRepo.StatusActive,
		}
		if err := s.companyRepo.Create(txCtx, company); err != nil {
			return err
		}

		// 更新 admin 的 company_id
		if err := s.userRepo.UpdateFields(txCtx, admin.ID, map[string]interface{}{
			"company_id": company.ID,
		}); err != nil {
			return err
		}

		// 将 admin 创建的所有子用户也关联到同一企业
		users, err := s.userRepo.ListAll(txCtx)
		if err != nil {
			return err
		}
		for _, u := range users {
			if u.ID != admin.ID && u.CreatedBy != nil && *u.CreatedBy == admin.ID && u.CompanyID == 0 {
				if err := s.userRepo.UpdateFields(txCtx, u.ID, map[string]interface{}{
					"company_id": company.ID,
				}); err != nil {
					return err
				}
			}
		}

		// 迁移业务数据：将 user_id 关联的数据改为 company_id
		db := repository.GetDB(txCtx, s.txManager.DB())
		db.Exec("UPDATE platform_auths SET company_id = ?, created_by = COALESCE(created_by, 0) WHERE company_id = 0")
		db.Exec("UPDATE orders SET company_id = ? WHERE company_id = 0", company.ID)
		db.Exec("UPDATE order_daily_stats SET company_id = ? WHERE company_id = 0", company.ID)
		db.Exec("UPDATE cash_flow_statements SET company_id = ? WHERE company_id = 0", company.ID)

		return nil
	})
}

// ensureLegacySchemaCompatibility 兼容旧版数据库中遗留的 user_id 列约束
func (s *Service) ensureLegacySchemaCompatibility(ctx context.Context) error {
	db := repository.GetDB(ctx, s.txManager.DB())
	if db == nil || db.Dialector.Name() != "mysql" {
		return nil
	}

	legacyTables := []string{
		"platform_auths",
		"orders",
		"order_daily_stats",
		"cash_flow_statements",
	}

	for _, tableName := range legacyTables {
		hasUserID, err := s.hasMySQLColumn(db, tableName, "user_id")
		if err != nil {
			return err
		}
		if !hasUserID {
			continue
		}

		// 保留历史列但放宽约束，避免新流程插入时因未传 user_id 报错
		if err := db.Exec("ALTER TABLE " + tableName + " MODIFY COLUMN user_id BIGINT UNSIGNED NULL DEFAULT NULL").Error; err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) hasMySQLColumn(db *gorm.DB, tableName, columnName string) (bool, error) {
	var count int64
	err := db.Raw(
		`SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?`,
		tableName,
		columnName,
	).Scan(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// migrateLegacyCompanyIDs 将历史 company_id 迁移为 90 开头 11 位，并同步更新关联表
func (s *Service) migrateLegacyCompanyIDs(ctx context.Context) error {
	return s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		db := repository.GetDB(txCtx, s.txManager.DB())
		if db == nil {
			return nil
		}

		var companies []Company
		if err := db.Order("id ASC").Find(&companies).Error; err != nil {
			return err
		}
		if len(companies) == 0 {
			return nil
		}

		var maxStandardID uint64
		if err := db.Model(&Company{}).
			Where("id >= ? AND id <= ?", companyRepo.CompanyIDMin, companyRepo.CompanyIDMax).
			Select("COALESCE(MAX(id), 0)").
			Scan(&maxStandardID).Error; err != nil {
			return err
		}
		if maxStandardID < uint64(companyRepo.CompanyIDMin-1) {
			maxStandardID = uint64(companyRepo.CompanyIDMin - 1)
		}

		for _, c := range companies {
			if companyRepo.IsStandardCompanyID(c.ID) {
				continue
			}

			maxStandardID++
			if maxStandardID > uint64(companyRepo.CompanyIDMax) {
				return fmt.Errorf("company_id 超出可分配范围，旧ID=%d", c.ID)
			}

			oldID := c.ID
			newID := uint(maxStandardID)
			if err := s.rebindCompanyID(txCtx, oldID, newID); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *Service) rebindCompanyID(ctx context.Context, oldID, newID uint) error {
	db := repository.GetDB(ctx, s.txManager.DB())

	// 先更新所有关联表，再更新 companies 主键
	updates := []struct {
		table  string
		column string
	}{
		{"users", "company_id"},
		{"platform_auths", "company_id"},
		{"orders", "company_id"},
		{"order_daily_stats", "company_id"},
		{"cash_flow_statements", "company_id"},
		{"platform_products", "platform_account_id"},
		{"product_mappings", "platform_account_id"},
	}

	for _, u := range updates {
		if err := db.Table(u.table).Where(u.column+" = ?", oldID).Update(u.column, newID).Error; err != nil {
			if isMissingTableOrColumnError(err) {
				continue
			}
			return err
		}
	}

	if err := db.Model(&Company{}).Where("id = ?", oldID).Update("id", newID).Error; err != nil {
		return err
	}
	return nil
}

func isMissingTableOrColumnError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "doesn't exist") ||
		strings.Contains(msg, "unknown column") ||
		strings.Contains(msg, "no such table") ||
		strings.Contains(msg, "no such column")
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
		if isUserManagePermission(p.Code) {
			if currentUser.IsSuperAdmin() && targetRole == RoleAdmin {
				assignable = append(assignable, p)
			}
			continue
		}
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
		return false
	}
	if currentUser.CompanyID != targetUser.CompanyID {
		return false
	}
	return currentUser.CanManageRole(targetUser.Role)
}

// ========== 包级函数（保持向后兼容） ==========

// InitDefaultSuperAdmin 初始化默认超级管理员（包级函数）
func InitDefaultSuperAdmin() error {
	if userService == nil {
		return errors.New("user service not initialized, call InitHandler first")
	}
	return userService.InitDefaultSuperAdmin()
}
