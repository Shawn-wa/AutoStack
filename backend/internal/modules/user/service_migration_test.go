package user

import (
	"context"
	"path/filepath"
	"testing"

	"autostack/internal/repository"
	companyRepo "autostack/internal/repository/company"
	userRepo "autostack/internal/repository/user"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMigrateExistingDataCommitsOrRollsBackBusinessUpdates(t *testing.T) {
	for _, failTable := range []string{"", "platform_auths", "orders", "order_daily_stats", "cash_flow_statements"} {
		name := failTable
		if name == "" {
			name = "success"
		}
		t.Run(name, func(t *testing.T) {
			db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "migration.db")), &gorm.Config{})
			if err != nil {
				t.Fatal(err)
			}
			connection, err := db.DB()
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = connection.Close() })
			if err := db.AutoMigrate(&Company{}, &User{}); err != nil {
				t.Fatal(err)
			}
			admin := User{Username: "legacy-admin", Email: "legacy@example.test", Role: RoleSuperAdmin}
			if err := db.Create(&admin).Error; err != nil {
				t.Fatal(err)
			}
			for _, table := range []string{"platform_auths", "orders", "order_daily_stats", "cash_flow_statements"} {
				if table == failTable {
					continue
				}
				if err := db.Exec("CREATE TABLE " + table + " (id INTEGER PRIMARY KEY, company_id BIGINT, created_by BIGINT)").Error; err != nil {
					t.Fatal(err)
				}
				if err := db.Exec("INSERT INTO " + table + " (id, company_id) VALUES (1, 0)").Error; err != nil {
					t.Fatal(err)
				}
			}
			svc := NewService(repository.NewTxManager(db), userRepo.NewUserRepository(db), companyRepo.NewCompanyRepository(db))
			err = svc.migrateExistingData(context.Background(), &admin)
			if (err != nil) != (failTable != "") {
				t.Fatalf("migration err=%v, failTable=%q", err, failTable)
			}
			var persisted User
			if err := db.First(&persisted, admin.ID).Error; err != nil {
				t.Fatal(err)
			}
			var companies int64
			if err := db.Model(&Company{}).Count(&companies).Error; err != nil {
				t.Fatal(err)
			}
			if failTable != "" {
				if persisted.CompanyID != 0 || companies != 0 {
					t.Fatalf("transaction did not roll back: companyID=%d companies=%d", persisted.CompanyID, companies)
				}
			} else if persisted.CompanyID < 90000000000 || companies != 1 {
				t.Fatalf("unexpected migrated company: %d, count=%d", persisted.CompanyID, companies)
			}
			for _, table := range []string{"platform_auths", "orders", "order_daily_stats", "cash_flow_statements"} {
				if table == failTable {
					continue
				}
				var companyID uint
				if err := db.Table(table).Select("company_id").Where("id = ?", 1).Scan(&companyID).Error; err != nil {
					t.Fatal(err)
				}
				if companyID != persisted.CompanyID {
					t.Fatalf("%s company_id=%d want=%d", table, companyID, persisted.CompanyID)
				}
			}
		})
	}
}
