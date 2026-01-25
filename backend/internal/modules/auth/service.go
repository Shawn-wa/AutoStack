package auth

import (
	"errors"

	"autostack/internal/modules/user"
	"autostack/internal/utils"
)

var (
	ErrUserNotFound    = errors.New("用户不存在")
	ErrInvalidPassword = errors.New("密码错误")
	ErrUserDisabled    = errors.New("用户已被禁用")
)

// Service 认证服务
type Service struct {
	jwtSecret     string
	jwtExpireHour int
}

// NewService 创建认证服务实例
func NewService(jwtSecret string, jwtExpireHour int) *Service {
	return &Service{
		jwtSecret:     jwtSecret,
		jwtExpireHour: jwtExpireHour,
	}
}

// getUserService 获取用户服务
func (s *Service) getUserService() *user.Service {
	return user.GetService()
}

// Login 用户登录
func (s *Service) Login(username, password string) (*user.User, string, error) {
	userSvc := s.getUserService()

	// 获取用户
	u, err := userSvc.GetUserByUsername(username)
	if err != nil {
		if err == user.ErrUserNotFound {
			return nil, "", ErrUserNotFound
		}
		return nil, "", err
	}

	// 检查用户状态
	if !u.IsActive() {
		return nil, "", ErrUserDisabled
	}

	// 验证密码
	if !utils.CheckPassword(password, u.PasswordHash) {
		return nil, "", ErrInvalidPassword
	}

	// 生成 JWT（包含权限信息）
	token, err := utils.GenerateTokenWithPermissions(
		u.ID,
		u.Username,
		u.Role,
		u.GetPermissions(),
		s.jwtSecret,
		s.jwtExpireHour,
	)
	if err != nil {
		return nil, "", err
	}

	return u, token, nil
}

// Register 用户注册
func (s *Service) Register(username, password, email string) (*user.User, error) {
	return s.getUserService().CreateUser(username, password, email, user.RoleUser)
}
