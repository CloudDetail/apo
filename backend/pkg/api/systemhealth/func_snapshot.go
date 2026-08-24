// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package systemhealth

import (
	"context"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/datagroup"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/services/common"
	"go.uber.org/zap"
)

const systemHealthSnapshotCalcVersion = "v1"

const (
	systemHealthSnapshotMaxBackfillBuckets = 24
	systemHealthSnapshotLockTTL            = 50 * time.Minute
)

type systemHealthSnapshotWindow struct {
	Start time.Time
	End   time.Time
}

func (h *handler) StartSnapshotScheduler(ctx context.Context, logger *zap.Logger) {
	if logger == nil {
		logger = zap.NewNop()
	}
	go h.runSystemHealthSnapshotScheduler(ctx, logger)
}

func (h *handler) runSystemHealthSnapshotScheduler(ctx context.Context, logger *zap.Logger) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		now := time.Now()
		if now.Minute() >= 10 {
			h.collectDueSystemHealthSnapshots(core.EmptyCtx(), now, logger)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func latestCompleteSystemHealthSnapshotWindow(now time.Time) (time.Time, time.Time) {
	end := now.Truncate(time.Hour)
	return end.Add(-time.Hour), end
}

func systemHealthSnapshotWindowsDue(latestSaved *time.Time, now time.Time, maxBackfillBuckets int) []systemHealthSnapshotWindow {
	if now.Minute() < 10 {
		return nil
	}
	if maxBackfillBuckets <= 0 {
		maxBackfillBuckets = 1
	}
	latestStart, _ := latestCompleteSystemHealthSnapshotWindow(now)
	firstStart := latestStart
	if latestSaved != nil && !latestSaved.IsZero() {
		firstStart = latestSaved.Add(time.Hour)
	}
	if firstStart.After(latestStart) {
		return nil
	}

	earliestStart := latestStart.Add(-time.Duration(maxBackfillBuckets-1) * time.Hour)
	if firstStart.Before(earliestStart) {
		firstStart = earliestStart
	}

	windows := make([]systemHealthSnapshotWindow, 0, int(latestStart.Sub(firstStart)/time.Hour)+1)
	for cursor := firstStart; !cursor.After(latestStart); cursor = cursor.Add(time.Hour) {
		windows = append(windows, systemHealthSnapshotWindow{Start: cursor, End: cursor.Add(time.Hour)})
	}
	return windows
}

func (h *handler) collectDueSystemHealthSnapshots(ctx core.Context, now time.Time, logger *zap.Logger) {
	if logger == nil {
		logger = zap.NewNop()
	}
	latestSaved, ok, err := h.dbRepo.LatestSystemHealthSnapshotBucketStart(ctx)
	if err != nil {
		logger.Warn("system-health snapshot latest bucket lookup failed", zap.Error(err))
		return
	}
	var latestSavedPtr *time.Time
	if ok {
		latestSavedPtr = &latestSaved
	}
	collectSystemHealthSnapshotWindows(
		ctx,
		systemHealthSnapshotWindowsDue(latestSavedPtr, now, systemHealthSnapshotMaxBackfillBuckets),
		logger,
		func(ctx core.Context, window systemHealthSnapshotWindow) (bool, error) {
			return h.dbRepo.TryClaimSystemHealthSnapshotBucket(ctx, window.Start, systemHealthSnapshotLockTTL)
		},
		func(ctx core.Context, window systemHealthSnapshotWindow) error {
			return h.collectSystemHealthSnapshot(ctx, window.Start, window.End, logger)
		},
	)
}

func collectSystemHealthSnapshotWindows(
	ctx core.Context,
	windows []systemHealthSnapshotWindow,
	logger *zap.Logger,
	claim func(core.Context, systemHealthSnapshotWindow) (bool, error),
	collect func(core.Context, systemHealthSnapshotWindow) error,
) {
	if logger == nil {
		logger = zap.NewNop()
	}
	for _, window := range windows {
		claimed, err := claim(ctx, window)
		if err != nil {
			logger.Warn("system-health snapshot lock claim failed",
				zap.Time("startTime", window.Start),
				zap.Time("endTime", window.End),
				zap.Error(err),
			)
			return
		}
		if !claimed {
			return
		}
		if err := collect(ctx, window); err != nil {
			return
		}
	}
}

func (h *handler) collectSystemHealthSnapshot(ctx core.Context, startTime time.Time, endTime time.Time, logger *zap.Logger) error {
	groups, err := h.snapshotSystemGroups(ctx)
	if err != nil {
		logger.Warn("system-health snapshot groups failed",
			zap.Time("startTime", startTime),
			zap.Time("endTime", endTime),
			zap.Error(err),
		)
		return err
	}
	systems, err := h.buildOverviewSystems(ctx, startTime, endTime, groups, "snapshot")
	if err != nil {
		logger.Warn("system-health snapshot overview build failed",
			zap.Time("startTime", startTime),
			zap.Time("endTime", endTime),
			zap.Error(err),
		)
		return err
	}
	thresholds, err := h.systemHealthConfig(ctx)
	if err != nil {
		logger.Warn("system-health snapshot config failed, using default threshold", zap.Error(err))
	}
	thresholds = normalizeSystemHealthScoreThresholds(thresholds)
	applySystemHealthScoreThresholds(systems, thresholds)

	systemSnapshots, serviceSnapshots := buildSystemHealthSnapshotRecords(systems, thresholds, startTime, endTime)
	if err := h.dbRepo.SaveSystemHealthSnapshots(ctx, systemSnapshots, serviceSnapshots); err != nil {
		logger.Warn("system-health snapshot save failed",
			zap.Time("startTime", startTime),
			zap.Time("endTime", endTime),
			zap.Int("systems", len(systemSnapshots)),
			zap.Int("services", len(serviceSnapshots)),
			zap.Error(err),
		)
		return err
	}
	logger.Info("system-health snapshot saved",
		zap.Time("startTime", startTime),
		zap.Time("endTime", endTime),
		zap.Int("systems", len(systemSnapshots)),
		zap.Int("services", len(serviceSnapshots)),
	)
	return nil
}

func (h *handler) snapshotSystemGroups(ctx core.Context) ([]*datagroup.DataGroupTreeNode, error) {
	root := common.DataGroupStorage
	if root == nil || root.DataGroupTreeNode == nil {
		rootNode, err := h.dbRepo.LoadDataGroupTree(ctx)
		if err != nil {
			return nil, err
		}
		return snapshotSystemGroupsFromRoot(rootNode), nil
	}
	return snapshotSystemGroupsFromRoot(root.DataGroupTreeNode), nil
}

func snapshotSystemGroupsFromRoot(root *datagroup.DataGroupTreeNode) []*datagroup.DataGroupTreeNode {
	return firstLevelSystemGroups(root)
}

func buildSystemHealthSnapshotRecords(
	systems []response.SystemHealthSystem,
	thresholds response.SystemHealthConfigResp,
	startTime time.Time,
	endTime time.Time,
) ([]database.SystemHealthSystemSnapshot, []database.SystemHealthServiceSnapshot) {
	thresholds = normalizeSystemHealthScoreThresholds(thresholds)
	systemSnapshots := make([]database.SystemHealthSystemSnapshot, 0, len(systems))
	serviceSnapshots := make([]database.SystemHealthServiceSnapshot, 0)
	for _, system := range systems {
		status := system.Status
		if status == "" {
			status = scoreHealthStatus(system.Score, thresholds)
		}
		if len(system.Services) == 0 {
			status = healthStatusGray
		}
		systemSnapshots = append(systemSnapshots, database.SystemHealthSystemSnapshot{
			BucketStart:              startTime,
			BucketEnd:                endTime,
			Granularity:              "hour",
			GroupID:                  system.GroupID,
			GroupName:                system.Name,
			Score:                    normalizeHealthScore(system.Score),
			SnapshotStatus:           status,
			SnapshotSevereThreshold:  thresholds.SevereAbnormalScoreThreshold,
			SnapshotGeneralThreshold: thresholds.GeneralAbnormalScoreThreshold,
			ServiceTotal:             len(system.Services),
			CalcVersion:              systemHealthSnapshotCalcVersion,
		})
		for _, service := range system.Services {
			serviceStatus := service.Status
			if serviceStatus == "" {
				serviceStatus = scoreHealthStatus(service.Score, thresholds)
			}
			serviceSnapshots = append(serviceSnapshots, database.SystemHealthServiceSnapshot{
				BucketStart:              startTime,
				BucketEnd:                endTime,
				Granularity:              "hour",
				GroupID:                  system.GroupID,
				GroupName:                system.Name,
				ServiceName:              service.Name,
				Score:                    normalizeHealthScore(service.Score),
				SnapshotStatus:           serviceStatus,
				SnapshotSevereThreshold:  thresholds.SevereAbnormalScoreThreshold,
				SnapshotGeneralThreshold: thresholds.GeneralAbnormalScoreThreshold,
				CalcVersion:              systemHealthSnapshotCalcVersion,
			})
		}
	}
	return systemSnapshots, serviceSnapshots
}
