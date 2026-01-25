package product

// ProductQuery 产品查询条件
type ProductQuery struct {
	Page     int
	PageSize int
	Keyword  string
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
