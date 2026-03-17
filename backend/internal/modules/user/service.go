package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"autostack/internal/repository"
	companyRepo "autostack/internal/repository/company"
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
	ErrRoleExists         = errors.New("角色已存在")
	ErrRoleNameExists     = errors.New("角色名已存在")
	ErrRoleNotFound       = errors.New("角色不存在")
	ErrRoleDisabled       = errors.New("角色已禁用")
	ErrRoleInUse          = errors.New("角色已被用户使用")
)

// Service 用户服务
type Service struct {
	txManager   repository.TxManager
	userRepo    userRepo.UserRepository
	companyRepo companyRepo.CompanyRepository
}

type PermissionMigrationResult struct {
	CompaniesTotal     int
	CompaniesProcessed int
	RebuildRoleBinding bool
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

// CreateUserWithPermissions 创建用户（管理员创建，权限按角色模板继承）
func (s *Service) CreateUserWithPermissions(username, password, email, role string, permissions []string, createdBy *uint, companyID uint) (*User, error) {
	ctx := context.Background()
	role = strings.TrimSpace(role)
	if role == "" {
		return nil, ErrRoleNotFound
	}

	if role != RoleSuperAdmin {
		roleDef, err := s.GetRoleDefinitionByCode(companyID, role)
		if err != nil {
			return nil, err
		}
		if roleDef == nil {
			return nil, ErrRoleNotFound
		}
		if roleDef.Enabled != 1 {
			return nil, ErrRoleDisabled
		}
	}

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

	// 权限统一在角色上维护，用户级 permissions 字段保留但不再写入。
	_ = permissions

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

	// 初始化该企业的默认角色权限模板
	if err := s.EnsureDefaultRolePermissions(resultCompany.ID); err != nil {
		return nil, nil, err
	}
	if err := s.EnsureDefaultRoles(resultCompany.ID); err != nil {
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

// EnsureDefaultRolePermissions 确保企业具备默认角色权限模板
func (s *Service) EnsureDefaultRolePermissions(companyID uint) error {
	ctx := context.Background()
	db := repository.GetDB(ctx, s.txManager.DB())

	if err := s.EnsureDefaultRoles(companyID); err != nil {
		return err
	}

	if err := s.ensurePermissionCatalogInitialized(ctx); err != nil {
		return err
	}
	codeIDMap, err := s.getPermissionCodeIDMap(ctx)
	if err != nil {
		return err
	}

	roles, err := s.ListRoleCodes(companyID)
	if err != nil {
		return err
	}
	for _, role := range roles {
		var count int64
		if err := db.Model(&RolePermissionBinding{}).
			Where("company_id = ? AND role = ?", companyID, role).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}

		// 尝试从旧版 JSON 表迁移
		var legacy RolePermission
		var seedCodes []string
		if err := db.Where("company_id = ? AND role = ?", companyID, role).First(&legacy).Error; err == nil {
			seedCodes = legacy.GetPermissions()
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if len(seedCodes) == 0 {
			seedCodes = DefaultRolePermissionsByRole(role)
		}

		if err := s.replaceRoleBindings(ctx, companyID, role, seedCodes, codeIDMap); err != nil {
			return err
		}
	}
	return nil
}

// GetRolePermissions 获取企业角色权限（super_admin 固定全权限）
func (s *Service) GetRolePermissions(companyID uint) (map[string][]string, error) {
	if err := s.EnsureDefaultRolePermissions(companyID); err != nil {
		return nil, err
	}
	ctx := context.Background()
	allCodes, err := s.getAllActionPermissionCodes(ctx)
	if err != nil {
		return nil, err
	}

	result := map[string][]string{RoleSuperAdmin: allCodes}
	roles, err := s.ListRoleCodes(companyID)
	if err != nil {
		return nil, err
	}
	for _, role := range roles {
		if role == RoleSuperAdmin {
			continue
		}
		codes, err := s.getRolePermissionCodes(ctx, companyID, role)
		if err != nil {
			return nil, err
		}
		result[role] = codes
	}
	return result, nil
}

// GetEffectivePermissions 获取用户实际权限（按角色模板）
func (s *Service) GetEffectivePermissions(u *User) []string {
	if u == nil {
		return []string{}
	}
	if u.IsSuperAdmin() {
		return DefaultRolePermissionsByRole(RoleSuperAdmin)
	}
	rolePerms, err := s.GetRolePermissions(u.CompanyID)
	if err != nil {
		return []string{}
	}
	if perms, ok := rolePerms[u.Role]; ok {
		return perms
	}
	return []string{}
}

// UpdateRolePermissions 更新角色权限（仅角色维度管理）
func (s *Service) UpdateRolePermissions(currentUser *User, role string, permissions []string) error {
	role = strings.TrimSpace(role)
	if role == RoleSuperAdmin {
		return ErrPermissionDenied
	}
	if !s.CanManageRole(currentUser, role) {
		return ErrPermissionDenied
	}
	ok, err := s.RoleExists(currentUser.CompanyID, role)
	if err != nil {
		return err
	}
	if !ok {
		return ErrRoleNotFound
	}

	if err := s.ValidatePermissions(currentUser, role, permissions); err != nil {
		return err
	}
	ctx := context.Background()
	codeIDMap, err := s.getPermissionCodeIDMap(ctx)
	if err != nil {
		return err
	}
	return s.replaceRoleBindings(ctx, currentUser.CompanyID, role, permissions, codeIDMap)
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

// RunPermissionMigration 执行权限目录初始化迁移（支持线上 API 触发，幂等）
func (s *Service) RunPermissionMigration(rebuildRoleBindings bool) (PermissionMigrationResult, error) {
	ctx := context.Background()
	result := PermissionMigrationResult{
		RebuildRoleBinding: rebuildRoleBindings,
	}

	// 先强制同步目录定义（按当前代码定义 upsert）
	if err := s.ensurePermissionCatalog(ctx); err != nil {
		return result, err
	}
	// 再做层级/名称迁移和下线节点清理
	if err := s.migrateReportHierarchy(ctx); err != nil {
		return result, err
	}
	if err := s.migratePermissionDisplayNames(ctx); err != nil {
		return result, err
	}
	if err := s.pruneRemovedPermissionNodes(ctx); err != nil {
		return result, err
	}

	companies, err := s.companyRepo.List(ctx)
	if err != nil {
		return result, err
	}
	result.CompaniesTotal = len(companies)

	for _, c := range companies {
		if rebuildRoleBindings {
			if err := s.EnsureDefaultRolePermissions(c.ID); err != nil {
				return result, err
			}
		} else {
			if err := s.EnsureDefaultRoles(c.ID); err != nil {
				return result, err
			}
		}
		result.CompaniesProcessed++
	}

	return result, nil
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
		_ = s.EnsureDefaultRolePermissions(existingAdmin.CompanyID)
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
	ctx := context.Background()
	if err := s.ensurePermissionCatalogInitialized(ctx); err != nil {
		modules := make(map[string][]PermissionDef)
		for _, p := range AllPermissions {
			modules[p.Module] = append(modules[p.Module], p)
		}
		return PermissionsResponse{
			Permissions: AllPermissions,
			Modules:     modules,
		}
	}

	db := repository.GetDB(ctx, s.txManager.DB())
	var rows []Permission
	if err := db.Where("node_type = ? AND enabled = ?", "action", 1).Order("id ASC").Find(&rows).Error; err != nil {
		modules := make(map[string][]PermissionDef)
		for _, p := range AllPermissions {
			modules[p.Module] = append(modules[p.Module], p)
		}
		return PermissionsResponse{
			Permissions: AllPermissions,
			Modules:     modules,
		}
	}

	permissions := make([]PermissionDef, 0, len(rows))
	modules := make(map[string][]PermissionDef)

	for _, row := range rows {
		code := row.Code
		if code == "" {
			continue
		}
		routeKey := ""
		action := row.Action
		if strings.HasPrefix(code, "route:") {
			trimmed := strings.TrimPrefix(code, "route:")
			lastIdx := strings.LastIndex(trimmed, ":")
			if lastIdx > 0 {
				routeKey = trimmed[:lastIdx]
				if action == "" {
					action = trimmed[lastIdx+1:]
				}
			}
		}
		if routeKey == "" {
			continue
		}
		module := routeKey
		if dot := strings.Index(routeKey, "."); dot > 0 {
			module = routeKey[:dot]
		}

		def := PermissionDef{
			Code:        code,
			Name:        row.Name,
			Module:      module,
			RouteKey:    routeKey,
			ParentRoute: module,
			Action:      action,
		}
		permissions = append(permissions, def)
		modules[module] = append(modules[module], def)
	}

	return PermissionsResponse{
		Permissions: permissions,
		Modules:     modules,
	}
}

// GetPermissionRouteTree 获取权限路由树（数据库）
func (s *Service) GetPermissionRouteTree() ([]PermissionRouteNode, error) {
	ctx := context.Background()
	if err := s.ensurePermissionCatalogInitialized(ctx); err != nil {
		return nil, err
	}

	db := repository.GetDB(ctx, s.txManager.DB())
	var rows []Permission
	if err := db.Where("enabled = ?", 1).Order("sort ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}

	permByID := make(map[uint]Permission, len(rows))
	nodeByID := make(map[uint]*PermissionRouteNode)
	children := make(map[uint][]uint)
	var roots []uint

	for _, r := range rows {
		permByID[r.ID] = r
		if r.NodeType == "action" {
			continue
		}
		n := &PermissionRouteNode{
			Key:   r.NodeKey,
			Name:  r.Name,
			Level: 1,
		}
		nodeByID[r.ID] = n
	}

	for _, r := range rows {
		if r.NodeType == "action" {
			continue
		}
		if r.ParentID != nil {
			if _, ok := nodeByID[*r.ParentID]; ok {
				children[*r.ParentID] = append(children[*r.ParentID], r.ID)
				continue
			}
		}
		roots = append(roots, r.ID)
	}

	getModuleKey := func(parentID uint) string {
		curr := permByID[parentID]
		moduleKey := curr.NodeKey
		for curr.ParentID != nil {
			up, ok := permByID[*curr.ParentID]
			if !ok {
				break
			}
			moduleKey = up.NodeKey
			curr = up
		}
		return moduleKey
	}

	for _, r := range rows {
		if r.NodeType != "action" || r.ParentID == nil {
			continue
		}
		parentNode, ok := nodeByID[*r.ParentID]
		if !ok {
			continue
		}
		parentPerm := permByID[*r.ParentID]
		moduleKey := getModuleKey(*r.ParentID)
		parentNode.Permissions = append(parentNode.Permissions, PermissionDef{
			Code:        r.Code,
			Name:        r.Name,
			Module:      moduleKey,
			RouteKey:    parentPerm.NodeKey,
			ParentRoute: moduleKey,
			Action:      r.Action,
		})
	}

	var build func(id uint, level int) PermissionRouteNode
	build = func(id uint, level int) PermissionRouteNode {
		out := *nodeByID[id]
		out.Level = level
		childIDs := children[id]
		out.Children = make([]PermissionRouteNode, 0, len(childIDs))
		for _, cid := range childIDs {
			out.Children = append(out.Children, build(cid, level+1))
		}
		return out
	}

	result := make([]PermissionRouteNode, 0, len(roots))
	for _, rid := range roots {
		if _, ok := nodeByID[rid]; !ok {
			continue
		}
		result = append(result, build(rid, 1))
	}
	return result, nil
}

// FilterPermissionRouteTree 按权限过滤路由树（仅保留当前用户可见功能节点）
func (s *Service) FilterPermissionRouteTree(routeTree []PermissionRouteNode, permissionCodes []string) []PermissionRouteNode {
	permSet := make(map[string]struct{}, len(permissionCodes))
	for _, code := range permissionCodes {
		permSet[code] = struct{}{}
	}

	var filter func(nodes []PermissionRouteNode) []PermissionRouteNode
	filter = func(nodes []PermissionRouteNode) []PermissionRouteNode {
		result := make([]PermissionRouteNode, 0, len(nodes))
		for _, n := range nodes {
			filteredPerms := make([]PermissionDef, 0, len(n.Permissions))
			for _, p := range n.Permissions {
				if _, ok := permSet[p.Code]; ok {
					filteredPerms = append(filteredPerms, p)
				}
			}
			filteredChildren := filter(n.Children)
			if len(filteredPerms) == 0 && len(filteredChildren) == 0 {
				continue
			}
			n.Permissions = filteredPerms
			n.Children = filteredChildren
			result = append(result, n)
		}
		return result
	}

	return filter(routeTree)
}

// GetAssignablePermissions 获取当前用户可分配的权限
func (s *Service) GetAssignablePermissions(currentUser *User, targetRole string) []PermissionDef {
	var assignable []PermissionDef

	for _, p := range AllPermissions {
		if targetRole == RoleAdmin && !currentUser.IsSuperAdmin() {
			continue
		}
		assignable = append(assignable, p)
	}

	return assignable
}

// ValidatePermissions 验证权限是否可被授予
func (s *Service) ValidatePermissions(currentUser *User, targetRole string, permissions []string) error {
	if err := s.ensurePermissionCatalogInitialized(context.Background()); err != nil {
		return err
	}
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

// EnsureDefaultRoles 初始化企业默认角色定义
func (s *Service) EnsureDefaultRoles(companyID uint) error {
	ctx := context.Background()
	db := repository.GetDB(ctx, s.txManager.DB())
	defaults := []RoleDefinition{
		{CompanyID: companyID, Role: RoleAdmin, RoleName: "管理员", Desc: "管理员", Enabled: 1, IsSystem: 1},
		{CompanyID: companyID, Role: RoleUser, RoleName: "普通用户", Desc: "普通用户", Enabled: 1, IsSystem: 1},
	}
	for _, d := range defaults {
		var existing RoleDefinition
		err := db.Where("company_id = ? AND role = ?", companyID, d.Role).First(&existing).Error
		if err == nil {
			updates := map[string]interface{}{
				"enabled": d.Enabled,
			}
			if strings.TrimSpace(existing.RoleName) == "" {
				updates["role_name"] = d.RoleName
			}
			if strings.TrimSpace(existing.Desc) == "" {
				updates["desc"] = d.Desc
			}
			if len(updates) > 0 {
				_ = db.Model(&RoleDefinition{}).Where("id = ?", existing.ID).Updates(updates).Error
			}
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := db.Create(&d).Error; err != nil {
			return err
		}
	}
	return nil
}

// ListRoleCodes 返回企业可分配角色编码（不含 super_admin）
func (s *Service) ListRoleCodes(companyID uint) ([]string, error) {
	ctx := context.Background()
	if err := s.EnsureDefaultRoles(companyID); err != nil {
		return nil, err
	}
	db := repository.GetDB(ctx, s.txManager.DB())
	var rows []RoleDefinition
	if err := db.Where("company_id = ?", companyID).Order("is_system DESC, role ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]string, 0, len(rows))
	for _, r := range rows {
		result = append(result, r.Role)
	}
	return result, nil
}

func (s *Service) RoleExists(companyID uint, role string) (bool, error) {
	ctx := context.Background()
	db := repository.GetDB(ctx, s.txManager.DB())
	var count int64
	if err := db.Model(&RoleDefinition{}).Where("company_id = ? AND role = ?", companyID, role).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *Service) RoleNameExists(companyID uint, roleName string, excludeID uint) (bool, error) {
	ctx := context.Background()
	db := repository.GetDB(ctx, s.txManager.DB())
	query := db.Model(&RoleDefinition{}).Where("company_id = ? AND role_name = ?", companyID, roleName)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *Service) GetRoleDefinitionByCode(companyID uint, role string) (*RoleDefinition, error) {
	ctx := context.Background()
	db := repository.GetDB(ctx, s.txManager.DB())
	var item RoleDefinition
	err := db.Where("company_id = ? AND role = ?", companyID, role).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// CanManageRole 是否可管理目标角色
func (s *Service) CanManageRole(currentUser *User, targetRole string) bool {
	if currentUser == nil {
		return false
	}
	if currentUser.IsSuperAdmin() {
		return targetRole != RoleSuperAdmin
	}
	if currentUser.Role == RoleAdmin {
		if targetRole == RoleSuperAdmin || targetRole == RoleAdmin {
			return false
		}
		return true
	}
	return false
}

// ListRoles 角色列表（含备注和权限数）
func (s *Service) ListRoles(companyID uint) ([]RoleItem, error) {
	ctx := context.Background()
	if err := s.EnsureDefaultRoles(companyID); err != nil {
		return nil, err
	}
	rolePerms, err := s.GetRolePermissions(companyID)
	if err != nil {
		return nil, err
	}
	db := repository.GetDB(ctx, s.txManager.DB())
	var rows []RoleDefinition
	if err := db.Where("company_id = ?", companyID).Order("is_system DESC, role ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]RoleItem, 0, len(rows))
	for _, r := range rows {
		perms := rolePerms[r.Role]
		result = append(result, RoleItem{
			ID:              r.ID,
			Role:            r.Role,
			RoleName:        r.RoleName,
			Description:     r.Desc,
			Enabled:         r.Enabled == 1,
			IsSystem:        r.IsSystem == 1,
			PermissionCount: len(perms),
		})
	}
	return result, nil
}

// CreateRole 创建自定义角色
func (s *Service) CreateRole(currentUser *User, role, roleName, description string, enabled int) error {
	if currentUser == nil || !currentUser.IsAdmin() {
		return ErrPermissionDenied
	}
	role = strings.ToLower(strings.TrimSpace(role))
	roleName = strings.TrimSpace(roleName)
	if role == "" || roleName == "" || role == RoleSuperAdmin {
		return ErrPermissionDenied
	}
	ok, err := s.RoleExists(currentUser.CompanyID, role)
	if err != nil {
		return err
	}
	if ok {
		return ErrRoleExists
	}
	nameExists, err := s.RoleNameExists(currentUser.CompanyID, roleName, 0)
	if err != nil {
		return err
	}
	if nameExists {
		return ErrRoleNameExists
	}
	if enabled != 0 && enabled != 1 {
		enabled = 1
	}

	ctx := context.Background()
	db := repository.GetDB(ctx, s.txManager.DB())
	roleDef := RoleDefinition{
		CompanyID: currentUser.CompanyID,
		Role:      role,
		RoleName:  roleName,
		Desc:      strings.TrimSpace(description),
		Enabled:   enabled,
		IsSystem:  0,
	}
	if err := db.Create(&roleDef).Error; err != nil {
		return err
	}

	// 新角色默认复用普通用户权限
	codeIDMap, err := s.getPermissionCodeIDMap(ctx)
	if err != nil {
		return err
	}
	return s.replaceRoleBindings(ctx, currentUser.CompanyID, role, DefaultRolePermissionsByRole(RoleUser), codeIDMap)
}

func (s *Service) UpdateRole(currentUser *User, id uint, role, roleName, description string, enabled int) error {
	if currentUser == nil || !currentUser.IsAdmin() {
		return ErrPermissionDenied
	}
	role = strings.ToLower(strings.TrimSpace(role))
	roleName = strings.TrimSpace(roleName)
	if role == "" || roleName == "" || role == RoleSuperAdmin {
		return ErrPermissionDenied
	}

	ctx := context.Background()
	db := repository.GetDB(ctx, s.txManager.DB())
	var existing RoleDefinition
	if err := db.Where("id = ? AND company_id = ?", id, currentUser.CompanyID).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}
		return err
	}
	if !s.CanManageRole(currentUser, existing.Role) {
		return ErrPermissionDenied
	}
	if existing.Role != role && !s.CanManageRole(currentUser, role) {
		return ErrPermissionDenied
	}

	codeExists, err := s.RoleExists(currentUser.CompanyID, role)
	if err != nil {
		return err
	}
	if codeExists && existing.Role != role {
		return ErrRoleExists
	}
	nameExists, err := s.RoleNameExists(currentUser.CompanyID, roleName, existing.ID)
	if err != nil {
		return err
	}
	if nameExists {
		return ErrRoleNameExists
	}
	if enabled != 0 && enabled != 1 {
		enabled = existing.Enabled
	}

	return s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		tx := repository.GetDB(txCtx, s.txManager.DB())
		if err := tx.Model(&RoleDefinition{}).Where("id = ?", existing.ID).Updates(map[string]interface{}{
			"role":      role,
			"role_name": roleName,
			"desc":      strings.TrimSpace(description),
			"enabled":   enabled,
		}).Error; err != nil {
			return err
		}
		if existing.Role != role {
			if err := tx.Model(&RolePermissionBinding{}).Where("company_id = ? AND role = ?", currentUser.CompanyID, existing.Role).Update("role", role).Error; err != nil {
				return err
			}
			if err := tx.Model(&User{}).Where("company_id = ? AND role = ?", currentUser.CompanyID, existing.Role).Update("role", role).Error; err != nil {
				return err
			}
			if err := tx.Model(&RolePermission{}).Where("company_id = ? AND role = ?", currentUser.CompanyID, existing.Role).Update("role", role).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Service) DeleteRole(currentUser *User, id uint) error {
	if currentUser == nil || !currentUser.IsAdmin() {
		return ErrPermissionDenied
	}
	ctx := context.Background()
	db := repository.GetDB(ctx, s.txManager.DB())
	var existing RoleDefinition
	if err := db.Where("id = ? AND company_id = ?", id, currentUser.CompanyID).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}
		return err
	}
	if !s.CanManageRole(currentUser, existing.Role) {
		return ErrPermissionDenied
	}

	var userCount int64
	if err := db.Model(&User{}).Where("company_id = ? AND role = ?", currentUser.CompanyID, existing.Role).Count(&userCount).Error; err != nil {
		return err
	}
	if userCount > 0 {
		return ErrRoleInUse
	}

	return s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		tx := repository.GetDB(txCtx, s.txManager.DB())
		if err := tx.Where("company_id = ? AND role = ?", currentUser.CompanyID, existing.Role).Delete(&RolePermissionBinding{}).Error; err != nil {
			return err
		}
		if err := tx.Where("company_id = ? AND role = ?", currentUser.CompanyID, existing.Role).Delete(&RolePermission{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&RoleDefinition{}, existing.ID).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *Service) ensurePermissionCatalog(ctx context.Context) error {
	return s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		tx := repository.GetDB(txCtx, s.txManager.DB())
		var upsertNode func(parentID *uint, node PermissionRouteNode, order int) (uint, error)
		upsertNode = func(parentID *uint, node PermissionRouteNode, order int) (uint, error) {
			nodeType := "route"
			if parentID == nil {
				nodeType = "module"
			}

			var item Permission
			err := tx.Where("node_key = ?", node.Key).First(&item).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				item = Permission{
					ParentID: parentID,
					NodeKey:  node.Key,
					Name:     node.Name,
					NodeType: nodeType,
					Sort:     order,
					Enabled:  1,
				}
				if err := tx.Create(&item).Error; err != nil {
					return 0, err
				}
			} else if err != nil {
				return 0, err
			} else {
				if err := tx.Model(&Permission{}).Where("id = ?", item.ID).Updates(map[string]interface{}{
					"parent_id":  parentID,
					"name":       node.Name,
					"node_type":  nodeType,
					"sort":       order,
					"enabled":    1,
					"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
				}).Error; err != nil {
					return 0, err
				}
			}

			for idx, p := range node.Permissions {
				var actionNode Permission
				err := tx.Where("node_key = ?", p.Code).First(&actionNode).Error
				if errors.Is(err, gorm.ErrRecordNotFound) {
					actionNode = Permission{
						ParentID: &item.ID,
						NodeKey:  p.Code,
						Code:     p.Code,
						Name:     p.Name,
						NodeType: "action",
						Action:   p.Action,
						Sort:     idx,
						Enabled:  1,
					}
					if err := tx.Create(&actionNode).Error; err != nil {
						return 0, err
					}
				} else if err != nil {
					return 0, err
				} else {
					if err := tx.Model(&Permission{}).Where("id = ?", actionNode.ID).Updates(map[string]interface{}{
						"parent_id":  item.ID,
						"code":       p.Code,
						"name":       p.Name,
						"node_type":  "action",
						"action":     p.Action,
						"sort":       idx,
						"enabled":    1,
						"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
					}).Error; err != nil {
						return 0, err
					}
				}
			}

			for idx, child := range node.Children {
				if _, err := upsertNode(&item.ID, child, idx); err != nil {
					return 0, err
				}
			}

			return item.ID, nil
		}

		for idx, root := range PermissionRouteTree {
			if _, err := upsertNode(nil, root, idx); err != nil {
				return err
			}
		}
		return nil
	})
}

// ensurePermissionCatalogInitialized 仅在权限目录为空时做一次初始化，避免页面读取触发回写
func (s *Service) ensurePermissionCatalogInitialized(ctx context.Context) error {
	db := repository.GetDB(ctx, s.txManager.DB())
	var count int64
	if err := db.Model(&Permission{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		if err := s.migrateReportHierarchy(ctx); err != nil {
			return err
		}
		if err := s.migratePermissionDisplayNames(ctx); err != nil {
			return err
		}
		return s.pruneRemovedPermissionNodes(ctx)
	}
	if err := s.ensurePermissionCatalog(ctx); err != nil {
		return err
	}
	if err := s.migrateReportHierarchy(ctx); err != nil {
		return err
	}
	if err := s.migratePermissionDisplayNames(ctx); err != nil {
		return err
	}
	return s.pruneRemovedPermissionNodes(ctx)
}

// migrateReportHierarchy 将财务/结算节点迁移到报表一级目录（幂等）
func (s *Service) migrateReportHierarchy(ctx context.Context) error {
	return s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		tx := repository.GetDB(txCtx, s.txManager.DB())

		var report Permission
		err := tx.Where("node_key = ? AND node_type = ?", userRepo.RouteReport, "module").First(&report).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			var maxSort int
			_ = tx.Model(&Permission{}).Where("node_type = ? AND parent_id IS NULL", "module").Select("COALESCE(MAX(sort), 0)").Scan(&maxSort).Error
			report = Permission{
				NodeKey:  userRepo.RouteReport,
				Name:     "报表",
				NodeType: "module",
				Sort:     maxSort + 1,
				Enabled:  1,
			}
			if err := tx.Create(&report).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		routeKeys := []string{userRepo.RouteOrderCashFlow, userRepo.RouteOrderSettlement}
		if err := tx.Model(&Permission{}).
			Where("node_key IN ? AND node_type = ?", routeKeys, "route").
			Updates(map[string]interface{}{
				"parent_id":  report.ID,
				"node_type":  "route",
				"enabled":    1,
				"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
			}).Error; err != nil {
			return err
		}
		return nil
	})
}

// migratePermissionDisplayNames 同步目录展示名称到数据库（幂等）
func (s *Service) migratePermissionDisplayNames(ctx context.Context) error {
	db := repository.GetDB(ctx, s.txManager.DB())
	updates := map[string]string{
		userRepo.RouteWarehouse: "仓库",
		userRepo.RouteShipping:  "物流",
	}
	for key, name := range updates {
		if err := db.Model(&Permission{}).
			Where("node_key = ? AND node_type = ?", key, "module").
			Updates(map[string]interface{}{
				"name":       name,
				"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
			}).Error; err != nil {
			return err
		}
	}
	return nil
}

// pruneRemovedPermissionNodes 删除已下线功能对应的历史权限节点及绑定
func (s *Service) pruneRemovedPermissionNodes(ctx context.Context) error {
	db := repository.GetDB(ctx, s.txManager.DB())
	var rows []Permission
	if err := db.Find(&rows).Error; err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}

	removedPrefixes := []string{
		"route:system.projects:",
		"route:system.deployments:",
		"route:system.templates:",
	}
	removedRouteKeys := map[string]struct{}{
		"system.projects":    {},
		"system.deployments": {},
		"system.templates":   {},
	}

	children := make(map[uint][]uint, len(rows))
	for _, r := range rows {
		if r.ParentID != nil {
			children[*r.ParentID] = append(children[*r.ParentID], r.ID)
		}
	}

	toDelete := make(map[uint]struct{})
	var mark func(id uint)
	mark = func(id uint) {
		if _, ok := toDelete[id]; ok {
			return
		}
		toDelete[id] = struct{}{}
		for _, cid := range children[id] {
			mark(cid)
		}
	}

	for _, r := range rows {
		if _, ok := removedRouteKeys[r.NodeKey]; ok {
			mark(r.ID)
			continue
		}
		for _, p := range removedPrefixes {
			if strings.HasPrefix(r.NodeKey, p) || strings.HasPrefix(r.Code, p) {
				mark(r.ID)
				break
			}
		}
	}

	if len(toDelete) == 0 {
		return nil
	}
	ids := make([]uint, 0, len(toDelete))
	for id := range toDelete {
		ids = append(ids, id)
	}

	return s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		tx := repository.GetDB(txCtx, s.txManager.DB())
		if err := tx.Where("permission_id IN ?", ids).Delete(&RolePermissionBinding{}).Error; err != nil {
			return err
		}
		if err := tx.Where("id IN ?", ids).Delete(&Permission{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *Service) getPermissionCodeIDMap(ctx context.Context) (map[string]uint, error) {
	db := repository.GetDB(ctx, s.txManager.DB())
	var rows []Permission
	if err := db.Where("node_type = ? AND enabled = ?", "action", 1).Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[string]uint, len(rows))
	for _, r := range rows {
		if r.Code != "" {
			result[r.Code] = r.ID
		}
	}
	return result, nil
}

func (s *Service) getAllActionPermissionCodes(ctx context.Context) ([]string, error) {
	db := repository.GetDB(ctx, s.txManager.DB())
	var rows []Permission
	if err := db.Where("node_type = ? AND enabled = ?", "action", 1).Order("id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	codes := make([]string, 0, len(rows))
	for _, r := range rows {
		if r.Code != "" {
			codes = append(codes, r.Code)
		}
	}
	return codes, nil
}

func (s *Service) getRolePermissionCodes(ctx context.Context, companyID uint, role string) ([]string, error) {
	db := repository.GetDB(ctx, s.txManager.DB())
	var codes []string
	if err := db.Table("role_permission_bindings AS b").
		Select("p.code").
		Joins("JOIN permissions p ON p.id = b.permission_id").
		Where("b.company_id = ? AND b.role = ? AND p.enabled = ?", companyID, role, 1).
		Order("p.id ASC").
		Pluck("p.code", &codes).Error; err != nil {
		return nil, err
	}
	return codes, nil
}

func (s *Service) replaceRoleBindings(ctx context.Context, companyID uint, role string, codes []string, codeIDMap map[string]uint) error {
	return s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		tx := repository.GetDB(txCtx, s.txManager.DB())
		if err := tx.Where("company_id = ? AND role = ?", companyID, role).Delete(&RolePermissionBinding{}).Error; err != nil {
			return err
		}

		seen := make(map[string]struct{})
		items := make([]RolePermissionBinding, 0, len(codes))
		for _, code := range codes {
			if _, ok := seen[code]; ok {
				continue
			}
			seen[code] = struct{}{}
			id, ok := codeIDMap[code]
			if !ok {
				continue
			}
			items = append(items, RolePermissionBinding{
				CompanyID:    companyID,
				Role:         role,
				PermissionID: id,
			})
		}
		if len(items) == 0 {
			return nil
		}
		return tx.Create(&items).Error
	})
}

// CanManageUser 检查当前用户是否可以管理目标用户
func (s *Service) CanManageUser(currentUser *User, targetUser *User) bool {
	if currentUser.ID == targetUser.ID {
		return false
	}
	if currentUser.CompanyID != targetUser.CompanyID {
		return false
	}
	return s.CanManageRole(currentUser, targetUser.Role)
}

// ========== 包级函数（保持向后兼容） ==========

// InitDefaultSuperAdmin 初始化默认超级管理员（包级函数）
func InitDefaultSuperAdmin() error {
	if userService == nil {
		return errors.New("user service not initialized, call InitHandler first")
	}
	return userService.InitDefaultSuperAdmin()
}
