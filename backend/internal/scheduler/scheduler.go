package scheduler

import (
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"

	"autostack/internal/commonBase/database"
	"autostack/internal/modules/order"
)

// commissionSyncResult 佣金同步结果
type commissionSyncResult struct {
	total   int
	updated int
}

var cronScheduler *cron.Cron

// Start 启动定时任务调度器
func Start() {
	cronScheduler = cron.New(cron.WithSeconds())

	// 每小时同步订单和佣金（每小时的第 5 分钟执行，避免整点高峰）
	_, err := cronScheduler.AddFunc("0 5 * * * *", syncAllAuthsOrdersAndCommission)
	if err != nil {
		log.Printf("[Scheduler] 添加定时任务失败: %v", err)
		return
	}

	cronScheduler.Start()
	log.Println("[Scheduler] 定时任务调度器已启动")
	log.Println("[Scheduler] - 每小时第5分钟同步所有授权的订单和佣金")
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

// syncAllAuthsOrdersAndCommission 同步所有活跃授权的订单和佣金
func syncAllAuthsOrdersAndCommission() {
	log.Println("[Scheduler] 开始执行定时同步任务...")

	db := database.GetDB()
	orderService := order.NewService()

	// 获取所有活跃的授权
	var auths []order.PlatformAuth
	if err := db.Where("status = ?", order.AuthStatusActive).Find(&auths).Error; err != nil {
		log.Printf("[Scheduler] 获取授权列表失败: %v", err)
		return
	}

	log.Printf("[Scheduler] 找到 %d 个活跃授权", len(auths))

	// 同步时间范围：最近2小时（确保覆盖）
	now := time.Now()
	since := now.Add(-2 * time.Hour)

	successCount := 0
	failCount := 0

	for _, auth := range auths {
		log.Printf("[Scheduler] 同步授权 ID=%d, 平台=%s, 店铺=%s", auth.ID, auth.Platform, auth.ShopName)

		// 同步订单
		result, err := orderService.SyncOrders(auth.ID, auth.UserID, since, now)
		if err != nil {
			log.Printf("[Scheduler] 同步订单失败 (授权ID=%d): %v", auth.ID, err)
			failCount++
			continue
		}

		log.Printf("[Scheduler] 订单同步完成 (授权ID=%d): 总计=%d, 新增=%d, 更新=%d",
			auth.ID, result.Total, result.Created, result.Updated)

		// 同步佣金：只同步已签收的订单（最近30天内签收的）
		commissionSince := now.Add(-30 * 24 * time.Hour)
		commissionResult, err := syncCommissionForDeliveredOrders(auth.ID, auth.UserID, commissionSince, now)
		if err != nil {
			log.Printf("[Scheduler] 同步佣金失败 (授权ID=%d): %v", auth.ID, err)
			// 佣金同步失败不影响整体
		} else {
			log.Printf("[Scheduler] 佣金同步完成 (授权ID=%d): 处理=%d, 更新=%d",
				auth.ID, commissionResult.total, commissionResult.updated)
		}

		successCount++
	}

	log.Printf("[Scheduler] 定时同步任务完成: 成功=%d, 失败=%d", successCount, failCount)
}

// syncCommissionForDeliveredOrders 只同步已签收订单的佣金
func syncCommissionForDeliveredOrders(authID, userID uint, since, to time.Time) (*commissionSyncResult, error) {
	db := database.GetDB()

	// 获取授权信息
	var auth order.PlatformAuth
	if err := db.Where("id = ? AND user_id = ?", authID, userID).First(&auth).Error; err != nil {
		return nil, fmt.Errorf("授权不存在")
	}

	// 获取平台适配器
	adapter := order.GetAdapter(auth.Platform)
	if adapter == nil {
		return nil, fmt.Errorf("平台 %s 适配器未找到", auth.Platform)
	}

	// 解密凭证
	credentials, err := order.Decrypt(auth.Credentials)
	if err != nil {
		return nil, fmt.Errorf("凭证解密失败: %w", err)
	}

	// 只查询已签收的订单（status = 'delivered'）
	var orders []order.Order
	query := db.Model(&order.Order{}).
		Where("platform_auth_id = ?", authID).
		Where("status = ?", order.OrderStatusDelivered).
		Where("order_time >= ? AND order_time <= ?", since, to)
	if err := query.Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("查询订单失败: %w", err)
	}

	// 提取订单号列表
	postingNumbers := make([]string, 0, len(orders))
	for _, ord := range orders {
		postingNumbers = append(postingNumbers, ord.PlatformOrderNo)
	}

	log.Printf("[Scheduler] 找到 %d 个已签收订单需要同步佣金", len(postingNumbers))

	if len(postingNumbers) == 0 {
		return &commissionSyncResult{total: 0, updated: 0}, nil
	}

	// 使用适配器获取佣金
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

	// 批量更新订单佣金
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
