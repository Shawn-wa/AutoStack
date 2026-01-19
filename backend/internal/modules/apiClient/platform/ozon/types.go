package ozon

// Credentials OZON凭证
type Credentials struct {
	ClientID           string `json:"client_id"`
	APIKey             string `json:"api_key"`
	SettlementCurrency string `json:"settlement_currency"` // 结算货币，默认RUB
}

// ========== 订单相关 ==========

// OrderListRequest 订单列表请求
type OrderListRequest struct {
	Dir    string          `json:"dir"`
	Filter OrderListFilter `json:"filter"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
	With   OrderListWith   `json:"with"`
}

// OrderListFilter 订单过滤条件
type OrderListFilter struct {
	Since  string `json:"since"`
	To     string `json:"to"`
	Status string `json:"status,omitempty"`
}

// OrderListWith 订单附加数据
type OrderListWith struct {
	AnalyticsData bool `json:"analytics_data"`
	FinancialData bool `json:"financial_data"`
}

// OrderListResponse 订单列表响应
type OrderListResponse struct {
	Result struct {
		Postings []Posting `json:"postings"`
	} `json:"result"`
}

// Posting 发货单
type Posting struct {
	PostingNumber  string         `json:"posting_number"`
	OrderID        int64          `json:"order_id"`
	OrderNumber    string         `json:"order_number"`
	Status         string         `json:"status"`
	InProcessAt    string         `json:"in_process_at"`
	ShipmentDate   string         `json:"shipment_date"`
	DeliveringDate string         `json:"delivering_date"`
	Products       []Product      `json:"products"`
	Customer       *Customer      `json:"customer,omitempty"`
	AddressInfo    *AddressInfo   `json:"addressee,omitempty"`
	FinancialData  *FinancialData `json:"financial_data,omitempty"`
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

// ========== 财务相关 ==========

// FinanceListRequest 财务列表请求
type FinanceListRequest struct {
	Filter   FinanceFilter `json:"filter"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

// FinanceFilter 财务过滤条件
type FinanceFilter struct {
	Date            DateRange `json:"date"`
	TransactionType string    `json:"transaction_type"`
}

// DateRange 日期范围
type DateRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// FinanceListResponse 财务列表响应
type FinanceListResponse struct {
	Result struct {
		Operations []FinanceOperation `json:"operations"`
		PageCount  int                `json:"page_count"`
		RowCount   int                `json:"row_count"`
	} `json:"result"`
}

// FinanceOperation 财务操作
type FinanceOperation struct {
	OperationID          int64          `json:"operation_id"`
	OperationType        string         `json:"operation_type"`
	OperationTypeName    string         `json:"operation_type_name"`
	OperationDate        string         `json:"operation_date"`
	Amount               float64        `json:"amount"`
	SaleCommission       float64        `json:"sale_commission"`
	AccrualsForSale      float64        `json:"accruals_for_sale"`
	DeliveryCharge       float64        `json:"delivery_charge"`
	ReturnDeliveryCharge float64        `json:"return_delivery_charge"`
	Posting              FinancePosting `json:"posting"`
}

// FinancePosting 财务关联的发货单
type FinancePosting struct {
	PostingNumber  string `json:"posting_number"`
	DeliverySchema string `json:"delivery_schema"`
	OrderDate      string `json:"order_date"`
	WarehouseID    int64  `json:"warehouse_id"`
}

// FinanceTotalsRequest 财务汇总请求
// API: POST /v3/finance/transaction/totals
type FinanceTotalsRequest struct {
	PostingNumber string `json:"posting_number"` // 发货单号
}

// FinanceTotalsResponse 财务汇总响应
type FinanceTotalsResponse struct {
	Result FinanceTotalsItem `json:"result"`
}

// FinanceTotalsItem 财务汇总项 - 单个订单的财务汇总数据
// API: POST /v3/finance/transaction/totals
// 文档: https://docs.ozon.ru/api/seller/#operation/FinanceAPI_FinanceTransactionTotalV3
type FinanceTotalsItem struct {
	AccrualsForSale         float64 `json:"accruals_for_sale"`         // 销售收入：卖家因销售商品获得的收入金额（正数，如 1500.00）
	SaleCommission          float64 `json:"sale_commission"`           // 销售佣金：Ozon平台从销售中收取的佣金费用（负数，如 -225.00）
	ProcessingAndDelivery   float64 `json:"processing_and_delivery"`   // 物流费用：商品处理、包装和配送的费用（负数，如 -150.00）
	RefundsAndCancellations float64 `json:"refunds_and_cancellations"` // 退款退货：退款及取消订单产生的费用扣减（通常为负数或0）
	ServicesAmount          float64 `json:"services_amount"`           // 平台服务费：Ozon提供的增值服务费用（负数，如广告费、仓储费）
	CompensationAmount      float64 `json:"compensation_amount"`       // 补偿金额：平台对卖家的补偿款项（正数，如物流损坏赔偿）
	MoneyTransfer           float64 `json:"money_transfer"`            // 资金划转：账户间资金转移记录（可正可负）
	OthersAmount            float64 `json:"others_amount"`             // 其他费用：未归类的其他杂项费用
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
	Code       string `json:"code"`        // 报告唯一标识符
	Status     string `json:"status"`      // 报告状态：waiting, processing, success, failed
	Error      string `json:"error"`       // 错误信息
	File       string `json:"file"`        // 报告文件下载链接
	ReportType string `json:"report_type"` // 报告类型
	CreatedAt  string `json:"created_at"`  // 创建时间
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
// 文档: https://docs.ozon.com/api/seller/en/#operation/ProductAPI_GetProductListV3
type ProductListRequest struct {
	Filter ProductListFilter `json:"filter"`
	LastID string            `json:"last_id,omitempty"`
	Limit  int               `json:"limit"`
}

// ProductListFilter 商品列表过滤条件
type ProductListFilter struct {
	OfferID    []string `json:"offer_id,omitempty"`
	ProductID  []int64  `json:"product_id,omitempty"`
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
	Total  int           `json:"total"`
}

// ProductItem 商品项
type ProductItem struct {
	ProductID int64  `json:"product_id"`
	OfferID   string `json:"offer_id"`
}

// ProductInfoRequest 商品详情请求
// API: POST /v2/product/info/list
type ProductInfoRequest struct {
	OfferID   []string `json:"offer_id,omitempty"`
	ProductID []int64  `json:"product_id,omitempty"`
}

// ProductInfoResponse 商品详情响应
type ProductInfoResponse struct {
	Result ProductInfoResult `json:"result"`
}

// ProductInfoResult 商品详情结果
type ProductInfoResult struct {
	Items []ProductInfoItem `json:"items"`
}

// ProductInfoItem 商品详情项
type ProductInfoItem struct {
	ID             int64            `json:"id"`
	Name           string           `json:"name"`
	OfferID        string           `json:"offer_id"`
	Barcode        string           `json:"barcode"`
	Price          string           `json:"price"` // 注意：Ozon API 返回的价格可能是字符串或数字，这里根据实际情况可能需要调整，但通常作为字符串处理更安全
	OldPrice       string           `json:"old_price"`
	PremiumPrice   string           `json:"premium_price"`
	Vat            string           `json:"vat"`
	Visible        bool             `json:"visible"`
	Stocks         ProductInfoStock `json:"stocks"`
	Status         ProductStatus    `json:"status"`
	MarketingPrice string           `json:"marketing_price"`
	CurrencyCode   string           `json:"currency_code"`
}

// ProductInfoStock 商品库存信息
type ProductInfoStock struct {
	Coming   int `json:"coming"`
	Present  int `json:"present"`
	Reserved int `json:"reserved"`
}

// ProductStatus 商品状态
type ProductStatus struct {
	State            string `json:"state"`
	StateFailDetails string `json:"state_fail_details"`
	StateName        string `json:"state_name"`
	StateDescription string `json:"state_description"`
	IsFailed         bool   `json:"is_failed"`
	IsCreated        bool   `json:"is_created"`
}
