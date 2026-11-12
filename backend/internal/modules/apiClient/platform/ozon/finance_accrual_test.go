package ozon

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestGetCommissionsForPostings(t *testing.T) {
	client := NewClient(&Credentials{ClientID: "test-client", APIKey: "test-key"}, nil)
	postings := make([]string, accrualPostingLimit+1)
	for i := range postings {
		postings[i] = fmt.Sprintf("posting-%d", i)
	}
	typesCalls, postingsCalls, retries := 0, 0, 0
	client.GetHTTPClient().GetClient().Transport = orderRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		defer req.Body.Close()
		if req.Method != http.MethodPost {
			t.Fatalf("unexpected method %s", req.Method)
		}
		switch req.URL.Path {
		case accrualTypesPath:
			typesCalls++
			body, err := io.ReadAll(req.Body)
			if err != nil || string(body) != "{}" {
				t.Fatalf("unexpected dictionary request: %s (%v)", body, err)
			}
			return accrualTestResponse(http.StatusOK, `{"accrual_types":[
				{"id":102,"name":"SaleCommission"},{"id":103,"name":"Logistic"},
				{"id":104,"name":"ReturnFlowLogistic"},{"id":105,"name":"ItemPacking"},{"id":106,"name":"ItemCompensation"},
				{"id":107,"name":"RfbsClientDeliveryCharge"}
			]}`), nil
		case accrualPostingsPath:
			if retries < 2 {
				retries++
				response := accrualTestResponse(http.StatusTooManyRequests, `{"code":8}`)
				response.Header = make(http.Header)
				response.Header.Set("Retry-After", "0")
				return response, nil
			}
			var body struct {
				PostingNumbers []string `json:"posting_numbers"`
			}
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			postingsCalls++
			switch postingsCalls {
			case 1:
				if len(body.PostingNumbers) != accrualPostingLimit || body.PostingNumbers[0] != "posting-0" {
					t.Fatalf("unexpected first batch: %v", body.PostingNumbers)
				}
				return accrualTestResponse(http.StatusOK, `{"posting_accruals":[
					{"posting_number":"posting-0","accruals":[
						{"type_id":102,"quantity":2,"seller_price":{"amount":"50.00","currency":"RUB"},"accrued":{"amount":"-10.00","currency":"RUB"}},
						{"type_id":102,"quantity":1,"seller_price":{"amount":"-50.00","currency":"RUB"},"accrued":{"amount":"5.00","currency":"RUB"}},
						{"type_id":103,"accrued":{"amount":"-5.50","currency":"RUB"}},
						{"type_id":104,"accrued":{"amount":"-3.00","currency":"RUB"}},
						{"type_id":105,"accrued":{"amount":"-1.00","currency":"RUB"}},
						{"type_id":106,"accrued":{"amount":"2.00","currency":"RUB"}},
						{"type_id":107,"accrued":{"amount":"-4.00","currency":"RUB"}},
						{"type_id":108,"accrued":{"amount":"-0.50","currency":"RUB"}}
					]},
					{"posting_number":"posting-1","accruals":[]},
					{"posting_number":"posting-2","accruals":[
						{"type_id":102,"quantity":1,"seller_price":{"amount":"0","currency":"RUB"},"accrued":{"amount":"0","currency":"RUB"}}
					]},
					{"posting_number":"unexpected","accruals":[{"type_id":101,"accrued":{"amount":"1.00","currency":"RUB"}}]}
				]}`), nil
			case 2:
				if len(body.PostingNumbers) != 1 || body.PostingNumbers[0] != postings[accrualPostingLimit] {
					t.Fatalf("unexpected last batch: %v", body.PostingNumbers)
				}
				return accrualTestResponse(http.StatusOK, `{"posting_accruals":[]}`), nil
			}
		}
		t.Fatalf("unexpected request: %s", req.URL.Path)
		return nil, nil
	})

	result, err := NewFinanceAPI(client).GetCommissionsForPostings(append(postings, "", postings[0]))
	if err != nil {
		t.Fatal(err)
	}
	want := CommissionData{
		AccrualsForSale: 50, SaleCommission: -5, ProcessingAndDelivery: -5.5, RefundsAndCancellations: -3,
		ServicesAmount: -1, CompensationAmount: 2, MoneyTransfer: -4, OthersAmount: -0.5,
		ProfitAmount: 33, CommissionCurrency: "RUB",
	}
	if len(result) != 2 || result["posting-0"] != want || result["posting-2"] != (CommissionData{CommissionCurrency: "RUB"}) {
		t.Fatalf("unexpected commission data: %+v", result)
	}
	if typesCalls != 1 || postingsCalls != 2 {
		t.Fatalf("unexpected request counts: %d types, %d postings", typesCalls, postingsCalls)
	}
}

func TestGetCommissionsForPostingsRejectsInvalidTotals(t *testing.T) {
	for _, tc := range []struct {
		name, body, wantError string
		status                int
	}{
		{
			name: "mixed currencies", status: http.StatusOK, wantError: "多种货币",
			body: `{"posting_accruals":[{"posting_number":"posting-1","accruals":[
				{"type_id":1,"accrued":{"amount":"5","currency":"RUB"}},
				{"type_id":1,"accrued":{"amount":"5","currency":"CNY"}}
			]}]}`,
		},
		{
			name: "missing amount", status: http.StatusOK, wantError: "金额无效",
			body: `{"posting_accruals":[{"posting_number":"posting-1","accruals":[{"type_id":1,"accrued":{"currency":"RUB"}}]}]}`,
		},
		{
			name: "rate limit", status: http.StatusTooManyRequests, wantError: "429",
			body: `{"code":8,"message":"Too many requests"}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := NewClient(&Credentials{ClientID: "test-client", APIKey: "test-key"}, nil)
			client.GetHTTPClient().GetClient().Transport = orderRoundTripFunc(func(req *http.Request) (*http.Response, error) {
				defer req.Body.Close()
				switch req.URL.Path {
				case accrualTypesPath:
					return accrualTestResponse(http.StatusOK, `{"accrual_types":[{"id":1,"name":"Logistic"}]}`), nil
				case accrualPostingsPath:
					return accrualTestResponse(tc.status, tc.body), nil
				default:
					t.Fatalf("unexpected request: %s", req.URL.Path)
					return nil, nil
				}
			})
			result, err := NewFinanceAPI(client).GetCommissionsForPostings([]string{"posting-1"})
			if result != nil || err == nil || !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("got (%+v, %v), want error containing %q", result, err, tc.wantError)
			}
		})
	}
}

func accrualTestResponse(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body))}
}
