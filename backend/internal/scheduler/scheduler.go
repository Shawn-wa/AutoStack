package scheduler

import (
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"

	"autostack/internal/commonBase/database"
	"autostack/internal/modules/order"
	"autostack/internal/modules/product"
)

type commissionSyncResult struct {
	total   int
	updated int
}

var cronScheduler *cron.Cron

// Start 启动定时任务调度器
func Start() {
	cronScheduler = cron.New(cron.WithSeconds())

	_, err := cronScheduler.AddFunc("0 5 * * * *", syncAllAuthsOrdersAndCommission)
	if err != nil {
		log.Printf("[Scheduler] 添加订单同步任务失败: %v", err)
		return
	}

	_, err = cronScheduler.AddFunc("0 10 */4 * * *", calculateOrderTrendStats)
	if err != nil {
		log.Printf("[Scheduler] 添加订单走势统计任务失败: %v", err)
		return
	}

	_, err = cronScheduler.AddFunc("0 */1 * * * *", processPendingSyncTasks)
	if err != nil {
		log.Printf("[Scheduler] 添加同步任务扫描失败: %v", err)
		return
	}

	_, err = cronScheduler.AddFunc("0 20 1 * * *", cleanOldSyncTasks)
	if err != nil {
		log.Printf("[Scheduler] 添加任务清理失败: %v", err)
		return
	}

	_, err = cronScheduler.AddFunc("0 0 2 * * *", syncAllPlatformProducts)
	if err != nil {
		log.Printf("[Scheduler] 添加平台产品同步任务失败: %v", err)
		return
	}

	cronScheduler.Start()
	log.Println("[Scheduler] 定时任务调度器已启动")
	log.Println("[Scheduler] - 每小时第5分钟同步所有授权的订单和佣金")
	log.Println("[Scheduler] - 每4小时第10分钟统计订单走势数据")
	log.Println("[Scheduler] - 每1分钟扫描并执行待处理的同步任务")
	log.Println("[Scheduler] - 每天凌晨1:20清理3个月前的同步任务记录")
	log.Println("[Scheduler] - 每天凌晨2点同步所有平台产品")
}

// Stop 停止调度器
func Stop() {
	if cronScheduler != nil {
		cronScheduler.Stop()
		log.Println("[Scheduler] 定时任务调度器已停止")
	}
}

// TriggerSync 手动触发一次同步（供 API 调用）
func TriggerSync() {
	go syncAllAuthsOrdersAndCommission()
}

// TriggerTrendStats 手动触发一次订单走势统计（供 API 调用）
func TriggerTrendStats() {
	go calculateOrderTrendStats()
}

// TriggerSyncTasks 手动触发一次同步任务扫描（供 API 调用）
func TriggerSyncTasks() {
	go processPendingSyncTasks()
}

// TriggerProductSync 手动触发一次平台产品同步（供 API 调用）
func TriggerProductSync() {
	go syncAllPlatformProducts()
}

// syncAllAuthsOrdersAndCommission 同步所有活跃授权的订单和佣金
func syncAllAuthsOrdersAndCommission() {
	log.Println("[Scheduler] 开始执行定时同步任务...")

	db := database.GetDB()
	orderService := order.GetService()

	var auths []order.PlatformAuth
	if err := db.Where("status = ?", order.AuthStatusActive).Find(&auths).Error; err != nil {
		log.Printf("[Scheduler] 获取授权列表失败: %v", err)
		return
	}

	log.Printf("[Scheduler] 找到 %d 个活跃授权", len(auths))

	now := time.Now()
	since := now.AddDate(0, -3, 0)

	successCount := 0
	failCount := 0

	for _, auth := range auths {
		log.Printf("[Scheduler] 同步授权 ID=%d, 平台=%s, 店铺=%s", auth.ID, auth.Platform, auth.ShopName)

		result, err := orderService.SyncOrders(auth.ID, auth.CompanyID, since, now, true)
		if err != nil {
			log.Printf("[Scheduler] 同步订单失败 (授权ID=%d): %v", auth.ID, err)
			failCount++
			continue
		}

		log.Printf("[Scheduler] 订单同步完成 (授权ID=%d): 总计=%d, 新增=%d, 更新=%d",
			auth.ID, result.Total, result.Created, result.Updated)

		commissionSince := now.Add(-30 * 24 * time.Hour)
		commissionResult, err := syncCommissionForDeliveredOrders(auth.ID, auth.CompanyID, commissionSince, now)
		if err != nil {
			log.Printf("[Scheduler] 同步佣金失败 (授权ID=%d): %v", auth.ID, err)
		} else {
			log.Printf("[Scheduler] 佣金同步完成 (授权ID=%d): 处理=%d, 更新=%d",
				auth.ID, commissionResult.total, commissionResult.updated)
		}

		successCount++
	}

	log.Printf("[Scheduler] 定时同步任务完成: 成功=%d, 失败=%d", successCount, failCount)
}

// calculateOrderTrendStats 按企业统计订单走势数据并存储到 order_daily_stats 表
func calculateOrderTrendStats() {
	log.Println("[Scheduler] 开始执行订单走势统计任务...")

	db := database.GetDB()

	type CompanyInfo struct {
		ID uint
	}
	var companies []CompanyInfo
	if err := db.Table("companies").Select("id").Find(&companies).Error; err != nil {
		log.Printf("[Scheduler] 获取企业列表失败: %v", err)
		return
	}

	log.Printf("[Scheduler] 找到 %d 个企业需要统计订单走势", len(companies))

	days := 30
	endDate := time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour)
	startDate := endDate.AddDate(0, 0, -days)

	totalSaved := 0
	totalUpdated := 0

	for _, company := range companies {
		type DailyStats struct {
			Date     time.Time
			Currency string
			Count    int64
			Amount   float64
		}

		var dailyStats []DailyStats
		err := db.Table("orders").
			Select(`
				DATE(order_time) as date, 
				currency,
				COUNT(*) as count, 
				COALESCE(SUM(total_amount), 0) as amount
			`).
			Where("company_id = ? AND order_time >= ? AND order_time < ?", company.ID, startDate, endDate).
			Group("DATE(order_time), currency").
			Order("date ASC").
			Scan(&dailyStats).Error

		if err != nil {
			log.Printf("[Scheduler] 企业 %d 订单走势统计失败: %v", company.ID, err)
			continue
		}

		log.Printf("[Scheduler] 企业 %d 共有 %d 条日期统计数据", company.ID, len(dailyStats))

		for _, stat := range dailyStats {
			statDate := stat.Date

			var existing order.OrderDailyStat
			result := db.Where("company_id = ? AND stat_date = ? AND currency = ?", company.ID, statDate, stat.Currency).First(&existing)

			if result.Error != nil {
				newStat := order.OrderDailyStat{
					CompanyID:   company.ID,
					StatDate:    statDate,
					Currency:    stat.Currency,
					OrderCount:  stat.Count,
					OrderAmount: stat.Amount,
				}
				if err := db.Create(&newStat).Error; err != nil {
					log.Printf("[Scheduler] 创建统计记录失败: %v", err)
					continue
				}
				totalSaved++
			} else {
				db.Model(&existing).Updates(map[string]interface{}{
					"order_count":  stat.Count,
					"order_amount": stat.Amount,
				})
				totalUpdated++
			}
		}

		var totalOrders int64
		var totalAmount float64
		for _, stat := range dailyStats {
			totalOrders += stat.Count
			totalAmount += stat.Amount
		}

		log.Printf("[Scheduler] 企业 %d 订单走势统计完成: 近%d天订单=%d, 总金额=%.2f",
			company.ID, days, totalOrders, totalAmount)
	}

	log.Printf("[Scheduler] 订单走势统计任务完成: 新增=%d, 更新=%d", totalSaved, totalUpdated)
}

// syncCommissionForDeliveredOrders 只同步已签收订单的佣金
func syncCommissionForDeliveredOrders(authID, companyID uint, since, to time.Time) (*commissionSyncResult, error) {
	db := database.GetDB()

	var auth order.PlatformAuth
	if err := db.Where("id = ? AND company_id = ?", authID, companyID).First(&auth).Error; err != nil {
		return nil, fmt.Errorf("授权不存在")
	}

	adapter := order.GetAdapter(auth.Platform)
	if adapter == nil {
		return nil, fmt.Errorf("平台 %s 适配器未找到", auth.Platform)
	}

	credentials, err := order.Decrypt(auth.Credentials)
	if err != nil {
		return nil, fmt.Errorf("凭证解密失败: %w", err)
	}

	var orders []order.Order
	query := db.Model(&order.Order{}).
		Where("platform_auth_id = ?", authID).
		Where("status = ?", order.OrderStatusDelivered).
		Where("order_time >= ? AND order_time <= ?", since, to)
	if err := query.Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("查询订单失败: %w", err)
	}

	postingNumbers := make([]string, 0, len(orders))
	for _, ord := range orders {
		postingNumbers = append(postingNumbers, ord.PlatformOrderNo)
	}

	log.Printf("[Scheduler] 找到 %d 个已签收订单需要同步佣金", len(postingNumbers))

	if len(postingNumbers) == 0 {
		return &commissionSyncResult{total: 0, updated: 0}, nil
	}

	var commissions map[string]*order.CommissionData
	if adapterWithOrders, ok := adapter.(order.PlatformAdapterWithOrders); ok {
		commissions, err = adapterWithOrders.GetCommissionsForOrders(credentials, postingNumbers, auth.ID)
	} else if adapterWithLog, ok := adapter.(order.PlatformAdapterWithLog); ok {
		commissions, err = adapterWithLog.GetCommissionsWithLog(credentials, since, to, auth.ID)
	} else {
		commissions, err = adapter.GetCommissions(credentials, since, to)
	}
	if err != nil {
		return nil, fmt.Errorf("获取佣金数据失败: %w", err)
	}

	result := &commissionSyncResult{
		total: len(postingNumbers),
	}

	now := time.Now()
	for postingNumber, commData := range commissions {
		updateResult := db.Model(&order.Order{}).
			Where("platform_order_no = ? AND platform_auth_id = ?", postingNumber, authID).
			Updates(map[string]interface{}{
				"accruals_for_sale":         commData.AccrualsForSale,
				"sale_commission":           commData.SaleCommission,
				"processing_and_delivery":   commData.ProcessingAndDelivery,
				"refunds_and_cancellations": commData.RefundsAndCancellations,
				"services_amount":           commData.ServicesAmount,
				"compensation_amount":       commData.CompensationAmount,
				"money_transfer":            commData.MoneyTransfer,
				"others_amount":             commData.OthersAmount,
				"profit_amount":             commData.ProfitAmount,
				"commission_currency":       commData.CommissionCurrency,
				"commission_synced_at":      &now,
			})
		if updateResult.RowsAffected > 0 {
			result.updated++
		}
	}

	return result, nil
}

// processPendingSyncTasks 扫描并执行待处理的同步任务
func processPendingSyncTasks() {
	log.Println("[Scheduler] 开始扫描待处理的同步任务...")

	productService := product.GetService()
	if productService != nil {
		productService.ProcessPendingTasks()
	} else {
		log.Println("[Scheduler] 产品服务未初始化，跳过同步任务扫描")
	}

	log.Println("[Scheduler] 同步任务扫描完成")
}

// cleanOldSyncTasks 清理3个月前的同步任务记录
func cleanOldSyncTasks() {
	log.Println("[Scheduler] 开始清理旧的同步任务记录...")

	before := time.Now().AddDate(0, -3, 0)

	productService := product.GetService()
	if productService == nil {
		log.Println("[Scheduler] 产品服务未初始化，跳过清理任务")
		return
	}

	deleted, err := productService.CleanOldTasks(before)
	if err != nil {
		log.Printf("[Scheduler] 清理同步任务记录失败: %v", err)
		return
	}

	log.Printf("[Scheduler] 同步任务记录清理完成: 删除 %d 条", deleted)
}

// syncAllPlatformProducts 同步所有活跃授权的平台产品
func syncAllPlatformProducts() {
	log.Println("[Scheduler] 开始执行平台产品同步任务...")

	db := database.GetDB()
	productService := product.GetService()

	if productService == nil {
		log.Println("[Scheduler] 产品服务未初始化，跳过平台产品同步")
		return
	}

	var auths []order.PlatformAuth
	if err := db.Where("status = ?", order.AuthStatusActive).Find(&auths).Error; err != nil {
		log.Printf("[Scheduler] 获取授权列表失败: %v", err)
		return
	}

	log.Printf("[Scheduler] 找到 %d 个活跃授权需要同步产品", len(auths))

	successCount := 0
	failCount := 0

	for _, auth := range auths {
		log.Printf("[Scheduler] 同步产品 - 授权 ID=%d, 平台=%s, 店铺=%s", auth.ID, auth.Platform, auth.ShopName)

		err := productService.SyncPlatformProducts(auth.ID)
		if err != nil {
			log.Printf("[Scheduler] 同步产品失败 (授权ID=%d): %v", auth.ID, err)
			failCount++
			continue
		}

		log.Printf("[Scheduler] 产品同步完成 (授权ID=%d)", auth.ID)
		successCount++
	}

	log.Printf("[Scheduler] 平台产品同步任务完成: 成功=%d, 失败=%d", successCount, failCount)
}
