// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"fmt"
	"testing"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	dbdriver "github.com/CloudDetail/apo/backend/pkg/repository/database/driver"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSystemHealthReportRecordsRetentionPruneAndFilter(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&SystemHealthReportRecord{}); err != nil {
		t.Fatalf("migrate report records: %v", err)
	}
	repo := &daoRepo{DB: &dbdriver.DB{DB: db}}
	ctx := core.EmptyCtx()
	base := time.Now().UTC().Truncate(time.Second)

	first := &SystemHealthReportRecord{
		RecordID:  "first-record",
		GroupID:   101,
		Payload:   `{"version":1}`,
		CreatedAt: base,
	}
	if err := repo.SaveSystemHealthReportRecord(ctx, first); err != nil {
		t.Fatalf("save first report: %v", err)
	}
	second := &SystemHealthReportRecord{
		RecordID:  "second-record",
		GroupID:   101,
		Payload:   `{"version":2}`,
		CreatedAt: base.Add(time.Second),
	}
	if err := repo.SaveSystemHealthReportRecord(ctx, second); err != nil {
		t.Fatalf("save second report: %v", err)
	}

	var count int64
	if err := db.Model(&SystemHealthReportRecord{}).Where("group_id = ?", 101).Count(&count).Error; err != nil {
		t.Fatalf("count generated reports: %v", err)
	}
	if count != 2 {
		t.Fatalf("each generated report must remain an independent record, got %d", count)
	}

	expired := &SystemHealthReportRecord{
		RecordID:  "expired-record",
		GroupID:   101,
		Payload:   `{"expired":true}`,
		CreatedAt: base.Add(-SystemHealthReportRecordRetention - time.Hour),
	}
	if err := db.Create(expired).Error; err != nil {
		t.Fatalf("seed expired report: %v", err)
	}

	items := make([]SystemHealthReportRecord, 0, MaxSystemHealthReportRecordsPerGroup)
	for i := 0; i < MaxSystemHealthReportRecordsPerGroup; i++ {
		items = append(items, SystemHealthReportRecord{
			RecordID:  fmt.Sprintf("record-%04d", i),
			GroupID:   101,
			Payload:   fmt.Sprintf(`{"index":%d}`, i),
			CreatedAt: base.Add(time.Duration(i+2) * time.Minute),
		})
	}
	if err := db.CreateInBatches(items, 100).Error; err != nil {
		t.Fatalf("seed report history: %v", err)
	}
	newest := &SystemHealthReportRecord{
		RecordID:  "record-0500",
		GroupID:   101,
		Payload:   `{"index":500}`,
		CreatedAt: base.Add(time.Duration(MaxSystemHealthReportRecordsPerGroup+2) * time.Minute),
	}
	if err := repo.SaveSystemHealthReportRecord(ctx, newest); err != nil {
		t.Fatalf("save newest report: %v", err)
	}
	if err := db.Model(&SystemHealthReportRecord{}).Where("record_id = ?", expired.RecordID).Count(&count).Error; err != nil {
		t.Fatalf("count expired reports: %v", err)
	}
	if count != 0 {
		t.Fatalf("expired report should be deleted on the next group save, got %d", count)
	}
	if err := db.Model(&SystemHealthReportRecord{}).Where("group_id = ?", 101).Count(&count).Error; err != nil {
		t.Fatalf("count pruned reports: %v", err)
	}
	if count != MaxSystemHealthReportRecordsPerGroup {
		t.Fatalf("expected %d retained reports, got %d", MaxSystemHealthReportRecordsPerGroup, count)
	}
	if _, err := repo.GetSystemHealthReportRecord(ctx, newest.RecordID); err != nil {
		t.Fatalf("newest report must be retained: %v", err)
	}

	if err := db.Create(&SystemHealthReportRecord{
		RecordID:  "expired-other-group",
		GroupID:   202,
		Payload:   `{}`,
		CreatedAt: base.Add(-SystemHealthReportRecordRetention - time.Hour),
	}).Error; err != nil {
		t.Fatalf("seed expired other-group report: %v", err)
	}
	if _, err := repo.GetSystemHealthReportRecord(ctx, "expired-other-group"); err == nil {
		t.Fatal("expired report detail must not be readable")
	}
	records, total, err := repo.ListSystemHealthReportRecords(ctx, []int64{202}, 1, 20)
	if err != nil {
		t.Fatalf("list expired group reports: %v", err)
	}
	if total != 0 || len(records) != 0 {
		t.Fatalf("expired reports must be excluded from history: total=%d records=%+v", total, records)
	}
	if err := repo.SaveSystemHealthReportRecord(ctx, &SystemHealthReportRecord{
		RecordID:  "other-group",
		GroupID:   202,
		Payload:   `{}`,
		CreatedAt: base,
	}); err != nil {
		t.Fatalf("save other group report: %v", err)
	}
	records, total, err = repo.ListSystemHealthReportRecords(ctx, []int64{202}, 1, 20)
	if err != nil {
		t.Fatalf("list authorized group reports: %v", err)
	}
	if total != 1 || len(records) != 1 || records[0].GroupID != 202 {
		t.Fatalf("unexpected authorized report list: total=%d records=%+v", total, records)
	}
	records, total, err = repo.ListSystemHealthReportRecords(ctx, nil, 1, 20)
	if err != nil {
		t.Fatalf("list empty authorized groups: %v", err)
	}
	if total != 0 || len(records) != 0 {
		t.Fatalf("empty authorization must return no reports, total=%d records=%+v", total, records)
	}
}
