package product

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"

	"autostack/internal/commonBase/database"
	"autostack/internal/modules/apiClient/platform"
	"autostack/internal/modules/apiClient/platform/ozon"
	"autostack/internal/modules/order"
)

// Service 产品服务
type Service struct{}

// ListProducts 获取本地产品列表
func (s *Service) ListProducts(page, pageSize int) ([]Product, int64, error) {
	var products []Product
	var total int64
	db := database.GetDB()

	if err := db.Model(&Product{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// CreateProduct 创建本地产品
func (s *Service) CreateProduct(req CreateProductRequest) (*Product, error) {
	db := database.GetDB()

	// 检查SKU是否已存在
	var count int64
	db.Model(&Product{}).Where("sku = ?", req.SKU).Count(&count)
	if count > 0 {
		return nil, errors.New("SKU已存在")
	}

	product := &Product{
		SKU:        req.SKU,
		Name:       req.Name,
		Image:      req.Image,
		CostPrice:  req.CostPrice,
		Weight:     req.Weight,
		Dimensions: req.Dimensions,
	}

	if err := db.Create(product).Error; err != nil {
		return nil, err
	}

	return product, nil
}

// UpdateProduct 更新本地产品
func (s *Service) UpdateProduct(id uint, req UpdateProductRequest) (*Product, error) {
	db := database.GetDB()
	var product Product

	if err := db.First(&product, id).Error; err != nil {
		return nil, errors.New("产品不存在")
	}

	updates := map[string]interface{}{
		"name":       req.Name,
		"image":      req.Image,
		"cost_price": req.CostPrice,
		"weight":     req.Weight,
		"dimensions": req.Dimensions,
	}

	if err := db.Model(&product).Updates(updates).Error; err != nil {
		return nil, err
	}

	return &product, nil
}

// DeleteProduct 删除本地产品
func (s *Service) DeleteProduct(id uint) error {
	db := database.GetDB()
	
	// 检查是否有映射关联
	var count int64
	db.Model(&ProductMapping{}).Where("product_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("该产品已关联平台产品，无法删除")
	}

	return db.Delete(&Product{}, id).Error
}

// ListPlatformProducts 获取平台产品列表
func (s *Service) ListPlatformProducts(platformAuthID uint, page, pageSize int) ([]PlatformProduct, int64, error) {
	var products []PlatformProduct
	var total int64
	db := database.GetDB()

	query := db.Model(&PlatformProduct{})
	if platformAuthID > 0 {
		query = query.Where("platform_auth_id = ?", platformAuthID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("ProductMapping.Product").Offset(offset).Limit(pageSize).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// MapProduct 关联产品
func (s *Service) MapProduct(req MapProductRequest) error {
	db := database.GetDB()

	// 检查平台产品是否存在
	var platformProduct PlatformProduct
	if err := db.First(&platformProduct, req.PlatformProductID).Error; err != nil {
		return errors.New("平台产品不存在")
	}

	// 检查本地产品是否存在
	var product Product
	if err := db.First(&product, req.ProductID).Error; err != nil {
		return errors.New("本地产品不存在")
	}

	// 检查是否已有映射
	var mapping ProductMapping
	err := db.Where("platform_product_id = ?", req.PlatformProductID).First(&mapping).Error
	if err == nil {
		// 更新现有映射
		mapping.ProductID = req.ProductID
		return db.Save(&mapping).Error
	}

	// 创建新映射
	newMapping := ProductMapping{
		PlatformProductID: req.PlatformProductID,
		ProductID:         req.ProductID,
	}

	return db.Create(&newMapping).Error
}

// UnmapProduct 解除关联
func (s *Service) UnmapProduct(platformProductID uint) error {
	db := database.GetDB()
	return db.Where("platform_product_id = ?", platformProductID).Delete(&ProductMapping{}).Error
}

// SyncPlatformProducts 同步平台产品
func (s *Service) SyncPlatformProducts(platformAuthID uint) error {
	db := database.GetDB()
	
	// 获取授权信息
	var auth order.PlatformAuth
	if err := db.First(&auth, platformAuthID).Error; err != nil {
		return errors.New("授权信息不存在")
	}

	if auth.Platform != order.PlatformOzon {
		return errors.New("目前仅支持 Ozon 平台同步")
	}

	// 解密凭证
	credentials, err := order.Decrypt(auth.Credentials)
	if err != nil {
		return fmt.Errorf("解密凭证失败: %v", err)
	}

	// 创建客户端
	client, err := createOzonClient(credentials, platformAuthID)
	if err != nil {
		return err
	}

	// 拉取产品列表
	// 分页拉取所有产品
	limit := 1000
	lastID := ""
	var allItems []ozon.ProductItem

	for {
		resp, err := client.GetProductList(limit, lastID)
		if err != nil {
			return fmt.Errorf("拉取产品列表失败: %v", err)
		}

		allItems = append(allItems, resp.Result.Items...)

		if len(resp.Result.Items) < limit || resp.Result.LastID == "" {
			break
		}
		lastID = resp.Result.LastID
	}

	if len(allItems) == 0 {
		return nil
	}

	// 批量获取产品详情（每次最多100个）
	batchSize := 100
	for i := 0; i < len(allItems); i += batchSize {
		end := i + batchSize
		if end > len(allItems) {
			end = len(allItems)
		}

		batchItems := allItems[i:end]
		var offerIDs []string
		for _, item := range batchItems {
			offerIDs = append(offerIDs, item.OfferID)
		}

		infoResp, err := client.GetProductInfo(offerIDs)
		if err != nil {
			// 记录错误但继续处理
			fmt.Printf("获取产品详情失败 (batch %d-%d): %v\n", i, end, err)
			continue
		}

		// 保存到数据库
		for _, info := range infoResp.Result.Items {
			price, _ := strconv.ParseFloat(info.Price, 64)
			
			// 查找或创建
			var pp PlatformProduct
			err := db.Where("platform_auth_id = ? AND platform_sku = ?", platformAuthID, info.OfferID).First(&pp).Error
			
			if errors.Is(err, gorm.ErrRecordNotFound) {
				pp = PlatformProduct{
					Platform:       order.PlatformOzon,
					PlatformAuthID: platformAuthID,
					PlatformSKU:    info.OfferID,
				}
			}

			// 更新字段
			pp.Name = info.Name
			pp.Stock = info.Stocks.Present
			pp.Price = price
			pp.Currency = info.CurrencyCode
			pp.Status = info.Status.State
			rawData, _ := json.Marshal(info)
			pp.RawData = string(rawData)

			if err := db.Save(&pp).Error; err != nil {
				fmt.Printf("保存平台产品失败 %s: %v\n", info.OfferID, err)
			}
		}
	}

	// 更新同步时间
	// 注意：PlatformAuth 模型中可能没有 LastProductSyncAt 字段，这里暂时复用 LastSyncAt 或忽略
	// 最好是在 PlatformAuth 添加字段，但为了简化，这里先不更新 Auth 表状态，只记录日志
	
	return nil
}

// 辅助函数：创建 Ozon 客户端
func createOzonClient(credentials string, platformAuthID uint) (*ozon.Client, error) {
	creds, err := ozon.ParseCredentials(credentials)
	if err != nil {
		return nil, err
	}

	db := database.GetDB()
	logger := platform.NewRequestLogger(db, platformAuthID, order.PlatformOzon)
	return ozon.NewClient(creds, logger), nil
}

// ========== 同步任务相关 ==========

// CreateSyncTask 创建同步任务
func (s *Service) CreateSyncTask(platformAuthID uint, taskType string) (*PlatformSyncTask, error) {
	db := database.GetDB()

	task := &PlatformSyncTask{
		PlatformAuthID: platformAuthID,
		TaskType:       taskType,
		Status:         SyncTaskStatusPending,
		MaxRetry:       5,
	}

	if err := db.Create(task).Error; err != nil {
		return nil, err
	}

	return task, nil
}

// ExecuteSyncTask 执行同步任务（根据类型分发）
func (s *Service) ExecuteSyncTask(task *PlatformSyncTask) error {
	switch task.TaskType {
	case SyncTaskTypeProduct:
		return s.SyncPlatformProducts(task.PlatformAuthID)
	// 其他任务类型可在此扩展
	default:
		return fmt.Errorf("未知的任务类型: %s", task.TaskType)
	}
}

// ProcessPendingTasks 扫描并处理待执行任务
func (s *Service) ProcessPendingTasks() {
	db := database.GetDB()
	lockID := fmt.Sprintf("worker-%d", time.Now().UnixNano())
	lockTimeout := 10 * time.Minute // 锁定超时时间

	// 查找待执行或锁定超时的任务
	var tasks []PlatformSyncTask
	now := time.Now()
	lockExpired := now.Add(-lockTimeout)

	db.Where(
		"(status = ? OR (status = ? AND locked_at < ?)) AND retry_count < max_retry",
		SyncTaskStatusPending, SyncTaskStatusRunning, lockExpired,
	).Order("created_at ASC").Limit(10).Find(&tasks)

	fmt.Printf("[SyncTask] 扫描到 %d 个待处理任务\n", len(tasks))

	for _, task := range tasks {
		taskID := task.ID // 保存任务ID

		// 尝试获取锁（乐观锁）
		result := db.Model(&PlatformSyncTask{}).
			Where("id = ? AND (status = ? OR (status = ? AND locked_at < ?))",
				taskID, SyncTaskStatusPending, SyncTaskStatusRunning, lockExpired).
			Updates(map[string]interface{}{
				"status":    SyncTaskStatusRunning,
				"locked_at": now,
				"locked_by": lockID,
			})

		if result.RowsAffected == 0 {
			// 未获取到锁，跳过
			fmt.Printf("[SyncTask] 任务 %d 获取锁失败，跳过\n", taskID)
			continue
		}

		fmt.Printf("[SyncTask] 任务 %d 开始执行，类型: %s\n", taskID, task.TaskType)

		// 更新开始时间
		startedAt := time.Now()
		db.Model(&PlatformSyncTask{}).Where("id = ?", taskID).Update("started_at", &startedAt)

		// 执行任务
		err := s.ExecuteSyncTask(&task)
		finishedAt := time.Now()

		if err != nil {
			// 任务失败
			newRetryCount := task.RetryCount + 1
			newStatus := SyncTaskStatusPending // 等待重试
			if newRetryCount >= task.MaxRetry {
				newStatus = SyncTaskStatusFailed // 达到最大重试次数
			}

			db.Model(&PlatformSyncTask{}).Where("id = ?", taskID).Updates(map[string]interface{}{
				"status":        newStatus,
				"retry_count":   newRetryCount,
				"error_message": err.Error(),
				"finished_at":   &finishedAt,
				"locked_at":     nil,
				"locked_by":     "",
			})

			fmt.Printf("[SyncTask] 任务 %d 执行失败 (重试 %d/%d): %v\n",
				taskID, newRetryCount, task.MaxRetry, err)
		} else {
			// 任务成功
			db.Model(&PlatformSyncTask{}).Where("id = ?", taskID).Updates(map[string]interface{}{
				"status":        SyncTaskStatusSuccess,
				"error_message": "",
				"finished_at":   &finishedAt,
				"locked_at":     nil,
				"locked_by":     "",
			})

			fmt.Printf("[SyncTask] 任务 %d 执行成功\n", taskID)
		}
	}
}

// CleanOldTasks 清理旧任务记录
func (s *Service) CleanOldTasks(before time.Time) (int64, error) {
	db := database.GetDB()

	result := db.Where("created_at < ? AND status IN ?", before,
		[]string{SyncTaskStatusSuccess, SyncTaskStatusFailed}).
		Delete(&PlatformSyncTask{})

	return result.RowsAffected, result.Error
}

// ListSyncTasks 获取同步任务列表
func (s *Service) ListSyncTasks(page, pageSize int, status string) ([]PlatformSyncTask, int64, error) {
	db := database.GetDB()
	var tasks []PlatformSyncTask
	var total int64

	query := db.Model(&PlatformSyncTask{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}
