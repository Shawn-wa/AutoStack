package ozon

import (
	"autostack/internal/modules/apiClient/platform"
	"encoding/json"
	"net/http"
)

// GetProductList 获取商品列表
// limit: 每页数量，默认 100，最大 1000
// lastID: 上一页最后一个商品的 ID，用于分页（第一页传空字符串）
func (c *Client) GetProductList(limit int, lastID string) (*ProductListResponse, error) {
	if limit <= 0 {
		limit = 100
	}

	reqBody := ProductListRequest{
		Limit:  limit,
		LastID: lastID,
		Filter: ProductListFilter{
			Visibility: "ALL",
		},
	}

	// API: POST /v3/product/list
	// 文档: https://docs.ozon.com/api/seller/en/#operation/ProductAPI_GetProductListV3
	resp, err := c.DoRequest(http.MethodPost, "/v3/product/list", reqBody, platform.RequestTypeProductList)
	if err != nil {
		return nil, err
	}

	var result ProductListResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetProductInfo 获取商品详细信息
// offerIds: 商品的 Offer ID 列表
func (c *Client) GetProductInfo(offerIds []string) (*ProductInfoResponse, error) {
	reqBody := ProductInfoRequest{
		OfferID: offerIds,
	}

	resp, err := c.DoRequest(http.MethodPost, "/v2/product/info/list", reqBody, platform.RequestTypeProductInfo)
	if err != nil {
		return nil, err
	}

	var result ProductInfoResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
