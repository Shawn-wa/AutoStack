package product

import (
	"autostack/internal/repository"
	inventoryRepo "autostack/internal/repository/inventory"
	productRepo "autostack/internal/repository/product"
	"autostack/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// service 产品服务实例
var service *Service

// InitHandler 初始化 Handler，注入 Service 依赖
// 应在服务器启动时调用
func InitHandler(db *gorm.DB) {
	txManager := repository.NewTxManager(db)

	service = NewService(
		txManager,
		productRepo.NewProductRepository(db),
		productRepo.NewPlatformProductRepository(db),
		productRepo.NewProductMappingRepository(db),
		productRepo.NewSyncTaskRepository(db),
		productRepo.NewProductSupplierRepository(db),
		inventoryRepo.NewWarehouseRepository(db),
		inventoryRepo.NewInventoryRepository(db),
		inventoryRepo.NewStockInOrderRepository(db),
		inventoryRepo.NewStockInOrderItemRepository(db),
	)
}

// GetService 获取服务实例（用于外部调用，如定时任务）
func GetService() *Service {
	return service
}

// ListProducts 获取本地产品列表
func ListProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")

	products, total, err := service.ListProducts(page, pageSize, keyword)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取产品列表失败")
		return
	}

	// 获取仓库ID列表并批量查询仓库名称
	warehouseMap := make(map[uint]string)
	warehouses, _ := service.ListWarehouses()
	for _, w := range warehouses {
		warehouseMap[w.ID] = w.Name
	}

	var list []ProductResponse
	for _, p := range products {
		list = append(list, ProductResponse{
			ID:            p.ID,
			WID:           p.WID,
			WarehouseName: warehouseMap[p.WID],
			SKU:           p.SKU,
			Name:          p.Name,
			Image:         p.Image,
			CostPrice:     p.CostPrice,
			Weight:        p.Weight,
			Dimensions:    p.Dimensions,
			CreatedAt:     p.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:     p.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	response.Success(c, http.StatusOK, "获取成功", ProductListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// CreateProduct 创建本地产品
func CreateProduct(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	product, err := service.CreateProduct(req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "创建成功", product)
}

// UpdateProduct 更新本地产品
func UpdateProduct(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	product, err := service.UpdateProduct(uint(id), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "更新成功", product)
}

// DeleteProduct 删除本地产品
func DeleteProduct(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := service.DeleteProduct(uint(id)); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "删除成功", nil)
}

// ListPlatformProducts 获取平台产品列表
func ListPlatformProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	authID, _ := strconv.Atoi(c.DefaultQuery("platform_auth_id", "0"))
	keyword := c.Query("keyword")

	products, total, err := service.ListPlatformProducts(uint(authID), keyword, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取平台产品列表失败")
		return
	}

	var list []PlatformProductResponse
	for _, p := range products {
		resp := PlatformProductResponse{
			ID:             p.ID,
			Platform:       p.Platform,
			PlatformAuthID: p.PlatformAuthID,
			PlatformSKU:    p.PlatformSKU,
			Name:           p.Name,
			Image:          p.Image,
			Stock:          p.Stock,
			Price:          p.Price,
			Currency:       p.Currency,
			Status:         p.Status,
			CreatedAt:      p.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:      p.UpdatedAt.Format("2006-01-02 15:04:05"),
		}

		if p.ProductMapping != nil {
			mapping := MappingResponse{
				ID:                p.ProductMapping.ID,
				PlatformProductID: p.ProductMapping.PlatformProductID,
				ProductID:         p.ProductMapping.ProductID,
			}
			if p.ProductMapping.Product != nil {
				mapping.Product = &ProductResponse{
					ID:   p.ProductMapping.Product.ID,
					SKU:  p.ProductMapping.Product.SKU,
					Name: p.ProductMapping.Product.Name,
				}
			}
			resp.ProductMapping = &mapping
		}

		list = append(list, resp)
	}

	response.Success(c, http.StatusOK, "获取成功", PlatformProductListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// MapProduct 关联产品
func MapProduct(c *gin.Context) {
	var req MapProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := service.MapProduct(req); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "关联成功", nil)
}

// UnmapProduct 解除关联
func UnmapProduct(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id")) // platform_product_id
	if err := service.UnmapProduct(uint(id)); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "解除关联成功", nil)
}

// SyncPlatformProducts 同步平台产品（通过任务队列）
func SyncPlatformProducts(c *gin.Context) {
	var req SyncProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 创建同步任务
	task, err := service.CreateSyncTask(req.PlatformAuthID, SyncTaskTypeProduct)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "创建同步任务失败")
		return
	}

	response.Success(c, http.StatusOK, "同步任务已创建", map[string]interface{}{
		"task_id": task.ID,
	})
}

// SyncPlatformProductsDirect 直接同步平台产品（不走任务队列）
func SyncPlatformProductsDirect(c *gin.Context) {
	var req SyncProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 直接执行同步
	if err := service.SyncPlatformProducts(req.PlatformAuthID); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "同步完成", nil)
}

// TriggerSyncTasks 手动触发执行待处理的同步任务
func TriggerSyncTasks(c *gin.Context) {
	go service.ProcessPendingTasks()
	response.Success(c, http.StatusOK, "同步任务处理已触发", nil)
}

// InitProducts 初始化本地产品（根据平台SKU生成）
func InitProducts(c *gin.Context) {
	var req InitProductsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := service.InitProductsFromPlatform(req.PlatformAuthID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "初始化完成", result)
}

// ListSyncTasks 获取同步任务列表
func ListSyncTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")

	tasks, total, err := service.ListSyncTasks(page, pageSize, status)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取任务列表失败")
		return
	}

	response.Success(c, http.StatusOK, "获取成功", map[string]interface{}{
		"list":      tasks,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ========== 入库单相关 ==========

// CreateStockInOrder 创建入库单
func CreateStockInOrder(c *gin.Context) {
	var req CreateStockInOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	order, err := service.CreateStockInOrder(req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 转换响应
	resp := StockInOrderResponse{
		ID:        order.ID,
		OrderNo:   order.OrderNo,
		Status:    order.Status,
		Remark:    order.Remark,
		CreatedAt: order.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	for _, item := range order.Items {
		resp.Items = append(resp.Items, StockInOrderItemResponse{
			ID:          item.ID,
			ProductID:   item.ProductID,
			SKU:         item.SKU,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
		})
	}

	response.Success(c, http.StatusCreated, "入库单创建成功", resp)
}

// ListStockInOrders 获取入库单列表
func ListStockInOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")

	orders, total, err := service.ListStockInOrders(page, pageSize, status)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取入库单列表失败")
		return
	}

	var list []StockInOrderResponse
	for _, order := range orders {
		warehouseName := ""
		if order.Warehouse != nil {
			warehouseName = order.Warehouse.Name
		}
		resp := StockInOrderResponse{
			ID:            order.ID,
			OrderNo:       order.OrderNo,
			WarehouseID:   order.WarehouseID,
			WarehouseName: warehouseName,
			Status:        order.Status,
			Remark:        order.Remark,
			CreatedAt:     order.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		for _, item := range order.Items {
			resp.Items = append(resp.Items, StockInOrderItemResponse{
				ID:          item.ID,
				ProductID:   item.ProductID,
				SKU:         item.SKU,
				ProductName: item.ProductName,
				Quantity:    item.Quantity,
			})
		}
		list = append(list, resp)
	}

	response.Success(c, http.StatusOK, "获取成功", StockInOrderListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// GetStockInOrder 获取入库单详情
func GetStockInOrder(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	order, err := service.GetStockInOrder(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	warehouseName := ""
	if order.Warehouse != nil {
		warehouseName = order.Warehouse.Name
	}
	resp := StockInOrderResponse{
		ID:            order.ID,
		OrderNo:       order.OrderNo,
		WarehouseID:   order.WarehouseID,
		WarehouseName: warehouseName,
		Status:        order.Status,
		Remark:        order.Remark,
		CreatedAt:     order.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	for _, item := range order.Items {
		resp.Items = append(resp.Items, StockInOrderItemResponse{
			ID:          item.ID,
			ProductID:   item.ProductID,
			SKU:         item.SKU,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
		})
	}

	response.Success(c, http.StatusOK, "获取成功", resp)
}

// ========== 仓库相关 ==========

// ListWarehouses 获取仓库列表（仅活跃状态，用于下拉选择）
func ListWarehouses(c *gin.Context) {
	warehouses, err := service.ListWarehouses()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取仓库列表失败")
		return
	}

	var list []WarehouseResponse
	for _, w := range warehouses {
		list = append(list, WarehouseResponse{
			ID:        w.ID,
			Code:      w.Code,
			Name:      w.Name,
			Type:      w.Type,
			Address:   w.Address,
			Status:    w.Status,
			CreatedAt: w.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	response.Success(c, http.StatusOK, "获取成功", list)
}

// ListAvailableWarehouses 获取当前用户可用的仓库列表（用于入库单等业务场景）
// 后续可根据用户权限进行过滤
func ListAvailableWarehouses(c *gin.Context) {
	// TODO: 从上下文获取用户信息，根据权限过滤仓库
	// userID := c.GetUint("user_id")
	// warehouses, err := service.ListWarehousesByUser(userID)

	// 当前返回所有活跃仓库
	warehouses, err := service.ListWarehouses()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取仓库列表失败")
		return
	}

	var list []WarehouseResponse
	for _, w := range warehouses {
		list = append(list, WarehouseResponse{
			ID:        w.ID,
			Code:      w.Code,
			Name:      w.Name,
			Type:      w.Type,
			Address:   w.Address,
			Status:    w.Status,
			CreatedAt: w.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	response.Success(c, http.StatusOK, "获取成功", WarehouseListResponse{
		List:  list,
		Total: int64(len(list)),
	})
}

// ListAllWarehouses 获取所有仓库（支持按类型筛选，用于仓库管理页面）
func ListAllWarehouses(c *gin.Context) {
	warehouseType := c.Query("type")

	warehouses, err := service.ListAllWarehouses(warehouseType)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取仓库列表失败")
		return
	}

	var list []WarehouseResponse
	for _, w := range warehouses {
		list = append(list, WarehouseResponse{
			ID:        w.ID,
			Code:      w.Code,
			Name:      w.Name,
			Type:      w.Type,
			Address:   w.Address,
			Status:    w.Status,
			CreatedAt: w.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	response.Success(c, http.StatusOK, "获取成功", list)
}

// CreateWarehouse 创建仓库
func CreateWarehouse(c *gin.Context) {
	var req CreateWarehouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	warehouse, err := service.CreateWarehouse(req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "创建成功", WarehouseResponse{
		ID:        warehouse.ID,
		Code:      warehouse.Code,
		Name:      warehouse.Name,
		Type:      warehouse.Type,
		Address:   warehouse.Address,
		Status:    warehouse.Status,
		CreatedAt: warehouse.CreatedAt.Format("2006-01-02 15:04:05"),
	})
}

// ========== 库存相关 ==========

// ListInventory 获取库存明细列表
func ListInventory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	warehouseID, _ := strconv.Atoi(c.DefaultQuery("warehouse_id", "0"))
	keyword := c.Query("keyword")

	inventories, total, err := service.ListInventory(uint(warehouseID), keyword, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取库存列表失败")
		return
	}

	var list []InventoryResponse
	for _, inv := range inventories {
		resp := InventoryResponse{
			ID:             inv.ID,
			ProductID:      inv.ProductID,
			WarehouseID:    inv.WarehouseID,
			SKU:            inv.SKU,
			AvailableStock: inv.AvailableStock,
			LockedStock:    inv.LockedStock,
			InTransitStock: inv.InTransitStock,
			TotalStock:     inv.TotalStock(),
			UpdatedAt:      inv.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		if inv.Product != nil {
			resp.ProductName = inv.Product.Name
			resp.ProductImage = inv.Product.Image
		}
		if inv.Warehouse != nil {
			resp.WarehouseCode = inv.Warehouse.Code
			resp.WarehouseName = inv.Warehouse.Name
		}
		list = append(list, resp)
	}

	response.Success(c, http.StatusOK, "获取成功", InventoryListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// UpdateInventory 更新库存
func UpdateInventory(c *gin.Context) {
	var req UpdateInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	inventory, err := service.UpdateInventory(req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	resp := InventoryResponse{
		ID:             inventory.ID,
		ProductID:      inventory.ProductID,
		WarehouseID:    inventory.WarehouseID,
		SKU:            inventory.SKU,
		AvailableStock: inventory.AvailableStock,
		LockedStock:    inventory.LockedStock,
		InTransitStock: inventory.InTransitStock,
		TotalStock:     inventory.TotalStock(),
		UpdatedAt:      inventory.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	if inventory.Product != nil {
		resp.ProductName = inventory.Product.Name
		resp.ProductImage = inventory.Product.Image
	}
	if inventory.Warehouse != nil {
		resp.WarehouseCode = inventory.Warehouse.Code
		resp.WarehouseName = inventory.Warehouse.Name
	}

	response.Success(c, http.StatusOK, "更新成功", resp)
}

// InitInventory 初始化库存（从产品表生成库存记录）
func InitInventory(c *gin.Context) {
	warehouseID, _ := strconv.Atoi(c.DefaultQuery("warehouse_id", "1"))

	created, err := service.InitInventoryFromProducts(uint(warehouseID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "初始化完成", map[string]int{
		"created": created,
	})
}

// ========== 供应商/采购信息相关 ==========

// ListSuppliers 获取供应商列表
func ListSuppliers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	productID, _ := strconv.Atoi(c.DefaultQuery("product_id", "0"))
	keyword := c.Query("keyword")
	status := c.Query("status")

	suppliers, total, err := service.ListSuppliers(uint(productID), keyword, status, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取供应商列表失败")
		return
	}

	var list []SupplierResponse
	for _, s := range suppliers {
		resp := SupplierResponse{
			ID:            s.ID,
			ProductID:     s.ProductID,
			SupplierName:  s.SupplierName,
			PurchaseLink:  s.PurchaseLink,
			UnitPrice:     s.UnitPrice,
			Currency:      s.Currency,
			MinOrderQty:   s.MinOrderQty,
			LeadTime:      s.LeadTime,
			EstimatedDays: s.EstimatedDays,
			Remark:        s.Remark,
			IsDefault:     s.IsDefault,
			Status:        s.Status,
			CreatedAt:     s.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:     s.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		if s.Product != nil {
			resp.ProductSKU = s.Product.SKU
			resp.ProductName = s.Product.Name
		}
		list = append(list, resp)
	}

	response.Success(c, http.StatusOK, "获取成功", SupplierListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// GetProductSuppliers 获取产品的供应商列表
func GetProductSuppliers(c *gin.Context) {
	productID, _ := strconv.Atoi(c.Param("id"))

	suppliers, err := service.ListSuppliersByProductID(uint(productID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取供应商列表失败")
		return
	}

	var list []SupplierResponse
	for _, s := range suppliers {
		list = append(list, SupplierResponse{
			ID:            s.ID,
			ProductID:     s.ProductID,
			SupplierName:  s.SupplierName,
			PurchaseLink:  s.PurchaseLink,
			UnitPrice:     s.UnitPrice,
			Currency:      s.Currency,
			MinOrderQty:   s.MinOrderQty,
			LeadTime:      s.LeadTime,
			EstimatedDays: s.EstimatedDays,
			Remark:        s.Remark,
			IsDefault:     s.IsDefault,
			Status:        s.Status,
			CreatedAt:     s.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:     s.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	response.Success(c, http.StatusOK, "获取成功", list)
}

// CreateSupplier 创建供应商
func CreateSupplier(c *gin.Context) {
	var req CreateSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	supplier, err := service.CreateSupplier(req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "创建成功", SupplierResponse{
		ID:            supplier.ID,
		ProductID:     supplier.ProductID,
		SupplierName:  supplier.SupplierName,
		PurchaseLink:  supplier.PurchaseLink,
		UnitPrice:     supplier.UnitPrice,
		Currency:      supplier.Currency,
		MinOrderQty:   supplier.MinOrderQty,
		LeadTime:      supplier.LeadTime,
		EstimatedDays: supplier.EstimatedDays,
		Remark:        supplier.Remark,
		IsDefault:     supplier.IsDefault,
		Status:        supplier.Status,
		CreatedAt:     supplier.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     supplier.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

// UpdateSupplier 更新供应商
func UpdateSupplier(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req UpdateSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	supplier, err := service.UpdateSupplier(uint(id), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "更新成功", SupplierResponse{
		ID:            supplier.ID,
		ProductID:     supplier.ProductID,
		SupplierName:  supplier.SupplierName,
		PurchaseLink:  supplier.PurchaseLink,
		UnitPrice:     supplier.UnitPrice,
		Currency:      supplier.Currency,
		MinOrderQty:   supplier.MinOrderQty,
		LeadTime:      supplier.LeadTime,
		EstimatedDays: supplier.EstimatedDays,
		Remark:        supplier.Remark,
		IsDefault:     supplier.IsDefault,
		Status:        supplier.Status,
		CreatedAt:     supplier.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     supplier.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

// DeleteSupplier 删除供应商
func DeleteSupplier(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := service.DeleteSupplier(uint(id)); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "删除成功", nil)
}
