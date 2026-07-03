// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package systemhealth

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/datagroup"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

func TestFirstLevelSystemGroupsOnlyUsesRootChildren(t *testing.T) {
	root := &datagroup.DataGroupTreeNode{
		DataGroup: datagroup.DataGroup{GroupID: 0, GroupName: "ALL"},
		SubGroups: []*datagroup.DataGroupTreeNode{
			{
				DataGroup: datagroup.DataGroup{GroupID: 101, GroupName: "订单系统"},
				SubGroups: []*datagroup.DataGroupTreeNode{
					{DataGroup: datagroup.DataGroup{GroupID: 201, GroupName: "订单子组"}},
				},
			},
			{DataGroup: datagroup.DataGroup{GroupID: 102, GroupName: "支付系统"}},
		},
	}

	groups := firstLevelSystemGroups(root)

	if len(groups) != 2 {
		t.Fatalf("expected only two first-level groups, got %d", len(groups))
	}
	if groups[0].GroupID != 101 || groups[1].GroupID != 102 {
		t.Fatalf("unexpected system groups: %+v", groups)
	}
	for _, group := range groups {
		if group.GroupID == 201 {
			t.Fatalf("nested group should not be treated as a system: %+v", group)
		}
	}
}

func serviceByName(t *testing.T, services []response.SystemHealthService, name string) response.SystemHealthService {
	t.Helper()
	for _, service := range services {
		if service.Name == name {
			return service
		}
	}
	t.Fatalf("expected service %q in %+v", name, services)
	return response.SystemHealthService{}
}

func TestBuildSystemHealthSystemAggregatesServiceStatus(t *testing.T) {
	group := datagroup.DataGroup{GroupID: 101, GroupName: "订单系统"}
	services := []*response.ServiceRYGResult{
		{
			ServiceName: "order-api",
			RYGResult: response.RYGResult{
				PercentScore: 40,
				Status:       response.YELLOW,
				ScoreDetail: []response.RYGScoreDetail{
					{Key: "latency", Score: 0, Detail: "latency too high"},
					{Key: "errorRate", Score: 2, Detail: "error rate warning"},
					{Key: "logErrorCount", Score: 3, Detail: "log ok"},
					{Key: "alert", Score: 0, Detail: "critical alert"},
					{Key: "replica", Score: 3, Detail: "replica ok"},
				},
			},
		},
		{
			ServiceName: "order-callback",
			RYGResult: response.RYGResult{
				PercentScore: 80,
				Status:       response.GREEN,
				ScoreDetail: []response.RYGScoreDetail{
					{Key: "latency", Score: 3, Detail: "latency ok"},
					{Key: "errorRate", Score: 3, Detail: "error rate ok"},
					{Key: "logErrorCount", Score: 3, Detail: "log ok"},
					{Key: "alert", Score: 3, Detail: "alert ok"},
					{Key: "replica", Score: 3, Detail: "replica ok"},
				},
			},
		},
	}

	system := buildSystemHealthSystem(group, services)

	if system.ID != 101 || system.GroupID != 101 || system.Name != "订单系统" {
		t.Fatalf("unexpected system identity: %+v", system)
	}
	if system.Score != 60 {
		t.Fatalf("expected average system score 60, got %d", system.Score)
	}
	if system.Metrics["latency"] != "red" {
		t.Fatalf("expected latency red, got %s", system.Metrics["latency"])
	}
	if system.Metrics["errorRate"] != "yellow" {
		t.Fatalf("expected errorRate yellow, got %s", system.Metrics["errorRate"])
	}
	if system.Metrics["errorLog"] != "green" {
		t.Fatalf("expected errorLog green, got %s", system.Metrics["errorLog"])
	}
	if system.Metrics["badRequests"] != "red" {
		t.Fatalf("expected badRequests red from alert score, got %s", system.Metrics["badRequests"])
	}
	if system.Metrics["resource"] != "green" {
		t.Fatalf("expected resource green from replica score, got %s", system.Metrics["resource"])
	}
	if system.Metrics["resourceShift"] != "gray" || system.Metrics["lockCost"] != "gray" {
		t.Fatalf("expected unsupported metrics to be gray, got %+v", system.Metrics)
	}
	if len(system.Services) != 2 {
		t.Fatalf("expected two services, got %d", len(system.Services))
	}
	if system.Services[0].Metrics["latency"] != "red" {
		t.Fatalf("expected service latency red, got %+v", system.Services[0].Metrics)
	}
}

func TestBuildSystemHealthSystemSortsServicesByLowScoreFirst(t *testing.T) {
	group := datagroup.DataGroup{GroupID: 101, GroupName: "订单系统"}
	services := []*response.ServiceRYGResult{
		{
			ServiceName: "healthy-api",
			RYGResult: response.RYGResult{
				PercentScore: 95,
			},
		},
		{
			ServiceName: "critical-api",
			RYGResult: response.RYGResult{
				PercentScore: 35,
			},
		},
		{
			ServiceName: "warning-api",
			RYGResult: response.RYGResult{
				PercentScore: 70,
			},
		},
	}

	system := buildSystemHealthSystem(group, services)

	if len(system.Services) != 3 {
		t.Fatalf("expected three services, got %d", len(system.Services))
	}
	got := []string{system.Services[0].Name, system.Services[1].Name, system.Services[2].Name}
	want := []string{"critical-api", "warning-api", "healthy-api"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected services sorted by low health score first, got %+v", got)
	}
}

func TestSortSystemHealthServicesByScoreHandlesZeroAndStableTies(t *testing.T) {
	services := []response.SystemHealthService{
		{Name: "healthy-api", Score: 100},
		{Name: "warning-a", Score: 66},
		{Name: "no-data", Score: 0},
		{Name: "warning-b", Score: 66},
		{Name: "critical", Score: 30},
	}

	sortSystemHealthServicesByScore(services)

	got := []string{services[0].Name, services[1].Name, services[2].Name, services[3].Name, services[4].Name}
	want := []string{"no-data", "critical", "warning-a", "warning-b", "healthy-api"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected zero, critical, stable equal scores, then healthy services, got %+v", got)
	}
}

func TestBuildSystemHealthSystemReturnsGrayMetricsWithoutServices(t *testing.T) {
	system := buildSystemHealthSystem(datagroup.DataGroup{GroupID: 101, GroupName: "空系统"}, nil)

	if system.Score != 0 {
		t.Fatalf("expected empty system score 0, got %d", system.Score)
	}
	for key, status := range system.Metrics {
		if status != "gray" {
			t.Fatalf("expected metric %s to be gray, got %s", key, status)
		}
	}
}

func TestBuildSystemHealthPendingIssuesAggregatesAllAccessibleSystems(t *testing.T) {
	systems := []response.SystemHealthSystem{
		{
			GroupID: 202,
			Name:    "支付系统",
			Score:   72,
			Services: []response.SystemHealthService{
				{
					Name:  "payment-api",
					Score: 65,
					Metrics: map[string]string{
						healthMetricLatency:   healthStatusYellow,
						healthMetricErrorRate: healthStatusGreen,
					},
					MetricDetails: map[string]response.SystemHealthMetricDetail{
						healthMetricLatency: {Status: healthStatusYellow, Value: "420ms", Note: "响应延时升高"},
					},
				},
			},
		},
		{
			GroupID: 101,
			Name:    "订单系统",
			Score:   38,
			Services: []response.SystemHealthService{
				{
					Name:  "order-api",
					Score: 35,
					Metrics: map[string]string{
						healthMetricErrorRate:   healthStatusRed,
						healthMetricLatency:     healthStatusYellow,
						healthMetricBadRequests: healthStatusYellow,
					},
					MetricDetails: map[string]response.SystemHealthMetricDetail{
						healthMetricErrorRate:   {Status: healthStatusRed, Value: "5.70%", Note: "错误率超过异常阈值"},
						healthMetricLatency:     {Status: healthStatusYellow, Value: "380ms", Note: "响应延时升高"},
						healthMetricBadRequests: {Status: healthStatusYellow, Value: "3", Note: "当前活跃告警数"},
					},
				},
			},
		},
	}

	issues := buildSystemHealthPendingIssues(systems, 20)

	if len(issues) != 4 {
		t.Fatalf("expected four pending issues, got %d: %+v", len(issues), issues)
	}
	first := issues[0]
	if first.GroupID != 101 || first.SystemName != "订单系统" || first.ServiceName != "order-api" || first.MetricKey != healthMetricErrorRate {
		t.Fatalf("expected red order-api errorRate issue first, got %+v", first)
	}
	if first.ID != "101:order-api:errorRate" {
		t.Fatalf("expected stable issue id, got %q", first.ID)
	}
	if first.Value != "5.70%" || first.Note != "错误率超过异常阈值" {
		t.Fatalf("expected metric detail copied, got %+v", first)
	}
	if issues[1].MetricKey != healthMetricLatency || issues[1].GroupID != 101 {
		t.Fatalf("expected lower-score yellow issue second, got %+v", issues[1])
	}
	badRequestsIssue := response.SystemHealthPendingIssue{}
	for _, issue := range issues {
		if issue.MetricKey == healthMetricBadRequests {
			badRequestsIssue = issue
			break
		}
	}
	if badRequestsIssue.MetricLabel != "活跃告警" {
		t.Fatalf("expected badRequests issue label to be active alerts, got %+v", badRequestsIssue)
	}
}

func TestBuildSystemHealthPendingIssuesHonorsLimit(t *testing.T) {
	metrics := map[string]string{
		healthMetricErrorRate:   healthStatusRed,
		healthMetricLatency:     healthStatusYellow,
		healthMetricBadRequests: healthStatusYellow,
	}
	system := response.SystemHealthSystem{
		GroupID: 101,
		Name:    "订单系统",
		Score:   50,
		Services: []response.SystemHealthService{
			{
				Name:          "order-api",
				Score:         40,
				Metrics:       metrics,
				MetricDetails: defaultSystemMetricDetails(),
			},
		},
	}

	issues := buildSystemHealthPendingIssues([]response.SystemHealthSystem{system}, 2)

	if len(issues) != 2 {
		t.Fatalf("expected limited pending issues, got %d", len(issues))
	}
}

func TestBuildSystemHealthOverviewSummaryUsesScoreThresholdOnly(t *testing.T) {
	scoreThresholds := response.SystemHealthConfigResp{
		SevereAbnormalScoreThreshold:  60,
		GeneralAbnormalScoreThreshold: 80,
	}
	systems := []response.SystemHealthSystem{
		{
			GroupID: 101,
			Name:    "订单系统",
			Score:   79,
			Services: []response.SystemHealthService{
				{Name: "order-api", Score: 79},
			},
			Metrics: map[string]string{
				healthMetricErrorRate: healthStatusGreen,
			},
		},
		{
			GroupID: 102,
			Name:    "支付系统",
			Score:   80,
			Services: []response.SystemHealthService{
				{Name: "payment-api", Score: 80},
			},
			Metrics: map[string]string{
				healthMetricLatency: healthStatusRed,
			},
		},
		{
			GroupID: 103,
			Name:    "库存系统",
			Score:   95,
			Services: []response.SystemHealthService{
				{Name: "inventory-api", Score: 95},
			},
			Metrics: map[string]string{
				healthMetricErrorLog: healthStatusYellow,
			},
		},
		{
			GroupID: 104,
			Name:    "空系统",
			Score:   0,
			Metrics: map[string]string{
				healthMetricLatency: healthStatusRed,
			},
		},
	}

	summary := buildSystemHealthOverviewSummary(systems, scoreThresholds)

	if summary.AbnormalSystems != 1 {
		t.Fatalf("expected only the below-threshold system to be abnormal, got %d", summary.AbnormalSystems)
	}
	if summary.AbnormalServices != 1 {
		t.Fatalf("expected only the below-threshold service to be abnormal, got %d", summary.AbnormalServices)
	}
	if summary.AbnormalMetrics != 3 {
		t.Fatalf("expected red/yellow metric counting to stay unchanged, got %d", summary.AbnormalMetrics)
	}
	if summary.SevereMetrics != 2 {
		t.Fatalf("expected red metric counting to stay unchanged, got %d", summary.SevereMetrics)
	}
	if summary.NoDataSystems != 1 {
		t.Fatalf("expected no-data system count to stay unchanged, got %d", summary.NoDataSystems)
	}
}

func TestScoreHealthStatusUsesConfiguredScoreThresholds(t *testing.T) {
	scoreThresholds := response.SystemHealthConfigResp{
		SevereAbnormalScoreThreshold:  45,
		GeneralAbnormalScoreThreshold: 82,
	}

	if got := scoreHealthStatus(44, scoreThresholds); got != healthStatusRed {
		t.Fatalf("expected score below severe threshold to be red, got %s", got)
	}
	if got := scoreHealthStatus(45, scoreThresholds); got != healthStatusYellow {
		t.Fatalf("expected score equal to severe threshold to be yellow, got %s", got)
	}
	if got := scoreHealthStatus(81, scoreThresholds); got != healthStatusYellow {
		t.Fatalf("expected score below general threshold to be yellow, got %s", got)
	}
	if got := scoreHealthStatus(82, scoreThresholds); got != healthStatusGreen {
		t.Fatalf("expected score equal to general threshold to be green, got %s", got)
	}
}

func TestBuildSystemHealthOverviewSummaryCountsNewAbnormalMetrics(t *testing.T) {
	scoreThresholds := response.SystemHealthConfigResp{
		SevereAbnormalScoreThreshold:  60,
		GeneralAbnormalScoreThreshold: 80,
	}
	current := []response.SystemHealthSystem{
		{
			GroupID: 101,
			Name:    "订单系统",
			Score:   72,
			Services: []response.SystemHealthService{
				{Name: "order-api"},
			},
			Metrics: map[string]string{
				healthMetricErrorRate: healthStatusRed,
				healthMetricLatency:   healthStatusYellow,
				healthMetricErrorLog:  healthStatusGreen,
			},
		},
		{
			GroupID: 102,
			Name:    "支付系统",
			Score:   92,
			Services: []response.SystemHealthService{
				{Name: "payment-api"},
			},
			Metrics: map[string]string{
				healthMetricErrorRate: healthStatusGreen,
				healthMetricLatency:   healthStatusYellow,
			},
		},
	}
	previous := []response.SystemHealthSystem{
		{
			GroupID: 101,
			Metrics: map[string]string{
				healthMetricErrorRate: healthStatusGreen,
				healthMetricLatency:   healthStatusYellow,
				healthMetricErrorLog:  healthStatusGreen,
			},
		},
		{
			GroupID: 102,
			Metrics: map[string]string{
				healthMetricErrorRate: healthStatusGreen,
				healthMetricLatency:   healthStatusGray,
			},
		},
	}

	summary := buildSystemHealthOverviewSummary(current, scoreThresholds, previous)

	if summary.AbnormalSystems != 1 {
		t.Fatalf("expected one abnormal system by score only, got %d", summary.AbnormalSystems)
	}
	if summary.AbnormalMetrics != 3 {
		t.Fatalf("expected three abnormal metrics, got %d", summary.AbnormalMetrics)
	}
	if summary.SevereMetrics != 1 {
		t.Fatalf("expected one severe metric, got %d", summary.SevereMetrics)
	}
	if summary.NewAbnormalMetrics != 2 {
		t.Fatalf("expected two newly abnormal metrics, got %d", summary.NewAbnormalMetrics)
	}
}

func TestNormalizeSystemHealthScoreThresholdsFallsBackToDefault(t *testing.T) {
	valid := normalizeSystemHealthScoreThresholds(response.SystemHealthConfigResp{
		SevereAbnormalScoreThreshold:  45,
		GeneralAbnormalScoreThreshold: 82,
	})
	if valid.SevereAbnormalScoreThreshold != 45 || valid.GeneralAbnormalScoreThreshold != 82 {
		t.Fatalf("expected valid thresholds to be kept, got %+v", valid)
	}

	invalid := normalizeSystemHealthScoreThresholds(response.SystemHealthConfigResp{
		SevereAbnormalScoreThreshold:  90,
		GeneralAbnormalScoreThreshold: 80,
	})
	if invalid.SevereAbnormalScoreThreshold != 60 || invalid.GeneralAbnormalScoreThreshold != 80 {
		t.Fatalf("expected invalid thresholds to fall back to defaults, got %+v", invalid)
	}
}

func TestNewAbnormalMetricWindowsUseRecentTenMinutes(t *testing.T) {
	endTime := time.Date(2026, 6, 22, 16, 30, 0, 0, time.UTC)

	currentStart, previousStart, previousEnd := newAbnormalMetricWindows(endTime)

	if !currentStart.Equal(endTime.Add(-10 * time.Minute)) {
		t.Fatalf("expected current start to be recent 10 minutes, got %s", currentStart)
	}
	if !previousStart.Equal(endTime.Add(-20 * time.Minute)) {
		t.Fatalf("expected previous start to be previous 10 minutes, got %s", previousStart)
	}
	if !previousEnd.Equal(currentStart) {
		t.Fatalf("expected previous end to match current start, got %s", previousEnd)
	}
}

func TestBuildOverviewSystemsForGroupsRunsConcurrentlyAndSorts(t *testing.T) {
	groups := []*datagroup.DataGroupTreeNode{
		{DataGroup: datagroup.DataGroup{GroupID: 2, GroupName: "slow"}},
		{DataGroup: datagroup.DataGroup{GroupID: 1, GroupName: "fast"}},
	}
	started := make(chan int64, len(groups))
	release := make(chan struct{})
	resultCh := make(chan []response.SystemHealthSystem, 1)
	errCh := make(chan error, 1)

	go func() {
		systems, err := buildOverviewSystemsForGroups(groups, 2, func(group *datagroup.DataGroupTreeNode) (response.SystemHealthSystem, error) {
			started <- group.GroupID
			<-release
			score := 90
			if group.GroupID == 2 {
				score = 10
			}
			return response.SystemHealthSystem{GroupID: group.GroupID, Score: score}, nil
		})
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- systems
	}()

	for i := 0; i < len(groups); i++ {
		select {
		case <-started:
		case <-time.After(200 * time.Millisecond):
			t.Fatalf("expected group %d to start before previous group was released", i+1)
		}
	}
	close(release)

	select {
	case err := <-errCh:
		t.Fatalf("expected systems, got error: %v", err)
	case systems := <-resultCh:
		if len(systems) != 2 {
			t.Fatalf("expected two systems, got %d", len(systems))
		}
		if systems[0].GroupID != 2 || systems[1].GroupID != 1 {
			t.Fatalf("expected systems sorted by score after concurrent build, got %+v", systems)
		}
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for concurrent system build")
	}
}

func TestBuildOverviewSystemsForGroupsKeepsGraySystemWhenBuilderFails(t *testing.T) {
	groups := []*datagroup.DataGroupTreeNode{
		{DataGroup: datagroup.DataGroup{GroupID: 1, GroupName: "orders"}},
	}

	systems, err := buildOverviewSystemsForGroups(groups, 1, func(group *datagroup.DataGroupTreeNode) (response.SystemHealthSystem, error) {
		return response.SystemHealthSystem{}, errors.New("build failed")
	})

	if err != nil {
		t.Fatalf("expected degraded systems, got error: %v", err)
	}
	if len(systems) != 1 {
		t.Fatalf("expected one degraded system, got %d", len(systems))
	}
	if systems[0].GroupID != 1 || systems[0].Name != "orders" {
		t.Fatalf("expected degraded system identity from group, got %+v", systems[0])
	}
	if systems[0].Status != healthStatusGray || len(systems[0].Services) != 0 {
		t.Fatalf("expected gray system without services, got %+v", systems[0])
	}
}

func TestSystemHealthOverviewWorkerContextUsesSnapshots(t *testing.T) {
	baseCtx := context.WithValue(context.Background(), "trace", "value")
	ctx := systemHealthOverviewWorkerContext{
		requestContext: baseCtx,
		requestID:      "request-1",
		routeMethod:    "GET",
		routePath:      "/api/system-health/overview",
		lang:           code.LANG_ZH,
		userID:         42,
	}

	if ctx.GetContext() != baseCtx {
		t.Fatalf("expected base context snapshot")
	}
	if ctx.LANG() != code.LANG_ZH {
		t.Fatalf("expected language snapshot, got %q", ctx.LANG())
	}
	if ctx.UserID() != 42 {
		t.Fatalf("expected user id snapshot, got %d", ctx.UserID())
	}
	if method, path := ctx.GetMethodPath(); method != "GET" || path != "/api/system-health/overview" {
		t.Fatalf("expected route snapshot, got method=%q path=%q", method, path)
	}
	if requestID, ok := ctx.Get(core.RequestIDKey); !ok || requestID != "request-1" {
		t.Fatalf("expected request id snapshot, got value=%v ok=%v", requestID, ok)
	}
	if value := ctx.Value("trace"); value != "value" {
		t.Fatalf("expected base context value, got %v", value)
	}
}

func TestSystemHealthOverviewExtraMetricTimeoutCapsCurrentAndHistoricalWindows(t *testing.T) {
	if timeout := systemHealthOverviewExtraMetricTimeout("current"); timeout != 2*time.Second {
		t.Fatalf("expected current window extra metrics timeout 2s, got %s", timeout)
	}
	if timeout := systemHealthOverviewExtraMetricTimeout("recent"); timeout != time.Second {
		t.Fatalf("expected recent window extra metrics timeout 1s, got %s", timeout)
	}
	if timeout := systemHealthOverviewExtraMetricTimeout("previous"); timeout != time.Second {
		t.Fatalf("expected previous window extra metrics timeout 1s, got %s", timeout)
	}
}

func TestApplySystemHealthExtraMetricsAggregatesResourceShiftAndLockCost(t *testing.T) {
	system := response.SystemHealthSystem{
		Metrics:       defaultSystemMetricStatus(),
		MetricDetails: defaultSystemMetricDetails(),
		Services: []response.SystemHealthService{
			{
				Name:          "order-api",
				Metrics:       defaultSystemMetricStatus(),
				MetricDetails: defaultSystemMetricDetails(),
			},
			{
				Name:          "order-worker",
				Metrics:       defaultSystemMetricStatus(),
				MetricDetails: defaultSystemMetricDetails(),
			},
		},
	}
	extra := systemHealthExtraMetricResults{
		ResourceShift: []prometheus.MetricResult{
			{
				Metric: prometheus.Labels{SvcName: "order-api"},
				Values: []prometheus.Points{{Value: 1.75}},
			},
			{
				Metric: prometheus.Labels{SvcName: "order-worker"},
				Values: []prometheus.Points{{Value: 1.05}},
			},
		},
		LockCost: []prometheus.MetricResult{
			{
				Metric: prometheus.Labels{SvcName: "order-api"},
				Values: []prometheus.Points{{Value: 12_000_000}},
			},
			{
				Metric: prometheus.Labels{SvcName: "order-worker"},
				Values: []prometheus.Points{{Value: 500_000}},
			},
		},
	}

	applySystemHealthExtraMetrics(&system, extra)

	if system.Services[0].Metrics["resourceShift"] != "red" {
		t.Fatalf("expected order-api resourceShift red, got %+v", system.Services[0].Metrics)
	}
	if system.Services[0].MetricDetails["resourceShift"].Value != "75.00%" {
		t.Fatalf("expected order-api resourceShift value 75.00%%, got %+v", system.Services[0].MetricDetails["resourceShift"])
	}
	if system.Services[0].Metrics["lockCost"] != "yellow" {
		t.Fatalf("expected order-api lockCost yellow, got %+v", system.Services[0].Metrics)
	}
	if system.Services[0].MetricDetails["lockCost"].Value != "12.00 ms/请求" {
		t.Fatalf("expected order-api lockCost value 12.00 ms/请求, got %+v", system.Services[0].MetricDetails["lockCost"])
	}
	if system.Services[1].Metrics["resourceShift"] != "green" || system.Services[1].Metrics["lockCost"] != "green" {
		t.Fatalf("expected order-worker extra metrics green, got %+v", system.Services[1].Metrics)
	}
	if system.Metrics["resourceShift"] != "red" || system.Metrics["lockCost"] != "yellow" {
		t.Fatalf("expected system to aggregate worst extra metrics, got %+v", system.Metrics)
	}
	if system.MetricDetails["resourceShift"].Value != "75.00%" {
		t.Fatalf("expected system resourceShift value to use worst service metric value, got %+v", system.MetricDetails["resourceShift"])
	}
	if system.MetricDetails["lockCost"].Value != "12.00 ms/请求" {
		t.Fatalf("expected system lockCost value to use worst service metric value, got %+v", system.MetricDetails["lockCost"])
	}
	if system.MetricDetails["resourceShift"].Note != "资源波动超过异常阈值" {
		t.Fatalf("expected resourceShift note to describe resource baseline, got %q", system.MetricDetails["resourceShift"].Note)
	}
	if system.MetricDetails["lockCost"].Note != "锁等待耗时升高" {
		t.Fatalf("expected lockCost note to describe futex cost, got %q", system.MetricDetails["lockCost"].Note)
	}
}

func TestApplySystemHealthExtraMetricsSortsServicesByLowScoreFirst(t *testing.T) {
	healthyMetrics := defaultSystemMetricStatus()
	for _, key := range systemHealthMetricKeys {
		healthyMetrics[key] = healthStatusGreen
	}
	system := response.SystemHealthSystem{
		Metrics:       defaultSystemMetricStatus(),
		MetricDetails: defaultSystemMetricDetails(),
		Services: []response.SystemHealthService{
			{
				Name:          "healthy-api",
				Metrics:       healthyMetrics,
				MetricDetails: defaultSystemMetricDetails(),
			},
			{
				Name:          "critical-api",
				Metrics:       defaultSystemMetricStatus(),
				MetricDetails: defaultSystemMetricDetails(),
			},
		},
	}
	extra := systemHealthExtraMetricResults{
		ResourceShift: []prometheus.MetricResult{
			{
				Metric: prometheus.Labels{SvcName: "critical-api"},
				Values: []prometheus.Points{{Value: 1.75}},
			},
		},
	}

	applySystemHealthExtraMetrics(&system, extra)

	got := []string{system.Services[0].Name, system.Services[1].Name}
	want := []string{"critical-api", "healthy-api"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected services sorted by low recalculated health score first, got %+v", got)
	}
}

func TestApplySystemHealthExtraMetricsMergesVMHostResourceDetails(t *testing.T) {
	system := response.SystemHealthSystem{
		Metrics:       defaultSystemMetricStatus(),
		MetricDetails: defaultSystemMetricDetails(),
		Services: []response.SystemHealthService{
			{Name: "order-api", Metrics: defaultSystemMetricStatus(), MetricDetails: defaultSystemMetricDetails()},
			{Name: "order-worker", Metrics: defaultSystemMetricStatus(), MetricDetails: defaultSystemMetricDetails()},
		},
	}

	scope := newSystemHealthVMHostResourceScope()
	scope.addInstance("order-api", &model.ServiceInstance{ServiceName: "order-api", NodeName: "vm-a", Pid: 2201})
	scope.addInstance("order-worker", &model.ServiceInstance{ServiceName: "order-worker", NodeName: "vm-b", Pid: 3301})

	extra := systemHealthExtraMetricResults{
		VMHostResource: systemHealthVMHostResourceResults{
			Scope: scope,
			CPU: []prometheus.MetricResult{
				{Metric: prometheus.Labels{NodeName: "vm-a"}, Values: []prometheus.Points{{Value: 0.72}}},
				{Metric: prometheus.Labels{NodeName: "vm-b"}, Values: []prometheus.Points{{Value: 0.32}}},
			},
			Memory: []prometheus.MetricResult{
				{Metric: prometheus.Labels{NodeName: "vm-a"}, Values: []prometheus.Points{{Value: 0.64}}},
				{Metric: prometheus.Labels{NodeName: "vm-b"}, Values: []prometheus.Points{{Value: 0.93}}},
			},
			Shift: []prometheus.MetricResult{
				{Metric: prometheus.Labels{NodeName: "vm-a"}, Values: []prometheus.Points{{Value: 1.35}}},
				{Metric: prometheus.Labels{NodeName: "vm-b"}, Values: []prometheus.Points{{Value: 1.05}}},
			},
		},
	}

	applySystemHealthExtraMetrics(&system, extra)

	orderAPI := serviceByName(t, system.Services, "order-api")
	apiResource := orderAPI.MetricDetails[healthMetricResource]
	if orderAPI.Metrics[healthMetricResource] != healthStatusYellow || apiResource.Value != "72.00%" || apiResource.Source != "host/vm-a" {
		t.Fatalf("expected order-api host CPU waterline yellow, metrics=%+v detail=%+v", orderAPI.Metrics, apiResource)
	}
	apiShift := orderAPI.MetricDetails[healthMetricResourceShift]
	if orderAPI.Metrics[healthMetricResourceShift] != healthStatusYellow || apiShift.Value != "35.00%" || apiShift.Source != "host/vm-a" {
		t.Fatalf("expected order-api host resource shift yellow, metrics=%+v detail=%+v", orderAPI.Metrics, apiShift)
	}

	orderWorker := serviceByName(t, system.Services, "order-worker")
	workerResource := orderWorker.MetricDetails[healthMetricResource]
	if orderWorker.Metrics[healthMetricResource] != healthStatusRed || workerResource.Value != "93.00%" || workerResource.Source != "host/vm-b" {
		t.Fatalf("expected order-worker host memory waterline red, metrics=%+v detail=%+v", orderWorker.Metrics, workerResource)
	}

	systemResource := system.MetricDetails[healthMetricResource]
	if system.Metrics[healthMetricResource] != healthStatusRed || systemResource.Value != "93.00%" || systemResource.Source != "host/vm-b" {
		t.Fatalf("expected system resource to use highest VM host waterline, metrics=%+v detail=%+v", system.Metrics, systemResource)
	}
}

func TestApplySystemHealthExtraMetricsKeepsWorseContainerOrVMResource(t *testing.T) {
	system := response.SystemHealthSystem{
		Metrics:       defaultSystemMetricStatus(),
		MetricDetails: defaultSystemMetricDetails(),
		Services: []response.SystemHealthService{
			{Name: "order-api", Metrics: defaultSystemMetricStatus(), MetricDetails: defaultSystemMetricDetails()},
		},
	}

	scope := newSystemHealthVMHostResourceScope()
	scope.addInstance("order-api", &model.ServiceInstance{ServiceName: "order-api", NodeName: "vm-a", Pid: 2201})

	extra := systemHealthExtraMetricResults{
		ResourceCPU: []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "order-api"}, Values: []prometheus.Points{{Value: 0.91}}},
		},
		VMHostResource: systemHealthVMHostResourceResults{
			Scope: scope,
			Memory: []prometheus.MetricResult{
				{Metric: prometheus.Labels{NodeName: "vm-a"}, Values: []prometheus.Points{{Value: 0.72}}},
			},
		},
	}

	applySystemHealthExtraMetrics(&system, extra)

	detail := system.Services[0].MetricDetails[healthMetricResource]
	if system.Services[0].Metrics[healthMetricResource] != healthStatusRed || detail.Value != "91.00%" || detail.Source != "" {
		t.Fatalf("expected worse container resource detail to win, metrics=%+v detail=%+v", system.Services[0].Metrics, detail)
	}
}

func TestApplySystemHealthExtraMetricsClampsServiceResourceWaterlineAtHundredPercent(t *testing.T) {
	system := response.SystemHealthSystem{
		Metrics:       defaultSystemMetricStatus(),
		MetricDetails: defaultSystemMetricDetails(),
		Services: []response.SystemHealthService{
			{Name: "order-api", Metrics: defaultSystemMetricStatus(), MetricDetails: defaultSystemMetricDetails()},
		},
	}

	extra := systemHealthExtraMetricResults{
		ResourceCPU: []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "order-api"}, Values: []prometheus.Points{{Value: 1.18}}},
		},
	}

	applySystemHealthExtraMetrics(&system, extra)

	detail := system.Services[0].MetricDetails[healthMetricResource]
	if system.Services[0].Metrics[healthMetricResource] != healthStatusRed || detail.Value != "100.00%" {
		t.Fatalf("expected service resource waterline to clamp at 100%%, metrics=%+v detail=%+v", system.Services[0].Metrics, detail)
	}
}

func TestApplySystemHealthExtraMetricsAddsVMHostShiftTrendToSystemTrend(t *testing.T) {
	system := response.SystemHealthSystem{
		Metrics:       defaultSystemMetricStatus(),
		MetricDetails: defaultSystemMetricDetails(),
		Services: []response.SystemHealthService{
			{Name: "order-api", Metrics: map[string]string{healthMetricResourceShift: healthStatusYellow}, MetricDetails: defaultSystemMetricDetails()},
		},
	}

	scope := newSystemHealthVMHostResourceScope()
	scope.addInstance("order-api", &model.ServiceInstance{ServiceName: "order-api", NodeName: "vm-a", Pid: 2201})

	extra := systemHealthExtraMetricResults{
		VMHostResource: systemHealthVMHostResourceResults{
			Scope: scope,
			ShiftTrend: []prometheus.MetricResult{
				{Metric: prometheus.Labels{NodeName: "vm-a"}, Values: []prometheus.Points{{TimeStamp: 1000, Value: 1.20}, {TimeStamp: 2000, Value: 1.75}}},
			},
		},
		SystemResourceShiftTrend: []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "container-api"}, Values: []prometheus.Points{{TimeStamp: 1000, Value: 1.45}, {TimeStamp: 2000, Value: 1.10}}},
		},
	}
	extra.SystemResourceShiftTrend = append(extra.SystemResourceShiftTrend, extra.VMHostResource.ShiftTrend...)

	applySystemHealthExtraMetrics(&system, extra)

	detail := system.MetricDetails[healthMetricResourceShift]
	if detail.Unit != "%" {
		t.Fatalf("expected resource shift trend unit percent, got %+v", detail)
	}
	if len(detail.Thresholds) != 2 || detail.Thresholds[0].Value != 30 || detail.Thresholds[1].Value != 60 {
		t.Fatalf("expected resource shift thresholds, got %+v", detail.Thresholds)
	}
	if len(detail.Series) != 2 {
		t.Fatalf("expected two trend points, got %+v", detail.Series)
	}
	if detail.Series[0].Timestamp != 1000 || detail.Series[0].Value != 45 {
		t.Fatalf("expected timestamp 1000 to use container 45%% shift, got %+v", detail.Series[0])
	}
	if detail.Series[1].Timestamp != 2000 || detail.Series[1].Value != 75 {
		t.Fatalf("expected timestamp 2000 to use VM host 75%% shift, got %+v", detail.Series[1])
	}
}

func TestApplySystemHealthExtraMetricsKeepsWorseServiceResourceDetailWhenVMHostIsLessSevere(t *testing.T) {
	serviceMetrics := defaultSystemMetricStatus()
	serviceMetrics[healthMetricResource] = healthStatusRed
	serviceDetails := defaultSystemMetricDetails()
	serviceDetails[healthMetricResource] = response.SystemHealthMetricDetail{
		Status: healthStatusRed,
		Value:  "--",
		Note:   "existing replica detail",
		Source: "replica",
	}
	system := response.SystemHealthSystem{
		Metrics:       defaultSystemMetricStatus(),
		MetricDetails: defaultSystemMetricDetails(),
		Services: []response.SystemHealthService{
			{Name: "order-api", Metrics: serviceMetrics, MetricDetails: serviceDetails},
		},
	}

	scope := newSystemHealthVMHostResourceScope()
	scope.addInstance("order-api", &model.ServiceInstance{ServiceName: "order-api", NodeName: "vm-a", Pid: 2201})
	extra := systemHealthExtraMetricResults{
		VMHostResource: systemHealthVMHostResourceResults{
			Scope: scope,
			CPU: []prometheus.MetricResult{
				{Metric: prometheus.Labels{NodeName: "vm-a"}, Values: []prometheus.Points{{Value: 0.72}}},
			},
		},
	}

	applySystemHealthExtraMetrics(&system, extra)

	resource := system.Services[0].MetricDetails[healthMetricResource]
	if system.Services[0].Metrics[healthMetricResource] != healthStatusRed || resource.Value != "--" || resource.Source != "replica" {
		t.Fatalf("expected existing red replica detail to remain, metrics=%+v detail=%+v", system.Services[0].Metrics, resource)
	}
}

func TestApplySystemHealthExtraMetricsFillsSystemResourceValueWhenVMHostIsLessSevere(t *testing.T) {
	apiMetrics := defaultSystemMetricStatus()
	apiMetrics[healthMetricResource] = healthStatusGreen
	apiDetails := defaultSystemMetricDetails()
	apiDetails[healthMetricResource] = response.SystemHealthMetricDetail{
		Status: healthStatusGreen,
		Value:  "42.00%",
		Source: "container",
	}
	workerMetrics := defaultSystemMetricStatus()
	workerMetrics[healthMetricResource] = healthStatusRed
	workerDetails := defaultSystemMetricDetails()
	workerDetails[healthMetricResource] = response.SystemHealthMetricDetail{
		Status: healthStatusRed,
		Value:  "--",
		Note:   "existing system detail",
		Source: "replica",
	}
	system := response.SystemHealthSystem{
		Metrics:       defaultSystemMetricStatus(),
		MetricDetails: defaultSystemMetricDetails(),
		Services: []response.SystemHealthService{
			{Name: "order-api", Metrics: apiMetrics, MetricDetails: apiDetails},
			{Name: "order-worker", Metrics: workerMetrics, MetricDetails: workerDetails},
		},
	}

	scope := newSystemHealthVMHostResourceScope()
	scope.addInstance("order-api", &model.ServiceInstance{ServiceName: "order-api", NodeName: "vm-a", Pid: 2201})
	extra := systemHealthExtraMetricResults{
		VMHostResource: systemHealthVMHostResourceResults{
			Scope: scope,
			CPU: []prometheus.MetricResult{
				{Metric: prometheus.Labels{NodeName: "vm-a"}, Values: []prometheus.Points{{Value: 0.72}}},
			},
		},
	}

	applySystemHealthExtraMetrics(&system, extra)

	resource := system.MetricDetails[healthMetricResource]
	if system.Metrics[healthMetricResource] != healthStatusRed || resource.Value != "72.00%" || resource.Source != "host/vm-a" || resource.Note != "资源副本健康存在风险" {
		t.Fatalf("expected existing red system detail to keep status and fill value, metrics=%+v detail=%+v", system.Metrics, resource)
	}
}

func TestApplySystemHealthExtraMetricsPopulatesSystemResourceValueWhenVMHostIsLessSevere(t *testing.T) {
	apiMetrics := defaultSystemMetricStatus()
	apiMetrics[healthMetricResource] = healthStatusRed
	apiDetails := defaultSystemMetricDetails()
	apiDetails[healthMetricResource] = response.SystemHealthMetricDetail{
		Status: healthStatusRed,
		Value:  "--",
		Note:   "existing system detail",
		Source: "replica",
	}
	system := response.SystemHealthSystem{
		Metrics:       defaultSystemMetricStatus(),
		MetricDetails: defaultSystemMetricDetails(),
		Services: []response.SystemHealthService{
			{Name: "order-api", Metrics: apiMetrics, MetricDetails: apiDetails},
		},
	}

	scope := newSystemHealthVMHostResourceScope()
	scope.addInstance("order-api", &model.ServiceInstance{ServiceName: "order-api", NodeName: "vm-a", Pid: 2201})
	extra := systemHealthExtraMetricResults{
		VMHostResource: systemHealthVMHostResourceResults{
			Scope: scope,
			Memory: []prometheus.MetricResult{
				{Metric: prometheus.Labels{NodeName: "vm-a"}, Values: []prometheus.Points{{Value: 0.695}}},
			},
		},
	}

	applySystemHealthExtraMetrics(&system, extra)

	resource := system.MetricDetails[healthMetricResource]
	if system.Metrics[healthMetricResource] != healthStatusRed || resource.Value != "69.50%" || resource.Source != "host/vm-a" || resource.Note != "资源副本健康存在风险" {
		t.Fatalf("expected existing red system detail to keep status and fill value, metrics=%+v detail=%+v", system.Metrics, resource)
	}
}

func TestApplySystemHealthResourceWaterlineKeepsLargestSameStatusNumericDetail(t *testing.T) {
	lower := response.SystemHealthService{
		Metrics: map[string]string{healthMetricResource: healthStatusYellow},
		MetricDetails: map[string]response.SystemHealthMetricDetail{
			healthMetricResource: {Status: healthStatusYellow, Value: "72.00%", Source: "container-a"},
		},
	}
	applySystemHealthResourceWaterline(&lower, response.SystemHealthMetricDetail{
		Status: healthStatusYellow,
		Value:  "75.00%",
		Source: "host/vm-a",
	})
	if detail := lower.MetricDetails[healthMetricResource]; lower.Metrics[healthMetricResource] != healthStatusYellow || detail.Value != "75.00%" || detail.Source != "host/vm-a" {
		t.Fatalf("expected larger same-status detail to replace current, metrics=%+v detail=%+v", lower.Metrics, detail)
	}

	higher := response.SystemHealthService{
		Metrics: map[string]string{healthMetricResource: healthStatusYellow},
		MetricDetails: map[string]response.SystemHealthMetricDetail{
			healthMetricResource: {Status: healthStatusYellow, Value: "75.00%", Source: "container-b"},
		},
	}
	applySystemHealthResourceWaterline(&higher, response.SystemHealthMetricDetail{
		Status: healthStatusYellow,
		Value:  "72.00%",
		Source: "host/vm-b",
	})
	if detail := higher.MetricDetails[healthMetricResource]; higher.Metrics[healthMetricResource] != healthStatusYellow || detail.Value != "75.00%" || detail.Source != "container-b" {
		t.Fatalf("expected larger existing same-status detail to remain, metrics=%+v detail=%+v", higher.Metrics, detail)
	}
}

func TestSystemMetricDetailDoesNotFallbackToServiceCountWithoutValues(t *testing.T) {
	services := []response.SystemHealthService{
		{
			Name:          "order-api",
			Metrics:       map[string]string{healthMetricLatency: healthStatusRed},
			MetricDetails: defaultSystemMetricDetails(),
		},
		{
			Name:          "order-worker",
			Metrics:       map[string]string{healthMetricLatency: healthStatusRed},
			MetricDetails: defaultSystemMetricDetails(),
		},
	}

	detail := systemMetricDetail(services, healthMetricLatency, healthStatusRed)

	if detail.Value != "--" {
		t.Fatalf("expected missing system metric value to stay empty instead of service count, got %q", detail.Value)
	}
	if strings.Contains(detail.Value, "服务") {
		t.Fatalf("system metric value should not fall back to service count, got %q", detail.Value)
	}
	if detail.Note != "响应延时超过异常阈值" {
		t.Fatalf("expected user-facing latency note, got %q", detail.Note)
	}
}

func TestApplySystemHealthExtraMetricsRecalculatesServicesWithoutExtraSamples(t *testing.T) {
	healthyMetrics := defaultSystemMetricStatus()
	for _, key := range []string{"resource", "errorRate", "latency", "errorLog", "badRequests"} {
		healthyMetrics[key] = healthStatusGreen
	}
	system := response.SystemHealthSystem{
		Metrics:       defaultSystemMetricStatus(),
		MetricDetails: defaultSystemMetricDetails(),
		Services: []response.SystemHealthService{
			{
				Name:          "order-api",
				Score:         12,
				Metrics:       healthyMetrics,
				MetricDetails: defaultSystemMetricDetails(),
			},
			{
				Name:          "order-worker",
				Score:         12,
				Metrics:       defaultSystemMetricStatus(),
				MetricDetails: defaultSystemMetricDetails(),
			},
		},
	}
	extra := systemHealthExtraMetricResults{
		ResourceShift: []prometheus.MetricResult{
			{
				Metric: prometheus.Labels{SvcName: "order-worker"},
				Values: []prometheus.Points{{Value: 1.75}},
			},
		},
	}

	applySystemHealthExtraMetrics(&system, extra)

	orderAPI := serviceByName(t, system.Services, "order-api")
	if orderAPI.Score != 100 {
		t.Fatalf("expected service without extra samples to be rescored from existing metrics, got %d", orderAPI.Score)
	}
}

func TestApplySystemHealthExtraMetricsUsesCurrentMetricValues(t *testing.T) {
	orderMetrics := defaultSystemMetricStatus()
	orderMetrics["resource"] = healthStatusRed
	orderMetrics["latency"] = healthStatusRed
	orderMetrics["errorRate"] = healthStatusYellow
	orderMetrics["errorLog"] = healthStatusGreen
	orderMetrics["badRequests"] = healthStatusYellow
	workerMetrics := defaultSystemMetricStatus()
	workerMetrics["resource"] = healthStatusGreen
	workerMetrics["latency"] = healthStatusYellow
	workerMetrics["errorRate"] = healthStatusGreen
	workerMetrics["errorLog"] = healthStatusGreen
	workerMetrics["badRequests"] = healthStatusGreen
	system := response.SystemHealthSystem{
		Metrics:       defaultSystemMetricStatus(),
		MetricDetails: defaultSystemMetricDetails(),
		Services: []response.SystemHealthService{
			{
				Name:          "order-api",
				Metrics:       orderMetrics,
				MetricDetails: defaultSystemMetricDetails(),
			},
			{
				Name:          "order-worker",
				Metrics:       workerMetrics,
				MetricDetails: defaultSystemMetricDetails(),
			},
		},
	}
	extra := systemHealthExtraMetricResults{
		ResourceCPU: []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "order-api"}, Values: []prometheus.Points{{Value: 0.63}}},
			{Metric: prometheus.Labels{SvcName: "order-worker"}, Values: []prometheus.Points{{Value: 0.42}}},
		},
		ResourceMemory: []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "order-api"}, Values: []prometheus.Points{{Value: 0.91}}},
			{Metric: prometheus.Labels{SvcName: "order-worker"}, Values: []prometheus.Points{{Value: 0.56}}},
		},
		Latency: []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "order-api"}, Values: []prometheus.Points{{Value: 900_000_000}}},
			{Metric: prometheus.Labels{SvcName: "order-worker"}, Values: []prometheus.Points{{Value: 120_000_000}}},
		},
		ErrorRate: []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "order-api"}, Values: []prometheus.Points{{Value: 0.123}}},
			{Metric: prometheus.Labels{SvcName: "order-worker"}, Values: []prometheus.Points{{Value: 0}}},
		},
		ErrorLog: []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "order-api"}, Values: []prometheus.Points{{Value: 42}}},
			{Metric: prometheus.Labels{SvcName: "order-worker"}, Values: []prometheus.Points{{Value: 0}}},
		},
		BadRequests: []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "order-api"}, Values: []prometheus.Points{{Value: 3}}},
			{Metric: prometheus.Labels{SvcName: "order-worker"}, Values: []prometheus.Points{{Value: 0}}},
		},
	}

	applySystemHealthExtraMetrics(&system, extra)

	if system.Services[0].MetricDetails["latency"].Value != "900ms" {
		t.Fatalf("expected service latency current value, got %+v", system.Services[0].MetricDetails["latency"])
	}
	if system.Services[0].MetricDetails["errorRate"].Value != "12.30%" {
		t.Fatalf("expected service errorRate current value, got %+v", system.Services[0].MetricDetails["errorRate"])
	}
	if system.Services[0].MetricDetails["errorLog"].Value != "42" {
		t.Fatalf("expected service errorLog current value, got %+v", system.Services[0].MetricDetails["errorLog"])
	}
	if system.Services[0].MetricDetails["badRequests"].Value != "3" {
		t.Fatalf("expected service badRequests current value, got %+v", system.Services[0].MetricDetails["badRequests"])
	}
	if system.Services[0].MetricDetails["resource"].Value != "91.00%" {
		t.Fatalf("expected service resource to use current waterline value, got %+v", system.Services[0].MetricDetails["resource"])
	}
	if system.Services[0].MetricDetails["resource"].Note != "当前内存资源水位" {
		t.Fatalf("expected service resource note to name the limiting resource, got %+v", system.Services[0].MetricDetails["resource"])
	}
	if system.MetricDetails["resource"].Value != "91.00%" {
		t.Fatalf("expected system resource to inherit riskiest current waterline, got %+v", system.MetricDetails["resource"])
	}
	if system.MetricDetails["latency"].Value != "900ms" {
		t.Fatalf("expected system latency to inherit riskiest current value, got %+v", system.MetricDetails["latency"])
	}
	if system.MetricDetails["badRequests"].Value != "3" {
		t.Fatalf("expected system badRequests to inherit active alert count, got %+v", system.MetricDetails["badRequests"])
	}
}

func TestBuildSystemHealthResourceWaterlinePQLJoinsContainerMetricsToServices(t *testing.T) {
	cpuQuery := buildSystemHealthCPUWaterlinePQL(
		prometheus.NewFilter().Equal("cluster_id", "dev").RegexMatch("svc_name", "order-api|order-worker"),
		"15m",
	)
	memoryQuery := buildSystemHealthMemoryWaterlinePQL(
		prometheus.NewFilter().Equal("cluster_id", "dev").RegexMatch("svc_name", "order-api|order-worker"),
		"15m",
	)

	for _, part := range []string{
		`container_cpu_usage_seconds_total{cluster_id="dev",container!="",container!="POD"}[15m]`,
		`container_spec_cpu_quota{cluster_id="dev",container!="",container!="POD"}[15m]`,
		`container_spec_cpu_period{cluster_id="dev",container!="",container!="POD"}[15m]`,
		`kindling_span_trace_duration_nanoseconds_count{cluster_id="dev",svc_name=~"order-api|order-worker",pod!=""}[15m]`,
		`max by (svc_name)`,
		`on (cluster_id,namespace,pod) group_left (svc_name)`,
	} {
		if !strings.Contains(cpuQuery, part) {
			t.Fatalf("expected CPU resource query to contain %q, got %s", part, cpuQuery)
		}
	}

	for _, part := range []string{
		`container_memory_working_set_bytes{cluster_id="dev",container!="",container!="POD"}[15m]`,
		`container_spec_memory_limit_bytes{cluster_id="dev",container!="",container!="POD"}[15m]`,
		`kindling_span_trace_duration_nanoseconds_count{cluster_id="dev",svc_name=~"order-api|order-worker",pod!=""}[15m]`,
		`max by (svc_name)`,
		`on (cluster_id,namespace,pod) group_left (svc_name)`,
	} {
		if !strings.Contains(memoryQuery, part) {
			t.Fatalf("expected memory resource query to contain %q, got %s", part, memoryQuery)
		}
	}
}

func TestSystemHealthVMHostResourceScopeKeepsOnlyVMInstances(t *testing.T) {
	scope := systemHealthVMHostResourceScope{
		NodesByService: map[string]map[string]struct{}{},
		SystemNodes:    map[string]struct{}{},
	}

	scope.addInstance("order-api", &model.ServiceInstance{ServiceName: "order-api", NodeName: "vm-a", Pid: 2201})
	scope.addInstance("order-api", &model.ServiceInstance{ServiceName: "order-api", NodeName: "vm-b", Pid: 2202})
	scope.addInstance("order-api", &model.ServiceInstance{ServiceName: "order-api", NodeName: "node-with-pod", PodName: "order-api-0", Pid: 1})
	scope.addInstance("order-worker", &model.ServiceInstance{ServiceName: "order-worker", NodeName: "node-with-container", ContainerId: "container-1", Pid: 3310})
	scope.addInstance("ignored", &model.ServiceInstance{ServiceName: "ignored", NodeName: "pid-one", Pid: 1})

	if !scope.hasNodes() {
		t.Fatalf("expected VM host scope to contain nodes")
	}
	if got := scope.sortedNodes(); !reflect.DeepEqual(got, []string{"vm-a", "vm-b"}) {
		t.Fatalf("expected only VM hosts, got %+v", got)
	}
	if got := scope.sortedServiceNodes("order-api"); !reflect.DeepEqual(got, []string{"vm-a", "vm-b"}) {
		t.Fatalf("expected order-api VM hosts, got %+v", got)
	}
	if got := scope.sortedServiceNodes("order-worker"); len(got) != 0 {
		t.Fatalf("expected container service to have no VM host nodes, got %+v", got)
	}
}

func TestSystemHealthVMHostResourceWaterlineDetailsMapHostsToServices(t *testing.T) {
	scope := newSystemHealthVMHostResourceScope()
	scope.addInstance("order-api", &model.ServiceInstance{ServiceName: "order-api", NodeName: "vm-a", Pid: 2201})
	scope.addInstance("order-api", &model.ServiceInstance{ServiceName: "order-api", NodeName: "vm-b", Pid: 2202})
	scope.addInstance("order-worker", &model.ServiceInstance{ServiceName: "order-worker", NodeName: "vm-b", Pid: 3301})

	details := systemHealthVMHostResourceWaterlineDetails(
		scope,
		[]prometheus.MetricResult{
			{Metric: prometheus.Labels{NodeName: "vm-a"}, Values: []prometheus.Points{{Value: 0.72}}},
			{Metric: prometheus.Labels{NodeName: "vm-b"}, Values: []prometheus.Points{{Value: 0.42}}},
		},
		[]prometheus.MetricResult{
			{Metric: prometheus.Labels{NodeName: "vm-a"}, Values: []prometheus.Points{{Value: 0.65}}},
			{Metric: prometheus.Labels{NodeName: "vm-b"}, Values: []prometheus.Points{{Value: 0.93}}},
		},
	)

	api := details["order-api"]
	if api.Status != healthStatusRed || api.Value != "93.00%" || api.Source != "host/vm-b" || api.Note != "Current host memory resource waterline" {
		t.Fatalf("expected order-api to use highest related host waterline, got %+v", api)
	}

	worker := details["order-worker"]
	if worker.Status != healthStatusRed || worker.Value != "93.00%" || worker.Source != "host/vm-b" {
		t.Fatalf("expected order-worker to inherit vm-b host waterline, got %+v", worker)
	}
}

func TestSystemHealthVMHostResourceWaterlineDetailsClampValuesAtHundredPercent(t *testing.T) {
	scope := newSystemHealthVMHostResourceScope()
	scope.addInstance("order-api", &model.ServiceInstance{ServiceName: "order-api", NodeName: "vm-a", Pid: 2201})

	details := systemHealthVMHostResourceWaterlineDetails(
		scope,
		[]prometheus.MetricResult{
			{Metric: prometheus.Labels{NodeName: "vm-a"}, Values: []prometheus.Points{{Value: 1.24}}},
		},
		nil,
	)

	api := details["order-api"]
	if api.Status != healthStatusRed || api.Value != "100.00%" || api.Source != "host/vm-a" {
		t.Fatalf("expected VM host resource waterline to clamp at 100%%, got %+v", api)
	}
}

func TestSystemHealthVMHostResourceShiftDetailsMapHostsToServices(t *testing.T) {
	scope := newSystemHealthVMHostResourceScope()
	scope.addInstance("order-api", &model.ServiceInstance{ServiceName: "order-api", NodeName: "vm-a", Pid: 2201})
	scope.addInstance("order-api", &model.ServiceInstance{ServiceName: "order-api", NodeName: "vm-b", Pid: 2202})
	scope.addInstance("order-worker", &model.ServiceInstance{ServiceName: "order-worker", NodeName: "vm-b", Pid: 3301})

	details := systemHealthVMHostResourceShiftDetails(scope, []prometheus.MetricResult{
		{Metric: prometheus.Labels{NodeName: "vm-a"}, Values: []prometheus.Points{{Value: 1.35}}},
		{Metric: prometheus.Labels{NodeName: "vm-b"}, Values: []prometheus.Points{{Value: 1.75}}},
	})

	api := details["order-api"]
	if api.Status != healthStatusRed || api.Value != "75.00%" || api.Source != "host/vm-b" || api.Note != "Host CPU/memory resource shift against 24h baseline" {
		t.Fatalf("expected order-api to use highest related host shift, got %+v", api)
	}

	worker := details["order-worker"]
	if worker.Status != healthStatusRed || worker.Value != "75.00%" || worker.Source != "host/vm-b" {
		t.Fatalf("expected order-worker to inherit vm-b host shift, got %+v", worker)
	}
}

func TestBuildSystemHealthVMHostResourcePQLUsesNodeExporterInstanceName(t *testing.T) {
	nodes := []string{"vm-a", "vm-b"}

	cpuQuery := buildSystemHealthVMHostCPUWaterlinePQL(nodes, "15m", "")
	memoryQuery := buildSystemHealthVMHostMemoryWaterlinePQL(nodes, "15m", "")
	shiftQuery := buildSystemHealthVMHostResourceShiftPQL(nodes, "15m")

	for _, part := range []string{
		`node_cpu_seconds_total{instance_name=~"vm-a|vm-b",mode!~"idle|iowait|steal"}[15m]`,
		`label_replace(rate(`,
		`"node_name", "$1", "instance_name", "(.*)"`,
		`sum by (node_name,cpu)`,
		`avg by (node_name)`,
	} {
		if !strings.Contains(cpuQuery, part) {
			t.Fatalf("expected VM host CPU query to contain %q, got %s", part, cpuQuery)
		}
	}

	for _, part := range []string{
		`node_memory_MemAvailable_bytes{instance_name=~"vm-a|vm-b"}[15m]`,
		`node_memory_MemTotal_bytes{instance_name=~"vm-a|vm-b"}[15m]`,
		`label_replace(avg_over_time(`,
		`"node_name", "$1", "instance_name", "(.*)"`,
		`max by (node_name)`,
	} {
		if !strings.Contains(memoryQuery, part) {
			t.Fatalf("expected VM host memory query to contain %q, got %s", part, memoryQuery)
		}
	}

	for _, part := range []string{
		`offset 24h`,
		`max by (node_name)`,
		`1 / (`,
		`node_cpu_seconds_total{instance_name=~"vm-a|vm-b",mode!~"idle|iowait|steal"}[15m]`,
		`node_memory_MemAvailable_bytes{instance_name=~"vm-a|vm-b"}[15m]`,
	} {
		if !strings.Contains(shiftQuery, part) {
			t.Fatalf("expected VM host shift query to contain %q, got %s", part, shiftQuery)
		}
	}
}

func TestBuildSystemHealthLockCostPQLForInstancesNarrowsPolarisByPods(t *testing.T) {
	query := buildSystemHealthLockCostPQLForInstances([]*model.ServiceInstance{
		{
			ServiceName: "ts-order-service",
			ClusterID:   "dev",
			Namespace:   "train-ticket",
			PodName:     "ts-order-service-abc",
		},
	}, "15m")

	mustContain := []string{
		`originx_thread_polaris_nanoseconds_sum{cluster_id="dev",namespace="train-ticket",pod="ts-order-service-abc",type="futex"}`,
		`originx_thread_polaris_nanoseconds_count{cluster_id="dev",namespace="train-ticket",pod="ts-order-service-abc",type="cpu"}`,
		`kindling_span_trace_duration_nanoseconds_count{cluster_id="dev",namespace="train-ticket",pod="ts-order-service-abc",svc_name="ts-order-service"}`,
		`on (cluster_id,namespace,pod) group_left (svc_name)`,
	}
	for _, part := range mustContain {
		if !strings.Contains(query, part) {
			t.Fatalf("expected query to contain %q, got %s", part, query)
		}
	}
}

func TestApplySystemHealthExtraMetricsAddsSystemTrendSeries(t *testing.T) {
	system := response.SystemHealthSystem{
		Metrics:       defaultSystemMetricStatus(),
		MetricDetails: defaultSystemMetricDetails(),
		Services: []response.SystemHealthService{
			{
				Name:          "order-api",
				Score:         80,
				Metrics:       map[string]string{healthMetricLatency: healthStatusYellow, healthMetricErrorRate: healthStatusRed, healthMetricErrorLog: healthStatusYellow},
				MetricDetails: defaultSystemMetricDetails(),
			},
			{
				Name:          "order-worker",
				Score:         90,
				Metrics:       map[string]string{healthMetricLatency: healthStatusGreen, healthMetricErrorRate: healthStatusGreen, healthMetricErrorLog: healthStatusGreen},
				MetricDetails: defaultSystemMetricDetails(),
			},
		},
	}
	extra := systemHealthExtraMetricResults{
		SystemLatency: []prometheus.MetricResult{
			{Values: []prometheus.Points{{Value: 480_000_000}}},
		},
		SystemErrorRate: []prometheus.MetricResult{
			{Values: []prometheus.Points{{Value: 0.057}}},
		},
		ErrorLog: []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "order-api"}, Values: []prometheus.Points{{Value: 4}}},
			{Metric: prometheus.Labels{SvcName: "order-worker"}, Values: []prometheus.Points{{Value: 3}}},
		},
		SystemLatencyTrend: []prometheus.MetricResult{
			{Values: []prometheus.Points{{TimeStamp: 101, Value: 120_000_000}, {TimeStamp: 102, Value: 360_000_000}}},
		},
		SystemErrorRateTrend: []prometheus.MetricResult{
			{Values: []prometheus.Points{{TimeStamp: 101, Value: 0.012}, {TimeStamp: 102, Value: 0.034}}},
		},
		SystemErrorLogTrend: []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "order-api"}, Values: []prometheus.Points{{TimeStamp: 101, Value: 2}, {TimeStamp: 102, Value: 5}}},
			{Metric: prometheus.Labels{SvcName: "order-worker"}, Values: []prometheus.Points{{TimeStamp: 101, Value: 1}, {TimeStamp: 102, Value: 4}}},
		},
	}

	applySystemHealthExtraMetrics(&system, extra)

	if system.MetricDetails[healthMetricLatency].Value != "480ms" {
		t.Fatalf("expected system latency current value to use system aggregate, got %+v", system.MetricDetails[healthMetricLatency])
	}
	if system.MetricDetails[healthMetricLatency].Unit != "ms" {
		t.Fatalf("expected latency unit ms, got %+v", system.MetricDetails[healthMetricLatency])
	}
	if got := system.MetricDetails[healthMetricLatency].Series; len(got) != 2 || got[0].Value != 120 || got[1].Value != 360 {
		t.Fatalf("expected latency trend converted to ms, got %+v", got)
	}
	if system.MetricDetails[healthMetricErrorRate].Value != "5.70%" {
		t.Fatalf("expected system error rate current value to use request-weighted aggregate, got %+v", system.MetricDetails[healthMetricErrorRate])
	}
	if got := system.MetricDetails[healthMetricErrorRate].Series; len(got) != 2 || got[0].Value != 1.2 || got[1].Value != 3.4 {
		t.Fatalf("expected error rate trend converted to percent, got %+v", got)
	}
	if system.MetricDetails[healthMetricErrorLog].Value != "7" {
		t.Fatalf("expected system error log current value to sum service logs, got %+v", system.MetricDetails[healthMetricErrorLog])
	}
	if got := system.MetricDetails[healthMetricErrorLog].Series; len(got) != 2 || got[0].Value != 3 || got[1].Value != 9 {
		t.Fatalf("expected error log trend to sum services per timestamp, got %+v", got)
	}
}

func TestApplySystemHealthExtraMetricsAddsResourceShiftTrendThresholds(t *testing.T) {
	system := response.SystemHealthSystem{
		Metrics:       defaultSystemMetricStatus(),
		MetricDetails: defaultSystemMetricDetails(),
		Services: []response.SystemHealthService{
			{
				Name:          "order-api",
				Score:         80,
				Metrics:       map[string]string{healthMetricResourceShift: healthStatusYellow},
				MetricDetails: defaultSystemMetricDetails(),
			},
			{
				Name:          "order-worker",
				Score:         90,
				Metrics:       map[string]string{healthMetricResourceShift: healthStatusGreen},
				MetricDetails: defaultSystemMetricDetails(),
			},
		},
	}
	extra := systemHealthExtraMetricResults{
		ResourceShift: []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "order-api"}, Values: []prometheus.Points{{Value: 1.45}}},
		},
		SystemResourceShiftTrend: []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "order-api"}, Values: []prometheus.Points{{TimeStamp: 201, Value: 1.10}, {TimeStamp: 202, Value: 1.75}}},
			{Metric: prometheus.Labels{SvcName: "order-worker"}, Values: []prometheus.Points{{TimeStamp: 201, Value: 1.35}, {TimeStamp: 202, Value: 1.20}}},
		},
	}

	applySystemHealthExtraMetrics(&system, extra)

	detail := system.MetricDetails[healthMetricResourceShift]
	if detail.Unit != "%" {
		t.Fatalf("expected resource shift unit percent, got %+v", detail)
	}
	if got := detail.Series; len(got) != 2 || got[0].Value != 35 || got[1].Value != 75 {
		t.Fatalf("expected resource shift trend to use max system shift percent, got %+v", got)
	}
	if len(detail.Thresholds) != 2 || detail.Thresholds[0].Value != 30 || detail.Thresholds[1].Value != 60 {
		t.Fatalf("expected resource shift threshold lines at 30 and 60, got %+v", detail.Thresholds)
	}
}

func TestApplySystemHealthExtraMetricsUsesSystemPodWaterlineSource(t *testing.T) {
	system := response.SystemHealthSystem{
		Metrics:       defaultSystemMetricStatus(),
		MetricDetails: defaultSystemMetricDetails(),
		Services: []response.SystemHealthService{
			{
				Name:          "order-api",
				Score:         95,
				Metrics:       map[string]string{healthMetricResource: healthStatusGreen},
				MetricDetails: defaultSystemMetricDetails(),
			},
		},
	}
	extra := systemHealthExtraMetricResults{
		SystemResourceCPU: []prometheus.MetricResult{
			{Metric: prometheus.Labels{ClusterID: "dev", Namespace: "prod", POD: "order-a"}, Values: []prometheus.Points{{Value: 0.72}}},
		},
		SystemResourceMemory: []prometheus.MetricResult{
			{Metric: prometheus.Labels{ClusterID: "dev", Namespace: "prod", POD: "order-b"}, Values: []prometheus.Points{{Value: 0.93}}},
		},
	}

	applySystemHealthExtraMetrics(&system, extra)

	detail := system.MetricDetails[healthMetricResource]
	if system.Metrics[healthMetricResource] != healthStatusRed || detail.Status != healthStatusRed {
		t.Fatalf("expected system resource waterline to become red, metrics=%+v detail=%+v", system.Metrics, detail)
	}
	if detail.Value != "93.00%" {
		t.Fatalf("expected highest Pod waterline value, got %+v", detail)
	}
	if detail.Source != "dev/prod/order-b" {
		t.Fatalf("expected waterline source to identify the limiting Pod, got %+v", detail)
	}
	if len(detail.Series) != 0 {
		t.Fatalf("resource waterline should remain current snapshot only, got %+v", detail.Series)
	}
}

func TestApplySystemHealthExtraMetricsClampsSystemPodWaterlineAtHundredPercent(t *testing.T) {
	system := response.SystemHealthSystem{
		Metrics:       defaultSystemMetricStatus(),
		MetricDetails: defaultSystemMetricDetails(),
		Services: []response.SystemHealthService{
			{
				Name:          "order-api",
				Score:         95,
				Metrics:       map[string]string{healthMetricResource: healthStatusGreen},
				MetricDetails: defaultSystemMetricDetails(),
			},
		},
	}
	extra := systemHealthExtraMetricResults{
		SystemResourceCPU: []prometheus.MetricResult{
			{Metric: prometheus.Labels{ClusterID: "dev", Namespace: "prod", POD: "order-a"}, Values: []prometheus.Points{{Value: 1.42}}},
		},
	}

	applySystemHealthExtraMetrics(&system, extra)

	detail := system.MetricDetails[healthMetricResource]
	if system.Metrics[healthMetricResource] != healthStatusRed || detail.Value != "100.00%" {
		t.Fatalf("expected system resource waterline to clamp at 100%%, got %+v", detail)
	}
}

func TestBuildSystemHealthSystemWaterlinePQLKeepsPodGranularity(t *testing.T) {
	query := buildSystemHealthSystemCPUWaterlinePQL(
		prometheus.NewFilter().Equal("cluster_id", "dev").RegexMatch("svc_name", "order-api|order-worker"),
		"15m",
	)

	for _, part := range []string{
		`sum by (cluster_id,namespace,pod) (rate(container_cpu_usage_seconds_total{cluster_id="dev",container!="",container!="POD"}[15m]))`,
		`kindling_span_trace_duration_nanoseconds_count{cluster_id="dev",svc_name=~"order-api|order-worker",pod!=""}[15m]`,
		`max by (cluster_id,namespace,pod)`,
		`on (cluster_id,namespace,pod)`,
	} {
		if !strings.Contains(query, part) {
			t.Fatalf("expected system CPU waterline query to contain %q, got %s", part, query)
		}
	}
	if strings.Contains(query, `group_left (svc_name)`) || strings.Contains(query, `max by (svc_name)`) {
		t.Fatalf("system waterline should keep Pod granularity instead of service representative grouping, got %s", query)
	}
}
