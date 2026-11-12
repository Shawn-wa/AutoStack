package ozon

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestGetOrderListV4(t *testing.T) {
	client := NewClient(&Credentials{ClientID: "test-client", APIKey: "test-key"}, nil)
	client.GetHTTPClient().GetClient().Transport = orderRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost || req.URL.Path != "/v4/posting/fbs/list" {
			t.Fatalf("unexpected request: %s %s", req.Method, req.URL.Path)
		}
		var body struct {
			SortDir string            `json:"sort_dir"`
			Filter  map[string]string `json:"filter"`
			Limit   int               `json:"limit"`
			Cursor  string            `json:"cursor"`
		}
		defer req.Body.Close()
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.SortDir != "ASC" || body.Cursor != "page-1" || body.Limit != 100 ||
			body.Filter["since"] != "2026-11-09T16:00:00Z" || body.Filter["to"] != "2026-11-10T16:00:00Z" {
			t.Fatalf("unexpected body: %+v", body)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body: io.NopCloser(strings.NewReader(`{
				"cursor": "page-2",
				"has_next": true,
				"postings": [{
					"posting_number": "test-posting",
					"shipment_date_without_delay": "2026-11-11T12:00:00Z",
					"products": [{"sku": 123, "quantity": 2, "price": {"amount": "12.50", "currency": "RUB"}}]
				}]
			}`)),
		}, nil
	})

	since := time.Date(2026, 11, 10, 0, 0, 0, 0, time.FixedZone("UTC+8", 8*60*60))
	result, err := NewOrderAPI(client).GetOrderList(since, since.Add(24*time.Hour), "page-1", 100)
	if err != nil {
		t.Fatal(err)
	}
	if !result.HasNext || result.Cursor != "page-2" || len(result.Postings) != 1 {
		t.Fatalf("unexpected page: %+v", result)
	}
	posting := result.Postings[0]
	if posting.PostingNumber != "test-posting" || posting.ShipmentDateWithoutDelay != "2026-11-11T12:00:00Z" || len(posting.Products) != 1 {
		t.Fatalf("unexpected posting: %+v", posting)
	}
	want := Product{Sku: 123, Quantity: 2, Price: "12.50", CurrencyCode: "RUB"}
	if got := posting.Products[0]; got != want {
		t.Errorf("product = %+v, want %+v", got, want)
	}
}

func TestOrderDetailV3Price(t *testing.T) {
	var response OrderDetailResponse
	body := `{"result":{"products":[{"price":"12.50","currency_code":"RUB"}]}}`
	if err := json.Unmarshal([]byte(body), &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Result.Products) != 1 {
		t.Fatalf("unexpected products: %+v", response.Result.Products)
	}
	want := Product{Price: "12.50", CurrencyCode: "RUB"}
	if got := response.Result.Products[0]; got != want {
		t.Errorf("product = %+v, want %+v", got, want)
	}
}

type orderRoundTripFunc func(*http.Request) (*http.Response, error)

func (f orderRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
