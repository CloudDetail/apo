// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package systemhealth

import (
	"math"
	"net/http"
	"sort"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"go.uber.org/zap"
)

const (
	systemHealthTrendJudgementCurrent  = "current"
	systemHealthTrendJudgementSnapshot = "snapshot"
	systemHealthTrendDefaultRange      = "last7d"
	systemHealthTrendViewSummary       = "summary"
	systemHealthTrendViewSystemScore   = "systemScore"
	systemHealthTrendViewDetail        = "detail"
)

type systemHealthTrendInterval struct {
	Start time.Time
	End   time.Time
}

type systemHealthTrendWindow struct {
	Range       string
	Granularity string
	Start       time.Time
	End         time.Time
	Intervals   []systemHealthTrendInterval
}

type systemHealthTrendSample struct {
	Timestamp         time.Time
	TotalSystems      int
	AbnormalSystems   int
	AttentionSystems  int
	SystemScore       float64
	SystemScoreRisk   bool
	HasSystemScore    bool
	AbnormalServices  int
	AttentionServices int
	HealthyServices   int
	NoDataServices    int
}

func (h *handler) GetTrend() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetSystemHealthTrendRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(http.StatusBadRequest, code.ParamBindError, err)
			return
		}

		resp, err := h.buildTrend(c, req, time.Now())
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, code.GetDataGroupError, err)
			return
		}
		c.Payload(resp)
	}
}

func (h *handler) GetSystemsTrend() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetSystemHealthSystemsTrendRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(http.StatusBadRequest, code.ParamBindError, err)
			return
		}

		resp, err := h.buildSystemsTrend(c, req, time.Now())
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, code.GetDataGroupError, err)
			return
		}
		c.Payload(resp)
	}
}

func (h *handler) buildTrend(ctx core.Context, req *request.GetSystemHealthTrendRequest, now time.Time) (response.SystemHealthTrendResp, error) {
	window := systemHealthTrendWindowFor(req.Range, now)
	judgement := normalizeSystemHealthTrendJudgement(req.Judgement)
	groups, err := h.visibleSystemGroups(ctx)
	if err != nil {
		return response.SystemHealthTrendResp{}, err
	}
	groupIDs := make([]int64, 0, len(groups))
	allowedGroups := make(map[int64]struct{}, len(groups))
	for _, group := range groups {
		allowedGroups[group.GroupID] = struct{}{}
		if req.GroupID == 0 || req.GroupID == group.GroupID {
			groupIDs = append(groupIDs, group.GroupID)
		}
	}
	if req.GroupID > 0 {
		if _, ok := allowedGroups[req.GroupID]; !ok {
			return response.SystemHealthTrendResp{}, nil
		}
	}

	thresholds, err := h.systemHealthConfig(ctx)
	if err != nil {
		systemHealthWarnCtx(ctx, "system-health trend config failed, using default threshold", zap.Error(err))
	}
	thresholds = normalizeSystemHealthScoreThresholds(thresholds)
	systemSnapshots, err := h.dbRepo.ListSystemHealthSystemSnapshots(ctx, groupIDs, window.Start, window.End)
	if err != nil {
		return response.SystemHealthTrendResp{}, err
	}
	var serviceSnapshots []database.SystemHealthServiceSnapshot
	if systemHealthTrendViewNeedsServiceSnapshots(req.View) {
		serviceSnapshots, err = h.dbRepo.ListSystemHealthServiceSnapshots(ctx, groupIDs, window.Start, window.End)
		if err != nil {
			return response.SystemHealthTrendResp{}, err
		}
	}

	return response.SystemHealthTrendResp{
		Range:       window.Range,
		Granularity: window.Granularity,
		Judgement:   judgement,
		Thresholds: response.SystemHealthTrendThreshold{
			Severe:  thresholds.SevereAbnormalScoreThreshold,
			General: thresholds.GeneralAbnormalScoreThreshold,
		},
		Points:      buildSystemHealthTrendPoints(systemSnapshots, serviceSnapshots, window.Intervals, thresholds, judgement),
		GeneratedAt: time.Now().Format(time.RFC3339),
		Source:      systemHealthSourceDataGroup,
	}, nil
}

func (h *handler) buildSystemsTrend(ctx core.Context, req *request.GetSystemHealthSystemsTrendRequest, now time.Time) (response.SystemHealthSystemsTrendResp, error) {
	window := systemHealthTrendWindowFor(req.Range, now)
	judgement := normalizeSystemHealthTrendJudgement(req.Judgement)
	groups, err := h.visibleSystemGroups(ctx)
	if err != nil {
		return response.SystemHealthSystemsTrendResp{}, err
	}
	allowedGroups := make(map[int64]struct{}, len(groups))
	for _, group := range groups {
		allowedGroups[group.GroupID] = struct{}{}
	}
	groupIDs := visibleRequestedSystemHealthGroupIDs(req.GroupIDs, allowedGroups)

	thresholds, err := h.systemHealthConfig(ctx)
	if err != nil {
		systemHealthWarnCtx(ctx, "system-health systems trend config failed, using default threshold", zap.Error(err))
	}
	thresholds = normalizeSystemHealthScoreThresholds(thresholds)
	systemSnapshots, err := h.dbRepo.ListSystemHealthSystemSnapshots(ctx, groupIDs, window.Start, window.End)
	if err != nil {
		return response.SystemHealthSystemsTrendResp{}, err
	}

	return response.SystemHealthSystemsTrendResp{
		Range:       window.Range,
		Granularity: window.Granularity,
		Judgement:   judgement,
		Groups:      buildSystemHealthSystemScoreTrendGroups(systemSnapshots, groupIDs, window.Intervals, thresholds, judgement),
		GeneratedAt: time.Now().Format(time.RFC3339),
		Source:      systemHealthSourceDataGroup,
	}, nil
}

func systemHealthTrendViewNeedsServiceSnapshots(view string) bool {
	switch view {
	case systemHealthTrendViewSummary, systemHealthTrendViewSystemScore:
		return false
	default:
		return true
	}
}

func visibleRequestedSystemHealthGroupIDs(requested []int64, allowed map[int64]struct{}) []int64 {
	groupIDs := make([]int64, 0, len(requested))
	seen := map[int64]struct{}{}
	for _, groupID := range requested {
		if groupID == 0 {
			continue
		}
		if _, ok := allowed[groupID]; !ok {
			continue
		}
		if _, ok := seen[groupID]; ok {
			continue
		}
		seen[groupID] = struct{}{}
		groupIDs = append(groupIDs, groupID)
	}
	return groupIDs
}

func buildSystemHealthSystemScoreTrendGroups(
	systemSnapshots []database.SystemHealthSystemSnapshot,
	groupIDs []int64,
	intervals []systemHealthTrendInterval,
	thresholds response.SystemHealthConfigResp,
	judgement string,
) []response.SystemHealthSystemTrendGroup {
	snapshotsByGroup := make(map[int64][]database.SystemHealthSystemSnapshot)
	for _, snapshot := range systemSnapshots {
		snapshotsByGroup[snapshot.GroupID] = append(snapshotsByGroup[snapshot.GroupID], snapshot)
	}
	groups := make([]response.SystemHealthSystemTrendGroup, 0, len(groupIDs))
	for _, groupID := range groupIDs {
		groups = append(groups, response.SystemHealthSystemTrendGroup{
			GroupID: groupID,
			Points:  buildSystemHealthTrendPoints(snapshotsByGroup[groupID], nil, intervals, thresholds, judgement),
		})
	}
	return groups
}

func buildSystemHealthTrendPoints(
	systemSnapshots []database.SystemHealthSystemSnapshot,
	serviceSnapshots []database.SystemHealthServiceSnapshot,
	intervals []systemHealthTrendInterval,
	thresholds response.SystemHealthConfigResp,
	judgement string,
) []response.SystemHealthTrendPoint {
	thresholds = normalizeSystemHealthScoreThresholds(thresholds)
	judgement = normalizeSystemHealthTrendJudgement(judgement)
	systemSamples := buildSystemHealthTrendSystemSamples(systemSnapshots, thresholds, judgement)
	serviceSamples := buildSystemHealthTrendServiceSamples(serviceSnapshots, thresholds, judgement)
	points := make([]response.SystemHealthTrendPoint, 0, len(intervals))
	for _, interval := range intervals {
		samplesByTimestamp := map[time.Time]systemHealthTrendSample{}
		for _, sample := range systemSamples {
			if !timeInTrendInterval(sample.Timestamp, interval) {
				continue
			}
			merged := samplesByTimestamp[sample.Timestamp]
			merged.Timestamp = sample.Timestamp
			merged.TotalSystems = sample.TotalSystems
			merged.AbnormalSystems = sample.AbnormalSystems
			merged.AttentionSystems = sample.AttentionSystems
			merged.SystemScore = sample.SystemScore
			merged.SystemScoreRisk = sample.SystemScoreRisk
			merged.HasSystemScore = sample.HasSystemScore
			samplesByTimestamp[sample.Timestamp] = merged
		}
		for _, sample := range serviceSamples {
			if !timeInTrendInterval(sample.Timestamp, interval) {
				continue
			}
			merged := samplesByTimestamp[sample.Timestamp]
			merged.Timestamp = sample.Timestamp
			merged.AbnormalServices = sample.AbnormalServices
			merged.AttentionServices = sample.AttentionServices
			merged.HealthyServices = sample.HealthyServices
			merged.NoDataServices = sample.NoDataServices
			samplesByTimestamp[sample.Timestamp] = merged
		}
		samples := make([]systemHealthTrendSample, 0, len(samplesByTimestamp))
		for _, sample := range samplesByTimestamp {
			samples = append(samples, sample)
		}
		sort.SliceStable(samples, func(i, j int) bool {
			return samples[i].Timestamp.Before(samples[j].Timestamp)
		})
		if len(samples) == 0 {
			points = append(points, response.SystemHealthTrendPoint{
				Timestamp:       interval.Start.UnixMilli(),
				MissingSnapshot: true,
			})
			continue
		}
		points = append(points, response.SystemHealthTrendPoint{
			Timestamp:        interval.Start.UnixMilli(),
			AbnormalSystems:  buildTrendCountStat(samples, func(sample systemHealthTrendSample) int { return sample.AbnormalSystems }),
			AttentionSystems: buildTrendCountStat(samples, func(sample systemHealthTrendSample) int { return sample.AttentionSystems }),
			AbnormalSystemRatio: buildTrendRatioStat(samples, func(sample systemHealthTrendSample) int {
				return sample.AbnormalSystems
			}, func(sample systemHealthTrendSample) int {
				return sample.TotalSystems
			}),
			AttentionSystemRatio: buildTrendRatioStat(samples, func(sample systemHealthTrendSample) int {
				return sample.AttentionSystems
			}, func(sample systemHealthTrendSample) int {
				return sample.TotalSystems
			}),
			SystemHealthScore: buildTrendScoreStat(samples),
			AbnormalServices:  buildTrendCountStat(samples, func(sample systemHealthTrendSample) int { return sample.AbnormalServices }),
			AttentionServices: buildTrendCountStat(samples, func(sample systemHealthTrendSample) int { return sample.AttentionServices }),
			HealthyServices:   buildTrendCountStat(samples, func(sample systemHealthTrendSample) int { return sample.HealthyServices }),
			NoDataServices:    buildTrendCountStat(samples, func(sample systemHealthTrendSample) int { return sample.NoDataServices }),
		})
	}
	return points
}

func buildSystemHealthTrendSystemSamples(
	snapshots []database.SystemHealthSystemSnapshot,
	thresholds response.SystemHealthConfigResp,
	judgement string,
) []systemHealthTrendSample {
	byTime := map[time.Time][]database.SystemHealthSystemSnapshot{}
	for _, snapshot := range snapshots {
		byTime[snapshot.BucketStart] = append(byTime[snapshot.BucketStart], snapshot)
	}
	timestamps := make([]time.Time, 0, len(byTime))
	for timestamp := range byTime {
		timestamps = append(timestamps, timestamp)
	}
	sort.SliceStable(timestamps, func(i, j int) bool {
		return timestamps[i].Before(timestamps[j])
	})

	samples := make([]systemHealthTrendSample, 0, len(timestamps))
	for _, timestamp := range timestamps {
		sample := systemHealthTrendSample{Timestamp: timestamp}
		sample.TotalSystems = len(byTime[timestamp])
		scoreTotal := 0
		scoreCount := 0
		for _, snapshot := range byTime[timestamp] {
			status := systemHealthSnapshotStatus(snapshot.Score, snapshot.SnapshotStatus, snapshot.SnapshotSevereThreshold, snapshot.SnapshotGeneralThreshold, thresholds, judgement)
			if status == healthStatusGray {
				continue
			}
			scoreTotal += snapshot.Score
			scoreCount++
			switch status {
			case healthStatusRed:
				sample.AbnormalSystems++
				sample.SystemScoreRisk = true
			case healthStatusYellow:
				sample.AttentionSystems++
				sample.SystemScoreRisk = true
			}
		}
		if scoreCount > 0 {
			sample.SystemScore = float64(scoreTotal) / float64(scoreCount)
			sample.HasSystemScore = true
		}
		samples = append(samples, sample)
	}
	return samples
}

func buildSystemHealthTrendServiceSamples(
	snapshots []database.SystemHealthServiceSnapshot,
	thresholds response.SystemHealthConfigResp,
	judgement string,
) []systemHealthTrendSample {
	byTime := map[time.Time][]database.SystemHealthServiceSnapshot{}
	for _, snapshot := range snapshots {
		byTime[snapshot.BucketStart] = append(byTime[snapshot.BucketStart], snapshot)
	}
	timestamps := make([]time.Time, 0, len(byTime))
	for timestamp := range byTime {
		timestamps = append(timestamps, timestamp)
	}
	sort.SliceStable(timestamps, func(i, j int) bool {
		return timestamps[i].Before(timestamps[j])
	})

	samples := make([]systemHealthTrendSample, 0, len(timestamps))
	for _, timestamp := range timestamps {
		sample := systemHealthTrendSample{Timestamp: timestamp}
		for _, snapshot := range byTime[timestamp] {
			status := systemHealthSnapshotStatus(snapshot.Score, snapshot.SnapshotStatus, snapshot.SnapshotSevereThreshold, snapshot.SnapshotGeneralThreshold, thresholds, judgement)
			switch status {
			case healthStatusRed:
				sample.AbnormalServices++
			case healthStatusYellow:
				sample.AttentionServices++
			case healthStatusGreen:
				sample.HealthyServices++
			case healthStatusGray:
				sample.NoDataServices++
			}
		}
		samples = append(samples, sample)
	}
	return samples
}

func buildTrendCountStat(samples []systemHealthTrendSample, value func(systemHealthTrendSample) int) response.SystemHealthTrendCountStat {
	if len(samples) == 0 {
		return response.SystemHealthTrendCountStat{}
	}
	total := 0
	peak := 0
	minValue := math.MaxFloat64
	active := 0
	last := 0
	for _, sample := range samples {
		current := value(sample)
		total += current
		if current > peak {
			peak = current
		}
		if float64(current) < minValue {
			minValue = float64(current)
		}
		if current > 0 {
			active++
		}
		last = current
	}
	return response.SystemHealthTrendCountStat{
		Avg:         float64(total) / float64(len(samples)),
		Peak:        peak,
		Min:         minValue,
		Last:        float64(last),
		ActiveRatio: float64(active) / float64(len(samples)),
		SampleCount: len(samples),
	}
}

func buildTrendRatioStat(
	samples []systemHealthTrendSample,
	numerator func(systemHealthTrendSample) int,
	denominator func(systemHealthTrendSample) int,
) response.SystemHealthTrendRatioStat {
	total := 0.0
	count := 0
	peak := 0.0
	minValue := math.MaxFloat64
	active := 0
	last := 0.0
	for _, sample := range samples {
		base := denominator(sample)
		if base <= 0 {
			continue
		}
		current := float64(numerator(sample)) / float64(base)
		total += current
		count++
		if current > peak {
			peak = current
		}
		if current < minValue {
			minValue = current
		}
		if current > 0 {
			active++
		}
		last = current
	}
	if count == 0 {
		return response.SystemHealthTrendRatioStat{}
	}
	return response.SystemHealthTrendRatioStat{
		Avg:         total / float64(count),
		Peak:        peak,
		Min:         minValue,
		Last:        last,
		ActiveRatio: float64(active) / float64(count),
		SampleCount: count,
	}
}

func buildTrendScoreStat(samples []systemHealthTrendSample) response.SystemHealthTrendScoreStat {
	total := 0.0
	count := 0
	minValue := math.MaxFloat64
	maxValue := -math.MaxFloat64
	last := 0.0
	riskSamples := 0
	for _, sample := range samples {
		if !sample.HasSystemScore {
			continue
		}
		total += sample.SystemScore
		count++
		if sample.SystemScore < minValue {
			minValue = sample.SystemScore
		}
		if sample.SystemScore > maxValue {
			maxValue = sample.SystemScore
		}
		if sample.SystemScoreRisk {
			riskSamples++
		}
		last = sample.SystemScore
	}
	if count == 0 {
		return response.SystemHealthTrendScoreStat{}
	}
	return response.SystemHealthTrendScoreStat{
		Avg:                 total / float64(count),
		Min:                 minValue,
		Max:                 maxValue,
		Last:                last,
		BelowThresholdRatio: float64(riskSamples) / float64(count),
		SampleCount:         count,
	}
}

func systemHealthSnapshotStatus(
	score int,
	snapshotStatus string,
	snapshotSevereThreshold int,
	snapshotGeneralThreshold int,
	currentThresholds response.SystemHealthConfigResp,
	judgement string,
) string {
	if snapshotStatus == healthStatusGray {
		return healthStatusGray
	}
	if judgement == systemHealthTrendJudgementSnapshot && isKnownHealthStatus(snapshotStatus) {
		return snapshotStatus
	}
	thresholds := currentThresholds
	if judgement == systemHealthTrendJudgementSnapshot {
		thresholds = normalizeSystemHealthScoreThresholds(response.SystemHealthConfigResp{
			SevereAbnormalScoreThreshold:  snapshotSevereThreshold,
			GeneralAbnormalScoreThreshold: snapshotGeneralThreshold,
		})
	}
	return scoreHealthStatus(score, thresholds)
}

func normalizeSystemHealthTrendJudgement(judgement string) string {
	if judgement == systemHealthTrendJudgementCurrent {
		return judgement
	}
	if judgement == systemHealthTrendJudgementSnapshot {
		return judgement
	}
	return systemHealthTrendJudgementSnapshot
}

func isKnownHealthStatus(status string) bool {
	return status == healthStatusRed || status == healthStatusYellow || status == healthStatusGreen || status == healthStatusGray
}

func timeInTrendInterval(value time.Time, interval systemHealthTrendInterval) bool {
	return !value.Before(interval.Start) && value.Before(interval.End)
}

func systemHealthTrendWindowFor(rangeKey string, now time.Time) systemHealthTrendWindow {
	if rangeKey == "" {
		rangeKey = systemHealthTrendDefaultRange
	}
	end := now
	var start time.Time
	step := time.Hour
	granularity := "hour"
	switch rangeKey {
	case "today":
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case "last24h":
		start = now.Add(-24 * time.Hour)
	case "thisWeek":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		start = dayStart.AddDate(0, 0, -(weekday - 1))
	case "last15d":
		start = now.AddDate(0, 0, -15)
		step = 24 * time.Hour
		granularity = "day"
	case "last30d":
		start = now.AddDate(0, 0, -30)
		step = 24 * time.Hour
		granularity = "day"
	case "last6mo":
		start = now.AddDate(0, -6, 0)
		step = 7 * 24 * time.Hour
		granularity = "week"
	case "last1y":
		start = now.AddDate(-1, 0, 0)
		step = 7 * 24 * time.Hour
		granularity = "week"
	default:
		rangeKey = "last7d"
		start = now.AddDate(0, 0, -7)
		step = 6 * time.Hour
		granularity = "6h"
	}
	return systemHealthTrendWindow{
		Range:       rangeKey,
		Granularity: granularity,
		Start:       alignTrendTime(start, step),
		End:         end,
		Intervals:   buildSystemHealthTrendIntervals(alignTrendTime(start, step), end, step),
	}
}

func buildSystemHealthTrendIntervals(start time.Time, end time.Time, step time.Duration) []systemHealthTrendInterval {
	if !end.After(start) {
		return nil
	}
	intervals := make([]systemHealthTrendInterval, 0)
	for cursor := start; cursor.Before(end); cursor = cursor.Add(step) {
		next := cursor.Add(step)
		intervals = append(intervals, systemHealthTrendInterval{Start: cursor, End: next})
	}
	return intervals
}

func alignTrendTime(value time.Time, step time.Duration) time.Time {
	dayStart := time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
	if step >= 24*time.Hour {
		return dayStart
	}
	return dayStart.Add(time.Duration(int64(value.Sub(dayStart)/step)) * step)
}
