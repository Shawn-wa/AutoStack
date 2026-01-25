package order

// 本文件为类型别名定义，实际实体已迁移至 repository 层
// 保持向后兼容，避免修改现有代码的导入路径

import (
	orderRepo "autostack/internal/repository/order"
	platformRepo "autostack/internal/repository/platform"
)

// ========== 平台域类型别名 ==========

// PlatformAuth 平台授权模型
type PlatformAuth = platformRepo.PlatformAuth

// OrdersRequestLog 订单请求日志模型
type OrdersRequestLog = platformRepo.OrdersRequestLog

// CashFlowStatement 现金流报表模型
type CashFlowStatement = platformRepo.CashFlowStatement

// ========== 订单域类型别名 ==========

// Order 订单模型
type Order = orderRepo.Order

// OrderItem 订单商品模型
type OrderItem = orderRepo.OrderItem

// OrderDailyStat 订单每日统计模型
type OrderDailyStat = orderRepo.OrderDailyStat

// CommissionData 佣金数据
type CommissionData = orderRepo.CommissionData

// ========== 平台常量别名 ==========

const (
	PlatformOzon = platformRepo.PlatformOzon
)

// ========== 授权状态常量别名 ==========

const (
	AuthStatusDisabled = platformRepo.AuthStatusDisabled
	AuthStatusActive   = platformRepo.AuthStatusActive
	AuthStatusExpired  = platformRepo.AuthStatusExpired
)

// ========== 请求日志类型常量别名 ==========

const (
	RequestTypeOrderList   = platformRepo.RequestTypeOrderList
	RequestTypeFinance     = platformRepo.RequestTypeFinance
	RequestTypeTestConnect = platformRepo.RequestTypeTestConnect
	RequestTypeCashFlow    = platformRepo.RequestTypeCashFlow
)
