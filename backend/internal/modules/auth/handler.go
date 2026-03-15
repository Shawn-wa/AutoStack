package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"autostack/internal/modules/user"
	"autostack/pkg/response"
)

var authService *Service

// InitService 初始化认证服务
func InitService(jwtSecret string, jwtExpireHour int) {
	authService = NewService(jwtSecret, jwtExpireHour)
}

// getCompanyName 获取企业名称
func getCompanyName(companyID uint) string {
	if companyID == 0 {
		return ""
	}
	userSvc := user.GetService()
	if userSvc == nil {
		return ""
	}
	company, err := userSvc.GetCompanyByID(companyID)
	if err != nil {
		return ""
	}
	return company.Name
}

// Login 用户登录
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	u, token, err := authService.Login(req.Username, req.Password)
	if err != nil {
		switch err {
		case ErrUserNotFound:
			response.Error(c, http.StatusUnauthorized, "用户不存在")
		case ErrInvalidPassword:
			response.Error(c, http.StatusUnauthorized, "密码错误")
		case ErrUserDisabled:
			response.Error(c, http.StatusForbidden, "用户已被禁用")
		default:
			response.Error(c, http.StatusInternalServerError, "登录失败")
		}
		return
	}

	response.Success(c, http.StatusOK, "登录成功", LoginResponse{
		Token: token,
		User: UserInfo{
			ID:          u.ID,
			CompanyID:   u.CompanyID,
			CompanyName: getCompanyName(u.CompanyID),
			Username:    u.Username,
			Email:       u.Email,
			Role:        u.Role,
			Permissions: u.GetPermissions(),
		},
	})
}

// Register 用户注册
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	u, company, err := authService.Register(req.Username, req.Password, req.Email, req.CompanyName)
	if err != nil {
		if err == user.ErrUserExists {
			response.Error(c, http.StatusConflict, "用户名或邮箱已存在")
			return
		}
		response.Error(c, http.StatusInternalServerError, "注册失败")
		return
	}

	response.Success(c, http.StatusCreated, "注册成功", UserInfo{
		ID:          u.ID,
		CompanyID:   company.ID,
		CompanyName: company.Name,
		Username:    u.Username,
		Email:       u.Email,
		Role:        u.Role,
	})
}
