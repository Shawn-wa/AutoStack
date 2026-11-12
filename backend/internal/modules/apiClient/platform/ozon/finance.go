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
// 文档路径: https://docs.ozon.ru/api/seller/#operation/ReportAPI_CreateMutualSettlementReport
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

	if createResp.Result.Code == "" {
		return result, fmt.Errorf("Ozon 未返回结算报告编号")
	}
	if maxRetries <= 0 {
		maxRetries = 10
	}

	var lastPollError error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(2 * time.Second)
		}

		reportInfo, err := api.GetReportInfo(createResp.Result.Code)
		lastPollError = err
		if err != nil {
			continue
		}

		result.ReportInfo = reportInfo
		switch reportInfo.Result.Status {
		case "success":
			if reportInfo.Result.File == "" {
				return result, fmt.Errorf("Ozon 未返回结算报告下载链接")
			}
			return result, nil
		case "failed":
			return result, fmt.Errorf("报告生成失败: %s", reportInfo.Result.Error)
		}
	}

	if lastPollError != nil {
		return result, fmt.Errorf("查询结算报告状态失败: %w", lastPollError)
	}
	return result, fmt.Errorf("报告生成超时，请稍后重试")
}
