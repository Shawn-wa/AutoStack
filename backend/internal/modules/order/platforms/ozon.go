package platforms

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"autostack/internal/modules/order"
)

const (
	ozonAPIBaseURL = "https://api-seller.ozon.ru"
)

// OzonCredentials Ozon凭证结构
type OzonCredentials struct {
	ClientID string `json:"client_id"`
	APIKey   string `json:"api_key"`
}

// OzonAdapter Ozon平台适配器
type OzonAdapter struct{}

// GetPlatformName 获取平台标识
func (a *OzonAdapter) GetPlatformName() string {
	return order.PlatformOzon
}

// GetPlatformLabel 获取平台显示名称
func (a *OzonAdapter) GetPlatformLabel() string {
	return "Ozon"
}

// GetCredentialFields 获取凭证字段定义
func (a *OzonAdapter) GetCredentialFields() []order.CredentialField {
	return []order.CredentialField{
		{Key: "client_id", Label: "Client ID", Type: "text", Required: true},
		{Key: "api_key", Label: "API Key", Type: "password", Required: true},
	}
}

// parseCredentials 解析凭证
func (a *OzonAdapter) parseCredentials(credentials string) (*OzonCredentials, error) {
	var creds OzonCredentials
	if err := json.Unmarshal([]byte(credentials), &creds); err != nil {
		return nil, errors.New("凭证格式错误")
	}
	if creds.ClientID == "" || creds.APIKey == "" {
		return nil, errors.New("凭证不完整")
	}
	return &creds, nil
}

// TestConnection 测试连接
func (a *OzonAdapter) TestConnection(credentials string) error {
	creds, err := a.parseCredentials(credentials)
	if err != nil {
		return err
	}

	// 调用一个简单的 API 来验证凭证
	// 使用获取卖家信息的接口
	req, err := http.NewRequest("POST", ozonAPIBaseURL+"/v3/posting/fbs/list", bytes.NewBuffer([]byte(`{
		"filter": {
			"since": "`+time.Now().Add(-24*time.Hour).Format(time.RFC3339)+`",
			"to": "`+time.Now().Format(time.RFC3339)+`"
		},
		"limit": 1,
		"offset": 0
	}`)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Id", creds.ClientID)
	req.Header.Set("Api-Key", creds.APIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return errors.New("授权失败：Client ID 或 API Key 无效")
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API请求失败: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

// OzonOrderListRequest Ozon订单列表请求
type OzonOrderListRequest struct {
	Dir    string              `json:"dir"`
	Filter OzonOrderListFilter `json:"filter"`
	Limit  int                 `json:"limit"`
	Offset int                 `json:"offset"`
	With   OzonOrderListWith   `json:"with"`
}

// OzonOrderListFilter Ozon订单过滤条件
type OzonOrderListFilter struct {
	Since  string `json:"since"`
	To     string `json:"to"`
	Status string `json:"status,omitempty"`
}

// OzonOrderListWith Ozon订单附加数据
type OzonOrderListWith struct {
	AnalyticsData bool `json:"analytics_data"`
	FinancialData bool `json:"financial_data"`
}

// OzonOrderListResponse Ozon订单列表响应
type OzonOrderListResponse struct {
	Result struct {
		Postings []OzonPosting `json:"postings"`
	} `json:"result"`
}

// OzonPosting Ozon发货单
type OzonPosting struct {
	PostingNumber  string             `json:"posting_number"`
	OrderID        int64              `json:"order_id"`
	OrderNumber    string             `json:"order_number"`
	Status         string             `json:"status"`
	InProcessAt    string             `json:"in_process_at"`
	ShipmentDate   string             `json:"shipment_date"`
	DeliveringDate string             `json:"delivering_date"`
	Products       []OzonProduct      `json:"products"`
	Customer       *OzonCustomer      `json:"customer,omitempty"`
	AddressInfo    *OzonAddressInfo   `json:"addressee,omitempty"`
	FinancialData  *OzonFinancialData `json:"financial_data,omitempty"`
}

// OzonProduct Ozon商品
type OzonProduct struct {
	Sku          int64  `json:"sku"`
	Name         string `json:"name"`
	Quantity     int    `json:"quantity"`
	OfferID      string `json:"offer_id"`
	Price        string `json:"price"`
	CurrencyCode string `json:"currency_code"`
}

// OzonCustomer Ozon客户
type OzonCustomer struct {
	Name    string      `json:"name"`
	Phone   string      `json:"phone"`
	Email   string      `json:"email"`
	Address OzonAddress `json:"address"`
}

// OzonAddress Ozon地址
type OzonAddress struct {
	Country     string `json:"country"`
	Region      string `json:"region"`
	City        string `json:"city"`
	ZipCode     string `json:"zip_code"`
	AddressTail string `json:"address_tail"`
}

// OzonAddressInfo Ozon收件人信息
type OzonAddressInfo struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	City    string `json:"city"`
	Region  string `json:"region"`
	Country string `json:"country"`
}

// OzonFinancialData Ozon财务数据
type OzonFinancialData struct {
	Products []OzonFinancialProduct `json:"products"`
}

// OzonFinancialProduct Ozon财务商品
type OzonFinancialProduct struct {
	Price        float64 `json:"price"`
	CurrencyCode string  `json:"currency_code"`
}

// SyncOrders 同步订单
func (a *OzonAdapter) SyncOrders(credentials string, since, to time.Time) ([]*order.Order, error) {
	creds, err := a.parseCredentials(credentials)
	if err != nil {
		return nil, err
	}

	var allOrders []*order.Order
	offset := 0
	limit := 100

	for {
		reqBody := OzonOrderListRequest{
			Dir: "ASC",
			Filter: OzonOrderListFilter{
				Since: since.Format(time.RFC3339),
				To:    to.Format(time.RFC3339),
			},
			Limit:  limit,
			Offset: offset,
			With: OzonOrderListWith{
				AnalyticsData: true,
				FinancialData: true,
			},
		}

		bodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			return nil, err
		}

		req, err := http.NewRequest("POST", ozonAPIBaseURL+"/v3/posting/fbs/list", bytes.NewBuffer(bodyBytes))
		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Client-Id", creds.ClientID)
		req.Header.Set("Api-Key", creds.APIKey)

		client := &http.Client{Timeout: 60 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("请求失败: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			return nil, fmt.Errorf("读取响应失败: %w", err)
		}

		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("API请求失败: %d - %s", resp.StatusCode, string(body))
		}

		var listResp OzonOrderListResponse
		if err := json.Unmarshal(body, &listResp); err != nil {
			return nil, fmt.Errorf("解析响应失败: %w", err)
		}

		if len(listResp.Result.Postings) == 0 {
			break
		}

		for _, posting := range listResp.Result.Postings {
			ord := a.convertToOrder(posting, string(body))
			allOrders = append(allOrders, ord)
		}

		if len(listResp.Result.Postings) < limit {
			break
		}

		offset += limit
	}

	return allOrders, nil
}

// convertToOrder 转换为统一订单格式
func (a *OzonAdapter) convertToOrder(posting OzonPosting, rawData string) *order.Order {
	ord := &order.Order{
		Platform:        order.PlatformOzon,
		PlatformOrderNo: posting.PostingNumber,
		PlatformStatus:  posting.Status,
		Status:          a.mapStatus(posting.Status),
		RawData:         rawData,
	}

	// 解析订单时间
	if posting.InProcessAt != "" {
		if t, err := time.Parse(time.RFC3339, posting.InProcessAt); err == nil {
			ord.OrderTime = &t
		}
	}

	// 解析发货时间
	if posting.ShipmentDate != "" {
		if t, err := time.Parse(time.RFC3339, posting.ShipmentDate); err == nil {
			ord.ShipTime = &t
		}
	}

	// 收件人信息
	if posting.AddressInfo != nil {
		ord.RecipientName = posting.AddressInfo.Name
		ord.RecipientPhone = posting.AddressInfo.Phone
		ord.Country = posting.AddressInfo.Country
		ord.Province = posting.AddressInfo.Region
		ord.City = posting.AddressInfo.City
	}

	if posting.Customer != nil {
		if ord.RecipientName == "" {
			ord.RecipientName = posting.Customer.Name
		}
		if ord.RecipientPhone == "" {
			ord.RecipientPhone = posting.Customer.Phone
		}
		ord.Country = posting.Customer.Address.Country
		ord.Province = posting.Customer.Address.Region
		ord.City = posting.Customer.Address.City
		ord.ZipCode = posting.Customer.Address.ZipCode
		ord.Address = posting.Customer.Address.AddressTail
	}

	// 商品信息
	var totalAmount float64
	for _, prod := range posting.Products {
		item := order.OrderItem{
			PlatformSku: prod.OfferID,
			Sku:         prod.OfferID,
			Name:        prod.Name,
			Quantity:    prod.Quantity,
			Currency:    prod.CurrencyCode,
		}
		// 解析价格
		if prod.Price != "" {
			fmt.Sscanf(prod.Price, "%f", &item.Price)
		}
		totalAmount += item.Price * float64(item.Quantity)
		ord.Items = append(ord.Items, item)
		if ord.Currency == "" && prod.CurrencyCode != "" {
			ord.Currency = prod.CurrencyCode
		}
	}
	ord.TotalAmount = totalAmount

	return ord
}

// mapStatus 映射状态
func (a *OzonAdapter) mapStatus(platformStatus string) string {
	if status, ok := order.OzonStatusMap[platformStatus]; ok {
		return status
	}
	return order.OrderStatusPending
}

func init() {
	// 自动注册适配器
	order.RegisterAdapter(&OzonAdapter{})
}
