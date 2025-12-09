package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"autostack/pkg/response"
)

// Health 健康检查
func Health(c *gin.Context) {
	response.Success(c, http.StatusOK, "服务运行正常", gin.H{
		"status":  "healthy",
		"service": "AutoStack",
		"version": "1.0.0",
	})
}
