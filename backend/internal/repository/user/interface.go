package user

import (
	"context"
)

// UserRepository 用户仓储接口
type UserRepository interface {
	// FindByID 根据ID查找用户
	FindByID(ctx context.Context, id uint) (*User, error)
	// FindByUsername 根据用户名查找用户
	FindByUsername(ctx context.Context, username string) (*User, error)
	// FindByRole 根据角色查找第一个用户
	FindByRole(ctx context.Context, role string) (*User, error)
	// List 分页查询用户列表
	List(ctx context.Context, query *UserQuery) ([]User, int64, error)
	// ListAll 查询所有用户
	ListAll(ctx context.Context) ([]User, error)
	// Create 创建用户
	Create(ctx context.Context, user *User) error
	// Update 更新用户
	Update(ctx context.Context, user *User) error
	// UpdateFields 更新指定字段
	UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error
	// Delete 删除用户
	Delete(ctx context.Context, id uint) (int64, error)
	// CountByUsernameOrEmail 统计用户名或邮箱的数量
	CountByUsernameOrEmail(ctx context.Context, username, email string) (int64, error)
	// Count 统计用户总数
	Count(ctx context.Context) (int64, error)
}
