package ozon

import (
	"encoding/json"
	"fmt"
	"time"

	"autostack/internal/modules/apiClient/platform"
)

// OrderAPI 订单API
type OrderAPI struct {
	client *Client
}

// NewOrderAPI 创建订单API
func NewOrderAPI(client *Client) *OrderAPI {
	return &OrderAPI{client: client}
}

// GetOrderList 获取订单列表
// API: POST /v3/posting/fbs/list
func (api *OrderAPI) GetOrderList(since, to time.Time, offset, limit int) (*OrderListResponse, error) {
	req := OrderListRequest{
		Dir: "ASC",
		Filter: OrderListFilter{
			Since: since.Format(time.RFC3339),
			To:    to.Format(time.RFC3339),
		},
		Limit:  limit,
		Offset: offset,
		With: OrderListWith{
			AnalyticsData: true,
			FinancialData: true,
		},
	}

	resp, err := api.client.DoRequest("POST", "/v3/posting/fbs/list", req, platform.RequestTypeOrderList)
	if err != nil {
		return nil, err
	}

	var result OrderListResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &result, nil
}

// TestConnection 测试连接
// 使用获取订单列表API验证凭证
func (api *OrderAPI) TestConnection() error {
	req := OrderListRequest{
		Dir: "ASC",
		Filter: OrderListFilter{
			Since: time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			To:    time.Now().Format(time.RFC3339),
		},
		Limit:  1,
		Offset: 0,
	}

	resp, err := api.client.DoRequest("POST", "/v3/posting/fbs/list", req, platform.RequestTypeTestConnect)
	if err != nil {
		// 检查是否是认证错误
		if resp != nil && (resp.StatusCode == 401 || resp.StatusCode == 403) {
			return fmt.Errorf("授权失败：Client ID 或 API Key 无效")
		}
		return err
	}

	return nil
}

