package service

import (
	"fmt"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/polarisanalyzer"
	prom "github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
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
	// sorted, unsorted, err :=
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
			ServiceName:          descendant.Service,
			EndPoint:             descendant.Endpoint,
			Distance:             descendant.Relevance,
			DistanceType:         sortType,
			DelaySource:          "self",
			REDMetricsStatus:     model.STATUS_NORMAL,
			LogMetricsStatus:     model.STATUS_NORMAL,
			InfrastructureStatus: model.STATUS_NORMAL,
			NetStatus:            model.STATUS_NORMAL,
			K8sStatus:            model.STATUS_NORMAL,
			LastUpdateTime:       nil,
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
		// 查询实例相关的告警信息
		events, _ := s.chRepo.GetAlertEventsSample(1, startTime, endTime,
			request.AlertFilter{Service: req.Service, Status: "firing"}, instanceList)

		// 按告警原因修改告警状态/
		for _, event := range events {
			switch event.Group {
			case clickhouse.INFRA_GROUP:
				descendantResp.InfrastructureStatus = model.STATUS_CRITICAL
				descendantResp.AlertReason.Add("infra", fmt.Sprintf("%s: %s", event.ReceivedTime.Format("15:04:05"), event.Name))
			case clickhouse.NETWORK_GROUP:
				descendantResp.NetStatus = model.STATUS_CRITICAL
				descendantResp.AlertReason.Add("net", fmt.Sprintf("%s: %s", event.ReceivedTime.Format("15:04:05"), event.Name))
			default:
				// 忽略 app 和 container 告警
				continue
			}
		}

		// 查询warning及以上级别的K8s事件
		k8sEvents, _ := s.chRepo.GetK8sAlertEventsSample(startTime, endTime, instanceList)
		if len(k8sEvents) > 0 {
			descendantResp.K8sStatus = model.STATUS_CRITICAL
			for _, event := range k8sEvents {
				info := fmt.Sprintf("%s: %s %s:%s", event.Timestamp.Format("15:04:05"), event.GetObjName(), event.GetReason(), event.Body)
				descendantResp.AlertReason.Add("k8s", info)
			}
		}

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

func (s *service) queryDescendantStatus(services []string, endpoints []string, startTime, endTime int64) (map[string]*DescendantStatus, error) {
	avgDepLatency, err := s.promRepo.QueryAggMetricsWithFilter(
		prom.PQLAvgDepLatencyWithFilters,
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

	var descendantStatusMap = make(map[string]*DescendantStatus)
	for _, metric := range avgLatency {
		status := &DescendantStatus{
			DepLatency:          -1,
			Latency:             metric.Values[0].Value,
			LatencyDoD:          -1,
			ErrorRateDoD:        -1,
			RequestPerSecondDoD: -1,
		}
		descendantStatusMap[metric.Metric.SvcName+"_"+metric.Metric.ContentKey] = status
	}

	for _, metric := range avgDepLatency {
		status, find := descendantStatusMap[metric.Metric.SvcName+"_"+metric.Metric.ContentKey]
		if find {
			status.DepLatency = metric.Values[0].Value
		}
	}

	for _, metric := range avgLatencyDoD {
		status, find := descendantStatusMap[metric.Metric.SvcName+"_"+metric.Metric.ContentKey]
		if find {
			status.LatencyDoD = metric.Values[0].Value
		}
	}

	for _, metric := range avgErrorRateDoD {
		status, find := descendantStatusMap[metric.Metric.SvcName+"_"+metric.Metric.ContentKey]
		if find {
			status.ErrorRateDoD = metric.Values[0].Value
		}
	}

	for _, metric := range avgRequestPerSecondDoD {
		status, find := descendantStatusMap[metric.Metric.SvcName+"_"+metric.Metric.ContentKey]
		if find {
			status.RequestPerSecondDoD = metric.Values[0].Value
		}
	}

	return descendantStatusMap, err
}

type DescendantStatus struct {
	DepLatency float64
	Latency    float64

	LatencyDoD          float64 // 延迟日同比
	ErrorRateDoD        float64 // 错误率日同比
	RequestPerSecondDoD float64 // 请求数日同比
}

func fillServiceDelaySourceAndREDAlarm(descendantResp *response.GetDescendantRelevanceResponse, descendantStatus map[string]*DescendantStatus, threshold database.Threshold) {
	descendantKey := descendantResp.ServiceName + "_" + descendantResp.EndPoint
	if status, ok := descendantStatus[descendantKey]; ok {
		if status.DepLatency > 0 && status.Latency > 0 {
			var depRatio = status.DepLatency / status.Latency
			if depRatio > 0.5 {
				descendantResp.DelaySource = "dependency"
			} else {
				descendantResp.DelaySource = "self"
			}
			delayDistribution := fmt.Sprintf("latency: %.2f, depLatency: %.2f(%.2f)", status.DepLatency, status.Latency, depRatio)
			descendantResp.AlertReason.Add("delaySource", delayDistribution)
		} else {
			descendantResp.DelaySource = "self"
		}

		if status.RequestPerSecondDoD < 0 {
			descendantResp.AlertReason.Add("RED", "TPS: 未采集到数据")
		} else if threshold.Tps > 0 && status.RequestPerSecondDoD*100 > (100+threshold.Tps) {
			descendantResp.REDMetricsStatus = model.STATUS_CRITICAL
			descendantResp.AlertReason.Add("RED", fmt.Sprintf("TPS: 请求TPS日同比: %.2f 高于设定阈值 %.2f;", status.RequestPerSecondDoD, (100+threshold.Tps)/100))
		}

		if status.LatencyDoD < 0 {
			descendantResp.AlertReason.Add("RED", "延迟: 未采集到数据")
		} else if threshold.Latency > 0 && status.LatencyDoD*100 > (100+threshold.Latency) {
			descendantResp.REDMetricsStatus = model.STATUS_CRITICAL
			descendantResp.AlertReason.Add("RED", fmt.Sprintf("延迟: 延迟日同比: %.2f 高于设定阈值 %.2f;", status.LatencyDoD, (100+threshold.Latency)/100))
		}

		if status.ErrorRateDoD < 0 {
			descendantResp.AlertReason.Add("RED", "错误率: 未采集到数据")
		} else if threshold.ErrorRate > 0 && status.ErrorRateDoD*100 < (100-threshold.ErrorRate) {
			descendantResp.REDMetricsStatus = model.STATUS_CRITICAL
			descendantResp.AlertReason.Add("RED", fmt.Sprintf("错误率: 错误率日同比: %.2f 低于设定阈值 %.2f;", status.ErrorRateDoD, (100-threshold.ErrorRate)/100))
		}
	} else {
		descendantResp.AlertReason.Add("RED", "时间段内未统计到应用延时,应用无请求或未监控,忽略RED告警;")
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
