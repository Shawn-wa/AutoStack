package shipping

// TemplateQuery 模板查询条件
type TemplateQuery struct {
	Page     int
	PageSize int
	Keyword  string // 按名称搜索
	Status   string // 按状态筛选
}
