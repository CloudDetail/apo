// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"errors"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SystemHealthSystemSnapshot struct {
	ID                       uint      `gorm:"primaryKey;autoIncrement"`
	BucketStart              time.Time `gorm:"column:bucket_start;uniqueIndex:idx_system_health_system_snapshots_bucket_group;index:idx_system_health_system_snapshots_bucket"`
	BucketEnd                time.Time `gorm:"column:bucket_end;index:idx_system_health_system_snapshots_bucket_end"`
	Granularity              string    `gorm:"column:granularity;type:varchar(16);not null;default:hour"`
	GroupID                  int64     `gorm:"column:group_id;uniqueIndex:idx_system_health_system_snapshots_bucket_group;index:idx_system_health_system_snapshots_group_bucket"`
	GroupName                string    `gorm:"column:group_name;type:varchar(255)"`
	Score                    int       `gorm:"column:score"`
	SnapshotStatus           string    `gorm:"column:snapshot_status;type:varchar(16)"`
	SnapshotSevereThreshold  int       `gorm:"column:snapshot_severe_threshold"`
	SnapshotGeneralThreshold int       `gorm:"column:snapshot_general_threshold"`
	ServiceTotal             int       `gorm:"column:service_total"`
	CalcVersion              string    `gorm:"column:calc_version;type:varchar(32)"`
	CreatedAt                time.Time `gorm:"column:created_at"`
	UpdatedAt                time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (SystemHealthSystemSnapshot) TableName() string {
	return "system_health_system_snapshots"
}

type SystemHealthServiceSnapshot struct {
	ID                       uint      `gorm:"primaryKey;autoIncrement"`
	BucketStart              time.Time `gorm:"column:bucket_start;uniqueIndex:idx_system_health_service_snapshots_bucket_group_service;index:idx_system_health_service_snapshots_bucket"`
	BucketEnd                time.Time `gorm:"column:bucket_end;index:idx_system_health_service_snapshots_bucket_end"`
	Granularity              string    `gorm:"column:granularity;type:varchar(16);not null;default:hour"`
	GroupID                  int64     `gorm:"column:group_id;uniqueIndex:idx_system_health_service_snapshots_bucket_group_service;index:idx_system_health_service_snapshots_group_bucket"`
	GroupName                string    `gorm:"column:group_name;type:varchar(255)"`
	ServiceName              string    `gorm:"column:service_name;type:varchar(512);uniqueIndex:idx_system_health_service_snapshots_bucket_group_service"`
	Score                    int       `gorm:"column:score"`
	SnapshotStatus           string    `gorm:"column:snapshot_status;type:varchar(16)"`
	SnapshotSevereThreshold  int       `gorm:"column:snapshot_severe_threshold"`
	SnapshotGeneralThreshold int       `gorm:"column:snapshot_general_threshold"`
	CalcVersion              string    `gorm:"column:calc_version;type:varchar(32)"`
	CreatedAt                time.Time `gorm:"column:created_at"`
	UpdatedAt                time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (SystemHealthServiceSnapshot) TableName() string {
	return "system_health_service_snapshots"
}

type SystemHealthSnapshotLock struct {
	BucketStart time.Time `gorm:"column:bucket_start;primaryKey"`
	LockedUntil time.Time `gorm:"column:locked_until;index"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (SystemHealthSnapshotLock) TableName() string {
	return "system_health_snapshot_locks"
}

func (repo *daoRepo) SaveSystemHealthSnapshots(ctx core.Context, systems []SystemHealthSystemSnapshot, services []SystemHealthServiceSnapshot) error {
	now := time.Now()
	for index := range systems {
		if systems[index].Granularity == "" {
			systems[index].Granularity = "hour"
		}
		if systems[index].CreatedAt.IsZero() {
			systems[index].CreatedAt = now
		}
	}
	for index := range services {
		if services[index].Granularity == "" {
			services[index].Granularity = "hour"
		}
		if services[index].CreatedAt.IsZero() {
			services[index].CreatedAt = now
		}
	}

	return repo.GetContextDB(ctx).Transaction(func(tx *gorm.DB) error {
		if len(systems) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "bucket_start"}, {Name: "group_id"}},
				DoUpdates: clause.AssignmentColumns([]string{
					"bucket_end",
					"granularity",
					"group_name",
					"score",
					"snapshot_status",
					"snapshot_severe_threshold",
					"snapshot_general_threshold",
					"service_total",
					"calc_version",
					"updated_at",
				}),
			}).Create(&systems).Error; err != nil {
				return err
			}
		}
		if len(services) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "bucket_start"}, {Name: "group_id"}, {Name: "service_name"}},
				DoUpdates: clause.AssignmentColumns([]string{
					"bucket_end",
					"granularity",
					"group_name",
					"score",
					"snapshot_status",
					"snapshot_severe_threshold",
					"snapshot_general_threshold",
					"calc_version",
					"updated_at",
				}),
			}).Create(&services).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (repo *daoRepo) LatestSystemHealthSnapshotBucketStart(ctx core.Context) (time.Time, bool, error) {
	var snapshot SystemHealthSystemSnapshot
	err := repo.GetContextDB(ctx).
		Order("bucket_start DESC").
		Limit(1).
		First(&snapshot).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return time.Time{}, false, nil
	}
	if err != nil {
		return time.Time{}, false, err
	}
	return snapshot.BucketStart, true, nil
}

func (repo *daoRepo) TryClaimSystemHealthSnapshotBucket(ctx core.Context, bucketStart time.Time, ttl time.Duration) (bool, error) {
	if ttl <= 0 {
		ttl = time.Hour
	}
	now := time.Now()
	lock := SystemHealthSnapshotLock{
		BucketStart: bucketStart,
		LockedUntil: now.Add(ttl),
		CreatedAt:   now,
	}
	result := repo.GetContextDB(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "bucket_start"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"locked_until": now.Add(ttl),
			"updated_at":   now,
		}),
		Where: clause.Where{Exprs: []clause.Expression{
			systemHealthSnapshotLockExpiredExpr(now),
		}},
	}).Create(&lock)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func systemHealthSnapshotLockExpiredExpr(now time.Time) clause.Expression {
	return clause.Expr{
		SQL:  `"system_health_snapshot_locks"."locked_until" < ?`,
		Vars: []interface{}{now},
	}
}

func (repo *daoRepo) ListSystemHealthSystemSnapshots(ctx core.Context, groupIDs []int64, startTime time.Time, endTime time.Time) ([]SystemHealthSystemSnapshot, error) {
	if len(groupIDs) == 0 {
		return []SystemHealthSystemSnapshot{}, nil
	}
	query := repo.GetContextDB(ctx).
		Where("bucket_start >= ? AND bucket_start < ?", startTime, endTime).
		Where("group_id IN ?", groupIDs).
		Order("bucket_start ASC, group_id ASC")
	var snapshots []SystemHealthSystemSnapshot
	err := query.Find(&snapshots).Error
	return snapshots, err
}

func (repo *daoRepo) ListSystemHealthServiceSnapshots(ctx core.Context, groupIDs []int64, startTime time.Time, endTime time.Time) ([]SystemHealthServiceSnapshot, error) {
	if len(groupIDs) == 0 {
		return []SystemHealthServiceSnapshot{}, nil
	}
	query := repo.GetContextDB(ctx).
		Where("bucket_start >= ? AND bucket_start < ?", startTime, endTime).
		Where("group_id IN ?", groupIDs).
		Order("bucket_start ASC, group_id ASC, service_name ASC")
	var snapshots []SystemHealthServiceSnapshot
	err := query.Find(&snapshots).Error
	return snapshots, err
}
