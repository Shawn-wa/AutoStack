package ozon

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestCashFlowStatementPagination(t *testing.T) {
	client := NewClient(&Credentials{ClientID: "test-client", APIKey: "test-key"}, nil)
	var page int
	client.GetHTTPClient().GetClient().Transport = orderRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost || req.URL.Path != "/v1/finance/cash-flow-statement/list" {
			t.Fatalf("unexpected request: %s %s", req.Method, req.URL.Path)
		}
		defer req.Body.Close()
		var body CashFlowStatementRequest
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		page++
		if body.Page != page || body.PageSize != 100 || body.Date.From != "2026-10-01T00:00:00Z" || body.Date.To != "2026-10-31T23:59:59.999Z" {
			t.Fatalf("unexpected request body: %+v", body)
		}
		payload := `{"result":{"page_count":2,"cash_flows":[{"period":{"id":1,"begin":"2026-10-01T00:00:00Z","end":"2026-10-15T23:59:59Z"},"orders_amount":123.45,"currency_code":"RUB"}]}}`
		if page == 2 {
			payload = `{"result":{"page_count":2,"cash_flows":[{"period":{"id":2,"begin":"2026-10-16T00:00:00Z","end":"2026-10-31T23:59:59Z"},"orders_amount":678.9,"currency_code":"RUB"}]}}`
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(payload))}, nil
	})
	items, err := NewFinanceAPI(client).GetAllCashFlowStatements(
		time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.October, 31, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatal(err)
	}
	if page != 2 || len(items) != 2 || items[0].OrdersAmount != 123.45 || items[1].Period.ID != 2 || items[1].CurrencyCode != "RUB" {
		t.Fatalf("got %d pages with items %+v", page, items)
	}
}

func TestMutualSettlementReport(t *testing.T) {
	const reportFile = "https://example.com/report.xlsx?token=test-token"
	const expiresAt = "2026-11-06T08:00:00Z"
	const createReport = `{"result":{"code":"report-code"}}`
	tests := []struct {
		name         string
		create       string
		polls        []string
		wantError    string
		wantFile     string
		wantExpires  string
		wantBalances int
	}{
		{
			name:   "ready after waiting",
			create: createReport,
			polls: []string{
				`{"result":{"code":"report-code","status":"waiting"}}`,
				`{"result":{"code":"report-code","status":"success","report_type":"MUTUAL_SETTLEMENT","file":"` + reportFile + `","expires_at":"` + expiresAt + `"}}`,
			},
			wantFile:    reportFile,
			wantExpires: expiresAt,
		},
		{
			name:      "missing report code",
			create:    `{"result":{}}`,
			wantError: "未返回结算报告编号",
		},
		{
			name:      "missing download link",
			create:    createReport,
			polls:     []string{`{"result":{"status":"success"}}`},
			wantError: "未返回结算报告下载链接",
		},
		{
			name:      "generation failure",
			create:    createReport,
			polls:     []string{`{"result":{"status":"failed","error":"test-report-error"}}`},
			wantError: "test-report-error",
		},
		{
			name:         "legacy direct response",
			create:       `{"result":{"details":[{"balance":12.5,"currency_code":"RUB"}]}}`,
			wantBalances: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(&Credentials{ClientID: "test-client", APIKey: "test-key"}, nil)
			var created bool
			var polls int
			client.GetHTTPClient().GetClient().Transport = orderRoundTripFunc(func(req *http.Request) (*http.Response, error) {
				defer req.Body.Close()
				var body map[string]string
				if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				if req.Method != http.MethodPost {
					t.Fatalf("unexpected method: %s", req.Method)
				}
				var payload string
				if !created {
					if req.URL.Path != "/v1/finance/mutual-settlement" || body["date"] != "2026-10" {
						t.Fatalf("unexpected create request: %s %v", req.URL.Path, body)
					}
					created = true
					payload = tt.create
				} else {
					if req.URL.Path != "/v1/report/info" || body["code"] != "report-code" || polls >= len(tt.polls) {
						t.Fatalf("unexpected poll request: %s %v", req.URL.Path, body)
					}
					payload = tt.polls[polls]
					polls++
				}
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(payload))}, nil
			})
			result, err := NewFinanceAPI(client).GetMutualSettlementWithReport(time.Time{}, time.Date(2026, time.October, 31, 0, 0, 0, 0, time.UTC), 2)
			if polls != len(tt.polls) {
				t.Errorf("got %d polls, want %d", polls, len(tt.polls))
			}
			if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Fatalf("got error %v, want %q", err, tt.wantError)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(result.CreateResponse.Result.Details) != tt.wantBalances {
				t.Errorf("unexpected balances: %+v", result.CreateResponse.Result.Details)
			}
			if tt.wantFile != "" && (result.ReportInfo == nil || result.ReportInfo.Result.File != tt.wantFile || result.ReportInfo.Result.ExpiresAt != tt.wantExpires) {
				t.Errorf("unexpected report info: %+v", result.ReportInfo)
			}
		})
	}
}
