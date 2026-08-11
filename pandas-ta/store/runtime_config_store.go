package store

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PandasTaRuntimeConfig 保存 pandas-ta runtime 的接入配置和最近健康态。
type PandasTaRuntimeConfig struct {
	ID                 string    `gorm:"column:id;type:varchar(64);primaryKey"`
	BaseURL            string    `gorm:"column:base_url;type:text;not null"`
	Enabled            bool      `gorm:"column:enabled;not null"`
	TimeoutMS          int       `gorm:"column:timeout_ms;not null"`
	HealthStatus       string    `gorm:"column:health_status;type:varchar(64);not null"`
	LastHealthAt       time.Time `gorm:"column:last_health_at"`
	LastHealthSummary  string    `gorm:"column:last_health_summary;type:text;not null"`
	DefaultPriceSource string    `gorm:"column:default_price_source;type:varchar(64);not null"`
	UpdatedBy          string    `gorm:"column:updated_by;type:varchar(128);not null"`
	CreatedAt          time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt          time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (PandasTaRuntimeConfig) TableName() string {
	return "admin_pandas_ta_runtime_configs"
}

type PandasTaRuntimeConfigStore struct {
	db *gorm.DB
}

func NewPandasTaRuntimeConfigStore(db *gorm.DB) *PandasTaRuntimeConfigStore {
	return &PandasTaRuntimeConfigStore{db: db}
}

func (s *PandasTaRuntimeConfigStore) InitTables() error {
	return s.db.AutoMigrate(&PandasTaRuntimeConfig{})
}

func (s *PandasTaRuntimeConfigStore) GetLatest() (*PandasTaRuntimeConfig, error) {
	var item PandasTaRuntimeConfig
	if err := s.db.Order("updated_at desc").Take(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *PandasTaRuntimeConfigStore) Upsert(item *PandasTaRuntimeConfig) error {
	if item == nil {
		return fmt.Errorf("pandas-ta runtime config is required")
	}
	item.BaseURL = strings.TrimSpace(item.BaseURL)
	item.HealthStatus = strings.TrimSpace(item.HealthStatus)
	item.LastHealthSummary = strings.TrimSpace(item.LastHealthSummary)
	item.DefaultPriceSource = strings.TrimSpace(item.DefaultPriceSource)
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
			if item.DefaultPriceSource == "" {
				item.DefaultPriceSource = "market"
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
	existing.DefaultPriceSource = item.DefaultPriceSource
	existing.UpdatedBy = item.UpdatedBy
	return s.db.Save(existing).Error
}
