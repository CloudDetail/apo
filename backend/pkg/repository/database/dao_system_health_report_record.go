// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
)

type SystemHealthReportRecord struct {
	RecordID         string    `gorm:"column:record_id;type:varchar(64);primaryKey"`
	GroupID          int64     `gorm:"column:group_id;index:idx_system_health_report_records_group_created"`
	Status           string    `gorm:"column:status;type:varchar(32)"`
	Score            int       `gorm:"column:score"`
	Summary          string    `gorm:"column:summary;type:text"`
	Window           string    `gorm:"column:window;type:varchar(64)"`
	StartTime        int64     `gorm:"column:start_time;index"`
	EndTime          int64     `gorm:"column:end_time;index"`
	GeneratedAt      string    `gorm:"column:generated_at;type:varchar(64)"`
	IssueCount       int       `gorm:"column:issue_count"`
	ActiveAlertCount int       `gorm:"column:active_alert_count"`
	AIStatus         string    `gorm:"column:ai_status;type:varchar(32)"`
	Payload          string    `gorm:"column:payload"`
	CreatedAt        time.Time `gorm:"column:created_at;index:idx_system_health_report_records_group_created"`
}

func (SystemHealthReportRecord) TableName() string {
	return "system_health_report_records"
}

func (repo *daoRepo) SaveSystemHealthReportRecord(ctx core.Context, record *SystemHealthReportRecord) error {
	if record == nil {
		return nil
	}
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now()
	}
	return repo.GetContextDB(ctx).Create(record).Error
}

func (repo *daoRepo) ListSystemHealthReportRecords(ctx core.Context, groupID int64, page int, pageSize int) ([]SystemHealthReportRecord, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	query := repo.GetContextDB(ctx).Model(&SystemHealthReportRecord{})
	if groupID > 0 {
		query = query.Where("group_id = ?", groupID)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var records []SystemHealthReportRecord
	err := query.
		Select("record_id", "group_id", "status", "score", "summary", "window", "start_time", "end_time", "generated_at", "issue_count", "active_alert_count", "ai_status", "created_at").
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&records).Error
	return records, total, err
}

func (repo *daoRepo) GetSystemHealthReportRecord(ctx core.Context, recordID string) (*SystemHealthReportRecord, error) {
	var record SystemHealthReportRecord
	err := repo.GetContextDB(ctx).Where("record_id = ?", recordID).First(&record).Error
	return &record, err
}
