package ozon

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"autostack/internal/modules/apiClient/platform"
)

const (
	// BaseURL OZON API基础URL
	BaseURL = "https://api-seller.ozon.ru"
)

// Client OZON API客户端
// 通过组合 BaseClient 继承公共请求处理逻辑
type Client struct {
	*platform.BaseClient // 组合基础客户端
	clientID             string
	apiKey               string
}

// NewClient 创建OZON客户端
func NewClient(creds *Credentials, logger *platform.RequestLogger) *Client {
	return &Client{
		BaseClient: platform.NewBaseClient(BaseURL, logger),
		clientID:   creds.ClientID,
		apiKey:     creds.APIKey,
	}
}

// ========== 实现 PlatformClient 接口 ==========

// SetAuthHeaders 设置OZON认证请求头
func (c *Client) SetAuthHeaders(req *http.Request) {
	req.Header.Set("Client-Id", c.clientID)
	req.Header.Set("Api-Key", c.apiKey)
}

// GetBaseURL 获取API基础URL
func (c *Client) GetBaseURL() string {
	return BaseURL
}

// GetMaskedAuthInfo 获取脱敏的认证信息
func (c *Client) GetMaskedAuthInfo() string {
	return fmt.Sprintf(`{"Client-Id": "%s", "Api-Key": "***"}`, c.clientID)
}

// ========== 便捷方法 ==========

// DoRequest 执行HTTP请求
// 封装调用，传入自身作为PlatformClient
func (c *Client) DoRequest(method, path string, body interface{}, requestType string) (*platform.APIResponse, error) {
	return c.BaseClient.DoRequest(c, method, path, body, requestType)
}

// GetClientID 获取Client ID
func (c *Client) GetClientID() string {
	return c.clientID
}

// ========== 凭证解析 ==========

// ParseCredentials 解析凭证JSON
func ParseCredentials(credentialsJSON string) (*Credentials, error) {
	var creds Credentials
	if err := json.Unmarshal([]byte(credentialsJSON), &creds); err != nil {
		return nil, errors.New("凭证格式错误")
	}
	if creds.ClientID == "" || creds.APIKey == "" {
		return nil, errors.New("凭证不完整")
	}
	if creds.SettlementCurrency == "" {
		creds.SettlementCurrency = "RUB"
	}
	return &creds, nil
}
