package store

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaLibRuntimeConfig struct {
	ID                  string    `gorm:"column:id;type:varchar(64);primaryKey"`
	BaseURL             string    `gorm:"column:base_url;type:text;not null"`
	Enabled             bool      `gorm:"column:enabled;not null"`
	TimeoutMS           int       `gorm:"column:timeout_ms;not null"`
	HealthStatus        string    `gorm:"column:health_status;type:varchar(64);not null"`
	LastHealthAt        time.Time `gorm:"column:last_health_at"`
	LastHealthSummary   string    `gorm:"column:last_health_summary;type:text;not null"`
	DefaultSymbolSource string    `gorm:"column:default_symbol_source;type:varchar(64);not null"`
	UpdatedBy           string    `gorm:"column:updated_by;type:varchar(128);not null"`
	CreatedAt           time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt           time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (TaLibRuntimeConfig) TableName() string {
	return "admin_ta_lib_runtime_configs"
}

type TaLibRuntimeConfigStore struct {
	db *gorm.DB
}

func NewTaLibRuntimeConfigStore(db *gorm.DB) *TaLibRuntimeConfigStore {
	return &TaLibRuntimeConfigStore{db: db}
}

func (s *TaLibRuntimeConfigStore) initTables() error {
	return s.db.AutoMigrate(&TaLibRuntimeConfig{})
}

// InitTables 对外暴露 ta-lib runtime 配置表初始化，兼容迁移期旧调用方。
func (s *TaLibRuntimeConfigStore) InitTables() error {
	return s.initTables()
}

func (s *TaLibRuntimeConfigStore) GetLatest() (*TaLibRuntimeConfig, error) {
	var item TaLibRuntimeConfig
	if err := s.db.Order("updated_at desc").Take(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *TaLibRuntimeConfigStore) Upsert(item *TaLibRuntimeConfig) error {
	if item == nil {
		return fmt.Errorf("ta-lib runtime config is required")
	}
	item.BaseURL = strings.TrimSpace(item.BaseURL)
	item.HealthStatus = strings.TrimSpace(item.HealthStatus)
	item.LastHealthSummary = strings.TrimSpace(item.LastHealthSummary)
	item.DefaultSymbolSource = strings.TrimSpace(item.DefaultSymbolSource)
	item.UpdatedBy = strings.TrimSpace(item.UpdatedBy)
	if item.UpdatedBy == "" {
		return fmt.Errorf("updated by is required")
	}
	if strings.TrimSpace(item.ID) == "" {
		item.ID = uuid.NewString()
	}
	existing, err := s.GetLatest()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			if item.HealthStatus == "" {
				item.HealthStatus = "unknown"
			}
			if item.DefaultSymbolSource == "" {
				item.DefaultSymbolSource = "market"
			}
			return s.db.Create(item).Error
		}
		return err
	}
	if item.BaseURL == "" {
		item.BaseURL = existing.BaseURL
	}
	existing.BaseURL = item.BaseURL
	existing.Enabled = item.Enabled
	existing.TimeoutMS = item.TimeoutMS
	existing.HealthStatus = item.HealthStatus
	existing.LastHealthAt = item.LastHealthAt
	existing.LastHealthSummary = item.LastHealthSummary
	existing.DefaultSymbolSource = item.DefaultSymbolSource
	existing.UpdatedBy = item.UpdatedBy
	return s.db.Save(existing).Error
}

type TaLibFunctionGovernance struct {
	ID                string    `gorm:"column:id;type:varchar(64);primaryKey"`
	FunctionKey       string    `gorm:"column:function_key;type:varchar(128);uniqueIndex;not null"`
	GroupName         string    `gorm:"column:group_name;type:varchar(128);not null"`
	DisplayName       string    `gorm:"column:display_name;type:varchar(128);not null"`
	DescriptionZH     string    `gorm:"column:description_zh;type:text;not null"`
	Enabled           bool      `gorm:"column:enabled;not null"`
	AgentEnabled      bool      `gorm:"column:agent_enabled;not null"`
	AdminTrialEnabled bool      `gorm:"column:admin_trial_enabled;not null"`
	DefaultParamsJSON string    `gorm:"column:default_params_json;type:text;not null"`
	RiskNote          string    `gorm:"column:risk_note;type:text;not null"`
	ReviewedBy        string    `gorm:"column:reviewed_by;type:varchar(128);not null"`
	ReviewedAt        time.Time `gorm:"column:reviewed_at"`
	CreatedAt         time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt         time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (TaLibFunctionGovernance) TableName() string {
	return "admin_ta_lib_function_governances"
}

type TaLibFunctionGovernanceStore struct {
	db *gorm.DB
}

func NewTaLibFunctionGovernanceStore(db *gorm.DB) *TaLibFunctionGovernanceStore {
	return &TaLibFunctionGovernanceStore{db: db}
}

func (s *TaLibFunctionGovernanceStore) initTables() error {
	return s.db.AutoMigrate(&TaLibFunctionGovernance{})
}

// InitTables 对外暴露函数治理表初始化，兼容迁移期旧调用方。
func (s *TaLibFunctionGovernanceStore) InitTables() error {
	return s.initTables()
}

func (s *TaLibFunctionGovernanceStore) ListAll() ([]TaLibFunctionGovernance, error) {
	items := make([]TaLibFunctionGovernance, 0, 16)
	if err := s.db.Order("function_key asc").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *TaLibFunctionGovernanceStore) Upsert(item *TaLibFunctionGovernance) error {
	if item == nil {
		return fmt.Errorf("ta-lib function governance is required")
	}
	item.FunctionKey = strings.TrimSpace(item.FunctionKey)
	item.GroupName = strings.TrimSpace(item.GroupName)
	item.DisplayName = strings.TrimSpace(item.DisplayName)
	item.DescriptionZH = strings.TrimSpace(item.DescriptionZH)
	item.DefaultParamsJSON = strings.TrimSpace(item.DefaultParamsJSON)
	item.RiskNote = strings.TrimSpace(item.RiskNote)
	item.ReviewedBy = strings.TrimSpace(item.ReviewedBy)
	if item.FunctionKey == "" {
		return fmt.Errorf("function key is required")
	}
	if item.DisplayName == "" {
		return fmt.Errorf("display name is required")
	}
	if item.ReviewedBy == "" {
		return fmt.Errorf("reviewed by is required")
	}
	if strings.TrimSpace(item.ID) == "" {
		item.ID = uuid.NewString()
	}
	var existing TaLibFunctionGovernance
	err := s.db.Where("function_key = ?", item.FunctionKey).Take(&existing).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			if item.ReviewedAt.IsZero() {
				item.ReviewedAt = time.Now().UTC()
			}
			return s.db.Create(item).Error
		}
		return err
	}
	existing.GroupName = item.GroupName
	existing.DisplayName = item.DisplayName
	existing.DescriptionZH = item.DescriptionZH
	existing.Enabled = item.Enabled
	existing.AgentEnabled = item.AgentEnabled
	existing.AdminTrialEnabled = item.AdminTrialEnabled
	existing.DefaultParamsJSON = item.DefaultParamsJSON
	existing.RiskNote = item.RiskNote
	existing.ReviewedBy = item.ReviewedBy
	if item.ReviewedAt.IsZero() {
		existing.ReviewedAt = time.Now().UTC()
	} else {
		existing.ReviewedAt = item.ReviewedAt
	}
	return s.db.Save(&existing).Error
}

type TaLibTrialHistory struct {
	ID                string    `gorm:"column:id;type:varchar(64);primaryKey"`
	Symbol            string    `gorm:"column:symbol;type:varchar(128);not null"`
	Timeframe         string    `gorm:"column:timeframe;type:varchar(64);not null"`
	FunctionKey       string    `gorm:"column:function_key;type:varchar(128);not null"`
	InputSummaryJSON  string    `gorm:"column:input_summary_json;type:text;not null"`
	ResultSummaryJSON string    `gorm:"column:result_summary_json;type:text;not null"`
	Status            string    `gorm:"column:status;type:varchar(64);not null"`
	ErrorMessage      string    `gorm:"column:error_message;type:text;not null"`
	CreatedBy         string    `gorm:"column:created_by;type:varchar(128);not null"`
	CreatedAt         time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (TaLibTrialHistory) TableName() string {
	return "admin_ta_lib_trial_histories"
}

type TaLibTrialHistoryStore struct {
	db *gorm.DB
}

func NewTaLibTrialHistoryStore(db *gorm.DB) *TaLibTrialHistoryStore {
	return &TaLibTrialHistoryStore{db: db}
}

func (s *TaLibTrialHistoryStore) initTables() error {
	return s.db.AutoMigrate(&TaLibTrialHistory{})
}

// InitTables 对外暴露试算历史表初始化，兼容迁移期旧调用方。
func (s *TaLibTrialHistoryStore) InitTables() error {
	return s.initTables()
}

func (s *TaLibTrialHistoryStore) Create(item *TaLibTrialHistory) error {
	if item == nil {
		return fmt.Errorf("ta-lib trial history is required")
	}
	item.Symbol = strings.TrimSpace(item.Symbol)
	item.Timeframe = strings.TrimSpace(item.Timeframe)
	item.FunctionKey = strings.TrimSpace(item.FunctionKey)
	item.InputSummaryJSON = strings.TrimSpace(item.InputSummaryJSON)
	item.ResultSummaryJSON = strings.TrimSpace(item.ResultSummaryJSON)
	item.Status = strings.TrimSpace(item.Status)
	item.ErrorMessage = strings.TrimSpace(item.ErrorMessage)
	item.CreatedBy = strings.TrimSpace(item.CreatedBy)
	if item.Symbol == "" || item.FunctionKey == "" {
		return fmt.Errorf("symbol and function key are required")
	}
	if strings.TrimSpace(item.ID) == "" {
		item.ID = uuid.NewString()
	}
	return s.db.Create(item).Error
}

func (s *TaLibTrialHistoryStore) ListRecent(limit int) ([]TaLibTrialHistory, error) {
	if limit <= 0 {
		limit = 20
	}
	items := make([]TaLibTrialHistory, 0, limit)
	if err := s.db.Order("created_at desc").Limit(limit).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
