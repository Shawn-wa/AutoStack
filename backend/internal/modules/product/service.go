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
func (s *Service) ListProducts(page, pageSize int, keyword string) ([]Product, int64, error) {
	var products []Product
	var total int64
	db := database.GetDB()

	query := db.Model(&Product{})
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("sku LIKE ? OR name LIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&products).Error; err != nil {
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
		WID:        req.WID,
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
		"wid":        req.WID,
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
func (s *Service) ListPlatformProducts(platformAuthID uint, keyword string, page, pageSize int) ([]PlatformProduct, int64, error) {
	var products []PlatformProduct
	var total int64
	db := database.GetDB()

	query := db.Model(&PlatformProduct{})
	if platformAuthID > 0 {
		query = query.Where("platform_auth_id = ?", platformAuthID)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		// 使用子查询搜索本地SKU，避免JOIN导致的重复记录和性能问题
		// 搜索条件：平台SKU、平台名称、关联的本地SKU
		query = query.Where(`(
			platform_sku LIKE ? OR 
			name LIKE ? OR 
			id IN (
				SELECT pm.platform_product_id 
				FROM product_mappings pm 
				JOIN products p ON p.id = pm.product_id 
				WHERE p.sku LIKE ?
			)
		)`, like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("ProductMapping.Product").Order("id DESC").Offset(offset).Limit(pageSize).Find(&products).Error; err != nil {
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

	// 如果未提供 PlatformAccountID，使用平台产品的 PlatformAccountID
	platformAccountID := req.PlatformAccountID
	if platformAccountID == 0 {
		platformAccountID = platformProduct.PlatformAccountID
	}

	// 检查是否已有相同的映射（使用复合唯一键）
	var mapping ProductMapping
	err := db.Where("wid = ? AND platform_account_id = ? AND product_id = ? AND platform_product_id = ?",
		req.WID, platformAccountID, req.ProductID, req.PlatformProductID).First(&mapping).Error
	if err == nil {
		// 已存在相同映射，无需更新
		return nil
	}

	// 创建新映射
	newMapping := ProductMapping{
		WID:               req.WID,
		PlatformAccountID: platformAccountID,
		ProductID:         req.ProductID,
		PlatformProductID: req.PlatformProductID,
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

		// 调试日志：打印返回的产品数量
		fmt.Printf("[DEBUG] 获取产品详情成功，返回 %d 个产品\n", len(infoResp.Items))
		if len(infoResp.Items) == 0 {
			// 打印原始响应以便调试
			rawResp, _ := json.Marshal(infoResp)
			fmt.Printf("[DEBUG] 产品详情响应为空，原始响应: %s\n", string(rawResp))
		}

		// 保存到数据库
		for _, info := range infoResp.Items {
			price, _ := strconv.ParseFloat(info.Price, 64)

			// 计算总库存 (累加所有仓库类型的库存)
			totalStock := 0
			for _, stock := range info.Stocks.Stocks {
				totalStock += stock.Present
			}

			// 查找或创建：使用 platform_account_id + unique_code 作为唯一键
			var pp PlatformProduct
			err := db.Where("platform_account_id = ? AND unique_code = ?", auth.UserID, info.OfferID).First(&pp).Error

			if errors.Is(err, gorm.ErrRecordNotFound) {
				pp = PlatformProduct{
					Platform:          order.PlatformOzon,
					PlatformAuthID:    platformAuthID,
					PlatformAccountID: auth.UserID,
					PlatformSKU:       info.OfferID,
					UniqueCode:        info.OfferID,
				}
			}

			// 更新字段
			pp.PlatformAuthID = platformAuthID // 始终更新为最新的授权ID
			pp.PlatformAccountID = auth.UserID
			pp.UniqueCode = info.OfferID
			pp.Name = info.Name
			pp.Stock = totalStock
			pp.Price = price
			pp.Currency = info.CurrencyCode
			// 保存主图：使用 GetPrimaryImageURL 方法获取
			pp.Image = info.GetPrimaryImageURL()
			// v3 API 无 status.state，使用 is_archived 判断状态
			if info.IsArchived {
				pp.Status = "archived"
			} else {
				pp.Status = "active"
			}
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

// CreateSyncTask 创建同步任务（自动去重，避免并发，手动触发提升优先级）
func (s *Service) CreateSyncTask(platformAuthID uint, taskType string) (*PlatformSyncTask, error) {
	db := database.GetDB()

	// 检查是否已存在相同的待处理或执行中的任务
	var existingTask PlatformSyncTask
	err := db.Where(
		"platform_auth_id = ? AND task_type = ? AND status IN ?",
		platformAuthID, taskType, []string{SyncTaskStatusPending, SyncTaskStatusRunning},
	).First(&existingTask).Error

	if err == nil {
		// 已存在相同任务，提升优先级+1
		db.Model(&PlatformSyncTask{}).Where("id = ?", existingTask.ID).
			Update("priority", existingTask.Priority+1)
		existingTask.Priority++
		fmt.Printf("[SyncTask] 任务 %d 已存在，优先级提升至 %d\n", existingTask.ID, existingTask.Priority)
		return &existingTask, nil
	}

	// 创建新任务，默认优先级10
	task := &PlatformSyncTask{
		PlatformAuthID: platformAuthID,
		TaskType:       taskType,
		Status:         SyncTaskStatusPending,
		Priority:       10,
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
	).Order("priority DESC, created_at ASC").Limit(10).Find(&tasks)

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

// InitProductsFromPlatform 根据平台SKU初始化本地产品
func (s *Service) InitProductsFromPlatform(platformAuthID uint) (*InitProductsResponse, error) {
	db := database.GetDB()
	result := &InitProductsResponse{}

	// 获取默认仓库ID（取第一个可用仓库，如果没有则为0）
	var defaultWarehouse Warehouse
	var defaultWID uint = 0
	if err := db.Where("status = ?", 1).First(&defaultWarehouse).Error; err == nil {
		defaultWID = defaultWarehouse.ID
	}

	// 查询平台产品
	query := db.Model(&PlatformProduct{})
	if platformAuthID > 0 {
		query = query.Where("platform_auth_id = ?", platformAuthID)
	}

	var platformProducts []PlatformProduct
	if err := query.Find(&platformProducts).Error; err != nil {
		return nil, fmt.Errorf("查询平台产品失败: %v", err)
	}

	result.TotalPlatformProducts = len(platformProducts)

	for _, pp := range platformProducts {
		// 检查是否已有完全相同的映射（wid + platform_account_id + product_id + platform_product_id）
		// 先查找或创建本地产品
		var localProduct Product
		err := db.Where("sku = ?", pp.PlatformSKU).First(&localProduct).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新的本地产品
			localProduct = Product{
				SKU:   pp.PlatformSKU,
				Name:  pp.Name,
				Image: pp.Image,
			}
			if err := db.Create(&localProduct).Error; err != nil {
				fmt.Printf("[InitProducts] 创建本地产品失败 %s: %v\n", pp.PlatformSKU, err)
				continue
			}
			result.CreatedProducts++
		} else if err != nil {
			fmt.Printf("[InitProducts] 查询本地产品失败 %s: %v\n", pp.PlatformSKU, err)
			continue
		}

		// 检查是否已有相同的映射（使用复合唯一键）
		var existingMapping ProductMapping
		if err := db.Where("wid = ? AND platform_account_id = ? AND product_id = ? AND platform_product_id = ?",
			defaultWID, pp.PlatformAccountID, localProduct.ID, pp.ID).First(&existingMapping).Error; err == nil {
			result.SkippedMapped++
			continue
		}

		// 创建映射关系
		mapping := ProductMapping{
			WID:               defaultWID,
			PlatformAccountID: pp.PlatformAccountID,
			ProductID:         localProduct.ID,
			PlatformProductID: pp.ID,
		}
		if err := db.Create(&mapping).Error; err != nil {
			fmt.Printf("[InitProducts] 创建映射失败 wid=%d, account=%d, product=%d, platform=%d: %v\n",
				defaultWID, pp.PlatformAccountID, localProduct.ID, pp.ID, err)
			continue
		}
		result.CreatedMappings++
	}

	fmt.Printf("[InitProducts] 完成: 总数=%d, 跳过(已映射)=%d, 新建产品=%d, 新建映射=%d\n",
		result.TotalPlatformProducts, result.SkippedMapped,
		result.CreatedProducts, result.CreatedMappings)

	return result, nil
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

// ========== 入库单相关 ==========

// generateStockInOrderNo 生成入库单号
func generateStockInOrderNo() string {
	return fmt.Sprintf("SI%s%04d", time.Now().Format("20060102150405"), time.Now().Nanosecond()%10000)
}

// CreateStockInOrder 创建入库单（同时更新库存）
func (s *Service) CreateStockInOrder(req CreateStockInOrderRequest) (*StockInOrder, error) {
	db := database.GetDB()

	// 验证仓库存在
	var warehouse Warehouse
	if err := db.First(&warehouse, req.WarehouseID).Error; err != nil {
		return nil, fmt.Errorf("仓库不存在")
	}

	// 开启事务
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建入库单
	order := &StockInOrder{
		OrderNo:     generateStockInOrderNo(),
		WarehouseID: req.WarehouseID,
		Status:      StockInStatusCompleted, // 直接完成
		Remark:      req.Remark,
	}

	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建入库单失败: %v", err)
	}

	// 创建明细并更新库存
	for _, item := range req.Items {
		// 查询产品
		var product Product
		if err := tx.First(&product, item.ProductID).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("产品不存在: %d", item.ProductID)
		}

		// 创建明细
		orderItem := &StockInOrderItem{
			StockInOrderID: order.ID,
			ProductID:      item.ProductID,
			SKU:            product.SKU,
			ProductName:    product.Name,
			Quantity:       item.Quantity,
		}
		if err := tx.Create(orderItem).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("创建入库明细失败: %v", err)
		}

		// 更新产品总库存
		if err := tx.Model(&Product{}).Where("id = ?", item.ProductID).
			Update("stock", gorm.Expr("stock + ?", item.Quantity)).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("更新产品库存失败: %v", err)
		}

		// 更新仓库库存明细（warehouse_center_inventory）
		var inventory WarehouseCenterInventory
		err := tx.Where("product_id = ? AND warehouse_id = ?", item.ProductID, req.WarehouseID).First(&inventory).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新的库存记录
			inventory = WarehouseCenterInventory{
				ProductID:      item.ProductID,
				WarehouseID:    req.WarehouseID,
				SKU:            product.SKU,
				AvailableStock: item.Quantity,
				LockedStock:    0,
				InTransitStock: 0,
			}
			if err := tx.Create(&inventory).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("创建仓库库存记录失败: %v", err)
			}
		} else if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("查询仓库库存失败: %v", err)
		} else {
			// 更新现有库存记录的可用库存
			if err := tx.Model(&WarehouseCenterInventory{}).Where("id = ?", inventory.ID).
				Update("available_stock", gorm.Expr("available_stock + ?", item.Quantity)).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("更新仓库库存失败: %v", err)
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("提交事务失败: %v", err)
	}

	// 重新加载完整数据
	db.Preload("Items").First(order, order.ID)

	return order, nil
}

// ListStockInOrders 获取入库单列表
func (s *Service) ListStockInOrders(page, pageSize int, status string) ([]StockInOrder, int64, error) {
	db := database.GetDB()
	var orders []StockInOrder
	var total int64

	query := db.Model(&StockInOrder{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Warehouse").Preload("Items").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// GetStockInOrder 获取入库单详情
func (s *Service) GetStockInOrder(id uint) (*StockInOrder, error) {
	db := database.GetDB()
	var order StockInOrder
	if err := db.Preload("Warehouse").Preload("Items").First(&order, id).Error; err != nil {
		return nil, errors.New("入库单不存在")
	}
	return &order, nil
}

// ========== 仓库相关 ==========

// ListWarehouses 获取仓库列表
func (s *Service) ListWarehouses() ([]Warehouse, error) {
	db := database.GetDB()
	var warehouses []Warehouse
	if err := db.Where("status = ?", WarehouseStatusActive).Order("id ASC").Find(&warehouses).Error; err != nil {
		return nil, err
	}
	return warehouses, nil
}

// ListAllWarehouses 获取所有仓库（支持按类型筛选）
func (s *Service) ListAllWarehouses(warehouseType string) ([]Warehouse, error) {
	db := database.GetDB()
	var warehouses []Warehouse

	query := db.Model(&Warehouse{})
	if warehouseType != "" && warehouseType != "all" {
		query = query.Where("type = ?", warehouseType)
	}

	if err := query.Order("id ASC").Find(&warehouses).Error; err != nil {
		return nil, err
	}
	return warehouses, nil
}

// CreateWarehouse 创建仓库
func (s *Service) CreateWarehouse(req CreateWarehouseRequest) (*Warehouse, error) {
	db := database.GetDB()

	// 检查编码是否已存在
	var count int64
	db.Model(&Warehouse{}).Where("code = ?", req.Code).Count(&count)
	if count > 0 {
		return nil, errors.New("仓库编码已存在")
	}

	// 默认仓库类型为本地仓
	warehouseType := req.Type
	if warehouseType == "" {
		warehouseType = WarehouseTypeLocal
	}

	warehouse := &Warehouse{
		Code:    req.Code,
		Name:    req.Name,
		Type:    warehouseType,
		Address: req.Address,
		Status:  WarehouseStatusActive,
	}

	if err := db.Create(warehouse).Error; err != nil {
		return nil, err
	}

	return warehouse, nil
}

// InitDefaultWarehouse 初始化默认仓库
func (s *Service) InitDefaultWarehouse() error {
	db := database.GetDB()
	var count int64
	db.Model(&Warehouse{}).Count(&count)
	if count > 0 {
		return nil // 已有仓库，无需初始化
	}

	defaultWarehouse := &Warehouse{
		Code:    "WH001",
		Name:    "默认仓库",
		Address: "",
		Status:  WarehouseStatusActive,
	}
	return db.Create(defaultWarehouse).Error
}

// ========== 库存相关 ==========

// ListInventory 获取库存明细列表
func (s *Service) ListInventory(warehouseID uint, keyword string, page, pageSize int) ([]WarehouseCenterInventory, int64, error) {
	db := database.GetDB()
	var inventories []WarehouseCenterInventory
	var total int64

	query := db.Model(&WarehouseCenterInventory{})
	if warehouseID > 0 {
		query = query.Where("warehouse_id = ?", warehouseID)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("sku LIKE ?", like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Product").Preload("Warehouse").
		Order("updated_at DESC").Offset(offset).Limit(pageSize).Find(&inventories).Error; err != nil {
		return nil, 0, err
	}

	return inventories, total, nil
}

// GetOrCreateInventory 获取或创建库存记录
func (s *Service) GetOrCreateInventory(productID, warehouseID uint) (*WarehouseCenterInventory, error) {
	db := database.GetDB()

	var inventory WarehouseCenterInventory
	err := db.Where("product_id = ? AND warehouse_id = ?", productID, warehouseID).First(&inventory).Error
	if err == nil {
		return &inventory, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 获取产品信息
	var product Product
	if err := db.First(&product, productID).Error; err != nil {
		return nil, errors.New("产品不存在")
	}

	// 创建新库存记录
	inventory = WarehouseCenterInventory{
		ProductID:      productID,
		WarehouseID:    warehouseID,
		SKU:            product.SKU,
		AvailableStock: 0,
		LockedStock:    0,
		InTransitStock: 0,
	}

	if err := db.Create(&inventory).Error; err != nil {
		return nil, err
	}

	return &inventory, nil
}

// UpdateInventory 更新库存
func (s *Service) UpdateInventory(req UpdateInventoryRequest) (*WarehouseCenterInventory, error) {
	db := database.GetDB()

	inventory, err := s.GetOrCreateInventory(req.ProductID, req.WarehouseID)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})
	if req.AvailableStock != nil {
		updates["available_stock"] = *req.AvailableStock
	}
	if req.LockedStock != nil {
		updates["locked_stock"] = *req.LockedStock
	}
	if req.InTransitStock != nil {
		updates["in_transit_stock"] = *req.InTransitStock
	}

	if len(updates) > 0 {
		if err := db.Model(inventory).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	// 重新加载
	db.Preload("Product").Preload("Warehouse").First(inventory, inventory.ID)

	return inventory, nil
}

// InitInventoryFromProducts 从产品表初始化库存（为所有产品创建默认仓库库存记录）
func (s *Service) InitInventoryFromProducts(warehouseID uint) (int, error) {
	db := database.GetDB()

	// 获取所有产品
	var products []Product
	if err := db.Find(&products).Error; err != nil {
		return 0, err
	}

	created := 0
	for _, p := range products {
		var count int64
		db.Model(&WarehouseCenterInventory{}).Where("product_id = ? AND warehouse_id = ?", p.ID, warehouseID).Count(&count)
		if count > 0 {
			continue // 已存在
		}

		inventory := WarehouseCenterInventory{
			ProductID:      p.ID,
			WarehouseID:    warehouseID,
			SKU:            p.SKU,
			AvailableStock: p.Stock, // 使用产品表的库存作为初始可用库存
			LockedStock:    0,
			InTransitStock: 0,
		}
		if err := db.Create(&inventory).Error; err != nil {
			fmt.Printf("[InitInventory] 创建库存记录失败 %s: %v\n", p.SKU, err)
			continue
		}
		created++
	}

	return created, nil
}
