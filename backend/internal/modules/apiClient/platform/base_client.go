package platform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// BaseClient 基础客户端
// 包含公共的HTTP请求处理逻辑，各平台客户端通过组合方式继承
type BaseClient struct {
	httpClient *HTTPClient
	logger     *RequestLogger
}

// NewBaseClient 创建基础客户端
func NewBaseClient(baseURL string, logger *RequestLogger) *BaseClient {
	httpClient := NewHTTPClient(
		WithBaseURL(baseURL),
		WithHeader("Content-Type", "application/json"),
	)

	return &BaseClient{
		httpClient: httpClient,
		logger:     logger,
	}
}

// GetHTTPClient 获取HTTP客户端
func (c *BaseClient) GetHTTPClient() *HTTPClient {
	return c.httpClient
}

// GetLogger 获取日志记录器
func (c *BaseClient) GetLogger() *RequestLogger {
	return c.logger
}

// DoRequest 执行HTTP请求（公共逻辑）
// client: 实现PlatformClient接口的具体平台客户端
// method: HTTP方法
// path: API路径
// body: 请求体
// requestType: 请求类型（用于日志）
func (c *BaseClient) DoRequest(client PlatformClient, method, path string, body interface{}, requestType string) (*APIResponse, error) {
	var bodyBytes []byte
	var err error

	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
	}

	url := client.GetBaseURL() + path
	req, err := http.NewRequest(method, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置公共请求头
	req.Header.Set("Content-Type", "application/json")

	// 调用子类实现设置认证头
	client.SetAuthHeaders(req)

	// 执行请求
	startTime := time.Now()
	resp, err := c.httpClient.GetClient().Do(req)
	duration := time.Since(startTime).Milliseconds()

	// 准备日志
	logEntry := &RequestLog{
		RequestType:    requestType,
		RequestURL:     url,
		RequestMethod:  method,
		RequestHeaders: client.GetMaskedAuthInfo(),
		RequestBody:    string(bodyBytes),
		Duration:       duration,
	}

	if err != nil {
		logEntry.ErrorMessage = err.Error()
		logEntry.ResponseStatus = 0
		c.logRequest(logEntry)
		return &APIResponse{Error: err, Duration: duration}, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logEntry.ErrorMessage = err.Error()
		logEntry.ResponseStatus = resp.StatusCode
		c.logRequest(logEntry)
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	logEntry.ResponseStatus = resp.StatusCode
	logEntry.ResponseBody = string(respBody)

	if resp.StatusCode != http.StatusOK {
		logEntry.ErrorMessage = fmt.Sprintf("API请求失败: %d", resp.StatusCode)
		c.logRequest(logEntry)
		return &APIResponse{
			StatusCode: resp.StatusCode,
			Body:       respBody,
			Duration:   duration,
		}, fmt.Errorf("API请求失败: %d - %s", resp.StatusCode, string(respBody))
	}

	// 记录成功日志
	c.logRequest(logEntry)

	return &APIResponse{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Duration:   duration,
	}, nil
}

// logRequest 记录请求日志
func (c *BaseClient) logRequest(log *RequestLog) {
	if c.logger != nil {
		c.logger.LogRequest(log)
	}
}
