package store

import "gorm.io/gorm"

// Models 返回 trendln 运行时负责的 GORM 模型。
func Models() []any {
	return []any{
		&TrendlnRuntimeConfig{},
		&TrendlnFunctionGovernance{},
		&TrendlnTrialHistory{},
	}
}

// InitAll 初始化 trendln runtime 表结构。
func InitAll(db *gorm.DB) error {
	if err := NewTrendlnRuntimeConfigStore(db).InitTables(); err != nil {
		return err
	}
	if err := NewTrendlnFunctionGovernanceStore(db).InitTables(); err != nil {
		return err
	}
	if err := NewTrendlnTrialHistoryStore(db).InitTables(); err != nil {
		return err
	}
	return nil
}
