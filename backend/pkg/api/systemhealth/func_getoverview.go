// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package systemhealth

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/datagroup"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/common"
	"github.com/CloudDetail/apo/backend/pkg/services/serviceoverview"
	"go.uber.org/zap"
)

const (
	systemHealthSourceDataGroup = "data_group"

	healthMetricResource      = "resource"
	healthMetricResourceShift = "resourceShift"
	healthMetricErrorRate     = "errorRate"
	healthMetricLatency       = "latency"
	healthMetricErrorLog      = "errorLog"
	healthMetricBadRequests   = "badRequests" // Legacy API key: active/firing alerts, not HTTP bad requests.
	healthMetricLockCost      = "lockCost"

	healthStatusGreen  = "green"
	healthStatusYellow = "yellow"
	healthStatusRed    = "red"
	healthStatusGray   = "gray"

	systemHealthOverviewGroupConcurrency = 4
	systemHealthOverviewCurrentTimeout   = 2 * time.Second
	systemHealthOverviewHistoryTimeout   = time.Second
	systemHealthDefaultSystemPageSize    = 50
	systemHealthDefaultServicePageSize   = 50
	systemHealthDefaultIssuePageSize     = 20
	systemHealthMaxPageSize              = 100
	systemHealthCurrentCacheTTL          = 30 * time.Second
	maxSystemHealthCurrentCacheEntries   = 128
	systemHealthCurrentCacheCalcVersion  = "current-v1"
	systemHealthErrorRateRedThreshold    = 0.05
	systemHealthErrorRateRedMinRequests  = 100
	systemHealthErrorRateRedMinErrors    = 5
	systemHealthLatencyRedMinRequests    = 20
	systemHealthLatencyYellowThreshold   = 500_000_000
	systemHealthLatencyRedThreshold      = 1_000_000_000
	systemHealthLatencyDoDYellow         = 10
	systemHealthLatencyDoDRed            = 50
	systemHealthErrorLogDoDYellow        = 5
	systemHealthErrorLogDoDRed           = 20
	systemHealthLockCostRedMinRequests   = 20
	systemHealthResourceShiftRedMinLevel = 30
	systemHealthResourceWaterlineYellow  = 70
	systemHealthResourceWaterlineRed     = 90
)

var systemHealthMetricKeys = []string{
	healthMetricResource,
	healthMetricResourceShift,
	healthMetricErrorRate,
	healthMetricLatency,
	healthMetricErrorLog,
	healthMetricBadRequests,
	healthMetricLockCost,
}

var systemHealthMetricLabels = map[string]string{
	healthMetricResource:      "资源水位",
	healthMetricResourceShift: "资源波动",
	healthMetricErrorRate:     "错误率",
	healthMetricLatency:       "响应延时",
	healthMetricErrorLog:      "错误日志",
	healthMetricBadRequests:   "活跃告警",
	healthMetricLockCost:      "锁等待",
}

var rygMetricToSystemMetric = map[string]string{
	"errorRate":     healthMetricErrorRate,
	"latency":       healthMetricLatency,
	"logErrorCount": healthMetricErrorLog,
	"alert":         healthMetricBadRequests,
}

func (h *handler) GetOverview() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetSystemHealthOverviewRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(http.StatusBadRequest, code.ParamBindError, err)
			return
		}

		resp, err := h.buildOverview(c, req)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, code.GetDataGroupError, err)
			return
		}
		if req.SummaryOnly {
			c.Payload(systemHealthOverviewSummaryOnlyResponse(resp))
			return
		}
		c.Payload(resp)
	}
}

func (h *handler) GetSystems() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.ListSystemHealthSystemsRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(http.StatusBadRequest, code.ParamBindError, err)
			return
		}

		resp, err := h.buildSystemsPage(c, req)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, code.GetDataGroupError, err)
			return
		}
		c.Payload(resp)
	}
}

func (h *handler) GetSystemServices() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.ListSystemHealthSystemServicesRequest)
		if err := c.ShouldBindURI(req); err != nil {
			c.AbortWithError(http.StatusBadRequest, code.ParamBindError, err)
			return
		}
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(http.StatusBadRequest, code.ParamBindError, err)
			return
		}

		resp, err := h.buildSystemServicesPage(c, req)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, code.GetDataGroupError, err)
			return
		}
		c.Payload(resp)
	}
}

func (h *handler) GetPendingIssues() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.ListSystemHealthPendingIssuesRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(http.StatusBadRequest, code.ParamBindError, err)
			return
		}

		resp, err := h.buildPendingIssuesPage(c, req)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, code.GetDataGroupError, err)
			return
		}
		c.Payload(resp)
	}
}

func (h *handler) GetNewAbnormalMetrics() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetSystemHealthNewAbnormalMetricsRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(http.StatusBadRequest, code.ParamBindError, err)
			return
		}

		resp, err := h.buildNewAbnormalMetrics(c, req)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, code.GetDataGroupError, err)
			return
		}
		c.Payload(resp)
	}
}

func (h *handler) GetSystemDetail() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetSystemHealthSystemDetailRequest)
		if err := c.ShouldBindURI(req); err != nil {
			c.AbortWithError(http.StatusBadRequest, code.ParamBindError, err)
			return
		}
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(http.StatusBadRequest, code.ParamBindError, err)
			return
		}

		if !shouldIncludeSystemHealthDetailServices(req.IncludeServices) {
			resp, err := h.buildSystemDetailHeader(c, req)
			if err != nil {
				c.AbortWithError(http.StatusBadRequest, code.GetDataGroupError, err)
				return
			}
			c.Payload(resp)
			return
		}

		resp, err := h.buildSystemDetail(c, req)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, code.GetDataGroupError, err)
			return
		}
		c.Payload(resp)
	}
}

func (h *handler) buildOverview(ctx core.Context, req *request.GetSystemHealthOverviewRequest) (response.SystemHealthOverviewResp, error) {
	buildStartedAt := time.Now()
	startTime, endTime, window := resolveTimeRange(&request.GetSystemHealthReportRequest{
		RelativeTime: req.RelativeTime,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
	})
	systemHealthInfo("system-health overview build started", systemHealthStageFields(ctx, buildStartedAt, "01_resolve_window",
		zap.String("window", window),
	)...)

	var currentLookup systemHealthCurrentCacheLookup
	hasCurrentLookup := false
	if req.SummaryOnly && req.SkipNewAbnormalMetrics {
		if current, lookup, ok, err := h.readOnlyCurrentHealthIfFresh(ctx, startTime, endTime, window); err != nil {
			return response.SystemHealthOverviewResp{}, err
		} else if ok {
			return response.SystemHealthOverviewResp{
				Summary:     current.Summary,
				Config:      current.Config,
				Window:      current.Window,
				GeneratedAt: current.GeneratedAt,
				Source:      current.Source,
			}, nil
		} else {
			currentLookup = lookup
			hasCurrentLookup = true
		}
	}

	var current systemHealthCurrentResult
	var groups []*datagroup.DataGroupTreeNode
	var err error
	if hasCurrentLookup {
		current, groups, err = h.buildCurrentHealthWithLookup(ctx, startTime, endTime, window, buildStartedAt, currentLookup)
	} else {
		current, groups, err = h.buildCurrentHealth(ctx, startTime, endTime, window, buildStartedAt)
	}
	if err != nil {
		return response.SystemHealthOverviewResp{}, err
	}
	summary := current.Summary
	if !req.SkipNewAbnormalMetrics {
		newAbnormalMetrics, err := h.calculateNewAbnormalMetrics(ctx, endTime, groups)
		if err != nil {
			systemHealthWarn("system-health overview new abnormal metrics failed", systemHealthStageFields(ctx, buildStartedAt, "04_new_abnormal_metrics",
				zap.Error(err),
			)...)
		} else {
			summary.NewAbnormalMetrics = newAbnormalMetrics
		}
	}

	resp := response.SystemHealthOverviewResp{
		Summary:     summary,
		Config:      current.Config,
		Window:      current.Window,
		GeneratedAt: current.GeneratedAt,
		Source:      current.Source,
	}
	if !req.SummaryOnly {
		resp.Systems = current.Systems
		if len(current.PendingIssues) > 20 {
			resp.PendingIssues = append([]response.SystemHealthPendingIssue(nil), current.PendingIssues[:20]...)
		} else {
			resp.PendingIssues = append([]response.SystemHealthPendingIssue(nil), current.PendingIssues...)
		}
	}
	systemHealthInfo("system-health overview build completed", systemHealthStageFields(ctx, buildStartedAt, "06_build_completed",
		zap.Int("systems", len(resp.Systems)),
		zap.Int("pendingIssues", len(resp.PendingIssues)),
	)...)
	return resp, nil
}

func systemHealthOverviewSummaryOnlyResponse(resp response.SystemHealthOverviewResp) response.SystemHealthOverviewSummaryResp {
	return response.SystemHealthOverviewSummaryResp{
		Summary:     resp.Summary,
		Config:      resp.Config,
		Window:      resp.Window,
		GeneratedAt: resp.GeneratedAt,
		Source:      resp.Source,
	}
}

func shouldIncludeSystemHealthDetailServices(includeServices *bool) bool {
	return includeServices == nil || *includeServices
}

func systemHealthSystemDetailHeaderResponse(resp response.SystemHealthSystemDetailResp) response.SystemHealthSystemDetailHeaderResp {
	return response.SystemHealthSystemDetailHeaderResp{
		System:      systemHealthSystemHeader(resp.System),
		Window:      resp.Window,
		GeneratedAt: resp.GeneratedAt,
		Source:      resp.Source,
	}
}

func systemHealthSystemDetailHeaderFromCurrent(current systemHealthCurrentResult, groupID int64) (response.SystemHealthSystemDetailHeaderResp, bool) {
	for _, system := range current.Systems {
		if system.GroupID != groupID {
			continue
		}
		return response.SystemHealthSystemDetailHeaderResp{
			System:      systemHealthSystemHeader(systemHealthSystemRow(system, current.Config)),
			Window:      current.Window,
			GeneratedAt: current.GeneratedAt,
			Source:      current.Source,
		}, true
	}
	return response.SystemHealthSystemDetailHeaderResp{}, false
}

func systemHealthSystemHeader(system response.SystemHealthSystem) response.SystemHealthSystemHeader {
	return response.SystemHealthSystemHeader{
		ID:                    system.ID,
		GroupID:               system.GroupID,
		Name:                  system.Name,
		Score:                 system.Score,
		Status:                system.Status,
		Metrics:               cloneStringMap(system.Metrics),
		MetricDetails:         cloneSystemHealthMetricDetails(system.MetricDetails),
		ServiceTotal:          system.ServiceTotal,
		AbnormalServiceCount:  system.AbnormalServiceCount,
		AttentionServiceCount: system.AttentionServiceCount,
		Source:                system.Source,
	}
}

type systemHealthCurrentResult struct {
	Summary       response.SystemHealthOverviewSummary
	Systems       []response.SystemHealthSystem
	PendingIssues []response.SystemHealthPendingIssue
	Config        response.SystemHealthConfigResp
	Window        string
	GeneratedAt   string
	Source        string
}

type systemHealthCurrentCache struct {
	mutex      sync.Mutex
	ttl        time.Duration
	maxEntries int
	entries    map[string]systemHealthCurrentCacheEntry
	inflight   map[string]*systemHealthCurrentCacheCall
}

type systemHealthCurrentCacheEntry struct {
	result    systemHealthCurrentResult
	createdAt time.Time
	expiresAt time.Time
}

type systemHealthCurrentCacheCall struct {
	wg     sync.WaitGroup
	result systemHealthCurrentResult
	err    error
}

type systemHealthCurrentCacheStatus string

const (
	systemHealthCurrentCacheStatusBypass systemHealthCurrentCacheStatus = "bypass"
	systemHealthCurrentCacheStatusHit    systemHealthCurrentCacheStatus = "hit"
	systemHealthCurrentCacheStatusMiss   systemHealthCurrentCacheStatus = "miss"
	systemHealthCurrentCacheStatusShared systemHealthCurrentCacheStatus = "shared"
)

type systemHealthCurrentCacheTelemetry struct {
	Status  systemHealthCurrentCacheStatus
	Entries int
}

type systemHealthCurrentCacheLookup struct {
	Key    string
	Groups []*datagroup.DataGroupTreeNode
	Config response.SystemHealthConfigResp
}

func newSystemHealthCurrentCache(ttl time.Duration, maxEntries int) *systemHealthCurrentCache {
	if ttl <= 0 {
		ttl = systemHealthCurrentCacheTTL
	}
	if maxEntries <= 0 {
		maxEntries = maxSystemHealthCurrentCacheEntries
	}
	return &systemHealthCurrentCache{
		ttl:        ttl,
		maxEntries: maxEntries,
		entries:    make(map[string]systemHealthCurrentCacheEntry),
		inflight:   make(map[string]*systemHealthCurrentCacheCall),
	}
}

func (c *systemHealthCurrentCache) getOrBuild(key string, now time.Time, build func() (systemHealthCurrentResult, error)) (systemHealthCurrentResult, error) {
	result, _, err := c.getOrBuildWithTelemetry(key, now, build)
	return result, err
}

func (c *systemHealthCurrentCache) getOrBuildWithTelemetry(key string, now time.Time, build func() (systemHealthCurrentResult, error)) (systemHealthCurrentResult, systemHealthCurrentCacheTelemetry, error) {
	if c == nil || key == "" {
		result, err := build()
		return result, systemHealthCurrentCacheTelemetry{Status: systemHealthCurrentCacheStatusBypass}, err
	}

	c.mutex.Lock()
	if entry, ok := c.entries[key]; ok && now.Before(entry.expiresAt) {
		result := cloneSystemHealthCurrentResult(entry.result)
		entries := len(c.entries)
		c.mutex.Unlock()
		return result, systemHealthCurrentCacheTelemetry{
			Status:  systemHealthCurrentCacheStatusHit,
			Entries: entries,
		}, nil
	}
	if call, ok := c.inflight[key]; ok {
		c.mutex.Unlock()
		call.wg.Wait()
		c.mutex.Lock()
		entries := len(c.entries)
		c.mutex.Unlock()
		return cloneSystemHealthCurrentResult(call.result), systemHealthCurrentCacheTelemetry{
			Status:  systemHealthCurrentCacheStatusShared,
			Entries: entries,
		}, call.err
	}
	call := &systemHealthCurrentCacheCall{}
	call.wg.Add(1)
	c.inflight[key] = call
	c.mutex.Unlock()

	result, err := build()
	call.result = cloneSystemHealthCurrentResult(result)
	call.err = err

	c.mutex.Lock()
	delete(c.inflight, key)
	if err == nil {
		c.setLocked(key, result, now)
	}
	entries := len(c.entries)
	c.mutex.Unlock()

	call.wg.Done()
	return cloneSystemHealthCurrentResult(result), systemHealthCurrentCacheTelemetry{
		Status:  systemHealthCurrentCacheStatusMiss,
		Entries: entries,
	}, err
}

// getReadOnlyIfFreshOrWait returns cached data without cloning; callers must only project from it.
func (c *systemHealthCurrentCache) getReadOnlyIfFreshOrWait(key string, now time.Time) (systemHealthCurrentResult, systemHealthCurrentCacheTelemetry, bool) {
	if c == nil || key == "" {
		return systemHealthCurrentResult{}, systemHealthCurrentCacheTelemetry{Status: systemHealthCurrentCacheStatusBypass}, false
	}

	c.mutex.Lock()
	if entry, ok := c.entries[key]; ok && now.Before(entry.expiresAt) {
		result := entry.result
		entries := len(c.entries)
		c.mutex.Unlock()
		return result, systemHealthCurrentCacheTelemetry{
			Status:  systemHealthCurrentCacheStatusHit,
			Entries: entries,
		}, true
	}
	if call, ok := c.inflight[key]; ok {
		c.mutex.Unlock()
		call.wg.Wait()
		c.mutex.Lock()
		entries := len(c.entries)
		c.mutex.Unlock()
		if call.err != nil {
			return systemHealthCurrentResult{}, systemHealthCurrentCacheTelemetry{
				Status:  systemHealthCurrentCacheStatusShared,
				Entries: entries,
			}, false
		}
		return call.result, systemHealthCurrentCacheTelemetry{
			Status:  systemHealthCurrentCacheStatusShared,
			Entries: entries,
		}, true
	}
	entries := len(c.entries)
	c.mutex.Unlock()
	return systemHealthCurrentResult{}, systemHealthCurrentCacheTelemetry{
		Status:  systemHealthCurrentCacheStatusMiss,
		Entries: entries,
	}, false
}

func (c *systemHealthCurrentCache) setLocked(key string, result systemHealthCurrentResult, now time.Time) {
	for existingKey, entry := range c.entries {
		if now.After(entry.expiresAt) {
			delete(c.entries, existingKey)
		}
	}
	if len(c.entries) >= c.maxEntries {
		var oldestKey string
		var oldestCreatedAt time.Time
		for existingKey, entry := range c.entries {
			if oldestKey == "" || entry.createdAt.Before(oldestCreatedAt) {
				oldestKey = existingKey
				oldestCreatedAt = entry.createdAt
			}
		}
		delete(c.entries, oldestKey)
	}
	c.entries[key] = systemHealthCurrentCacheEntry{
		result:    cloneSystemHealthCurrentResult(result),
		createdAt: now,
		expiresAt: now.Add(c.ttl),
	}
}

func systemHealthCurrentCacheKey(userID int64, groups []*datagroup.DataGroupTreeNode, startTime time.Time, endTime time.Time, window string, config response.SystemHealthConfigResp) string {
	config = normalizeSystemHealthScoreThresholds(config)
	cacheStartTime, cacheEndTime := systemHealthCurrentCacheRange(startTime, endTime, window)
	return strings.Join([]string{
		systemHealthCurrentCacheCalcVersion,
		strconv.FormatInt(userID, 10),
		systemHealthCurrentGroupSignature(groups),
		strings.TrimSpace(window),
		strconv.FormatInt(cacheStartTime.UnixMicro(), 10),
		strconv.FormatInt(cacheEndTime.UnixMicro(), 10),
		strconv.Itoa(config.SevereAbnormalScoreThreshold),
		strconv.Itoa(config.GeneralAbnormalScoreThreshold),
	}, "|")
}

func systemHealthCurrentCacheRange(startTime time.Time, endTime time.Time, window string) (time.Time, time.Time) {
	if strings.TrimSpace(window) == "custom" {
		return startTime, endTime
	}
	duration := endTime.Sub(startTime)
	if duration <= 0 || systemHealthCurrentCacheTTL <= 0 {
		return startTime, endTime
	}
	cacheEndTime := endTime.Truncate(systemHealthCurrentCacheTTL)
	return cacheEndTime.Add(-duration), cacheEndTime
}

func systemHealthCurrentGroupSignature(groups []*datagroup.DataGroupTreeNode) string {
	groupIDs := make([]int64, 0, len(groups))
	for _, group := range groups {
		if group == nil {
			continue
		}
		groupIDs = append(groupIDs, group.GroupID)
	}
	sort.Slice(groupIDs, func(i, j int) bool {
		return groupIDs[i] < groupIDs[j]
	})
	parts := make([]string, 0, len(groupIDs))
	for _, groupID := range groupIDs {
		parts = append(parts, strconv.FormatInt(groupID, 10))
	}
	return strings.Join(parts, ",")
}

func (h *handler) systemHealthCurrentCacheLookup(ctx core.Context, startTime time.Time, endTime time.Time, window string) (systemHealthCurrentCacheLookup, error) {
	groups, err := h.visibleSystemGroups(ctx)
	if err != nil {
		return systemHealthCurrentCacheLookup{}, err
	}
	scoreThresholds, err := h.systemHealthConfig(ctx)
	if err != nil {
		systemHealthWarnCtx(ctx, "system-health current cache lookup config failed, using default threshold", zap.Error(err))
	}
	scoreThresholds = normalizeSystemHealthScoreThresholds(scoreThresholds)
	return systemHealthCurrentCacheLookup{
		Key:    systemHealthCurrentCacheKey(ctx.UserID(), groups, startTime, endTime, window, scoreThresholds),
		Groups: groups,
		Config: scoreThresholds,
	}, nil
}

func (h *handler) readOnlyCurrentHealthIfFresh(ctx core.Context, startTime time.Time, endTime time.Time, window string) (systemHealthCurrentResult, systemHealthCurrentCacheLookup, bool, error) {
	lookup, err := h.systemHealthCurrentCacheLookup(ctx, startTime, endTime, window)
	if err != nil {
		return systemHealthCurrentResult{}, systemHealthCurrentCacheLookup{}, false, err
	}
	current, _, ok := h.currentHealthCache.getReadOnlyIfFreshOrWait(lookup.Key, time.Now())
	return current, lookup, ok, nil
}

func systemHealthVisibleGroupByID(groups []*datagroup.DataGroupTreeNode, groupID int64) *datagroup.DataGroupTreeNode {
	for _, group := range groups {
		if group != nil && group.GroupID == groupID {
			return group
		}
	}
	return nil
}

func cloneSystemHealthCurrentResult(result systemHealthCurrentResult) systemHealthCurrentResult {
	result.Systems = cloneSystemHealthSystems(result.Systems)
	if result.PendingIssues != nil {
		result.PendingIssues = append([]response.SystemHealthPendingIssue(nil), result.PendingIssues...)
	}
	return result
}

func cloneSystemHealthSystems(systems []response.SystemHealthSystem) []response.SystemHealthSystem {
	if systems == nil {
		return nil
	}
	cloned := make([]response.SystemHealthSystem, len(systems))
	for i := range systems {
		cloned[i] = cloneSystemHealthSystem(systems[i])
	}
	return cloned
}

func cloneSystemHealthSystem(system response.SystemHealthSystem) response.SystemHealthSystem {
	system.Metrics = cloneStringMap(system.Metrics)
	system.MetricDetails = cloneSystemHealthMetricDetails(system.MetricDetails)
	if system.Services != nil {
		services := system.Services
		system.Services = make([]response.SystemHealthService, len(services))
		for i := range services {
			system.Services[i] = cloneSystemHealthService(services[i])
		}
	}
	return system
}

func cloneSystemHealthService(service response.SystemHealthService) response.SystemHealthService {
	service.Metrics = cloneStringMap(service.Metrics)
	service.MetricDetails = cloneSystemHealthMetricDetails(service.MetricDetails)
	return service
}

func cloneSystemHealthServices(services []response.SystemHealthService) []response.SystemHealthService {
	if services == nil {
		return nil
	}
	cloned := make([]response.SystemHealthService, len(services))
	for i := range services {
		cloned[i] = cloneSystemHealthService(services[i])
	}
	return cloned
}

func cloneStringMap(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}
	cloned := make(map[string]string, len(input))
	for key, value := range input {
		cloned[key] = value
	}
	return cloned
}

func cloneSystemHealthMetricDetails(input map[string]response.SystemHealthMetricDetail) map[string]response.SystemHealthMetricDetail {
	if input == nil {
		return nil
	}
	cloned := make(map[string]response.SystemHealthMetricDetail, len(input))
	for key, detail := range input {
		if detail.Series != nil {
			detail.Series = append([]response.SystemHealthMetricPoint(nil), detail.Series...)
		}
		if detail.Thresholds != nil {
			detail.Thresholds = append([]response.SystemHealthMetricThreshold(nil), detail.Thresholds...)
		}
		if detail.Badges != nil {
			detail.Badges = append([]response.SystemHealthMetricBadge(nil), detail.Badges...)
		}
		cloned[key] = detail
	}
	return cloned
}

func (h *handler) buildCurrentHealth(ctx core.Context, startTime time.Time, endTime time.Time, window string, buildStartedAt time.Time) (systemHealthCurrentResult, []*datagroup.DataGroupTreeNode, error) {
	lookup, err := h.systemHealthCurrentCacheLookup(ctx, startTime, endTime, window)
	if err != nil {
		systemHealthWarn("system-health current cache lookup failed", systemHealthStageFields(ctx, buildStartedAt, "02_current_lookup",
			zap.Error(err),
		)...)
		return systemHealthCurrentResult{}, nil, err
	}
	return h.buildCurrentHealthWithLookup(ctx, startTime, endTime, window, buildStartedAt, lookup)
}

func (h *handler) buildCurrentHealthWithLookup(ctx core.Context, startTime time.Time, endTime time.Time, window string, buildStartedAt time.Time, lookup systemHealthCurrentCacheLookup) (systemHealthCurrentResult, []*datagroup.DataGroupTreeNode, error) {
	groups := lookup.Groups
	scoreThresholds := normalizeSystemHealthScoreThresholds(lookup.Config)
	cacheKey := lookup.Key
	now := time.Now()
	cacheStartedAt := time.Now()
	result, cacheTelemetry, err := h.currentHealthCache.getOrBuildWithTelemetry(cacheKey, now, func() (systemHealthCurrentResult, error) {
		stageStartedAt := time.Now()
		systems, err := h.buildOverviewSystems(ctx, startTime, endTime, groups, "current")
		if err != nil {
			systemHealthWarn("system-health current systems failed", systemHealthStageFields(ctx, stageStartedAt, "04_current_systems",
				zap.Error(err),
			)...)
			return systemHealthCurrentResult{}, err
		}
		systemHealthInfo("system-health current systems built", systemHealthStageFields(ctx, stageStartedAt, "04_current_systems",
			zap.Int("systems", len(systems)),
		)...)
		applySystemHealthScoreThresholds(systems, scoreThresholds)
		return systemHealthCurrentResult{
			Summary:       buildSystemHealthOverviewSummary(systems, scoreThresholds),
			Systems:       systems,
			PendingIssues: buildSystemHealthPendingIssues(systems, 0),
			Config:        scoreThresholds,
			Window:        window,
			GeneratedAt:   now.Format(time.RFC3339),
			Source:        systemHealthSourceDataGroup,
		}, nil
	})
	if err != nil {
		return systemHealthCurrentResult{}, nil, err
	}
	systemHealthInfo("system-health current cache lookup completed", systemHealthStageFields(ctx, cacheStartedAt, "04_current_cache",
		zap.String("cacheStatus", string(cacheTelemetry.Status)),
		zap.String("cacheKeyHash", systemHealthShortHash(cacheKey)),
		zap.Duration("cacheTTL", systemHealthCurrentCacheTTL),
		zap.Int("cacheEntries", cacheTelemetry.Entries),
		zap.String("window", window),
		zap.Int64("startTime", startTime.UnixMicro()),
		zap.Int64("endTime", endTime.UnixMicro()),
		zap.Int("visibleGroups", len(groups)),
		zap.Int("systems", len(result.Systems)),
		zap.Int("services", result.Summary.Services),
		zap.Int("pendingIssues", len(result.PendingIssues)),
		zap.String("generatedAt", result.GeneratedAt),
	)...)
	return result, groups, nil
}

func (h *handler) buildSystemsPage(ctx core.Context, req *request.ListSystemHealthSystemsRequest) (response.SystemHealthSystemsResp, error) {
	startTime, endTime, window := resolveTimeRange(&request.GetSystemHealthReportRequest{
		RelativeTime: req.RelativeTime,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
	})

	current, lookup, ok, err := h.readOnlyCurrentHealthIfFresh(ctx, startTime, endTime, window)
	if err != nil {
		return response.SystemHealthSystemsResp{}, err
	}
	if ok {
		return buildSystemHealthSystemsPage(current.Systems, systemHealthListOptions{
			Page:     req.Page,
			PageSize: req.PageSize,
			Keyword:  req.Keyword,
			Status:   req.Status,
			Sort:     req.Sort,
		}, current.Config, current.Window, current.GeneratedAt), nil
	}

	current, _, err = h.buildCurrentHealthWithLookup(ctx, startTime, endTime, window, time.Now(), lookup)
	if err != nil {
		return response.SystemHealthSystemsResp{}, err
	}

	return buildSystemHealthSystemsPage(current.Systems, systemHealthListOptions{
		Page:     req.Page,
		PageSize: req.PageSize,
		Keyword:  req.Keyword,
		Status:   req.Status,
		Sort:     req.Sort,
	}, current.Config, current.Window, current.GeneratedAt), nil
}

func (h *handler) buildSystemServicesPage(ctx core.Context, req *request.ListSystemHealthSystemServicesRequest) (response.SystemHealthServicesResp, error) {
	startTime, endTime, window := resolveTimeRange(&request.GetSystemHealthReportRequest{
		RelativeTime: req.RelativeTime,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
	})
	current, lookup, ok, err := h.readOnlyCurrentHealthIfFresh(ctx, startTime, endTime, window)
	if err != nil {
		return response.SystemHealthServicesResp{}, err
	}
	if ok {
		if resp, ok := buildSystemHealthServicesPageFromCurrent(current, req.GroupID, systemHealthListOptions{
			Page:     req.Page,
			PageSize: req.PageSize,
			Keyword:  req.Keyword,
			Status:   req.Status,
			Sort:     req.Sort,
		}); ok {
			return resp, nil
		}
	}

	matched := systemHealthVisibleGroupByID(lookup.Groups, req.GroupID)
	if matched == nil {
		return response.SystemHealthServicesResp{}, fmt.Errorf("system group %d not found or no permission", req.GroupID)
	}
	detail, err := h.buildSystemDetailForGroup(ctx, matched, startTime, endTime, window, lookup.Config)
	if err != nil {
		return response.SystemHealthServicesResp{}, err
	}
	return buildSystemHealthServicesPage(detail.System, systemHealthListOptions{
		Page:     req.Page,
		PageSize: req.PageSize,
		Keyword:  req.Keyword,
		Status:   req.Status,
		Sort:     req.Sort,
	}, lookup.Config, detail.Window, detail.GeneratedAt), nil
}

func (h *handler) buildSystemDetailHeader(ctx core.Context, req *request.GetSystemHealthSystemDetailRequest) (response.SystemHealthSystemDetailHeaderResp, error) {
	startTime, endTime, window := resolveTimeRange(&request.GetSystemHealthReportRequest{
		RelativeTime: req.RelativeTime,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
	})
	current, lookup, ok, err := h.readOnlyCurrentHealthIfFresh(ctx, startTime, endTime, window)
	if err != nil {
		return response.SystemHealthSystemDetailHeaderResp{}, err
	}
	if ok {
		if resp, ok := systemHealthSystemDetailHeaderFromCurrent(current, req.GroupID); ok {
			return resp, nil
		}
		return response.SystemHealthSystemDetailHeaderResp{}, fmt.Errorf("system group %d not found or no permission", req.GroupID)
	}

	current, _, err = h.buildCurrentHealthWithLookup(ctx, startTime, endTime, window, time.Now(), lookup)
	if err != nil {
		return response.SystemHealthSystemDetailHeaderResp{}, err
	}
	if resp, ok := systemHealthSystemDetailHeaderFromCurrent(current, req.GroupID); ok {
		return resp, nil
	}
	return response.SystemHealthSystemDetailHeaderResp{}, fmt.Errorf("system group %d not found or no permission", req.GroupID)
}

func (h *handler) buildPendingIssuesPage(ctx core.Context, req *request.ListSystemHealthPendingIssuesRequest) (response.SystemHealthPendingIssuesResp, error) {
	startTime, endTime, window := resolveTimeRange(&request.GetSystemHealthReportRequest{
		RelativeTime: req.RelativeTime,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
	})

	current, lookup, ok, err := h.readOnlyCurrentHealthIfFresh(ctx, startTime, endTime, window)
	if err != nil {
		return response.SystemHealthPendingIssuesResp{}, err
	}
	if ok {
		return buildSystemHealthPendingIssuesPageFromIssues(current.PendingIssues, systemHealthPendingIssueListOptions{
			Page:     req.Page,
			PageSize: req.PageSize,
			Keyword:  req.Keyword,
			Status:   req.Status,
			GroupID:  req.GroupID,
		}, current.Window, current.GeneratedAt), nil
	}

	current, _, err = h.buildCurrentHealthWithLookup(ctx, startTime, endTime, window, time.Now(), lookup)
	if err != nil {
		return response.SystemHealthPendingIssuesResp{}, err
	}

	return buildSystemHealthPendingIssuesPageFromIssues(current.PendingIssues, systemHealthPendingIssueListOptions{
		Page:     req.Page,
		PageSize: req.PageSize,
		Keyword:  req.Keyword,
		Status:   req.Status,
		GroupID:  req.GroupID,
	}, current.Window, current.GeneratedAt), nil
}

func (h *handler) buildNewAbnormalMetrics(ctx core.Context, req *request.GetSystemHealthNewAbnormalMetricsRequest) (response.SystemHealthNewAbnormalMetricsResp, error) {
	_, endTime, window := resolveTimeRange(&request.GetSystemHealthReportRequest{
		RelativeTime: req.RelativeTime,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
	})

	groups, err := h.visibleSystemGroups(ctx)
	if err != nil {
		return response.SystemHealthNewAbnormalMetricsResp{}, err
	}
	newAbnormalMetrics, err := h.calculateNewAbnormalMetrics(ctx, endTime, groups)
	if err != nil {
		return response.SystemHealthNewAbnormalMetricsResp{}, err
	}
	return response.SystemHealthNewAbnormalMetricsResp{
		NewAbnormalMetrics: newAbnormalMetrics,
		Window:             window,
		GeneratedAt:        time.Now().Format(time.RFC3339),
		Source:             systemHealthSourceDataGroup,
	}, nil
}

func (h *handler) calculateNewAbnormalMetrics(ctx core.Context, endTime time.Time, groups []*datagroup.DataGroupTreeNode) (int, error) {
	recentStartTime, previousStartTime, previousEndTime := newAbnormalMetricWindows(endTime)

	stageStartedAt := time.Now()
	recentSystems, err := h.buildOverviewSystems(ctx, recentStartTime, endTime, groups, "recent")
	if err != nil {
		systemHealthWarn("system-health overview recent systems failed", systemHealthStageFields(ctx, stageStartedAt, "04_recent_systems",
			zap.Error(err),
		)...)
		return 0, err
	}
	systemHealthInfo("system-health overview recent systems built", systemHealthStageFields(ctx, stageStartedAt, "04_recent_systems",
		zap.Int("systems", len(recentSystems)),
	)...)

	stageStartedAt = time.Now()
	previousSystems, err := h.buildOverviewSystems(ctx, previousStartTime, previousEndTime, groups, "previous")
	if err != nil {
		systemHealthWarn("system-health overview previous systems failed", systemHealthStageFields(ctx, stageStartedAt, "05_previous_systems",
			zap.Error(err),
		)...)
		return 0, err
	}
	systemHealthInfo("system-health overview previous systems built", systemHealthStageFields(ctx, stageStartedAt, "05_previous_systems",
		zap.Int("systems", len(previousSystems)),
	)...)

	return countNewAbnormalMetrics(recentSystems, previousSystems), nil
}

func (h *handler) buildOverviewSystems(ctx core.Context, startTime time.Time, endTime time.Time, groups []*datagroup.DataGroupTreeNode, windowLabel string) ([]response.SystemHealthSystem, error) {
	rygService := serviceoverview.New(zap.NewNop(), h.chRepo, h.dbRepo, h.promRepo)
	workerCtx := newSystemHealthOverviewWorkerContext(ctx)
	extraMetricTimeout := systemHealthOverviewExtraMetricTimeout(windowLabel)
	return buildOverviewSystemsForGroups(groups, systemHealthOverviewGroupConcurrency, func(group *datagroup.DataGroupTreeNode) (response.SystemHealthSystem, error) {
		stageStartedAt := time.Now()
		serviceScope, err := h.systemHealthServiceScope(workerCtx, group.GroupID)
		if err != nil {
			systemHealthWarn("system-health overview group service scope failed", systemHealthStageFields(workerCtx, stageStartedAt, "build_group_service_scope",
				zap.String("windowLabel", windowLabel),
				zap.Int64("groupId", group.GroupID),
				zap.String("groupName", group.GroupName),
				zap.Error(err),
			)...)
			return response.SystemHealthSystem{}, err
		}

		stageStartedAt = time.Now()
		light, err := rygService.GetServicesRYGLightStatus(workerCtx, &request.GetRygLightRequest{
			StartTime: startTime.UnixMicro(),
			EndTime:   endTime.UnixMicro(),
			GroupID:   group.GroupID,
		})
		if err != nil {
			systemHealthWarn("system-health overview group ryg failed", systemHealthStageFields(workerCtx, stageStartedAt, "build_group_ryg",
				zap.String("windowLabel", windowLabel),
				zap.Int64("groupId", group.GroupID),
				zap.String("groupName", group.GroupName),
				zap.Error(err),
			)...)
			light.ServiceList = nil
		} else {
			systemHealthInfo("system-health overview group ryg built", systemHealthStageFields(workerCtx, stageStartedAt, "build_group_ryg",
				zap.String("windowLabel", windowLabel),
				zap.Int64("groupId", group.GroupID),
				zap.String("groupName", group.GroupName),
				zap.Int("services", len(light.ServiceList)),
			)...)
		}

		system := buildSystemHealthSystemWithServiceScope(group.DataGroup, serviceScope, light.ServiceList)
		stageStartedAt = time.Now()
		extraCtx := workerCtx
		cancel := func() {}
		if extraMetricTimeout > 0 {
			timeoutCtx, timeoutCancel := context.WithTimeout(workerCtx.GetContext(), extraMetricTimeout)
			extraCtx = workerCtx.withRequestContext(timeoutCtx)
			cancel = timeoutCancel
		}
		extraMetrics, err := h.querySystemHealthExtraMetrics(extraCtx, startTime, endTime, group.GroupID, system.Services)
		cancel()
		if err == nil {
			applySystemHealthExtraMetrics(&system, extraMetrics)
			systemHealthInfo("system-health overview group extra metrics applied", systemHealthStageFields(workerCtx, stageStartedAt, "build_group_extra_metrics",
				zap.String("windowLabel", windowLabel),
				zap.Int64("groupId", group.GroupID),
				zap.String("groupName", group.GroupName),
			)...)
		} else {
			systemHealthWarn("system-health overview group extra metrics skipped", systemHealthStageFields(workerCtx, stageStartedAt, "build_group_extra_metrics",
				zap.String("windowLabel", windowLabel),
				zap.Int64("groupId", group.GroupID),
				zap.String("groupName", group.GroupName),
				zap.Error(err),
			)...)
		}
		return system, nil
	})
}

func systemHealthOverviewExtraMetricTimeout(windowLabel string) time.Duration {
	if windowLabel == "current" {
		return systemHealthOverviewCurrentTimeout
	}
	return systemHealthOverviewHistoryTimeout
}

type overviewSystemBuilder func(group *datagroup.DataGroupTreeNode) (response.SystemHealthSystem, error)

func buildOverviewSystemsForGroups(groups []*datagroup.DataGroupTreeNode, concurrency int, build overviewSystemBuilder) ([]response.SystemHealthSystem, error) {
	systems := make([]response.SystemHealthSystem, len(groups))
	if len(groups) == 0 {
		return systems, nil
	}
	if concurrency <= 0 {
		concurrency = 1
	}
	if concurrency > len(groups) {
		concurrency = len(groups)
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)
	for index, group := range groups {
		index, group := index, group
		sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				<-sem
			}()

			system, err := build(group)
			if err != nil {
				systems[index] = buildSystemHealthSystem(group.DataGroup, nil)
				return
			}
			systems[index] = system
		}()
	}
	wg.Wait()
	sort.SliceStable(systems, func(i, j int) bool {
		if systems[i].Score == systems[j].Score {
			return systems[i].GroupID < systems[j].GroupID
		}
		return systems[i].Score < systems[j].Score
	})

	return systems, nil
}

type systemHealthOverviewWorkerContext struct {
	core.Context
	requestContext context.Context
	requestID      string
	routeMethod    string
	routePath      string
	lang           string
	userID         int64
}

func newSystemHealthOverviewWorkerContext(ctx core.Context) systemHealthOverviewWorkerContext {
	requestContext := context.Background()
	requestID := ""
	routeMethod := ""
	routePath := ""
	lang := code.LANG_ZH
	var userID int64
	if ctx != nil {
		if request := ctx.Request(); request != nil && request.Context() != nil {
			requestContext = request.Context()
		}
		requestID = systemHealthRequestID(ctx)
		routeMethod, routePath = ctx.GetMethodPath()
		lang = ctx.LANG()
		userID = ctx.UserID()
	}
	return systemHealthOverviewWorkerContext{
		Context:        ctx,
		requestContext: requestContext,
		requestID:      requestID,
		routeMethod:    routeMethod,
		routePath:      routePath,
		lang:           lang,
		userID:         userID,
	}
}

func (c systemHealthOverviewWorkerContext) withRequestContext(requestContext context.Context) systemHealthOverviewWorkerContext {
	c.requestContext = requestContext
	return c
}

func (c systemHealthOverviewWorkerContext) GetContext() context.Context {
	if c.requestContext == nil {
		return context.Background()
	}
	return c.requestContext
}

func (c systemHealthOverviewWorkerContext) Get(key string) (any, bool) {
	if key == core.RequestIDKey && c.requestID != "" {
		return c.requestID, true
	}
	return nil, false
}

func (c systemHealthOverviewWorkerContext) GetMethodPath() (method string, path string) {
	return c.routeMethod, c.routePath
}

func (c systemHealthOverviewWorkerContext) LANG() string {
	if c.lang == "" {
		return code.LANG_ZH
	}
	return c.lang
}

func (c systemHealthOverviewWorkerContext) UserID() int64 {
	return c.userID
}

func (c systemHealthOverviewWorkerContext) Done() <-chan struct{} {
	return c.GetContext().Done()
}

func (c systemHealthOverviewWorkerContext) Deadline() (deadline time.Time, ok bool) {
	return c.GetContext().Deadline()
}

func (c systemHealthOverviewWorkerContext) Err() error {
	return c.GetContext().Err()
}

func (c systemHealthOverviewWorkerContext) Value(key any) any {
	if stringKey, ok := key.(string); ok {
		if value, exists := c.Get(stringKey); exists {
			return value
		}
	}
	return c.GetContext().Value(key)
}

func (h *handler) buildSystemDetail(ctx core.Context, req *request.GetSystemHealthSystemDetailRequest) (response.SystemHealthSystemDetailResp, error) {
	startTime, endTime, window := resolveTimeRange(&request.GetSystemHealthReportRequest{
		RelativeTime: req.RelativeTime,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
	})

	groups, err := h.visibleSystemGroups(ctx)
	if err != nil {
		return response.SystemHealthSystemDetailResp{}, err
	}
	matched := systemHealthVisibleGroupByID(groups, req.GroupID)
	if matched == nil {
		return response.SystemHealthSystemDetailResp{}, fmt.Errorf("system group %d not found or no permission", req.GroupID)
	}
	scoreThresholds, err := h.systemHealthConfig(ctx)
	if err != nil {
		systemHealthWarnCtx(ctx, "system-health detail config failed, using default threshold", zap.Error(err))
	}
	return h.buildSystemDetailForGroup(ctx, matched, startTime, endTime, window, scoreThresholds)
}

func (h *handler) buildSystemDetailForGroup(ctx core.Context, matched *datagroup.DataGroupTreeNode, startTime time.Time, endTime time.Time, window string, scoreThresholds response.SystemHealthConfigResp) (response.SystemHealthSystemDetailResp, error) {
	if matched == nil {
		return response.SystemHealthSystemDetailResp{}, fmt.Errorf("system group not found or no permission")
	}
	rygService := serviceoverview.New(zap.NewNop(), h.chRepo, h.dbRepo, h.promRepo)
	light, err := rygService.GetServicesRYGLightStatus(ctx, &request.GetRygLightRequest{
		StartTime: startTime.UnixMicro(),
		EndTime:   endTime.UnixMicro(),
		GroupID:   matched.GroupID,
	})
	if err != nil {
		systemHealthWarnCtx(ctx, "system-health detail ryg failed", zap.Int64("groupId", matched.GroupID), zap.Error(err))
		light.ServiceList = nil
	}

	serviceScope, err := h.systemHealthServiceScope(ctx, matched.GroupID)
	if err != nil {
		return response.SystemHealthSystemDetailResp{}, err
	}
	system := buildSystemHealthSystemWithServiceScope(matched.DataGroup, serviceScope, light.ServiceList)
	extraMetrics, err := h.querySystemHealthExtraMetrics(ctx, startTime, endTime, matched.GroupID, system.Services)
	if err == nil {
		applySystemHealthExtraMetrics(&system, extraMetrics)
	}
	applySystemHealthScoreThreshold(&system, normalizeSystemHealthScoreThresholds(scoreThresholds))

	return response.SystemHealthSystemDetailResp{
		System:      system,
		Window:      window,
		GeneratedAt: time.Now().Format(time.RFC3339),
		Source:      systemHealthSourceDataGroup,
	}, nil
}

func (h *handler) visibleSystemGroups(ctx core.Context) ([]*datagroup.DataGroupTreeNode, error) {
	permGroupIDs, err := h.dbRepo.GetDataGroupIDsByUserId(ctx, ctx.UserID())
	if err != nil {
		return nil, err
	}

	root := common.DataGroupStorage
	if root == nil || root.DataGroupTreeNode == nil {
		rootNode, err := h.dbRepo.LoadDataGroupTree(ctx)
		if err != nil {
			return nil, err
		}
		return editableFirstLevelSystemGroups(rootNode.CloneWithPermission(permGroupIDs)), nil
	}
	return editableFirstLevelSystemGroups(root.CloneWithPermission(permGroupIDs)), nil
}

func firstLevelSystemGroups(root *datagroup.DataGroupTreeNode) []*datagroup.DataGroupTreeNode {
	if root == nil {
		return []*datagroup.DataGroupTreeNode{}
	}
	groups := append([]*datagroup.DataGroupTreeNode{}, root.SubGroups...)
	sort.SliceStable(groups, func(i, j int) bool {
		return groups[i].GroupID < groups[j].GroupID
	})
	return groups
}

func editableFirstLevelSystemGroups(root *datagroup.DataGroupTreeNode) []*datagroup.DataGroupTreeNode {
	groups := firstLevelSystemGroups(root)
	result := make([]*datagroup.DataGroupTreeNode, 0, len(groups))
	for _, group := range groups {
		if group.PermissionType == datagroup.DATA_GROUP_PERMISSION_TYPE_VIEW ||
			group.PermissionType == datagroup.DATA_GROUP_PERMISSION_TYPE_EDIT {
			result = append(result, group)
		}
	}
	return result
}

func (h *handler) systemHealthServiceScope(ctx core.Context, groupID int64) ([]string, error) {
	selected := []string{"APO_ALL_DATA"}
	if groupID != 0 {
		var err error
		selected, err = h.dbRepo.GetScopeIDsSelectedByGroupID(ctx, groupID)
		if err != nil {
			return nil, err
		}
	}
	if len(selected) == 0 {
		return []string{}, nil
	}

	var scopeTree *datagroup.DataScopeTree
	if common.DataGroupStorage != nil && common.DataGroupStorage.DataScopeTree != nil {
		scopeTree = common.DataGroupStorage.DataScopeTree
	} else {
		var err error
		scopeTree, err = h.dbRepo.LoadScopes(ctx)
		if err != nil {
			return nil, err
		}
	}
	if scopeTree == nil {
		return []string{}, nil
	}

	_, leafs := scopeTree.CloneWithCategory(selected, datagroup.DATASOURCE_CATEGORY_APM)
	serviceSet := make(map[string]struct{}, len(leafs))
	for labels := range leafs {
		if labels.Service == "" {
			continue
		}
		serviceSet[labels.Service] = struct{}{}
	}
	services := make([]string, 0, len(serviceSet))
	for service := range serviceSet {
		services = append(services, service)
	}
	sort.Strings(services)
	return services, nil
}

func buildSystemHealthSystem(group datagroup.DataGroup, services []*response.ServiceRYGResult) response.SystemHealthSystem {
	return buildSystemHealthSystemWithServiceScope(group, nil, services)
}

func buildSystemHealthSystemWithServiceScope(group datagroup.DataGroup, serviceScope []string, services []*response.ServiceRYGResult) response.SystemHealthSystem {
	system := response.SystemHealthSystem{
		ID:            group.GroupID,
		GroupID:       group.GroupID,
		Name:          group.GroupName,
		Status:        healthStatusGray,
		Metrics:       defaultSystemMetricStatus(),
		MetricDetails: defaultSystemMetricDetails(),
		Services:      make([]response.SystemHealthService, 0, len(services)),
		Source:        systemHealthSourceDataGroup,
	}

	serviceNames := normalizedSystemHealthServiceScope(serviceScope)
	if len(serviceNames) == 0 {
		serviceNames = serviceNamesFromRYGResults(services)
	}
	if len(serviceNames) == 0 {
		return system
	}

	rygByService := map[string]*response.ServiceRYGResult{}
	for _, service := range services {
		if service == nil || service.ServiceName == "" {
			continue
		}
		rygByService[service.ServiceName] = service
	}

	for _, serviceName := range serviceNames {
		if rygService, ok := rygByService[serviceName]; ok {
			system.Services = append(system.Services, buildSystemHealthService(rygService))
		} else {
			system.Services = append(system.Services, buildNoDataSystemHealthService(serviceName))
		}
	}

	recalculateSystemHealthSystem(&system)
	sortSystemHealthServicesByScore(system.Services)

	return system
}

func normalizedSystemHealthServiceScope(serviceScope []string) []string {
	serviceSet := make(map[string]struct{}, len(serviceScope))
	for _, serviceName := range serviceScope {
		if serviceName == "" {
			continue
		}
		serviceSet[serviceName] = struct{}{}
	}
	services := make([]string, 0, len(serviceSet))
	for serviceName := range serviceSet {
		services = append(services, serviceName)
	}
	sort.Strings(services)
	return services
}

func serviceNamesFromRYGResults(services []*response.ServiceRYGResult) []string {
	serviceNames := make([]string, 0, len(services))
	for _, service := range services {
		if service == nil || service.ServiceName == "" {
			continue
		}
		serviceNames = append(serviceNames, service.ServiceName)
	}
	return normalizedSystemHealthServiceScope(serviceNames)
}

func buildSystemHealthService(service *response.ServiceRYGResult) response.SystemHealthService {
	score := normalizeHealthScore(service.PercentScore)
	result := response.SystemHealthService{
		Name:          service.ServiceName,
		Score:         score,
		Status:        scoreHealthStatus(score),
		Metrics:       defaultSystemMetricStatus(),
		MetricDetails: defaultSystemMetricDetails(),
	}

	for _, detail := range service.ScoreDetail {
		metricKey, ok := rygMetricToSystemMetric[detail.Key]
		if !ok {
			continue
		}
		status := rygScoreHealthStatus(detail.Score)
		result.Metrics[metricKey] = worseHealthStatus(result.Metrics[metricKey], status)
		result.MetricDetails[metricKey] = response.SystemHealthMetricDetail{
			Status: result.Metrics[metricKey],
			Value:  metricStatusText(result.Metrics[metricKey]),
			Note:   detail.Detail,
		}
	}

	return result
}

func buildNoDataSystemHealthService(serviceName string) response.SystemHealthService {
	return response.SystemHealthService{
		Name:          serviceName,
		Score:         0,
		Status:        healthStatusGray,
		Metrics:       defaultSystemMetricStatus(),
		MetricDetails: defaultSystemMetricDetails(),
	}
}

func sortSystemHealthServicesByScore(services []response.SystemHealthService) {
	sort.SliceStable(services, func(i, j int) bool {
		leftNoData := isNoDataSystemHealthService(services[i])
		rightNoData := isNoDataSystemHealthService(services[j])
		if leftNoData != rightNoData {
			return !leftNoData
		}
		return services[i].Score < services[j].Score
	})
}

func hasScoredSystemHealthServices(services []response.SystemHealthService) bool {
	for _, service := range services {
		if !isNoDataSystemHealthService(service) {
			return true
		}
	}
	return false
}

func isNoDataSystemHealthService(service response.SystemHealthService) bool {
	if service.Status == healthStatusGray {
		return true
	}
	if len(service.Metrics) == 0 {
		return false
	}
	for _, key := range systemHealthMetricKeys {
		status := service.Metrics[key]
		if status != "" && status != healthStatusGray {
			return false
		}
	}
	return true
}

func defaultSystemMetricStatus() map[string]string {
	metrics := make(map[string]string, len(systemHealthMetricKeys))
	for _, key := range systemHealthMetricKeys {
		metrics[key] = healthStatusGray
	}
	return metrics
}

func defaultSystemMetricDetails() map[string]response.SystemHealthMetricDetail {
	details := make(map[string]response.SystemHealthMetricDetail, len(systemHealthMetricKeys))
	for _, key := range systemHealthMetricKeys {
		details[key] = response.SystemHealthMetricDetail{
			Status: healthStatusGray,
			Value:  "--",
			Note:   "暂无数据",
		}
	}
	return details
}

func newAbnormalMetricWindows(endTime time.Time) (time.Time, time.Time, time.Time) {
	currentStart := endTime.Add(-10 * time.Minute)
	previousStart := currentStart.Add(-10 * time.Minute)
	return currentStart, previousStart, currentStart
}

func systemHealthTrendStep(startTime time.Time, endTime time.Time) time.Duration {
	duration := endTime.Sub(startTime)
	if duration <= 0 {
		return time.Minute
	}
	step := duration / 30
	if step < time.Minute {
		return time.Minute
	}
	if step > time.Hour {
		return time.Hour
	}
	return step
}

type systemHealthListOptions struct {
	Page     int
	PageSize int
	Keyword  string
	Status   string
	Sort     string
}

type systemHealthPendingIssueListOptions struct {
	Page     int
	PageSize int
	Keyword  string
	Status   string
	GroupID  int64
}

func buildSystemHealthSystemsPage(
	systems []response.SystemHealthSystem,
	options systemHealthListOptions,
	scoreThresholds response.SystemHealthConfigResp,
	window string,
	generatedAt string,
) response.SystemHealthSystemsResp {
	page, pageSize := normalizeSystemHealthPage(options.Page, options.PageSize, systemHealthDefaultSystemPageSize)
	scoreThresholds = normalizeSystemHealthScoreThresholds(scoreThresholds)
	filtered := make([]response.SystemHealthSystem, 0, len(systems))
	for _, system := range systems {
		if !systemHealthKeywordMatches(options.Keyword, system.Name) {
			continue
		}
		if !systemHealthTargetMatchesStatus(system.Status, system.Score, options.Status, scoreThresholds) {
			continue
		}
		row := systemHealthSystemRow(system, scoreThresholds)
		filtered = append(filtered, row)
	}
	sortSystemHealthSystemRows(filtered, options.Sort)
	total := len(filtered)
	return response.SystemHealthSystemsResp{
		Systems:     cloneSystemHealthSystems(pageSystemHealthSystems(filtered, page, pageSize)),
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
		Window:      window,
		GeneratedAt: generatedAt,
		Source:      systemHealthSourceDataGroup,
	}
}

func buildSystemHealthServicesPage(
	system response.SystemHealthSystem,
	options systemHealthListOptions,
	scoreThresholds response.SystemHealthConfigResp,
	window string,
	generatedAt string,
) response.SystemHealthServicesResp {
	page, pageSize := normalizeSystemHealthPage(options.Page, options.PageSize, systemHealthDefaultServicePageSize)
	scoreThresholds = normalizeSystemHealthScoreThresholds(scoreThresholds)
	filtered := make([]response.SystemHealthService, 0, len(system.Services))
	for _, service := range system.Services {
		if !systemHealthKeywordMatches(options.Keyword, service.Name, service.Desc) {
			continue
		}
		if !systemHealthTargetMatchesStatus(service.Status, service.Score, options.Status, scoreThresholds) {
			continue
		}
		filtered = append(filtered, service)
	}
	sortSystemHealthServices(filtered, options.Sort)
	total := len(filtered)
	return response.SystemHealthServicesResp{
		Services:    cloneSystemHealthServices(pageSystemHealthServices(filtered, page, pageSize)),
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
		Window:      window,
		GeneratedAt: generatedAt,
		Source:      systemHealthSourceDataGroup,
	}
}

func buildSystemHealthServicesPageFromCurrent(
	current systemHealthCurrentResult,
	groupID int64,
	options systemHealthListOptions,
) (response.SystemHealthServicesResp, bool) {
	for _, system := range current.Systems {
		if system.GroupID != groupID {
			continue
		}
		return buildSystemHealthServicesPage(system, options, current.Config, current.Window, current.GeneratedAt), true
	}
	return response.SystemHealthServicesResp{}, false
}

func buildSystemHealthPendingIssuesPage(
	systems []response.SystemHealthSystem,
	options systemHealthPendingIssueListOptions,
	window string,
	generatedAt string,
) response.SystemHealthPendingIssuesResp {
	return buildSystemHealthPendingIssuesPageFromIssues(buildSystemHealthPendingIssues(systems, 0), options, window, generatedAt)
}

func buildSystemHealthPendingIssuesPageFromIssues(
	issues []response.SystemHealthPendingIssue,
	options systemHealthPendingIssueListOptions,
	window string,
	generatedAt string,
) response.SystemHealthPendingIssuesResp {
	page, pageSize := normalizeSystemHealthPage(options.Page, options.PageSize, systemHealthDefaultIssuePageSize)
	filtered := make([]response.SystemHealthPendingIssue, 0, len(issues))
	for _, issue := range issues {
		if options.GroupID != 0 && issue.GroupID != options.GroupID {
			continue
		}
		if !systemHealthPendingIssueMatchesStatus(issue, options.Status) {
			continue
		}
		if !systemHealthKeywordMatches(options.Keyword, issue.SystemName, issue.ServiceName, issue.MetricLabel, issue.Note) {
			continue
		}
		filtered = append(filtered, issue)
	}
	total := len(filtered)
	return response.SystemHealthPendingIssuesResp{
		Issues:      pageSystemHealthPendingIssues(filtered, page, pageSize),
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
		Window:      window,
		GeneratedAt: generatedAt,
		Source:      systemHealthSourceDataGroup,
	}
}

func normalizeSystemHealthPage(page int, pageSize int, defaultPageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > systemHealthMaxPageSize {
		pageSize = systemHealthMaxPageSize
	}
	return page, pageSize
}

func systemHealthKeywordMatches(keyword string, values ...string) bool {
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	if keyword == "" {
		return true
	}
	for _, value := range values {
		if strings.Contains(strings.ToLower(value), keyword) {
			return true
		}
	}
	return false
}

func systemHealthTargetMatchesStatus(status string, score int, filter string, thresholds response.SystemHealthConfigResp) bool {
	filter = normalizeSystemHealthListStatus(filter)
	if filter == "" || filter == "all" {
		return true
	}
	if status == "" {
		status = scoreHealthStatus(score, thresholds)
	}
	return status == filter
}

func systemHealthPendingIssueMatchesStatus(issue response.SystemHealthPendingIssue, filter string) bool {
	filter = normalizeSystemHealthListStatus(filter)
	if filter == "" || filter == "all" {
		return true
	}
	return issue.Status == filter
}

func normalizeSystemHealthListStatus(status string) string {
	switch strings.TrimSpace(status) {
	case "", "all":
		return "all"
	case "abnormal", healthStatusRed:
		return healthStatusRed
	case "attention", healthStatusYellow:
		return healthStatusYellow
	case "healthy", healthStatusGreen:
		return healthStatusGreen
	case "noData", "nodata", healthStatusGray:
		return healthStatusGray
	default:
		return "all"
	}
}

func systemHealthSystemRow(system response.SystemHealthSystem, thresholds response.SystemHealthConfigResp) response.SystemHealthSystem {
	populateSystemHealthServiceCounts(&system, thresholds)
	system.Services = nil
	return system
}

func isNoDataSystemHealthSystem(system response.SystemHealthSystem) bool {
	return system.Status == healthStatusGray
}

func hasSystemHealthAggregateMetricStatus(metrics map[string]string) bool {
	_, ok := scoreFromSystemMetricStatuses(metrics)
	return ok
}

func populateSystemHealthServiceCounts(system *response.SystemHealthSystem, thresholds response.SystemHealthConfigResp) {
	if system == nil {
		return
	}
	thresholds = normalizeSystemHealthScoreThresholds(thresholds)
	system.ServiceTotal = len(system.Services)
	system.AbnormalServiceCount = 0
	system.AttentionServiceCount = 0
	for _, service := range system.Services {
		if isNoDataSystemHealthService(service) {
			continue
		}
		score := normalizeHealthScore(service.Score)
		if score < thresholds.SevereAbnormalScoreThreshold {
			system.AbnormalServiceCount++
		} else if score < thresholds.GeneralAbnormalScoreThreshold {
			system.AttentionServiceCount++
		}
	}
}

func sortSystemHealthSystemRows(systems []response.SystemHealthSystem, sortKey string) {
	sort.SliceStable(systems, func(i, j int) bool {
		switch sortKey {
		case "score_desc":
			leftNoData := isNoDataSystemHealthSystem(systems[i])
			rightNoData := isNoDataSystemHealthSystem(systems[j])
			if leftNoData != rightNoData {
				return !leftNoData
			}
			if systems[i].Score == systems[j].Score {
				return systems[i].GroupID < systems[j].GroupID
			}
			return systems[i].Score > systems[j].Score
		case "name_asc":
			return systems[i].Name < systems[j].Name
		default:
			leftNoData := isNoDataSystemHealthSystem(systems[i])
			rightNoData := isNoDataSystemHealthSystem(systems[j])
			if leftNoData != rightNoData {
				return !leftNoData
			}
			if systems[i].Score == systems[j].Score {
				return systems[i].GroupID < systems[j].GroupID
			}
			return systems[i].Score < systems[j].Score
		}
	})
}

func sortSystemHealthServices(services []response.SystemHealthService, sortKey string) {
	sort.SliceStable(services, func(i, j int) bool {
		switch sortKey {
		case "score_desc":
			return services[i].Score > services[j].Score
		case "name_asc":
			return services[i].Name < services[j].Name
		default:
			leftNoData := isNoDataSystemHealthService(services[i])
			rightNoData := isNoDataSystemHealthService(services[j])
			if leftNoData != rightNoData {
				return !leftNoData
			}
			if services[i].Score == services[j].Score {
				return services[i].Name < services[j].Name
			}
			return services[i].Score < services[j].Score
		}
	})
}

func pageSystemHealthSystems(systems []response.SystemHealthSystem, page int, pageSize int) []response.SystemHealthSystem {
	start := (page - 1) * pageSize
	if start >= len(systems) {
		return []response.SystemHealthSystem{}
	}
	end := start + pageSize
	if end > len(systems) {
		end = len(systems)
	}
	return systems[start:end]
}

func pageSystemHealthServices(services []response.SystemHealthService, page int, pageSize int) []response.SystemHealthService {
	start := (page - 1) * pageSize
	if start >= len(services) {
		return []response.SystemHealthService{}
	}
	end := start + pageSize
	if end > len(services) {
		end = len(services)
	}
	return services[start:end]
}

func pageSystemHealthPendingIssues(issues []response.SystemHealthPendingIssue, page int, pageSize int) []response.SystemHealthPendingIssue {
	start := (page - 1) * pageSize
	if start >= len(issues) {
		return []response.SystemHealthPendingIssue{}
	}
	end := start + pageSize
	if end > len(issues) {
		end = len(issues)
	}
	return issues[start:end]
}

func buildSystemHealthOverviewSummary(systems []response.SystemHealthSystem, scoreThresholds response.SystemHealthConfigResp, previousSystems ...[]response.SystemHealthSystem) response.SystemHealthOverviewSummary {
	scoreThresholds = normalizeSystemHealthScoreThresholds(scoreThresholds)
	summary := response.SystemHealthOverviewSummary{TotalSystems: len(systems)}
	for _, system := range systems {
		summary.Services += len(system.Services)
		hasSystemAggregateMetrics := hasSystemHealthAggregateMetricStatus(system.Metrics)
		if !hasSystemAggregateMetrics {
			summary.NoDataSystems++
		}
		if hasSystemAggregateMetrics {
			if system.Score < scoreThresholds.SevereAbnormalScoreThreshold {
				summary.LowScoreSystems++
			}
			if system.Score < scoreThresholds.GeneralAbnormalScoreThreshold {
				summary.AbnormalSystems++
			}
		}
		for _, service := range system.Services {
			if isNoDataSystemHealthService(service) {
				continue
			}
			if service.Score < scoreThresholds.GeneralAbnormalScoreThreshold {
				summary.AbnormalServices++
			}
		}
		for _, status := range system.Metrics {
			if isAbnormalMetricStatus(status) {
				summary.AbnormalMetrics++
			}
			if status == healthStatusRed {
				summary.SevereMetrics++
			}
		}
	}
	if len(previousSystems) > 0 {
		summary.NewAbnormalMetrics = countNewAbnormalMetrics(systems, previousSystems[0])
	}
	return summary
}

func countNewAbnormalMetrics(systems []response.SystemHealthSystem, previousSystems []response.SystemHealthSystem) int {
	count := 0
	previousMetrics := mapPreviousSystemMetrics(previousSystems)
	for _, system := range systems {
		for key, status := range system.Metrics {
			if isAbnormalMetricStatus(status) && !isAbnormalMetricStatus(previousMetrics[systemOverviewKey(system)][key]) {
				count++
			}
		}
	}
	return count
}

func mapPreviousSystemMetrics(previousSystems ...[]response.SystemHealthSystem) map[string]map[string]string {
	result := make(map[string]map[string]string)
	if len(previousSystems) == 0 {
		return result
	}
	for _, system := range previousSystems[0] {
		result[systemOverviewKey(system)] = system.Metrics
	}
	return result
}

func systemOverviewKey(system response.SystemHealthSystem) string {
	if system.GroupID != 0 {
		return strconv.FormatInt(system.GroupID, 10)
	}
	return system.Name
}

func isAbnormalMetricStatus(status string) bool {
	return status == healthStatusRed || status == healthStatusYellow
}

func buildSystemHealthPendingIssues(systems []response.SystemHealthSystem, limit int) []response.SystemHealthPendingIssue {
	issues := make([]response.SystemHealthPendingIssue, 0)
	for _, system := range systems {
		for _, service := range system.Services {
			for _, key := range systemHealthMetricKeys {
				status := service.Metrics[key]
				if status != healthStatusRed && status != healthStatusYellow {
					continue
				}
				detail := service.MetricDetails[key]
				issues = append(issues, response.SystemHealthPendingIssue{
					ID:           systemHealthPendingIssueID(system.GroupID, service.Name, key),
					GroupID:      system.GroupID,
					SystemName:   system.Name,
					ServiceName:  service.Name,
					MetricKey:    key,
					MetricLabel:  systemHealthMetricLabels[key],
					Status:       status,
					Value:        firstNonEmpty(detail.Value, "--"),
					Note:         firstNonEmpty(detail.Note, "暂无指标说明"),
					ServiceScore: service.Score,
					SystemScore:  system.Score,
				})
			}
		}
	}

	sort.SliceStable(issues, func(i, j int) bool {
		left := issues[i]
		right := issues[j]
		if healthStatusSeverity(left.Status) != healthStatusSeverity(right.Status) {
			return healthStatusSeverity(left.Status) > healthStatusSeverity(right.Status)
		}
		if left.ServiceScore != right.ServiceScore {
			return left.ServiceScore < right.ServiceScore
		}
		if left.SystemScore != right.SystemScore {
			return left.SystemScore < right.SystemScore
		}
		if left.SystemName != right.SystemName {
			return left.SystemName < right.SystemName
		}
		if left.ServiceName != right.ServiceName {
			return left.ServiceName < right.ServiceName
		}
		return systemHealthMetricOrder(left.MetricKey) < systemHealthMetricOrder(right.MetricKey)
	})

	if limit > 0 && len(issues) > limit {
		return issues[:limit]
	}
	return issues
}

func systemHealthPendingIssueID(groupID int64, serviceName string, metricKey string) string {
	return fmt.Sprintf("%d:%s:%s", groupID, serviceName, metricKey)
}

func systemHealthMetricOrder(metricKey string) int {
	for index, key := range systemHealthMetricKeys {
		if key == metricKey {
			return index
		}
	}
	return len(systemHealthMetricKeys)
}

func normalizeHealthScore(score int) int {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

func scoreHealthStatus(score int, configs ...response.SystemHealthConfigResp) string {
	scoreThresholds := defaultSystemHealthScoreThresholds()
	if len(configs) > 0 {
		scoreThresholds = normalizeSystemHealthScoreThresholds(configs[0])
	}
	if score < scoreThresholds.SevereAbnormalScoreThreshold {
		return healthStatusRed
	}
	if score < scoreThresholds.GeneralAbnormalScoreThreshold {
		return healthStatusYellow
	}
	return healthStatusGreen
}

func applySystemHealthScoreThresholds(systems []response.SystemHealthSystem, config response.SystemHealthConfigResp) {
	for i := range systems {
		applySystemHealthScoreThreshold(&systems[i], config)
	}
}

func applySystemHealthScoreThreshold(system *response.SystemHealthSystem, config response.SystemHealthConfigResp) {
	if system == nil {
		return
	}
	config = normalizeSystemHealthScoreThresholds(config)
	if !hasSystemHealthAggregateMetricStatus(system.Metrics) {
		system.Status = healthStatusGray
	} else {
		system.Status = scoreHealthStatus(system.Score, config)
	}
	for i := range system.Services {
		if isNoDataSystemHealthService(system.Services[i]) {
			system.Services[i].Status = healthStatusGray
			continue
		}
		system.Services[i].Status = scoreHealthStatus(system.Services[i].Score, config)
	}
	populateSystemHealthServiceCounts(system, config)
}

func rygScoreHealthStatus(score int) string {
	if score >= 3 {
		return healthStatusGreen
	}
	if score >= 1 {
		return healthStatusYellow
	}
	if score == 0 {
		return healthStatusRed
	}
	return healthStatusGray
}

func worseHealthStatus(left string, right string) string {
	if left == "" || left == healthStatusGray {
		return right
	}
	if healthStatusSeverity(right) > healthStatusSeverity(left) {
		return right
	}
	return left
}

func healthStatusSeverity(status string) int {
	switch status {
	case healthStatusRed:
		return 3
	case healthStatusYellow:
		return 2
	case healthStatusGray:
		return 1
	case healthStatusGreen:
		return 0
	default:
		return -1
	}
}

func metricStatusText(status string) string {
	switch status {
	case healthStatusRed:
		return "异常"
	case healthStatusYellow:
		return "关注"
	case healthStatusGreen:
		return "正常"
	default:
		return "--"
	}
}

func systemMetricDetail(services []response.SystemHealthService, key string, status string) response.SystemHealthMetricDetail {
	_, serviceName, representativeBadges := systemMetricRepresentativeValue(services, key, status)
	badges := []response.SystemHealthMetricBadge(nil)
	if serviceName != "" {
		badges = append(badges, response.SystemHealthMetricBadge{
			Label:  systemMetricContributorBadgeLabel(key),
			Value:  serviceName,
			Status: status,
		})
	}
	badges = append(badges, representativeBadges...)
	return response.SystemHealthMetricDetail{
		Status: status,
		Value:  "--",
		Note:   systemMetricNote(key, status),
		Badges: badges,
	}
}

func systemMetricContributorOnlyDetail(services []response.SystemHealthService, key string) (response.SystemHealthMetricDetail, bool) {
	status, serviceName, representativeBadges := systemMetricRepresentativeMetric(services, key)
	if serviceName == "" && len(representativeBadges) == 0 {
		return response.SystemHealthMetricDetail{}, false
	}
	badges := []response.SystemHealthMetricBadge(nil)
	if serviceName != "" {
		badges = append(badges, response.SystemHealthMetricBadge{
			Label:  systemMetricContributorBadgeLabel(key),
			Value:  serviceName,
			Status: status,
		})
	}
	badges = append(badges, representativeBadges...)
	return response.SystemHealthMetricDetail{
		Status: healthStatusGray,
		Value:  "--",
		Note:   "服务级指标仅用于异常来源分析",
		Badges: badges,
	}, true
}

func systemMetricRepresentativeValue(services []response.SystemHealthService, key string, status string) (string, string, []response.SystemHealthMetricBadge) {
	_, value, serviceName, badges := systemMetricRepresentativeValueByStatus(services, key, status)
	return value, serviceName, badges
}

func systemMetricRepresentativeMetric(services []response.SystemHealthService, key string) (string, string, []response.SystemHealthMetricBadge) {
	status, _, serviceName, badges := systemMetricRepresentativeValueByStatus(services, key, "")
	return status, serviceName, badges
}

func systemMetricRepresentativeValueByStatus(services []response.SystemHealthService, key string, expectedStatus string) (string, string, string, []response.SystemHealthMetricBadge) {
	type candidate struct {
		value       string
		serviceName string
		badges      []response.SystemHealthMetricBadge
		rank        float64
		status      string
	}

	var selected *candidate
	for _, service := range services {
		status := service.Metrics[key]
		detail := service.MetricDetails[key]
		if detail.Status != "" {
			status = detail.Status
		}
		if status == "" {
			status = healthStatusGray
		}
		if expectedStatus != "" && status != expectedStatus {
			continue
		}
		rank, ok := parseSystemMetricDetailRank(detail.Value)
		if !ok {
			continue
		}
		if selected == nil || rank > selected.rank {
			next := candidate{
				value:       detail.Value,
				serviceName: service.Name,
				badges:      append([]response.SystemHealthMetricBadge(nil), detail.Badges...),
				rank:        rank,
				status:      status,
			}
			selected = &next
		}
	}
	if selected != nil {
		return selected.status, selected.value, selected.serviceName, selected.badges
	}
	return healthStatusGray, "--", "", nil
}

func systemMetricContributorBadgeLabel(key string) string {
	switch key {
	case healthMetricResource:
		return "水位最高服务"
	case healthMetricResourceShift:
		return "主要贡献服务"
	case healthMetricErrorRate:
		return "错误率最高服务"
	case healthMetricLatency:
		return "响应最慢服务"
	case healthMetricErrorLog:
		return "错误日志最多服务"
	case healthMetricBadRequests:
		return "活跃告警最多服务"
	case healthMetricLockCost:
		return "锁等待最高服务"
	default:
		return "指标最高服务"
	}
}

func parseSystemMetricDetailRank(value string) (float64, bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || trimmed == "--" {
		return 0, false
	}

	var builder strings.Builder
	for _, char := range trimmed {
		if (char >= '0' && char <= '9') || char == '.' || char == '-' || char == '+' {
			builder.WriteRune(char)
			continue
		}
		if builder.Len() > 0 {
			break
		}
	}
	if builder.Len() == 0 {
		return 0, false
	}

	rank, err := strconv.ParseFloat(builder.String(), 64)
	if err != nil || math.IsNaN(rank) {
		return 0, false
	}
	return rank, true
}

func systemMetricNote(key string, status string) string {
	if status == healthStatusGray {
		return "暂无数据"
	}
	switch key {
	case healthMetricResource:
		return statusMetricNote(status, "资源水位正常", "资源水位需要关注", "资源水位超过异常阈值")
	case healthMetricResourceShift:
		return statusMetricNote(status, "资源波动处于稳定区间", "资源波动接近异常阈值", "资源波动超过异常阈值")
	case healthMetricErrorRate:
		return statusMetricNote(status, "错误率处于稳定区间", "错误率接近异常阈值", "错误率超过异常阈值")
	case healthMetricLatency:
		return statusMetricNote(status, "响应延时稳定", "响应延时升高", "响应延时超过异常阈值")
	case healthMetricErrorLog:
		return statusMetricNote(status, "错误日志数稳定", "错误日志数波动增加", "错误日志数超过异常阈值")
	case healthMetricBadRequests:
		return statusMetricNote(status, "活跃告警处于稳定区间", "活跃告警数升高", "活跃告警超过异常阈值")
	case healthMetricLockCost:
		return statusMetricNote(status, "锁等待稳定", "锁等待耗时升高", "锁等待耗时超过异常阈值")
	default:
		return statusMetricNote(status, "指标处于稳定区间", "指标需要关注", "指标超过异常阈值")
	}
}

func statusMetricNote(status string, green string, yellow string, red string) string {
	switch status {
	case healthStatusGreen:
		return green
	case healthStatusYellow:
		return yellow
	case healthStatusRed:
		return red
	default:
		return "暂无数据"
	}
}

type systemHealthExtraMetricResults struct {
	ResourceCPU              []prometheus.MetricResult
	ResourceMemory           []prometheus.MetricResult
	SystemResourceCPU        []prometheus.MetricResult
	SystemResourceMemory     []prometheus.MetricResult
	Latency                  []prometheus.MetricResult
	ErrorRate                []prometheus.MetricResult
	RequestCount             []prometheus.MetricResult
	ErrorCount               []prometheus.MetricResult
	ErrorLog                 []prometheus.MetricResult
	BadRequests              map[string]systemHealthAlertSummary
	ResourceShift            []prometheus.MetricResult
	LockCost                 []prometheus.MetricResult
	SystemLatency            []prometheus.MetricResult
	SystemErrorRate          []prometheus.MetricResult
	SystemRequestCount       []prometheus.MetricResult
	SystemErrorCount         []prometheus.MetricResult
	SystemLatencyTrend       []prometheus.MetricResult
	SystemLatencyDoD         []prometheus.MetricResult
	SystemErrorRateTrend     []prometheus.MetricResult
	SystemErrorLogTrend      []prometheus.MetricResult
	SystemErrorLogDoD        []prometheus.MetricResult
	SystemResourceShift      []prometheus.MetricResult
	SystemResourceShiftTrend []prometheus.MetricResult
	VMHostResource           systemHealthVMHostResourceResults
}

type systemHealthAlertSummary struct {
	Count            uint64
	RedCount         uint64
	YellowCount      uint64
	GrayCount        uint64
	HighestSeverity  model.SeverityLevel
	AffectedServices int
}

type systemHealthVMHostResourceResults struct {
	Scope      systemHealthVMHostResourceScope
	CPU        []prometheus.MetricResult
	Memory     []prometheus.MetricResult
	Shift      []prometheus.MetricResult
	ShiftTrend []prometheus.MetricResult
}

type systemHealthVMHostResourceScope struct {
	NodesByService map[string]map[string]struct{}
	SystemNodes    map[string]struct{}
}

func newSystemHealthVMHostResourceScope() systemHealthVMHostResourceScope {
	return systemHealthVMHostResourceScope{
		NodesByService: map[string]map[string]struct{}{},
		SystemNodes:    map[string]struct{}{},
	}
}

func (s *systemHealthVMHostResourceScope) addInstance(serviceName string, instance *model.ServiceInstance) {
	if instance == nil || instance.PodName != "" || instance.ContainerId != "" || instance.NodeName == "" || instance.Pid <= 1 {
		return
	}
	if s.NodesByService == nil {
		s.NodesByService = map[string]map[string]struct{}{}
	}
	if s.SystemNodes == nil {
		s.SystemNodes = map[string]struct{}{}
	}
	if serviceName == "" {
		serviceName = instance.ServiceName
	}
	if serviceName == "" {
		return
	}
	if _, ok := s.NodesByService[serviceName]; !ok {
		s.NodesByService[serviceName] = map[string]struct{}{}
	}
	s.NodesByService[serviceName][instance.NodeName] = struct{}{}
	s.SystemNodes[instance.NodeName] = struct{}{}
}

func (s systemHealthVMHostResourceScope) hasNodes() bool {
	return len(s.SystemNodes) > 0
}

func (s systemHealthVMHostResourceScope) sortedNodes() []string {
	return systemHealthSortedVMHostNodes(s.SystemNodes)
}

func (s systemHealthVMHostResourceScope) sortedServiceNodes(serviceName string) []string {
	return systemHealthSortedVMHostNodes(s.NodesByService[serviceName])
}

func systemHealthSortedVMHostNodes(nodes map[string]struct{}) []string {
	result := make([]string, 0, len(nodes))
	for node := range nodes {
		result = append(result, node)
	}
	sort.Strings(result)
	return result
}

func (h *handler) buildSystemHealthVMHostResourceScope(ctx core.Context, startTime, endTime time.Time, groupFilter prometheus.PQLFilter, services []response.SystemHealthService) (systemHealthVMHostResourceScope, error) {
	scope := newSystemHealthVMHostResourceScope()
	serviceNames := systemHealthServiceNames(services)
	if len(serviceNames) == 0 {
		return scope, nil
	}

	spanFilter := prometheus.Clone(groupFilter).RegexMatch(prometheus.ServiceNameKey, prometheus.RegexMultipleValue(serviceNames...))
	instancesByService, err := h.promRepo.GetMultiSVCInstanceListByPQLFilter(ctx, startTime.UnixMicro(), endTime.UnixMicro(), spanFilter)
	if err != nil {
		return scope, err
	}
	for _, serviceName := range serviceNames {
		serviceInstances := instancesByService[serviceName]
		if serviceInstances == nil {
			continue
		}
		for _, instance := range serviceInstances.GetInstances() {
			scope.addInstance(serviceName, instance)
		}
	}
	return scope, nil
}

func (h *handler) querySystemHealthExtraMetrics(ctx core.Context, startTime time.Time, endTime time.Time, groupID int64, services []response.SystemHealthService) (systemHealthExtraMetricResults, error) {
	groupFilter, err := common.GetPQLFilterByGroupID(ctx, h.dbRepo, "apm", groupID)
	if err != nil {
		return systemHealthExtraMetricResults{}, err
	}
	window := prometheus.VecFromS2E(startTime.UnixMicro(), endTime.UnixMicro())
	queryTime := endTime
	trendStep := systemHealthTrendStep(startTime, endTime)
	trendStepMicroS := int64(trendStep / time.Microsecond)

	result := systemHealthExtraMetricResults{}
	if latency, err := h.promRepo.QueryMetricsWithPQLFilter(ctx, prometheus.PQLAvgLatencyWithPQLFilter, startTime.UnixMicro(), endTime.UnixMicro(), prometheus.SVCGranularity, groupFilter); err == nil {
		result.Latency = latency
	}
	if latency, err := h.promRepo.QueryMetricsWithPQLFilter(ctx, prometheus.PQLAvgLatencyWithPQLFilter, startTime.UnixMicro(), endTime.UnixMicro(), prometheus.Granularity(""), groupFilter); err == nil {
		result.SystemLatency = latency
	}
	if latencyTrend, err := h.promRepo.QueryRangeMetricsWithPQLFilter(ctx, prometheus.PQLAvgLatencyWithPQLFilter, startTime.UnixMicro(), endTime.UnixMicro(), trendStepMicroS, prometheus.Granularity(""), groupFilter); err == nil {
		result.SystemLatencyTrend = latencyTrend
	}
	if latencyDoD, err := h.promRepo.QueryMetricsWithPQLFilter(ctx, prometheus.DayOnDayTemplate(prometheus.PQLAvgLatencyWithPQLFilter), startTime.UnixMicro(), endTime.UnixMicro(), prometheus.Granularity(""), groupFilter); err == nil {
		result.SystemLatencyDoD = latencyDoD
	}
	if errorRate, err := h.promRepo.QueryMetricsWithPQLFilter(ctx, prometheus.PQLAvgErrorRateWithPQLFilter, startTime.UnixMicro(), endTime.UnixMicro(), prometheus.SVCGranularity, groupFilter); err == nil {
		result.ErrorRate = errorRate
	}
	if errorRate, err := h.promRepo.QueryMetricsWithPQLFilter(ctx, prometheus.PQLAvgErrorRateWithPQLFilter, startTime.UnixMicro(), endTime.UnixMicro(), prometheus.Granularity(""), groupFilter); err == nil {
		result.SystemErrorRate = errorRate
	}
	if requestCount, err := h.promRepo.QueryMetricsWithPQLFilter(ctx, prometheus.PQLRequestCountWithPQLFilter, startTime.UnixMicro(), endTime.UnixMicro(), prometheus.SVCGranularity, groupFilter); err == nil {
		result.RequestCount = requestCount
	}
	if requestCount, err := h.promRepo.QueryMetricsWithPQLFilter(ctx, prometheus.PQLRequestCountWithPQLFilter, startTime.UnixMicro(), endTime.UnixMicro(), prometheus.Granularity(""), groupFilter); err == nil {
		result.SystemRequestCount = requestCount
	}
	if errorCount, err := h.promRepo.QueryMetricsWithPQLFilter(ctx, prometheus.PQLErrorCountWithPQLFilter, startTime.UnixMicro(), endTime.UnixMicro(), prometheus.SVCGranularity, groupFilter); err == nil {
		result.ErrorCount = errorCount
	}
	if errorCount, err := h.promRepo.QueryMetricsWithPQLFilter(ctx, prometheus.PQLErrorCountWithPQLFilter, startTime.UnixMicro(), endTime.UnixMicro(), prometheus.Granularity(""), groupFilter); err == nil {
		result.SystemErrorCount = errorCount
	}
	if errorRateTrend, err := h.promRepo.QueryRangeMetricsWithPQLFilter(ctx, prometheus.PQLAvgErrorRateWithPQLFilter, startTime.UnixMicro(), endTime.UnixMicro(), trendStepMicroS, prometheus.Granularity(""), groupFilter); err == nil {
		result.SystemErrorRateTrend = errorRateTrend
	}
	if errorLog, err := h.promRepo.QueryMetricsWithPQLFilter(ctx, prometheus.PQLAvgLogErrorCountCombineEndpointsInfoWithPQLFilter, startTime.UnixMicro(), endTime.UnixMicro(), prometheus.SVCGranularity, groupFilter); err == nil {
		result.ErrorLog = errorLog
	}
	if errorLogTrend, err := h.promRepo.QueryRangeMetricsWithPQLFilter(ctx, prometheus.PQLAvgLogErrorCountCombineEndpointsInfoWithPQLFilter, startTime.UnixMicro(), endTime.UnixMicro(), trendStepMicroS, prometheus.SVCGranularity, groupFilter); err == nil {
		result.SystemErrorLogTrend = errorLogTrend
	}
	if errorLogDoD, err := h.promRepo.QueryMetricsWithPQLFilter(ctx, prometheus.DayOnDayTemplate(prometheus.PQLAvgLogErrorCountCombineEndpointsInfoWithPQLFilter), startTime.UnixMicro(), endTime.UnixMicro(), prometheus.Granularity(""), groupFilter); err == nil {
		result.SystemErrorLogDoD = errorLogDoD
	}
	if badRequests, err := h.querySystemHealthBadRequestsByService(ctx, startTime, endTime, groupFilter, services); err == nil {
		result.BadRequests = badRequests
	}

	if resourceCPU, err := h.promRepo.QueryData(ctx, queryTime, buildSystemHealthCPUWaterlinePQL(groupFilter, window)); err == nil {
		result.ResourceCPU = resourceCPU
	}
	if resourceMemory, err := h.promRepo.QueryData(ctx, queryTime, buildSystemHealthMemoryWaterlinePQL(groupFilter, window)); err == nil {
		result.ResourceMemory = resourceMemory
	}
	if resourceCPU, err := h.promRepo.QueryData(ctx, queryTime, buildSystemHealthSystemCPUWaterlinePQL(groupFilter, window)); err == nil {
		result.SystemResourceCPU = resourceCPU
	}
	if resourceMemory, err := h.promRepo.QueryData(ctx, queryTime, buildSystemHealthSystemMemoryWaterlinePQL(groupFilter, window)); err == nil {
		result.SystemResourceMemory = resourceMemory
	}

	resourceShiftQuery := buildSystemHealthResourceShiftPQL(groupFilter, window)
	if resourceShift, err := h.promRepo.QueryData(ctx, queryTime, resourceShiftQuery); err == nil {
		result.ResourceShift = resourceShift
	}
	if resourceShift, err := h.promRepo.QueryData(ctx, queryTime, buildSystemHealthSystemCPUResourceShiftPQL(groupFilter, window)); err == nil {
		result.SystemResourceShift = append(result.SystemResourceShift, resourceShift...)
	}
	if resourceShift, err := h.promRepo.QueryData(ctx, queryTime, buildSystemHealthSystemMemoryResourceShiftPQL(groupFilter, window)); err == nil {
		result.SystemResourceShift = append(result.SystemResourceShift, resourceShift...)
	}
	if resourceShiftTrend, err := h.promRepo.QueryRangeData(ctx, startTime, endTime, buildSystemHealthSystemCPUResourceShiftPQL(groupFilter, prometheus.VecFromDuration(trendStep)), trendStep); err == nil {
		result.SystemResourceShiftTrend = append(result.SystemResourceShiftTrend, resourceShiftTrend...)
	}
	if resourceShiftTrend, err := h.promRepo.QueryRangeData(ctx, startTime, endTime, buildSystemHealthSystemMemoryResourceShiftPQL(groupFilter, prometheus.VecFromDuration(trendStep)), trendStep); err == nil {
		result.SystemResourceShiftTrend = append(result.SystemResourceShiftTrend, resourceShiftTrend...)
	}
	if vmHostScope, err := h.buildSystemHealthVMHostResourceScope(ctx, startTime, endTime, groupFilter, services); err == nil && vmHostScope.hasNodes() {
		nodes := vmHostScope.sortedNodes()
		result.VMHostResource.Scope = vmHostScope
		if resourceCPU, err := h.promRepo.QueryData(ctx, queryTime, buildSystemHealthVMHostCPUWaterlinePQL(nodes, window, "")); err == nil {
			result.VMHostResource.CPU = resourceCPU
		}
		if resourceMemory, err := h.promRepo.QueryData(ctx, queryTime, buildSystemHealthVMHostMemoryWaterlinePQL(nodes, window, "")); err == nil {
			result.VMHostResource.Memory = resourceMemory
		}
		hostShiftQuery := buildSystemHealthVMHostResourceShiftPQL(nodes, window)
		if resourceShift, err := h.promRepo.QueryData(ctx, queryTime, hostShiftQuery); err == nil {
			result.VMHostResource.Shift = resourceShift
		}
		hostShiftTrendQuery := buildSystemHealthVMHostResourceShiftPQL(nodes, prometheus.VecFromDuration(trendStep))
		if resourceShiftTrend, err := h.promRepo.QueryRangeData(ctx, startTime, endTime, hostShiftTrendQuery, trendStep); err == nil {
			result.VMHostResource.ShiftTrend = resourceShiftTrend
			result.SystemResourceShiftTrend = append(result.SystemResourceShiftTrend, resourceShiftTrend...)
		}
	} else if err != nil {
		systemHealthWarnCtx(ctx, "system-health VM host resource scope skipped", zap.Error(err))
	}

	lockCostQuery := buildSystemHealthLockCostPQL(groupFilter, window)
	if lockCost, err := h.promRepo.QueryData(ctx, queryTime, lockCostQuery); err == nil {
		result.LockCost = lockCost
	} else if lockCost, fallbackErr := h.querySystemHealthLockCostByServiceInstances(ctx, startTime, endTime, groupFilter, services, window); fallbackErr == nil {
		result.LockCost = lockCost
	}
	return result, nil
}

func buildSystemHealthCPUWaterlinePQL(filter prometheus.PQLFilter, window string) string {
	resourceFilter, spanFilter := splitSystemHealthResourceFilters(filter)
	usageFilter := prometheus.Clone(resourceFilter).NotEqual("container", "").NotEqual("container", "POD")
	quotaFilter := prometheus.Clone(resourceFilter).NotEqual("container", "").NotEqual("container", "POD")
	periodFilter := prometheus.Clone(resourceFilter).NotEqual("container", "").NotEqual("container", "POD")
	usage := fmt.Sprintf(
		`sum by (cluster_id,namespace,pod) (rate(%s))`,
		systemHealthRangeSelector("container_cpu_usage_seconds_total", usageFilter, window, ""),
	)
	limit := fmt.Sprintf(
		`sum by (cluster_id,namespace,pod) ((avg_over_time(%s) > 0) / (avg_over_time(%s) > 0))`,
		systemHealthRangeSelector("container_spec_cpu_quota", quotaFilter, window, ""),
		systemHealthRangeSelector("container_spec_cpu_period", periodFilter, window, ""),
	)
	return systemHealthPodMetricJoinServiceMax(fmt.Sprintf(`(%s) / (%s)`, usage, limit), spanFilter, window, "")
}

func buildSystemHealthMemoryWaterlinePQL(filter prometheus.PQLFilter, window string) string {
	resourceFilter, spanFilter := splitSystemHealthResourceFilters(filter)
	usageFilter := prometheus.Clone(resourceFilter).NotEqual("container", "").NotEqual("container", "POD")
	limitFilter := prometheus.Clone(resourceFilter).NotEqual("container", "").NotEqual("container", "POD")
	usage := fmt.Sprintf(
		`sum by (cluster_id,namespace,pod) (avg_over_time(%s))`,
		systemHealthRangeSelector("container_memory_working_set_bytes", usageFilter, window, ""),
	)
	limit := fmt.Sprintf(
		`sum by (cluster_id,namespace,pod) (avg_over_time(%s) > 0)`,
		systemHealthRangeSelector("container_spec_memory_limit_bytes", limitFilter, window, ""),
	)
	return systemHealthPodMetricJoinServiceMax(fmt.Sprintf(`(%s) / (%s)`, usage, limit), spanFilter, window, "")
}

func buildSystemHealthSystemCPUWaterlinePQL(filter prometheus.PQLFilter, window string) string {
	resourceFilter, spanFilter := splitSystemHealthResourceFilters(filter)
	usageFilter := prometheus.Clone(resourceFilter).NotEqual("container", "").NotEqual("container", "POD")
	quotaFilter := prometheus.Clone(resourceFilter).NotEqual("container", "").NotEqual("container", "POD")
	periodFilter := prometheus.Clone(resourceFilter).NotEqual("container", "").NotEqual("container", "POD")
	usage := fmt.Sprintf(
		`sum by (cluster_id,namespace,pod) (rate(%s))`,
		systemHealthRangeSelector("container_cpu_usage_seconds_total", usageFilter, window, ""),
	)
	limit := fmt.Sprintf(
		`sum by (cluster_id,namespace,pod) ((avg_over_time(%s) > 0) / (avg_over_time(%s) > 0))`,
		systemHealthRangeSelector("container_spec_cpu_quota", quotaFilter, window, ""),
		systemHealthRangeSelector("container_spec_cpu_period", periodFilter, window, ""),
	)
	return systemHealthPodMetricFilterBySystemPods(fmt.Sprintf(`(%s) / (%s)`, usage, limit), spanFilter, window, "")
}

func buildSystemHealthSystemMemoryWaterlinePQL(filter prometheus.PQLFilter, window string) string {
	resourceFilter, spanFilter := splitSystemHealthResourceFilters(filter)
	usageFilter := prometheus.Clone(resourceFilter).NotEqual("container", "").NotEqual("container", "POD")
	limitFilter := prometheus.Clone(resourceFilter).NotEqual("container", "").NotEqual("container", "POD")
	usage := fmt.Sprintf(
		`sum by (cluster_id,namespace,pod) (avg_over_time(%s))`,
		systemHealthRangeSelector("container_memory_working_set_bytes", usageFilter, window, ""),
	)
	limit := fmt.Sprintf(
		`sum by (cluster_id,namespace,pod) (avg_over_time(%s) > 0)`,
		systemHealthRangeSelector("container_spec_memory_limit_bytes", limitFilter, window, ""),
	)
	return systemHealthPodMetricFilterBySystemPods(fmt.Sprintf(`(%s) / (%s)`, usage, limit), spanFilter, window, "")
}

func buildSystemHealthResourceShiftPQL(filter prometheus.PQLFilter, window string) string {
	resourceFilter, spanFilter := splitSystemHealthResourceFilters(filter)
	cpuRatio := systemHealthResourceRatioPQL(
		`rate(%s)`,
		"container_cpu_usage_seconds_total",
		resourceFilter,
		spanFilter,
		window,
	)
	memoryRatio := systemHealthResourceRatioPQL(
		`avg_over_time(%s)`,
		"container_memory_working_set_bytes",
		resourceFilter,
		spanFilter,
		window,
	)
	return fmt.Sprintf(
		`max by (svc_name) (((%s) >= 1) or (1 / (%s)) or ((%s) >= 1) or (1 / (%s)))`,
		cpuRatio,
		cpuRatio,
		memoryRatio,
		memoryRatio,
	)
}

func buildSystemHealthSystemCPUResourceShiftPQL(filter prometheus.PQLFilter, window string) string {
	resourceFilter, spanFilter := splitSystemHealthResourceFilters(filter)
	return systemHealthSystemResourceShiftPQL(
		`rate(%s)`,
		"container_cpu_usage_seconds_total",
		resourceFilter,
		spanFilter,
		window,
	)
}

func buildSystemHealthSystemMemoryResourceShiftPQL(filter prometheus.PQLFilter, window string) string {
	resourceFilter, spanFilter := splitSystemHealthResourceFilters(filter)
	return systemHealthSystemResourceShiftPQL(
		`avg_over_time(%s)`,
		"container_memory_working_set_bytes",
		resourceFilter,
		spanFilter,
		window,
	)
}

func systemHealthSystemResourceShiftPQL(aggTemplate string, metric string, resourceFilter prometheus.PQLFilter, spanFilter prometheus.PQLFilter, window string) string {
	ratio := systemHealthSystemResourceRatioPQL(aggTemplate, metric, resourceFilter, spanFilter, window)
	return fmt.Sprintf(`((%s) >= 1) or (1 / (%s))`, ratio, ratio)
}

func buildSystemHealthVMHostCPUWaterlinePQL(nodes []string, window string, offset string) string {
	filter := systemHealthVMHostNodeFilter(nodes)
	filter.AddPatternFilter("mode!~", "idle|iowait|steal")
	return fmt.Sprintf(
		`avg by (node_name) (sum by (node_name,cpu) (label_replace(rate(%s), "node_name", "$1", "instance_name", "(.*)")))`,
		systemHealthRangeSelector("node_cpu_seconds_total", filter, window, offset),
	)
}

func buildSystemHealthVMHostMemoryWaterlinePQL(nodes []string, window string, offset string) string {
	available := fmt.Sprintf(
		`label_replace(avg_over_time(%s), "node_name", "$1", "instance_name", "(.*)")`,
		systemHealthRangeSelector("node_memory_MemAvailable_bytes", systemHealthVMHostNodeFilter(nodes), window, offset),
	)
	total := fmt.Sprintf(
		`label_replace(avg_over_time(%s), "node_name", "$1", "instance_name", "(.*)")`,
		systemHealthRangeSelector("node_memory_MemTotal_bytes", systemHealthVMHostNodeFilter(nodes), window, offset),
	)
	return fmt.Sprintf(`max by (node_name) (1 - ((%s) / (%s)))`, available, total)
}

func buildSystemHealthVMHostResourceShiftPQL(nodes []string, window string) string {
	cpuRatio := fmt.Sprintf(
		`(%s) / (%s)`,
		buildSystemHealthVMHostCPUWaterlinePQL(nodes, window, ""),
		buildSystemHealthVMHostCPUWaterlinePQL(nodes, window, "offset 24h"),
	)
	memoryRatio := fmt.Sprintf(
		`(%s) / (%s)`,
		buildSystemHealthVMHostMemoryWaterlinePQL(nodes, window, ""),
		buildSystemHealthVMHostMemoryWaterlinePQL(nodes, window, "offset 24h"),
	)
	return fmt.Sprintf(
		`max by (node_name) (((%s) >= 1) or (1 / (%s)) or ((%s) >= 1) or (1 / (%s)))`,
		cpuRatio,
		cpuRatio,
		memoryRatio,
		memoryRatio,
	)
}

func systemHealthVMHostNodeFilter(nodes []string) prometheus.PQLFilter {
	return prometheus.NewFilter().RegexMatch("instance_name", prometheus.RegexMultipleValue(nodes...))
}

func buildSystemHealthLockCostPQL(filter prometheus.PQLFilter, window string) string {
	resourceFilter, spanFilter := splitSystemHealthResourceFilters(filter)
	return buildSystemHealthLockCostPQLWithFilters(resourceFilter, spanFilter, window)
}

func buildSystemHealthLockCostPQLWithFilters(resourceFilter prometheus.PQLFilter, spanFilter prometheus.PQLFilter, window string) string {
	futexCost := systemHealthPolarisByServicePQL(
		"originx_thread_polaris_nanoseconds_sum",
		prometheus.Clone(resourceFilter).Equal("type", "futex"),
		spanFilter,
		window,
		"",
	)
	requestCount := fmt.Sprintf(
		`sum by (svc_name) (increase(%s))`,
		systemHealthRangeSelector("kindling_span_trace_duration_nanoseconds_count", spanFilter, window, ""),
	)
	polarisBaseline := systemHealthPolarisByServicePQL(
		"originx_thread_polaris_nanoseconds_count",
		prometheus.Clone(resourceFilter).Equal("type", "cpu"),
		spanFilter,
		window,
		"",
	)
	return fmt.Sprintf(`((%s) / (%s)) or ((%s) * 0 + 0)`, futexCost, requestCount, polarisBaseline)
}

func (h *handler) querySystemHealthLockCostByServiceInstances(ctx core.Context, startTime time.Time, endTime time.Time, groupFilter prometheus.PQLFilter, services []response.SystemHealthService, window string) ([]prometheus.MetricResult, error) {
	serviceNames := systemHealthServiceNames(services)
	if len(serviceNames) == 0 {
		return nil, nil
	}

	spanFilter := prometheus.Clone(groupFilter).RegexMatch(prometheus.ServiceNameKey, prometheus.RegexMultipleValue(serviceNames...))
	instancesByService, err := h.promRepo.GetMultiSVCInstanceListByPQLFilter(ctx, startTime.UnixMicro(), endTime.UnixMicro(), spanFilter)
	if err != nil {
		return nil, err
	}

	instances := systemHealthFlattenServiceInstances(instancesByService, serviceNames)
	if len(instances) == 0 {
		return nil, nil
	}

	return h.promRepo.QueryData(ctx, endTime, buildSystemHealthLockCostPQLForInstances(instances, window))
}

func buildSystemHealthLockCostPQLForInstances(instances []*model.ServiceInstance, window string) string {
	resourceFilter, spanFilter := systemHealthInstanceFilters(instances)
	return buildSystemHealthLockCostPQLWithFilters(resourceFilter, spanFilter, window)
}

func (h *handler) querySystemHealthBadRequestsByService(ctx core.Context, startTime time.Time, endTime time.Time, groupFilter prometheus.PQLFilter, services []response.SystemHealthService) (map[string]systemHealthAlertSummary, error) {
	if h.chRepo == nil {
		return nil, nil
	}
	serviceNames := systemHealthServiceNames(services)
	if len(serviceNames) == 0 {
		return nil, nil
	}

	spanFilter := prometheus.Clone(groupFilter).RegexMatch(prometheus.ServiceNameKey, prometheus.RegexMultipleValue(serviceNames...))
	instancesByService, err := h.promRepo.GetMultiSVCInstanceListByPQLFilter(ctx, startTime.UnixMicro(), endTime.UnixMicro(), spanFilter)
	if err != nil {
		return nil, err
	}

	instances := systemHealthFlattenServiceInstances(instancesByService, serviceNames)
	alertCounts, err := h.chRepo.GetAlertEventCountGroupByInstance(
		ctx,
		startTime,
		endTime,
		request.AlertFilter{Status: "firing"},
		&model.RelatedInstances{SIs: instances},
	)
	if err != nil {
		return nil, err
	}

	results := make(map[string]systemHealthAlertSummary, len(serviceNames))
	for _, serviceName := range serviceNames {
		serviceInstances := instancesByService[serviceName]
		if serviceInstances == nil {
			continue
		}
		levelCounts := serviceoverview.GroupAlertEventCountListByInstance(alertCounts, serviceInstances.GetInstances())
		summary := systemHealthAlertSummaryFromLevelCounts(levelCounts)
		if summary.Count > 0 {
			summary.AffectedServices = 1
		}
		results[serviceName] = summary
	}
	return results, nil
}

func systemHealthResourceRatioPQL(aggTemplate string, metric string, resourceFilter prometheus.PQLFilter, spanFilter prometheus.PQLFilter, window string) string {
	now := systemHealthResourceByServicePQL(aggTemplate, metric, resourceFilter, spanFilter, window, "")
	lastDay := systemHealthResourceByServicePQL(aggTemplate, metric, resourceFilter, spanFilter, window, "offset 24h")
	return fmt.Sprintf(`(%s) / (%s)`, now, lastDay)
}

func systemHealthSystemResourceRatioPQL(aggTemplate string, metric string, resourceFilter prometheus.PQLFilter, spanFilter prometheus.PQLFilter, window string) string {
	now := systemHealthResourceBySystemPQL(aggTemplate, metric, resourceFilter, spanFilter, window, "")
	lastDay := systemHealthResourceBySystemPQL(aggTemplate, metric, resourceFilter, spanFilter, window, "offset 24h")
	return fmt.Sprintf(`(%s) / (%s)`, now, lastDay)
}

func systemHealthResourceByServicePQL(aggTemplate string, metric string, resourceFilter prometheus.PQLFilter, spanFilter prometheus.PQLFilter, window string, offset string) string {
	resourceSeries := fmt.Sprintf(
		`sum by (cluster_id,namespace,pod) (`+aggTemplate+`)`,
		systemHealthRangeSelector(metric, prometheus.Clone(resourceFilter).NotEqual("container", "").NotEqual("container", "POD"), window, offset),
	)
	return systemHealthPodMetricJoinService(resourceSeries, spanFilter, window, offset)
}

func systemHealthResourceBySystemPQL(aggTemplate string, metric string, resourceFilter prometheus.PQLFilter, spanFilter prometheus.PQLFilter, window string, offset string) string {
	resourceSeries := fmt.Sprintf(
		`sum by (cluster_id,namespace,pod) (`+aggTemplate+`)`,
		systemHealthRangeSelector(metric, prometheus.Clone(resourceFilter).NotEqual("container", "").NotEqual("container", "POD"), window, offset),
	)
	return fmt.Sprintf(`sum (%s)`, systemHealthPodMetricFilterBySystemPods(resourceSeries, spanFilter, window, offset))
}

func systemHealthPolarisByServicePQL(metric string, resourceFilter prometheus.PQLFilter, spanFilter prometheus.PQLFilter, window string, offset string) string {
	polarisSeries := fmt.Sprintf(
		`sum by (cluster_id,namespace,pod) (increase(%s))`,
		systemHealthRangeSelector(metric, resourceFilter, window, offset),
	)
	return systemHealthPodMetricJoinService(polarisSeries, spanFilter, window, offset)
}

func systemHealthPodMetricJoinService(podMetric string, spanFilter prometheus.PQLFilter, window string, offset string) string {
	spanSeries := fmt.Sprintf(
		`group by (cluster_id,namespace,pod,svc_name) (last_over_time(%s))`,
		systemHealthRangeSelector("kindling_span_trace_duration_nanoseconds_count", prometheus.Clone(spanFilter).NotEqual("pod", ""), window, offset),
	)
	return fmt.Sprintf(`sum by (svc_name) ((%s) * on (cluster_id,namespace,pod) group_left (svc_name) (%s))`, podMetric, spanSeries)
}

func systemHealthPodMetricJoinServiceMax(podMetric string, spanFilter prometheus.PQLFilter, window string, offset string) string {
	spanSeries := fmt.Sprintf(
		`group by (cluster_id,namespace,pod,svc_name) (last_over_time(%s))`,
		systemHealthRangeSelector("kindling_span_trace_duration_nanoseconds_count", prometheus.Clone(spanFilter).NotEqual("pod", ""), window, offset),
	)
	return fmt.Sprintf(`max by (svc_name) ((%s) * on (cluster_id,namespace,pod) group_left (svc_name) (%s))`, podMetric, spanSeries)
}

func systemHealthPodMetricFilterBySystemPods(podMetric string, spanFilter prometheus.PQLFilter, window string, offset string) string {
	spanSeries := fmt.Sprintf(
		`group by (cluster_id,namespace,pod) (last_over_time(%s))`,
		systemHealthRangeSelector("kindling_span_trace_duration_nanoseconds_count", prometheus.Clone(spanFilter).NotEqual("pod", ""), window, offset),
	)
	return fmt.Sprintf(`max by (cluster_id,namespace,pod) ((%s) * on (cluster_id,namespace,pod) (%s))`, podMetric, spanSeries)
}

func splitSystemHealthResourceFilters(filter prometheus.PQLFilter) (prometheus.PQLFilter, prometheus.PQLFilter) {
	spanFilter := prometheus.Clone(filter)
	resourceFilter, _ := prometheus.Clone(filter).SplitFilters([]string{
		prometheus.ServiceNameKey,
		prometheus.ContentKeyKey,
		"is_error",
	})
	return resourceFilter, spanFilter
}

func systemHealthRangeSelector(metric string, filter prometheus.PQLFilter, window string, offset string) string {
	filterText := ""
	if filter != nil {
		filterText = filter.String()
	}
	if offset == "" {
		return fmt.Sprintf(`%s{%s}[%s]`, metric, filterText, window)
	}
	return fmt.Sprintf(`%s{%s}[%s] %s`, metric, filterText, window, offset)
}

func systemHealthServiceNames(services []response.SystemHealthService) []string {
	names := make([]string, 0, len(services))
	seen := make(map[string]struct{}, len(services))
	for _, service := range services {
		if service.Name == "" {
			continue
		}
		if _, ok := seen[service.Name]; ok {
			continue
		}
		seen[service.Name] = struct{}{}
		names = append(names, service.Name)
	}
	sort.Strings(names)
	return names
}

func systemHealthFlattenServiceInstances(instancesByService map[string]*model.ServiceInstances, serviceNames []string) []*model.ServiceInstance {
	instances := make([]*model.ServiceInstance, 0)
	for _, serviceName := range serviceNames {
		serviceInstances := instancesByService[serviceName]
		if serviceInstances == nil {
			continue
		}
		for _, instance := range serviceInstances.GetInstances() {
			if instance == nil || instance.PodName == "" {
				continue
			}
			instance.ServiceName = serviceName
			instances = append(instances, instance)
		}
	}
	return instances
}

func systemHealthInstanceFilters(instances []*model.ServiceInstance) (prometheus.PQLFilter, prometheus.PQLFilter) {
	resourceFilters := make([]prometheus.PQLFilter, 0, len(instances))
	spanFilters := make([]prometheus.PQLFilter, 0, len(instances))
	seen := make(map[string]struct{}, len(instances))
	for _, instance := range instances {
		if instance == nil || instance.PodName == "" {
			continue
		}
		key := fmt.Sprintf("%s\x00%s\x00%s\x00%s", instance.ClusterID, instance.Namespace, instance.PodName, instance.ServiceName)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		resourceFilter := prometheus.NewFilter()
		if instance.ClusterID != "" {
			resourceFilter.Equal("cluster_id", instance.ClusterID)
		}
		if instance.Namespace != "" {
			resourceFilter.Equal("namespace", instance.Namespace)
		}
		resourceFilter.Equal("pod", instance.PodName)
		resourceFilters = append(resourceFilters, resourceFilter)

		spanFilter := prometheus.NewFilter()
		if instance.ClusterID != "" {
			spanFilter.Equal("cluster_id", instance.ClusterID)
		}
		if instance.Namespace != "" {
			spanFilter.Equal("namespace", instance.Namespace)
		}
		spanFilter.Equal("pod", instance.PodName)
		if instance.ServiceName != "" {
			spanFilter.Equal(prometheus.ServiceNameKey, instance.ServiceName)
		}
		spanFilters = append(spanFilters, spanFilter)
	}
	return prometheus.Or(resourceFilters...), prometheus.Or(spanFilters...)
}

func systemHealthAlertCount(levelCounts model.AlertEventLevelCountMap) uint64 {
	total := uint64(0)
	for _, counts := range levelCounts {
		for _, count := range counts {
			total += count
		}
	}
	return total
}

func systemHealthAlertSummaryFromLevelCounts(levelCounts model.AlertEventLevelCountMap) systemHealthAlertSummary {
	summary := systemHealthAlertSummary{}
	for _, counts := range levelCounts {
		for severity, count := range counts {
			summary.Count += count
			if severity > summary.HighestSeverity {
				summary.HighestSeverity = severity
			}
			switch severity {
			case model.SeverityLevelCritical, model.SeverityLevelError:
				summary.RedCount += count
			case model.SeverityLevelWarning:
				summary.YellowCount += count
			default:
				summary.GrayCount += count
			}
		}
	}
	return summary
}

func systemHealthAlertSummaryStatus(summary systemHealthAlertSummary) string {
	if summary.Count == 0 {
		return healthStatusGreen
	}
	if summary.RedCount > 0 {
		return healthStatusRed
	}
	if summary.YellowCount > 0 {
		return healthStatusYellow
	}
	return healthStatusGray
}

func systemHealthMergeAlertSummary(total systemHealthAlertSummary, next systemHealthAlertSummary) systemHealthAlertSummary {
	total.Count += next.Count
	total.RedCount += next.RedCount
	total.YellowCount += next.YellowCount
	total.GrayCount += next.GrayCount
	total.AffectedServices += next.AffectedServices
	if next.HighestSeverity > total.HighestSeverity {
		total.HighestSeverity = next.HighestSeverity
	}
	return total
}

func applySystemHealthExtraMetrics(system *response.SystemHealthSystem, extra systemHealthExtraMetricResults) {
	if system == nil || len(system.Services) == 0 {
		return
	}

	requestCountByService := systemHealthCurrentValueByService(extra.RequestCount)
	errorCountByService := systemHealthCurrentValueByService(extra.ErrorCount)
	resourceByService := systemHealthResourceWaterlineDetails(extra.ResourceCPU, extra.ResourceMemory)
	vmHostResourceByService := systemHealthVMHostResourceWaterlineDetails(extra.VMHostResource.Scope, extra.VMHostResource.CPU, extra.VMHostResource.Memory)
	latencyByService := systemHealthLatencyDetails(extra.Latency)
	errorRateByService := systemHealthErrorRateDetails(extra.ErrorRate, requestCountByService, errorCountByService)
	errorLogByService := systemHealthCountDetails(extra.ErrorLog, "当前时间窗口错误日志数")
	badRequestsByService := systemHealthAlertDetails(extra.BadRequests)
	resourceShiftByService := systemHealthResourceShiftDetails(extra.ResourceShift, resourceByService)
	vmHostResourceShiftByService := systemHealthVMHostResourceShiftDetails(extra.VMHostResource.Scope, extra.VMHostResource.Shift, vmHostResourceByService)
	lockCostByService := systemHealthLockCostDetails(extra.LockCost, requestCountByService)
	for i := range system.Services {
		service := &system.Services[i]
		if detail, ok := resourceByService[service.Name]; ok {
			applySystemHealthResourceWaterline(service, detail)
		}
		if detail, ok := latencyByService[service.Name]; ok {
			detail = applySystemHealthLowRequestStatusCap(detail, service.Metrics[healthMetricLatency], requestCountByService[service.Name], systemHealthLatencyRedMinRequests)
			applySystemHealthCurrentValue(service, healthMetricLatency, detail)
		}
		if detail, ok := errorRateByService[service.Name]; ok {
			applySystemHealthCurrentValue(service, healthMetricErrorRate, detail)
		}
		if detail, ok := errorLogByService[service.Name]; ok {
			applySystemHealthCurrentValue(service, healthMetricErrorLog, detail)
		}
		if detail, ok := badRequestsByService[service.Name]; ok {
			applySystemHealthCurrentValue(service, healthMetricBadRequests, detail)
		}
		if detail, ok := resourceShiftByService[service.Name]; ok {
			applySystemHealthResourceShift(service, detail)
		}
		if detail, ok := vmHostResourceByService[service.Name]; ok {
			applySystemHealthResourceWaterline(service, detail)
		}
		if detail, ok := vmHostResourceShiftByService[service.Name]; ok {
			applySystemHealthResourceShift(service, detail)
		}
		if detail, ok := lockCostByService[service.Name]; ok {
			service.Metrics[healthMetricLockCost] = detail.Status
			service.MetricDetails[healthMetricLockCost] = detail
		}
		service.Score = scoreFromMetricStatuses(service.Metrics)
		service.Status = scoreHealthStatus(service.Score)
	}

	recalculateSystemHealthSystem(system)
	applySystemHealthSystemMetricDetails(system, extra)
	sortSystemHealthServicesByScore(system.Services)
}

func applySystemHealthResourceWaterline(service *response.SystemHealthService, detail response.SystemHealthMetricDetail) {
	current := service.MetricDetails[healthMetricResource]
	if _, nextOK := parseSystemMetricDetailRank(detail.Value); nextOK {
		if _, currentOK := parseSystemMetricDetailRank(current.Value); !currentOK {
			service.Metrics[healthMetricResource] = detail.Status
			service.MetricDetails[healthMetricResource] = detail
			return
		}
	}
	applySystemHealthMergedDetail(service, healthMetricResource, detail)
}

func applySystemHealthResourceShift(service *response.SystemHealthService, detail response.SystemHealthMetricDetail) {
	applySystemHealthMergedDetail(service, healthMetricResourceShift, detail)
}

func applySystemHealthMergedDetail(service *response.SystemHealthService, key string, detail response.SystemHealthMetricDetail) {
	currentStatus := service.Metrics[key]
	current := service.MetricDetails[key]
	if shouldKeepCurrentSystemHealthDetail(currentStatus, current, detail) {
		return
	}
	if currentStatus == "" {
		currentStatus = healthStatusGray
	}
	detail.Status = worseHealthStatus(currentStatus, detail.Status)
	service.Metrics[key] = detail.Status
	service.MetricDetails[key] = detail
}

func shouldKeepCurrentSystemHealthDetail(currentStatus string, current response.SystemHealthMetricDetail, next response.SystemHealthMetricDetail) bool {
	if currentStatus == "" {
		currentStatus = current.Status
	}
	if currentStatus == "" {
		currentStatus = healthStatusGray
	}
	nextStatus := next.Status
	if nextStatus == "" {
		nextStatus = healthStatusGray
	}

	mergedStatus := worseHealthStatus(currentStatus, nextStatus)
	if mergedStatus != nextStatus {
		return true
	}
	if mergedStatus != currentStatus {
		return false
	}

	currentRank, currentOK := parseSystemMetricDetailRank(current.Value)
	nextRank, nextOK := parseSystemMetricDetailRank(next.Value)
	if currentOK && nextOK {
		return currentRank > nextRank
	}
	if currentOK {
		return true
	}
	if nextOK {
		return false
	}
	return current.Value != "" && current.Value != "--"
}

func applySystemHealthCurrentValue(service *response.SystemHealthService, key string, detail response.SystemHealthMetricDetail) {
	status := service.Metrics[key]
	if status == "" {
		status = healthStatusGray
	}
	if detail.Status != "" {
		status = detail.Status
		service.Metrics[key] = status
	}
	detail.Status = status
	service.MetricDetails[key] = detail
}

func applySystemHealthLowRequestStatusCap(detail response.SystemHealthMetricDetail, currentStatus string, requestCount systemHealthMetricSample, minRequests float64) response.SystemHealthMetricDetail {
	if requestCount.OK {
		detail.Badges = append(detail.Badges, response.SystemHealthMetricBadge{
			Label:  "请求数",
			Value:  formatCountValue(requestCount.Value),
			Status: healthStatusGreen,
		})
	}
	status := detail.Status
	if status == "" {
		status = currentStatus
	}
	if status == healthStatusRed && requestCount.OK && requestCount.Value < minRequests {
		detail.Status = healthStatusYellow
		detail.Badges = append(detail.Badges, response.SystemHealthMetricBadge{
			Label:  "样本不足",
			Value:  "不一票红",
			Status: healthStatusYellow,
		})
	}
	return detail
}

func applySystemHealthSystemMetricDetails(system *response.SystemHealthSystem, extra systemHealthExtraMetricResults) {
	if system == nil {
		return
	}

	if detail, ok := systemHealthSystemResourceWaterlineDetail(extra.SystemResourceCPU, extra.SystemResourceMemory); ok {
		applySystemHealthSystemResourceWaterline(system, detail)
	}
	if detail, ok := systemHealthVMHostSystemResourceWaterlineDetail(extra.VMHostResource.Scope, extra.VMHostResource.CPU, extra.VMHostResource.Memory); ok {
		applySystemHealthSystemResourceWaterline(system, detail)
	}

	if detail, ok := systemHealthSingleLatencyDetail(extra.SystemLatency, extra.SystemLatencyDoD, systemHealthSingleCurrentValue(extra.SystemRequestCount)); ok {
		applySystemHealthSystemAggregateValue(system, healthMetricLatency, detail)
	}
	applySystemHealthSystemTrend(system, healthMetricLatency, "ms", systemHealthSingleSeries(extra.SystemLatencyTrend, nanosecondsToMilliseconds), systemHealthLatencyThresholds())

	if detail, ok := systemHealthSingleErrorRateDetail(extra.SystemErrorRate, "系统请求错误率", systemHealthSingleCurrentValue(extra.SystemRequestCount), systemHealthSingleCurrentValue(extra.SystemErrorCount)); ok {
		applySystemHealthSystemAggregateValue(system, healthMetricErrorRate, detail)
	}
	applySystemHealthSystemTrendWithLabel(
		system,
		healthMetricErrorRate,
		"%",
		"系统错误率",
		systemHealthSingleSeries(extra.SystemErrorRateTrend, ratioToPercent),
		systemHealthErrorRateThresholds(),
	)

	if count, ok := systemHealthCurrentSum(extra.ErrorLog); ok {
		applySystemHealthSystemAggregateValue(system, healthMetricErrorLog, systemHealthSystemErrorLogDetail(count, extra.SystemErrorLogDoD))
	}
	applySystemHealthSystemTrend(system, healthMetricErrorLog, "count", systemHealthSumSeries(extra.SystemErrorLogTrend, identityMetricValue), nil)

	if summary, ok := systemHealthSystemAlertSummary(extra.BadRequests); ok {
		applySystemHealthSystemAggregateValue(system, healthMetricBadRequests, systemHealthAlertDetail(summary, "系统当前活跃告警数"))
	}

	if detail, ok := systemHealthSystemResourceShiftDetail(extra.SystemResourceShift, system.MetricDetails[healthMetricResource]); ok {
		if vmDetail, vmOK := systemHealthSystemResourceShiftDetail(extra.VMHostResource.Shift, system.MetricDetails[healthMetricResource]); vmOK {
			detail = worseSystemHealthAggregateDetail(detail, vmDetail)
		}
		applySystemHealthSystemAggregateValue(system, healthMetricResourceShift, detail)
	} else if detail, ok := systemHealthSystemResourceShiftDetail(extra.VMHostResource.Shift, system.MetricDetails[healthMetricResource]); ok {
		applySystemHealthSystemAggregateValue(system, healthMetricResourceShift, detail)
	}
	applySystemHealthSystemTrend(
		system,
		healthMetricResourceShift,
		"%",
		systemHealthMaxResourceShiftSeries(extra.SystemResourceShiftTrend),
		systemHealthResourceShiftThresholds(),
	)
	if detail, ok := systemHealthSystemLockCostDetail(extra.LockCost, systemHealthCurrentValueByService(extra.RequestCount)); ok {
		applySystemHealthSystemAggregateValue(system, healthMetricLockCost, detail)
	}
	recalculateSystemHealthAggregateScore(system)
}

func applySystemHealthSystemResourceWaterline(system *response.SystemHealthSystem, detail response.SystemHealthMetricDetail) {
	currentStatus := system.Metrics[healthMetricResource]
	current := system.MetricDetails[healthMetricResource]
	if _, nextOK := parseSystemMetricDetailRank(detail.Value); nextOK {
		if _, currentOK := parseSystemMetricDetailRank(current.Value); !currentOK {
			system.Metrics[healthMetricResource] = detail.Status
			detail.Badges = mergeSystemHealthMetricBadges(detail.Badges, current.Badges)
			system.MetricDetails[healthMetricResource] = detail
			return
		}
	}
	if shouldKeepCurrentSystemHealthDetail(currentStatus, current, detail) {
		if merged, ok := supplementSystemHealthMetricDetailValue(current, detail); ok {
			system.MetricDetails[healthMetricResource] = merged
		}
		return
	}
	if currentStatus == "" {
		currentStatus = healthStatusGray
	}
	detail.Status = worseHealthStatus(currentStatus, detail.Status)
	detail.Badges = mergeSystemHealthMetricBadges(detail.Badges, current.Badges)
	system.Metrics[healthMetricResource] = detail.Status
	system.MetricDetails[healthMetricResource] = detail
}

func supplementSystemHealthMetricDetailValue(current response.SystemHealthMetricDetail, next response.SystemHealthMetricDetail) (response.SystemHealthMetricDetail, bool) {
	if _, currentOK := parseSystemMetricDetailRank(current.Value); currentOK {
		return current, false
	}

	_, nextOK := parseSystemMetricDetailRank(next.Value)
	if !nextOK {
		return current, false
	}

	current.Value = next.Value
	if current.Unit == "" {
		current.Unit = next.Unit
	}
	if current.Source == "" {
		current.Source = next.Source
	}
	if current.Note == "" {
		current.Note = next.Note
	}
	return current, true
}

func applySystemHealthSystemValue(system *response.SystemHealthSystem, key string, detail response.SystemHealthMetricDetail) {
	status := system.Metrics[key]
	if status == "" {
		status = healthStatusGray
	}
	if detail.Status != "" {
		status = worseHealthStatus(status, detail.Status)
		system.Metrics[key] = status
	}
	detail.Status = status
	if current, ok := system.MetricDetails[key]; ok {
		detail.Series = current.Series
		detail.Thresholds = current.Thresholds
		if detail.SeriesUnit == "" {
			detail.SeriesUnit = current.SeriesUnit
		}
		if detail.SeriesLabel == "" {
			detail.SeriesLabel = current.SeriesLabel
		}
		if detail.Unit == "" {
			detail.Unit = current.Unit
		}
		if detail.Source == "" {
			detail.Source = current.Source
		}
	}
	system.MetricDetails[key] = detail
}

func applySystemHealthSystemAggregateValue(system *response.SystemHealthSystem, key string, detail response.SystemHealthMetricDetail) {
	status := system.Metrics[key]
	if status == "" {
		status = healthStatusGray
	}
	if detail.Status != "" {
		status = detail.Status
		system.Metrics[key] = status
	}
	detail.Status = status
	if current, ok := system.MetricDetails[key]; ok {
		detail.Series = current.Series
		detail.Thresholds = current.Thresholds
		if detail.SeriesUnit == "" {
			detail.SeriesUnit = current.SeriesUnit
		}
		if detail.SeriesLabel == "" {
			detail.SeriesLabel = current.SeriesLabel
		}
		if detail.Unit == "" {
			detail.Unit = current.Unit
		}
		if detail.Source == "" {
			detail.Source = current.Source
		}
		if len(detail.Badges) == 0 {
			detail.Badges = current.Badges
		} else {
			detail.Badges = mergeSystemHealthMetricBadges(detail.Badges, current.Badges)
		}
	}
	system.MetricDetails[key] = detail
}

func mergeSystemHealthMetricBadges(primary []response.SystemHealthMetricBadge, secondary []response.SystemHealthMetricBadge) []response.SystemHealthMetricBadge {
	if len(primary) == 0 {
		if secondary == nil {
			return nil
		}
		return append([]response.SystemHealthMetricBadge(nil), secondary...)
	}
	if len(secondary) == 0 {
		return primary
	}
	merged := append([]response.SystemHealthMetricBadge(nil), primary...)
	seen := make(map[string]struct{}, len(merged)+len(secondary))
	for _, badge := range merged {
		seen[badge.Label+"\x00"+badge.Value] = struct{}{}
	}
	for _, badge := range secondary {
		key := badge.Label + "\x00" + badge.Value
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		merged = append(merged, badge)
	}
	return merged
}

func worseSystemHealthAggregateDetail(left response.SystemHealthMetricDetail, right response.SystemHealthMetricDetail) response.SystemHealthMetricDetail {
	leftStatus := left.Status
	if leftStatus == "" {
		leftStatus = healthStatusGray
	}
	rightStatus := right.Status
	if rightStatus == "" {
		rightStatus = healthStatusGray
	}
	if healthStatusSeverity(rightStatus) > healthStatusSeverity(leftStatus) {
		return right
	}
	if healthStatusSeverity(rightStatus) < healthStatusSeverity(leftStatus) {
		return left
	}
	leftRank, leftOK := parseSystemMetricDetailRank(left.Value)
	rightRank, rightOK := parseSystemMetricDetailRank(right.Value)
	if rightOK && (!leftOK || rightRank > leftRank) {
		return right
	}
	return left
}

func applySystemHealthSystemTrend(system *response.SystemHealthSystem, key string, unit string, series []response.SystemHealthMetricPoint, thresholds []response.SystemHealthMetricThreshold) {
	applySystemHealthSystemTrendWithLabel(system, key, unit, "", series, thresholds)
}

func applySystemHealthSystemTrendWithLabel(system *response.SystemHealthSystem, key string, unit string, label string, series []response.SystemHealthMetricPoint, thresholds []response.SystemHealthMetricThreshold) {
	if len(series) == 0 && len(thresholds) == 0 {
		return
	}
	detail := system.MetricDetails[key]
	if detail.Status == "" {
		detail.Status = system.Metrics[key]
	}
	if detail.Status == "" {
		detail.Status = healthStatusGray
	}
	if unit != "" {
		detail.SeriesUnit = unit
		if detail.Unit == "" {
			detail.Unit = unit
		}
	}
	if label != "" {
		detail.SeriesLabel = label
	}
	if len(series) > 0 {
		detail.Series = series
	}
	if len(thresholds) > 0 {
		detail.Thresholds = thresholds
	}
	system.MetricDetails[key] = detail
}

func systemHealthSingleValueDetail(results []prometheus.MetricResult, formatter func(float64) string, unit string, note string) (response.SystemHealthMetricDetail, bool) {
	for _, result := range results {
		if len(result.Values) == 0 {
			continue
		}
		value := result.Values[0].Value
		if !isValidSystemHealthMetricValue(value) {
			continue
		}
		return response.SystemHealthMetricDetail{
			Value: formatter(value),
			Unit:  unit,
			Note:  note,
		}, true
	}
	return response.SystemHealthMetricDetail{}, false
}

func systemHealthSingleLatencyDetail(results []prometheus.MetricResult, dodResults []prometheus.MetricResult, requestCount systemHealthMetricSample) (response.SystemHealthMetricDetail, bool) {
	for _, result := range results {
		if len(result.Values) == 0 {
			continue
		}
		value := result.Values[0].Value
		if !isValidSystemHealthMetricValue(value) {
			continue
		}
		status := systemHealthLatencyStatus(value)
		badges := []response.SystemHealthMetricBadge(nil)
		if changePercent, ok := systemHealthSingleChangePercent(dodResults); ok {
			status = worseHealthStatus(status, systemHealthChangeStatus(changePercent, systemHealthLatencyDoDYellow, systemHealthLatencyDoDRed))
			badges = append(badges, response.SystemHealthMetricBadge{
				Label:  "24h变化",
				Value:  formatPercentValue(changePercent),
				Status: systemHealthChangeStatus(changePercent, systemHealthLatencyDoDYellow, systemHealthLatencyDoDRed),
			})
		}
		detail := response.SystemHealthMetricDetail{
			Status: status,
			Value:  formatMillisecondsValue(value),
			Unit:   "ms",
			Note:   "系统请求平均响应延时",
			Badges: badges,
		}
		return applySystemHealthLowRequestStatusCap(detail, status, requestCount, systemHealthLatencyRedMinRequests), true
	}
	return response.SystemHealthMetricDetail{}, false
}

func systemHealthSingleErrorRateDetail(results []prometheus.MetricResult, note string, requestCount systemHealthMetricSample, errorCount systemHealthMetricSample) (response.SystemHealthMetricDetail, bool) {
	for _, result := range results {
		if len(result.Values) == 0 {
			continue
		}
		value := result.Values[0].Value
		if !isValidSystemHealthMetricValue(value) {
			continue
		}
		return response.SystemHealthMetricDetail{
			Status: systemHealthCurrentErrorRateStatus(value, requestCount, errorCount),
			Value:  formatPercentValue(value * 100),
			Unit:   "%",
			Note:   note,
			Badges: systemHealthTrafficBadges(requestCount, errorCount, systemHealthErrorRateRedHasLowSample(value, requestCount, errorCount)),
		}, true
	}
	return response.SystemHealthMetricDetail{}, false
}

func systemHealthSystemErrorLogDetail(count float64, dodResults []prometheus.MetricResult) response.SystemHealthMetricDetail {
	status := healthStatusGreen
	badges := []response.SystemHealthMetricBadge(nil)
	if count > 0 {
		status = healthStatusYellow
	}
	if changePercent, ok := systemHealthSingleChangePercent(dodResults); ok {
		status = systemHealthChangeStatus(changePercent, systemHealthErrorLogDoDYellow, systemHealthErrorLogDoDRed)
		badges = append(badges, response.SystemHealthMetricBadge{
			Label:  "24h变化",
			Value:  formatPercentValue(changePercent),
			Status: status,
		})
	}
	return response.SystemHealthMetricDetail{
		Status: status,
		Value:  formatCountValue(count),
		Unit:   "count",
		Note:   "系统窗口错误日志总数",
		Badges: badges,
	}
}

func systemHealthCurrentSum(results []prometheus.MetricResult) (float64, bool) {
	total := 0.0
	found := false
	for _, result := range results {
		if len(result.Values) == 0 {
			continue
		}
		value := result.Values[0].Value
		if !isValidSystemHealthMetricValue(value) {
			continue
		}
		total += value
		found = true
	}
	return total, found
}

func systemHealthSingleChangePercent(results []prometheus.MetricResult) (float64, bool) {
	for _, result := range results {
		if len(result.Values) == 0 {
			continue
		}
		value := result.Values[0].Value
		if math.IsNaN(value) || value < 0 {
			continue
		}
		return ratioToShiftPercent(value), true
	}
	return 0, false
}

func systemHealthSingleSeries(results []prometheus.MetricResult, transform func(float64) float64) []response.SystemHealthMetricPoint {
	for _, result := range results {
		series := systemHealthSeriesFromPoints(result.Values, transform)
		if len(series) > 0 {
			return series
		}
	}
	return nil
}

func systemHealthSumSeries(results []prometheus.MetricResult, transform func(float64) float64) []response.SystemHealthMetricPoint {
	valuesByTimestamp := make(map[int64]float64)
	for _, result := range results {
		for _, point := range result.Values {
			if !isValidSystemHealthMetricValue(point.Value) {
				continue
			}
			valuesByTimestamp[point.TimeStamp] += transform(point.Value)
		}
	}
	return systemHealthSortedSeries(valuesByTimestamp)
}

func systemHealthMaxSeries(results []prometheus.MetricResult, transform func(float64) float64) []response.SystemHealthMetricPoint {
	valuesByTimestamp := make(map[int64]float64)
	seen := make(map[int64]bool)
	for _, result := range results {
		for _, point := range result.Values {
			if !isValidSystemHealthMetricValue(point.Value) {
				continue
			}
			value := transform(point.Value)
			if !seen[point.TimeStamp] || value > valuesByTimestamp[point.TimeStamp] {
				valuesByTimestamp[point.TimeStamp] = value
				seen[point.TimeStamp] = true
			}
		}
	}
	return systemHealthSortedSeries(valuesByTimestamp)
}

func systemHealthMaxResourceShiftSeries(results []prometheus.MetricResult) []response.SystemHealthMetricPoint {
	valuesByTimestamp := make(map[int64]float64)
	seen := make(map[int64]bool)
	for _, result := range results {
		for _, point := range result.Values {
			if !isValidSystemHealthMetricValue(point.Value) || point.Value < 1 {
				continue
			}
			value := ratioToShiftPercent(point.Value)
			if !seen[point.TimeStamp] || value > valuesByTimestamp[point.TimeStamp] {
				valuesByTimestamp[point.TimeStamp] = value
				seen[point.TimeStamp] = true
			}
		}
	}
	return systemHealthSortedSeries(valuesByTimestamp)
}

func systemHealthSeriesFromPoints(points []prometheus.Points, transform func(float64) float64) []response.SystemHealthMetricPoint {
	valuesByTimestamp := make(map[int64]float64, len(points))
	for _, point := range points {
		if !isValidSystemHealthMetricValue(point.Value) {
			continue
		}
		valuesByTimestamp[point.TimeStamp] = transform(point.Value)
	}
	return systemHealthSortedSeries(valuesByTimestamp)
}

func systemHealthSortedSeries(valuesByTimestamp map[int64]float64) []response.SystemHealthMetricPoint {
	if len(valuesByTimestamp) == 0 {
		return nil
	}
	timestamps := make([]int64, 0, len(valuesByTimestamp))
	for timestamp := range valuesByTimestamp {
		timestamps = append(timestamps, timestamp)
	}
	sort.Slice(timestamps, func(i, j int) bool {
		return timestamps[i] < timestamps[j]
	})
	series := make([]response.SystemHealthMetricPoint, 0, len(timestamps))
	for _, timestamp := range timestamps {
		series = append(series, response.SystemHealthMetricPoint{
			Timestamp: timestamp,
			Value:     roundSystemHealthMetricValue(valuesByTimestamp[timestamp]),
		})
	}
	return series
}

func systemHealthResourceShiftThresholds() []response.SystemHealthMetricThreshold {
	return []response.SystemHealthMetricThreshold{
		{Label: "关注线", Value: 30, Status: healthStatusYellow},
		{Label: "异常线", Value: 60, Status: healthStatusRed},
	}
}

func systemHealthErrorRateThresholds() []response.SystemHealthMetricThreshold {
	return []response.SystemHealthMetricThreshold{
		{Label: "异常线", Value: systemHealthErrorRateRedThreshold * 100, Status: healthStatusRed},
	}
}

func systemHealthLatencyThresholds() []response.SystemHealthMetricThreshold {
	return []response.SystemHealthMetricThreshold{
		{Label: "关注线", Value: nanosecondsToMilliseconds(systemHealthLatencyYellowThreshold), Status: healthStatusYellow},
		{Label: "异常线", Value: nanosecondsToMilliseconds(systemHealthLatencyRedThreshold), Status: healthStatusRed},
	}
}

func systemHealthLatencyStatus(value float64) string {
	if math.IsInf(value, 0) || value >= systemHealthLatencyRedThreshold {
		return healthStatusRed
	}
	if value >= systemHealthLatencyYellowThreshold {
		return healthStatusYellow
	}
	return healthStatusGreen
}

func systemHealthChangeStatus(changePercent float64, yellowThreshold float64, redThreshold float64) string {
	if math.IsInf(changePercent, 0) || changePercent >= redThreshold {
		return healthStatusRed
	}
	if changePercent >= yellowThreshold {
		return healthStatusYellow
	}
	return healthStatusGreen
}

func nanosecondsToMilliseconds(value float64) float64 {
	return value / 1_000_000
}

func ratioToPercent(value float64) float64 {
	return value * 100
}

func ratioToShiftPercent(value float64) float64 {
	return (value - 1) * 100
}

func identityMetricValue(value float64) float64 {
	return value
}

func isValidSystemHealthMetricValue(value float64) bool {
	return !math.IsNaN(value) && value >= 0
}

func roundSystemHealthMetricValue(value float64) float64 {
	if math.IsInf(value, 0) {
		return value
	}
	return math.Round(value*100) / 100
}

type systemHealthMetricSample struct {
	Value float64
	OK    bool
}

func systemHealthCurrentValueByService(results []prometheus.MetricResult) map[string]systemHealthMetricSample {
	values := make(map[string]systemHealthMetricSample, len(results))
	for _, result := range results {
		if len(result.Values) == 0 || result.Metric.SvcName == "" {
			continue
		}
		value := result.Values[0].Value
		if !isValidSystemHealthMetricValue(value) {
			continue
		}
		values[result.Metric.SvcName] = systemHealthMetricSample{Value: value, OK: true}
	}
	return values
}

func systemHealthSingleCurrentValue(results []prometheus.MetricResult) systemHealthMetricSample {
	for _, result := range results {
		if len(result.Values) == 0 {
			continue
		}
		value := result.Values[0].Value
		if !isValidSystemHealthMetricValue(value) {
			continue
		}
		return systemHealthMetricSample{Value: value, OK: true}
	}
	return systemHealthMetricSample{}
}

func systemHealthLatencyDetails(results []prometheus.MetricResult) map[string]response.SystemHealthMetricDetail {
	details := make(map[string]response.SystemHealthMetricDetail, len(results))
	for _, result := range results {
		if len(result.Values) == 0 || result.Metric.SvcName == "" {
			continue
		}
		value := result.Values[0].Value
		if math.IsNaN(value) || value < 0 {
			continue
		}
		details[result.Metric.SvcName] = response.SystemHealthMetricDetail{
			Status: systemHealthLatencyStatus(value),
			Value:  formatMillisecondsValue(value),
			Note:   "当前平均响应延时",
		}
	}
	return details
}

func systemHealthErrorRateDetails(results []prometheus.MetricResult, requestCounts map[string]systemHealthMetricSample, errorCounts map[string]systemHealthMetricSample) map[string]response.SystemHealthMetricDetail {
	details := make(map[string]response.SystemHealthMetricDetail, len(results))
	for _, result := range results {
		if len(result.Values) == 0 || result.Metric.SvcName == "" {
			continue
		}
		value := result.Values[0].Value
		if math.IsNaN(value) || value < 0 {
			continue
		}
		requestCount := requestCounts[result.Metric.SvcName]
		errorCount := errorCounts[result.Metric.SvcName]
		details[result.Metric.SvcName] = response.SystemHealthMetricDetail{
			Status: systemHealthCurrentErrorRateStatus(value, requestCount, errorCount),
			Value:  formatPercentValue(value * 100),
			Note:   "当前请求错误率",
			Badges: systemHealthTrafficBadges(requestCount, errorCount, systemHealthErrorRateRedHasLowSample(value, requestCount, errorCount)),
		}
	}
	return details
}

func systemHealthCurrentErrorRateStatus(value float64, requestCount systemHealthMetricSample, errorCount systemHealthMetricSample) string {
	if requestCount.OK && requestCount.Value <= 0 {
		return healthStatusGray
	}
	if math.IsInf(value, 0) || value >= systemHealthErrorRateRedThreshold {
		if systemHealthErrorRateHasLowSample(requestCount, errorCount) {
			return healthStatusYellow
		}
		return healthStatusRed
	}
	if value > 0 {
		return healthStatusYellow
	}
	return healthStatusGreen
}

func systemHealthErrorRateRedHasLowSample(value float64, requestCount systemHealthMetricSample, errorCount systemHealthMetricSample) bool {
	if !math.IsInf(value, 0) && value < systemHealthErrorRateRedThreshold {
		return false
	}
	return systemHealthErrorRateHasLowSample(requestCount, errorCount)
}

func systemHealthErrorRateHasLowSample(requestCount systemHealthMetricSample, errorCount systemHealthMetricSample) bool {
	if requestCount.OK && requestCount.Value < systemHealthErrorRateRedMinRequests {
		return true
	}
	if errorCount.OK && errorCount.Value < systemHealthErrorRateRedMinErrors {
		return true
	}
	return false
}

func systemHealthTrafficBadges(requestCount systemHealthMetricSample, errorCount systemHealthMetricSample, lowSample bool) []response.SystemHealthMetricBadge {
	badges := make([]response.SystemHealthMetricBadge, 0, 3)
	if requestCount.OK {
		badges = append(badges, response.SystemHealthMetricBadge{
			Label:  "请求数",
			Value:  formatCountValue(requestCount.Value),
			Status: healthStatusGreen,
		})
	}
	if errorCount.OK {
		badges = append(badges, response.SystemHealthMetricBadge{
			Label:  "失败数",
			Value:  formatCountValue(errorCount.Value),
			Status: systemHealthErrorCountBadgeStatus(errorCount.Value),
		})
	}
	if lowSample {
		badges = append(badges, response.SystemHealthMetricBadge{
			Label:  "样本不足",
			Value:  "不一票红",
			Status: healthStatusYellow,
		})
	}
	return badges
}

func systemHealthErrorCountBadgeStatus(value float64) string {
	if value >= systemHealthErrorRateRedMinErrors {
		return healthStatusRed
	}
	if value > 0 {
		return healthStatusYellow
	}
	return healthStatusGreen
}

func systemHealthCountDetails(results []prometheus.MetricResult, note string) map[string]response.SystemHealthMetricDetail {
	details := make(map[string]response.SystemHealthMetricDetail, len(results))
	for _, result := range results {
		if len(result.Values) == 0 || result.Metric.SvcName == "" {
			continue
		}
		value := result.Values[0].Value
		if math.IsNaN(value) || value < 0 {
			continue
		}
		details[result.Metric.SvcName] = response.SystemHealthMetricDetail{
			Value: formatCountValue(value),
			Note:  note,
		}
	}
	return details
}

func systemHealthAlertDetails(summaries map[string]systemHealthAlertSummary) map[string]response.SystemHealthMetricDetail {
	details := make(map[string]response.SystemHealthMetricDetail, len(summaries))
	for serviceName, summary := range summaries {
		if serviceName == "" {
			continue
		}
		details[serviceName] = systemHealthAlertDetail(summary, "当前活跃告警数")
	}
	return details
}

func systemHealthSystemAlertSummary(summaries map[string]systemHealthAlertSummary) (systemHealthAlertSummary, bool) {
	if len(summaries) == 0 {
		return systemHealthAlertSummary{}, false
	}
	total := systemHealthAlertSummary{}
	for _, summary := range summaries {
		total = systemHealthMergeAlertSummary(total, summary)
	}
	return total, true
}

func systemHealthAlertDetail(summary systemHealthAlertSummary, note string) response.SystemHealthMetricDetail {
	return response.SystemHealthMetricDetail{
		Status: systemHealthAlertSummaryStatus(summary),
		Value:  formatCountValue(float64(summary.Count)),
		Unit:   "count",
		Note:   note,
		Badges: systemHealthAlertBadges(summary),
	}
}

func systemHealthAlertBadges(summary systemHealthAlertSummary) []response.SystemHealthMetricBadge {
	if summary.Count == 0 {
		return nil
	}
	badges := []response.SystemHealthMetricBadge{
		{Label: "红色告警", Value: formatCountValue(float64(summary.RedCount)), Status: healthStatusRed},
		{Label: "黄色告警", Value: formatCountValue(float64(summary.YellowCount)), Status: healthStatusYellow},
	}
	if summary.AffectedServices > 0 {
		badges = append(badges, response.SystemHealthMetricBadge{
			Label:  "影响服务",
			Value:  formatCountValue(float64(summary.AffectedServices)),
			Status: systemHealthAlertSummaryStatus(summary),
		})
	}
	return badges
}

func systemHealthResourceWaterlineDetails(cpuResults []prometheus.MetricResult, memoryResults []prometheus.MetricResult) map[string]response.SystemHealthMetricDetail {
	type resourceWaterlineCandidate struct {
		value float64
		note  string
	}

	candidates := make(map[string]resourceWaterlineCandidate, len(cpuResults)+len(memoryResults))
	collect := func(results []prometheus.MetricResult, note string) {
		for _, result := range results {
			if len(result.Values) == 0 || result.Metric.SvcName == "" {
				continue
			}
			value := result.Values[0].Value
			if math.IsNaN(value) || value < 0 {
				continue
			}
			current, ok := candidates[result.Metric.SvcName]
			if !ok || value > current.value {
				candidates[result.Metric.SvcName] = resourceWaterlineCandidate{
					value: value,
					note:  note,
				}
			}
		}
	}
	collect(cpuResults, "当前 CPU 资源水位")
	collect(memoryResults, "当前内存资源水位")

	details := make(map[string]response.SystemHealthMetricDetail, len(candidates))
	for serviceName, candidate := range candidates {
		status := healthStatusGreen
		if math.IsInf(candidate.value, 0) || candidate.value >= 0.9 {
			status = healthStatusRed
		} else if candidate.value >= 0.7 {
			status = healthStatusYellow
		}
		details[serviceName] = response.SystemHealthMetricDetail{
			Status: status,
			Value:  formatWaterlinePercentValue(candidate.value * 100),
			Note:   candidate.note,
		}
	}
	return details
}

func systemHealthSystemResourceWaterlineDetail(cpuResults []prometheus.MetricResult, memoryResults []prometheus.MetricResult) (response.SystemHealthMetricDetail, bool) {
	samples := systemHealthSystemResourceWaterlineSamples(cpuResults, memoryResults)
	return systemHealthAggregateResourceWaterlineDetail(samples)
}

func systemHealthResourceSource(labels prometheus.Labels) string {
	parts := make([]string, 0, 3)
	if labels.ClusterID != "" {
		parts = append(parts, labels.ClusterID)
	}
	if labels.Namespace != "" {
		parts = append(parts, labels.Namespace)
	}
	if labels.POD != "" {
		parts = append(parts, labels.POD)
	} else if labels.NodeName != "" {
		parts = append(parts, labels.NodeName)
	}
	return strings.Join(parts, "/")
}

type systemHealthResourceWaterlineSample struct {
	source string
	value  float64
}

func systemHealthSystemResourceWaterlineSamples(cpuResults []prometheus.MetricResult, memoryResults []prometheus.MetricResult) []systemHealthResourceWaterlineSample {
	valuesBySource := make(map[string]float64, len(cpuResults)+len(memoryResults))
	collect := func(results []prometheus.MetricResult, fallbackPrefix string) {
		for index, result := range results {
			if len(result.Values) == 0 {
				continue
			}
			value := result.Values[0].Value
			if !isValidSystemHealthMetricValue(value) {
				continue
			}
			source := systemHealthResourceSource(result.Metric)
			if source == "" {
				source = fmt.Sprintf("%s-sample-%d", fallbackPrefix, index)
			}
			percent := clampSystemHealthPercent(value * 100)
			if current, ok := valuesBySource[source]; !ok || percent > current {
				valuesBySource[source] = percent
			}
		}
	}
	collect(cpuResults, "cpu")
	collect(memoryResults, "memory")

	sources := make([]string, 0, len(valuesBySource))
	for source := range valuesBySource {
		sources = append(sources, source)
	}
	sort.Strings(sources)
	samples := make([]systemHealthResourceWaterlineSample, 0, len(sources))
	for _, source := range sources {
		samples = append(samples, systemHealthResourceWaterlineSample{
			source: source,
			value:  valuesBySource[source],
		})
	}
	return samples
}

func systemHealthAggregateResourceWaterlineDetail(samples []systemHealthResourceWaterlineSample) (response.SystemHealthMetricDetail, bool) {
	if len(samples) == 0 {
		return response.SystemHealthMetricDetail{}, false
	}

	values := make([]float64, 0, len(samples))
	total := 0.0
	highCount := 0
	redCount := 0
	for _, sample := range samples {
		value := clampSystemHealthPercent(sample.value)
		values = append(values, value)
		total += value
		if value >= systemHealthResourceWaterlineYellow {
			highCount++
		}
		if value >= systemHealthResourceWaterlineRed {
			redCount++
		}
	}
	avg := total / float64(len(values))
	p95 := percentileSystemHealthValue(values, 0.95)
	displayValue := math.Max(avg, p95)
	status := systemHealthAggregateResourceWaterlineStatus(avg, p95)

	badges := []response.SystemHealthMetricBadge{
		{Label: "平均水位", Value: formatWaterlinePercentValue(avg), Status: systemHealthResourceWaterlineStatus(avg)},
		{Label: "高水位实例", Value: fmt.Sprintf("%d/%d", highCount, len(values)), Status: systemHealthResourceWaterlineCountStatus(highCount, redCount)},
		{Label: "P95水位", Value: formatWaterlinePercentValue(p95), Status: systemHealthResourceWaterlineStatus(p95)},
	}
	if redCount > 0 {
		badges = append(badges, response.SystemHealthMetricBadge{
			Label:  "异常实例",
			Value:  fmt.Sprintf("%d/%d", redCount, len(values)),
			Status: healthStatusRed,
		})
	}

	return response.SystemHealthMetricDetail{
		Status: status,
		Value:  formatWaterlinePercentValue(displayValue),
		Unit:   "%",
		Note:   "系统资源水位聚合",
		Badges: badges,
	}, true
}

func percentileSystemHealthValue(values []float64, percentile float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	if percentile <= 0 {
		return sorted[0]
	}
	if percentile >= 1 {
		return sorted[len(sorted)-1]
	}
	index := int(math.Ceil(percentile*float64(len(sorted)))) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(sorted) {
		index = len(sorted) - 1
	}
	return sorted[index]
}

func systemHealthAggregateResourceWaterlineStatus(avg float64, p95 float64) string {
	if math.IsInf(avg, 0) || math.IsInf(p95, 0) || avg >= systemHealthResourceWaterlineRed || p95 >= systemHealthResourceWaterlineRed {
		return healthStatusRed
	}
	if avg >= systemHealthResourceWaterlineYellow || p95 >= systemHealthResourceWaterlineYellow {
		return healthStatusYellow
	}
	return healthStatusGreen
}

func systemHealthResourceWaterlineStatus(value float64) string {
	if math.IsInf(value, 0) || value >= systemHealthResourceWaterlineRed {
		return healthStatusRed
	}
	if value >= systemHealthResourceWaterlineYellow {
		return healthStatusYellow
	}
	return healthStatusGreen
}

func systemHealthResourceWaterlineCountStatus(highCount int, redCount int) string {
	if redCount > 0 {
		return healthStatusRed
	}
	if highCount > 0 {
		return healthStatusYellow
	}
	return healthStatusGreen
}

func clampSystemHealthPercent(value float64) float64 {
	if math.IsInf(value, 0) || value > 100 {
		return 100
	}
	if math.IsNaN(value) || value < 0 {
		return 0
	}
	return value
}

type systemHealthVMHostResourceCandidate struct {
	value  float64
	note   string
	source string
}

func systemHealthVMHostResourceWaterlineDetails(scope systemHealthVMHostResourceScope, cpuResults []prometheus.MetricResult, memoryResults []prometheus.MetricResult) map[string]response.SystemHealthMetricDetail {
	candidates := systemHealthVMHostResourceWaterlineCandidates(cpuResults, memoryResults)
	details := make(map[string]response.SystemHealthMetricDetail, len(scope.NodesByService))
	for serviceName := range scope.NodesByService {
		detail, ok := systemHealthVMHostDetailForServiceNodes(scope.sortedServiceNodes(serviceName), candidates, 70, 90)
		if ok {
			details[serviceName] = detail
		}
	}
	return details
}

func systemHealthVMHostSystemResourceWaterlineDetail(scope systemHealthVMHostResourceScope, cpuResults []prometheus.MetricResult, memoryResults []prometheus.MetricResult) (response.SystemHealthMetricDetail, bool) {
	candidates := systemHealthVMHostResourceWaterlineCandidates(cpuResults, memoryResults)
	samples := make([]systemHealthResourceWaterlineSample, 0, len(scope.SystemNodes))
	for _, nodeName := range scope.sortedNodes() {
		candidate, ok := candidates[nodeName]
		if !ok {
			continue
		}
		samples = append(samples, systemHealthResourceWaterlineSample{
			source: candidate.source,
			value:  candidate.value,
		})
	}
	return systemHealthAggregateResourceWaterlineDetail(samples)
}

func systemHealthVMHostResourceWaterlineCandidates(cpuResults []prometheus.MetricResult, memoryResults []prometheus.MetricResult) map[string]systemHealthVMHostResourceCandidate {
	candidates := make(map[string]systemHealthVMHostResourceCandidate, len(cpuResults)+len(memoryResults))
	collect := func(results []prometheus.MetricResult, note string) {
		for _, result := range results {
			if len(result.Values) == 0 {
				continue
			}
			nodeName := systemHealthVMHostNodeName(result.Metric)
			if nodeName == "" {
				continue
			}
			value := result.Values[0].Value
			if !isValidSystemHealthMetricValue(value) {
				continue
			}
			percent := value * 100
			current, ok := candidates[nodeName]
			if !ok || percent > current.value {
				candidates[nodeName] = systemHealthVMHostResourceCandidate{
					value:  percent,
					note:   note,
					source: systemHealthVMHostResourceSource(nodeName),
				}
			}
		}
	}
	collect(cpuResults, "Current host CPU resource waterline")
	collect(memoryResults, "Current host memory resource waterline")
	return candidates
}

func systemHealthVMHostResourceShiftDetails(scope systemHealthVMHostResourceScope, results []prometheus.MetricResult, waterlineDetails map[string]response.SystemHealthMetricDetail) map[string]response.SystemHealthMetricDetail {
	candidates := systemHealthVMHostResourceShiftCandidates(results)
	details := make(map[string]response.SystemHealthMetricDetail, len(scope.NodesByService))
	for serviceName := range scope.NodesByService {
		detail, ok := systemHealthVMHostDetailForServiceNodes(scope.sortedServiceNodes(serviceName), candidates, 30, 60)
		if ok {
			details[serviceName] = capSystemHealthResourceShiftByWaterline(detail, waterlineDetails[serviceName])
		}
	}
	return details
}

func systemHealthVMHostResourceShiftCandidates(results []prometheus.MetricResult) map[string]systemHealthVMHostResourceCandidate {
	candidates := make(map[string]systemHealthVMHostResourceCandidate, len(results))
	for _, result := range results {
		if len(result.Values) == 0 {
			continue
		}
		nodeName := systemHealthVMHostNodeName(result.Metric)
		if nodeName == "" {
			continue
		}
		value := result.Values[0].Value
		if math.IsNaN(value) || value < 1 {
			continue
		}
		shiftPercent := ratioToShiftPercent(value)
		current, ok := candidates[nodeName]
		if !ok || shiftPercent > current.value {
			candidates[nodeName] = systemHealthVMHostResourceCandidate{
				value:  shiftPercent,
				note:   "Host CPU/memory resource shift against 24h baseline",
				source: systemHealthVMHostResourceSource(nodeName),
			}
		}
	}
	return candidates
}

func systemHealthVMHostDetailForServiceNodes(nodes []string, candidates map[string]systemHealthVMHostResourceCandidate, yellowThreshold float64, redThreshold float64) (response.SystemHealthMetricDetail, bool) {
	var selected *systemHealthVMHostResourceCandidate
	for _, nodeName := range nodes {
		candidate, ok := candidates[nodeName]
		if !ok {
			continue
		}
		if selected == nil || candidate.value > selected.value {
			selected = &candidate
		}
	}
	if selected == nil {
		return response.SystemHealthMetricDetail{}, false
	}

	status := healthStatusGreen
	if math.IsInf(selected.value, 0) || selected.value >= redThreshold {
		status = healthStatusRed
	} else if selected.value >= yellowThreshold {
		status = healthStatusYellow
	}
	return response.SystemHealthMetricDetail{
		Status: status,
		Value:  formatWaterlinePercentValue(selected.value),
		Unit:   "%",
		Source: selected.source,
		Note:   selected.note,
	}, true
}

func systemHealthVMHostNodeName(labels prometheus.Labels) string {
	return labels.NodeName
}

func systemHealthVMHostResourceSource(nodeName string) string {
	if nodeName == "" {
		return ""
	}
	return "host/" + nodeName
}

func systemHealthResourceShiftDetails(results []prometheus.MetricResult, waterlineDetails map[string]response.SystemHealthMetricDetail) map[string]response.SystemHealthMetricDetail {
	details := make(map[string]response.SystemHealthMetricDetail, len(results))
	for _, result := range results {
		if len(result.Values) == 0 || result.Metric.SvcName == "" {
			continue
		}
		value := result.Values[0].Value
		if math.IsNaN(value) || value < 1 {
			continue
		}
		shiftPercent := (value - 1) * 100
		status := healthStatusGreen
		if math.IsInf(value, 0) || shiftPercent >= 60 {
			status = healthStatusRed
		} else if shiftPercent >= 30 {
			status = healthStatusYellow
		}
		detail := response.SystemHealthMetricDetail{
			Status: status,
			Value:  formatPercentValue(shiftPercent),
			Note:   "CPU/内存资源相对 24h 前基线波动",
		}
		details[result.Metric.SvcName] = capSystemHealthResourceShiftByWaterline(detail, waterlineDetails[result.Metric.SvcName])
	}
	return details
}

func systemHealthSystemResourceShiftDetail(results []prometheus.MetricResult, waterline response.SystemHealthMetricDetail) (response.SystemHealthMetricDetail, bool) {
	selected := 0.0
	found := false
	for _, result := range results {
		if len(result.Values) == 0 {
			continue
		}
		value := result.Values[0].Value
		if !isValidSystemHealthMetricValue(value) || value < 1 {
			continue
		}
		shiftPercent := ratioToShiftPercent(value)
		if !found || shiftPercent > selected {
			selected = shiftPercent
			found = true
		}
	}
	if !found {
		return response.SystemHealthMetricDetail{}, false
	}

	status := healthStatusGreen
	if math.IsInf(selected, 0) || selected >= 60 {
		status = healthStatusRed
	} else if selected >= 30 {
		status = healthStatusYellow
	}
	detail := response.SystemHealthMetricDetail{
		Status: status,
		Value:  formatPercentValue(selected),
		Unit:   "%",
		Note:   "系统资源消耗相对 24h 前基线波动",
	}
	return capSystemHealthResourceShiftByWaterline(detail, waterline), true
}

func capSystemHealthResourceShiftByWaterline(detail response.SystemHealthMetricDetail, waterline response.SystemHealthMetricDetail) response.SystemHealthMetricDetail {
	waterlinePercent, ok := parseSystemMetricDetailRank(waterline.Value)
	if !ok {
		return detail
	}
	detail.Badges = append(detail.Badges, response.SystemHealthMetricBadge{
		Label:  "当前水位",
		Value:  formatWaterlinePercentValue(waterlinePercent),
		Status: waterline.Status,
	})
	if detail.Status == healthStatusRed && waterlinePercent < systemHealthResourceShiftRedMinLevel {
		detail.Status = healthStatusYellow
		detail.Badges = append(detail.Badges, response.SystemHealthMetricBadge{
			Label:  "低水位波动",
			Value:  "不一票红",
			Status: healthStatusYellow,
		})
	}
	return detail
}

func systemHealthLockCostDetails(results []prometheus.MetricResult, requestCounts map[string]systemHealthMetricSample) map[string]response.SystemHealthMetricDetail {
	details := make(map[string]response.SystemHealthMetricDetail, len(results))
	for _, result := range results {
		if len(result.Values) == 0 || result.Metric.SvcName == "" {
			continue
		}
		value := result.Values[0].Value
		if math.IsNaN(value) || value < 0 {
			continue
		}
		status := healthStatusGreen
		if math.IsInf(value, 0) || value >= 20_000_000 {
			status = healthStatusRed
		} else if value >= 5_000_000 {
			status = healthStatusYellow
		}
		detail := response.SystemHealthMetricDetail{
			Status: status,
			Value:  formatMillisecondsPerRequest(value),
			Note:   "由 futex 锁等待耗时按请求数均摊计算",
		}
		details[result.Metric.SvcName] = applySystemHealthLowRequestStatusCap(detail, status, requestCounts[result.Metric.SvcName], systemHealthLockCostRedMinRequests)
	}
	return details
}

func systemHealthSystemLockCostDetail(results []prometheus.MetricResult, requestCounts map[string]systemHealthMetricSample) (response.SystemHealthMetricDetail, bool) {
	totalRequests := 0.0
	totalLockCost := 0.0
	for _, result := range results {
		if len(result.Values) == 0 || result.Metric.SvcName == "" {
			continue
		}
		value := result.Values[0].Value
		if math.IsNaN(value) || value < 0 {
			continue
		}
		requestCount := requestCounts[result.Metric.SvcName]
		if !requestCount.OK || requestCount.Value <= 0 {
			continue
		}
		totalRequests += requestCount.Value
		totalLockCost += value * requestCount.Value
	}
	if totalRequests <= 0 {
		return response.SystemHealthMetricDetail{}, false
	}

	value := totalLockCost / totalRequests
	status := healthStatusGreen
	if math.IsInf(value, 0) || value >= 20_000_000 {
		status = healthStatusRed
	} else if value >= 5_000_000 {
		status = healthStatusYellow
	}
	detail := response.SystemHealthMetricDetail{
		Status: status,
		Value:  formatMillisecondsPerRequest(value),
		Unit:   "ms/request",
		Note:   "系统 futex 锁等待耗时按请求数均摊",
		Badges: []response.SystemHealthMetricBadge{
			{
				Label:  "总锁等待",
				Value:  formatMillisecondsValue(totalLockCost),
				Status: status,
			},
		},
	}
	detail = applySystemHealthLowRequestStatusCap(detail, status, systemHealthMetricSample{Value: totalRequests, OK: true}, systemHealthLockCostRedMinRequests)
	return detail, true
}

func recalculateSystemHealthSystem(system *response.SystemHealthSystem) {
	if system == nil {
		return
	}

	system.Score = 0
	system.Status = healthStatusGray
	system.Metrics = defaultSystemMetricStatus()
	system.MetricDetails = defaultSystemMetricDetails()
	for _, key := range systemHealthMetricKeys {
		if detail, ok := systemMetricContributorOnlyDetail(system.Services, key); ok {
			system.MetricDetails[key] = detail
		}
	}
}

func recalculateSystemHealthAggregateScore(system *response.SystemHealthSystem) {
	if system == nil {
		return
	}
	score, ok := scoreFromSystemMetricStatuses(system.Metrics)
	if !ok {
		system.Score = 0
		system.Status = healthStatusGray
		return
	}
	system.Score = score
	system.Status = scoreHealthStatus(score)
}

func scoreFromMetricStatuses(metrics map[string]string) int {
	score, _ := scoreFromMetricStatusesWithData(metrics)
	return score
}

func scoreFromSystemMetricStatuses(metrics map[string]string) (int, bool) {
	return scoreFromMetricStatusesWithData(metrics)
}

func scoreFromMetricStatusesWithData(metrics map[string]string) (int, bool) {
	total := 0
	count := 0
	for _, key := range systemHealthMetricKeys {
		switch metrics[key] {
		case healthStatusGreen:
			total += 100
			count++
		case healthStatusYellow:
			total += 66
			count++
		case healthStatusRed:
			count++
		}
	}
	if count == 0 {
		return 0, false
	}
	return int(math.Round(float64(total) / float64(count))), true
}

func formatPercentValue(value float64) string {
	if math.IsInf(value, 0) {
		return "Inf%"
	}
	return fmt.Sprintf("%.2f%%", value)
}

func formatWaterlinePercentValue(value float64) string {
	if math.IsNaN(value) {
		return "--"
	}
	if math.IsInf(value, 0) || value > 100 {
		value = 100
	}
	if value < 0 {
		value = 0
	}
	return fmt.Sprintf("%.2f%%", value)
}

func formatMillisecondsValue(value float64) string {
	if math.IsInf(value, 0) {
		return "Infms"
	}
	milliseconds := value / 1_000_000
	if milliseconds >= 10 {
		return fmt.Sprintf("%.0fms", milliseconds)
	}
	return fmt.Sprintf("%.2fms", milliseconds)
}

func formatMillisecondsPerRequest(value float64) string {
	if math.IsInf(value, 0) {
		return "Inf ms/请求"
	}
	return fmt.Sprintf("%.2f ms/请求", value/1_000_000)
}

func formatCountValue(value float64) string {
	if math.IsInf(value, 0) {
		return "Inf"
	}
	rounded := math.Round(value)
	if math.Abs(value-rounded) < 0.005 {
		return fmt.Sprintf("%.0f", rounded)
	}
	return fmt.Sprintf("%.2f", value)
}
