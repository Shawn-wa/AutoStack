package order

import "sync"

// OrderStatus 订单状态枚举
type OrderStatus struct {
	Value       string // 状态值（存储在数据库中）
	Name        string // 显示名称
	Description string // 状态描述
}

// 系统标准订单状态定义
var (
	StatusPending     = OrderStatus{Value: "pending", Name: "待处理", Description: "订单已创建，等待处理"}
	StatusReadyToShip = OrderStatus{Value: "ready_to_ship", Name: "待发货", Description: "订单已处理，等待发货"}
	StatusShipped     = OrderStatus{Value: "shipped", Name: "已发货", Description: "订单已发货，运输中"}
	StatusDelivered   = OrderStatus{Value: "delivered", Name: "已签收", Description: "订单已送达签收"}
	StatusCancelled   = OrderStatus{Value: "cancelled", Name: "已取消", Description: "订单已取消"}
)

// 状态值常量（便于直接使用）
const (
	OrderStatusPending     = "pending"
	OrderStatusReadyToShip = "ready_to_ship"
	OrderStatusShipped     = "shipped"
	OrderStatusDelivered   = "delivered"
	OrderStatusCancelled   = "cancelled"
)

// AllOrderStatuses 所有订单状态列表
var AllOrderStatuses = []OrderStatus{
	StatusPending,
	StatusReadyToShip,
	StatusShipped,
	StatusDelivered,
	StatusCancelled,
}

// statusRegistry 平台状态映射注册表
// 结构: map[platform]map[platformStatus]systemStatus
var statusRegistry = make(map[string]map[string]string)
var statusMutex sync.RWMutex

// RegisterPlatformStatusMapping 注册平台状态映射
// platform: 平台标识（如 "ozon"）
// mappings: 平台状态到系统状态的映射 map[platformStatus]systemStatus
func RegisterPlatformStatusMapping(platform string, mappings map[string]string) {
	statusMutex.Lock()
	defer statusMutex.Unlock()
	statusRegistry[platform] = mappings
}

// MapPlatformStatus 将平台状态映射为系统状态
// platform: 平台标识
// platformStatus: 平台原始状态
// 返回系统状态值，如果未找到映射则返回默认状态（pending）
func MapPlatformStatus(platform, platformStatus string) string {
	statusMutex.RLock()
	defer statusMutex.RUnlock()

	if platformMappings, ok := statusRegistry[platform]; ok {
		if systemStatus, ok := platformMappings[platformStatus]; ok {
			return systemStatus
		}
	}
	return OrderStatusPending // 默认返回待处理状态
}

// GetOrderStatusByValue 根据状态值获取状态枚举
func GetOrderStatusByValue(value string) *OrderStatus {
	for _, status := range AllOrderStatuses {
		if status.Value == value {
			return &status
		}
	}
	return nil
}

// GetOrderStatusName 获取状态显示名称
func GetOrderStatusName(value string) string {
	if status := GetOrderStatusByValue(value); status != nil {
		return status.Name
	}
	return value
}

// GetPlatformStatusMappings 获取指定平台的状态映射（用于调试或展示）
func GetPlatformStatusMappings(platform string) map[string]string {
	statusMutex.RLock()
	defer statusMutex.RUnlock()

	if mappings, ok := statusRegistry[platform]; ok {
		// 返回副本，避免外部修改
		result := make(map[string]string, len(mappings))
		for k, v := range mappings {
			result[k] = v
		}
		return result
	}
	return nil
}

