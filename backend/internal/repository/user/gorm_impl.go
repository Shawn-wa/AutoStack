package user

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"autostack/internal/repository"
)

// ========== UserRepository 实现 ==========

type gormUserRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓储实例
func NewUserRepository(db *gorm.DB) UserRepository {
	return &gormUserRepository{db: db}
}

func (r *gormUserRepository) getDB(ctx context.Context) *gorm.DB {
	return repository.GetDB(ctx, r.db)
}

func (r *gormUserRepository) FindByID(ctx context.Context, id uint) (*User, error) {
	var user User
	if err := r.getDB(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *gormUserRepository) FindByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	if err := r.getDB(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *gormUserRepository) FindByRole(ctx context.Context, role string) (*User, error) {
	var user User
	if err := r.getDB(ctx).Where("role = ?", role).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *gormUserRepository) List(ctx context.Context, query *UserQuery) ([]User, int64, error) {
	var users []User
	var total int64
	db := r.getDB(ctx)

	q := db.Model(&User{})
	if query.Role != "" {
		q = q.Where("role = ?", query.Role)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		q = q.Where("username LIKE ? OR email LIKE ?", like, like)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	if err := q.Offset(offset).Limit(query.PageSize).Order("id DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *gormUserRepository) ListAll(ctx context.Context) ([]User, error) {
	var users []User
	if err := r.getDB(ctx).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *gormUserRepository) Create(ctx context.Context, user *User) error {
	return r.getDB(ctx).Create(user).Error
}

func (r *gormUserRepository) Update(ctx context.Context, user *User) error {
	return r.getDB(ctx).Save(user).Error
}

func (r *gormUserRepository) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	return r.getDB(ctx).Model(&User{}).Where("id = ?", id).Updates(fields).Error
}

func (r *gormUserRepository) Delete(ctx context.Context, id uint) (int64, error) {
	result := r.getDB(ctx).Delete(&User{}, id)
	return result.RowsAffected, result.Error
}

func (r *gormUserRepository) CountByUsernameOrEmail(ctx context.Context, username, email string) (int64, error) {
	var count int64
	if err := r.getDB(ctx).Model(&User{}).Where("username = ? OR email = ?", username, email).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *gormUserRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.getDB(ctx).Model(&User{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// ========== 错误定义 ==========

var (
	ErrUserNotFound = errors.New("用户不存在")
	ErrUserExists   = errors.New("用户名或邮箱已存在")
)
