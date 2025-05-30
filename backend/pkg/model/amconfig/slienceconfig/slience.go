// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package slienceconfig

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
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
	var val []byte
	switch s := value.(type) {
	case string:
		val = []byte(s)
	case []byte:
		val = s
	default:
		return fmt.Errorf("failed to scan JSONField, expected string or []byte, got %T", value)
	}
	return json.Unmarshal(val, e)
}
