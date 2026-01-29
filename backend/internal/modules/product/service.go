package product

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"

	"autostack/internal/modules/apiClient/platform"
	"autostack/internal/modules/apiClient/platform/ozon"
	"autostack/internal/modules/order"
	"autostack/internal/repository"
	inventoryRepo "autostack/internal/repository/inventory"
	productRepo "autostack/internal/repository/product"
)

// Service 产品服务
type Service struct {
	txManager           repository.TxManager
	productRepo         productRepo.ProductRepository
	platformProductRepo productRepo.PlatformProductRepository
	mappingRepo         productRepo.ProductMappingRepository
	syncTaskRepo        productRepo.SyncTaskRepository
	supplierRepo        productRepo.ProductSupplierRepository
	warehouseRepo       inventoryRepo.WarehouseRepository
	inventoryRepo       inventoryRepo.InventoryRepository
	stockInOrderRepo    inventoryRepo.StockInOrderRepository
	stockInItemRepo     inventoryRepo.StockInOrderItemRepository
}

// NewService 创建产品服务实例
func NewService(
	txManager repository.TxManager,
	productRepo productRepo.ProductRepository,
	platformProductRepo productRepo.PlatformProductRepository,
	mappingRepo productRepo.ProductMappingRepository,
	syncTaskRepo productRepo.SyncTaskRepository,
	supplierRepo productRepo.ProductSupplierRepository,
	warehouseRepo inventoryRepo.WarehouseRepository,
	inventoryRepo inventoryRepo.InventoryRepository,
	stockInOrderRepo inventoryRepo.StockInOrderRepository,
	stockInItemRepo inventoryRepo.StockInOrderItemRepository,
) *Service {
	return &Service{
		txManager:           txManager,
		productRepo:         productRepo,
		platformProductRepo: platformProductRepo,
		mappingRepo:         mappingRepo,
		syncTaskRepo:        syncTaskRepo,
		supplierRepo:        supplierRepo,
		warehouseRepo:       warehouseRepo,
		inventoryRepo:       inventoryRepo,
		stockInOrderRepo:    stockInOrderRepo,
		stockInItemRepo:     stockInItemRepo,
	}
}

// ListProducts 获取本地产品列表
func (s *Service) ListProducts(page, pageSize int, keyword string) ([]Product, int64, error) {
	ctx := context.Background()
	return s.productRepo.List(ctx, &productRepo.ProductQuery{
		Page:     page,
		PageSize: pageSize,
		Keyword:  keyword,
	})
}

// CreateProduct 创建本地产品
func (s *Service) CreateProduct(req CreateProductRequest) (*Product, error) {
	ctx := context.Background()

	// 检查SKU是否已存在
	count, err := s.productRepo.CountBySKU(ctx, req.SKU)
	if err != nil {
		return nil, err
	}
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

	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, err
	}

	return product, nil
}

// UpdateProduct 更新本地产品
func (s *Service) UpdateProduct(id uint, req UpdateProductRequest) (*Product, error) {
	ctx := context.Background()

	product, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("产品不存在")
		}
		return nil, err
	}

	updates := map[string]interface{}{
		"wid":        req.WID,
		"name":       req.Name,
		"image":      req.Image,
		"cost_price": req.CostPrice,
		"weight":     req.Weight,
		"dimensions": req.Dimensions,
	}

	if err := s.productRepo.UpdateFields(ctx, id, updates); err != nil {
		return nil, err
	}

	// 重新加载产品
	product, _ = s.productRepo.FindByID(ctx, id)
	return product, nil
}

// DeleteProduct 删除本地产品
func (s *Service) DeleteProduct(id uint) error {
	ctx := context.Background()

	// 检查是否有映射关联
	count, err := s.mappingRepo.CountByProductID(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("该产品已关联平台产品，无法删除")
	}

	return s.productRepo.Delete(ctx, id)
}

// ListPlatformProducts 获取平台产品列表
func (s *Service) ListPlatformProducts(platformAuthID uint, keyword string, page, pageSize int) ([]PlatformProduct, int64, error) {
	ctx := context.Background()
	return s.platformProductRepo.List(ctx, &productRepo.PlatformProductQuery{
		Page:           page,
		PageSize:       pageSize,
		PlatformAuthID: platformAuthID,
		Keyword:        keyword,
	})
}

// MapProduct 关联产品
func (s *Service) MapProduct(req MapProductRequest) error {
	ctx := context.Background()

	// 检查平台产品是否存在
	platformProduct, err := s.platformProductRepo.FindByID(ctx, req.PlatformProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("平台产品不存在")
		}
		return err
	}

	// 检查本地产品是否存在
	_, err = s.productRepo.FindByID(ctx, req.ProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("本地产品不存在")
		}
		return err
	}

	// 如果未提供 PlatformAccountID，使用平台产品的 PlatformAccountID
	platformAccountID := req.PlatformAccountID
	if platformAccountID == 0 {
		platformAccountID = platformProduct.PlatformAccountID
	}

	// 检查是否已有相同的映射（使用复合唯一键）
	_, err = s.mappingRepo.FindByCompositeKey(ctx, req.WID, platformAccountID, req.ProductID, req.PlatformProductID)
	if err == nil {
		// 已存在相同映射，无需更新
		return nil
	}

	// 创建新映射
	newMapping := &ProductMapping{
		WID:               req.WID,
		PlatformAccountID: platformAccountID,
		ProductID:         req.ProductID,
		PlatformProductID: req.PlatformProductID,
	}

	return s.mappingRepo.Create(ctx, newMapping)
}

// UnmapProduct 解除关联
func (s *Service) UnmapProduct(platformProductID uint) error {
	ctx := context.Background()
	return s.mappingRepo.DeleteByPlatformProductID(ctx, platformProductID)
}

// SyncPlatformProducts 同步平台产品
func (s *Service) SyncPlatformProducts(platformAuthID uint) error {
	ctx := context.Background()
	db := s.txManager.DB()

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
	client, err := s.createOzonClient(credentials, platformAuthID)
	if err != nil {
		return err
	}

	// 拉取产品列表
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
			fmt.Printf("获取产品详情失败 (batch %d-%d): %v\n", i, end, err)
			continue
		}

		fmt.Printf("[DEBUG] 获取产品详情成功，返回 %d 个产品\n", len(infoResp.Items))
		if len(infoResp.Items) == 0 {
			rawResp, _ := json.Marshal(infoResp)
			fmt.Printf("[DEBUG] 产品详情响应为空，原始响应: %s\n", string(rawResp))
		}

		// 保存到数据库
		for _, info := range infoResp.Items {
			price, _ := strconv.ParseFloat(info.Price, 64)

			// 计算总库存
			totalStock := 0
			for _, stock := range info.Stocks.Stocks {
				totalStock += stock.Present
			}

			// 查找或创建
			pp, err := s.platformProductRepo.FindByAccountAndUniqueCode(ctx, auth.UserID, info.OfferID)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				pp = &PlatformProduct{
					Platform:          order.PlatformOzon,
					PlatformAuthID:    platformAuthID,
					PlatformAccountID: auth.UserID,
					PlatformSKU:       info.OfferID,
					UniqueCode:        info.OfferID,
				}
			} else if err != nil {
				fmt.Printf("查询平台产品失败 %s: %v\n", info.OfferID, err)
				continue
			}

			// 更新字段
			pp.PlatformAuthID = platformAuthID
			pp.PlatformAccountID = auth.UserID
			pp.UniqueCode = info.OfferID
			pp.Name = info.Name
			pp.Stock = totalStock
			pp.Price = price
			pp.Currency = info.CurrencyCode
			pp.Image = info.GetPrimaryImageURL()
			if info.IsArchived {
				pp.Status = "archived"
			} else {
				pp.Status = "active"
			}
			rawData, _ := json.Marshal(info)
			pp.RawData = string(rawData)

			if err := s.platformProductRepo.Save(ctx, pp); err != nil {
				fmt.Printf("保存平台产品失败 %s: %v\n", info.OfferID, err)
			}
		}
	}

	return nil
}

// createOzonClient 创建 Ozon 客户端
func (s *Service) createOzonClient(credentials string, platformAuthID uint) (*ozon.Client, error) {
	creds, err := ozon.ParseCredentials(credentials)
	if err != nil {
		return nil, err
	}

	db := s.txManager.DB()
	logger := platform.NewRequestLogger(db, platformAuthID, order.PlatformOzon)
	return ozon.NewClient(creds, logger), nil
}

// ========== 同步任务相关 ==========

// CreateSyncTask 创建同步任务
func (s *Service) CreateSyncTask(platformAuthID uint, taskType string) (*PlatformSyncTask, error) {
	ctx := context.Background()

	// 检查是否已存在相同的待处理或执行中的任务
	existingTask, err := s.syncTaskRepo.FindPendingOrRunning(ctx, platformAuthID, taskType)
	if err == nil {
		// 已存在相同任务，提升优先级+1
		if err := s.syncTaskRepo.UpdatePriority(ctx, existingTask.ID, existingTask.Priority+1); err != nil {
			return nil, err
		}
		existingTask.Priority++
		fmt.Printf("[SyncTask] 任务 %d 已存在，优先级提升至 %d\n", existingTask.ID, existingTask.Priority)
		return existingTask, nil
	}

	// 创建新任务
	task := &PlatformSyncTask{
		PlatformAuthID: platformAuthID,
		TaskType:       taskType,
		Status:         SyncTaskStatusPending,
		Priority:       10,
		MaxRetry:       5,
	}

	if err := s.syncTaskRepo.Create(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}

// ExecuteSyncTask 执行同步任务
func (s *Service) ExecuteSyncTask(task *PlatformSyncTask) error {
	switch task.TaskType {
	case SyncTaskTypeProduct:
		return s.SyncPlatformProducts(task.PlatformAuthID)
	default:
		return fmt.Errorf("未知的任务类型: %s", task.TaskType)
	}
}

// ProcessPendingTasks 扫描并处理待执行任务
func (s *Service) ProcessPendingTasks() {
	ctx := context.Background()
	lockID := fmt.Sprintf("worker-%d", time.Now().UnixNano())
	lockTimeout := 10 * time.Minute

	// 查找待执行任务
	tasks, err := s.syncTaskRepo.FindPendingTasks(ctx, lockTimeout, 10)
	if err != nil {
		fmt.Printf("[SyncTask] 查询待处理任务失败: %v\n", err)
		return
	}

	fmt.Printf("[SyncTask] 扫描到 %d 个待处理任务\n", len(tasks))

	for _, task := range tasks {
		taskID := task.ID

		// 尝试获取锁
		locked, err := s.syncTaskRepo.TryLock(ctx, taskID, lockID, lockTimeout)
		if err != nil || !locked {
			fmt.Printf("[SyncTask] 任务 %d 获取锁失败，跳过\n", taskID)
			continue
		}

		fmt.Printf("[SyncTask] 任务 %d 开始执行，类型: %s\n", taskID, task.TaskType)

		// 更新开始时间
		startedAt := time.Now()
		s.syncTaskRepo.UpdateStatus(ctx, taskID, map[string]interface{}{"started_at": &startedAt})

		// 执行任务
		err = s.ExecuteSyncTask(&task)
		finishedAt := time.Now()

		if err != nil {
			// 任务失败
			newRetryCount := task.RetryCount + 1
			newStatus := SyncTaskStatusPending
			if newRetryCount >= task.MaxRetry {
				newStatus = SyncTaskStatusFailed
			}

			s.syncTaskRepo.UpdateStatus(ctx, taskID, map[string]interface{}{
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
			s.syncTaskRepo.UpdateStatus(ctx, taskID, map[string]interface{}{
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
	ctx := context.Background()
	return s.syncTaskRepo.DeleteOldTasks(ctx, before, []string{SyncTaskStatusSuccess, SyncTaskStatusFailed})
}

// InitProductsFromPlatform 根据平台SKU初始化本地产品
func (s *Service) InitProductsFromPlatform(platformAuthID uint) (*InitProductsResponse, error) {
	ctx := context.Background()
	result := &InitProductsResponse{}

	// 获取默认仓库ID
	var defaultWID uint = 0
	defaultWarehouse, err := s.warehouseRepo.FindFirstActive(ctx)
	if err == nil {
		defaultWID = defaultWarehouse.ID
	}

	// 查询平台产品
	var platformProducts []PlatformProduct
	if platformAuthID > 0 {
		platformProducts, err = s.platformProductRepo.ListByAuthID(ctx, platformAuthID)
	} else {
		platformProducts, _, err = s.platformProductRepo.List(ctx, &productRepo.PlatformProductQuery{
			Page:     1,
			PageSize: 100000, // 获取所有
		})
	}
	if err != nil {
		return nil, fmt.Errorf("查询平台产品失败: %v", err)
	}

	result.TotalPlatformProducts = len(platformProducts)

	for _, pp := range platformProducts {
		// 查找或创建本地产品
		localProduct, err := s.productRepo.FindBySKU(ctx, pp.PlatformSKU)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新的本地产品
			localProduct = &Product{
				SKU:   pp.PlatformSKU,
				Name:  pp.Name,
				Image: pp.Image,
			}
			if err := s.productRepo.Create(ctx, localProduct); err != nil {
				fmt.Printf("[InitProducts] 创建本地产品失败 %s: %v\n", pp.PlatformSKU, err)
				continue
			}
			result.CreatedProducts++
		} else if err != nil {
			fmt.Printf("[InitProducts] 查询本地产品失败 %s: %v\n", pp.PlatformSKU, err)
			continue
		}

		// 检查是否已有相同的映射
		_, err = s.mappingRepo.FindByCompositeKey(ctx, defaultWID, pp.PlatformAccountID, localProduct.ID, pp.ID)
		if err == nil {
			result.SkippedMapped++
			continue
		}

		// 创建映射关系
		mapping := &ProductMapping{
			WID:               defaultWID,
			PlatformAccountID: pp.PlatformAccountID,
			ProductID:         localProduct.ID,
			PlatformProductID: pp.ID,
		}
		if err := s.mappingRepo.Create(ctx, mapping); err != nil {
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
	ctx := context.Background()
	return s.syncTaskRepo.List(ctx, &productRepo.SyncTaskQuery{
		Page:     page,
		PageSize: pageSize,
		Status:   status,
	})
}

// ========== 入库单相关 ==========

// generateStockInOrderNo 生成入库单号
func generateStockInOrderNo() string {
	return fmt.Sprintf("SI%s%04d", time.Now().Format("20060102150405"), time.Now().Nanosecond()%10000)
}

// CreateStockInOrder 创建入库单（同时更新库存）
func (s *Service) CreateStockInOrder(req CreateStockInOrderRequest) (*StockInOrder, error) {
	ctx := context.Background()

	// 验证仓库存在
	_, err := s.warehouseRepo.FindByID(ctx, req.WarehouseID)
	if err != nil {
		return nil, fmt.Errorf("仓库不存在")
	}

	var resultOrder *StockInOrder

	// 使用事务管理器
	err = s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// 创建入库单
		stockInOrder := &StockInOrder{
			OrderNo:     generateStockInOrderNo(),
			WarehouseID: req.WarehouseID,
			Status:      StockInStatusCompleted,
			Remark:      req.Remark,
		}

		if err := s.stockInOrderRepo.Create(txCtx, stockInOrder); err != nil {
			return fmt.Errorf("创建入库单失败: %v", err)
		}

		// 创建明细并更新库存
		for _, item := range req.Items {
			// 查询产品
			product, err := s.productRepo.FindByID(txCtx, item.ProductID)
			if err != nil {
				return fmt.Errorf("产品不存在: %d", item.ProductID)
			}

			// 创建明细
			orderItem := &StockInOrderItem{
				StockInOrderID: stockInOrder.ID,
				ProductID:      item.ProductID,
				SKU:            product.SKU,
				ProductName:    product.Name,
				Quantity:       item.Quantity,
			}
			if err := s.stockInItemRepo.Create(txCtx, orderItem); err != nil {
				return fmt.Errorf("创建入库明细失败: %v", err)
			}

			// 更新产品总库存
			if err := s.productRepo.UpdateStock(txCtx, item.ProductID, item.Quantity); err != nil {
				return fmt.Errorf("更新产品库存失败: %v", err)
			}

			// 更新仓库库存明细
			inventory, err := s.inventoryRepo.FindByProductAndWarehouse(txCtx, item.ProductID, req.WarehouseID)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 创建新的库存记录
				inventory = &WarehouseCenterInventory{
					ProductID:      item.ProductID,
					WarehouseID:    req.WarehouseID,
					SKU:            product.SKU,
					AvailableStock: item.Quantity,
					LockedStock:    0,
					InTransitStock: 0,
				}
				if err := s.inventoryRepo.Create(txCtx, inventory); err != nil {
					return fmt.Errorf("创建仓库库存记录失败: %v", err)
				}
			} else if err != nil {
				return fmt.Errorf("查询仓库库存失败: %v", err)
			} else {
				// 更新现有库存记录
				if err := s.inventoryRepo.UpdateAvailableStock(txCtx, inventory.ID, item.Quantity); err != nil {
					return fmt.Errorf("更新仓库库存失败: %v", err)
				}
			}
		}

		resultOrder = stockInOrder
		return nil
	})

	if err != nil {
		return nil, err
	}

	// 重新加载完整数据
	resultOrder, _ = s.stockInOrderRepo.FindByIDWithDetails(ctx, resultOrder.ID)

	return resultOrder, nil
}

// ListStockInOrders 获取入库单列表
func (s *Service) ListStockInOrders(page, pageSize int, status string) ([]StockInOrder, int64, error) {
	ctx := context.Background()
	return s.stockInOrderRepo.List(ctx, &inventoryRepo.StockInOrderQuery{
		Page:     page,
		PageSize: pageSize,
		Status:   status,
	})
}

// GetStockInOrder 获取入库单详情
func (s *Service) GetStockInOrder(id uint) (*StockInOrder, error) {
	ctx := context.Background()
	order, err := s.stockInOrderRepo.FindByIDWithDetails(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("入库单不存在")
		}
		return nil, err
	}
	return order, nil
}

// ========== 仓库相关 ==========

// ListWarehouses 获取仓库列表
func (s *Service) ListWarehouses() ([]Warehouse, error) {
	ctx := context.Background()
	return s.warehouseRepo.ListActive(ctx)
}

// ListAllWarehouses 获取所有仓库（支持按类型筛选）
func (s *Service) ListAllWarehouses(warehouseType string) ([]Warehouse, error) {
	ctx := context.Background()
	return s.warehouseRepo.List(ctx, &inventoryRepo.WarehouseQuery{
		Type: warehouseType,
	})
}

// CreateWarehouse 创建仓库
func (s *Service) CreateWarehouse(req CreateWarehouseRequest) (*Warehouse, error) {
	ctx := context.Background()

	// 检查编码是否已存在
	count, err := s.warehouseRepo.CountByCode(ctx, req.Code)
	if err != nil {
		return nil, err
	}
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

	if err := s.warehouseRepo.Create(ctx, warehouse); err != nil {
		return nil, err
	}

	return warehouse, nil
}

// InitDefaultWarehouse 初始化默认仓库
func (s *Service) InitDefaultWarehouse() error {
	ctx := context.Background()

	count, err := s.warehouseRepo.Count(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // 已有仓库，无需初始化
	}

	defaultWarehouse := &Warehouse{
		Code:    "WH001",
		Name:    "默认仓库",
		Address: "",
		Status:  WarehouseStatusActive,
	}
	return s.warehouseRepo.Create(ctx, defaultWarehouse)
}

// ========== 库存相关 ==========

// ListInventory 获取库存明细列表
func (s *Service) ListInventory(warehouseID uint, keyword string, page, pageSize int) ([]WarehouseCenterInventory, int64, error) {
	ctx := context.Background()
	return s.inventoryRepo.List(ctx, &inventoryRepo.InventoryQuery{
		Page:        page,
		PageSize:    pageSize,
		WarehouseID: warehouseID,
		Keyword:     keyword,
	})
}

// GetOrCreateInventory 获取或创建库存记录
func (s *Service) GetOrCreateInventory(productID, warehouseID uint) (*WarehouseCenterInventory, error) {
	ctx := context.Background()

	inventory, err := s.inventoryRepo.FindByProductAndWarehouse(ctx, productID, warehouseID)
	if err == nil {
		return inventory, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 获取产品信息
	product, err := s.productRepo.FindByID(ctx, productID)
	if err != nil {
		return nil, errors.New("产品不存在")
	}

	// 创建新库存记录
	inventory = &WarehouseCenterInventory{
		ProductID:      productID,
		WarehouseID:    warehouseID,
		SKU:            product.SKU,
		AvailableStock: 0,
		LockedStock:    0,
		InTransitStock: 0,
	}

	if err := s.inventoryRepo.Create(ctx, inventory); err != nil {
		return nil, err
	}

	return inventory, nil
}

// UpdateInventory 更新库存
func (s *Service) UpdateInventory(req UpdateInventoryRequest) (*WarehouseCenterInventory, error) {
	ctx := context.Background()

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
		if err := s.inventoryRepo.UpdateFields(ctx, inventory.ID, updates); err != nil {
			return nil, err
		}
	}

	// 重新加载
	inventory, _ = s.inventoryRepo.FindByID(ctx, inventory.ID)

	return inventory, nil
}

// InitInventoryFromProducts 从产品表初始化库存
func (s *Service) InitInventoryFromProducts(warehouseID uint) (int, error) {
	ctx := context.Background()

	// 获取所有产品
	products, err := s.productRepo.FindAll(ctx)
	if err != nil {
		return 0, err
	}

	created := 0
	for _, p := range products {
		count, err := s.inventoryRepo.CountByProductAndWarehouse(ctx, p.ID, warehouseID)
		if err != nil {
			continue
		}
		if count > 0 {
			continue // 已存在
		}

		inventory := &WarehouseCenterInventory{
			ProductID:      p.ID,
			WarehouseID:    warehouseID,
			SKU:            p.SKU,
			AvailableStock: p.Stock,
			LockedStock:    0,
			InTransitStock: 0,
		}
		if err := s.inventoryRepo.Create(ctx, inventory); err != nil {
			fmt.Printf("[InitInventory] 创建库存记录失败 %s: %v\n", p.SKU, err)
			continue
		}
		created++
	}

	return created, nil
}

// ========== 供应商/采购信息相关 ==========

// ListSuppliersByProductID 获取产品的供应商列表
func (s *Service) ListSuppliersByProductID(productID uint) ([]ProductSupplier, error) {
	ctx := context.Background()
	return s.supplierRepo.FindByProductID(ctx, productID)
}

// ListSuppliers 分页获取供应商列表
func (s *Service) ListSuppliers(productID uint, keyword, status string, page, pageSize int) ([]ProductSupplier, int64, error) {
	ctx := context.Background()
	return s.supplierRepo.List(ctx, &productRepo.ProductSupplierQuery{
		Page:      page,
		PageSize:  pageSize,
		ProductID: productID,
		Keyword:   keyword,
		Status:    status,
	})
}

// GetSupplier 获取供应商详情
func (s *Service) GetSupplier(id uint) (*ProductSupplier, error) {
	ctx := context.Background()
	return s.supplierRepo.FindByID(ctx, id)
}

// CreateSupplier 创建供应商
func (s *Service) CreateSupplier(req CreateSupplierRequest) (*ProductSupplier, error) {
	ctx := context.Background()

	// 验证产品是否存在
	_, err := s.productRepo.FindByID(ctx, req.ProductID)
	if err != nil {
		return nil, errors.New("产品不存在")
	}

	currency := req.Currency
	if currency == "" {
		currency = "CNY"
	}

	supplier := &ProductSupplier{
		ProductID:     req.ProductID,
		SupplierName:  req.SupplierName,
		PurchaseLink:  req.PurchaseLink,
		UnitPrice:     req.UnitPrice,
		Currency:      currency,
		MinOrderQty:   req.MinOrderQty,
		LeadTime:      req.LeadTime,
		EstimatedDays: req.EstimatedDays,
		Remark:        req.Remark,
		IsDefault:     req.IsDefault,
		Status:        SupplierStatusActive,
	}

	if err := s.supplierRepo.Create(ctx, supplier); err != nil {
		return nil, err
	}

	// 如果设置为默认，则更新其他供应商的默认状态
	if req.IsDefault {
		_ = s.supplierRepo.SetDefault(ctx, req.ProductID, supplier.ID)
	}

	return supplier, nil
}

// UpdateSupplier 更新供应商
func (s *Service) UpdateSupplier(id uint, req UpdateSupplierRequest) (*ProductSupplier, error) {
	ctx := context.Background()

	supplier, err := s.supplierRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("供应商不存在")
	}

	if req.SupplierName != "" {
		supplier.SupplierName = req.SupplierName
	}
	supplier.PurchaseLink = req.PurchaseLink
	supplier.UnitPrice = req.UnitPrice
	if req.Currency != "" {
		supplier.Currency = req.Currency
	}
	supplier.MinOrderQty = req.MinOrderQty
	supplier.LeadTime = req.LeadTime
	supplier.EstimatedDays = req.EstimatedDays
	supplier.Remark = req.Remark
	supplier.IsDefault = req.IsDefault
	if req.Status != "" {
		supplier.Status = req.Status
	}

	if err := s.supplierRepo.Update(ctx, supplier); err != nil {
		return nil, err
	}

	// 如果设置为默认，则更新其他供应商的默认状态
	if req.IsDefault {
		_ = s.supplierRepo.SetDefault(ctx, supplier.ProductID, supplier.ID)
	}

	return supplier, nil
}

// DeleteSupplier 删除供应商
func (s *Service) DeleteSupplier(id uint) error {
	ctx := context.Background()
	return s.supplierRepo.Delete(ctx, id)
}

// SetDefaultSupplier 设置默认供应商
func (s *Service) SetDefaultSupplier(productID, supplierID uint) error {
	ctx := context.Background()
	return s.supplierRepo.SetDefault(ctx, productID, supplierID)
}
