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
	healthMetricBadRequests   = "badRequests"
	healthMetricLockCost      = "lockCost"

	healthStatusGreen  = "green"
	healthStatusYellow = "yellow"
	healthStatusRed    = "red"
	healthStatusGray   = "gray"

	systemHealthOverviewGroupConcurrency = 4
	systemHealthOverviewCurrentTimeout   = 2 * time.Second
	systemHealthOverviewHistoryTimeout   = time.Second
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
	healthMetricBadRequests:   "异常请求",
	healthMetricLockCost:      "锁等待",
}

var rygMetricToSystemMetric = map[string]string{
	"replica":       healthMetricResource,
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

	stageStartedAt := time.Now()
	groups, err := h.visibleSystemGroups(ctx)
	if err != nil {
		systemHealthWarn("system-health overview visible groups failed", systemHealthStageFields(ctx, stageStartedAt, "02_visible_groups",
			zap.Error(err),
		)...)
		return response.SystemHealthOverviewResp{}, err
	}
	systemHealthInfo("system-health overview visible groups loaded", systemHealthStageFields(ctx, stageStartedAt, "02_visible_groups",
		zap.Int("groups", len(groups)),
	)...)

	stageStartedAt = time.Now()
	systems, err := h.buildOverviewSystems(ctx, startTime, endTime, groups, "current")
	if err != nil {
		systemHealthWarn("system-health overview systems failed", systemHealthStageFields(ctx, stageStartedAt, "03_current_systems",
			zap.Error(err),
		)...)
		return response.SystemHealthOverviewResp{}, err
	}
	systemHealthInfo("system-health overview systems built", systemHealthStageFields(ctx, stageStartedAt, "03_current_systems",
		zap.Int("systems", len(systems)),
	)...)

	summary := buildSystemHealthOverviewSummary(systems)
	recentStartTime, previousStartTime, previousEndTime := newAbnormalMetricWindows(endTime)
	stageStartedAt = time.Now()
	recentSystems, recentErr := h.buildOverviewSystems(ctx, recentStartTime, endTime, groups, "recent")
	if recentErr != nil {
		systemHealthWarn("system-health overview recent systems failed", systemHealthStageFields(ctx, stageStartedAt, "04_recent_systems",
			zap.Error(recentErr),
		)...)
	} else {
		systemHealthInfo("system-health overview recent systems built", systemHealthStageFields(ctx, stageStartedAt, "04_recent_systems",
			zap.Int("systems", len(recentSystems)),
		)...)
	}

	stageStartedAt = time.Now()
	previousSystems, previousErr := h.buildOverviewSystems(ctx, previousStartTime, previousEndTime, groups, "previous")
	if previousErr != nil {
		systemHealthWarn("system-health overview previous systems failed", systemHealthStageFields(ctx, stageStartedAt, "05_previous_systems",
			zap.Error(previousErr),
		)...)
	} else {
		systemHealthInfo("system-health overview previous systems built", systemHealthStageFields(ctx, stageStartedAt, "05_previous_systems",
			zap.Int("systems", len(previousSystems)),
		)...)
	}

	if recentErr == nil && previousErr == nil {
		summary.NewAbnormalMetrics = countNewAbnormalMetrics(recentSystems, previousSystems)
	}

	resp := response.SystemHealthOverviewResp{
		Summary:       summary,
		Systems:       systems,
		PendingIssues: buildSystemHealthPendingIssues(systems, 20),
		Window:        window,
		GeneratedAt:   time.Now().Format(time.RFC3339),
		Source:        systemHealthSourceDataGroup,
	}
	systemHealthInfo("system-health overview build completed", systemHealthStageFields(ctx, buildStartedAt, "06_build_completed",
		zap.Int("systems", len(resp.Systems)),
		zap.Int("pendingIssues", len(resp.PendingIssues)),
	)...)
	return resp, nil
}

func (h *handler) buildOverviewSystems(ctx core.Context, startTime time.Time, endTime time.Time, groups []*datagroup.DataGroupTreeNode, windowLabel string) ([]response.SystemHealthSystem, error) {
	rygService := serviceoverview.New(zap.NewNop(), h.chRepo, h.dbRepo, h.promRepo)
	workerCtx := newSystemHealthOverviewWorkerContext(ctx)
	extraMetricTimeout := systemHealthOverviewExtraMetricTimeout(windowLabel)
	return buildOverviewSystemsForGroups(groups, systemHealthOverviewGroupConcurrency, func(group *datagroup.DataGroupTreeNode) (response.SystemHealthSystem, error) {
		stageStartedAt := time.Now()
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
			return response.SystemHealthSystem{}, err
		}
		systemHealthInfo("system-health overview group ryg built", systemHealthStageFields(workerCtx, stageStartedAt, "build_group_ryg",
			zap.String("windowLabel", windowLabel),
			zap.Int64("groupId", group.GroupID),
			zap.String("groupName", group.GroupName),
			zap.Int("services", len(light.ServiceList)),
		)...)

		system := buildSystemHealthSystem(group.DataGroup, light.ServiceList)
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

	var matched *datagroup.DataGroupTreeNode
	groups, err := h.visibleSystemGroups(ctx)
	if err != nil {
		return response.SystemHealthSystemDetailResp{}, err
	}
	for _, group := range groups {
		if group.GroupID == req.GroupID {
			matched = group
			break
		}
	}
	if matched == nil {
		return response.SystemHealthSystemDetailResp{}, fmt.Errorf("system group %d not found or no permission", req.GroupID)
	}

	rygService := serviceoverview.New(zap.NewNop(), h.chRepo, h.dbRepo, h.promRepo)
	light, err := rygService.GetServicesRYGLightStatus(ctx, &request.GetRygLightRequest{
		StartTime: startTime.UnixMicro(),
		EndTime:   endTime.UnixMicro(),
		GroupID:   matched.GroupID,
	})
	if err != nil {
		return response.SystemHealthSystemDetailResp{}, err
	}

	system := buildSystemHealthSystem(matched.DataGroup, light.ServiceList)
	extraMetrics, err := h.querySystemHealthExtraMetrics(ctx, startTime, endTime, matched.GroupID, system.Services)
	if err == nil {
		applySystemHealthExtraMetrics(&system, extraMetrics)
	}

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

func buildSystemHealthSystem(group datagroup.DataGroup, services []*response.ServiceRYGResult) response.SystemHealthSystem {
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

	if len(services) == 0 {
		return system
	}

	scoreTotal := 0
	serviceCount := 0
	aggregateMetrics := make(map[string]string, len(systemHealthMetricKeys))
	for _, service := range services {
		if service == nil {
			continue
		}
		serviceHealth := buildSystemHealthService(service)
		system.Services = append(system.Services, serviceHealth)
		scoreTotal += serviceHealth.Score
		serviceCount++
		for key, status := range serviceHealth.Metrics {
			if status == healthStatusGray {
				continue
			}
			aggregateMetrics[key] = worseHealthStatus(aggregateMetrics[key], status)
		}
	}

	if serviceCount == 0 {
		return system
	}

	system.Score = int(math.Round(float64(scoreTotal) / float64(serviceCount)))
	system.Status = scoreHealthStatus(system.Score)
	for _, key := range systemHealthMetricKeys {
		if status, ok := aggregateMetrics[key]; ok {
			system.Metrics[key] = status
			system.MetricDetails[key] = systemMetricDetail(system.Services, key, status)
		}
	}

	return system
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

func buildSystemHealthOverviewSummary(systems []response.SystemHealthSystem, previousSystems ...[]response.SystemHealthSystem) response.SystemHealthOverviewSummary {
	summary := response.SystemHealthOverviewSummary{TotalSystems: len(systems)}
	for _, system := range systems {
		summary.Services += len(system.Services)
		if len(system.Services) == 0 {
			summary.NoDataSystems++
		}
		if len(system.Services) > 0 {
			if system.Score < 60 {
				summary.LowScoreSystems++
			}
			if system.Score < 80 || hasAbnormalMetric(system.Metrics) {
				summary.AbnormalSystems++
			}
		}
		for _, service := range system.Services {
			if service.Score < 80 || hasAbnormalMetric(service.Metrics) {
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

func hasAbnormalMetric(metrics map[string]string) bool {
	for _, status := range metrics {
		if isAbnormalMetricStatus(status) {
			return true
		}
	}
	return false
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

func scoreHealthStatus(score int) string {
	if score < 60 {
		return healthStatusRed
	}
	if score < 80 {
		return healthStatusYellow
	}
	return healthStatusGreen
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
	return response.SystemHealthMetricDetail{
		Status: status,
		Value:  systemMetricRepresentativeValue(services, key, status),
		Note:   systemMetricNote(key, status),
	}
}

func systemMetricRepresentativeValue(services []response.SystemHealthService, key string, status string) string {
	type candidate struct {
		value string
		rank  float64
	}

	var selected *candidate
	for _, service := range services {
		if service.Metrics[key] != status {
			continue
		}
		detail := service.MetricDetails[key]
		rank, ok := parseSystemMetricDetailRank(detail.Value)
		if !ok {
			continue
		}
		if selected == nil || rank > selected.rank {
			next := candidate{value: detail.Value, rank: rank}
			selected = &next
		}
	}
	if selected != nil {
		return selected.value
	}
	return "--"
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
		return statusMetricNote(status, "资源副本健康稳定", "资源副本健康需要关注", "资源副本健康存在风险")
	case healthMetricResourceShift:
		return statusMetricNote(status, "资源波动处于稳定区间", "资源波动接近异常阈值", "资源波动超过异常阈值")
	case healthMetricErrorRate:
		return statusMetricNote(status, "错误率处于稳定区间", "错误率接近异常阈值", "错误率超过异常阈值")
	case healthMetricLatency:
		return statusMetricNote(status, "响应延时稳定", "响应延时升高", "响应延时超过异常阈值")
	case healthMetricErrorLog:
		return statusMetricNote(status, "错误日志数稳定", "错误日志数波动增加", "错误日志数超过异常阈值")
	case healthMetricBadRequests:
		return statusMetricNote(status, "异常请求处于稳定区间", "异常请求数升高", "异常请求超过异常阈值")
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
	ErrorLog                 []prometheus.MetricResult
	BadRequests              []prometheus.MetricResult
	ResourceShift            []prometheus.MetricResult
	LockCost                 []prometheus.MetricResult
	SystemLatency            []prometheus.MetricResult
	SystemErrorRate          []prometheus.MetricResult
	SystemLatencyTrend       []prometheus.MetricResult
	SystemErrorRateTrend     []prometheus.MetricResult
	SystemErrorLogTrend      []prometheus.MetricResult
	SystemResourceShiftTrend []prometheus.MetricResult
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
	if errorRate, err := h.promRepo.QueryMetricsWithPQLFilter(ctx, prometheus.PQLAvgErrorRateWithPQLFilter, startTime.UnixMicro(), endTime.UnixMicro(), prometheus.SVCGranularity, groupFilter); err == nil {
		result.ErrorRate = errorRate
	}
	if errorRate, err := h.promRepo.QueryMetricsWithPQLFilter(ctx, prometheus.PQLAvgErrorRateWithPQLFilter, startTime.UnixMicro(), endTime.UnixMicro(), prometheus.Granularity(""), groupFilter); err == nil {
		result.SystemErrorRate = errorRate
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
	resourceShiftTrendQuery := buildSystemHealthResourceShiftPQL(groupFilter, prometheus.VecFromDuration(trendStep))
	if resourceShiftTrend, err := h.promRepo.QueryRangeData(ctx, startTime, endTime, resourceShiftTrendQuery, trendStep); err == nil {
		result.SystemResourceShiftTrend = resourceShiftTrend
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

func buildSystemHealthVMHostCPUWaterlinePQL(nodes []string, window string, offset string) string {
	filter := systemHealthVMHostNodeFilter(nodes)
	filter.AddPatternFilter("mode!~", "idle|iowait|steal")
	return fmt.Sprintf(
		`avg by (node_name) (label_replace(rate(%s), "node_name", "$1", "instance_name", "(.*)"))`,
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

func (h *handler) querySystemHealthBadRequestsByService(ctx core.Context, startTime time.Time, endTime time.Time, groupFilter prometheus.PQLFilter, services []response.SystemHealthService) ([]prometheus.MetricResult, error) {
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

	results := make([]prometheus.MetricResult, 0, len(serviceNames))
	for _, serviceName := range serviceNames {
		serviceInstances := instancesByService[serviceName]
		if serviceInstances == nil {
			continue
		}
		levelCounts := serviceoverview.GroupAlertEventCountListByInstance(alertCounts, serviceInstances.GetInstances())
		results = append(results, prometheus.MetricResult{
			Metric: prometheus.Labels{SvcName: serviceName},
			Values: []prometheus.Points{{
				Value: float64(systemHealthAlertCount(levelCounts)),
			}},
		})
	}
	return results, nil
}

func systemHealthResourceRatioPQL(aggTemplate string, metric string, resourceFilter prometheus.PQLFilter, spanFilter prometheus.PQLFilter, window string) string {
	now := systemHealthResourceByServicePQL(aggTemplate, metric, resourceFilter, spanFilter, window, "")
	lastDay := systemHealthResourceByServicePQL(aggTemplate, metric, resourceFilter, spanFilter, window, "offset 24h")
	return fmt.Sprintf(`(%s) / (%s)`, now, lastDay)
}

func systemHealthResourceByServicePQL(aggTemplate string, metric string, resourceFilter prometheus.PQLFilter, spanFilter prometheus.PQLFilter, window string, offset string) string {
	resourceSeries := fmt.Sprintf(
		`sum by (cluster_id,namespace,pod) (`+aggTemplate+`)`,
		systemHealthRangeSelector(metric, prometheus.Clone(resourceFilter).NotEqual("container", "").NotEqual("container", "POD"), window, offset),
	)
	return systemHealthPodMetricJoinService(resourceSeries, spanFilter, window, offset)
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

func applySystemHealthExtraMetrics(system *response.SystemHealthSystem, extra systemHealthExtraMetricResults) {
	if system == nil || len(system.Services) == 0 {
		return
	}

	resourceByService := systemHealthResourceWaterlineDetails(extra.ResourceCPU, extra.ResourceMemory)
	latencyByService := systemHealthLatencyDetails(extra.Latency)
	errorRateByService := systemHealthErrorRateDetails(extra.ErrorRate)
	errorLogByService := systemHealthCountDetails(extra.ErrorLog, "当前时间窗口错误日志数")
	badRequestsByService := systemHealthCountDetails(extra.BadRequests, "当前活动告警数")
	resourceShiftByService := systemHealthResourceShiftDetails(extra.ResourceShift)
	lockCostByService := systemHealthLockCostDetails(extra.LockCost)
	for i := range system.Services {
		service := &system.Services[i]
		if detail, ok := resourceByService[service.Name]; ok {
			applySystemHealthResourceWaterline(service, detail)
		}
		if detail, ok := latencyByService[service.Name]; ok {
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
			service.Metrics[healthMetricResourceShift] = detail.Status
			service.MetricDetails[healthMetricResourceShift] = detail
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
}

func applySystemHealthResourceWaterline(service *response.SystemHealthService, detail response.SystemHealthMetricDetail) {
	status := detail.Status
	replicaStatus := service.Metrics[healthMetricResource]
	mergedStatus := worseHealthStatus(replicaStatus, status)
	if mergedStatus != status && replicaStatus != "" && replicaStatus != healthStatusGray {
		detail.Note = strings.Join([]string{detail.Note, systemMetricNote(healthMetricResource, replicaStatus)}, "；")
	}
	detail.Status = mergedStatus
	service.Metrics[healthMetricResource] = mergedStatus
	service.MetricDetails[healthMetricResource] = detail
}

func applySystemHealthCurrentValue(service *response.SystemHealthService, key string, detail response.SystemHealthMetricDetail) {
	status := service.Metrics[key]
	if status == "" {
		status = healthStatusGray
	}
	detail.Status = status
	service.MetricDetails[key] = detail
}

func applySystemHealthSystemMetricDetails(system *response.SystemHealthSystem, extra systemHealthExtraMetricResults) {
	if system == nil {
		return
	}

	if detail, ok := systemHealthSystemResourceWaterlineDetail(extra.SystemResourceCPU, extra.SystemResourceMemory); ok {
		existingStatus := system.Metrics[healthMetricResource]
		mergedStatus := worseHealthStatus(existingStatus, detail.Status)
		if mergedStatus != detail.Status && existingStatus != "" && existingStatus != healthStatusGray {
			detail.Note = strings.Join([]string{detail.Note, systemMetricNote(healthMetricResource, existingStatus)}, "；")
		}
		detail.Status = mergedStatus
		system.Metrics[healthMetricResource] = mergedStatus
		system.MetricDetails[healthMetricResource] = detail
	}

	if detail, ok := systemHealthSingleValueDetail(extra.SystemLatency, formatMillisecondsValue, "ms", "系统请求平均响应延时"); ok {
		applySystemHealthSystemValue(system, healthMetricLatency, detail)
	}
	applySystemHealthSystemTrend(system, healthMetricLatency, "ms", systemHealthSingleSeries(extra.SystemLatencyTrend, nanosecondsToMilliseconds), nil)

	if detail, ok := systemHealthSingleValueDetail(extra.SystemErrorRate, func(value float64) string {
		return formatPercentValue(value * 100)
	}, "%", "系统请求错误率"); ok {
		applySystemHealthSystemValue(system, healthMetricErrorRate, detail)
	}
	applySystemHealthSystemTrend(system, healthMetricErrorRate, "%", systemHealthSingleSeries(extra.SystemErrorRateTrend, ratioToPercent), nil)

	if count, ok := systemHealthCurrentSum(extra.ErrorLog); ok {
		applySystemHealthSystemValue(system, healthMetricErrorLog, response.SystemHealthMetricDetail{
			Value: formatCountValue(count),
			Unit:  "count",
			Note:  "系统窗口错误日志总数",
		})
	}
	applySystemHealthSystemTrend(system, healthMetricErrorLog, "count", systemHealthSumSeries(extra.SystemErrorLogTrend, identityMetricValue), nil)

	applySystemHealthSystemTrend(
		system,
		healthMetricResourceShift,
		"%",
		systemHealthMaxResourceShiftSeries(extra.SystemResourceShiftTrend),
		systemHealthResourceShiftThresholds(),
	)
}

func applySystemHealthSystemValue(system *response.SystemHealthSystem, key string, detail response.SystemHealthMetricDetail) {
	status := system.Metrics[key]
	if status == "" {
		status = healthStatusGray
	}
	detail.Status = status
	if current, ok := system.MetricDetails[key]; ok {
		detail.Series = current.Series
		detail.Thresholds = current.Thresholds
		if detail.Unit == "" {
			detail.Unit = current.Unit
		}
		if detail.Source == "" {
			detail.Source = current.Source
		}
	}
	system.MetricDetails[key] = detail
}

func applySystemHealthSystemTrend(system *response.SystemHealthSystem, key string, unit string, series []response.SystemHealthMetricPoint, thresholds []response.SystemHealthMetricThreshold) {
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
		detail.Unit = unit
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
			Value: formatMillisecondsValue(value),
			Note:  "当前平均响应延时",
		}
	}
	return details
}

func systemHealthErrorRateDetails(results []prometheus.MetricResult) map[string]response.SystemHealthMetricDetail {
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
			Value: formatPercentValue(value * 100),
			Note:  "当前请求错误率",
		}
	}
	return details
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
			Value:  formatPercentValue(candidate.value * 100),
			Note:   candidate.note,
		}
	}
	return details
}

func systemHealthSystemResourceWaterlineDetail(cpuResults []prometheus.MetricResult, memoryResults []prometheus.MetricResult) (response.SystemHealthMetricDetail, bool) {
	type resourceWaterlineCandidate struct {
		value  float64
		note   string
		source string
	}

	var selected *resourceWaterlineCandidate
	collect := func(results []prometheus.MetricResult, note string) {
		for _, result := range results {
			if len(result.Values) == 0 {
				continue
			}
			value := result.Values[0].Value
			if !isValidSystemHealthMetricValue(value) {
				continue
			}
			if selected == nil || value > selected.value {
				selected = &resourceWaterlineCandidate{
					value:  value,
					note:   note,
					source: systemHealthResourceSource(result.Metric),
				}
			}
		}
	}
	collect(cpuResults, "当前 CPU 资源水位")
	collect(memoryResults, "当前内存资源水位")

	if selected == nil {
		return response.SystemHealthMetricDetail{}, false
	}

	status := healthStatusGreen
	if math.IsInf(selected.value, 0) || selected.value >= 0.9 {
		status = healthStatusRed
	} else if selected.value >= 0.7 {
		status = healthStatusYellow
	}

	return response.SystemHealthMetricDetail{
		Status: status,
		Value:  formatPercentValue(selected.value * 100),
		Unit:   "%",
		Source: selected.source,
		Note:   selected.note,
	}, true
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

func systemHealthResourceShiftDetails(results []prometheus.MetricResult) map[string]response.SystemHealthMetricDetail {
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
		details[result.Metric.SvcName] = response.SystemHealthMetricDetail{
			Status: status,
			Value:  formatPercentValue(shiftPercent),
			Note:   "CPU/内存资源相对 24h 前基线波动",
		}
	}
	return details
}

func systemHealthLockCostDetails(results []prometheus.MetricResult) map[string]response.SystemHealthMetricDetail {
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
		details[result.Metric.SvcName] = response.SystemHealthMetricDetail{
			Status: status,
			Value:  formatMillisecondsPerRequest(value),
			Note:   "由 futex 锁等待耗时按请求数均摊计算",
		}
	}
	return details
}

func recalculateSystemHealthSystem(system *response.SystemHealthSystem) {
	if system == nil || len(system.Services) == 0 {
		return
	}

	scoreTotal := 0
	aggregateMetrics := make(map[string]string, len(systemHealthMetricKeys))
	for _, service := range system.Services {
		scoreTotal += service.Score
		for key, status := range service.Metrics {
			if status == healthStatusGray {
				continue
			}
			aggregateMetrics[key] = worseHealthStatus(aggregateMetrics[key], status)
		}
	}
	system.Score = int(math.Round(float64(scoreTotal) / float64(len(system.Services))))
	system.Status = scoreHealthStatus(system.Score)
	for _, key := range systemHealthMetricKeys {
		if status, ok := aggregateMetrics[key]; ok {
			system.Metrics[key] = status
			system.MetricDetails[key] = systemMetricDetail(system.Services, key, status)
		}
	}
}

func scoreFromMetricStatuses(metrics map[string]string) int {
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
		return 0
	}
	return int(math.Round(float64(total) / float64(count)))
}

func formatPercentValue(value float64) string {
	if math.IsInf(value, 0) {
		return "Inf%"
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
