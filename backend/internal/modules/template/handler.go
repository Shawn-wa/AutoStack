package template

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"autostack/pkg/response"
)

// CreateTemplateRequest 创建模板请求
type CreateTemplateRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Type        string `json:"type" binding:"required"` // docker, k8s, compose
	Content     string `json:"content" binding:"required"`
}

// ListTemplates 模板列表
func ListTemplates(c *gin.Context) {
	// 预置的部署模板
	templates := []gin.H{
		{
			"id":          1,
			"name":        "Nginx 静态站点",
			"description": "快速部署静态网站",
			"type":        "docker",
			"icon":        "nginx",
		},
		{
			"id":          2,
			"name":        "Node.js 应用",
			"description": "部署 Node.js 应用程序",
			"type":        "docker",
			"icon":        "nodejs",
		},
		{
			"id":          3,
			"name":        "Go 微服务",
			"description": "部署 Go 语言微服务",
			"type":        "docker",
			"icon":        "go",
		},
		{
			"id":          4,
			"name":        "MySQL 数据库",
			"description": "部署 MySQL 数据库实例",
			"type":        "docker",
			"icon":        "mysql",
		},
		{
			"id":          5,
			"name":        "Redis 缓存",
			"description": "部署 Redis 缓存服务",
			"type":        "docker",
			"icon":        "redis",
		},
		{
			"id":          6,
			"name":        "全栈应用",
			"description": "前后端 + 数据库完整方案",
			"type":        "compose",
			"icon":        "stack",
		},
	}
	response.Success(c, http.StatusOK, "获取成功", gin.H{
		"list":  templates,
		"total": len(templates),
	})
}

// CreateTemplate 创建模板
func CreateTemplate(c *gin.Context) {
	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	// TODO: 实现实际的创建逻辑
	response.Success(c, http.StatusCreated, "创建成功", gin.H{
		"id":          100,
		"name":        req.Name,
		"description": req.Description,
		"type":        req.Type,
	})
}

// GetTemplate 获取模板详情
func GetTemplate(c *gin.Context) {
	id := c.Param("id")
	// TODO: 实现实际的查询逻辑
	response.Success(c, http.StatusOK, "获取成功", gin.H{
		"id":          id,
		"name":        "Nginx 静态站点",
		"description": "快速部署静态网站",
		"type":        "docker",
		"content":     "FROM nginx:alpine\nCOPY . /usr/share/nginx/html",
	})
}
