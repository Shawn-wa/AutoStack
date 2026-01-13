package ozon

import (
	"encoding/json"
	"fmt"
	"time"

	"autostack/internal/modules/apiClient/platform"
)

// FinanceAPI 财务API
type FinanceAPI struct {
	client *Client
}

// NewFinanceAPI 创建财务API
func NewFinanceAPI(client *Client) *FinanceAPI {
	return &FinanceAPI{client: client}
}

// GetTransactionList 获取财务交易列表
// API: POST /v3/finance/transaction/list
func (api *FinanceAPI) GetTransactionList(since, to time.Time, page, pageSize int) (*FinanceListResponse, error) {
	req := FinanceListRequest{
		Filter: FinanceFilter{
			Date: DateRange{
				From: since.Format(time.RFC3339),
				To:   to.Format(time.RFC3339),
			},
			TransactionType: "all",
		},
		Page:     page,
		PageSize: pageSize,
	}

	resp, err := api.client.DoRequest("POST", "/v3/finance/transaction/list", req, platform.RequestTypeFinance)
	if err != nil {
		return nil, err
	}

	var result FinanceListResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &result, nil
}

// GetTransactionTotals 获取单个订单的财务汇总
// API: POST /v3/finance/transaction/totals
func (api *FinanceAPI) GetTransactionTotals(postingNumber string) (*FinanceTotalsResponse, error) {
	req := FinanceTotalsRequest{
		PostingNumber: postingNumber,
	}

	resp, err := api.client.DoRequest("POST", "/v3/finance/transaction/totals", req, platform.RequestTypeFinance)
	if err != nil {
		return nil, err
	}

	var result FinanceTotalsResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &result, nil
}

// CommissionData 佣金数据（聚合结果）
type CommissionData struct {
	AccrualsForSale         float64 // 销售收入
	SaleCommission          float64 // 销售佣金
	ProcessingAndDelivery   float64 // 加工和配送费
	RefundsAndCancellations float64 // 退款和取消
	ServicesAmount          float64 // 服务费
	CompensationAmount      float64 // 补偿金额
	MoneyTransfer           float64 // 资金转账
	OthersAmount            float64 // 其他金额
	ProfitAmount            float64 // 订单利润额（所有项目汇总）
	CommissionCurrency      string  // 货币
}

// GetAllCommissions 获取所有佣金数据（分页获取全部）
func (api *FinanceAPI) GetAllCommissions(since, to time.Time, currency string) (map[string]*CommissionData, error) {
	result := make(map[string]*CommissionData)
	page := 1
	pageSize := 1000

	for {
		resp, err := api.GetTransactionList(since, to, page, pageSize)
		if err != nil {
			return nil, err
		}

		if len(resp.Result.Operations) == 0 {
			break
		}

		// 按 posting_number 聚合佣金数据
		for _, op := range resp.Result.Operations {
			postingNumber := op.Posting.PostingNumber
			if postingNumber == "" {
				continue
			}

			if _, exists := result[postingNumber]; !exists {
				result[postingNumber] = &CommissionData{
					CommissionCurrency: currency,
				}
			}

			// 累加各项费用（此方法已废弃，仅保留基本字段兼容）
			result[postingNumber].AccrualsForSale += op.AccrualsForSale
			result[postingNumber].SaleCommission += op.SaleCommission
			// 计算利润
			result[postingNumber].ProfitAmount += op.Amount
		}

		// 检查是否还有更多页
		if page >= resp.Result.PageCount {
			break
		}
		page++
	}

	return result, nil
}

// GetSingleOrderCommission 获取单个订单的佣金数据
func (api *FinanceAPI) GetSingleOrderCommission(postingNumber, currency string) (*CommissionData, error) {
	resp, err := api.GetTransactionTotals(postingNumber)
	if err != nil {
		return nil, err
	}

	item := resp.Result
	commData := &CommissionData{
		CommissionCurrency:      currency,
		AccrualsForSale:         item.AccrualsForSale,
		SaleCommission:          item.SaleCommission,
		ProcessingAndDelivery:   item.ProcessingAndDelivery,
		RefundsAndCancellations: item.RefundsAndCancellations,
		ServicesAmount:          item.ServicesAmount,
		CompensationAmount:      item.CompensationAmount,
		MoneyTransfer:           item.MoneyTransfer,
		OthersAmount:            item.OthersAmount,
	}

	// 计算订单利润额 = 所有项目的总和
	commData.ProfitAmount = commData.AccrualsForSale +
		commData.SaleCommission +
		commData.ProcessingAndDelivery +
		commData.RefundsAndCancellations +
		commData.ServicesAmount +
		commData.CompensationAmount +
		commData.MoneyTransfer +
		commData.OthersAmount

	return commData, nil
}

// GetCashFlowStatementList 获取现金流报表列表
// API: POST /v1/finance/cash-flow-statement/list
func (api *FinanceAPI) GetCashFlowStatementList(since, to time.Time, page, pageSize int) (*CashFlowStatementResponse, error) {
	// 构造日期范围：开始日期的 00:00:00 到结束日期的 23:59:59
	sinceDate := time.Date(since.Year(), since.Month(), since.Day(), 0, 0, 0, 0, time.UTC)
	toDate := time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 999000000, time.UTC)

	req := CashFlowStatementRequest{
		Date: DateRange{
			From: sinceDate.Format(time.RFC3339Nano),
			To:   toDate.Format(time.RFC3339Nano),
		},
		Page:     page,
		PageSize: pageSize,
	}

	resp, err := api.client.DoRequest("POST", "/v1/finance/cash-flow-statement/list", req, platform.RequestTypeCashFlow)
	if err != nil {
		return nil, err
	}

	var result CashFlowStatementResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &result, nil
}

// GetAllCashFlowStatements 获取所有现金流报表（分页获取全部）
func (api *FinanceAPI) GetAllCashFlowStatements(since, to time.Time) ([]CashFlowItem, error) {
	var allItems []CashFlowItem
	page := 1
	pageSize := 100

	for {
		resp, err := api.GetCashFlowStatementList(since, to, page, pageSize)
		if err != nil {
			return nil, err
		}

		if len(resp.Result.CashFlows) == 0 {
			break
		}

		allItems = append(allItems, resp.Result.CashFlows...)

		// 检查是否还有更多页
		if page >= resp.Result.PageCount {
			break
		}
		page++
	}

	return allItems, nil
}

// GetMutualSettlement 获取结算报告
// API: POST /v1/finance/mutual-settlement
// 文档路径: https://docs.ozon.ru/api/seller/#operation/FinanceAPI_MutualSettlementReport
// 对应页面: https://seller.ozon.ru/app/finances/balance (余额页面)
func (api *FinanceAPI) GetMutualSettlement(since, to time.Time) (*MutualSettlementResponse, error) {
	// 使用结束日期的年月作为查询日期（格式：YYYY-MM）
	dateStr := to.Format("2006-01")

	req := MutualSettlementRequest{
		Date: dateStr,
	}

	resp, err := api.client.DoRequest("POST", "/v1/finance/mutual-settlement", req, platform.RequestTypeMutualSettlement)
	if err != nil {
		return nil, err
	}

	var result MutualSettlementResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &result, nil
}

// GetReportInfo 获取报告信息
// API: POST /v1/report/info
// 文档路径: https://docs.ozon.ru/api/seller/#operation/ReportAPI_ReportInfo
// 用于查询报告生成状态和获取下载链接
func (api *FinanceAPI) GetReportInfo(code string) (*ReportInfoResponse, error) {
	req := ReportInfoRequest{
		Code: code,
	}

	resp, err := api.client.DoRequest("POST", "/v1/report/info", req, platform.RequestTypeReportInfo)
	if err != nil {
		return nil, err
	}

	var result ReportInfoResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &result, nil
}

// GetMutualSettlementWithReport 获取结算报告（包含轮询报告状态）
// 先创建报告，然后轮询获取报告结果
func (api *FinanceAPI) GetMutualSettlementWithReport(since, to time.Time, maxRetries int) (*MutualSettlementFullResponse, error) {
	// 1. 创建结算报告
	createResp, err := api.GetMutualSettlement(since, to)
	if err != nil {
		return nil, fmt.Errorf("创建结算报告失败: %w", err)
	}

	result := &MutualSettlementFullResponse{
		CreateResponse: createResp,
	}

	// 如果直接返回了 details，说明是同步返回，无需轮询
	if len(createResp.Result.Details) > 0 {
		return result, nil
	}

	// 如果返回了 code，需要轮询报告状态
	if createResp.Result.Code != "" {
		if maxRetries <= 0 {
			maxRetries = 10
		}

		for i := 0; i < maxRetries; i++ {
			time.Sleep(2 * time.Second)

			reportInfo, err := api.GetReportInfo(createResp.Result.Code)
			if err != nil {
				continue
			}

			result.ReportInfo = reportInfo

			switch reportInfo.Result.Status {
			case "success":
				return result, nil
			case "failed":
				return result, fmt.Errorf("报告生成失败: %s", reportInfo.Result.Error)
			case "waiting", "processing":
				// 继续等待
				continue
			}
		}

		return result, fmt.Errorf("报告生成超时，请稍后重试")
	}

	return result, nil
}
