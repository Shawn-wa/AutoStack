package repository

import (
	"context"

	"gorm.io/gorm"
)

// ctxKey 用于 context 中存储事务的键类型
type ctxKey string

const txKey ctxKey = "db_tx"

// WithTx 将事务注入 context
func WithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey, tx)
}

// GetDB 从 context 获取数据库连接
// 如果 context 中存在事务，返回事务连接；否则返回默认数据库连接
func GetDB(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey).(*gorm.DB); ok && tx != nil {
		return tx
	}
	return defaultDB
}

// HasTx 检查 context 中是否存在事务
func HasTx(ctx context.Context) bool {
	tx, ok := ctx.Value(txKey).(*gorm.DB)
	return ok && tx != nil
}
