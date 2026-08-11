package store

import "gorm.io/gorm"

// Models 返回 pandas-ta 运行时负责的 GORM 模型。
func Models() []any {
	return []any{
		&PandasTaRuntimeConfig{},
		&PandasTaFunctionGovernance{},
		&PandasTaTrialHistory{},
	}
}

// InitAll 初始化 pandas-ta runtime 表结构。
func InitAll(db *gorm.DB) error {
	if err := NewPandasTaRuntimeConfigStore(db).InitTables(); err != nil {
		return err
	}
	if err := NewPandasTaFunctionGovernanceStore(db).InitTables(); err != nil {
		return err
	}
	if err := NewPandasTaTrialHistoryStore(db).InitTables(); err != nil {
		return err
	}
	return nil
}
