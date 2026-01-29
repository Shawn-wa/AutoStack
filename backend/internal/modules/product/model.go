package product

// 本文件为类型别名定义，实际实体已迁移至 repository 层
// 保持向后兼容，避免修改现有代码的导入路径

import (
	inventoryRepo "autostack/internal/repository/inventory"
	productRepo "autostack/internal/repository/product"
)

// ========== 产品域类型别名 ==========

// Product 本地产品模型
type Product = productRepo.Product

// PlatformProduct 平台产品模型
type PlatformProduct = productRepo.PlatformProduct

// ProductMapping 产品映射关系
type ProductMapping = productRepo.ProductMapping

// PlatformSyncTask 平台同步任务
type PlatformSyncTask = productRepo.PlatformSyncTask

// ProductSupplier 产品供应商/采购信息
type ProductSupplier = productRepo.ProductSupplier

// ========== 库存域类型别名 ==========

// Warehouse 仓库
type Warehouse = inventoryRepo.Warehouse

// WarehouseCenterInventory 仓库库存明细
type WarehouseCenterInventory = inventoryRepo.WarehouseCenterInventory

// StockInOrder 入库单
type StockInOrder = inventoryRepo.StockInOrder

// StockInOrderItem 入库单明细
type StockInOrderItem = inventoryRepo.StockInOrderItem

// ========== 同步任务常量别名 ==========

const (
	SyncTaskTypeProduct    = productRepo.SyncTaskTypeProduct
	SyncTaskTypeOrder      = productRepo.SyncTaskTypeOrder
	SyncTaskTypeCommission = productRepo.SyncTaskTypeCommission
	SyncTaskTypeCashFlow   = productRepo.SyncTaskTypeCashFlow

	SyncTaskStatusPending = productRepo.SyncTaskStatusPending
	SyncTaskStatusRunning = productRepo.SyncTaskStatusRunning
	SyncTaskStatusSuccess = productRepo.SyncTaskStatusSuccess
	SyncTaskStatusFailed  = productRepo.SyncTaskStatusFailed
)

// ========== 供应商常量别名 ==========

const (
	SupplierStatusActive   = productRepo.SupplierStatusActive
	SupplierStatusInactive = productRepo.SupplierStatusInactive
)

// ========== 入库单状态常量别名 ==========

const (
	StockInStatusPending   = inventoryRepo.StockInStatusPending
	StockInStatusCompleted = inventoryRepo.StockInStatusCompleted
	StockInStatusCancelled = inventoryRepo.StockInStatusCancelled
)

// ========== 仓库常量别名 ==========

const (
	WarehouseStatusActive   = inventoryRepo.WarehouseStatusActive
	WarehouseStatusInactive = inventoryRepo.WarehouseStatusInactive

	WarehouseTypeLocal    = inventoryRepo.WarehouseTypeLocal
	WarehouseTypeOverseas = inventoryRepo.WarehouseTypeOverseas
	WarehouseTypeFBA      = inventoryRepo.WarehouseTypeFBA
	WarehouseTypeThird    = inventoryRepo.WarehouseTypeThird
	WarehouseTypeVirtual  = inventoryRepo.WarehouseTypeVirtual
)
