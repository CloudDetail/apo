// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"strings"
	"testing"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	dbdriver "github.com/CloudDetail/apo/backend/pkg/repository/database/driver"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func TestListSystemHealthSnapshotsRequiresGroupIDs(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&SystemHealthSystemSnapshot{}, &SystemHealthServiceSnapshot{}, &SystemHealthSnapshotLock{}); err != nil {
		t.Fatalf("migrate snapshots: %v", err)
	}
	repo := &daoRepo{DB: &dbdriver.DB{DB: db}}
	start := time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)

	if err := repo.SaveSystemHealthSnapshots(
		core.EmptyCtx(),
		[]SystemHealthSystemSnapshot{{BucketStart: start, BucketEnd: end, GroupID: 101, GroupName: "订单系统", Score: 90}},
		[]SystemHealthServiceSnapshot{{BucketStart: start, BucketEnd: end, GroupID: 101, GroupName: "订单系统", ServiceName: "order-api", Score: 90}},
	); err != nil {
		t.Fatalf("save snapshots: %v", err)
	}

	systemSnapshots, err := repo.ListSystemHealthSystemSnapshots(core.EmptyCtx(), nil, start, end)
	if err != nil {
		t.Fatalf("list system snapshots: %v", err)
	}
	if len(systemSnapshots) != 0 {
		t.Fatalf("expected no system snapshots without visible group ids, got %d", len(systemSnapshots))
	}

	serviceSnapshots, err := repo.ListSystemHealthServiceSnapshots(core.EmptyCtx(), nil, start, end)
	if err != nil {
		t.Fatalf("list service snapshots: %v", err)
	}
	if len(serviceSnapshots) != 0 {
		t.Fatalf("expected no service snapshots without visible group ids, got %d", len(serviceSnapshots))
	}
}

func TestLatestSystemHealthSnapshotBucketStart(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&SystemHealthSystemSnapshot{}, &SystemHealthServiceSnapshot{}, &SystemHealthSnapshotLock{}); err != nil {
		t.Fatalf("migrate snapshots: %v", err)
	}
	repo := &daoRepo{DB: &dbdriver.DB{DB: db}}
	start := time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC)

	_, ok, err := repo.LatestSystemHealthSnapshotBucketStart(core.EmptyCtx())
	if err != nil {
		t.Fatalf("latest empty snapshots: %v", err)
	}
	if ok {
		t.Fatal("expected no latest snapshot before saving data")
	}

	if err := repo.SaveSystemHealthSnapshots(
		core.EmptyCtx(),
		[]SystemHealthSystemSnapshot{
			{BucketStart: start, BucketEnd: start.Add(time.Hour), GroupID: 101, GroupName: "订单系统", Score: 90},
			{BucketStart: start.Add(2 * time.Hour), BucketEnd: start.Add(3 * time.Hour), GroupID: 101, GroupName: "订单系统", Score: 80},
		},
		nil,
	); err != nil {
		t.Fatalf("save snapshots: %v", err)
	}

	latest, ok, err := repo.LatestSystemHealthSnapshotBucketStart(core.EmptyCtx())
	if err != nil {
		t.Fatalf("latest snapshots: %v", err)
	}
	if !ok {
		t.Fatal("expected latest snapshot after saving data")
	}
	if !latest.Equal(start.Add(2 * time.Hour)) {
		t.Fatalf("expected latest bucket %s, got %s", start.Add(2*time.Hour), latest)
	}
}

func TestTryClaimSystemHealthSnapshotBucketHonorsLockTTL(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&SystemHealthSystemSnapshot{}, &SystemHealthServiceSnapshot{}, &SystemHealthSnapshotLock{}); err != nil {
		t.Fatalf("migrate snapshots: %v", err)
	}
	repo := &daoRepo{DB: &dbdriver.DB{DB: db}}
	bucket := time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC)

	claimed, err := repo.TryClaimSystemHealthSnapshotBucket(core.EmptyCtx(), bucket, time.Hour)
	if err != nil {
		t.Fatalf("first claim: %v", err)
	}
	if !claimed {
		t.Fatal("expected first claim to acquire the bucket")
	}

	claimed, err = repo.TryClaimSystemHealthSnapshotBucket(core.EmptyCtx(), bucket, time.Hour)
	if err != nil {
		t.Fatalf("second claim: %v", err)
	}
	if claimed {
		t.Fatal("expected unexpired lock to reject a second claim")
	}

	if err := db.Model(&SystemHealthSnapshotLock{}).
		Where("bucket_start = ?", bucket).
		Update("locked_until", time.Now().Add(-time.Minute)).Error; err != nil {
		t.Fatalf("expire lock: %v", err)
	}

	claimed, err = repo.TryClaimSystemHealthSnapshotBucket(core.EmptyCtx(), bucket, time.Hour)
	if err != nil {
		t.Fatalf("expired claim: %v", err)
	}
	if !claimed {
		t.Fatal("expected expired lock to be claimable")
	}
}

func TestSystemHealthSnapshotLockExpiredExprQualifiesColumnForPostgres(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: "user=apo dbname=apo sslmode=disable",
	}), &gorm.Config{DryRun: true, DisableAutomaticPing: true, SkipDefaultTransaction: true})
	if err != nil {
		t.Fatalf("open postgres dry run: %v", err)
	}
	now := time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC)
	lock := SystemHealthSnapshotLock{
		BucketStart: now,
		LockedUntil: now.Add(time.Hour),
		CreatedAt:   now,
	}

	result := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "bucket_start"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"locked_until": now.Add(time.Hour),
			"updated_at":   now,
		}),
		Where: clause.Where{Exprs: []clause.Expression{
			systemHealthSnapshotLockExpiredExpr(now),
		}},
	}).Create(&lock)
	if result.Error != nil {
		t.Fatalf("build postgres dry-run lock SQL: %v", result.Error)
	}
	stmt := result.Statement
	sql := stmt.SQL.String()

	if !strings.Contains(sql, `"system_health_snapshot_locks"."locked_until" < `) {
		t.Fatalf("expected qualified locked_until in postgres lock SQL, got %s", sql)
	}
	if strings.Contains(sql, " WHERE locked_until < ") {
		t.Fatalf("expected no ambiguous unqualified locked_until in postgres lock SQL, got %s", sql)
	}
}
