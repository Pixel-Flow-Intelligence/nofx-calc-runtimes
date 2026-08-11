package store

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TrendlnFunctionGovernance 控制结构分析能力是否对 admin/agent 可用。
type TrendlnFunctionGovernance struct {
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

func (TrendlnFunctionGovernance) TableName() string {
	return "admin_trendln_function_governances"
}

type TrendlnFunctionGovernanceStore struct {
	db *gorm.DB
}

func NewTrendlnFunctionGovernanceStore(db *gorm.DB) *TrendlnFunctionGovernanceStore {
	return &TrendlnFunctionGovernanceStore{db: db}
}

func (s *TrendlnFunctionGovernanceStore) InitTables() error {
	return s.db.AutoMigrate(&TrendlnFunctionGovernance{})
}

func (s *TrendlnFunctionGovernanceStore) ListAll() ([]TrendlnFunctionGovernance, error) {
	items := make([]TrendlnFunctionGovernance, 0, 16)
	if err := s.db.Order("function_key asc").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *TrendlnFunctionGovernanceStore) Upsert(item *TrendlnFunctionGovernance) error {
	if item == nil {
		return fmt.Errorf("trendln function governance is required")
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
	var existing TrendlnFunctionGovernance
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
