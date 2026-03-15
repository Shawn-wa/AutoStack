package product

// ProductQuery 产品查询条件
type ProductQuery struct {
	CompanyID  uint
	Page        int
	PageSize    int
	Keyword     string
	WarehouseID uint // 仓库ID筛选
}

// PlatformProductQuery 平台产品查询条件
type PlatformProductQuery struct {
	CompanyID      uint
	Page           int
	PageSize       int
	Platform       string
	PlatformAuthID uint
	Keyword        string
	MappedFilter   string // all, mapped, unmapped
}

// SyncTaskQuery 同步任务查询条件
type SyncTaskQuery struct {
	CompanyID uint
	Page     int
	PageSize int
	Status   string
}

// ProductSupplierQuery 产品供应商查询条件
type ProductSupplierQuery struct {
	CompanyID uint
	Page      int
	PageSize  int
	ProductID uint   // 按产品ID筛选
	Keyword   string // 按供应商名称搜索
	Status    string // 按状态筛选
}
