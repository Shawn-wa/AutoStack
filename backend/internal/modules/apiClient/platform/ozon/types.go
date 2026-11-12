package ozon

import "encoding/json"

// Credentials OZON凭证
type Credentials struct {
	ClientID           string `json:"client_id"`
	APIKey             string `json:"api_key"`
	SettlementCurrency string `json:"settlement_currency"` // 结算货币，默认RUB
}

// ========== 订单相关 ==========

// OrderListRequest 订单列表请求
type OrderListRequest struct {
	SortDir string          `json:"sort_dir"`
	Filter  OrderListFilter `json:"filter"`
	Limit   int             `json:"limit"`
	Cursor  string          `json:"cursor,omitempty"`
	With    OrderListWith   `json:"with"`
}

// OrderListFilter 订单过滤条件
type OrderListFilter struct {
	Since string `json:"since"`
	To    string `json:"to"`
}

// OrderListWith 订单附加数据
type OrderListWith struct {
	AnalyticsData bool `json:"analytics_data"`
	FinancialData bool `json:"financial_data"`
}

// OrderListResponse 订单列表响应
type OrderListResponse struct {
	Postings []Posting `json:"postings"`
	HasNext  bool      `json:"has_next"`
	Cursor   string    `json:"cursor"`
}

// OrderDetailRequest 订单详情请求
// API: POST /v3/posting/fbs/get
// 文档: https://docs.ozon.ru/api/seller/zh/#operation/PostingAPI_GetFbsPostingV3
type OrderDetailRequest struct {
	PostingNumber string        `json:"posting_number"`
	With          OrderListWith `json:"with"`
}

// OrderDetailResponse 订单详情响应
type OrderDetailResponse struct {
	Result Posting `json:"result"`
}

// Posting 发货单
type Posting struct {
	PostingNumber            string         `json:"posting_number"`
	OrderID                  int64          `json:"order_id"`
	OrderNumber              string         `json:"order_number"`
	Status                   string         `json:"status"`
	InProcessAt              string         `json:"in_process_at"`
	ShipmentDate             string         `json:"shipment_date"`               // 预计发货日期
	ShipmentDateWithoutDelay string         `json:"shipment_date_without_delay"` // 发货截止时间（逾期时间）
	DeliveringDate           string         `json:"delivering_date"`
	Products                 []Product      `json:"products"`
	Customer                 *Customer      `json:"customer,omitempty"`
	AddressInfo              *AddressInfo   `json:"addressee,omitempty"`
	FinancialData            *FinancialData `json:"financial_data,omitempty"`
}

// Product 商品
type Product struct {
	Sku          int64  `json:"sku"`
	Name         string `json:"name"`
	Quantity     int    `json:"quantity"`
	OfferID      string `json:"offer_id"`
	Price        string `json:"price"`
	CurrencyCode string `json:"currency_code"`
}

// UnmarshalJSON 兼容 v3 订单详情的字符串价格和 v4 订单列表的价格对象。
func (p *Product) UnmarshalJSON(data []byte) error {
	type plainProduct Product
	var wire struct {
		*plainProduct
		Price json.RawMessage `json:"price"`
	}
	wire.plainProduct = (*plainProduct)(p)
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	if len(wire.Price) == 0 {
		return nil
	}
	if wire.Price[0] != '{' {
		return json.Unmarshal(wire.Price, &p.Price)
	}

	var price struct {
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
	}
	if err := json.Unmarshal(wire.Price, &price); err != nil {
		return err
	}
	p.Price = price.Amount
	p.CurrencyCode = price.Currency
	return nil
}

// Customer 客户
type Customer struct {
	Name    string  `json:"name"`
	Phone   string  `json:"phone"`
	Email   string  `json:"email"`
	Address Address `json:"address"`
}

// Address 地址
type Address struct {
	Country     string `json:"country"`
	Region      string `json:"region"`
	City        string `json:"city"`
	ZipCode     string `json:"zip_code"`
	AddressTail string `json:"address_tail"`
}

// AddressInfo 收件人信息
type AddressInfo struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	City    string `json:"city"`
	Region  string `json:"region"`
	Country string `json:"country"`
}

// FinancialData 财务数据
type FinancialData struct {
	Products []FinancialProduct `json:"products"`
}

// FinancialProduct 财务商品
type FinancialProduct struct {
	Price        float64 `json:"price"`
	CurrencyCode string  `json:"currency_code"`
}

// DateRange 日期范围
type DateRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// ========== 现金流报表相关 ==========

// CashFlowStatementRequest 现金流报表请求
// API: POST /v1/finance/cash-flow-statement/list
type CashFlowStatementRequest struct {
	Date     DateRange `json:"date"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
}

// CashFlowStatementResponse 现金流报表响应
type CashFlowStatementResponse struct {
	Result CashFlowResult `json:"result"`
}

// CashFlowResult 现金流结果
type CashFlowResult struct {
	CashFlows []CashFlowItem `json:"cash_flows"`
	PageCount int            `json:"page_count"`
}

// CashFlowItem 现金流条目 - 按周期汇总的财务数据
// API: POST /v1/finance/cash-flow-statement/list
// 文档: https://docs.ozon.ru/api/seller/#operation/FinanceAPI_FinanceCashFlowStatementList
type CashFlowItem struct {
	Period                      CashFlowPeriod `json:"period"`                          // 报告周期：包含周期ID、开始和结束时间
	OrdersAmount                float64        `json:"orders_amount"`                   // 订单销售金额：该周期内订单的总销售额（正数）
	ReturnsAmount               float64        `json:"returns_amount"`                  // 退货金额：该周期内退货产生的金额（负数）
	CommissionAmount            float64        `json:"commission_amount"`               // 平台佣金：Ozon收取的销售佣金总额（负数）
	ServicesAmount              float64        `json:"services_amount"`                 // 服务费用：平台提供的各类服务费用（负数，如广告费、仓储费）
	ItemDeliveryAndReturnAmount float64        `json:"item_delivery_and_return_amount"` // 物流费用：商品配送和退货物流费用（负数）
	CurrencyCode                string         `json:"currency_code"`                   // 货币代码：结算货币，如 RUB、USD、CNY 等
}

// CashFlowPeriod 报告周期
type CashFlowPeriod struct {
	ID    int64  `json:"id"`
	Begin string `json:"begin"`
	End   string `json:"end"`
}

// ========== 结算报告相关 ==========

// MutualSettlementRequest 结算报告请求
// API: POST /v1/finance/mutual-settlement
// 注意：date 字段是字符串格式 "YYYY-MM"（年-月），不是完整日期
type MutualSettlementRequest struct {
	Date string `json:"date"` // 日期，格式：YYYY-MM（如 2026-01）
}

// MutualSettlementResponse 结算报告响应（创建报告）
type MutualSettlementResponse struct {
	Result MutualSettlementResult `json:"result"`
}

// MutualSettlementResult 结算报告结果（创建报告返回的 code）
type MutualSettlementResult struct {
	Code    string                   `json:"code"`    // 报告唯一标识符
	Details []MutualSettlementDetail `json:"details"` // 结算明细（如果直接返回）
}

// ========== 报告信息相关 ==========

// ReportInfoRequest 报告信息请求
// API: POST /v1/report/info
type ReportInfoRequest struct {
	Code string `json:"code"` // 报告唯一标识符
}

// ReportInfoResponse 报告信息响应
type ReportInfoResponse struct {
	Result ReportInfoResult `json:"result"`
}

// ReportInfoResult 报告信息结果
type ReportInfoResult struct {
	Code       string `json:"code"`                 // 报告唯一标识符
	Status     string `json:"status"`               // 报告状态：waiting, processing, success, failed
	Error      string `json:"error"`                // 错误信息
	File       string `json:"file"`                 // XLSX 报告文件下载链接
	ReportType string `json:"report_type"`          // 报告类型
	CreatedAt  string `json:"created_at"`           // 创建时间
	ExpiresAt  string `json:"expires_at,omitempty"` // 下载链接过期时间
}

// MutualSettlementFullResponse 结算报告完整响应（包含创建响应和报告信息）
type MutualSettlementFullResponse struct {
	CreateResponse *MutualSettlementResponse `json:"create_response"` // 创建报告响应
	ReportInfo     *ReportInfoResponse       `json:"report_info"`     // 报告信息（如果是异步生成）
}

// MutualSettlementDetail 结算明细
// API: POST /v1/finance/mutual-settlement
// 对应页面: https://seller.ozon.ru/app/finances/balance
type MutualSettlementDetail struct {
	BalanceAtBeginning float64 `json:"balance_at_beginning"` // 期初余额（如：一月月初 -6,163）
	AccruedAmount      float64 `json:"accrued_amount"`       // 应计金额（如：已于一月应计 36,476）
	PaidAmount         float64 `json:"paid_amount"`          // 已支付金额（如：已于一月支付 0）
	Balance            float64 `json:"balance"`              // 当前余额（如：当前余额 30,313）
	CurrencyCode       string  `json:"currency_code"`        // 货币代码
	PeriodName         string  `json:"period_name"`          // 周期名称
}

// ========== 产品相关 ==========

// ProductListRequest 商品列表请求
// API: POST /v3/product/list
// 文档: https://docs.ozon.ru/api/seller/#operation/ProductAPI_GetProductList
type ProductListRequest struct {
	Filter ProductListFilter `json:"filter"`
	LastID string            `json:"last_id,omitempty"`
	Limit  int               `json:"limit"`
}

// ProductListFilter 商品列表过滤条件
type ProductListFilter struct {
	OfferID    []string `json:"offer_id,omitempty"`
	ProductID  []string `json:"product_id,omitempty"`
	SKUs       []string `json:"skus,omitempty"`
	Visibility string   `json:"visibility,omitempty"` // ALL, VISIBLE, INVISIBLE, EMPTY_STOCK, NOT_MODERATED, MODERATED, DISABLED, STATE_FAILED, READY_TO_SUPPLY, VALIDATION_STATE_PENDING, VALIDATION_STATE_FAIL, VALIDATION_STATE_SUCCESS, TO_SUPPLY, IN_SALE, REMOVED_FROM_SALE, BANNED, OVERPRICED, CRITICALLY_OVERPRICED, EMPTY_BARCODE, BARCODE_EXIST, QUARANTINE, ARCHIVED, OVERPRICED_WITH_STOCK, PARTIAL_APPROVED, IMAGE_ABSENT, MODERATION_BLOCK
}

// ProductListResponse 商品列表响应
type ProductListResponse struct {
	Result ProductListResult `json:"result"`
}

// ProductListResult 商品列表结果
type ProductListResult struct {
	Items  []ProductItem `json:"items"`
	LastID string        `json:"last_id"`
	Total  int64         `json:"total_items"`
}

// UnmarshalJSON 优先读取 total_items，兼容仍返回 total 的响应。
func (r *ProductListResult) UnmarshalJSON(data []byte) error {
	type resultFields ProductListResult
	wire := struct {
		*resultFields
		TotalItems  *int64 `json:"total_items"`
		LegacyTotal *int64 `json:"total"`
	}{resultFields: (*resultFields)(r)}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	if wire.TotalItems != nil {
		r.Total = *wire.TotalItems
	} else if wire.LegacyTotal != nil {
		r.Total = *wire.LegacyTotal
	}
	return nil
}

// ProductItem 商品项
type ProductItem struct {
	ProductID int64  `json:"product_id"`
	OfferID   string `json:"offer_id"`
	SKU       int64  `json:"sku"`
}

// ProductInfoRequest 商品详情请求
// API: POST /v3/product/info/list
// 文档: https://docs.ozon.ru/api/seller/#operation/ProductAPI_GetProductInfoList
type ProductInfoRequest struct {
	OfferID   []string `json:"offer_id,omitempty"`
	ProductID []string `json:"product_id,omitempty"`
	SKU       []string `json:"sku,omitempty"`
}

// ProductInfoResponse 商品详情响应 (v3 API 直接返回 items，无 result 包装)
type ProductInfoResponse struct {
	Items []ProductInfoItem `json:"items"`
}

// ProductInfoItem 商品详情项
type ProductInfoItem struct {
	ID                    int64                    `json:"id"`
	SKU                   int64                    `json:"sku"`
	Name                  string                   `json:"name"`
	OfferID               string                   `json:"offer_id"`
	Barcodes              []string                 `json:"barcodes"`
	Price                 string                   `json:"price"`
	DeclaredPrice         *ProductDeclaredPrice    `json:"declared_price,omitempty"`
	OldPrice              string                   `json:"old_price"`
	MinPrice              string                   `json:"min_price"`
	MarketingPrice        string                   `json:"marketing_price"`
	Vat                   string                   `json:"vat"`
	CurrencyCode          string                   `json:"currency_code"`
	Stocks                ProductInfoStocks        `json:"stocks"`
	Errors                []ProductError           `json:"errors"`
	VisibilityDetails     ProductVisibilityDetails `json:"visibility_details"`
	PriceIndexes          ProductPriceIndexes      `json:"price_indexes"`
	VolumeWeight          float64                  `json:"volume_weight"`
	IsArchived            bool                     `json:"is_archived"`
	IsAutoarchived        bool                     `json:"is_autoarchived"`
	IsDiscounted          bool                     `json:"is_discounted"`
	IsPrepaymentAllowed   bool                     `json:"is_prepayment_allowed"`
	HasDiscountedFboItem  bool                     `json:"has_discounted_fbo_item"`
	DiscountedFboStocks   int                      `json:"discounted_fbo_stocks"`
	DescriptionCategoryID int64                    `json:"description_category_id"`
	TypeID                int64                    `json:"type_id"`
	CreatedAt             string                   `json:"created_at"`
	UpdatedAt             string                   `json:"updated_at"`
	Images                []string                 `json:"images"`
	PrimaryImage          interface{}              `json:"primary_image"` // 主图URL (可能是字符串或数组)
}

// ProductDeclaredPrice 商品申报价值。
type ProductDeclaredPrice struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

// GetPrimaryImageURL 获取主图URL
func (p *ProductInfoItem) GetPrimaryImageURL() string {
	switch v := p.PrimaryImage.(type) {
	case string:
		return v
	case []interface{}:
		if len(v) > 0 {
			if s, ok := v[0].(string); ok {
				return s
			}
		}
	}
	// 如果 primary_image 无效，使用 images[0]
	if len(p.Images) > 0 {
		return p.Images[0]
	}
	return ""
}

// ProductInfoStocks 商品库存信息 (v3 结构)
type ProductInfoStocks struct {
	HasStock bool               `json:"has_stock"`
	Stocks   []ProductStockItem `json:"stocks"`
}

// ProductStockItem 库存明细项
type ProductStockItem struct {
	Type     string `json:"type"`     // fbo, fbs 等
	Present  int    `json:"present"`  // 当前库存
	Reserved int    `json:"reserved"` // 预留库存
	SKU      int64  `json:"sku"`
	Source   string `json:"source"`
}

// ProductError 商品错误信息
type ProductError struct {
	Code        string `json:"code"`
	Field       string `json:"field"`
	AttributeID int64  `json:"attribute_id"`
	State       string `json:"state"`
	Level       string `json:"level"`
}

// ProductVisibilityDetails 商品可见性详情
type ProductVisibilityDetails struct {
	HasPrice bool `json:"has_price"`
	HasStock bool `json:"has_stock"`
}

// ProductPriceIndexes 价格指数
type ProductPriceIndexes struct {
	ColorIndex        string                 `json:"color_index"`
	ExternalIndexData ProductExternalIndexes `json:"external_index_data"`
}

// ProductExternalIndexes 外部价格指数
type ProductExternalIndexes struct {
	MinimalPrice         string  `json:"minimal_price"`
	MinimalPriceCurrency string  `json:"minimal_price_currency"`
	PriceIndexValue      float64 `json:"price_index_value"`
}
