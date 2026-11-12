package ozon

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestProductListTotalItems(t *testing.T) {
	for _, tt := range []struct {
		name   string
		counts string
		want   int64
	}{
		{name: "current", counts: `"total_items":3000000000`, want: 3000000000},
		{name: "legacy", counts: `"total":42`, want: 42},
		{name: "current takes precedence", counts: `"total":42,"total_items":0`, want: 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			body := `{"result":{"items":[{"offer_id":"offer-1","product_id":123,"sku":456}],"last_id":"next",` + tt.counts + `}}`
			client := NewClient(&Credentials{ClientID: "test-client", APIKey: "test-key"}, nil)
			client.GetHTTPClient().GetClient().Transport = orderRoundTripFunc(func(req *http.Request) (*http.Response, error) {
				defer req.Body.Close()
				var request ProductListRequest
				if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
					t.Fatal(err)
				}
				if req.Method != http.MethodPost || req.URL.Path != "/v3/product/list" ||
					request.Limit != 100 || request.LastID != "previous" || request.Filter.Visibility != "ALL" {
					t.Fatalf("unexpected request: %s %s %+v", req.Method, req.URL.Path, request)
				}
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body))}, nil
			})
			response, err := client.GetProductList(100, "previous")
			if err != nil {
				t.Fatal(err)
			}
			result := response.Result
			if result.Total != tt.want || result.LastID != "next" || len(result.Items) != 1 {
				t.Fatalf("unexpected product list result: %+v", result)
			}
			if want := (ProductItem{OfferID: "offer-1", ProductID: 123, SKU: 456}); result.Items[0] != want {
				t.Errorf("product = %+v, want %+v", result.Items[0], want)
			}
		})
	}
}

func TestProductInfoPriceAndStock(t *testing.T) {
	body := `{"items":[{
		"id":123,"sku":456,"offer_id":"offer-1","price":"19.99","currency_code":"CNY",
		"declared_price":{"amount":"12.00","currency":"CNY"},
		"primary_image":["https://example.test/product.jpg"],
		"stocks":{"has_stock":true,"stocks":[{"present":10,"reserved":2,"sku":456,"source":"fbs"}]}
	}]}`
	client := NewClient(&Credentials{ClientID: "test-client", APIKey: "test-key"}, nil)
	client.GetHTTPClient().GetClient().Transport = orderRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		defer req.Body.Close()
		var request ProductInfoRequest
		if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		if req.Method != http.MethodPost || req.URL.Path != "/v3/product/info/list" ||
			len(request.OfferID) != 1 || request.OfferID[0] != "offer-1" {
			t.Fatalf("unexpected request: %s %s %+v", req.Method, req.URL.Path, request)
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body))}, nil
	})
	response, err := client.GetProductInfo([]string{"offer-1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Items) != 1 {
		t.Fatalf("unexpected product details: %+v", response.Items)
	}
	item := response.Items[0]
	if item.Price != "19.99" || item.CurrencyCode != "CNY" || item.SKU != 456 ||
		item.GetPrimaryImageURL() != "https://example.test/product.jpg" {
		t.Errorf("unexpected product details: %+v", item)
	}
	if want := (ProductDeclaredPrice{Amount: "12.00", Currency: "CNY"}); item.DeclaredPrice == nil || *item.DeclaredPrice != want {
		t.Errorf("declared price = %+v, want %+v", item.DeclaredPrice, want)
	}
	if want := (ProductStockItem{Present: 10, Reserved: 2, SKU: 456, Source: "fbs"}); len(item.Stocks.Stocks) != 1 || item.Stocks.Stocks[0] != want {
		t.Errorf("stocks = %+v, want %+v", item.Stocks.Stocks, want)
	}
}
