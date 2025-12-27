package order

import (
	"time"
)

// PlatformAdapter 平台适配器接口
type PlatformAdapter interface {
	// GetPlatformName 获取平台标识
	GetPlatformName() string
	// GetPlatformLabel 获取平台显示名称
	GetPlatformLabel() string
	// GetCredentialFields 获取凭证字段定义
	GetCredentialFields() []CredentialField
	// TestConnection 测试连接
	TestConnection(credentials string) error
	// SyncOrders 同步订单
	SyncOrders(credentials string, since, to time.Time) ([]*Order, error)
	// GetCommissions 获取佣金信息
	GetCommissions(credentials string, since, to time.Time) (map[string]*CommissionData, error)
}

// PlatformAdapterWithLog 带日志记录的平台适配器接口（可选实现）
type PlatformAdapterWithLog interface {
	PlatformAdapter
	// TestConnectionWithLog 测试连接（带日志记录）
	TestConnectionWithLog(credentials string, platformAuthID uint) error
	// SyncOrdersWithLog 同步订单（带日志记录）
	SyncOrdersWithLog(credentials string, since, to time.Time, platformAuthID uint) ([]*Order, error)
	// GetCommissionsWithLog 获取佣金信息（带日志记录）
	GetCommissionsWithLog(credentials string, since, to time.Time, platformAuthID uint) (map[string]*CommissionData, error)
}

// 注册的适配器
var adapters = make(map[string]PlatformAdapter)

// RegisterAdapter 注册适配器
func RegisterAdapter(adapter PlatformAdapter) {
	adapters[adapter.GetPlatformName()] = adapter
}

// GetAdapter 获取适配器
func GetAdapter(platform string) PlatformAdapter {
	return adapters[platform]
}

// GetAllPlatforms 获取所有支持的平台
func GetAllPlatforms() []PlatformInfo {
	platforms := make([]PlatformInfo, 0, len(adapters))
	for _, adapter := range adapters {
		platforms = append(platforms, PlatformInfo{
			Name:   adapter.GetPlatformName(),
			Label:  adapter.GetPlatformLabel(),
			Fields: adapter.GetCredentialFields(),
		})
	}
	return platforms
}

// InitAdapters 初始化所有适配器
// 注意：适配器通过各自的 init() 函数自动注册
func InitAdapters() {
	// 适配器已通过 platforms 包的 init() 自动注册
}
