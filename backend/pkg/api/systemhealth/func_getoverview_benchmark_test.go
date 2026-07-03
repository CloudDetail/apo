// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package systemhealth

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

type systemHealthBenchmarkScale struct {
	name           string
	systemCount    int
	servicesPerSys int
}

var systemHealthBenchmarkScales = []systemHealthBenchmarkScale{
	{name: "10_systems_20_services_each", systemCount: 10, servicesPerSys: 20},
	{name: "50_systems_100_services_each", systemCount: 50, servicesPerSys: 100},
	{name: "200_systems_300_services_each", systemCount: 200, servicesPerSys: 300},
}

func BenchmarkSystemHealthCurrentCacheHitCloneByCardinality(b *testing.B) {
	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)
	for _, scale := range systemHealthBenchmarkScales {
		b.Run(scale.name, func(b *testing.B) {
			systems := benchmarkSystemHealthSystems(scale.systemCount, scale.servicesPerSys)
			cache := newSystemHealthCurrentCache(time.Minute, 8)
			_, _, err := cache.getOrBuildWithTelemetry("benchmark-key", now, func() (systemHealthCurrentResult, error) {
				return benchmarkSystemHealthCurrentResult(systems), nil
			})
			if err != nil {
				b.Fatalf("prime cache: %v", err)
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result, telemetry, err := cache.getOrBuildWithTelemetry("benchmark-key", now.Add(time.Second), func() (systemHealthCurrentResult, error) {
					b.Fatal("expected cache hit")
					return systemHealthCurrentResult{}, nil
				})
				if err != nil {
					b.Fatalf("cache hit: %v", err)
				}
				if telemetry.Status != systemHealthCurrentCacheStatusHit {
					b.Fatalf("expected cache hit, got %+v", telemetry)
				}
				if len(result.Systems) != scale.systemCount {
					b.Fatalf("expected %d systems, got %d", scale.systemCount, len(result.Systems))
				}
			}
		})
	}
}

func BenchmarkSystemHealthCurrentCacheReadOnlyHitByCardinality(b *testing.B) {
	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)
	for _, scale := range systemHealthBenchmarkScales {
		b.Run(scale.name, func(b *testing.B) {
			systems := benchmarkSystemHealthSystems(scale.systemCount, scale.servicesPerSys)
			cache := newSystemHealthCurrentCache(time.Minute, 8)
			_, _, err := cache.getOrBuildWithTelemetry("benchmark-key", now, func() (systemHealthCurrentResult, error) {
				return benchmarkSystemHealthCurrentResult(systems), nil
			})
			if err != nil {
				b.Fatalf("prime cache: %v", err)
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result, telemetry, ok := cache.getReadOnlyIfFreshOrWait("benchmark-key", now.Add(time.Second))
				if !ok {
					b.Fatal("expected read-only cache hit")
				}
				if telemetry.Status != systemHealthCurrentCacheStatusHit {
					b.Fatalf("expected cache hit, got %+v", telemetry)
				}
				if len(result.Systems) != scale.systemCount {
					b.Fatalf("expected %d systems, got %d", scale.systemCount, len(result.Systems))
				}
			}
		})
	}
}

func BenchmarkSystemHealthSystemsPageByCardinality(b *testing.B) {
	thresholds := benchmarkSystemHealthConfig()
	for _, scale := range systemHealthBenchmarkScales {
		b.Run(scale.name, func(b *testing.B) {
			systems := benchmarkSystemHealthSystems(scale.systemCount, scale.servicesPerSys)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				resp := buildSystemHealthSystemsPage(systems, systemHealthListOptions{
					Page:     1,
					PageSize: 50,
					Status:   "all",
					Sort:     "score_asc",
				}, thresholds, "15m", "2026-07-02T10:00:00Z")
				if resp.Total != scale.systemCount {
					b.Fatalf("expected %d systems, got %d", scale.systemCount, resp.Total)
				}
			}
		})
	}
}

func BenchmarkSystemHealthPendingIssuesBuildByCardinality(b *testing.B) {
	for _, scale := range systemHealthBenchmarkScales {
		b.Run(scale.name, func(b *testing.B) {
			systems := benchmarkSystemHealthSystems(scale.systemCount, scale.servicesPerSys)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				issues := buildSystemHealthPendingIssues(systems, 0)
				if len(issues) == 0 {
					b.Fatal("expected pending issues")
				}
			}
		})
	}
}

func BenchmarkSystemHealthPendingIssuesPageFromCachedIssuesByCardinality(b *testing.B) {
	for _, scale := range systemHealthBenchmarkScales {
		b.Run(scale.name, func(b *testing.B) {
			issues := buildSystemHealthPendingIssues(benchmarkSystemHealthSystems(scale.systemCount, scale.servicesPerSys), 0)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				resp := buildSystemHealthPendingIssuesPageFromIssues(issues, systemHealthPendingIssueListOptions{
					Page:     1,
					PageSize: 20,
					Status:   "all",
				}, "15m", "2026-07-02T10:00:00Z")
				if resp.Total != len(issues) {
					b.Fatalf("expected %d issues, got %d", len(issues), resp.Total)
				}
			}
		})
	}
}

func BenchmarkSystemHealthJSONMarshalByCardinality(b *testing.B) {
	thresholds := benchmarkSystemHealthConfig()
	for _, scale := range systemHealthBenchmarkScales {
		b.Run(scale.name+"/systems_page", func(b *testing.B) {
			resp := buildSystemHealthSystemsPage(benchmarkSystemHealthSystems(scale.systemCount, scale.servicesPerSys), systemHealthListOptions{
				Page:     1,
				PageSize: 50,
				Status:   "all",
				Sort:     "score_asc",
			}, thresholds, "15m", "2026-07-02T10:00:00Z")

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				payload, err := json.Marshal(resp)
				if err != nil {
					b.Fatalf("marshal systems page: %v", err)
				}
				if len(payload) == 0 {
					b.Fatal("expected systems payload")
				}
			}
		})
		b.Run(scale.name+"/pending_issues_page", func(b *testing.B) {
			issues := buildSystemHealthPendingIssues(benchmarkSystemHealthSystems(scale.systemCount, scale.servicesPerSys), 0)
			resp := buildSystemHealthPendingIssuesPageFromIssues(issues, systemHealthPendingIssueListOptions{
				Page:     1,
				PageSize: 20,
				Status:   "all",
			}, "15m", "2026-07-02T10:00:00Z")

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				payload, err := json.Marshal(resp)
				if err != nil {
					b.Fatalf("marshal pending issues page: %v", err)
				}
				if len(payload) == 0 {
					b.Fatal("expected pending issues payload")
				}
			}
		})
	}
}

func benchmarkSystemHealthCurrentResult(systems []response.SystemHealthSystem) systemHealthCurrentResult {
	thresholds := benchmarkSystemHealthConfig()
	return systemHealthCurrentResult{
		Summary:       buildSystemHealthOverviewSummary(systems, thresholds),
		Systems:       systems,
		PendingIssues: buildSystemHealthPendingIssues(systems, 0),
		Config:        thresholds,
		Window:        "15m",
		GeneratedAt:   "2026-07-02T10:00:00Z",
		Source:        systemHealthSourceDataGroup,
	}
}

func benchmarkSystemHealthSystems(systemCount int, servicesPerSystem int) []response.SystemHealthSystem {
	systems := make([]response.SystemHealthSystem, 0, systemCount)
	for systemIndex := 0; systemIndex < systemCount; systemIndex++ {
		groupID := int64(100000 + systemIndex)
		services := make([]response.SystemHealthService, 0, servicesPerSystem)
		for serviceIndex := 0; serviceIndex < servicesPerSystem; serviceIndex++ {
			services = append(services, benchmarkSystemHealthService(systemIndex, serviceIndex))
		}
		system := response.SystemHealthSystem{
			ID:            groupID,
			GroupID:       groupID,
			Name:          fmt.Sprintf("system-%04d", systemIndex),
			Services:      services,
			Metrics:       defaultSystemMetricStatus(),
			MetricDetails: defaultSystemMetricDetails(),
			Source:        systemHealthSourceDataGroup,
		}
		recalculateSystemHealthSystem(&system)
		applySystemHealthScoreThreshold(&system, benchmarkSystemHealthConfig())
		systems = append(systems, system)
	}
	return systems
}

func benchmarkSystemHealthService(systemIndex int, serviceIndex int) response.SystemHealthService {
	score := 95
	status := healthStatusGreen
	if serviceIndex%10 == 0 {
		score = 45
		status = healthStatusRed
	} else if serviceIndex%5 == 0 {
		score = 72
		status = healthStatusYellow
	}
	metrics := make(map[string]string, len(systemHealthMetricKeys))
	details := make(map[string]response.SystemHealthMetricDetail, len(systemHealthMetricKeys))
	for metricIndex, key := range systemHealthMetricKeys {
		metricStatus := healthStatusGreen
		if status == healthStatusRed && metricIndex%3 == 0 {
			metricStatus = healthStatusRed
		} else if status == healthStatusYellow && metricIndex%4 == 0 {
			metricStatus = healthStatusYellow
		}
		metrics[key] = metricStatus
		details[key] = response.SystemHealthMetricDetail{
			Status: metricStatus,
			Value:  fmt.Sprintf("%d", serviceIndex+metricIndex),
			Note:   fmt.Sprintf("metric %s for service %d-%d", key, systemIndex, serviceIndex),
			Series: []response.SystemHealthMetricPoint{
				{Timestamp: 1782972000000000, Value: float64(serviceIndex + metricIndex)},
				{Timestamp: 1782972060000000, Value: float64(serviceIndex + metricIndex + 1)},
			},
			Thresholds: []response.SystemHealthMetricThreshold{
				{Label: "severe", Value: 65, Status: healthStatusRed},
				{Label: "general", Value: 85, Status: healthStatusYellow},
			},
		}
	}
	return response.SystemHealthService{
		Name:          fmt.Sprintf("service-%04d-%04d", systemIndex, serviceIndex),
		Desc:          fmt.Sprintf("synthetic service %d-%d", systemIndex, serviceIndex),
		Score:         score,
		Status:        status,
		Metrics:       metrics,
		MetricDetails: details,
	}
}

func benchmarkSystemHealthConfig() response.SystemHealthConfigResp {
	return response.SystemHealthConfigResp{
		SevereAbnormalScoreThreshold:  65,
		GeneralAbnormalScoreThreshold: 85,
	}
}
