package scheduler

import (
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"autostack/internal/commonBase/database"
	"autostack/internal/config"
	"autostack/internal/modules/order"
	"gorm.io/gorm/logger"
)

// Exercise only local aggregation, without starting the scheduler or sync jobs.
func TestCalculateOrderTrendStatsReplacesLegacyCache(t *testing.T) {
	previousLocal := time.Local
	time.Local = time.FixedZone("scheduler-test-CST", 8*60*60)
	t.Cleanup(func() { time.Local = previousLocal })
	if err := database.InitDB(&config.DatabaseConfig{Driver: "sqlite", DSN: filepath.Join(t.TempDir(), "scheduler.db")}); err != nil {
		t.Fatal(err)
	}
	db := database.GetDB()
	db.Logger = logger.Default.LogMode(logger.Silent)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	if err := db.AutoMigrate(&order.Order{}, &order.OrderDailyStat{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("CREATE TABLE companies (id integer PRIMARY KEY)").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("INSERT INTO companies VALUES (1), (2)").Error; err != nil {
		t.Fatal(err)
	}
	today := time.Now().Format("2006-01-02")
	for _, stored := range []string{today, today + " 00:00:00+08:00"} {
		if err := db.Exec("INSERT INTO order_daily_stats (company_id, stat_date, currency, order_count, order_amount) VALUES (1, ?, 'RUB', 99, 999)", stored).Error; err != nil {
			t.Fatal(err)
		}
	}
	for company := 1; company <= 2; company++ {
		if err := db.Exec("INSERT INTO orders (company_id, platform_auth_id, platform, platform_order_no, status, order_time, currency, total_amount) VALUES (?, 1, 'ozon', ?, 'pending', ?, 'RUB', 10)", company, today+strconv.Itoa(company), today+" 00:00:00+00:00").Error; err != nil {
			t.Fatal(err)
		}
	}
	for run := 0; run < 2; run++ {
		calculateOrderTrendStats()
		var stats []order.OrderDailyStat
		if err := db.Order("company_id").Find(&stats).Error; err != nil {
			t.Fatal(err)
		}
		if len(stats) != 2 {
			t.Fatalf("run %d: expected one cache row per company, got %+v", run, stats)
		}
		for i, stat := range stats {
			if stat.CompanyID != uint(i+1) || stat.OrderCount != 1 || stat.OrderAmount != 10 || stat.StatDate.Format("2006-01-02") != today {
				t.Fatalf("run %d: unexpected stat %+v", run, stat)
			}
		}
	}
}
