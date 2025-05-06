package slienceconfig

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

type AlertSlienceConfig struct {
	ID int `gorm:"primryKey;auto_increment" json:"id"`

	AlertID   string    `gorm:"column:alert_id" json:"alertId"`
	AlertName string    `gorm:"alert_name" json:"alertName"`
	Group     string    `gorm:"group" json:"group"`
	Tags      TagsStr   `gorm:"tags" json:"tags"`
	For       string    `gorm:"for" json:"for"`
	StartAt   time.Time `gorm:"start_at" json:"startAt"`
	EndAt     time.Time `gorm:"end_at" json:"endAt"`
}

func (s AlertSlienceConfig) TableName() string {
	return "alert_slients"
}

func (s *AlertSlienceConfig) IsSlient(alert *alert.AlertEvent) bool {
	if s.AlertID == alert.AlertID &&
		s.StartAt.Before(alert.UpdateTime) &&
		s.EndAt.After(alert.UpdateTime) {
		return true
	}
	return false
}

type TagsStr map[string]string

func (e TagsStr) Value() (driver.Value, error) {
	if e == nil {
		return "", nil
	}
	val, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	return string(val), nil
}

func (e *TagsStr) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	val, ok := value.(string)
	if !ok {
		return fmt.Errorf("failed to scan JSONField, expected string, got %T", value)
	}
	if len(val) == 0 {
		return nil
	}
	return json.Unmarshal([]byte(val), e)
}
