package ozon

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"autostack/internal/modules/apiClient/platform"
)

const orderListPath = "/v4/posting/fbs/list"

// OrderAPI 订单API
type OrderAPI struct {
	client *Client
}

// NewOrderAPI 创建订单API
func NewOrderAPI(client *Client) *OrderAPI {
	return &OrderAPI{client: client}
}

// GetOrderList 获取订单列表
func (api *OrderAPI) GetOrderList(since, to time.Time, cursor string, limit int) (*OrderListResponse, error) {
	req := newOrderListRequest(since, to, limit)
	req.Cursor = cursor
	req.With = OrderListWith{
		AnalyticsData: true,
		FinancialData: true,
	}

	resp, err := api.client.DoRequest(http.MethodPost, orderListPath, req, platform.RequestTypeOrderList)
	if err != nil {
		return nil, err
	}

	var result OrderListResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &result, nil
}

// GetOrderDetail 获取单个订单详情
// API: POST /v3/posting/fbs/get
// 文档: https://docs.ozon.ru/api/seller/zh/#operation/PostingAPI_GetFbsPostingV3
func (api *OrderAPI) GetOrderDetail(postingNumber string) (*OrderDetailResponse, error) {
	req := OrderDetailRequest{
		PostingNumber: postingNumber,
		With: OrderListWith{
			AnalyticsData: true,
			FinancialData: true,
		},
	}

	resp, err := api.client.DoRequest("POST", "/v3/posting/fbs/get", req, platform.RequestTypeOrderList)
	if err != nil {
		return nil, err
	}

	var result OrderDetailResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &result, nil
}

// TestConnection 测试连接
// 使用获取订单列表API验证凭证
func (api *OrderAPI) TestConnection() error {
	now := time.Now()
	req := newOrderListRequest(now.Add(-24*time.Hour), now, 1)
	resp, err := api.client.DoRequest(http.MethodPost, orderListPath, req, platform.RequestTypeTestConnect)
	if err != nil && resp != nil && (resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden) {
		return fmt.Errorf("授权失败：Client ID 或 API Key 无效")
	}
	return err
}

func newOrderListRequest(since, to time.Time, limit int) OrderListRequest {
	return OrderListRequest{
		SortDir: "ASC",
		Filter: OrderListFilter{
			Since: since.UTC().Format(time.RFC3339),
			To:    to.UTC().Format(time.RFC3339),
		},
		Limit: limit,
	}
}
