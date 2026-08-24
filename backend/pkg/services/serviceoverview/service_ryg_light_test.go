// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

import (
	"strings"
	"testing"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

type rygLightPromRepo struct {
	prometheus.Repo
}

func (r *rygLightPromRepo) QueryMetricsWithPQLFilter(ctx core.Context, pqlTpl prometheus.PQLTemplate, startTime int64, endTime int64, gran prometheus.Granularity, filter prometheus.PQLFilter) ([]prometheus.MetricResult, error) {
	query := pqlTpl("1m", string(gran), filter, "")
	if strings.Contains(query, "kindling_span_trace_duration_nanoseconds_sum") && !strings.Contains(query, "offset 24h") {
		return []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "order-api"}, Values: []prometheus.Points{{Value: 100_000_000}}},
		}, nil
	}
	if strings.Contains(query, `is_error="true"`) && !strings.Contains(query, "offset 24h") {
		return []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "order-api"}, Values: []prometheus.Points{{Value: 1}}},
		}, nil
	}
	return nil, nil
}

func (r *rygLightPromRepo) GetInstanceListByPQLFilter(ctx core.Context, startTime int64, endTime int64, filter prometheus.PQLFilter) (*model.ServiceInstances, error) {
	instances := model.NewServiceInstances()
	instances.AddInstances([]*model.ServiceInstance{
		{ServiceName: "order-api", PodName: "order-api-1"},
		{ServiceName: "order-api", PodName: "order-api-2"},
	})
	return instances, nil
}

type rygLightDBRepo struct {
	database.Repo
}

type rygLightClickhouseRepo struct {
	clickhouse.Repo
}

func (r *rygLightClickhouseRepo) GetAlertEventCountGroupByInstance(ctx core.Context, startTime time.Time, endTime time.Time, filter request.AlertFilter, instances *model.RelatedInstances) ([]model.AlertEventCount, error) {
	return nil, nil
}

func TestBuildRYGServiceInstanceFilterKeepsDataGroupScope(t *testing.T) {
	groupFilter := prometheus.NewFilter().
		Equal("cluster_id", "cluster-a").
		Equal("namespace", "orders")
	baseFilter := prometheus.NewFilter().
		RegexMatch(prometheus.ServiceNameKey, "order.*")

	filter := buildRYGServiceInstanceFilter(groupFilter, baseFilter, "order-api")
	got := filter.String()

	for _, want := range []string{
		`cluster_id="cluster-a"`,
		`namespace="orders"`,
		`svc_name=~"order.*"`,
		`svc_name="order-api"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected scoped service instance filter to contain %q, got %s", want, got)
		}
	}
}

func TestExposeRYGLightStatusMarksCurrentErrorsWithoutDayOverDayAsRed(t *testing.T) {
	currentErrorRate := 100.0
	status := RYGLightStatus{
		ErrorRateAvg: &currentErrorRate,
		Instances: []*model.ServiceInstance{
			{},
			{},
		},
	}

	result := status.ExposeRYGLightStatus(code.LANG_ZH)

	for _, detail := range result.ScoreDetail {
		if detail.Key != "errorRate" {
			continue
		}
		if detail.Score != 0 {
			t.Fatalf("expected current errors without day-over-day data to mark errorRate red, got %+v", detail)
		}
		return
	}
	t.Fatalf("expected errorRate score detail, got %+v", result.ScoreDetail)
}

func TestGetServicesRYGLightStatusUsesCurrentErrorRateWhenDayOverDayMissing(t *testing.T) {
	service := &service{
		dbRepo:   &rygLightDBRepo{},
		promRepo: &rygLightPromRepo{},
		chRepo:   &rygLightClickhouseRepo{},
	}

	result, err := service.GetServicesRYGLightStatus(core.EmptyCtx(), &request.GetRygLightRequest{
		StartTime: 1_000_000,
		EndTime:   2_000_000,
	})
	if err != nil {
		t.Fatalf("GetServicesRYGLightStatus returned error: %v", err)
	}
	if len(result.ServiceList) != 1 {
		t.Fatalf("expected one service, got %+v", result.ServiceList)
	}

	for _, detail := range result.ServiceList[0].ScoreDetail {
		if detail.Key != "errorRate" {
			continue
		}
		if detail.Score != 0 {
			t.Fatalf("expected 100%% current error rate without day-over-day data to mark errorRate red, got %+v", detail)
		}
		return
	}
	t.Fatalf("expected errorRate score detail, got %+v", result.ServiceList[0].ScoreDetail)
}
