package order

import (
	"maps"
	"path/filepath"
	"testing"
	"time"

	"autostack/internal/commonBase/database"
	"autostack/internal/config"
	productRepo "autostack/internal/repository/product"
)

func TestGetOrderSummaryGroupsByOrderItemSKU(t *testing.T) {
	if err := database.InitDB(&config.DatabaseConfig{
		Driver: "sqlite",
		DSN:    filepath.Join(t.TempDir(), "summary.db"),
	}); err != nil {
		t.Fatal(err)
	}
	db := database.GetDB()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&Order{}, &OrderItem{}, &productRepo.PlatformProduct{}, &productRepo.ProductMapping{}, &productRepo.Product{}); err != nil {
		t.Fatal(err)
	}

	orderTime := time.Date(2026, time.November, 12, 15, 0, 0, 0, time.UTC)
	order := Order{
		CompanyID:       1,
		PlatformAuthID:  1,
		Platform:        "ozon",
		PlatformOrderNo: "test-order",
		Status:          "pending",
		OrderTime:       &orderTime,
		Items: []OrderItem{
			{Sku: "offer-a", PlatformSku: "100", Quantity: 2, Price: 10, Currency: "RUB"},
			{Sku: "offer-a", PlatformSku: "101", Quantity: 1, Price: 10, Currency: "RUB"},
			{Sku: "offer-b", PlatformSku: "100", Quantity: 4, Price: 5, Currency: "RUB"},
		},
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatal(err)
	}

	items, err := NewService().GetOrderSummary(order.CompanyID, &OrderSummaryRequest{
		StartTime: "2026-11-01",
		EndTime:   "2026-11-12",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d summary rows, want 2", len(items))
	}

	type totals struct {
		quantity int
		amount   float64
	}
	got := make(map[string]totals, len(items))
	for _, item := range items {
		if len(item.PlatformSKUs) != 1 {
			t.Fatalf("want one platform SKU per unmapped item, got %v", item.PlatformSKUs)
		}
		got[item.PlatformSKUs[0]] = totals{quantity: item.Quantity, amount: item.Amount}
	}
	want := map[string]totals{
		"offer-a": {quantity: 3, amount: 30},
		"offer-b": {quantity: 4, amount: 20},
	}
	if !maps.Equal(got, want) {
		t.Fatalf("summary totals by SKU = %v, want %v", got, want)
	}
}
