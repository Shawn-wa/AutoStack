package platform

import "net/http"

// PlatformClient 平台客户端接口
// 各平台客户端需实现此接口以支持公共请求处理逻辑
type PlatformClient interface {
	// SetAuthHeaders 设置认证请求头（由子类实现）
	SetAuthHeaders(req *http.Request)

	// GetBaseURL 获取API基础URL
	GetBaseURL() string

	// GetMaskedAuthInfo 获取脱敏的认证信息（用于日志记录）
	GetMaskedAuthInfo() string
}

// APIResponse API响应结构
type APIResponse struct {
	StatusCode int
	Body       []byte
	Duration   int64
	Error      error
}
