// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package systemhealth

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
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

func systemHealthMetricBadgesContain(badges []response.SystemHealthMetricBadge, label string, value string) bool {
	for _, badge := range badges {
		if badge.Label == label && badge.Value == value {
			return true
		}
	}
	return false
}

func TestBuildSystemHealthSystemKeepsSystemMetricsGrayUntilAggregateData(t *testing.T) {
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
	if system.Score != 0 || system.Status != healthStatusGray {
		t.Fatalf("expected system score/status to stay no-data until aggregate metrics arrive, got score=%d status=%s", system.Score, system.Status)
	}
	for key, status := range system.Metrics {
		if status != "gray" {
			t.Fatalf("expected system metric %s to stay gray until aggregate data arrives, got %+v", key, system.Metrics)
		}
	}
	if len(system.Services) != 2 {
		t.Fatalf("expected two services, got %d", len(system.Services))
	}
	if system.Services[0].Metrics["latency"] != "red" {
		t.Fatalf("expected service latency red, got %+v", system.Services[0].Metrics)
	}
}

func TestApplySystemHealthExtraMetricsDoesNotFallbackToServiceStatusWhenAggregateMissing(t *testing.T) {
	system := response.SystemHealthSystem{
		Metrics:       defaultSystemMetricStatus(),
		MetricDetails: defaultSystemMetricDetails(),
		Services: []response.SystemHealthService{
			{
				Name:          "small-api",
				Metrics:       defaultSystemMetricStatus(),
				MetricDetails: defaultSystemMetricDetails(),
			},
			{
				Name:          "bulk-api",
				Metrics:       defaultSystemMetricStatus(),
				MetricDetails: defaultSystemMetricDetails(),
			},
		},
	}
	extra := systemHealthExtraMetricResults{
		Latency: []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "small-api"}, Values: []prometheus.Points{{Value: 1_200_000_000}}},
			{Metric: prometheus.Labels{SvcName: "bulk-api"}, Values: []prometheus.Points{{Value: 80_000_000}}},
		},
		RequestCount: []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "small-api"}, Values: []prometheus.Points{{Value: 100}}},
			{Metric: prometheus.Labels{SvcName: "bulk-api"}, Values: []prometheus.Points{{Value: 1000}}},
		},
	}

	applySystemHealthExtraMetrics(&system, extra)

	if system.Metrics[healthMetricLatency] != healthStatusGray {
		t.Fatalf("expected missing aggregate latency to stay gray instead of service red, metrics=%+v detail=%+v", system.Metrics, system.MetricDetails[healthMetricLatency])
	}
	if system.Score != 0 || system.Status != healthStatusGray {
		t.Fatalf("expected missing aggregate metrics to leave system score/status no-data, got score=%d status=%s", system.Score, system.Status)
	}
	detail := system.MetricDetails[healthMetricLatency]
	if detail.Status != healthStatusGray || detail.Value != "--" {
		t.Fatalf("expected missing aggregate latency detail to remain a no-data main value, got %+v", detail)
	}
	if !systemHealthMetricBadgesContain(detail.Badges, "响应最慢服务", "small-api") {
		t.Fatalf("expected service contributor to remain available for tooltip, got %+v", detail.Badges)
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

func TestBuildSystemHealthSystemWithServiceScopeIncludesNoDataServices(t *testing.T) {
	group := datagroup.DataGroup{GroupID: 101, GroupName: "订单系统"}
	services := []*response.ServiceRYGResult{
		{
			ServiceName: "order-api",
			RYGResult: response.RYGResult{
				PercentScore: 40,
				ScoreDetail: []response.RYGScoreDetail{
					{Key: "latency", Score: 0, Detail: "latency too high"},
				},
			},
		},
	}

	system := buildSystemHealthSystemWithServiceScope(group, []string{"order-api", "order-worker"}, services)

	if len(system.Services) != 2 {
		t.Fatalf("expected scoped no-data service to be included, got %d services: %+v", len(system.Services), system.Services)
	}
	orderWorker := serviceByName(t, system.Services, "order-worker")
	if orderWorker.Status != healthStatusGray || orderWorker.Score != 0 {
		t.Fatalf("expected missing RYG service to be gray no-data, got %+v", orderWorker)
	}
	for key, status := range orderWorker.Metrics {
		if status != healthStatusGray {
			t.Fatalf("expected no-data service metric %s to be gray, got %s", key, status)
		}
	}
	if system.Score != 0 || system.Status != healthStatusGray {
		t.Fatalf("expected system score/status to stay no-data until aggregate metrics arrive, got score=%d status=%s", system.Score, system.Status)
	}
}

func TestBuildSystemHealthSystemWithServiceScopeKeepsAllNoDataSystemGray(t *testing.T) {
	group := datagroup.DataGroup{GroupID: 101, GroupName: "订单系统"}

	system := buildSystemHealthSystemWithServiceScope(group, []string{"order-api", "order-worker"}, nil)

	if len(system.Services) != 2 {
		t.Fatalf("expected scoped services to be included, got %d services: %+v", len(system.Services), system.Services)
	}
	if system.Status != healthStatusGray || system.Score != 0 {
		t.Fatalf("expected all-no-data system to stay gray, got %+v", system)
	}
	for _, service := range system.Services {
		if service.Status != healthStatusGray {
			t.Fatalf("expected all services to be gray, got %+v", system.Services)
		}
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

func TestBuildSystemHealthSystemsPageKeepsGlobalServiceCountsWhilePaging(t *testing.T) {
	scoreThresholds := response.SystemHealthConfigResp{
		SevereAbnormalScoreThreshold:  60,
		GeneralAbnormalScoreThreshold: 80,
	}
	systems := []response.SystemHealthSystem{
		{
			GroupID: 101,
			Name:    "订单系统",
			Score:   42,
			Services: []response.SystemHealthService{
				{Name: "order-api", Score: 42, Status: healthStatusRed},
				{Name: "order-worker", Score: 72, Status: healthStatusYellow},
				{Name: "order-cache", Score: 91, Status: healthStatusGreen},
			},
		},
		{
			GroupID: 102,
			Name:    "支付系统",
			Score:   74,
			Services: []response.SystemHealthService{
				{Name: "payment-api", Score: 74, Status: healthStatusYellow},
			},
		},
		{
			GroupID: 103,
			Name:    "库存系统",
			Score:   95,
			Services: []response.SystemHealthService{
				{Name: "inventory-api", Score: 95, Status: healthStatusGreen},
			},
		},
	}

	resp := buildSystemHealthSystemsPage(systems, systemHealthListOptions{
		Page:     1,
		PageSize: 1,
		Status:   "all",
		Sort:     "score_asc",
	}, scoreThresholds, "last30m", "2026-07-01T10:00:00Z")

	if resp.Total != 3 || resp.Page != 1 || resp.PageSize != 1 {
		t.Fatalf("expected first page metadata over all systems, got %+v", resp)
	}
	if len(resp.Systems) != 1 {
		t.Fatalf("expected one paged system row, got %d", len(resp.Systems))
	}
	row := resp.Systems[0]
	if row.GroupID != 101 {
		t.Fatalf("expected lowest-score system on first page, got %+v", row)
	}
	if len(row.Services) != 0 {
		t.Fatalf("expected pageable system row to omit service payload, got %d services", len(row.Services))
	}
	if row.ServiceTotal != 3 || row.AbnormalServiceCount != 1 || row.AttentionServiceCount != 1 {
		t.Fatalf("expected service counts from all services in the system, got %+v", row)
	}

	summary := buildSystemHealthOverviewSummary(systems, scoreThresholds)
	if summary.TotalSystems != 3 || summary.Services != 5 {
		t.Fatalf("expected global summary to still use all systems and services, got %+v", summary)
	}
}

func TestBuildSystemHealthSystemsPageSortsNoDataSystemsLast(t *testing.T) {
	systems := []response.SystemHealthSystem{
		{GroupID: 101, Name: "无数据系统", Score: 0, Status: healthStatusGray},
		{GroupID: 102, Name: "真实零分系统", Score: 0, Status: healthStatusRed},
		{GroupID: 103, Name: "异常系统", Score: 30, Status: healthStatusRed},
		{GroupID: 104, Name: "健康系统", Score: 95, Status: healthStatusGreen},
	}

	resp := buildSystemHealthSystemsPage(systems, systemHealthListOptions{
		Page:     1,
		PageSize: 10,
		Status:   "all",
		Sort:     "score_asc",
	}, defaultSystemHealthScoreThresholds(), "last30m", "2026-07-01T10:00:00Z")

	got := make([]int64, 0, len(resp.Systems))
	for _, system := range resp.Systems {
		got = append(got, system.GroupID)
	}
	want := []int64{102, 103, 104, 101}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected scored systems sorted by low health score with no-data last, got %+v", got)
	}
}

func TestBuildSystemHealthServicesPageReturnsRequestedPageAndTotal(t *testing.T) {
	system := response.SystemHealthSystem{
		GroupID: 101,
		Name:    "订单系统",
		Score:   75,
		Status:  healthStatusYellow,
		Services: []response.SystemHealthService{
			{Name: "healthy-api", Score: 95, Status: healthStatusGreen},
			{Name: "critical-api", Score: 35, Status: healthStatusRed},
			{Name: "warning-api", Score: 70, Status: healthStatusYellow},
		},
	}

	resp := buildSystemHealthServicesPage(system, systemHealthListOptions{
		Page:     2,
		PageSize: 1,
		Status:   "all",
		Sort:     "score_asc",
	}, defaultSystemHealthScoreThresholds(), "last30m", "2026-07-01T10:00:00Z")

	if resp.Total != 3 || resp.Page != 2 || resp.PageSize != 1 {
		t.Fatalf("expected second service page metadata, got %+v", resp)
	}
	if len(resp.Services) != 1 || resp.Services[0].Name != "warning-api" {
		t.Fatalf("expected second sorted service page to contain warning-api, got %+v", resp.Services)
	}
}

func TestBuildSystemHealthServicesPageSortsNoDataServicesLast(t *testing.T) {
	system := response.SystemHealthSystem{
		GroupID: 101,
		Name:    "订单系统",
		Score:   75,
		Status:  healthStatusYellow,
		Services: []response.SystemHealthService{
			{Name: "no-data-api", Score: 0, Status: healthStatusGray},
			{Name: "computed-zero-api", Score: 0, Status: healthStatusRed},
			{Name: "critical-api", Score: 30, Status: healthStatusRed},
			{Name: "healthy-api", Score: 95, Status: healthStatusGreen},
		},
	}

	resp := buildSystemHealthServicesPage(system, systemHealthListOptions{
		Page:     1,
		PageSize: 10,
		Status:   "all",
		Sort:     "score_asc",
	}, defaultSystemHealthScoreThresholds(), "last30m", "2026-07-01T10:00:00Z")

	got := make([]string, 0, len(resp.Services))
	for _, service := range resp.Services {
		got = append(got, service.Name)
	}
	want := []string{"computed-zero-api", "critical-api", "healthy-api", "no-data-api"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected scored services sorted by low health score with no-data last, got %+v", got)
	}
}

func TestBuildSystemHealthServicesPageFiltersWithConfiguredThresholds(t *testing.T) {
	system := response.SystemHealthSystem{
		GroupID: 101,
		Name:    "订单系统",
		Services: []response.SystemHealthService{
			{Name: "borderline-api", Score: 80},
			{Name: "healthy-api", Score: 95},
		},
	}
	scoreThresholds := response.SystemHealthConfigResp{
		SevereAbnormalScoreThreshold:  50,
		GeneralAbnormalScoreThreshold: 90,
	}

	resp := buildSystemHealthServicesPage(system, systemHealthListOptions{
		Page:     1,
		PageSize: 10,
		Status:   "attention",
	}, scoreThresholds, "last30m", "2026-07-01T10:00:00Z")

	if resp.Total != 1 || len(resp.Services) != 1 || resp.Services[0].Name != "borderline-api" {
		t.Fatalf("expected custom thresholds to classify borderline-api as attention, got %+v", resp)
	}
}

func TestBuildSystemHealthServicesPageFromCurrentFindsSystemAndUsesCurrentMetadata(t *testing.T) {
	current := systemHealthCurrentResult{
		Systems: []response.SystemHealthSystem{
			{
				GroupID: 101,
				Name:    "订单系统",
				Services: []response.SystemHealthService{
					{Name: "borderline-api", Score: 80},
					{Name: "healthy-api", Score: 95},
				},
			},
		},
		Config: response.SystemHealthConfigResp{
			SevereAbnormalScoreThreshold:  50,
			GeneralAbnormalScoreThreshold: 90,
		},
		Window:      "last30m",
		GeneratedAt: "2026-07-01T10:00:00Z",
	}

	resp, ok := buildSystemHealthServicesPageFromCurrent(current, 101, systemHealthListOptions{
		Page:     1,
		PageSize: 10,
		Status:   "attention",
	})

	if !ok {
		t.Fatal("expected system services page to be built from current health result")
	}
	if resp.Window != current.Window || resp.GeneratedAt != current.GeneratedAt {
		t.Fatalf("expected current metadata to be preserved, got %+v", resp)
	}
	if resp.Total != 1 || len(resp.Services) != 1 || resp.Services[0].Name != "borderline-api" {
		t.Fatalf("expected current thresholds to classify borderline-api as attention, got %+v", resp)
	}
}

func TestBuildSystemHealthSystemsPageClonesPagedRows(t *testing.T) {
	systems := []response.SystemHealthSystem{
		{
			GroupID: 101,
			Name:    "订单系统",
			Score:   72,
			Status:  healthStatusYellow,
			Metrics: map[string]string{healthMetricLatency: healthStatusYellow},
			MetricDetails: map[string]response.SystemHealthMetricDetail{
				healthMetricLatency: {
					Status: healthStatusYellow,
					Value:  "800ms",
					Series: []response.SystemHealthMetricPoint{{Timestamp: 1782972000000000, Value: 800}},
				},
			},
		},
	}

	resp := buildSystemHealthSystemsPage(systems, systemHealthListOptions{
		Page:     1,
		PageSize: 10,
		Status:   "all",
	}, defaultSystemHealthScoreThresholds(), "last30m", "2026-07-01T10:00:00Z")
	resp.Systems[0].Metrics[healthMetricLatency] = healthStatusGreen
	detail := resp.Systems[0].MetricDetails[healthMetricLatency]
	detail.Series[0].Value = 1

	if systems[0].Metrics[healthMetricLatency] != healthStatusYellow {
		t.Fatalf("expected systems page row metrics to be isolated, got %+v", systems[0].Metrics)
	}
	if systems[0].MetricDetails[healthMetricLatency].Series[0].Value != 800 {
		t.Fatalf("expected systems page row metric details to be isolated, got %+v", systems[0].MetricDetails[healthMetricLatency])
	}
}

func TestBuildSystemHealthServicesPageClonesPagedServices(t *testing.T) {
	system := response.SystemHealthSystem{
		GroupID: 101,
		Name:    "订单系统",
		Services: []response.SystemHealthService{
			{
				Name:    "order-api",
				Score:   72,
				Status:  healthStatusYellow,
				Metrics: map[string]string{healthMetricLatency: healthStatusYellow},
				MetricDetails: map[string]response.SystemHealthMetricDetail{
					healthMetricLatency: {
						Status: healthStatusYellow,
						Value:  "800ms",
						Series: []response.SystemHealthMetricPoint{{Timestamp: 1782972000000000, Value: 800}},
					},
				},
			},
		},
	}

	resp := buildSystemHealthServicesPage(system, systemHealthListOptions{
		Page:     1,
		PageSize: 10,
		Status:   "all",
	}, defaultSystemHealthScoreThresholds(), "last30m", "2026-07-01T10:00:00Z")
	resp.Services[0].Metrics[healthMetricLatency] = healthStatusGreen
	detail := resp.Services[0].MetricDetails[healthMetricLatency]
	detail.Series[0].Value = 1

	if system.Services[0].Metrics[healthMetricLatency] != healthStatusYellow {
		t.Fatalf("expected service page metrics to be isolated, got %+v", system.Services[0].Metrics)
	}
	if system.Services[0].MetricDetails[healthMetricLatency].Series[0].Value != 800 {
		t.Fatalf("expected service page metric details to be isolated, got %+v", system.Services[0].MetricDetails[healthMetricLatency])
	}
}

func TestBuildSystemHealthPendingIssuesPageIncludesIssuesOutsideServicePage(t *testing.T) {
	systems := []response.SystemHealthSystem{
		{
			GroupID: 101,
			Name:    "订单系统",
			Score:   75,
			Services: []response.SystemHealthService{
				{
					Name:          "healthy-api",
					Score:         95,
					Status:        healthStatusGreen,
					Metrics:       map[string]string{healthMetricLatency: healthStatusGreen},
					MetricDetails: defaultSystemMetricDetails(),
				},
				{
					Name:    "critical-api",
					Score:   35,
					Status:  healthStatusRed,
					Metrics: map[string]string{healthMetricErrorRate: healthStatusRed},
					MetricDetails: map[string]response.SystemHealthMetricDetail{
						healthMetricErrorRate: {Status: healthStatusRed, Value: "8.1%", Note: "错误率过高"},
					},
				},
			},
		},
	}

	resp := buildSystemHealthPendingIssuesPage(systems, systemHealthPendingIssueListOptions{
		Page:     1,
		PageSize: 20,
		Status:   "all",
	}, "last30m", "2026-07-01T10:00:00Z")

	if resp.Total != 1 || len(resp.Issues) != 1 {
		t.Fatalf("expected pending issue outside the first service row to be included, got %+v", resp)
	}
	if resp.Issues[0].ServiceName != "critical-api" || resp.Issues[0].MetricKey != healthMetricErrorRate {
		t.Fatalf("expected critical-api errorRate issue, got %+v", resp.Issues[0])
	}
}

func TestSystemHealthOverviewSummaryOnlyResponseOmitsHeavyLists(t *testing.T) {
	full := response.SystemHealthOverviewResp{
		Summary: response.SystemHealthOverviewSummary{
			TotalSystems: 2,
			Services:     3,
		},
		Systems: []response.SystemHealthSystem{
			{GroupID: 101, Name: "订单系统"},
		},
		PendingIssues: []response.SystemHealthPendingIssue{
			{ID: "101:order-api:latency", GroupID: 101, ServiceName: "order-api"},
		},
		Config: response.SystemHealthConfigResp{
			SevereAbnormalScoreThreshold:  60,
			GeneralAbnormalScoreThreshold: 80,
		},
		Window:      "last30m",
		GeneratedAt: "2026-07-02T10:00:00Z",
		Source:      systemHealthSourceDataGroup,
	}

	payload, err := json.Marshal(systemHealthOverviewSummaryOnlyResponse(full))
	if err != nil {
		t.Fatalf("marshal summary-only response: %v", err)
	}
	payloadText := string(payload)
	if strings.Contains(payloadText, "systems") || strings.Contains(payloadText, "pendingIssues") {
		t.Fatalf("expected summary-only response to omit systems and pendingIssues, got %s", payloadText)
	}
	if !strings.Contains(payloadText, "summary") || !strings.Contains(payloadText, "config") {
		t.Fatalf("expected summary-only response to keep summary and config, got %s", payloadText)
	}
	if !strings.Contains(payloadText, "window") || !strings.Contains(payloadText, "generatedAt") || !strings.Contains(payloadText, "source") {
		t.Fatalf("expected summary-only response to keep metadata, got %s", payloadText)
	}
}

func TestSystemHealthSystemDetailHeaderResponseOmitsServices(t *testing.T) {
	full := response.SystemHealthSystemDetailResp{
		System: response.SystemHealthSystem{
			GroupID:               101,
			Name:                  "订单系统",
			Score:                 72,
			Status:                healthStatusYellow,
			Metrics:               defaultSystemMetricStatus(),
			MetricDetails:         defaultSystemMetricDetails(),
			Services:              []response.SystemHealthService{{Name: "order-api"}},
			ServiceTotal:          12,
			AbnormalServiceCount:  2,
			AttentionServiceCount: 3,
			Source:                systemHealthSourceDataGroup,
		},
		Window:      "last30m",
		GeneratedAt: "2026-07-02T10:00:00Z",
		Source:      systemHealthSourceDataGroup,
	}

	payload, err := json.Marshal(systemHealthSystemDetailHeaderResponse(full))
	if err != nil {
		t.Fatalf("marshal detail header response: %v", err)
	}
	payloadText := string(payload)
	if strings.Contains(payloadText, "services") {
		t.Fatalf("expected detail header response to omit services, got %s", payloadText)
	}
	if !strings.Contains(payloadText, "serviceTotal") ||
		!strings.Contains(payloadText, "abnormalServiceCount") ||
		!strings.Contains(payloadText, "attentionServiceCount") {
		t.Fatalf("expected detail header response to keep service counts, got %s", payloadText)
	}
}

func TestSystemHealthSystemDetailHeaderFromCurrentUsesCachedSystemRows(t *testing.T) {
	current := systemHealthCurrentResult{
		Systems: []response.SystemHealthSystem{
			{
				GroupID: 101,
				Name:    "订单系统",
				Score:   72,
				Status:  healthStatusYellow,
				Metrics: map[string]string{
					healthMetricLatency: healthStatusYellow,
				},
				MetricDetails: map[string]response.SystemHealthMetricDetail{
					healthMetricLatency: {Status: healthStatusYellow, Value: "800ms"},
				},
				Services: []response.SystemHealthService{
					{Name: "order-api", Score: 50, Status: healthStatusRed},
					{Name: "pay-api", Score: 75, Status: healthStatusYellow},
					{Name: "user-api", Score: 95, Status: healthStatusGreen},
				},
				Source: systemHealthSourceDataGroup,
			},
		},
		Config: response.SystemHealthConfigResp{
			SevereAbnormalScoreThreshold:  60,
			GeneralAbnormalScoreThreshold: 80,
		},
		Window:      "15m",
		GeneratedAt: "2026-07-02T10:00:00Z",
		Source:      systemHealthSourceDataGroup,
	}

	resp, ok := systemHealthSystemDetailHeaderFromCurrent(current, 101)
	if !ok {
		t.Fatal("expected current health result to contain requested system")
	}
	if resp.System.GroupID != 101 || resp.System.Name != "订单系统" {
		t.Fatalf("unexpected system header: %+v", resp.System)
	}
	if resp.System.ServiceTotal != 3 ||
		resp.System.AbnormalServiceCount != 1 ||
		resp.System.AttentionServiceCount != 1 {
		t.Fatalf("expected service counts from current system services, got %+v", resp.System)
	}
	if resp.Window != "15m" || resp.GeneratedAt != "2026-07-02T10:00:00Z" || resp.Source != systemHealthSourceDataGroup {
		t.Fatalf("expected current metadata, got %+v", resp)
	}

	payload, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal detail header response: %v", err)
	}
	if strings.Contains(string(payload), "services") {
		t.Fatalf("expected current detail header to omit services, got %s", payload)
	}
	if _, ok := systemHealthSystemDetailHeaderFromCurrent(current, 202); ok {
		t.Fatal("expected missing system to return false")
	}
}

func TestSystemHealthSystemDetailHeaderFromCurrentClonesHeaderMaps(t *testing.T) {
	current := systemHealthCurrentResult{
		Systems: []response.SystemHealthSystem{
			{
				GroupID: 101,
				Name:    "订单系统",
				Score:   72,
				Status:  healthStatusYellow,
				Metrics: map[string]string{healthMetricLatency: healthStatusYellow},
				MetricDetails: map[string]response.SystemHealthMetricDetail{
					healthMetricLatency: {
						Status: healthStatusYellow,
						Value:  "800ms",
						Series: []response.SystemHealthMetricPoint{{Timestamp: 1782972000000000, Value: 800}},
					},
				},
			},
		},
		Config: defaultSystemHealthScoreThresholds(),
	}

	resp, ok := systemHealthSystemDetailHeaderFromCurrent(current, 101)
	if !ok {
		t.Fatal("expected current health result to contain requested system")
	}
	resp.System.Metrics[healthMetricLatency] = healthStatusGreen
	detail := resp.System.MetricDetails[healthMetricLatency]
	detail.Series[0].Value = 1

	if current.Systems[0].Metrics[healthMetricLatency] != healthStatusYellow {
		t.Fatalf("expected header metrics to be isolated, got %+v", current.Systems[0].Metrics)
	}
	if current.Systems[0].MetricDetails[healthMetricLatency].Series[0].Value != 800 {
		t.Fatalf("expected header metric details to be isolated, got %+v", current.Systems[0].MetricDetails[healthMetricLatency])
	}
}

func TestBuildSystemHealthOverviewSummaryUsesAggregateSystemScore(t *testing.T) {
	scoreThresholds := response.SystemHealthConfigResp{
		SevereAbnormalScoreThreshold:  60,
		GeneralAbnormalScoreThreshold: 80,
	}
	systems := []response.SystemHealthSystem{
		{
			GroupID: 101,
			Name:    "订单系统",
			Score:   100,
			Services: []response.SystemHealthService{
				{Name: "order-api", Score: 40},
			},
			Metrics: map[string]string{
				healthMetricErrorRate: healthStatusGreen,
			},
		},
		{
			GroupID: 102,
			Name:    "支付系统",
			Score:   0,
			Services: []response.SystemHealthService{
				{Name: "payment-api", Score: 100},
			},
			Metrics: map[string]string{
				healthMetricLatency: healthStatusRed,
			},
		},
		{
			GroupID: 103,
			Name:    "库存系统",
			Score:   0,
			Services: []response.SystemHealthService{
				{Name: "inventory-api", Score: 10},
			},
			Metrics: defaultSystemMetricStatus(),
		},
	}

	summary := buildSystemHealthOverviewSummary(systems, scoreThresholds)

	if summary.AbnormalSystems != 1 {
		t.Fatalf("expected only the aggregate-red system to be abnormal, got %d", summary.AbnormalSystems)
	}
	if summary.LowScoreSystems != 1 {
		t.Fatalf("expected only the aggregate-red system to be low-score, got %d", summary.LowScoreSystems)
	}
	if summary.AbnormalServices != 2 {
		t.Fatalf("expected service abnormal count to remain service-based, got %d", summary.AbnormalServices)
	}
	if summary.AbnormalMetrics != 1 {
		t.Fatalf("expected red/yellow metric counting to stay unchanged, got %d", summary.AbnormalMetrics)
	}
	if summary.SevereMetrics != 1 {
		t.Fatalf("expected red metric counting to stay unchanged, got %d", summary.SevereMetrics)
	}
	if summary.NoDataSystems != 1 {
		t.Fatalf("expected system without aggregate metrics to be no-data, got %d", summary.NoDataSystems)
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

func TestApplySystemHealthScoreThresholdKeepsMissingAggregateSystemGray(t *testing.T) {
	system := response.SystemHealthSystem{
		Score:   0,
		Status:  healthStatusGray,
		Metrics: defaultSystemMetricStatus(),
		Services: []response.SystemHealthService{
			{Name: "order-api", Score: 40, Status: healthStatusYellow},
		},
	}

	applySystemHealthScoreThreshold(&system, response.SystemHealthConfigResp{
		SevereAbnormalScoreThreshold:  60,
		GeneralAbnormalScoreThreshold: 80,
	})

	if system.Status != healthStatusGray {
		t.Fatalf("expected system without aggregate metrics to stay gray, got %+v", system)
	}
	if system.AbnormalServiceCount != 1 {
		t.Fatalf("expected service counts to still use service score thresholds, got %+v", system)
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

func TestSystemHealthCurrentCacheReturnsCachedResultWithoutRebuild(t *testing.T) {
	cache := newSystemHealthCurrentCache(30*time.Second, 8)
	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)
	builds := 0

	first, err := cache.getOrBuild("same-key", now, func() (systemHealthCurrentResult, error) {
		builds++
		return systemHealthCurrentResult{
			Systems:     []response.SystemHealthSystem{{GroupID: 101, Name: "订单系统"}},
			Summary:     response.SystemHealthOverviewSummary{TotalSystems: 1},
			Window:      "15m",
			GeneratedAt: "2026-07-02T10:00:00Z",
			Source:      systemHealthSourceDataGroup,
		}, nil
	})
	if err != nil {
		t.Fatalf("unexpected first build error: %v", err)
	}

	second, err := cache.getOrBuild("same-key", now.Add(time.Second), func() (systemHealthCurrentResult, error) {
		t.Fatal("expected cached result to avoid rebuilding")
		return systemHealthCurrentResult{}, nil
	})
	if err != nil {
		t.Fatalf("unexpected cache hit error: %v", err)
	}
	if builds != 1 {
		t.Fatalf("expected one build, got %d", builds)
	}
	if len(first.Systems) != 1 || len(second.Systems) != 1 || second.Systems[0].GroupID != 101 {
		t.Fatalf("unexpected cached systems: first=%+v second=%+v", first.Systems, second.Systems)
	}
}

func TestSystemHealthCurrentCacheHitPreservesServiceDetails(t *testing.T) {
	cache := newSystemHealthCurrentCache(30*time.Second, 8)
	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)

	_, err := cache.getOrBuild("same-key", now, func() (systemHealthCurrentResult, error) {
		return systemHealthCurrentResult{
			Systems: []response.SystemHealthSystem{
				{
					GroupID: 101,
					Name:    "订单系统",
					Services: []response.SystemHealthService{
						{
							Name:    "order-api",
							Score:   42,
							Status:  healthStatusRed,
							Metrics: map[string]string{healthMetricLatency: healthStatusRed},
							MetricDetails: map[string]response.SystemHealthMetricDetail{
								healthMetricLatency: {
									Status: healthStatusRed,
									Value:  "1200ms",
									Series: []response.SystemHealthMetricPoint{{Timestamp: 1782972000000000, Value: 1200}},
								},
							},
						},
					},
				},
			},
		}, nil
	})
	if err != nil {
		t.Fatalf("unexpected first build error: %v", err)
	}

	cached, err := cache.getOrBuild("same-key", now.Add(time.Second), func() (systemHealthCurrentResult, error) {
		t.Fatal("expected cache hit")
		return systemHealthCurrentResult{}, nil
	})
	if err != nil {
		t.Fatalf("unexpected cache hit error: %v", err)
	}
	if len(cached.Systems) != 1 || len(cached.Systems[0].Services) != 1 {
		t.Fatalf("expected cached service to be preserved, got %+v", cached.Systems)
	}
	service := cached.Systems[0].Services[0]
	if service.Name != "order-api" || service.Metrics[healthMetricLatency] != healthStatusRed {
		t.Fatalf("expected cached service details to be preserved, got %+v", service)
	}
	if got := service.MetricDetails[healthMetricLatency].Series; len(got) != 1 || got[0].Value != 1200 {
		t.Fatalf("expected cached service metric series to be preserved, got %+v", got)
	}

	cached.Systems[0].Services[0].Name = "mutated"
	cached.Systems[0].Services[0].Metrics[healthMetricLatency] = healthStatusGreen
	again, err := cache.getOrBuild("same-key", now.Add(2*time.Second), func() (systemHealthCurrentResult, error) {
		t.Fatal("expected cache hit after mutation")
		return systemHealthCurrentResult{}, nil
	})
	if err != nil {
		t.Fatalf("unexpected second cache hit error: %v", err)
	}
	if again.Systems[0].Services[0].Name != "order-api" || again.Systems[0].Services[0].Metrics[healthMetricLatency] != healthStatusRed {
		t.Fatalf("expected cached data to be isolated from caller mutation, got %+v", again.Systems[0].Services[0])
	}
}

func TestSystemHealthCurrentCacheReadOnlyHitDoesNotCloneFullResult(t *testing.T) {
	cache := newSystemHealthCurrentCache(30*time.Second, 8)
	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)

	_, err := cache.getOrBuild("same-key", now, func() (systemHealthCurrentResult, error) {
		return systemHealthCurrentResult{
			Systems: []response.SystemHealthSystem{
				{
					GroupID: 101,
					Services: []response.SystemHealthService{
						{Name: "order-api", Score: 80},
					},
				},
			},
		}, nil
	})
	if err != nil {
		t.Fatalf("unexpected first build error: %v", err)
	}

	cached, telemetry, ok := cache.getReadOnlyIfFreshOrWait("same-key", now.Add(time.Second))
	if !ok {
		t.Fatal("expected read-only cache hit")
	}
	if telemetry.Status != systemHealthCurrentCacheStatusHit {
		t.Fatalf("expected read-only hit telemetry, got %+v", telemetry)
	}
	entry := cache.entries["same-key"]
	if &cached.Systems[0] != &entry.result.Systems[0] {
		t.Fatal("expected read-only cache hit to return cached systems without cloning")
	}
	if &cached.Systems[0].Services[0] != &entry.result.Systems[0].Services[0] {
		t.Fatal("expected read-only cache hit to return cached services without cloning")
	}
}

func TestSystemHealthCurrentCacheLookupCarriesGroupsForMissReuse(t *testing.T) {
	groups := []*datagroup.DataGroupTreeNode{
		{DataGroup: datagroup.DataGroup{GroupID: 101, GroupName: "订单系统"}},
	}
	lookup := systemHealthCurrentCacheLookup{
		Key:    "cache-key",
		Groups: groups,
		Config: defaultSystemHealthScoreThresholds(),
	}

	if lookup.Key == "" || len(lookup.Groups) != 1 || lookup.Groups[0].GroupID != 101 {
		t.Fatalf("expected cache lookup to carry key and visible groups, got %+v", lookup)
	}
}

func TestSystemHealthVisibleGroupByIDFindsOnlyVisibleGroup(t *testing.T) {
	groups := []*datagroup.DataGroupTreeNode{
		{DataGroup: datagroup.DataGroup{GroupID: 101, GroupName: "订单系统"}},
		{DataGroup: datagroup.DataGroup{GroupID: 202, GroupName: "支付系统"}},
	}

	if group := systemHealthVisibleGroupByID(groups, 202); group == nil || group.GroupID != 202 {
		t.Fatalf("expected visible group 202, got %+v", group)
	}
	if group := systemHealthVisibleGroupByID(groups, 303); group != nil {
		t.Fatalf("expected missing group to return nil, got %+v", group)
	}
}

func TestSystemHealthCurrentCacheSeparatesDifferentKeys(t *testing.T) {
	cache := newSystemHealthCurrentCache(30*time.Second, 8)
	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)
	builds := 0
	build := func(groupID int64) func() (systemHealthCurrentResult, error) {
		return func() (systemHealthCurrentResult, error) {
			builds++
			return systemHealthCurrentResult{
				Systems: []response.SystemHealthSystem{{GroupID: groupID}},
			}, nil
		}
	}

	first, err := cache.getOrBuild("key-1", now, build(101))
	if err != nil {
		t.Fatalf("unexpected first build error: %v", err)
	}
	second, err := cache.getOrBuild("key-2", now.Add(time.Second), build(202))
	if err != nil {
		t.Fatalf("unexpected second build error: %v", err)
	}

	if builds != 2 {
		t.Fatalf("expected different keys to build separately, got %d builds", builds)
	}
	if first.Systems[0].GroupID != 101 || second.Systems[0].GroupID != 202 {
		t.Fatalf("unexpected keyed results: first=%+v second=%+v", first.Systems, second.Systems)
	}
}

func TestSystemHealthCurrentCacheSingleflightsConcurrentBuilds(t *testing.T) {
	cache := newSystemHealthCurrentCache(30*time.Second, 8)
	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)
	started := make(chan struct{})
	release := make(chan struct{})
	var builds int32
	var wg sync.WaitGroup
	errCh := make(chan error, 2)
	results := make(chan systemHealthCurrentResult, 2)

	build := func() (systemHealthCurrentResult, error) {
		if atomic.AddInt32(&builds, 1) == 1 {
			close(started)
		}
		<-release
		return systemHealthCurrentResult{
			Systems: []response.SystemHealthSystem{{GroupID: 101}},
		}, nil
	}
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := cache.getOrBuild("same-key", now, build)
			if err != nil {
				errCh <- err
				return
			}
			results <- result
		}()
	}

	select {
	case <-started:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected first build to start")
	}
	time.Sleep(20 * time.Millisecond)
	close(release)
	wg.Wait()
	close(errCh)
	close(results)

	for err := range errCh {
		t.Fatalf("unexpected concurrent cache error: %v", err)
	}
	if got := atomic.LoadInt32(&builds); got != 1 {
		t.Fatalf("expected one shared build, got %d", got)
	}
	if len(results) != 2 {
		t.Fatalf("expected both callers to receive result, got %d", len(results))
	}
}

func TestSystemHealthCurrentCacheTelemetryReportsMissHitAndShared(t *testing.T) {
	cache := newSystemHealthCurrentCache(30*time.Second, 8)
	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)
	builds := 0

	_, telemetry, err := cache.getOrBuildWithTelemetry("same-key", now, func() (systemHealthCurrentResult, error) {
		builds++
		return systemHealthCurrentResult{
			Systems: []response.SystemHealthSystem{{GroupID: 101}},
		}, nil
	})
	if err != nil {
		t.Fatalf("unexpected first build error: %v", err)
	}
	if telemetry.Status != systemHealthCurrentCacheStatusMiss {
		t.Fatalf("expected cache miss telemetry, got %+v", telemetry)
	}

	_, telemetry, err = cache.getOrBuildWithTelemetry("same-key", now.Add(time.Second), func() (systemHealthCurrentResult, error) {
		t.Fatal("expected hit to avoid rebuild")
		return systemHealthCurrentResult{}, nil
	})
	if err != nil {
		t.Fatalf("unexpected cache hit error: %v", err)
	}
	if telemetry.Status != systemHealthCurrentCacheStatusHit {
		t.Fatalf("expected cache hit telemetry, got %+v", telemetry)
	}
	if builds != 1 {
		t.Fatalf("expected one build after hit, got %d", builds)
	}

	sharedCache := newSystemHealthCurrentCache(30*time.Second, 8)
	started := make(chan struct{})
	release := make(chan struct{})
	telemetryCh := make(chan systemHealthCurrentCacheTelemetry, 2)
	errCh := make(chan error, 2)
	var sharedBuilds int32
	var wg sync.WaitGroup
	build := func() (systemHealthCurrentResult, error) {
		if atomic.AddInt32(&sharedBuilds, 1) == 1 {
			close(started)
		}
		<-release
		return systemHealthCurrentResult{
			Systems: []response.SystemHealthSystem{{GroupID: 202}},
		}, nil
	}
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, telemetry, err := sharedCache.getOrBuildWithTelemetry("shared-key", now, build)
			if err != nil {
				errCh <- err
				return
			}
			telemetryCh <- telemetry
		}()
	}

	select {
	case <-started:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected shared build to start")
	}
	time.Sleep(20 * time.Millisecond)
	close(release)
	wg.Wait()
	close(errCh)
	close(telemetryCh)

	for err := range errCh {
		t.Fatalf("unexpected shared cache error: %v", err)
	}
	statuses := map[systemHealthCurrentCacheStatus]int{}
	for telemetry := range telemetryCh {
		statuses[telemetry.Status]++
	}
	if statuses[systemHealthCurrentCacheStatusMiss] != 1 || statuses[systemHealthCurrentCacheStatusShared] != 1 {
		t.Fatalf("expected one miss and one shared telemetry, got %+v", statuses)
	}
	if got := atomic.LoadInt32(&sharedBuilds); got != 1 {
		t.Fatalf("expected one shared build, got %d", got)
	}
}

func TestSystemHealthCurrentCacheExpiresEntries(t *testing.T) {
	cache := newSystemHealthCurrentCache(30*time.Second, 8)
	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)
	builds := 0
	build := func(groupID int64) func() (systemHealthCurrentResult, error) {
		return func() (systemHealthCurrentResult, error) {
			builds++
			return systemHealthCurrentResult{
				Systems: []response.SystemHealthSystem{{GroupID: groupID}},
			}, nil
		}
	}

	if _, err := cache.getOrBuild("key", now, build(101)); err != nil {
		t.Fatalf("unexpected first build error: %v", err)
	}
	second, err := cache.getOrBuild("key", now.Add(31*time.Second), build(202))
	if err != nil {
		t.Fatalf("unexpected rebuild error: %v", err)
	}
	if builds != 2 {
		t.Fatalf("expected cache entry to expire and rebuild, got %d builds", builds)
	}
	if second.Systems[0].GroupID != 202 {
		t.Fatalf("expected rebuilt result after expiry, got %+v", second.Systems)
	}
}

func TestSystemHealthCurrentCacheKeyIncludesPermissionWindowAndThresholds(t *testing.T) {
	start := time.Date(2026, 7, 2, 9, 45, 0, 0, time.UTC)
	end := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)
	groups := []*datagroup.DataGroupTreeNode{
		{DataGroup: datagroup.DataGroup{GroupID: 101}},
		{DataGroup: datagroup.DataGroup{GroupID: 202}},
	}
	reversedGroups := []*datagroup.DataGroupTreeNode{
		{DataGroup: datagroup.DataGroup{GroupID: 202}},
		{DataGroup: datagroup.DataGroup{GroupID: 101}},
	}
	thresholds := response.SystemHealthConfigResp{
		SevereAbnormalScoreThreshold:  60,
		GeneralAbnormalScoreThreshold: 80,
	}

	base := systemHealthCurrentCacheKey(7, groups, start, end, "15m", thresholds)
	if base == "" {
		t.Fatal("expected non-empty cache key")
	}
	if reordered := systemHealthCurrentCacheKey(7, reversedGroups, start, end, "15m", thresholds); reordered != base {
		t.Fatalf("expected group order not to change permission signature, base=%q reordered=%q", base, reordered)
	}
	if otherUser := systemHealthCurrentCacheKey(8, groups, start, end, "15m", thresholds); otherUser == base {
		t.Fatal("expected user id to partition cache keys")
	}
	if otherGroups := systemHealthCurrentCacheKey(7, groups[:1], start, end, "15m", thresholds); otherGroups == base {
		t.Fatal("expected visible group ids to partition cache keys")
	}
	if otherWindow := systemHealthCurrentCacheKey(7, groups, start.Add(time.Minute), end, "15m", thresholds); otherWindow == base {
		t.Fatal("expected time window to partition cache keys")
	}
	if otherWindowLabel := systemHealthCurrentCacheKey(7, groups, start, end, "30m", thresholds); otherWindowLabel == base {
		t.Fatal("expected window label to partition cache keys")
	}
	otherThresholds := response.SystemHealthConfigResp{
		SevereAbnormalScoreThreshold:  50,
		GeneralAbnormalScoreThreshold: 80,
	}
	if otherConfig := systemHealthCurrentCacheKey(7, groups, start, end, "15m", otherThresholds); otherConfig == base {
		t.Fatal("expected score thresholds to partition cache keys")
	}
}

func TestSystemHealthCurrentCacheKeyBucketsRelativeWindows(t *testing.T) {
	firstEnd := time.Date(2026, 7, 2, 10, 0, 5, 0, time.UTC)
	secondEnd := firstEnd.Add(20 * time.Second)
	nextBucketEnd := firstEnd.Add(31 * time.Second)
	groups := []*datagroup.DataGroupTreeNode{{DataGroup: datagroup.DataGroup{GroupID: 101}}}
	thresholds := response.SystemHealthConfigResp{
		SevereAbnormalScoreThreshold:  60,
		GeneralAbnormalScoreThreshold: 80,
	}

	first := systemHealthCurrentCacheKey(7, groups, firstEnd.Add(-15*time.Minute), firstEnd, "15m", thresholds)
	second := systemHealthCurrentCacheKey(7, groups, secondEnd.Add(-15*time.Minute), secondEnd, "15m", thresholds)
	if first != second {
		t.Fatalf("expected relative current keys within TTL bucket to match, first=%q second=%q", first, second)
	}

	nextBucket := systemHealthCurrentCacheKey(7, groups, nextBucketEnd.Add(-15*time.Minute), nextBucketEnd, "15m", thresholds)
	if nextBucket == first {
		t.Fatal("expected relative current key in next TTL bucket to differ")
	}

	customFirst := systemHealthCurrentCacheKey(7, groups, firstEnd.Add(-15*time.Minute), firstEnd, "custom", thresholds)
	customSecond := systemHealthCurrentCacheKey(7, groups, secondEnd.Add(-15*time.Minute), secondEnd, "custom", thresholds)
	if customFirst == customSecond {
		t.Fatal("expected custom current keys to keep exact timestamps")
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
		SystemResourceShift: []prometheus.MetricResult{
			{
				Values: []prometheus.Points{{Value: 1.30}},
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
		RequestCount: []prometheus.MetricResult{
			{
				Metric: prometheus.Labels{SvcName: "order-api"},
				Values: []prometheus.Points{{Value: 10}},
			},
			{
				Metric: prometheus.Labels{SvcName: "order-worker"},
				Values: []prometheus.Points{{Value: 1190}},
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
	if system.Metrics["resourceShift"] != "yellow" || system.Metrics["lockCost"] != "green" {
		t.Fatalf("expected system to use aggregate extra metrics instead of worst service, got %+v", system.Metrics)
	}
	if system.MetricDetails["resourceShift"].Value != "30.00%" {
		t.Fatalf("expected system resourceShift value to use system aggregate metric value, got %+v", system.MetricDetails["resourceShift"])
	}
	if system.MetricDetails["lockCost"].Value != "0.60 ms/请求" {
		t.Fatalf("expected system lockCost value to use request-weighted aggregate, got %+v", system.MetricDetails["lockCost"])
	}
	if system.MetricDetails["resourceShift"].Note != "系统资源消耗相对 24h 前基线波动" {
		t.Fatalf("expected resourceShift note to describe resource baseline, got %q", system.MetricDetails["resourceShift"].Note)
	}
	if system.MetricDetails["lockCost"].Note != "系统 futex 锁等待耗时按请求数均摊" {
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
	if system.Metrics[healthMetricResource] != healthStatusRed || systemResource.Value != "93.00%" || systemResource.Source != "" {
		t.Fatalf("expected system resource to use aggregate VM host waterline, metrics=%+v detail=%+v", system.Metrics, systemResource)
	}
	if got := systemResource.Badges; len(got) < 3 || got[0].Label != "平均水位" || got[0].Value != "82.50%" || got[1].Label != "高水位实例" || got[1].Value != "2/2" || got[2].Label != "P95水位" || got[2].Value != "93.00%" {
		t.Fatalf("expected system resource badges to expose p95 and affected host ratio, got %+v", got)
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

func TestApplySystemHealthExtraMetricsReplacesReplicaResourceWithWaterlineDetail(t *testing.T) {
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
	if system.Services[0].Metrics[healthMetricResource] != healthStatusYellow || resource.Value != "72.00%" || resource.Source != "host/vm-a" {
		t.Fatalf("expected resource waterline to replace replica detail, metrics=%+v detail=%+v", system.Services[0].Metrics, resource)
	}
}

func TestApplySystemHealthExtraMetricsIgnoresReplicaWhenBuildingSystemResourceDetail(t *testing.T) {
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
	if system.Metrics[healthMetricResource] != healthStatusYellow || resource.Value != "72.00%" || resource.Source != "" || resource.Note != "系统资源水位聚合" {
		t.Fatalf("expected system resource to use aggregate waterline status and detail, metrics=%+v detail=%+v", system.Metrics, resource)
	}
}

func TestApplySystemHealthExtraMetricsDoesNotKeepReplicaRedWhenWaterlineIsGreen(t *testing.T) {
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
	if system.Metrics[healthMetricResource] != healthStatusGreen || resource.Value != "69.50%" || resource.Source != "" || resource.Note != "系统资源水位聚合" {
		t.Fatalf("expected aggregate green waterline to replace replica red status, metrics=%+v detail=%+v", system.Metrics, resource)
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

func TestSystemMetricDetailUsesMetricSpecificContributorBadge(t *testing.T) {
	services := []response.SystemHealthService{
		{
			Name:    "order-api",
			Metrics: map[string]string{healthMetricErrorRate: healthStatusRed, healthMetricResourceShift: healthStatusYellow},
			MetricDetails: map[string]response.SystemHealthMetricDetail{
				healthMetricErrorRate:     {Status: healthStatusRed, Value: "12.30%"},
				healthMetricResourceShift: {Status: healthStatusYellow, Value: "28.00%"},
			},
		},
	}

	errorRate := systemMetricDetail(services, healthMetricErrorRate, healthStatusRed)
	if errorRate.Value != "--" {
		t.Fatalf("expected system metric detail to leave main value for system aggregate, got %q", errorRate.Value)
	}
	if len(errorRate.Badges) != 1 || errorRate.Badges[0].Label != "错误率最高服务" || errorRate.Badges[0].Value != "order-api" {
		t.Fatalf("expected errorRate contributor badge to name the highest error-rate service, got %+v", errorRate.Badges)
	}

	resourceShift := systemMetricDetail(services, healthMetricResourceShift, healthStatusYellow)
	if resourceShift.Value != "--" {
		t.Fatalf("expected system metric detail to leave resourceShift main value for system aggregate, got %q", resourceShift.Value)
	}
	if len(resourceShift.Badges) != 1 || resourceShift.Badges[0].Label != "主要贡献服务" || resourceShift.Badges[0].Value != "order-api" {
		t.Fatalf("expected resourceShift contributor badge to name the largest-shift service, got %+v", resourceShift.Badges)
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
		SystemResourceCPU: []prometheus.MetricResult{
			{Metric: prometheus.Labels{POD: "order-api"}, Values: []prometheus.Points{{Value: 0.63}}},
			{Metric: prometheus.Labels{POD: "order-worker"}, Values: []prometheus.Points{{Value: 0.42}}},
		},
		SystemResourceMemory: []prometheus.MetricResult{
			{Metric: prometheus.Labels{POD: "order-api"}, Values: []prometheus.Points{{Value: 0.91}}},
			{Metric: prometheus.Labels{POD: "order-worker"}, Values: []prometheus.Points{{Value: 0.56}}},
		},
		Latency: []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "order-api"}, Values: []prometheus.Points{{Value: 900_000_000}}},
			{Metric: prometheus.Labels{SvcName: "order-worker"}, Values: []prometheus.Points{{Value: 120_000_000}}},
		},
		SystemLatency: []prometheus.MetricResult{
			{Values: []prometheus.Points{{Value: 480_000_000}}},
		},
		ErrorRate: []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "order-api"}, Values: []prometheus.Points{{Value: 0.123}}},
			{Metric: prometheus.Labels{SvcName: "order-worker"}, Values: []prometheus.Points{{Value: 0}}},
		},
		ErrorLog: []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "order-api"}, Values: []prometheus.Points{{Value: 42}}},
			{Metric: prometheus.Labels{SvcName: "order-worker"}, Values: []prometheus.Points{{Value: 0}}},
		},
		BadRequests: map[string]systemHealthAlertSummary{
			"order-api": {
				Count:            3,
				RedCount:         1,
				YellowCount:      2,
				HighestSeverity:  model.SeverityLevelCritical,
				AffectedServices: 1,
			},
			"order-worker": {},
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
		t.Fatalf("expected system resource to use aggregate status-driving waterline, got %+v", system.MetricDetails["resource"])
	}
	if system.MetricDetails["latency"].Value != "480ms" {
		t.Fatalf("expected system latency to use aggregate current value, got %+v", system.MetricDetails["latency"])
	}
	if system.MetricDetails["badRequests"].Value != "3" {
		t.Fatalf("expected system badRequests to inherit active alert count, got %+v", system.MetricDetails["badRequests"])
	}
	if system.MetricDetails["badRequests"].Status != healthStatusRed {
		t.Fatalf("expected system badRequests to use alert severity, got %+v", system.MetricDetails["badRequests"])
	}
	if got := system.MetricDetails["badRequests"].Badges; len(got) < 3 || got[0].Label != "红色告警" || got[0].Value != "1" || got[1].Label != "黄色告警" || got[1].Value != "2" || got[2].Label != "影响服务" || got[2].Value != "1" {
		t.Fatalf("expected system badRequests badges to expose severity distribution, got %+v", got)
	}
	if !systemHealthMetricBadgesContain(system.MetricDetails["badRequests"].Badges, "活跃告警最多服务", "order-api") {
		t.Fatalf("expected system badRequests badges to keep top alert service contributor, got %+v", system.MetricDetails["badRequests"].Badges)
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
	}, nil)

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
		SystemRequestCount: []prometheus.MetricResult{
			{Values: []prometheus.Points{{Value: 1000}}},
		},
		SystemErrorCount: []prometheus.MetricResult{
			{Values: []prometheus.Points{{Value: 57}}},
		},
		ErrorRate: []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "order-api"}, Values: []prometheus.Points{{Value: 0.123}}},
			{Metric: prometheus.Labels{SvcName: "order-worker"}, Values: []prometheus.Points{{Value: 0.017}}},
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
		t.Fatalf("expected system error rate current value to use system aggregate, got %+v", system.MetricDetails[healthMetricErrorRate])
	}
	if got := system.MetricDetails[healthMetricErrorRate].Series; len(got) != 2 || got[0].Value != 1.2 || got[1].Value != 3.4 {
		t.Fatalf("expected error rate trend to use system aggregate percent, got %+v", got)
	}
	if system.MetricDetails[healthMetricErrorRate].SeriesUnit != "%" || system.MetricDetails[healthMetricErrorRate].SeriesLabel != "系统错误率" {
		t.Fatalf("expected error rate chart metadata to describe system aggregate, got %+v", system.MetricDetails[healthMetricErrorRate])
	}
	if system.MetricDetails[healthMetricErrorLog].Value != "7" {
		t.Fatalf("expected system error log current value to sum service logs, got %+v", system.MetricDetails[healthMetricErrorLog])
	}
	if got := system.MetricDetails[healthMetricErrorLog].Series; len(got) != 2 || got[0].Value != 3 || got[1].Value != 9 {
		t.Fatalf("expected error log trend to sum services per timestamp, got %+v", got)
	}
}

func TestApplySystemHealthExtraMetricsMarksHundredPercentCurrentErrorRateRed(t *testing.T) {
	system := response.SystemHealthSystem{
		Metrics:       defaultSystemMetricStatus(),
		MetricDetails: defaultSystemMetricDetails(),
		Services: []response.SystemHealthService{
			{
				Name:          "order-api",
				Metrics:       map[string]string{healthMetricErrorRate: healthStatusGreen},
				MetricDetails: defaultSystemMetricDetails(),
			},
		},
	}
	extra := systemHealthExtraMetricResults{
		ErrorRate: []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "order-api"}, Values: []prometheus.Points{{Value: 1}}},
		},
		SystemErrorRate: []prometheus.MetricResult{
			{Values: []prometheus.Points{{Value: 1}}},
		},
	}

	applySystemHealthExtraMetrics(&system, extra)

	service := system.Services[0]
	if service.Metrics[healthMetricErrorRate] != healthStatusRed || service.MetricDetails[healthMetricErrorRate].Status != healthStatusRed {
		t.Fatalf("expected service current 100%% error rate to be red, metrics=%+v detail=%+v", service.Metrics, service.MetricDetails[healthMetricErrorRate])
	}
	if system.Metrics[healthMetricErrorRate] != healthStatusRed || system.MetricDetails[healthMetricErrorRate].Status != healthStatusRed {
		t.Fatalf("expected system current 100%% error rate to be red, metrics=%+v detail=%+v", system.Metrics, system.MetricDetails[healthMetricErrorRate])
	}
	if system.MetricDetails[healthMetricErrorRate].Value != "100.00%" {
		t.Fatalf("expected system error rate value to stay visible, got %+v", system.MetricDetails[healthMetricErrorRate])
	}
}

func TestApplySystemHealthExtraMetricsCapsLowSampleErrorRateAtYellow(t *testing.T) {
	system := response.SystemHealthSystem{
		Metrics:       defaultSystemMetricStatus(),
		MetricDetails: defaultSystemMetricDetails(),
		Services: []response.SystemHealthService{
			{
				Name:          "order-api",
				Metrics:       map[string]string{healthMetricErrorRate: healthStatusGreen},
				MetricDetails: defaultSystemMetricDetails(),
			},
		},
	}
	extra := systemHealthExtraMetricResults{
		ErrorRate: []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "order-api"}, Values: []prometheus.Points{{Value: 1}}},
		},
		RequestCount: []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "order-api"}, Values: []prometheus.Points{{Value: 1}}},
		},
		ErrorCount: []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "order-api"}, Values: []prometheus.Points{{Value: 1}}},
		},
		SystemErrorRate: []prometheus.MetricResult{
			{Values: []prometheus.Points{{Value: 1}}},
		},
		SystemRequestCount: []prometheus.MetricResult{
			{Values: []prometheus.Points{{Value: 1}}},
		},
		SystemErrorCount: []prometheus.MetricResult{
			{Values: []prometheus.Points{{Value: 1}}},
		},
	}

	applySystemHealthExtraMetrics(&system, extra)

	service := system.Services[0]
	if service.Metrics[healthMetricErrorRate] != healthStatusYellow || service.MetricDetails[healthMetricErrorRate].Status != healthStatusYellow {
		t.Fatalf("expected low-sample 100%% service error rate to be yellow, metrics=%+v detail=%+v", service.Metrics, service.MetricDetails[healthMetricErrorRate])
	}
	if system.Metrics[healthMetricErrorRate] != healthStatusYellow || system.MetricDetails[healthMetricErrorRate].Status != healthStatusYellow {
		t.Fatalf("expected low-sample 100%% system error rate to be yellow, metrics=%+v detail=%+v", system.Metrics, system.MetricDetails[healthMetricErrorRate])
	}
	if got := service.MetricDetails[healthMetricErrorRate].Badges; len(got) != 3 || got[0].Label != "请求数" || got[0].Value != "1" || got[1].Label != "失败数" || got[1].Value != "1" || got[2].Label != "样本不足" {
		t.Fatalf("expected low-sample error rate badges to expose request and error counts, got %+v", got)
	}
	if got := system.MetricDetails[healthMetricErrorRate].Badges; len(got) < 3 || got[0].Label != "请求数" || got[0].Value != "1" || got[1].Label != "失败数" || got[1].Value != "1" || got[2].Label != "样本不足" {
		t.Fatalf("expected system low-sample error rate badges to expose aggregate request and error counts, got %+v", got)
	}
	if !systemHealthMetricBadgesContain(system.MetricDetails[healthMetricErrorRate].Badges, "错误率最高服务", "order-api") {
		t.Fatalf("expected system low-sample error rate badges to keep highest error-rate service contributor, got %+v", system.MetricDetails[healthMetricErrorRate].Badges)
	}
}

func TestApplySystemHealthExtraMetricsKeepsHighSampleErrorRateRed(t *testing.T) {
	system := response.SystemHealthSystem{
		Metrics:       defaultSystemMetricStatus(),
		MetricDetails: defaultSystemMetricDetails(),
		Services: []response.SystemHealthService{
			{
				Name:          "order-api",
				Metrics:       map[string]string{healthMetricErrorRate: healthStatusGreen},
				MetricDetails: defaultSystemMetricDetails(),
			},
		},
	}
	extra := systemHealthExtraMetricResults{
		ErrorRate: []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "order-api"}, Values: []prometheus.Points{{Value: 0.10}}},
		},
		RequestCount: []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "order-api"}, Values: []prometheus.Points{{Value: 200}}},
		},
		ErrorCount: []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "order-api"}, Values: []prometheus.Points{{Value: 20}}},
		},
		SystemErrorRate: []prometheus.MetricResult{
			{Values: []prometheus.Points{{Value: 0.10}}},
		},
		SystemRequestCount: []prometheus.MetricResult{
			{Values: []prometheus.Points{{Value: 200}}},
		},
		SystemErrorCount: []prometheus.MetricResult{
			{Values: []prometheus.Points{{Value: 20}}},
		},
	}

	applySystemHealthExtraMetrics(&system, extra)

	if system.Services[0].Metrics[healthMetricErrorRate] != healthStatusRed {
		t.Fatalf("expected high-sample service error rate to remain red, got %+v", system.Services[0].Metrics)
	}
	if system.Metrics[healthMetricErrorRate] != healthStatusRed {
		t.Fatalf("expected high-sample system error rate to remain red, got %+v", system.Metrics)
	}
}

func TestApplySystemHealthExtraMetricsUsesAggregateErrorRateForSystemStatus(t *testing.T) {
	system := response.SystemHealthSystem{
		Metrics:       defaultSystemMetricStatus(),
		MetricDetails: defaultSystemMetricDetails(),
		Services: []response.SystemHealthService{
			{
				Name:          "small-api",
				Metrics:       map[string]string{healthMetricErrorRate: healthStatusGreen},
				MetricDetails: defaultSystemMetricDetails(),
			},
			{
				Name:          "bulk-api",
				Metrics:       map[string]string{healthMetricErrorRate: healthStatusGreen},
				MetricDetails: defaultSystemMetricDetails(),
			},
		},
	}
	extra := systemHealthExtraMetricResults{
		ErrorRate: []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "small-api"}, Values: []prometheus.Points{{Value: 1}}},
			{Metric: prometheus.Labels{SvcName: "bulk-api"}, Values: []prometheus.Points{{Value: 0}}},
		},
		RequestCount: []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "small-api"}, Values: []prometheus.Points{{Value: 1}}},
			{Metric: prometheus.Labels{SvcName: "bulk-api"}, Values: []prometheus.Points{{Value: 999}}},
		},
		ErrorCount: []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "small-api"}, Values: []prometheus.Points{{Value: 1}}},
			{Metric: prometheus.Labels{SvcName: "bulk-api"}, Values: []prometheus.Points{{Value: 0}}},
		},
		SystemErrorRate: []prometheus.MetricResult{
			{Values: []prometheus.Points{{Value: 0.001}}},
		},
		SystemRequestCount: []prometheus.MetricResult{
			{Values: []prometheus.Points{{Value: 1000}}},
		},
		SystemErrorCount: []prometheus.MetricResult{
			{Values: []prometheus.Points{{Value: 1}}},
		},
	}

	applySystemHealthExtraMetrics(&system, extra)

	smallAPI := serviceByName(t, system.Services, "small-api")
	if smallAPI.Metrics[healthMetricErrorRate] != healthStatusYellow {
		t.Fatalf("expected low-sample failing service to be yellow, got %+v", smallAPI.Metrics)
	}
	if system.Metrics[healthMetricErrorRate] != healthStatusYellow {
		t.Fatalf("expected system aggregate error rate to be yellow instead of one-vote red, got %+v", system.Metrics)
	}
	if system.MetricDetails[healthMetricErrorRate].Value != "0.10%" {
		t.Fatalf("expected system error rate value to use total failures / total requests, got %+v", system.MetricDetails[healthMetricErrorRate])
	}
}

func TestApplySystemHealthExtraMetricsUsesAggregateLatencyStatus(t *testing.T) {
	system := response.SystemHealthSystem{
		Metrics:       defaultSystemMetricStatus(),
		MetricDetails: defaultSystemMetricDetails(),
		Services: []response.SystemHealthService{
			{
				Name:          "small-api",
				Metrics:       map[string]string{healthMetricLatency: healthStatusRed},
				MetricDetails: map[string]response.SystemHealthMetricDetail{healthMetricLatency: {Status: healthStatusRed, Value: "1200ms"}},
			},
			{
				Name:          "bulk-api",
				Metrics:       map[string]string{healthMetricLatency: healthStatusGreen},
				MetricDetails: map[string]response.SystemHealthMetricDetail{healthMetricLatency: {Status: healthStatusGreen, Value: "80ms"}},
			},
		},
	}
	extra := systemHealthExtraMetricResults{
		SystemLatency: []prometheus.MetricResult{
			{Values: []prometheus.Points{{Value: 200_000_000}}},
		},
		SystemLatencyDoD: []prometheus.MetricResult{
			{Values: []prometheus.Points{{Value: 1.03}}},
		},
		SystemRequestCount: []prometheus.MetricResult{
			{Values: []prometheus.Points{{Value: 1000}}},
		},
	}

	applySystemHealthExtraMetrics(&system, extra)

	if system.Metrics[healthMetricLatency] != healthStatusGreen {
		t.Fatalf("expected system latency to use aggregate status instead of worst service, got metrics=%+v detail=%+v", system.Metrics, system.MetricDetails[healthMetricLatency])
	}
	detail := system.MetricDetails[healthMetricLatency]
	if detail.Value != "200ms" || detail.Status != healthStatusGreen {
		t.Fatalf("expected system latency aggregate value/status, got %+v", detail)
	}
	if !systemHealthMetricBadgesContain(detail.Badges, "响应最慢服务", "small-api") {
		t.Fatalf("expected latency tooltip badges to keep slowest service contributor, got %+v", detail.Badges)
	}
	if !systemHealthMetricBadgesContain(detail.Badges, "24h变化", "3.00%") {
		t.Fatalf("expected latency tooltip badges to expose aggregate 24h change, got %+v", detail.Badges)
	}
}

func TestApplySystemHealthExtraMetricsUsesAggregateErrorLogStatus(t *testing.T) {
	system := response.SystemHealthSystem{
		Metrics:       defaultSystemMetricStatus(),
		MetricDetails: defaultSystemMetricDetails(),
		Services: []response.SystemHealthService{
			{
				Name:          "small-api",
				Metrics:       map[string]string{healthMetricErrorLog: healthStatusRed},
				MetricDetails: map[string]response.SystemHealthMetricDetail{healthMetricErrorLog: {Status: healthStatusRed, Value: "2"}},
			},
			{
				Name:          "bulk-api",
				Metrics:       map[string]string{healthMetricErrorLog: healthStatusGreen},
				MetricDetails: map[string]response.SystemHealthMetricDetail{healthMetricErrorLog: {Status: healthStatusGreen, Value: "0"}},
			},
		},
	}
	extra := systemHealthExtraMetricResults{
		ErrorLog: []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "small-api"}, Values: []prometheus.Points{{Value: 2}}},
			{Metric: prometheus.Labels{SvcName: "bulk-api"}, Values: []prometheus.Points{{Value: 0}}},
		},
		SystemErrorLogDoD: []prometheus.MetricResult{
			{Values: []prometheus.Points{{Value: 1.02}}},
		},
	}

	applySystemHealthExtraMetrics(&system, extra)

	if system.Metrics[healthMetricErrorLog] != healthStatusGreen {
		t.Fatalf("expected system errorLog to use aggregate 24h change instead of worst service, got metrics=%+v detail=%+v", system.Metrics, system.MetricDetails[healthMetricErrorLog])
	}
	detail := system.MetricDetails[healthMetricErrorLog]
	if detail.Value != "2" || detail.Status != healthStatusGreen {
		t.Fatalf("expected system errorLog aggregate value/status, got %+v", detail)
	}
	if !systemHealthMetricBadgesContain(detail.Badges, "错误日志最多服务", "small-api") {
		t.Fatalf("expected errorLog tooltip badges to keep top log contributor, got %+v", detail.Badges)
	}
	if !systemHealthMetricBadgesContain(detail.Badges, "24h变化", "2.00%") {
		t.Fatalf("expected errorLog tooltip badges to expose aggregate 24h change, got %+v", detail.Badges)
	}
}

func TestApplySystemHealthExtraMetricsScoresSystemFromAggregateMetricStatuses(t *testing.T) {
	serviceMetrics := defaultSystemMetricStatus()
	for _, key := range systemHealthMetricKeys {
		serviceMetrics[key] = healthStatusGreen
	}
	system := response.SystemHealthSystem{
		Metrics:       defaultSystemMetricStatus(),
		MetricDetails: defaultSystemMetricDetails(),
		Services: []response.SystemHealthService{
			{
				Name:          "healthy-api",
				Metrics:       serviceMetrics,
				MetricDetails: defaultSystemMetricDetails(),
			},
		},
	}
	extra := systemHealthExtraMetricResults{
		SystemErrorRate: []prometheus.MetricResult{
			{Values: []prometheus.Points{{Value: 0.10}}},
		},
		SystemLatency: []prometheus.MetricResult{
			{Values: []prometheus.Points{{Value: 700_000_000}}},
		},
		SystemRequestCount: []prometheus.MetricResult{
			{Values: []prometheus.Points{{Value: 200}}},
		},
		SystemErrorCount: []prometheus.MetricResult{
			{Values: []prometheus.Points{{Value: 20}}},
		},
	}

	applySystemHealthExtraMetrics(&system, extra)

	if system.Metrics[healthMetricErrorRate] != healthStatusRed || system.Metrics[healthMetricLatency] != healthStatusYellow {
		t.Fatalf("expected aggregate metric statuses to drive system metrics, got %+v", system.Metrics)
	}
	if system.Score != 33 || system.Status != healthStatusRed {
		t.Fatalf("expected system score/status from aggregate metrics only, got score=%d status=%s metrics=%+v", system.Score, system.Status, system.Metrics)
	}
}

func TestApplySystemHealthExtraMetricsCapsLowWaterlineResourceShiftAtYellow(t *testing.T) {
	system := response.SystemHealthSystem{
		Metrics:       defaultSystemMetricStatus(),
		MetricDetails: defaultSystemMetricDetails(),
		Services: []response.SystemHealthService{
			{
				Name:          "order-api",
				Metrics:       defaultSystemMetricStatus(),
				MetricDetails: defaultSystemMetricDetails(),
			},
		},
	}
	extra := systemHealthExtraMetricResults{
		ResourceCPU: []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "order-api"}, Values: []prometheus.Points{{Value: 0.05}}},
		},
		ResourceShift: []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "order-api"}, Values: []prometheus.Points{{Value: 2}}},
		},
		SystemResourceCPU: []prometheus.MetricResult{
			{Values: []prometheus.Points{{Value: 0.05}}},
		},
		SystemResourceShift: []prometheus.MetricResult{
			{Values: []prometheus.Points{{Value: 2}}},
		},
	}

	applySystemHealthExtraMetrics(&system, extra)

	service := system.Services[0]
	if service.Metrics[healthMetricResourceShift] != healthStatusYellow || service.MetricDetails[healthMetricResourceShift].Status != healthStatusYellow {
		t.Fatalf("expected low-waterline resource shift to be capped at yellow, metrics=%+v detail=%+v", service.Metrics, service.MetricDetails[healthMetricResourceShift])
	}
	if system.Metrics[healthMetricResourceShift] != healthStatusYellow {
		t.Fatalf("expected low-waterline system resource shift to be yellow, got %+v", system.Metrics)
	}
	if got := service.MetricDetails[healthMetricResourceShift].Badges; len(got) == 0 || got[0].Label != "当前水位" || got[0].Value != "5.00%" {
		t.Fatalf("expected resource shift badge to expose current waterline, got %+v", got)
	}
}

func TestApplySystemHealthExtraMetricsCapsLowSampleLockCostAtYellow(t *testing.T) {
	system := response.SystemHealthSystem{
		Metrics:       defaultSystemMetricStatus(),
		MetricDetails: defaultSystemMetricDetails(),
		Services: []response.SystemHealthService{
			{
				Name:          "order-api",
				Metrics:       defaultSystemMetricStatus(),
				MetricDetails: defaultSystemMetricDetails(),
			},
		},
	}
	extra := systemHealthExtraMetricResults{
		LockCost: []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "order-api"}, Values: []prometheus.Points{{Value: 30_000_000}}},
		},
		RequestCount: []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "order-api"}, Values: []prometheus.Points{{Value: 1}}},
		},
	}

	applySystemHealthExtraMetrics(&system, extra)

	service := system.Services[0]
	if service.Metrics[healthMetricLockCost] != healthStatusYellow || service.MetricDetails[healthMetricLockCost].Status != healthStatusYellow {
		t.Fatalf("expected low-sample lock cost to be capped at yellow, metrics=%+v detail=%+v", service.Metrics, service.MetricDetails[healthMetricLockCost])
	}
	if system.Metrics[healthMetricLockCost] != healthStatusYellow {
		t.Fatalf("expected low-sample system lock cost to be yellow, got %+v", system.Metrics)
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

func TestApplySystemHealthExtraMetricsUsesAggregateSystemPodWaterline(t *testing.T) {
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
		t.Fatalf("expected displayed Pod waterline to match the status-driving value, got %+v", detail)
	}
	if detail.Source != "" {
		t.Fatalf("expected system waterline main detail not to point at a single Pod source, got %+v", detail)
	}
	if got := detail.Badges; len(got) < 4 || got[0].Label != "平均水位" || got[0].Value != "82.50%" || got[1].Label != "高水位实例" || got[1].Value != "2/2" || got[2].Label != "P95水位" || got[2].Value != "93.00%" || got[3].Label != "异常实例" || got[3].Value != "1/2" {
		t.Fatalf("expected aggregate waterline badges to expose p95 and affected ratios, got %+v", got)
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
