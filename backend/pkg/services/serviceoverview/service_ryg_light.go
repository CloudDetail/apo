package serviceoverview

import (
	"fmt"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	prom "github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

func (s *service) GetServicesRYGLightStatus(startTime time.Time, endTime time.Time, filter EndpointsFilter) (response.ServiceRYGLightRes, error) {
	var servicesMap = &servicesRYGLightMap{
		MetricGroupList: []*RYGLightStatus{},
		MetricGroupMap:  map[prom.ServiceKey]*RYGLightStatus{},
	}

	startMicroTS := startTime.UnixMicro()
	endMicroTs := endTime.UnixMicro()

	filters := extractEndpointFilters(filter)

	// FIX 用于展示没有LatencyDOD的服务
	avgLatency, err := s.promRepo.QueryAggMetricsWithFilter(
		prom.PQLAvgLatencyWithFilters,
		startMicroTS, endMicroTs,
		prom.SVCGranularity, filters...)
	if err == nil {
		servicesMap.MergeMetricResults(prom.AVG, prom.LATENCY, avgLatency)
	}

	avgLatencyDoD, err := s.promRepo.QueryAggMetricsWithFilter(
		prom.DayOnDay(prom.PQLAvgLatencyWithFilters),
		startMicroTS, endMicroTs,
		prom.SVCGranularity, filters...)
	if err == nil {
		servicesMap.MergeMetricResults(prom.DOD, prom.LATENCY, avgLatencyDoD)
	}

	avgErrorRateDoD, err := s.promRepo.QueryAggMetricsWithFilter(
		prom.DayOnDay(prom.PQLAvgErrorRateWithFilters),
		startMicroTS, endMicroTs,
		prom.SVCGranularity, filters...)
	if err == nil {
		servicesMap.MergeMetricResults(prom.DOD, prom.ERROR_RATE, avgErrorRateDoD)
	}

	avgLogErrorCountDoD, err := s.promRepo.QueryAggMetricsWithFilter(
		prom.DayOnDay(prom.PQLAvgLogErrorCountWithFilters),
		startMicroTS, endMicroTs,
		prom.SVCGranularity, filters...)
	if err == nil {
		servicesMap.MergeMetricResults(prom.DOD, prom.LOG_ERROR_COUNT, avgLogErrorCountDoD)
	}

	var resp = response.ServiceRYGLightRes{
		ServiceList: []*response.ServiceRYGResult{},
	}

	alertEventCount, _ := s.chRepo.GetAlertEventCountGroupByInstance(startTime, endTime, request.AlertFilter{Status: "firing"}, nil)
	for svcKey, status := range servicesMap.MetricGroupMap {
		instances, err := s.promRepo.GetInstanceList(startMicroTS, endMicroTs, svcKey.SvcName, "")
		if err != nil {
			continue
		}

		status.Instances = instances.GetInstances()

		if alertEventCount != nil {
			status.AlertEventLevelCountMap = GroupAlertEventCountListByInstance(alertEventCount, status.Instances)
		}

		resp.ServiceList = append(resp.ServiceList, &response.ServiceRYGResult{
			ServiceName: svcKey.SvcName,
			RYGResult:   *status.ExposeRYGLightStatus(),
		})
	}

	return resp, nil
}

// RYGLightStatus
// Red/Green/Yellow Status
type RYGLightStatus struct {
	// From Prometheus
	LatencyAvg *float64 // 平均延时 用于统计服务

	LatencyDoD       *float64 // 延迟日同比
	ErrorRateDoD     *float64 // 错误率日同比
	LogErrorCountDoD *float64 // 日志错误数日同比

	Instances []*model.ServiceInstance

	// From Clickhouse
	model.AlertStatus             // 告警状态
	model.AlertEventLevelCountMap // 告警事件级别统计
}

func (s *RYGLightStatus) ExposeRYGLightStatus() *response.RYGResult {
	var res = &response.RYGResult{}

	latencyScore := ScoreFromDoD(s.LatencyDoD, 10, 20, 50)
	if latencyScore >= 0 {
		res.Score += latencyScore
		res.ScoreDetail = append(res.ScoreDetail, response.RYGScoreDetail{
			Key:    "latency",
			Score:  latencyScore,
			Detail: fmt.Sprintf("latency 同比增长 %.2f%%", *s.LatencyDoD),
		})
	}

	errorRateScore := ScoreFromDoD(s.ErrorRateDoD, 5, 10, 20)
	if errorRateScore >= 0 {
		res.Score += errorRateScore
		res.ScoreDetail = append(res.ScoreDetail, response.RYGScoreDetail{
			Key:    "errorRate",
			Score:  errorRateScore,
			Detail: fmt.Sprintf("errorRate 同比增长 %.2f%%", *s.ErrorRateDoD),
		})
	}

	logErrorCountScore := ScoreFromDoD(s.LogErrorCountDoD, 5, 10, 20)
	if logErrorCountScore >= 0 {
		res.Score += logErrorCountScore
		res.ScoreDetail = append(res.ScoreDetail, response.RYGScoreDetail{
			Key:    "logErrorCount",
			Score:  logErrorCountScore,
			Detail: fmt.Sprintf("logErrorCount 同比增长 %.2f%%", *s.LogErrorCountDoD),
		})
	}

	alertScore := AlertScore(&s.AlertStatus, &s.AlertEventLevelCountMap)
	if alertScore >= 0 {
		res.Score += alertScore
		res.ScoreDetail = append(res.ScoreDetail, response.RYGScoreDetail{
			Key:    "alert",
			Score:  alertScore,
			Detail: getAlertScoreDetail(alertScore),
		})
	}

	if len(s.Instances) < 2 {
		res.Score += 3
		res.ScoreDetail = append(res.ScoreDetail, response.RYGScoreDetail{
			Key:    "replica",
			Score:  3,
			Detail: "应用实例数小于2, 存在服务不可用风险",
		})
	}

	if res.Score > 10 {
		res.Status = response.RED
	} else if res.Score > 3 {
		res.Status = response.YELLOW
	} else {
		res.Status = response.GREEN
	}
	return res
}

func ScoreFromDoD(value *float64, l3, l2, l1 float64) int {
	if value == nil {
		return -1
	}
	if *value > l1 {
		return 3
	} else if *value > l2 {
		return 2
	} else if *value > l3 {
		return 1
	}
	return 0
}

func AlertScore(status *model.AlertStatus, eventLevelCountMap *model.AlertEventLevelCountMap) int {
	if status == nil || status.IsAllNormal() {
		return 0
	}

	if eventLevelCountMap == nil {
		return 0
	}

	var warningCount int
	for _, eventLevelCounts := range *eventLevelCountMap {
		count, find := eventLevelCounts[model.SeverityLevelCritical]
		if find && count > 0 {
			return 3
		}
		count, find = eventLevelCounts[model.SeverityLevelError]
		if find && count > 0 {
			return 3
		}

		count, find = eventLevelCounts[model.SeverityLevelWarning]
		if find && count > 0 {
			warningCount++
		}
	}
	if warningCount >= 2 {
		return 2
	} else if warningCount >= 1 {
		return 1
	}
	return 0
}

func getAlertScoreDetail(score int) string {
	if score >= 3 {
		return "存在高优先级告警,系统存在严重问题"
	} else if score >= 2 {
		return "存在不同种类的低优先级告警"
	} else if score >= 1 {
		return "存在低优先级告警,无明显影响"
	} else {
		return "无告警,系统运行正常"
	}
}

var _ prom.MetricGroup = &RYGLightStatus{}

type servicesRYGLightMap = prom.MetricGroupMap[prom.ServiceKey, *RYGLightStatus]

func (s *RYGLightStatus) AppendGroupIfNotExist(metricGroup prom.MGroupName, metricName prom.MName) bool {
	return metricName == prom.LATENCY
}

func (s *RYGLightStatus) InitEmptyGroup(key prom.ConvertFromLabels) prom.MetricGroup {
	return &RYGLightStatus{}
}

func (s *RYGLightStatus) SetValue(groupName prom.MGroupName, metricName prom.MName, value float64) {
	if groupName == prom.AVG {
		if metricName == prom.LATENCY {
			micro := value / 1e3
			s.LatencyAvg = &micro
		}
		return
	}

	radio := (value - 1) * 100
	switch metricName {
	case prom.LATENCY:
		s.LatencyDoD = &radio
	case prom.ERROR_RATE:
		s.ErrorRateDoD = &radio
	case prom.LOG_ERROR_COUNT:
		s.LogErrorCountDoD = &radio
	}
}

func GroupAlertEventCountListByInstance(events []model.AlertEventCount, instances []*model.ServiceInstance) model.AlertEventLevelCountMap {
	var res = make(model.AlertEventLevelCountMap)
	for _, event := range events {
		for _, instance := range instances {
			if instance.MatchSvcTags(event.Group, event.Tags) {
				res.Add(clickhouse.GetAlertType(event.Group), event.Severity, event.AlarmCount)
				break
			}
		}
	}

	return res
}
