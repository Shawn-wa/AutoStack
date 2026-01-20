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
