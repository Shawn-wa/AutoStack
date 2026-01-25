package inventory

// WarehouseQuery 仓库查询条件
type WarehouseQuery struct {
	Type   string // 仓库类型筛选
	Status string // 状态筛选
}

// InventoryQuery 库存查询条件
type InventoryQuery struct {
	Page        int
	PageSize    int
	WarehouseID uint
	Keyword     string
}

// StockInOrderQuery 入库单查询条件
type StockInOrderQuery struct {
	Page     int
	PageSize int
	Status   string
}

// StockSummary 库存汇总
type StockSummary struct {
	SKU            string
	AvailableStock int
}
