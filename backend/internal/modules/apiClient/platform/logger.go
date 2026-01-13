package platform

import (
	"time"

	"gorm.io/gorm"
)

// RequestLog 请求日志模型
type RequestLog struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	PlatformAuthID uint      `gorm:"index;not null" json:"platform_auth_id"`
	Platform       string    `gorm:"size:50;not null;index" json:"platform"`
	RequestType    string    `gorm:"size:50;not null;index" json:"request_type"`
	RequestURL     string    `gorm:"size:500;not null" json:"request_url"`
	RequestMethod  string    `gorm:"size:10;not null" json:"request_method"`
	RequestHeaders string    `gorm:"type:text" json:"-"`
	RequestBody    string    `gorm:"type:longtext" json:"request_body"`
	ResponseStatus int       `gorm:"not null" json:"response_status"`
	ResponseBody   string    `gorm:"type:longtext" json:"response_body"`
	Duration       int64     `gorm:"not null" json:"duration"`
	ErrorMessage   string    `gorm:"size:1000" json:"error_message"`
	CreatedAt      time.Time `gorm:"index" json:"created_at"`
}

// TableName 表名
func (RequestLog) TableName() string {
	return "orders_request_log"
}

// RequestLogger 请求日志记录器
type RequestLogger struct {
	db             *gorm.DB
	platformAuthID uint
	platform       string
}

// NewRequestLogger 创建请求日志记录器
func NewRequestLogger(db *gorm.DB, platformAuthID uint, platform string) *RequestLogger {
	return &RequestLogger{
		db:             db,
		platformAuthID: platformAuthID,
		platform:       platform,
	}
}

// LogRequest 记录请求日志
func (l *RequestLogger) LogRequest(log *RequestLog) error {
	if l.db == nil {
		return nil
	}
	log.PlatformAuthID = l.platformAuthID
	log.Platform = l.platform
	log.CreatedAt = time.Now()
	return l.db.Create(log).Error
}

// RequestType 请求类型常量
const (
	RequestTypeTestConnect      = "TestConnect"
	RequestTypeOrderList        = "OrderList"
	RequestTypeFinance          = "Finance"
	RequestTypeCashFlow         = "CashFlow"
	RequestTypeMutualSettlement = "MutualSettlement"
	RequestTypeReportInfo       = "ReportInfo"
)
