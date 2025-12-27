package platforms

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"autostack/internal/modules/order"
)

// 创建自定义HTTP客户端，用于访问外部API
func newHTTPClient() *http.Client {
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
	return &http.Client{
		Timeout:   60 * time.Second,
		Transport: transport,
	}
}

const (
	ozonAPIBaseURL = "https://api-seller.ozon.ru"
)

// OzonCredentials Ozon凭证结构
type OzonCredentials struct {
	ClientID           string `json:"client_id"`
	APIKey             string `json:"api_key"`
	SettlementCurrency string `json:"settlement_currency"` // 结算货币，默认CNY
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
		{Key: "settlement_currency", Label: "结算货币", Type: "text", Required: false},
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
	return a.TestConnectionWithLog(credentials, 0)
}

// TestConnectionWithLog 测试连接（带日志记录）
func (a *OzonAdapter) TestConnectionWithLog(credentials string, platformAuthID uint) error {
	creds, err := a.parseCredentials(credentials)
	if err != nil {
		return err
	}

	// 调用一个简单的 API 来验证凭证
	requestURL := ozonAPIBaseURL + "/v3/posting/fbs/list"
	requestBody := `{
		"filter": {
			"since": "` + time.Now().Add(-24*time.Hour).Format(time.RFC3339) + `",
			"to": "` + time.Now().Format(time.RFC3339) + `"
		},
		"limit": 1,
		"offset": 0
	}`

	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Id", creds.ClientID)
	req.Header.Set("Api-Key", creds.APIKey)

	startTime := time.Now()
	client := newHTTPClient()
	resp, err := client.Do(req)
	duration := time.Since(startTime).Milliseconds()

	// 记录请求日志
	logEntry := &order.OrdersRequestLog{
		PlatformAuthID: platformAuthID,
		Platform:       order.PlatformOzon,
		RequestType:    order.RequestTypeTestConnect,
		RequestURL:     requestURL,
		RequestMethod:  "POST",
		RequestHeaders: a.maskHeaders(creds.ClientID),
		RequestBody:    requestBody,
		Duration:       duration,
		CreatedAt:      time.Now(),
	}

	if err != nil {
		logEntry.ErrorMessage = err.Error()
		logEntry.ResponseStatus = 0
		order.SaveRequestLog(logEntry)
		return fmt.Errorf("连接失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	logEntry.ResponseStatus = resp.StatusCode
	logEntry.ResponseBody = string(body)

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		logEntry.ErrorMessage = "授权失败"
		order.SaveRequestLog(logEntry)
		return errors.New("授权失败：Client ID 或 API Key 无效")
	}

	if resp.StatusCode != 200 {
		logEntry.ErrorMessage = fmt.Sprintf("API请求失败: %d", resp.StatusCode)
		order.SaveRequestLog(logEntry)
		return fmt.Errorf("API请求失败: %d - %s", resp.StatusCode, string(body))
	}

	order.SaveRequestLog(logEntry)
	return nil
}

// maskHeaders 脱敏请求头
func (a *OzonAdapter) maskHeaders(clientID string) string {
	return fmt.Sprintf(`{"Client-Id": "%s", "Api-Key": "***"}`, clientID)
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

// OzonFinanceRequest Finance API请求
type OzonFinanceRequest struct {
	Filter   OzonFinanceFilter `json:"filter"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

// OzonFinanceFilter Finance过滤条件
type OzonFinanceFilter struct {
	Date            OzonDateRange `json:"date"`
	OperationType   []string      `json:"operation_type,omitempty"`
	PostingNumber   string        `json:"posting_number,omitempty"`
	TransactionType string        `json:"transaction_type"`
}

// OzonDateRange 日期范围
type OzonDateRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// OzonFinanceResponse Finance API响应
type OzonFinanceResponse struct {
	Result struct {
		Operations []OzonFinanceOperation `json:"operations"`
		PageCount  int                    `json:"page_count"`
		RowCount   int                    `json:"row_count"`
	} `json:"result"`
}

// OzonFinanceOperation 财务操作记录
type OzonFinanceOperation struct {
	OperationID          int64                `json:"operation_id"`
	OperationType        string               `json:"operation_type"`
	OperationDate        string               `json:"operation_date"`
	OperationTypeName    string               `json:"operation_type_name"`
	SaleCommission       float64              `json:"sale_commission"`
	AccrualsForSale      float64              `json:"accruals_for_sale"`
	DeliveryCharge       float64              `json:"delivery_charge"`
	ReturnDeliveryCharge float64              `json:"return_delivery_charge"`
	Amount               float64              `json:"amount"`
	Type                 string               `json:"type"`
	Posting              OzonFinancePosting   `json:"posting"`
	Services             []OzonFinanceService `json:"services"`
}

// OzonFinanceService 财务服务项
type OzonFinanceService struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// OzonFinancePosting 财务关联的发货单
type OzonFinancePosting struct {
	PostingNumber  string `json:"posting_number"`
	DeliverySchema string `json:"delivery_schema"`
	OrderDate      string `json:"order_date"`
	WarehouseID    int64  `json:"warehouse_id"`
}

// SyncOrders 同步订单
func (a *OzonAdapter) SyncOrders(credentials string, since, to time.Time) ([]*order.Order, error) {
	return a.SyncOrdersWithLog(credentials, since, to, 0)
}

// SyncOrdersWithLog 同步订单（带日志记录）
func (a *OzonAdapter) SyncOrdersWithLog(credentials string, since, to time.Time, platformAuthID uint) ([]*order.Order, error) {
	creds, err := a.parseCredentials(credentials)
	if err != nil {
		return nil, err
	}

	var allOrders []*order.Order
	offset := 0
	limit := 100
	requestURL := ozonAPIBaseURL + "/v3/posting/fbs/list"

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

		req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(bodyBytes))
		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Client-Id", creds.ClientID)
		req.Header.Set("Api-Key", creds.APIKey)

		startTime := time.Now()
		client := newHTTPClient()
		resp, err := client.Do(req)
		duration := time.Since(startTime).Milliseconds()

		// 准备日志条目
		logEntry := &order.OrdersRequestLog{
			PlatformAuthID: platformAuthID,
			Platform:       order.PlatformOzon,
			RequestType:    order.RequestTypeOrderList,
			RequestURL:     requestURL,
			RequestMethod:  "POST",
			RequestHeaders: a.maskHeaders(creds.ClientID),
			RequestBody:    string(bodyBytes),
			Duration:       duration,
			CreatedAt:      time.Now(),
		}

		if err != nil {
			logEntry.ErrorMessage = err.Error()
			logEntry.ResponseStatus = 0
			order.SaveRequestLog(logEntry)
			return nil, fmt.Errorf("请求失败: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		logEntry.ResponseStatus = resp.StatusCode
		logEntry.ResponseBody = string(body)

		if err != nil {
			logEntry.ErrorMessage = err.Error()
			order.SaveRequestLog(logEntry)
			return nil, fmt.Errorf("读取响应失败: %w", err)
		}

		if resp.StatusCode != 200 {
			logEntry.ErrorMessage = fmt.Sprintf("API请求失败: %d", resp.StatusCode)
			order.SaveRequestLog(logEntry)
			return nil, fmt.Errorf("API请求失败: %d - %s", resp.StatusCode, string(body))
		}

		// 保存成功的请求日志
		order.SaveRequestLog(logEntry)

		var listResp OzonOrderListResponse
		if err := json.Unmarshal(body, &listResp); err != nil {
			return nil, fmt.Errorf("解析响应失败: %w", err)
		}

		// 调试日志
		log.Printf("[OZON] 同步订单: offset=%d, 获取到 %d 条订单", offset, len(listResp.Result.Postings))

		if len(listResp.Result.Postings) == 0 {
			break
		}

		for _, posting := range listResp.Result.Postings {
			// 只序列化当前订单的数据
			postingData, _ := json.Marshal(posting)
			ord := a.convertToOrder(posting, string(postingData))
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

// GetCommissions 获取佣金信息
func (a *OzonAdapter) GetCommissions(credentials string, since, to time.Time) (map[string]*order.CommissionData, error) {
	return a.GetCommissionsWithLog(credentials, since, to, 0)
}

// GetCommissionsWithLog 获取佣金信息（带日志记录）
func (a *OzonAdapter) GetCommissionsWithLog(credentials string, since, to time.Time, platformAuthID uint) (map[string]*order.CommissionData, error) {
	creds, err := a.parseCredentials(credentials)
	if err != nil {
		return nil, err
	}

	// 获取结算货币，默认RUB（Ozon财务数据以卢布为单位）
	settlementCurrency := creds.SettlementCurrency
	if settlementCurrency == "" {
		settlementCurrency = "RUB"
	}

	result := make(map[string]*order.CommissionData)
	page := 1
	pageSize := 1000
	requestURL := ozonAPIBaseURL + "/v3/finance/transaction/list"

	for {
		reqBody := OzonFinanceRequest{
			Filter: OzonFinanceFilter{
				Date: OzonDateRange{
					From: since.Format(time.RFC3339),
					To:   to.Format(time.RFC3339),
				},
				TransactionType: "all",
			},
			Page:     page,
			PageSize: pageSize,
		}

		bodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			return nil, err
		}

		req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(bodyBytes))
		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Client-Id", creds.ClientID)
		req.Header.Set("Api-Key", creds.APIKey)

		startTime := time.Now()
		client := newHTTPClient()
		resp, err := client.Do(req)
		duration := time.Since(startTime).Milliseconds()

		// 准备日志条目
		logEntry := &order.OrdersRequestLog{
			PlatformAuthID: platformAuthID,
			Platform:       order.PlatformOzon,
			RequestType:    order.RequestTypeFinance,
			RequestURL:     requestURL,
			RequestMethod:  "POST",
			RequestHeaders: a.maskHeaders(creds.ClientID),
			RequestBody:    string(bodyBytes),
			Duration:       duration,
			CreatedAt:      time.Now(),
		}

		if err != nil {
			logEntry.ErrorMessage = err.Error()
			logEntry.ResponseStatus = 0
			order.SaveRequestLog(logEntry)
			return nil, fmt.Errorf("请求失败: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		logEntry.ResponseStatus = resp.StatusCode
		logEntry.ResponseBody = string(body)

		if err != nil {
			logEntry.ErrorMessage = err.Error()
			order.SaveRequestLog(logEntry)
			return nil, fmt.Errorf("读取响应失败: %w", err)
		}

		if resp.StatusCode != 200 {
			logEntry.ErrorMessage = fmt.Sprintf("API请求失败: %d", resp.StatusCode)
			order.SaveRequestLog(logEntry)
			return nil, fmt.Errorf("API请求失败: %d - %s", resp.StatusCode, string(body))
		}

		// 保存成功的请求日志
		order.SaveRequestLog(logEntry)

		var financeResp OzonFinanceResponse
		if err := json.Unmarshal(body, &financeResp); err != nil {
			return nil, fmt.Errorf("解析响应失败: %w", err)
		}

		if len(financeResp.Result.Operations) == 0 {
			break
		}

		// 按 posting_number 聚合佣金数据
		for _, op := range financeResp.Result.Operations {
			postingNumber := op.Posting.PostingNumber
			if postingNumber == "" {
				continue
			}

			if _, exists := result[postingNumber]; !exists {
				result[postingNumber] = &order.CommissionData{
					CommissionCurrency: settlementCurrency,
				}
			}

			// 累加各项费用
			result[postingNumber].SaleCommission += op.SaleCommission
			result[postingNumber].AccrualsForSale += op.AccrualsForSale
			result[postingNumber].DeliveryCharge += op.DeliveryCharge
			result[postingNumber].ReturnDeliveryCharge += op.ReturnDeliveryCharge
			result[postingNumber].CommissionAmount += op.Amount
		}

		// 检查是否还有更多页
		if page >= financeResp.Result.PageCount {
			break
		}
		page++
	}

	return result, nil
}

func init() {
	// 自动注册适配器
	order.RegisterAdapter(&OzonAdapter{})
}
