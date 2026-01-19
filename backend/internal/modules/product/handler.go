package product

import (
	"autostack/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var service = &Service{}

// ListProducts 获取本地产品列表
func ListProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	products, total, err := service.ListProducts(page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取产品列表失败")
		return
	}

	var list []ProductResponse
	for _, p := range products {
		list = append(list, ProductResponse{
			ID:         p.ID,
			SKU:        p.SKU,
			Name:       p.Name,
			Image:      p.Image,
			CostPrice:  p.CostPrice,
			Weight:     p.Weight,
			Dimensions: p.Dimensions,
			CreatedAt:  p.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:  p.UpdatedAt.Format("2006-01-02 15:04:05"),
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

	products, total, err := service.ListPlatformProducts(uint(authID), page, pageSize)
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

// SyncPlatformProducts 同步平台产品
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

// TriggerSyncTasks 手动触发执行待处理的同步任务
func TriggerSyncTasks(c *gin.Context) {
	go service.ProcessPendingTasks()
	response.Success(c, http.StatusOK, "同步任务处理已触发", nil)
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
