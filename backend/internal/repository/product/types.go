package product

// ProductQuery 产品查询条件
type ProductQuery struct {
	Page        int
	PageSize    int
	Keyword     string
	WarehouseID uint // 仓库ID筛选
}

// PlatformProductQuery 平台产品查询条件
type PlatformProductQuery struct {
	Page           int
	PageSize       int
	PlatformAuthID uint
	Keyword        string
}

// SyncTaskQuery 同步任务查询条件
type SyncTaskQuery struct {
	Page     int
	PageSize int
	Status   string
}

// ProductSupplierQuery 产品供应商查询条件
type ProductSupplierQuery struct {
	Page      int
	PageSize  int
	ProductID uint   // 按产品ID筛选
	Keyword   string // 按供应商名称搜索
	Status    string // 按状态筛选
}
