package order

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGetDailyStatsSQLite(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { sqlDB.Close() })
	if err := db.Exec("CREATE TABLE orders (company_id integer, order_time datetime, currency text, total_amount real)").Error; err != nil {
		t.Fatal(err)
	}
	for _, row := range []struct {
		company        int
		date, currency string
		amount         float64
	}{
		{1, "2026-11-10 00:00:00+00:00", "RUB", 10},
		{1, "2026-11-10 12:30:00+00:00", "RUB", 20},
		{1, "2026-11-10 23:00:00-05:00", "USD", 5},
		{1, "2026-11-11 09:30:00+00:00", "RUB", 40},
		{1, "2026-11-09 23:59:59+00:00", "RUB", 100},
		{1, "2026-11-12 00:00:00+08:00", "RUB", 100},
		{2, "2026-11-10 12:00:00+00:00", "RUB", 100},
	} {
		if err := db.Exec("INSERT INTO orders VALUES (?, ?, ?, ?)", row.company, row.date, row.currency, row.amount).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Exec("INSERT INTO orders VALUES (1, '2026-11-10 01:00:00+00:00', NULL, 3), (1, '2026-11-10 02:00:00+00:00', '', 4)").Error; err != nil {
		t.Fatal(err)
	}
	location := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2026, 11, 10, 0, 0, 0, 0, location)
	repo := NewOrderDailyStatRepository(db)
	stats, err := repo.GetDailyStats(context.Background(), 1, start, start.AddDate(0, 0, 2))
	if err != nil {
		t.Fatal(err)
	}
	if len(stats) != 4 {
		t.Fatalf("got %#v; want four date/currency groups", stats)
	}
	want := map[string]struct {
		count  int64
		amount float64
	}{
		"2026-11-10/": {2, 7}, "2026-11-10/RUB": {2, 30}, "2026-11-10/USD": {1, 5}, "2026-11-11/RUB": {1, 40},
	}
	for _, stat := range stats {
		key := stat.Date.Format("2006-01-02") + "/" + stat.Currency
		expected, ok := want[key]
		if !ok || stat.Count != expected.count || stat.Amount != expected.amount {
			t.Errorf("unexpected group %#v", stat)
		}
		if stat.Date.Location() != location || stat.Date.Hour() != 0 {
			t.Errorf("date must be midnight in range location: %v", stat.Date)
		}
		delete(want, key)
	}
	if len(want) != 0 {
		t.Errorf("missing groups: %v", want)
	}
	stats, err = repo.GetDailyStats(context.Background(), 99, start, start.AddDate(0, 0, 2))
	if err != nil || len(stats) != 0 {
		t.Fatalf("empty company: stats=%v err=%v", stats, err)
	}
}

func TestParseDailyStatDate(t *testing.T) {
	location := time.FixedZone("UTC+8", 8*60*60)
	want := time.Date(2026, 11, 10, 0, 0, 0, 0, location)
	for _, value := range []interface{}{"2026-11-10", []byte("2026-11-10"), time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC)} {
		got, err := parseDailyStatDate(value, location)
		if err != nil || !got.Equal(want) || got.Location() != location {
			t.Errorf("parse %T: got %v, %v; want %v", value, got, err, want)
		}
	}
	for _, value := range []interface{}{nil, 42, "2026-02-30", []byte("invalid"), time.Time{}} {
		if _, err := parseDailyStatDate(value, location); err == nil {
			t.Errorf("expected invalid date error for %#v", value)
		}
	}
}

func TestDailyStatCacheDateRepresentations(t *testing.T) {
	for _, stored := range []string{"2026-11-10", "2026-11-10 00:00:00+00:00", "2026-11-10 00:00:00+08:00"} {
		t.Run(stored, func(t *testing.T) {
			db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
			if err != nil {
				t.Fatal(err)
			}
			sqlDB, err := db.DB()
			if err != nil {
				t.Fatal(err)
			}
			sqlDB.SetMaxOpenConns(1)
			t.Cleanup(func() { sqlDB.Close() })
			if err := db.AutoMigrate(&OrderDailyStat{}); err != nil {
				t.Fatal(err)
			}
			for _, date := range []string{"2026-11-09", stored, "2026-11-11"} {
				if err := db.Exec("INSERT INTO order_daily_stats (company_id, stat_date, currency, order_count, order_amount) VALUES (1, ?, 'RUB', 2, 30)", date).Error; err != nil {
					t.Fatal(err)
				}
			}
			start := time.Date(2026, 11, 10, 0, 0, 0, 0, time.FixedZone("UTC+8", 8*60*60))
			repo := NewOrderDailyStatRepository(db)
			stat, err := repo.FindByCompanyDateCurrency(context.Background(), 1, start, "RUB")
			if err != nil || stat.StatDate.Format("2006-01-02") != "2026-11-10" {
				t.Fatalf("lookup = %+v, %v", stat, err)
			}
			stats, err := repo.ListByCompanyAndDateRange(context.Background(), 1, start, start.AddDate(0, 0, 1), "RUB")
			if err != nil || len(stats) != 1 || stats[0].ID != stat.ID {
				t.Fatalf("range = %+v, %v", stats, err)
			}
		})
	}
}
