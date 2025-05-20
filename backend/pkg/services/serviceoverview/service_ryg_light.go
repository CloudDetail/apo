// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

import (
	"fmt"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	prom "github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

func (s *service) GetServicesRYGLightStatus(ctx core.Context, startTime time.Time, endTime time.Time, filter EndpointsFilter) (response.ServiceRYGLightRes, error) {
	var servicesMap = &servicesRYGLightMap{
		MetricGroupList: []*RYGLightStatus{},
		MetricGroupMap:  map[prom.ServiceKey]*RYGLightStatus{},
	}

	startMicroTS := startTime.UnixMicro()
	endMicroTs := endTime.UnixMicro()

	filters := filter.ExtractFilterStr()

	// FIX for showing services without LatencyDay-over-Day Growth Rate
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
		prom.DayOnDay(prom.PQLAvgLogErrorCountCombineEndpointsInfoWithFilters),
		startMicroTS, endMicroTs,
		prom.SVCGranularity, filters...)
	if err == nil {
		servicesMap.MergeMetricResults(prom.DOD, prom.LOG_ERROR_COUNT, avgLogErrorCountDoD)
	}

	var resp = response.ServiceRYGLightRes{
		ServiceList: []*response.ServiceRYGResult{},
	}

	alertEventCount, _ := s.chRepo.GetAlertEventCountGroupByInstance(ctx, startTime,
		endTime,
		request.AlertFilter{Status: "firing"},
		nil,
	)
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
			RYGResult:   *status.ExposeRYGLightStatus(ctx.LANG()),
		})
	}

	return resp, nil
}

// RYGLightStatus
// Red/Green/Yellow Status
type RYGLightStatus struct {
	// From Prometheus
	LatencyAvg       *float64 // average latency for statistical services
	ErrorRateAvg     *float64 // The average error rate is used get the current error situation
	LogErrorCountAvg *float64 // average number of log errors is used to get the current log error condition

	LatencyDoD       *float64 // Delay Day-over-Day Growth Rate
	ErrorRateDoD     *float64 // Error Rate Day-over-Day Growth Rate
	LogErrorCountDoD *float64 // log error Day-over-Day Growth Rate

	Instances []*model.ServiceInstance

	// From Clickhouse
	model.AlertStatus             // alarm status
	model.AlertEventLevelCountMap // Alarm event level statistics
}

func (s *RYGLightStatus) ExposeRYGLightStatus(lang string) *response.RYGResult {
	var res = &response.RYGResult{}

	latencyScore := ScoreFromDoD(s.LatencyDoD, 10, 20, 50)
	if latencyScore >= 0 {
		res.Score += latencyScore

		var detail string
		if lang == code.LANG_EN {
			detail = fmt.Sprintf("latency day-on-day growth %.2f%%", *s.LatencyDoD)
		} else {
			detail = fmt.Sprintf("latency 同比增长 %.2f%%", *s.LatencyDoD)
		}
		res.ScoreDetail = append(res.ScoreDetail, response.RYGScoreDetail{
			Key:    "latency",
			Score:  latencyScore,
			Detail: detail,
		})
	} else if latencyScore == -1 {
		res.Score += 3

		var detail string
		if lang == code.LANG_EN {
			detail = "Failed to get yesterday's latency metric, skipping check"
		} else {
			detail = "未获取到昨日延时,跳过检查"
		}

		res.ScoreDetail = append(res.ScoreDetail, response.RYGScoreDetail{
			Key:    "latency",
			Score:  3,
			Detail: detail,
		})
	}

	errorRateScore := ScoreFromDoD(s.ErrorRateDoD, 5, 10, 20)
	if errorRateScore >= 0 {
		res.Score += errorRateScore

		var detail string
		if lang == code.LANG_EN {
			detail = fmt.Sprintf("errorRate day-on-day growth %.2f%%", *s.ErrorRateDoD)
		} else {
			detail = fmt.Sprintf("errorRate 同比增长 %.2f%%", *s.ErrorRateDoD)
		}
		res.ScoreDetail = append(res.ScoreDetail, response.RYGScoreDetail{
			Key:    "errorRate",
			Score:  errorRateScore,
			Detail: detail,
		})
	} else if s.ErrorRateAvg != nil && *s.ErrorRateAvg > 0 {
		var detail string
		if lang == code.LANG_EN {
			detail = fmt.Sprintf("errorRate There were no errors at the same time yesterday. Today's error rate: %.2f%%", *s.ErrorRateAvg)
		} else {
			detail = fmt.Sprintf("errorRate 昨日同时期无错误,今日错误率: %.2f%%", *s.ErrorRateAvg)
		}

		res.ScoreDetail = append(res.ScoreDetail, response.RYGScoreDetail{
			Key:    "errorRate",
			Score:  0,
			Detail: detail,
		})
	} else {

		var detail string
		if lang == code.LANG_EN {
			detail = "errorRate No errors occurred today."
		} else {
			detail = "errorRate 今日未发生错误"
		}

		res.Score += 3
		res.ScoreDetail = append(res.ScoreDetail, response.RYGScoreDetail{
			Key:    "errorRate",
			Score:  3,
			Detail: detail,
		})
	}

	logErrorCountScore := ScoreFromDoD(s.LogErrorCountDoD, 5, 10, 20)
	if logErrorCountScore >= 0 {
		res.Score += logErrorCountScore

		var detail string
		if lang == code.LANG_EN {
			detail = fmt.Sprintf("logErrorCount day-on-day growth %.2f%%", *s.LogErrorCountDoD)
		} else {
			detail = fmt.Sprintf("logErrorCount 同比增长 %.2f%%", *s.LogErrorCountDoD)
		}

		res.ScoreDetail = append(res.ScoreDetail, response.RYGScoreDetail{
			Key:    "logErrorCount",
			Score:  logErrorCountScore,
			Detail: detail,
		})
	} else if s.LogErrorCountAvg != nil && *s.LogErrorCountAvg > 0 {
		var detail string
		if lang == code.LANG_EN {
			detail = fmt.Sprintf("logErrorCount There were no errors at the same time yesterday. Today's error count: %.2f%%", *s.LogErrorCountDoD)
		} else {
			detail = fmt.Sprintf("logErrorCount 昨日同时期无错误,今日日志错误数: %.0f%%", *s.ErrorRateAvg)
		}

		res.ScoreDetail = append(res.ScoreDetail, response.RYGScoreDetail{
			Key:    "logErrorCount",
			Score:  0,
			Detail: detail,
		})
	} else {
		res.Score += 3

		var detail string
		if lang == code.LANG_EN {
			detail = "logErrorCount No errors occurred today."
		} else {
			detail = "logErrorCount 今日未产生错误日志"
		}

		res.ScoreDetail = append(res.ScoreDetail, response.RYGScoreDetail{
			Key:    "logErrorCount",
			Score:  3,
			Detail: detail,
		})
	}

	alertScore := AlertScore(&s.AlertStatus, &s.AlertEventLevelCountMap)
	if alertScore >= 0 {
		res.Score += alertScore
		res.ScoreDetail = append(res.ScoreDetail, response.RYGScoreDetail{
			Key:    "alert",
			Score:  alertScore,
			Detail: getAlertScoreDetail(lang, alertScore),
		})
	}

	if len(s.Instances) < 2 {
		var detail string
		if lang == code.LANG_EN {
			detail = "The number of application instances is less than 2, there is a risk of service unavailability"
		} else {
			detail = "应用实例数小于2, 存在服务不可用风险"
		}

		res.ScoreDetail = append(res.ScoreDetail, response.RYGScoreDetail{
			Key:    "replica",
			Score:  0,
			Detail: detail,
		})
	} else {
		var detail string
		if lang == code.LANG_EN {
			detail = "Number of application instances is greater than 2, and the service has available replicas"
		} else {
			detail = "应用实例数大于2, 服务有可用副本"
		}

		res.Score += 3
		res.ScoreDetail = append(res.ScoreDetail, response.RYGScoreDetail{
			Key:    "replica",
			Score:  3,
			Detail: detail,
		})
	}

	res.PercentScore = res.Score * 100 / response.MAX_RYG_SCORE

	if res.PercentScore >= 80 {
		res.Status = response.GREEN
	} else if res.PercentScore < 80 && res.PercentScore >= 40 {
		res.Status = response.YELLOW
	} else if res.PercentScore < 40 {
		res.Status = response.RED
	}
	return res
}

func ScoreFromDoD(value *float64, l1, l2, l3 float64) int {
	if value == nil {
		return -1
	}
	if *value < l1 {
		return 3
	} else if *value < l2 {
		return 2
	} else if *value < l3 {
		return 1
	}
	return 0
}

func AlertScore(status *model.AlertStatus, eventLevelCountMap *model.AlertEventLevelCountMap) int {
	if status == nil || status.IsAllNormal() {
		return 3
	}

	if eventLevelCountMap == nil {
		return 3
	}

	var warningCount int
	for _, eventLevelCounts := range *eventLevelCountMap {
		count, find := eventLevelCounts[model.SeverityLevelCritical]
		if find && count > 0 {
			return 0
		}
		count, find = eventLevelCounts[model.SeverityLevelError]
		if find && count > 0 {
			return 0
		}

		count, find = eventLevelCounts[model.SeverityLevelWarning]
		if find && count > 0 {
			warningCount++
		}
	}
	if warningCount >= 2 {
		return 1
	} else if warningCount >= 1 {
		return 2
	}
	return 3
}

func getAlertScoreDetail(lang string, score int) string {
	if score <= 0 {
		if lang == code.LANG_EN {
			return "There is a high-priority alert, the system has serious problems"
		} else {
			return "存在高优先级告警,系统存在严重问题"
		}
	} else if score <= 1 {
		if lang == code.LANG_EN {
			return "There are different types of low-priority alerts"
		} else {
			return "存在不同种类的低优先级告警"
		}
	} else if score <= 2 {
		if lang == code.LANG_EN {
			return "There are low-priority alerts with no significant impact."
		} else {
			return "存在低优先级告警,无明显影响"
		}
	} else {

		if lang == code.LANG_EN {
			return "No alerts, the system is running normally."
		} else {
			return "无告警,系统运行正常"
		}
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

func (s *RYGLightStatus) SetValues(roupName prom.MGroupName, metricName prom.MName, points []prom.Points) {
	// Do Nothing
}

func (s *RYGLightStatus) SetValue(groupName prom.MGroupName, metricName prom.MName, value float64) {
	if groupName == prom.AVG {
		switch metricName {
		case prom.LATENCY:
			micro := value / 1e3
			s.LatencyAvg = &micro
		case prom.ERROR_RATE:
			errorRate := value * 100
			s.ErrorRateAvg = &errorRate
		case prom.LOG_ERROR_COUNT:
			s.LogErrorCountAvg = &value
		}
		return
	} else if groupName == prom.DOD {
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
