package slienceconfig

import (
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

type AlertSlienceConfig struct {
	ID int `gorm:"primryKey;auto_increment" json:"id"`

	AlertID string    `gorm:"column:alert_id" json:"alertId"`
	For     string    `gorm:"for" json:"for"`
	StartAt time.Time `gorm:"start_at" json:"startAt"`
	EndAt   time.Time `gorm:"end_at" json:"endAt"`
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
