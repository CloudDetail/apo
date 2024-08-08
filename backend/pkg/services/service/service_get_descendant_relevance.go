package service

import (
	"fmt"
	"log"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/services/serviceoverview"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
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

	unsortedDescendantWithTarget := append(unsortedDescendant, polarisanalyzer.LatencyRelevance{
		Service:  req.Service,
		Endpoint: req.Endpoint,
	})

	// 按延时相似度排序
	sorted, unsorted, err := s.polRepo.SortDescendantByLatencyRelevance(
		req.StartTime, req.EndTime, prom.VecFromDuration(time.Duration(req.Step)*time.Microsecond),
		req.Service, req.Endpoint,
		unsortedDescendantWithTarget,
	)

	if err != nil {
		// TODO 排序失败,输出日志,但正常继续返回
	}

	// 将未能排序成功的下游添加到descendants后(可能是没有北极星指标)
	for _, descendant := range unsorted {
		sorted = append(unsorted, polarisanalyzer.LatencyRelevance{
			Service:  descendant.Service,
			Endpoint: descendant.Endpoint,
		})
	}

	var resp []response.GetDescendantRelevanceResponse
	descendantStatus, err := s.queryDescendantStatus(services, endpoints, req.StartTime, req.EndTime)

	threshold, err := s.dbRepo.GetOrCreateThreshold("", "", database.GLOBAL)
	for _, descendant := range sorted {
		var descendantResp = response.GetDescendantRelevanceResponse{
			ServiceName:      descendant.Service,
			EndPoint:         descendant.Endpoint,
			Distance:         descendant.Relevance,
			DelaySource:      "self",
			REDMetricsStatus: model.STATUS_NORMAL,
			// TODO 查询日志,K8s,基础设施,网络告警和最后部署时间
			LogMetricsStatus:     model.STATUS_NORMAL,
			InfrastructureStatus: model.STATUS_NORMAL,
			NetStatus:            model.STATUS_NORMAL,
			K8sStatus:            model.STATUS_NORMAL,
			LastUpdateTime:       nil,
		}
		//获取每个endpoint下的所有实例
		var instances []serviceoverview.Instance
		startTime := time.Unix(req.StartTime/1000000, 0)
		endTime := time.Unix(req.EndTime/1000000, 0)
		step := time.Duration(req.Step * 1000)
		var Res []prom.MetricResult
		query := prom.QueryNodeName(descendant.Service, descendant.Endpoint)
		Res, err := s.promRepo.QueryData(endTime, query)
		if err != nil || Res == nil {
			continue
		}
		for _, result := range Res {
			contentKey := result.Metric.ContentKey
			serviceName := result.Metric.SvcName
			nodeName := result.Metric.NodeName
			pod := result.Metric.POD
			pid := result.Metric.PID
			found := false
			for _, Instance := range instances {
				if Instance.ContentKey == contentKey && Instance.SvcName == serviceName {
					found = true
					break
				}
			}
			if !found && contentKey == descendant.Endpoint && serviceName == descendant.Service {
				newInstance := serviceoverview.Instance{
					ContentKey: contentKey,
					SvcName:    serviceName,
					Pod:        pod,
					NodeName:   nodeName,
					Pid:        pid,
				}
				instances = append(instances, newInstance)
			}
		}
		var searchName []string
		var NodeNames []string
		var Pids []string
		var Pods []string
		for _, instance := range instances {
			if instance.Pod != "" {
				searchName = append(searchName, instance.Pod)
				Pods = append(Pods, instance.Pod)
			}
			if instance.NodeName != "" {
				searchName = append(searchName, instance.NodeName)
				NodeNames = append(NodeNames, instance.NodeName)
			}
			if instance.Pid != "" {
				Pids = append(Pids, instance.Pid)
			}
		}

		if len(searchName) > 0 {
			isAlert, err := s.chRepo.InfrastructureAlert(startTime, endTime, searchName)
			if err != nil {
				log.Printf("Failed to query InfrastructureAlert: %v", err)
			}
			if isAlert {
				descendantResp.InfrastructureStatus = model.STATUS_CRITICAL
			}
			isAlert, err = s.chRepo.K8sAlert(startTime, endTime, searchName)
			if err != nil {
				log.Printf("Failed to query K8sAlert: %v", err)
			}
			if isAlert {
				descendantResp.K8sStatus = model.STATUS_CRITICAL
			}
		}
		if len(Pods) > 0 || len(Pids) > 0 {
			isAlert, err := s.chRepo.NetworkAlert(startTime, endTime, Pods, NodeNames, Pids)
			if err != nil {
				log.Printf("Failed to query NetworkAlert: %v", err)
			}
			if isAlert {
				descendantResp.NetStatus = model.STATUS_CRITICAL
			}
		}
		if len(Pids) > 0 {
			startTimeMap, _ := s.promRepo.QueryProcessStartTime(startTime, endTime, step, Pids)
			latestStartTime, found := serviceoverview.GetLatestStartTime(startTimeMap, instances)
			if found {
				descendantResp.LastUpdateTime = new(int64)
				*descendantResp.LastUpdateTime = latestStartTime * 1e6
			}
		}
		descendantKey := descendant.Service + "_" + descendant.Endpoint
		if status, ok := descendantStatus[descendantKey]; ok {
			if status.DepLatency > 0 && status.Latency > 0 {
				var depRatio = status.DepLatency / status.Latency
				if depRatio > 0.5 {
					descendantResp.DelaySource = "dependency"
					descendantResp.DelayDistribution = fmt.Sprintf("latency: %.2f, depLatency: %.2f(%.2f)", status.DepLatency, status.Latency, depRatio)
				} else {
					descendantResp.DelaySource = "self"
					descendantResp.DelayDistribution = fmt.Sprintf("latency: %.2f, depLatency: %.2f(%.2f)", status.DepLatency, status.Latency, depRatio)
				}
			} else {
				descendantResp.DelaySource = "self"
			}

			if status.RequestPerSecondDoD < 0 {
				descendantResp.REDAlarmReason = "未统计到TPS日同比,忽略TPS告警;"
			} else if threshold.Tps > 0 && status.RequestPerSecondDoD*100 > (100+threshold.Tps) {
				descendantResp.REDMetricsStatus = model.STATUS_CRITICAL
				descendantResp.REDAlarmReason = fmt.Sprintf("%s 请求TPS日同比: %.2f 高于设定阈值 %.2f;", descendantResp.REDAlarmReason, status.RequestPerSecondDoD, (100+threshold.Latency)/100)
			}

			if status.LatencyDoD < 0 {
				descendantResp.REDAlarmReason = descendantResp.REDAlarmReason + "未统计到延迟日同比,忽略延迟告警"
			} else if threshold.Latency > 0 && status.LatencyDoD*100 > (100+threshold.Latency) {
				descendantResp.REDMetricsStatus = model.STATUS_CRITICAL
				descendantResp.REDAlarmReason = fmt.Sprintf("%s 延迟日同比: %.2f 高于设定阈值 %.2f;", descendantResp.REDAlarmReason, status.LatencyDoD, (100+threshold.Latency)/100)
			}

			if status.ErrorRateDoD < 0 {
				descendantResp.REDAlarmReason = descendantResp.REDAlarmReason + "未统计错误率日同比,忽略错误率告警;"
			} else if threshold.ErrorRate > 0 && status.ErrorRateDoD*100 < (100-threshold.ErrorRate) {
				descendantResp.REDMetricsStatus = model.STATUS_CRITICAL
				descendantResp.REDAlarmReason = fmt.Sprintf("%s 错误率日同比: %.2f 低于设定阈值 %.2f;", descendantResp.REDAlarmReason, status.ErrorRateDoD, (100-threshold.ErrorRate)/100)
			}
		} else {
			descendantResp.REDAlarmReason = "时间段内未统计到应用延时,应用无请求或未监控,忽略RED告警;"
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
		prom.ServiceRegexPQLFilter, prom.MultipleValue(services...),
		prom.ContentKeyRegexPQLFilter, prom.MultipleValue(endpoints...))
	if err != nil {
		return nil, err
	}

	avgLatency, err := s.promRepo.QueryAggMetricsWithFilter(
		prom.PQLAvgLatencyWithFilters,
		startTime, endTime,
		prom.EndpointGranularity,
		prom.ServiceRegexPQLFilter, prom.MultipleValue(services...),
		prom.ContentKeyRegexPQLFilter, prom.MultipleValue(endpoints...))
	if err != nil {
		return nil, err
	}

	avgLatencyDoD, err := s.promRepo.QueryAggMetricsWithFilter(
		prom.DayOnDay(prom.PQLAvgLatencyWithFilters),
		startTime, endTime,
		prom.EndpointGranularity,
		prom.ServiceRegexPQLFilter, prom.MultipleValue(services...),
		prom.ContentKeyRegexPQLFilter, prom.MultipleValue(endpoints...))
	if err != nil {
		return nil, err
	}

	avgErrorRateDoD, err := s.promRepo.QueryAggMetricsWithFilter(
		prom.DayOnDay(prom.PQLAvgErrorRateWithFilters),
		startTime, endTime,
		prom.EndpointGranularity,
		prom.ServiceRegexPQLFilter, prom.MultipleValue(services...),
		prom.ContentKeyRegexPQLFilter, prom.MultipleValue(endpoints...))
	if err != nil {
		return nil, err
	}
	avgRequestPerSecondDoD, err := s.promRepo.QueryAggMetricsWithFilter(
		prom.DayOnDay(prom.PQLAvgTPSWithFilters),
		startTime, endTime,
		prom.EndpointGranularity,
		prom.ServiceRegexPQLFilter, prom.MultipleValue(services...),
		prom.ContentKeyRegexPQLFilter, prom.MultipleValue(endpoints...))

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
