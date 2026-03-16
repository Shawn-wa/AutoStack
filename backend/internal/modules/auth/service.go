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

	u, err := userSvc.GetUserByUsername(username)
	if err != nil {
		if err == user.ErrUserNotFound {
			return nil, "", ErrUserNotFound
		}
		return nil, "", err
	}

	if !u.IsActive() {
		return nil, "", ErrUserDisabled
	}

	if !utils.CheckPassword(password, u.PasswordHash) {
		return nil, "", ErrInvalidPassword
	}

	token, err := utils.GenerateTokenWithPermissions(
		u.ID,
		u.CompanyID,
		u.Username,
		u.Role,
		userSvc.GetEffectivePermissions(u),
		s.jwtSecret,
		s.jwtExpireHour,
	)
	if err != nil {
		return nil, "", err
	}

	return u, token, nil
}

// Register 用户注册（创建企业 + 超级管理员）
func (s *Service) Register(username, password, email, companyName string) (*user.User, *user.Company, error) {
	return s.getUserService().RegisterWithCompany(username, password, email, companyName)
}
