package main

import (
	"fmt"
	"log"
	"strings"

	"autostack/internal/commonBase/database"
	"autostack/internal/config"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	if err := database.InitDB(&cfg.Database); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	db := database.GetDB()

	businessTables := []string{
		"platform_auths",
		"orders",
		"order_items",
		"orders_request_log",
		"cash_flow_statements",
		"order_daily_stats",
		"platform_sync_tasks",
		"platform_products",
		"product_mappings",
		"products",
		"product_suppliers",
		"warehouse_center_inventory",
		"stock_in_orders",
		"stock_in_order_items",
		"warehouses",
		"shipping_templates",
		"shipping_template_rules",
		"product_shipping_templates",
		"platform_product_shipping_templates",
	}

	// 1) 所有业务表补 company_id 列
	for _, table := range businessTables {
		if !hasTable(db, table) {
			log.Printf("表 %s 不存在，跳过", table)
			continue
		}
		if err := ensureCompanyIDColumn(db, table); err != nil {
			log.Fatalf("表 %s 添加 company_id 失败: %v", table, err)
		}
	}

	// 2) 按关系回填 company_id
	backfills := []string{
		// 平台授权
		`UPDATE platform_auths pa JOIN users u ON pa.user_id = u.id
		 SET pa.company_id = u.company_id
		 WHERE pa.company_id = 0 AND u.company_id <> 0`,
		`UPDATE platform_auths pa JOIN users u ON pa.created_by = u.id
		 SET pa.company_id = u.company_id
		 WHERE pa.company_id = 0 AND u.company_id <> 0`,

		// 订单与明细/日志/统计
		`UPDATE orders o JOIN platform_auths pa ON o.platform_auth_id = pa.id
		 SET o.company_id = pa.company_id
		 WHERE o.company_id = 0 AND pa.company_id <> 0`,
		`UPDATE order_items oi JOIN orders o ON oi.order_id = o.id
		 SET oi.company_id = o.company_id
		 WHERE oi.company_id = 0 AND o.company_id <> 0`,
		`UPDATE orders_request_log l JOIN platform_auths pa ON l.platform_auth_id = pa.id
		 SET l.company_id = pa.company_id
		 WHERE l.company_id = 0 AND pa.company_id <> 0`,
		`UPDATE order_daily_stats s JOIN users u ON s.user_id = u.id
		 SET s.company_id = u.company_id
		 WHERE s.company_id = 0 AND u.company_id <> 0`,

		// 现金流与同步任务
		`UPDATE cash_flow_statements c JOIN platform_auths pa ON c.platform_auth_id = pa.id
		 SET c.company_id = pa.company_id
		 WHERE c.company_id = 0 AND pa.company_id <> 0`,
		`UPDATE platform_sync_tasks t JOIN platform_auths pa ON t.platform_auth_id = pa.id
		 SET t.company_id = pa.company_id
		 WHERE t.company_id = 0 AND pa.company_id <> 0`,

		// 平台产品及映射
		`UPDATE platform_products pp JOIN platform_auths pa ON pp.platform_auth_id = pa.id
		 SET pp.company_id = pa.company_id
		 WHERE pp.company_id = 0 AND pa.company_id <> 0`,
		`UPDATE product_mappings pm JOIN platform_products pp ON pm.platform_product_id = pp.id
		 SET pm.company_id = pp.company_id
		 WHERE pm.company_id = 0 AND pp.company_id <> 0`,

		// 本地产品链路
		`UPDATE products p
		 JOIN (
		   SELECT pm.product_id, MIN(pm.company_id) AS company_id
		   FROM product_mappings pm
		   WHERE pm.company_id <> 0
		   GROUP BY pm.product_id
		 ) m ON p.id = m.product_id
		 SET p.company_id = m.company_id
		 WHERE p.company_id = 0`,
		`UPDATE product_suppliers s JOIN products p ON s.product_id = p.id
		 SET s.company_id = p.company_id
		 WHERE s.company_id = 0 AND p.company_id <> 0`,
		`UPDATE warehouse_center_inventory i JOIN products p ON i.product_id = p.id
		 SET i.company_id = p.company_id
		 WHERE i.company_id = 0 AND p.company_id <> 0`,
		`UPDATE stock_in_order_items si JOIN products p ON si.product_id = p.id
		 SET si.company_id = p.company_id
		 WHERE si.company_id = 0 AND p.company_id <> 0`,
		`UPDATE stock_in_orders so
		 JOIN (
		   SELECT stock_in_order_id, MIN(company_id) AS company_id
		   FROM stock_in_order_items
		   WHERE company_id <> 0
		   GROUP BY stock_in_order_id
		 ) x ON so.id = x.stock_in_order_id
		 SET so.company_id = x.company_id
		 WHERE so.company_id = 0`,
		`UPDATE warehouses w
		 JOIN (
		   SELECT warehouse_id, MIN(company_id) AS company_id
		   FROM stock_in_orders
		   WHERE company_id <> 0
		   GROUP BY warehouse_id
		 ) x ON w.id = x.warehouse_id
		 SET w.company_id = x.company_id
		 WHERE w.company_id = 0`,

		// 运费模板链路
		`UPDATE product_shipping_templates pst JOIN products p ON pst.product_id = p.id
		 SET pst.company_id = p.company_id
		 WHERE pst.company_id = 0 AND p.company_id <> 0`,
		`UPDATE platform_product_shipping_templates ppst JOIN platform_products pp ON ppst.platform_product_id = pp.id
		 SET ppst.company_id = pp.company_id
		 WHERE ppst.company_id = 0 AND pp.company_id <> 0`,
		`UPDATE shipping_template_rules r JOIN shipping_templates t ON r.template_id = t.id
		 SET r.company_id = t.company_id
		 WHERE r.company_id = 0 AND t.company_id <> 0`,
		`UPDATE shipping_templates t
		 JOIN (
		   SELECT shipping_template_id, MIN(company_id) AS company_id FROM (
		     SELECT shipping_template_id, company_id FROM product_shipping_templates WHERE company_id <> 0
		     UNION ALL
		     SELECT shipping_template_id, company_id FROM platform_product_shipping_templates WHERE company_id <> 0
		   ) u GROUP BY shipping_template_id
		 ) x ON t.id = x.shipping_template_id
		 SET t.company_id = x.company_id
		 WHERE t.company_id = 0`,
	}

	for _, sql := range backfills {
		// 仅当语句涉及的关键列存在时执行；缺列/缺表自动跳过
		if err := execBackfillSQL(db, sql); err != nil {
			log.Printf("回填语句执行失败（已跳过）: %v", err)
		}
	}

	// 3) 对仍为 0 的 company_id 做兜底（尽量与创建人/管理员企业保持一致）
	if err := execBackfillSQL(db, `
		UPDATE platform_auths pa
		JOIN users u ON pa.created_by = u.id
		SET pa.company_id = u.company_id
		WHERE pa.company_id = 0 AND u.company_id <> 0
	`); err != nil {
		log.Printf("platform_auths 兜底失败（已跳过）: %v", err)
	}
	if err := execBackfillSQL(db, `
		UPDATE orders o
		JOIN platform_auths pa ON o.platform_auth_id = pa.id
		SET o.company_id = pa.company_id
		WHERE o.company_id = 0 AND pa.company_id <> 0
	`); err != nil {
		log.Printf("orders 兜底失败（已跳过）: %v", err)
	}

	fmt.Println("全业务表 company_id 迁移完成")
}

func hasTable(db *gorm.DB, tableName string) bool {
	var count int64
	if err := db.Raw(
		`SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?`,
		tableName,
	).Scan(&count).Error; err != nil {
		log.Printf("检查表 %s 失败: %v", tableName, err)
		return false
	}
	return count > 0
}

func hasColumn(db *gorm.DB, tableName, column string) bool {
	var count int64
	if err := db.Raw(
		`SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?`,
		tableName, column,
	).Scan(&count).Error; err != nil {
		log.Printf("检查列 %s.%s 失败: %v", tableName, column, err)
		return false
	}
	return count > 0
}

func ensureCompanyIDColumn(db *gorm.DB, tableName string) error {
	if hasColumn(db, tableName, "company_id") {
		fmt.Printf("%s.company_id 已存在，跳过新增\n", tableName)
		return nil
	}

	if err := db.Exec("ALTER TABLE " + tableName + " ADD COLUMN company_id BIGINT UNSIGNED NOT NULL DEFAULT 0").Error; err != nil {
		return err
	}
	fmt.Printf("已新增 %s.company_id\n", tableName)

	// 尝试创建索引，不强制
	indexName := "idx_" + tableName + "_company_id"
	if len(indexName) > 60 {
		indexName = indexName[:60]
	}
	_ = db.Exec("ALTER TABLE " + tableName + " ADD INDEX " + indexName + " (company_id)").Error
	return nil
}

func execBackfillSQL(db *gorm.DB, sql string) error {
	sql = strings.TrimSpace(sql)
	if sql == "" {
		return nil
	}
	return db.Exec(sql).Error
}
