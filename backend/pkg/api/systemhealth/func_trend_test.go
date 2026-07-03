// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package systemhealth

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/datagroup"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

func TestBuildSystemHealthTrendPointsUsesCurrentThresholdsAndPreservesRisk(t *testing.T) {
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	mid := start.Add(time.Hour)
	end := start.Add(2 * time.Hour)
	intervals := []systemHealthTrendInterval{{Start: start, End: end}}
	thresholds := response.SystemHealthConfigResp{
		SevereAbnormalScoreThreshold:  60,
		GeneralAbnormalScoreThreshold: 80,
	}
	systemSnapshots := []database.SystemHealthSystemSnapshot{
		{BucketStart: start, BucketEnd: mid, GroupID: 101, Score: 55, SnapshotStatus: healthStatusGreen, SnapshotSevereThreshold: 40, SnapshotGeneralThreshold: 50},
		{BucketStart: start, BucketEnd: mid, GroupID: 102, Score: 82, SnapshotStatus: healthStatusGreen, SnapshotSevereThreshold: 40, SnapshotGeneralThreshold: 50},
		{BucketStart: mid, BucketEnd: end, GroupID: 101, Score: 72, SnapshotStatus: healthStatusGreen, SnapshotSevereThreshold: 40, SnapshotGeneralThreshold: 50},
		{BucketStart: mid, BucketEnd: end, GroupID: 102, Score: 90, SnapshotStatus: healthStatusGreen, SnapshotSevereThreshold: 40, SnapshotGeneralThreshold: 50},
	}
	serviceSnapshots := []database.SystemHealthServiceSnapshot{
		{BucketStart: start, BucketEnd: mid, GroupID: 101, ServiceName: "order-api", Score: 50, SnapshotStatus: healthStatusGreen, SnapshotSevereThreshold: 40, SnapshotGeneralThreshold: 50},
		{BucketStart: start, BucketEnd: mid, GroupID: 101, ServiceName: "order-worker", Score: 75, SnapshotStatus: healthStatusGreen, SnapshotSevereThreshold: 40, SnapshotGeneralThreshold: 50},
		{BucketStart: start, BucketEnd: mid, GroupID: 102, ServiceName: "pay-api", Score: 95, SnapshotStatus: healthStatusGreen, SnapshotSevereThreshold: 40, SnapshotGeneralThreshold: 50},
		{BucketStart: mid, BucketEnd: end, GroupID: 101, ServiceName: "order-api", Score: 90, SnapshotStatus: healthStatusGreen, SnapshotSevereThreshold: 40, SnapshotGeneralThreshold: 50},
		{BucketStart: mid, BucketEnd: end, GroupID: 101, ServiceName: "order-worker", Score: 75, SnapshotStatus: healthStatusGreen, SnapshotSevereThreshold: 40, SnapshotGeneralThreshold: 50},
		{BucketStart: mid, BucketEnd: end, GroupID: 102, ServiceName: "pay-api", Score: 95, SnapshotStatus: healthStatusGreen, SnapshotSevereThreshold: 40, SnapshotGeneralThreshold: 50},
		{BucketStart: mid, BucketEnd: end, GroupID: 102, ServiceName: "pay-cache", Score: 0, SnapshotStatus: healthStatusGray, SnapshotSevereThreshold: 40, SnapshotGeneralThreshold: 50},
	}

	points := buildSystemHealthTrendPoints(systemSnapshots, serviceSnapshots, intervals, thresholds, systemHealthTrendJudgementCurrent)

	if len(points) != 1 {
		t.Fatalf("expected one trend point, got %d", len(points))
	}
	point := points[0]
	assertFloat(t, point.AbnormalSystems.Avg, 0.5)
	assertInt(t, point.AbnormalSystems.Peak, 1)
	assertFloat(t, point.AbnormalSystems.Last, 0)
	assertFloat(t, point.AbnormalSystems.ActiveRatio, 0.5)
	assertInt(t, point.AttentionSystems.Peak, 1)
	assertFloat(t, point.AttentionSystems.Last, 1)
	assertFloat(t, point.AttentionSystems.ActiveRatio, 0.5)
	assertFloat(t, point.AbnormalSystemRatio.Avg, 0.25)
	assertFloat(t, point.AbnormalSystemRatio.Peak, 0.5)
	assertFloat(t, point.AbnormalSystemRatio.Last, 0)
	assertFloat(t, point.AttentionSystemRatio.Avg, 0.25)
	assertFloat(t, point.AttentionSystemRatio.Peak, 0.5)
	assertFloat(t, point.AttentionSystemRatio.Last, 0.5)
	assertFloat(t, point.SystemHealthScore.Avg, 74.75)
	assertFloat(t, point.SystemHealthScore.Min, 68.5)
	assertFloat(t, point.SystemHealthScore.Last, 81)
	assertFloat(t, point.SystemHealthScore.BelowThresholdRatio, 1)
	assertFloat(t, point.AbnormalServices.Avg, 0.5)
	assertInt(t, point.AbnormalServices.Peak, 1)
	assertFloat(t, point.AttentionServices.Avg, 1)
	assertInt(t, point.AttentionServices.Peak, 1)
	assertFloat(t, point.AttentionServices.ActiveRatio, 1)
	assertFloat(t, point.HealthyServices.Avg, 1.5)
	assertFloat(t, point.HealthyServices.Min, 1)
	assertFloat(t, point.HealthyServices.Last, 2)
	assertFloat(t, point.NoDataServices.Avg, 0.5)
	assertInt(t, point.NoDataServices.Peak, 1)
	assertFloat(t, point.NoDataServices.Last, 1)
}

func TestBuildSystemHealthTrendPointsCalculatesSystemRatiosAgainstTotalSnapshots(t *testing.T) {
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	intervals := []systemHealthTrendInterval{{Start: start, End: end}}
	thresholds := response.SystemHealthConfigResp{
		SevereAbnormalScoreThreshold:  60,
		GeneralAbnormalScoreThreshold: 80,
	}
	systemSnapshots := []database.SystemHealthSystemSnapshot{
		{BucketStart: start, BucketEnd: end, GroupID: 101, Score: 50, SnapshotStatus: healthStatusRed},
		{BucketStart: start, BucketEnd: end, GroupID: 102, Score: 70, SnapshotStatus: healthStatusYellow},
		{BucketStart: start, BucketEnd: end, GroupID: 103, Score: 90, SnapshotStatus: healthStatusGreen},
		{BucketStart: start, BucketEnd: end, GroupID: 104, Score: 0, SnapshotStatus: healthStatusGray},
	}

	points := buildSystemHealthTrendPoints(systemSnapshots, nil, intervals, thresholds, systemHealthTrendJudgementSnapshot)

	if len(points) != 1 {
		t.Fatalf("expected one trend point, got %d", len(points))
	}
	assertFloat(t, points[0].AbnormalSystemRatio.Peak, 0.25)
	assertFloat(t, points[0].AttentionSystemRatio.Peak, 0.25)
	assertInt(t, points[0].AbnormalSystemRatio.SampleCount, 1)
	assertInt(t, points[0].AttentionSystemRatio.SampleCount, 1)
}

func TestBuildSystemHealthTrendPointsKeepsIntervalsWithoutSnapshots(t *testing.T) {
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	second := start.Add(time.Hour)
	end := start.Add(2 * time.Hour)
	intervals := []systemHealthTrendInterval{
		{Start: start, End: second},
		{Start: second, End: end},
	}
	thresholds := response.SystemHealthConfigResp{
		SevereAbnormalScoreThreshold:  60,
		GeneralAbnormalScoreThreshold: 80,
	}
	systemSnapshots := []database.SystemHealthSystemSnapshot{
		{BucketStart: second, BucketEnd: end, GroupID: 101, Score: 95, SnapshotStatus: healthStatusGreen},
	}

	points := buildSystemHealthTrendPoints(systemSnapshots, nil, intervals, thresholds, systemHealthTrendJudgementCurrent)

	if len(points) != 2 {
		t.Fatalf("expected one trend point per interval, got %d", len(points))
	}
	if points[0].Timestamp != start.UnixMilli() || !points[0].MissingSnapshot {
		t.Fatalf("expected empty first interval to be marked missing, got %+v", points[0])
	}
	if points[0].NoDataServices.SampleCount != 0 {
		t.Fatalf("expected missing snapshot not to count as no-data services, got %+v", points[0].NoDataServices)
	}
	if points[1].Timestamp != second.UnixMilli() || points[1].MissingSnapshot {
		t.Fatalf("expected sampled second interval, got %+v", points[1])
	}
}

func TestBuildSystemHealthTrendPointsCanUseSnapshotJudgement(t *testing.T) {
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	intervals := []systemHealthTrendInterval{{Start: start, End: end}}
	thresholds := response.SystemHealthConfigResp{
		SevereAbnormalScoreThreshold:  60,
		GeneralAbnormalScoreThreshold: 80,
	}
	serviceSnapshots := []database.SystemHealthServiceSnapshot{
		{BucketStart: start, BucketEnd: end, GroupID: 101, ServiceName: "order-api", Score: 75, SnapshotStatus: healthStatusGreen, SnapshotSevereThreshold: 50, SnapshotGeneralThreshold: 70},
	}

	currentPoints := buildSystemHealthTrendPoints(nil, serviceSnapshots, intervals, thresholds, systemHealthTrendJudgementCurrent)
	snapshotPoints := buildSystemHealthTrendPoints(nil, serviceSnapshots, intervals, thresholds, systemHealthTrendJudgementSnapshot)

	if currentPoints[0].AttentionServices.Peak != 1 {
		t.Fatalf("expected current thresholds to classify score 75 as attention, got %+v", currentPoints[0].AttentionServices)
	}
	if snapshotPoints[0].AttentionServices.Peak != 0 || snapshotPoints[0].HealthyServices.Last != 1 {
		t.Fatalf("expected snapshot thresholds to keep score 75 healthy, got attention=%+v healthy=%+v", snapshotPoints[0].AttentionServices, snapshotPoints[0].HealthyServices)
	}
}

func TestNormalizeSystemHealthTrendJudgementDefaultsToSnapshot(t *testing.T) {
	if got := normalizeSystemHealthTrendJudgement(""); got != systemHealthTrendJudgementSnapshot {
		t.Fatalf("expected empty judgement to default to snapshot, got %s", got)
	}
	if got := normalizeSystemHealthTrendJudgement("unknown"); got != systemHealthTrendJudgementSnapshot {
		t.Fatalf("expected unknown judgement to default to snapshot, got %s", got)
	}
	if got := normalizeSystemHealthTrendJudgement(systemHealthTrendJudgementCurrent); got != systemHealthTrendJudgementCurrent {
		t.Fatalf("expected explicit current judgement to be preserved, got %s", got)
	}
}

func TestSystemHealthTrendViewControlsServiceSnapshotNeed(t *testing.T) {
	if systemHealthTrendViewNeedsServiceSnapshots(systemHealthTrendViewSummary) {
		t.Fatalf("summary trend view should not load service snapshots")
	}
	if systemHealthTrendViewNeedsServiceSnapshots(systemHealthTrendViewSystemScore) {
		t.Fatalf("systemScore trend view should not load service snapshots")
	}
	if !systemHealthTrendViewNeedsServiceSnapshots(systemHealthTrendViewDetail) {
		t.Fatalf("detail trend view should load service snapshots")
	}
	if !systemHealthTrendViewNeedsServiceSnapshots("") {
		t.Fatalf("compatibility trend view should keep loading service snapshots")
	}
}

func TestBuildSystemHealthSystemScoreTrendGroupsKeepsRequestedGroupOrder(t *testing.T) {
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	intervals := []systemHealthTrendInterval{{Start: start, End: end}}
	thresholds := response.SystemHealthConfigResp{
		SevereAbnormalScoreThreshold:  60,
		GeneralAbnormalScoreThreshold: 80,
	}
	systemSnapshots := []database.SystemHealthSystemSnapshot{
		{BucketStart: start, BucketEnd: end, GroupID: 101, Score: 91, SnapshotStatus: healthStatusGreen},
		{BucketStart: start, BucketEnd: end, GroupID: 102, Score: 52, SnapshotStatus: healthStatusRed},
	}

	groups := buildSystemHealthSystemScoreTrendGroups(
		systemSnapshots,
		[]int64{102, 101, 103},
		intervals,
		thresholds,
		systemHealthTrendJudgementSnapshot,
	)

	if len(groups) != 3 {
		t.Fatalf("expected one trend group per requested id, got %d", len(groups))
	}
	if groups[0].GroupID != 102 || groups[1].GroupID != 101 || groups[2].GroupID != 103 {
		t.Fatalf("expected requested group order to be preserved, got %+v", groups)
	}
	if len(groups[0].Points) != 1 || groups[0].Points[0].SystemHealthScore.Last != 52 {
		t.Fatalf("expected group 102 score trend point, got %+v", groups[0])
	}
	if len(groups[2].Points) != 1 || !groups[2].Points[0].MissingSnapshot {
		t.Fatalf("expected missing snapshot point for requested group without data, got %+v", groups[2])
	}
}

func TestBuildSystemHealthSnapshotRecordsStoresSystemAndServiceScoresWithThresholds(t *testing.T) {
	start := time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	thresholds := response.SystemHealthConfigResp{
		SevereAbnormalScoreThreshold:  60,
		GeneralAbnormalScoreThreshold: 80,
	}
	systems := []response.SystemHealthSystem{
		{
			GroupID: 101,
			Name:    "订单系统",
			Score:   72,
			Status:  healthStatusYellow,
			Services: []response.SystemHealthService{
				{Name: "order-api", Score: 55, Status: healthStatusRed},
				{Name: "order-worker", Score: 90, Status: healthStatusGreen},
			},
		},
		{
			GroupID: 102,
			Name:    "空系统",
			Score:   0,
			Status:  healthStatusGray,
		},
	}

	systemSnapshots, serviceSnapshots := buildSystemHealthSnapshotRecords(systems, thresholds, start, end)

	if len(systemSnapshots) != 2 {
		t.Fatalf("expected two system snapshots, got %d", len(systemSnapshots))
	}
	if systemSnapshots[0].GroupID != 101 || systemSnapshots[0].Score != 72 || systemSnapshots[0].SnapshotStatus != healthStatusYellow {
		t.Fatalf("unexpected first system snapshot: %+v", systemSnapshots[0])
	}
	if systemSnapshots[0].ServiceTotal != 2 || systemSnapshots[0].SnapshotSevereThreshold != 60 || systemSnapshots[0].SnapshotGeneralThreshold != 80 {
		t.Fatalf("expected service total and thresholds stored, got %+v", systemSnapshots[0])
	}
	if systemSnapshots[1].SnapshotStatus != healthStatusGray {
		t.Fatalf("expected no-data system to stay gray, got %+v", systemSnapshots[1])
	}
	if len(serviceSnapshots) != 2 {
		t.Fatalf("expected two service snapshots, got %d", len(serviceSnapshots))
	}
	if serviceSnapshots[0].GroupID != 101 || serviceSnapshots[0].ServiceName != "order-api" || serviceSnapshots[0].Score != 55 || serviceSnapshots[0].SnapshotStatus != healthStatusRed {
		t.Fatalf("unexpected first service snapshot: %+v", serviceSnapshots[0])
	}
	if serviceSnapshots[1].SnapshotGeneralThreshold != 80 {
		t.Fatalf("expected thresholds stored on service snapshot, got %+v", serviceSnapshots[1])
	}
}

func TestSystemHealthTrendWindowAlignsSixHourBucketsToLocalDay(t *testing.T) {
	location := time.FixedZone("CST", 8*60*60)
	now := time.Date(2026, 7, 1, 11, 30, 0, 0, location)

	window := systemHealthTrendWindowFor("last7d", now)

	if len(window.Intervals) == 0 {
		t.Fatal("expected trend intervals")
	}
	first := window.Intervals[0].Start
	if first.Hour() != 6 || first.Minute() != 0 {
		t.Fatalf("expected first 6h bucket to align to local 06:00, got %s", first.Format(time.RFC3339))
	}
}

func TestSystemHealthTrendWindowSupportsPastFifteenDays(t *testing.T) {
	location := time.FixedZone("CST", 8*60*60)
	now := time.Date(2026, 7, 1, 11, 30, 0, 0, location)

	window := systemHealthTrendWindowFor("last15d", now)

	if window.Range != "last15d" {
		t.Fatalf("expected last15d range to be preserved, got %s", window.Range)
	}
	if window.Granularity != "day" {
		t.Fatalf("expected last15d to use daily granularity, got %s", window.Granularity)
	}
	assertTime(t, window.Start, time.Date(2026, 6, 16, 0, 0, 0, 0, location))
	if len(window.Intervals) != 16 {
		t.Fatalf("expected aligned daily intervals covering the past 15 days, got %d", len(window.Intervals))
	}
	assertTime(t, window.Intervals[0].Start, window.Start)
	assertTime(t, window.Intervals[0].End, time.Date(2026, 6, 17, 0, 0, 0, 0, location))
}

func TestSnapshotSystemGroupsFromRootIgnoresUserPermissionFiltering(t *testing.T) {
	root := &datagroup.DataGroupTreeNode{
		DataGroup: datagroup.DataGroup{GroupID: 0, GroupName: "ALL"},
		SubGroups: []*datagroup.DataGroupTreeNode{
			{DataGroup: datagroup.DataGroup{GroupID: 102, GroupName: "支付系统"}},
			{DataGroup: datagroup.DataGroup{GroupID: 101, GroupName: "订单系统"}},
		},
	}

	groups := snapshotSystemGroupsFromRoot(root)

	if len(groups) != 2 {
		t.Fatalf("expected all first-level groups for background snapshots, got %d", len(groups))
	}
	if groups[0].GroupID != 101 || groups[1].GroupID != 102 {
		t.Fatalf("expected groups sorted by id without permission filtering, got %+v", groups)
	}
}

func TestSystemHealthSnapshotWindowsDueBackfillsMissingCompleteHours(t *testing.T) {
	location := time.FixedZone("CST", 8*60*60)
	now := time.Date(2026, 7, 1, 5, 10, 0, 0, location)
	latestSaved := time.Date(2026, 7, 1, 1, 0, 0, 0, location)

	windows := systemHealthSnapshotWindowsDue(&latestSaved, now, 24)

	if len(windows) != 3 {
		t.Fatalf("expected three missing windows, got %d", len(windows))
	}
	assertTime(t, windows[0].Start, time.Date(2026, 7, 1, 2, 0, 0, 0, location))
	assertTime(t, windows[2].Start, time.Date(2026, 7, 1, 4, 0, 0, 0, location))
}

func TestSystemHealthSnapshotWindowsDueCapsLongBackfillAndWaitsUntilMinuteTen(t *testing.T) {
	location := time.FixedZone("CST", 8*60*60)
	latestSaved := time.Date(2026, 6, 1, 0, 0, 0, 0, location)

	tooEarly := systemHealthSnapshotWindowsDue(&latestSaved, time.Date(2026, 7, 1, 5, 9, 0, 0, location), 24)
	if len(tooEarly) != 0 {
		t.Fatalf("expected no windows before minute 10, got %d", len(tooEarly))
	}

	windows := systemHealthSnapshotWindowsDue(&latestSaved, time.Date(2026, 7, 1, 5, 10, 0, 0, location), 3)
	if len(windows) != 3 {
		t.Fatalf("expected capped backfill windows, got %d", len(windows))
	}
	assertTime(t, windows[0].Start, time.Date(2026, 7, 1, 2, 0, 0, 0, location))
	assertTime(t, windows[2].Start, time.Date(2026, 7, 1, 4, 0, 0, 0, location))
}

func TestSystemHealthSnapshotWindowsDueStartsAtLatestCompleteHourWithoutHistory(t *testing.T) {
	location := time.FixedZone("CST", 8*60*60)
	windows := systemHealthSnapshotWindowsDue(nil, time.Date(2026, 7, 1, 5, 10, 0, 0, location), 24)

	if len(windows) != 1 {
		t.Fatalf("expected one latest complete window without history, got %d", len(windows))
	}
	assertTime(t, windows[0].Start, time.Date(2026, 7, 1, 4, 0, 0, 0, location))
}

func TestCollectSystemHealthSnapshotWindowsStopsBeforeLaterBucketsOnUnclaimedLock(t *testing.T) {
	start := time.Date(2026, 7, 1, 1, 0, 0, 0, time.UTC)
	windows := []systemHealthSnapshotWindow{
		{Start: start, End: start.Add(time.Hour)},
		{Start: start.Add(time.Hour), End: start.Add(2 * time.Hour)},
		{Start: start.Add(2 * time.Hour), End: start.Add(3 * time.Hour)},
	}
	var claimed []time.Time
	var collected []time.Time

	collectSystemHealthSnapshotWindows(
		core.EmptyCtx(),
		windows,
		nil,
		func(_ core.Context, window systemHealthSnapshotWindow) (bool, error) {
			claimed = append(claimed, window.Start)
			return len(claimed) == 1, nil
		},
		func(_ core.Context, window systemHealthSnapshotWindow) error {
			collected = append(collected, window.Start)
			return nil
		},
	)

	if len(claimed) != 2 {
		t.Fatalf("expected to claim until the first unclaimed bucket, got %d claims", len(claimed))
	}
	if len(collected) != 1 {
		t.Fatalf("expected only the first bucket collected before an unclaimed lock, got %d", len(collected))
	}
	assertTime(t, claimed[1], start.Add(time.Hour))
	assertTime(t, collected[0], start)
}

func TestCollectSystemHealthSnapshotWindowsStopsBeforeLaterBucketsOnCollectFailure(t *testing.T) {
	start := time.Date(2026, 7, 1, 1, 0, 0, 0, time.UTC)
	windows := []systemHealthSnapshotWindow{
		{Start: start, End: start.Add(time.Hour)},
		{Start: start.Add(time.Hour), End: start.Add(2 * time.Hour)},
		{Start: start.Add(2 * time.Hour), End: start.Add(3 * time.Hour)},
	}
	var claimed []time.Time
	var collected []time.Time

	collectSystemHealthSnapshotWindows(
		core.EmptyCtx(),
		windows,
		nil,
		func(_ core.Context, window systemHealthSnapshotWindow) (bool, error) {
			claimed = append(claimed, window.Start)
			return true, nil
		},
		func(_ core.Context, window systemHealthSnapshotWindow) error {
			collected = append(collected, window.Start)
			if len(collected) == 2 {
				return errors.New("collect failed")
			}
			return nil
		},
	)

	if len(claimed) != 2 {
		t.Fatalf("expected claims to stop at the failed collection bucket, got %d", len(claimed))
	}
	if len(collected) != 2 {
		t.Fatalf("expected collections to stop at the failed bucket, got %d", len(collected))
	}
	assertTime(t, claimed[1], start.Add(time.Hour))
	assertTime(t, collected[1], start.Add(time.Hour))
}

func assertInt(t *testing.T, got int, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("expected %d, got %d", want, got)
	}
}

func assertFloat(t *testing.T, got float64, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.0001 {
		t.Fatalf("expected %.4f, got %.4f", want, got)
	}
}

func assertTime(t *testing.T, got time.Time, want time.Time) {
	t.Helper()
	if !got.Equal(want) {
		t.Fatalf("expected %s, got %s", want.Format(time.RFC3339), got.Format(time.RFC3339))
	}
}
