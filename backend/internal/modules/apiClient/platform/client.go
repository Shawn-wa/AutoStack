package platform

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

// HTTPClient 通用HTTP客户端配置
type HTTPClient struct {
	client  *http.Client
	baseURL string
	headers map[string]string
}

// ClientOption 客户端配置选项
type ClientOption func(*HTTPClient)

// WithBaseURL 设置基础URL
func WithBaseURL(url string) ClientOption {
	return func(c *HTTPClient) {
		c.baseURL = url
	}
}

// WithHeader 添加请求头
func WithHeader(key, value string) ClientOption {
	return func(c *HTTPClient) {
		c.headers[key] = value
	}
}

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *HTTPClient) {
		c.client.Timeout = timeout
	}
}

// NewHTTPClient 创建新的HTTP客户端
func NewHTTPClient(opts ...ClientOption) *HTTPClient {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   15 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
	}

	c := &HTTPClient{
		client: &http.Client{
			Timeout:   60 * time.Second,
			Transport: transport,
		},
		headers: make(map[string]string),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// GetClient 获取底层HTTP客户端
func (c *HTTPClient) GetClient() *http.Client {
	return c.client
}

// GetBaseURL 获取基础URL
func (c *HTTPClient) GetBaseURL() string {
	return c.baseURL
}

// GetHeaders 获取默认请求头
func (c *HTTPClient) GetHeaders() map[string]string {
	return c.headers
}
