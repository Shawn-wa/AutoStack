package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// TxManager 事务管理器接口
type TxManager interface {
	// WithTransaction 在事务中执行函数，自动处理 commit/rollback
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	// DB 获取数据库连接
	DB() *gorm.DB
}

// gormTxManager GORM 事务管理器实现
type gormTxManager struct {
	db *gorm.DB
}

// NewTxManager 创建事务管理器
func NewTxManager(db *gorm.DB) TxManager {
	return &gormTxManager{db: db}
}

// WithTransaction 在事务中执行函数
// 如果 context 中已存在事务，则复用该事务（嵌套事务场景）
// 如果函数返回错误，自动回滚；否则自动提交
func (m *gormTxManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// 如果已在事务中，直接执行（复用现有事务）
	if HasTx(ctx) {
		return fn(ctx)
	}

	// 开启新事务
	tx := m.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("开启事务失败: %w", tx.Error)
	}

	// 将事务注入 context
	txCtx := WithTx(ctx, tx)

	// 执行业务函数
	if err := fn(txCtx); err != nil {
		// 回滚事务
		if rbErr := tx.Rollback().Error; rbErr != nil {
			return fmt.Errorf("回滚事务失败: %v, 原始错误: %w", rbErr, err)
		}
		return err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// DB 获取数据库连接
func (m *gormTxManager) DB() *gorm.DB {
	return m.db
}
