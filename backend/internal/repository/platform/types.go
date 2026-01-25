package platform

// PlatformAuthQuery 平台授权查询条件
type PlatformAuthQuery struct {
	Page     int
	PageSize int
	UserID   uint
	Platform string
	Status   *int
}

// RequestLogQuery 请求日志查询条件
type RequestLogQuery struct {
	Page           int
	PageSize       int
	PlatformAuthID uint
	RequestType    string
}

// CashFlowQuery 现金流查询条件
type CashFlowQuery struct {
	Page           int
	PageSize       int
	UserID         uint
	PlatformAuthID uint
}
