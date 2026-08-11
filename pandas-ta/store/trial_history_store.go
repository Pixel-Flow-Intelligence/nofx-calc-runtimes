package store

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PandasTaTrialHistory 记录每次试算的摘要，便于 admin 追溯。
type PandasTaTrialHistory struct {
	ID               string    `gorm:"column:id;type:varchar(64);primaryKey"`
	Symbol           string    `gorm:"column:symbol;type:varchar(128);not null"`
	Timeframe        string    `gorm:"column:timeframe;type:varchar(64);not null"`
	FunctionKey      string    `gorm:"column:function_key;type:varchar(128);not null"`
	InputSummaryJSON string    `gorm:"column:input_summary_json;type:text;not null"`
	ResultSummaryJSON string   `gorm:"column:result_summary_json;type:text;not null"`
	Status           string    `gorm:"column:status;type:varchar(64);not null"`
	ErrorMessage     string    `gorm:"column:error_message;type:text;not null"`
	CreatedBy        string    `gorm:"column:created_by;type:varchar(128);not null"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (PandasTaTrialHistory) TableName() string {
	return "admin_pandas_ta_trial_histories"
}

type PandasTaTrialHistoryStore struct {
	db *gorm.DB
}

func NewPandasTaTrialHistoryStore(db *gorm.DB) *PandasTaTrialHistoryStore {
	return &PandasTaTrialHistoryStore{db: db}
}

func (s *PandasTaTrialHistoryStore) InitTables() error {
	return s.db.AutoMigrate(&PandasTaTrialHistory{})
}

func (s *PandasTaTrialHistoryStore) Create(item *PandasTaTrialHistory) error {
	if item == nil {
		return fmt.Errorf("pandas-ta trial history is required")
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

func (s *PandasTaTrialHistoryStore) ListRecent(limit int) ([]PandasTaTrialHistory, error) {
	if limit <= 0 {
		limit = 20
	}
	items := make([]PandasTaTrialHistory, 0, limit)
	if err := s.db.Order("created_at desc").Limit(limit).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
