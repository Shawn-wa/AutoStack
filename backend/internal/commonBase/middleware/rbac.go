package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"autostack/pkg/response"
)

// 角色常量
const (
	RoleSuperAdmin = "super_admin"
	RoleAdmin      = "admin"
	RoleUser       = "user"
)

// RequireSuperAdmin 要求超级管理员权限的中间件
func RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			response.Error(c, http.StatusUnauthorized, "未授权")
			c.Abort()
			return
		}

		roleStr, ok := role.(string)
		if !ok || roleStr != RoleSuperAdmin {
			response.Error(c, http.StatusForbidden, "需要超级管理员权限")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin 要求管理员权限的中间件（包含超级管理员）
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			response.Error(c, http.StatusUnauthorized, "未授权")
			c.Abort()
			return
		}

		roleStr, ok := role.(string)
		if !ok || (roleStr != RoleAdmin && roleStr != RoleSuperAdmin) {
			response.Error(c, http.StatusForbidden, "需要管理员权限")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole 要求特定角色的中间件
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			response.Error(c, http.StatusUnauthorized, "未授权")
			c.Abort()
			return
		}

		roleStr, ok := role.(string)
		if !ok {
			response.Error(c, http.StatusInternalServerError, "角色类型错误")
			c.Abort()
			return
		}

		// 超级管理员拥有所有角色权限
		if roleStr == RoleSuperAdmin {
			c.Next()
			return
		}

		// 检查是否拥有允许的角色之一
		for _, allowedRole := range roles {
			if roleStr == allowedRole {
				c.Next()
				return
			}
		}

		response.Error(c, http.StatusForbidden, "权限不足")
		c.Abort()
	}
}

// RequirePermission 要求特定权限的中间件
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			response.Error(c, http.StatusUnauthorized, "未授权")
			c.Abort()
			return
		}

		roleStr, ok := role.(string)
		if !ok {
			response.Error(c, http.StatusInternalServerError, "角色类型错误")
			c.Abort()
			return
		}

		// 超级管理员拥有所有权限
		if roleStr == RoleSuperAdmin {
			c.Next()
			return
		}

		// 检查用户权限
		permsRaw, exists := c.Get("permissions")
		if !exists {
			response.Error(c, http.StatusForbidden, "权限不足")
			c.Abort()
			return
		}

		// 解析权限
		var permissions []string
		switch v := permsRaw.(type) {
		case string:
			if err := json.Unmarshal([]byte(v), &permissions); err != nil {
				response.Error(c, http.StatusInternalServerError, "权限解析错误")
				c.Abort()
				return
			}
		case []string:
			permissions = v
		case []interface{}:
			for _, p := range v {
				if ps, ok := p.(string); ok {
					permissions = append(permissions, ps)
				}
			}
		}

		// 检查是否有该权限
		for _, perm := range permissions {
			if perm == permission {
				c.Next()
				return
			}
		}

		response.Error(c, http.StatusForbidden, "权限不足")
		c.Abort()
	}
}

// RequireAnyPermission 要求任一权限的中间件
func RequireAnyPermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			response.Error(c, http.StatusUnauthorized, "未授权")
			c.Abort()
			return
		}

		roleStr, ok := role.(string)
		if !ok {
			response.Error(c, http.StatusInternalServerError, "角色类型错误")
			c.Abort()
			return
		}

		// 超级管理员拥有所有权限
		if roleStr == RoleSuperAdmin {
			c.Next()
			return
		}

		// 获取用户权限
		permsRaw, exists := c.Get("permissions")
		if !exists {
			response.Error(c, http.StatusForbidden, "权限不足")
			c.Abort()
			return
		}

		// 解析权限
		var userPerms []string
		switch v := permsRaw.(type) {
		case string:
			if err := json.Unmarshal([]byte(v), &userPerms); err != nil {
				response.Error(c, http.StatusInternalServerError, "权限解析错误")
				c.Abort()
				return
			}
		case []string:
			userPerms = v
		case []interface{}:
			for _, p := range v {
				if ps, ok := p.(string); ok {
					userPerms = append(userPerms, ps)
				}
			}
		}

		// 检查是否有任一权限
		permMap := make(map[string]bool)
		for _, perm := range userPerms {
			permMap[perm] = true
		}

		for _, required := range permissions {
			if permMap[required] {
				c.Next()
				return
			}
		}

		response.Error(c, http.StatusForbidden, "权限不足")
		c.Abort()
	}
}
