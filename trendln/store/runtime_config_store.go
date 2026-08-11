package store

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TrendlnRuntimeConfig 保存 trendln runtime 的接入配置与最近健康态。
type TrendlnRuntimeConfig struct {
	ID                string    `gorm:"column:id;type:varchar(64);primaryKey"`
	BaseURL           string    `gorm:"column:base_url;type:text;not null"`
	Enabled           bool      `gorm:"column:enabled;not null"`
	TimeoutMS         int       `gorm:"column:timeout_ms;not null"`
	HealthStatus      string    `gorm:"column:health_status;type:varchar(64);not null"`
	LastHealthAt      time.Time `gorm:"column:last_health_at"`
	LastHealthSummary string    `gorm:"column:last_health_summary;type:text;not null"`
	DefaultPriceSource string   `gorm:"column:default_price_source;type:varchar(64);not null"`
	UpdatedBy         string    `gorm:"column:updated_by;type:varchar(128);not null"`
	CreatedAt         time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt         time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (TrendlnRuntimeConfig) TableName() string {
	return "admin_trendln_runtime_configs"
}

type TrendlnRuntimeConfigStore struct {
	db *gorm.DB
}

func NewTrendlnRuntimeConfigStore(db *gorm.DB) *TrendlnRuntimeConfigStore {
	return &TrendlnRuntimeConfigStore{db: db}
}

func (s *TrendlnRuntimeConfigStore) InitTables() error {
	return s.db.AutoMigrate(&TrendlnRuntimeConfig{})
}

func (s *TrendlnRuntimeConfigStore) GetLatest() (*TrendlnRuntimeConfig, error) {
	var item TrendlnRuntimeConfig
	if err := s.db.Order("updated_at desc").Take(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *TrendlnRuntimeConfigStore) Upsert(item *TrendlnRuntimeConfig) error {
	if item == nil {
		return fmt.Errorf("trendln runtime config is required")
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
