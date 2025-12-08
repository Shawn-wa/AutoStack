package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"autostack/pkg/response"
)

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"required,email"`
}

// Login 用户登录
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	// TODO: 实现实际的登录逻辑
	response.Success(c, http.StatusOK, "登录成功", gin.H{
		"token": "jwt-token-placeholder",
		"user": gin.H{
			"id":       1,
			"username": req.Username,
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

	// TODO: 实现实际的注册逻辑
	response.Success(c, http.StatusCreated, "注册成功", gin.H{
		"id":       1,
		"username": req.Username,
		"email":    req.Email,
	})
}

