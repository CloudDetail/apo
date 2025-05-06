package slienceconfig

import (
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

type AlertSlienceConfig struct {
	ID int `gorm:"primaryKey;auto_increment"`

	AlertID string    `gorm:"primaryKey;column:alert_id"`
	For     string    `gorm:"for"`
	StartAt time.Time `gorm:"start_at"`
	EndAt   time.Time `gorm:"end_at"`
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
