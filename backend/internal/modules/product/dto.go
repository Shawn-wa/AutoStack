package product

// CreateProductRequest 创建产品请求
type CreateProductRequest struct {
	SKU        string  `json:"sku" binding:"required"`
	Name       string  `json:"name" binding:"required"`
	Image      string  `json:"image"`
	CostPrice  float64 `json:"cost_price"`
	Weight     float64 `json:"weight"`
	Dimensions string  `json:"dimensions"`
}

// UpdateProductRequest 更新产品请求
type UpdateProductRequest struct {
	Name       string  `json:"name"`
	Image      string  `json:"image"`
	CostPrice  float64 `json:"cost_price"`
	Weight     float64 `json:"weight"`
	Dimensions string  `json:"dimensions"`
}

// ProductResponse 产品响应
type ProductResponse struct {
	ID         uint    `json:"id"`
	SKU        string  `json:"sku"`
	Name       string  `json:"name"`
	Image      string  `json:"image"`
	CostPrice  float64 `json:"cost_price"`
	Weight     float64 `json:"weight"`
	Dimensions string  `json:"dimensions"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

// ProductListResponse 产品列表响应
type ProductListResponse struct {
	List     []ProductResponse `json:"list"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

// PlatformProductResponse 平台产品响应
type PlatformProductResponse struct {
	ID             uint             `json:"id"`
	Platform       string           `json:"platform"`
	PlatformAuthID uint             `json:"platform_auth_id"`
	PlatformSKU    string           `json:"platform_sku"`
	Name           string           `json:"name"`
	Image          string           `json:"image"`
	Stock          int              `json:"stock"`
	Price          float64          `json:"price"`
	Currency       string           `json:"currency"`
	Status         string           `json:"status"`
	CreatedAt      string           `json:"created_at"`
	UpdatedAt      string           `json:"updated_at"`
	ProductMapping *MappingResponse `json:"product_mapping,omitempty"`
}

// PlatformProductListResponse 平台产品列表响应
type PlatformProductListResponse struct {
	List     []PlatformProductResponse `json:"list"`
	Total    int64                     `json:"total"`
	Page     int                       `json:"page"`
	PageSize int                       `json:"page_size"`
}

// MappingResponse 映射关系响应
type MappingResponse struct {
	ID                uint             `json:"id"`
	PlatformProductID uint             `json:"platform_product_id"`
	ProductID         uint             `json:"product_id"`
	Product           *ProductResponse `json:"product,omitempty"`
}

// MapProductRequest 映射产品请求
type MapProductRequest struct {
	PlatformProductID uint `json:"platform_product_id" binding:"required"`
	ProductID         uint `json:"product_id" binding:"required"`
}

// SyncProductRequest 同步产品请求
type SyncProductRequest struct {
	PlatformAuthID uint `json:"platform_auth_id" binding:"required"`
}

// InitProductsRequest 初始化本地产品请求
type InitProductsRequest struct {
	PlatformAuthID uint `json:"platform_auth_id"` // 可选，指定平台授权ID；0表示全部
}

// InitProductsResponse 初始化本地产品响应
type InitProductsResponse struct {
	TotalPlatformProducts int `json:"total_platform_products"` // 平台产品总数
	SkippedMapped         int `json:"skipped_mapped"`          // 跳过（已有映射）
	SkippedExisting       int `json:"skipped_existing"`        // 跳过（SKU已存在但已关联）
	CreatedProducts       int `json:"created_products"`        // 新创建的本地产品数
	CreatedMappings       int `json:"created_mappings"`        // 新创建的映射数
}

// ========== 入库单相关 ==========

// StockInOrderItemRequest 入库单明细请求
type StockInOrderItemRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,min=1"`
}

// CreateStockInOrderRequest 创建入库单请求
type CreateStockInOrderRequest struct {
	WarehouseID uint                      `json:"warehouse_id" binding:"required"`
	Items       []StockInOrderItemRequest `json:"items" binding:"required,min=1"`
	Remark      string                    `json:"remark"`
}

// StockInOrderItemResponse 入库单明细响应
type StockInOrderItemResponse struct {
	ID          uint   `json:"id"`
	ProductID   uint   `json:"product_id"`
	SKU         string `json:"sku"`
	ProductName string `json:"product_name"`
	Quantity    int    `json:"quantity"`
}

// StockInOrderResponse 入库单响应
type StockInOrderResponse struct {
	ID            uint                       `json:"id"`
	OrderNo       string                     `json:"order_no"`
	WarehouseID   uint                       `json:"warehouse_id"`
	WarehouseName string                     `json:"warehouse_name"`
	Status        string                     `json:"status"`
	Remark        string                     `json:"remark"`
	Items         []StockInOrderItemResponse `json:"items"`
	CreatedAt     string                     `json:"created_at"`
}

// StockInOrderListResponse 入库单列表响应
type StockInOrderListResponse struct {
	List     []StockInOrderResponse `json:"list"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
}

// ========== 仓库相关 ==========

// CreateWarehouseRequest 创建仓库请求
type CreateWarehouseRequest struct {
	Code    string `json:"code" binding:"required"`
	Name    string `json:"name" binding:"required"`
	Type    string `json:"type"`    // 仓库类型：local/overseas/fba/third/virtual
	Address string `json:"address"`
}

// WarehouseResponse 仓库响应
type WarehouseResponse struct {
	ID        uint   `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Address   string `json:"address"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// WarehouseListResponse 仓库列表响应
type WarehouseListResponse struct {
	List  []WarehouseResponse `json:"list"`
	Total int64               `json:"total"`
}

// ========== 库存相关 ==========

// InventoryResponse 库存明细响应
type InventoryResponse struct {
	ID             uint   `json:"id"`
	ProductID      uint   `json:"product_id"`
	WarehouseID    uint   `json:"warehouse_id"`
	SKU            string `json:"sku"`
	ProductName    string `json:"product_name"`
	ProductImage   string `json:"product_image"`
	WarehouseCode  string `json:"warehouse_code"`
	WarehouseName  string `json:"warehouse_name"`
	AvailableStock int    `json:"available_stock"`
	LockedStock    int    `json:"locked_stock"`
	InTransitStock int    `json:"in_transit_stock"`
	TotalStock     int    `json:"total_stock"`
	UpdatedAt      string `json:"updated_at"`
}

// InventoryListResponse 库存列表响应
type InventoryListResponse struct {
	List     []InventoryResponse `json:"list"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}

// UpdateInventoryRequest 更新库存请求
type UpdateInventoryRequest struct {
	ProductID      uint `json:"product_id" binding:"required"`
	WarehouseID    uint `json:"warehouse_id" binding:"required"`
	AvailableStock *int `json:"available_stock"`
	LockedStock    *int `json:"locked_stock"`
	InTransitStock *int `json:"in_transit_stock"`
}
