package store

import "gorm.io/gorm"

// Models 返回 ta-lib 运行时负责的 GORM 模型。
func Models() []any {
	return []any{
		&TaLibRuntimeConfig{},
		&TaLibFunctionGovernance{},
		&TaLibTrialHistory{},
	}
}

// InitAll 初始化 ta-lib 运行时表结构。
func InitAll(db *gorm.DB) error {
	if err := NewTaLibRuntimeConfigStore(db).InitTables(); err != nil {
		return err
	}
	if err := NewTaLibFunctionGovernanceStore(db).InitTables(); err != nil {
		return err
	}
	if err := NewTaLibTrialHistoryStore(db).InitTables(); err != nil {
		return err
	}
	return nil
}
