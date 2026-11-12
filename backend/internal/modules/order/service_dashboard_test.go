package order

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"autostack/internal/commonBase/database"
	"autostack/internal/config"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const dashboardCompanyID uint = 90000000001

// Global database and time.Local are deliberately used without parallel tests.
func dashboardTestDB(t *testing.T) (*gorm.DB, time.Time) {
	t.Helper()
	previousLocal := time.Local
	time.Local = time.FixedZone("dashboard-test-CST", 8*60*60)
	t.Cleanup(func() { time.Local = previousLocal })
	if err := database.InitDB(&config.DatabaseConfig{Driver: "sqlite", DSN: filepath.Join(t.TempDir(), "dashboard.db")}); err != nil {
		t.Fatal(err)
	}
	db := database.GetDB()
	db.Logger = logger.Default.LogMode(logger.Silent)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&Order{}, &OrderDailyStat{}, &PlatformAuth{}); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	return db, time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
}

func dashboardOrder(t *testing.T, db *gorm.DB, company uint, at time.Time, currency string, amount float64) {
	t.Helper()
	row := Order{CompanyID: company, PlatformAuthID: company, Platform: "ozon", PlatformOrderNo: fmt.Sprintf("%d-%d-%s-%g", company, at.UnixNano(), currency, amount), Status: OrderStatusPending, OrderTime: &at, Currency: currency, TotalAmount: amount}
	if err := db.Create(&row).Error; err != nil {
		t.Fatal(err)
	}
}

func dashboardCache(t *testing.T, db *gorm.DB) []OrderDailyStat {
	t.Helper()
	var rows []OrderDailyStat
	if err := db.Order("id").Find(&rows).Error; err != nil {
		t.Fatal(err)
	}
	return rows
}

func TestDashboardTrendRefreshAndFallback(t *testing.T) {
	db, today := dashboardTestDB(t)
	first := today.AddDate(0, 0, -6)
	// The +08 first-day midnight must not be rounded to UTC midnight or moved
	// to the previous calendar date by SQLite's DATE(time-with-offset).
	dashboardOrder(t, db, dashboardCompanyID, first, "RUB", 10)
	dashboardOrder(t, db, dashboardCompanyID, first.Add(time.Hour), "RUB", 20)
	dashboardOrder(t, db, dashboardCompanyID, first.Add(2*time.Hour), "USD", 7)
	dashboardOrder(t, db, dashboardCompanyID, today.Add(3*time.Hour), "RUB", 40)
	dashboardOrder(t, db, dashboardCompanyID+1, first.Add(time.Hour), "RUB", 999)
	dashboardOrder(t, db, dashboardCompanyID, today.AddDate(0, 0, 1), "RUB", 800)
	dashboardOrder(t, db, dashboardCompanyID, today.AddDate(0, 0, -30), "RUB", 700)
	service := NewService()
	want := make([]OrderTrendItem, 7)
	for i := range want {
		want[i].Date = first.AddDate(0, 0, i).Format("2006-01-02")
	}
	want[0].Count, want[0].Amount = 2, 30
	want[6].Count, want[6].Amount = 1, 40
	fallback, err := service.GetOrderTrend(dashboardCompanyID, 7, "RUB")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(fallback.Items, want) {
		t.Fatalf("fallback = %+v, want %+v", fallback.Items, want)
	}
	// Seed another company's cache first, proving force refresh does not delete it.
	if err := service.InitOrderTrendStats(dashboardCompanyID+1, false); err != nil {
		t.Fatal(err)
	}
	if err := service.InitOrderTrendStats(dashboardCompanyID, false); err != nil {
		t.Fatal(err)
	}
	rows := dashboardCache(t, db)
	if len(rows) != 4 {
		t.Fatalf("cache has %d rows, want 4 date/currency/company groups", len(rows))
	}
	for _, force := range []bool{false, true, true} {
		if err := service.InitOrderTrendStats(dashboardCompanyID, force); err != nil {
			t.Fatal(err)
		}
		if got := dashboardCache(t, db); len(got) != 4 {
			t.Fatalf("force=%v cache row count = %d, want 4", force, len(got))
		}
		cached, err := service.GetOrderTrend(dashboardCompanyID, 7, "")
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(cached, fallback) {
			t.Fatalf("force=%v cached = %+v, fallback = %+v", force, cached.Items, fallback.Items)
		}
	}
	usd, err := service.GetOrderTrend(dashboardCompanyID, 7, "USD")
	if err != nil {
		t.Fatal(err)
	}
	if usd.Items[0].Count != 1 || usd.Items[0].Amount != 7 || usd.Items[6].Count != 0 {
		t.Fatalf("USD trend = %+v", usd.Items)
	}
	other, err := service.GetOrderTrend(dashboardCompanyID+1, 7, "RUB")
	if err != nil {
		t.Fatal(err)
	}
	if other.Items[0].Amount != 999 {
		t.Fatalf("other company cache altered: %+v", other.Items)
	}
	// A nonforced call must retain cached values; a forced refresh must recompute.
	dashboardOrder(t, db, dashboardCompanyID, first.Add(4*time.Hour), "RUB", 5)
	if err := service.InitOrderTrendStats(dashboardCompanyID, false); err != nil {
		t.Fatal(err)
	}
	unchanged, err := service.GetOrderTrend(dashboardCompanyID, 7, "RUB")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(unchanged, fallback) {
		t.Fatal("nonforced initialization changed existing cache")
	}
	if err := service.InitOrderTrendStats(dashboardCompanyID, true); err != nil {
		t.Fatal(err)
	}
	refreshed, err := service.GetOrderTrend(dashboardCompanyID, 7, "RUB")
	if err != nil {
		t.Fatal(err)
	}
	if refreshed.Items[0].Count != 3 || refreshed.Items[0].Amount != 35 {
		t.Fatalf("refresh = %+v", refreshed.Items[0])
	}
}

func TestDashboardRefreshFailurePreservesCache(t *testing.T) {
	for _, scenario := range []struct{ name, sql, errorText string }{
		{"query", "DROP TABLE orders", "no such table"},
		{"delete", "CREATE TRIGGER fail_cache_delete BEFORE DELETE ON order_daily_stats BEGIN SELECT RAISE(ABORT, 'injected delete failure'); END", "injected delete failure"},
		{"insert", "CREATE TRIGGER fail_cache_insert BEFORE INSERT ON order_daily_stats BEGIN SELECT RAISE(ABORT, 'injected insert failure'); END", "injected insert failure"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			db, today := dashboardTestDB(t)
			dashboardOrder(t, db, dashboardCompanyID, today.Add(time.Hour), "RUB", 15)
			old := []OrderDailyStat{{CompanyID: dashboardCompanyID, StatDate: today, Currency: "RUB", OrderCount: 9, OrderAmount: 99}, {CompanyID: dashboardCompanyID + 1, StatDate: today, Currency: "USD", OrderCount: 8, OrderAmount: 88}}
			if err := db.Create(&old).Error; err != nil {
				t.Fatal(err)
			}
			before := dashboardCache(t, db)
			if err := db.Exec(scenario.sql).Error; err != nil {
				t.Fatal(err)
			}
			if err := NewService().InitOrderTrendStats(dashboardCompanyID, true); err == nil || !strings.Contains(err.Error(), scenario.errorText) {
				t.Fatalf("refresh error = %v, want %q", err, scenario.errorText)
			}
			if after := dashboardCache(t, db); !reflect.DeepEqual(before, after) {
				t.Fatalf("failed refresh changed cache: before=%+v after=%+v", before, after)
			}
		})
	}
}

func TestDashboardEmptyCompany(t *testing.T) {
	db, _ := dashboardTestDB(t)
	service := NewService()
	for _, force := range []bool{false, true} {
		if err := service.InitOrderTrendStats(dashboardCompanyID, force); err != nil {
			t.Fatal(err)
		}
	}
	if len(dashboardCache(t, db)) != 0 {
		t.Fatal("empty company created cache rows")
	}
	trend, err := service.GetOrderTrend(dashboardCompanyID, 7, "RUB")
	if err != nil {
		t.Fatal(err)
	}
	if len(trend.Items) != 7 {
		t.Fatalf("got %d trend days", len(trend.Items))
	}
	for _, item := range trend.Items {
		if item.Count != 0 || item.Amount != 0 {
			t.Fatalf("nonzero empty trend: %+v", item)
		}
	}
	stats, err := service.GetDashboardStats(dashboardCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	if stats.TotalOrders != 0 || stats.TotalAuths != 0 {
		t.Fatalf("nonzero empty stats: %+v", stats)
	}
}

func TestDashboardReportsDatabaseErrors(t *testing.T) {
	for _, table := range []string{"orders", "platform_auths", "order_daily_stats"} {
		t.Run(table, func(t *testing.T) {
			db, _ := dashboardTestDB(t)
			if err := db.Exec("DROP TABLE " + table).Error; err != nil {
				t.Fatal(err)
			}
			if table == "order_daily_stats" {
				if err := NewService().InitOrderTrendStats(dashboardCompanyID, false); err == nil {
					t.Fatal("cache count error was swallowed")
				}
			} else {
				if _, err := NewService().GetDashboardStats(dashboardCompanyID); err == nil {
					t.Fatal("dashboard query error was swallowed")
				}
				if table == "orders" {
					if _, err := NewService().GetOrderTrend(dashboardCompanyID, 7, "RUB"); err == nil {
						t.Fatal("fallback query error was swallowed")
					}
				}
			}
		})
	}
}

func TestDashboardInitialRefreshDateScan(t *testing.T) {
	db, today := dashboardTestDB(t)
	dashboardOrder(t, db, dashboardCompanyID, today.Add(12*time.Hour), "RUB", 123)
	if err := NewService().InitOrderTrendStats(dashboardCompanyID, false); err != nil {
		t.Fatalf("initial refresh must accept the SQL date aggregate: %v", err)
	}
	rows := dashboardCache(t, db)
	if len(rows) != 1 || rows[0].StatDate.Format("2006-01-02") != today.Format("2006-01-02") || rows[0].OrderCount != 1 || rows[0].OrderAmount != 123 {
		t.Fatalf("initial cache = %+v", rows)
	}
}

func TestDashboardReadsDateOnlyCache(t *testing.T) {
	db, today := dashboardTestDB(t)
	first := today.AddDate(0, 0, -6)
	for _, row := range []struct {
		day    time.Time
		amount float64
	}{
		{first.AddDate(0, 0, -1), 1000}, {first, 10}, {today, 20}, {today.AddDate(0, 0, 1), 2000},
	} {
		if err := db.Exec("INSERT INTO order_daily_stats (company_id, stat_date, currency, order_count, order_amount) VALUES (?, ?, ?, ?, ?)", dashboardCompanyID, row.day.Format("2006-01-02"), "RUB", 1, row.amount).Error; err != nil {
			t.Fatal(err)
		}
	}
	got, err := NewService().GetOrderTrend(dashboardCompanyID, 7, "RUB")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Items) != 7 || got.Items[0] != (OrderTrendItem{Date: first.Format("2006-01-02"), Count: 1, Amount: 10}) || got.Items[6] != (OrderTrendItem{Date: today.Format("2006-01-02"), Count: 1, Amount: 20}) {
		t.Fatalf("date-only cache boundaries = %+v", got.Items)
	}
	for _, item := range got.Items[1:6] {
		if item.Count != 0 || item.Amount != 0 {
			t.Fatalf("unexpected date-only cache item: %+v", item)
		}
	}
}

func TestDashboardRefreshIncludesUTCMidnight(t *testing.T) {
	db, today := dashboardTestDB(t)
	first := today.AddDate(0, 0, -29)
	utcMidnight := time.Date(first.Year(), first.Month(), first.Day(), 0, 0, 0, 0, time.UTC)
	dashboardOrder(t, db, dashboardCompanyID, utcMidnight, "RUB", 17)
	service := NewService()
	before, err := service.GetOrderTrend(dashboardCompanyID, 30, "RUB")
	if err != nil {
		t.Fatal(err)
	}
	want := OrderTrendItem{Date: first.Format("2006-01-02"), Count: 1, Amount: 17}
	if before.Items[0] != want {
		t.Fatalf("UTC midnight fallback = %+v, want %+v", before.Items[0], want)
	}
	if err := service.InitOrderTrendStats(dashboardCompanyID, true); err != nil {
		t.Fatal(err)
	}
	after, err := service.GetOrderTrend(dashboardCompanyID, 30, "RUB")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("UTC midnight refresh changed trend: before=%+v after=%+v", before, after)
	}
	if rows := dashboardCache(t, db); len(rows) != 1 || rows[0].OrderCount != 1 {
		t.Fatalf("UTC midnight cache = %+v", rows)
	}
}

func TestDashboardUnknownCurrencyRemainsSeparate(t *testing.T) {
	db, today := dashboardTestDB(t)
	at := today.Add(12 * time.Hour)
	dashboardOrder(t, db, dashboardCompanyID, at, "RUB", 30)
	// SQL insertion preserves legacy NULL and empty-string currency values;
	// the cache entity must not apply its RUB default to these unknown groups.
	for i, currency := range []any{nil, ""} {
		if err := db.Exec("INSERT INTO orders (company_id, platform_auth_id, platform, platform_order_no, status, order_time, currency, total_amount) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", dashboardCompanyID, dashboardCompanyID, "ozon", fmt.Sprintf("unknown-%d", i), OrderStatusPending, at, currency, 10+i).Error; err != nil {
			t.Fatal(err)
		}
	}
	service := NewService()
	before, err := service.GetOrderTrend(dashboardCompanyID, 7, "RUB")
	if err != nil {
		t.Fatal(err)
	}
	if before.Items[6].Count != 1 || before.Items[6].Amount != 30 {
		t.Fatalf("unknown currency leaked into RUB fallback: %+v", before.Items)
	}
	for _, force := range []bool{false, true} {
		if err := service.InitOrderTrendStats(dashboardCompanyID, force); err != nil {
			t.Fatalf("refresh force=%v: %v", force, err)
		}
		rows := dashboardCache(t, db)
		if len(rows) != 2 {
			t.Fatalf("got %d currency groups, want 2: %+v", len(rows), rows)
		}
		for _, row := range rows {
			switch row.Currency {
			case "":
				if row.OrderCount != 2 || row.OrderAmount != 21 {
					t.Fatalf("unknown currency aggregate = %+v", row)
				}
			case "RUB":
				if row.OrderCount != 1 || row.OrderAmount != 30 {
					t.Fatalf("RUB aggregate = %+v", row)
				}
			default:
				t.Fatalf("unexpected currency: %+v", row)
			}
		}
		after, err := service.GetOrderTrend(dashboardCompanyID, 7, "RUB")
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(before, after) {
			t.Fatalf("currency normalization changed RUB trend: before=%+v after=%+v", before, after)
		}
	}
}
