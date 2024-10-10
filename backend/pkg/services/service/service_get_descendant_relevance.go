package service

import (
	"fmt"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/polarisanalyzer"
	prom "github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/serviceoverview"
)

// GetDescendantRelevance implements Service.
func (s *service) GetDescendantRelevance(req *request.GetDescendantRelevanceRequest) ([]response.GetDescendantRelevanceResponse, error) {
	// 查询所有子孙节点
	nodes, err := s.chRepo.ListDescendantNodes(req)
	if err != nil {
		return nil, err
	}

	if len(nodes) == 0 {
		return make([]response.GetDescendantRelevanceResponse, 0), nil
	}

	unsortedDescendant := make([]polarisanalyzer.LatencyRelevance, 0, len(nodes))
	var services, endpoints []string
	for _, node := range nodes {
		unsortedDescendant = append(unsortedDescendant, polarisanalyzer.LatencyRelevance{
			Service:  node.Service,
			Endpoint: node.Endpoint,
		})
		services = append(services, node.Service)
		endpoints = append(endpoints, node.Endpoint)
	}

	// 按延时相似度排序
	sortResp, err := s.polRepo.SortDescendantByLatencyRelevance(
		req.StartTime, req.EndTime, prom.VecFromDuration(time.Duration(req.Step)*time.Microsecond),
		req.Service, req.Endpoint,
		unsortedDescendant,
	)
	var sortResult []polarisanalyzer.LatencyRelevance
	var sortType string
	if err != nil || sortResp == nil {
		sortResult = unsortedDescendant
		sortType = "net_failed"
	} else {
		sortResult = sortResp.SortedDescendant
		sortType = sortResp.DistanceType
		// 将未能排序成功的下游添加到descendants后(可能是没有北极星指标)
		for _, descendant := range sortResp.UnsortedDescendant {
			sortResult = append(sortResult, polarisanalyzer.LatencyRelevance{
				Service:  descendant.Service,
				Endpoint: descendant.Endpoint,
			})
		}
	}

	var resp []response.GetDescendantRelevanceResponse
	descendantStatus, err := s.queryDescendantStatus(services, endpoints, req.StartTime, req.EndTime)
	if err != nil {
		// TODO 添加日志,查询RED指标失败
	}
	threshold, err := s.dbRepo.GetOrCreateThreshold("", "", database.GLOBAL)
	if err != nil {
		// TODO 添加日志,查询阈值失败
	}
	for _, descendant := range sortResult {
		var descendantResp = response.GetDescendantRelevanceResponse{
			ServiceName:      descendant.Service,
			EndPoint:         descendant.Endpoint,
			Distance:         descendant.Relevance,
			DistanceType:     sortType,
			DelaySource:      "unknown",
			REDMetricsStatus: model.STATUS_NORMAL,
			AlertStatus:      model.NORMAL_ALERT_STATUS,
			AlertReason:      model.AlertReason{},
			LastUpdateTime:   nil,
		}

		// 填充延时源和RED告警 (DelaySource/REDMetricsStatus)
		fillServiceDelaySourceAndREDAlarm(&descendantResp, descendantStatus, threshold)

		// 获取每个endpoint下的所有实例
		instances, err := s.promRepo.GetInstanceList(req.StartTime, req.EndTime, descendant.Service, descendant.Endpoint)
		if err != nil {
			// TODO deal error
			continue
		}

		startTime := time.UnixMicro(req.StartTime)
		endTime := time.UnixMicro(req.EndTime)

		instanceList := instances.GetInstances()

		// 填充告警状态
		descendantResp.AlertStatusCH = serviceoverview.GetAlertStatusCH(
			s.chRepo, &descendantResp.AlertReason, nil,
			[]string{}, descendant.Service, instanceList,
			startTime, endTime,
		)

		// 查询并填充进程启动时间
		startTSmap, _ := s.promRepo.QueryProcessStartTime(startTime, endTime, instanceList)
		latestStartTime := getLatestStartTime(startTSmap) * 1e6
		if latestStartTime > 0 {
			descendantResp.LastUpdateTime = &latestStartTime
		}
		resp = append(resp, descendantResp)
	}

	return resp, nil
}

func (s *service) queryDescendantStatus(services []string, endpoints []string, startTime, endTime int64) (*DescendantStatusMap, error) {
	avgDepLatency, err := s.promRepo.QueryAggMetricsWithFilter(
		prom.WithDefaultIFPolarisMetricExits(prom.PQLAvgDepLatencyWithFilters, prom.DefaultDepLatency),
		startTime, endTime,
		prom.EndpointGranularity,
		prom.ServiceRegexPQLFilter, prom.RegexMultipleValue(services...),
		prom.ContentKeyRegexPQLFilter, prom.RegexMultipleValue(endpoints...))
	if err != nil {
		return nil, err
	}

	avgLatency, err := s.promRepo.QueryAggMetricsWithFilter(
		prom.PQLAvgLatencyWithFilters,
		startTime, endTime,
		prom.EndpointGranularity,
		prom.ServiceRegexPQLFilter, prom.RegexMultipleValue(services...),
		prom.ContentKeyRegexPQLFilter, prom.RegexMultipleValue(endpoints...))
	if err != nil {
		return nil, err
	}

	avgLatencyDoD, err := s.promRepo.QueryAggMetricsWithFilter(
		prom.DayOnDay(prom.PQLAvgLatencyWithFilters),
		startTime, endTime,
		prom.EndpointGranularity,
		prom.ServiceRegexPQLFilter, prom.RegexMultipleValue(services...),
		prom.ContentKeyRegexPQLFilter, prom.RegexMultipleValue(endpoints...))
	if err != nil {
		return nil, err
	}

	avgErrorRateDoD, err := s.promRepo.QueryAggMetricsWithFilter(
		prom.DayOnDay(prom.PQLAvgErrorRateWithFilters),
		startTime, endTime,
		prom.EndpointGranularity,
		prom.ServiceRegexPQLFilter, prom.RegexMultipleValue(services...),
		prom.ContentKeyRegexPQLFilter, prom.RegexMultipleValue(endpoints...))
	if err != nil {
		return nil, err
	}
	avgRequestPerSecondDoD, err := s.promRepo.QueryAggMetricsWithFilter(
		prom.DayOnDay(prom.PQLAvgTPSWithFilters),
		startTime, endTime,
		prom.EndpointGranularity,
		prom.ServiceRegexPQLFilter, prom.RegexMultipleValue(services...),
		prom.ContentKeyRegexPQLFilter, prom.RegexMultipleValue(endpoints...))

	if err != nil {
		return nil, err
	}

	var descendantStatusMap = &DescendantStatusMap{
		MetricGroupList: []*DescendantStatus{},
		MetricGroupMap:  map[prom.EndpointKey]*DescendantStatus{},
	}

	descendantStatusMap.MergeMetricResults(prom.AVG, prom.LATENCY, avgLatency)
	descendantStatusMap.MergeMetricResults(prom.AVG, prom.DEP_LATENCY, avgDepLatency)
	descendantStatusMap.MergeMetricResults(prom.DOD, prom.LATENCY, avgLatencyDoD)
	descendantStatusMap.MergeMetricResults(prom.DOD, prom.ERROR_RATE, avgErrorRateDoD)
	descendantStatusMap.MergeMetricResults(prom.DOD, prom.THROUGHPUT, avgRequestPerSecondDoD)

	return descendantStatusMap, err
}

type DescendantStatusMap = prom.MetricGroupMap[prom.EndpointKey, *DescendantStatus]

func (s *DescendantStatus) InitEmptyGroup(_ prom.ConvertFromLabels) prom.MetricGroup {
	return &DescendantStatus{
		DepLatency:          nil,
		Latency:             nil,
		LatencyDoD:          nil,
		ErrorRateDoD:        nil,
		RequestPerSecondDoD: nil,
	}
}

func (s *DescendantStatus) AppendGroupIfNotExist(_ prom.MGroupName, metricName prom.MName) bool {
	return metricName == prom.LATENCY
}

func (s *DescendantStatus) SetValue(metricGroup prom.MGroupName, metricName prom.MName, value float64) {
	switch metricGroup {
	case prom.AVG:
		switch metricName {
		case prom.LATENCY:
			micros := value / 1e3
			s.Latency = &micros
		case prom.DEP_LATENCY:
			micros := value / 1e3
			s.DepLatency = &micros
		}
	case prom.DOD:
		radio := (value - 1) * 100
		switch metricName {
		case prom.LATENCY:
			s.LatencyDoD = &radio
		case prom.ERROR_RATE:
			s.ErrorRateDoD = &radio
		case prom.THROUGHPUT:
			s.RequestPerSecondDoD = &radio
		}
	}
}

type DescendantStatus struct {
	DepLatency *float64
	Latency    *float64

	LatencyDoD          *float64 // 延迟日同比
	ErrorRateDoD        *float64 // 错误率日同比
	RequestPerSecondDoD *float64 // 请求数日同比
}

func fillServiceDelaySourceAndREDAlarm(descendantResp *response.GetDescendantRelevanceResponse, descendantStatus *DescendantStatusMap, threshold database.Threshold) {
	ts := time.Now()
	descendantKey := prom.EndpointKey{
		ContentKey: descendantResp.EndPoint,
		SvcName:    descendantResp.ServiceName,
	}

	if status, ok := descendantStatus.MetricGroupMap[descendantKey]; ok {
		if status.DepLatency != nil && status.Latency != nil {
			var depRatio = *status.DepLatency / *status.Latency
			if depRatio > 0.5 {
				descendantResp.DelaySource = "dependency"
				delayDistribution := fmt.Sprintf("总延时: %.2f, 外部依赖延时: %.2f(%.2f)", *status.Latency, *status.DepLatency, depRatio)
				descendantResp.AlertReason.Add(model.DelaySourceAlert, model.AlertDetail{
					Timestamp:    ts.UnixMicro(),
					AlertObject:  descendantResp.ServiceName,
					AlertReason:  "外部依赖延时占总延时超过50%",
					AlertMessage: delayDistribution,
				})
			} else {
				descendantResp.DelaySource = "self"
			}
		} else {
			descendantResp.DelaySource = "unknown"
		}

		if status.RequestPerSecondDoD == nil {
			descendantResp.AlertReason.Add(model.REDMetricsAlert, model.AlertDetail{
				Timestamp:    ts.UnixMicro(),
				AlertObject:  descendantResp.ServiceName,
				AlertReason:  "TPS未采集到数据",
				AlertMessage: "",
			})
		} else if threshold.Tps > 0 && *status.RequestPerSecondDoD > threshold.Tps {
			descendantResp.REDMetricsStatus = model.STATUS_CRITICAL
			descendantResp.AlertReason.Add(model.REDMetricsAlert, model.AlertDetail{
				Timestamp:    ts.UnixMicro(),
				AlertObject:  descendantResp.ServiceName,
				AlertReason:  "TPS变化超过日同比阈值",
				AlertMessage: fmt.Sprintf("请求TPS日同比增长: %.2f%% 高于设定阈值 %.2f%%;", *status.RequestPerSecondDoD, threshold.Tps),
			})
		}

		if status.LatencyDoD == nil {
			descendantResp.AlertReason.Add(model.REDMetricsAlert, model.AlertDetail{
				Timestamp:    ts.UnixMicro(),
				AlertObject:  descendantResp.ServiceName,
				AlertReason:  "延迟未采集到数据",
				AlertMessage: "",
			})
		} else if threshold.Latency > 0 && *status.LatencyDoD > threshold.Latency {
			descendantResp.REDMetricsStatus = model.STATUS_CRITICAL
			descendantResp.AlertReason.Add(model.REDMetricsAlert, model.AlertDetail{
				Timestamp:    ts.UnixMicro(),
				AlertObject:  descendantResp.ServiceName,
				AlertReason:  "延时变化超过日同比阈值",
				AlertMessage: fmt.Sprintf("延迟日同比增长: %.2f%% 高于设定阈值 %.2f%%;", *status.LatencyDoD, threshold.Latency),
			})
		}

		if status.ErrorRateDoD == nil {
			descendantResp.AlertReason.Add(model.REDMetricsAlert, model.AlertDetail{
				Timestamp:    ts.UnixMicro(),
				AlertObject:  descendantResp.ServiceName,
				AlertReason:  "错误率未采集到数据",
				AlertMessage: "",
			})
		} else if threshold.ErrorRate > 0 && *status.ErrorRateDoD > threshold.ErrorRate {
			descendantResp.REDMetricsStatus = model.STATUS_CRITICAL
			descendantResp.AlertReason.Add(model.REDMetricsAlert, model.AlertDetail{
				Timestamp:    ts.UnixMicro(),
				AlertObject:  descendantResp.ServiceName,
				AlertReason:  "错误率变化超过日同比阈值",
				AlertMessage: fmt.Sprintf("错误率日同比增长: %.2f%% 高于设定阈值 %.2f%%;", *status.ErrorRateDoD, threshold.ErrorRate),
			})
		}
	} else {
		descendantResp.AlertReason.Add(model.REDMetricsAlert, model.AlertDetail{
			Timestamp:    ts.UnixMicro(),
			AlertObject:  descendantResp.ServiceName,
			AlertReason:  "时间段内未统计到应用延时,应用无请求或未监控,忽略RED告警;",
			AlertMessage: "",
		})
	}
}

func getLatestStartTime(startTSmap map[model.ServiceInstance]int64) int64 {
	var latestStartTime int64 = -1
	for _, startTime := range startTSmap {
		if startTime > latestStartTime {
			latestStartTime = startTime
		}
	}
	return latestStartTime
}
