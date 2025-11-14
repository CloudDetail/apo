// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"fmt"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/polarisanalyzer"
	prom "github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/common"
	"github.com/CloudDetail/apo/backend/pkg/services/serviceoverview"
	"go.uber.org/zap"
)

// GetDescendantRelevance implements Service.
func (s *service) GetDescendantRelevance(ctx core.Context, req *request.GetDescendantRelevanceRequest) ([]response.GetDescendantRelevanceResponse, error) {
	// Query all descendant nodes
	nodes, err := s.chRepo.ListDescendantNodes(ctx, req)
	if err != nil {
		return nil, err
	}

	nodes, err = common.MarkTopologyNodeInGroup(ctx, s.dbRepo, req.GroupID, nodes)
	if err != nil {
		return nil, err
	}
	if len(nodes.Nodes) == 0 {
		return make([]response.GetDescendantRelevanceResponse, 0), nil
	}

	unsortedDescendant := make([]polarisanalyzer.Relevance, 0, len(nodes.Nodes))
	descendants := make([]polarisanalyzer.ServiceNode, 0, len(nodes.Nodes))
	var isTracedMap = make(map[polarisanalyzer.ServiceNode]bool)

	var svcTmp = make(map[string]struct{})
	var contentKeyTmp = make(map[string]struct{})
	var svcList, contentKeyList []string
	for _, node := range nodes.Nodes {
		svcNode := polarisanalyzer.ServiceNode{
			Service:  node.Service,
			Endpoint: node.Endpoint,
			Group:    node.Group,
			System:   node.System,
		}
		unsortedDescendant = append(unsortedDescendant, polarisanalyzer.Relevance{
			ServiceNode: svcNode,
			Relevance:   0,
		})
		descendants = append(descendants, svcNode)
		isTracedMap[svcNode] = node.IsTraced

		if _, ok := svcTmp[svcNode.Service]; !ok {
			svcList = append(svcList, svcNode.Service)
			svcTmp[svcNode.Service] = struct{}{}
		}

		if _, ok := contentKeyTmp[svcNode.Endpoint]; !ok {
			contentKeyList = append(contentKeyList, svcNode.Endpoint)
			contentKeyTmp[svcNode.Endpoint] = struct{}{}
		}
	}

	// Sort by Delay Similarity
	sortResp, err := s.polRepo.SortDescendantByRelevance(
		req.StartTime, req.EndTime, prom.VecFromDuration(time.Duration(req.Step)*time.Microsecond),
		req.ClusterIDs, req.Service, req.Endpoint,
		descendants, "",
	)
	var sortResult []polarisanalyzer.Relevance
	var sortType string
	if err != nil || sortResp == nil {
		sortResult = unsortedDescendant
		sortType = "net_failed"
	} else {
		sortResult = sortResp.SortedDescendant
		sortType = sortResp.DistanceType
		// Add the downstream that failed to sort successfully to the descendants (maybe there is no Polaris metric)
		sortResult = append(sortResult, sortResp.UnsortedDescendant...)
	}

	var resp []response.GetDescendantRelevanceResponse

	groupFilter, err := common.GetPQLFilterByGroupID(ctx, s.dbRepo, "", req.GroupID)
	if err != nil {
		return nil, err
	}

	svcFilter := prom.RegexMatchFilter(prom.ServiceNameKey, prom.RegexMultipleValue(svcList...)).
		RegexMatch(prom.ContentKeyKey, prom.RegexMultipleValue(contentKeyList...))
	pqlFilter := prom.And(groupFilter, svcFilter)

	descendantStatus, err := s.queryDescendantStatus(ctx, pqlFilter, req.StartTime, req.EndTime)
	if err != nil {
		s.logger.Error("Failed to query RED metric", zap.Error(err))
	}
	threshold, err := s.dbRepo.GetOrCreateThreshold(ctx, "", "", database.GLOBAL)
	if err != nil {
		s.logger.Error("Failed to query the threshold", zap.Error(err))
	}
	for _, descendant := range sortResult {
		var descendantResp = response.GetDescendantRelevanceResponse{
			ServiceName:      descendant.Service,
			EndPoint:         descendant.Endpoint,
			Group:            descendant.Group,
			IsTraced:         isTracedMap[descendant.ServiceNode],
			Distance:         descendant.Relevance,
			DistanceType:     sortType,
			DelaySource:      "unknown",
			REDMetricsStatus: model.STATUS_NORMAL,
			AlertStatus:      model.NORMAL_ALERT_STATUS,
			AlertReason:      model.AlertReason{},
			LastUpdateTime:   nil,
		}

		// Fill delay source and RED alarm (DelaySource/REDMetricsStatus)
		fillServiceDelaySourceAndREDAlarm(&descendantResp, descendantStatus, threshold)

		// Get all instances under each endpoint
		instances, err := s.promRepo.GetInstanceListByPQLFilter(ctx, req.StartTime, req.EndTime, pqlFilter)
		if err != nil {
			s.logger.Error("Failed to query instance list", zap.Error(err))
		}

		startTime := time.UnixMicro(req.StartTime)
		endTime := time.UnixMicro(req.EndTime)

		instanceList := instances.GetInstances()

		// fill alarm status
		descendantResp.AlertStatusCH = serviceoverview.GetAlertStatusCH(
			ctx,
			s.chRepo, &descendantResp.AlertReason, nil,
			[]string{}, descendant.Service, instanceList,
			startTime, endTime,
		)

		// Query and populate the process start time
		startTSmap, _ := s.promRepo.QueryProcessStartTime(ctx, startTime, endTime, instanceList)
		latestStartTime := getLatestStartTime(startTSmap) * 1e6
		if latestStartTime > 0 {
			descendantResp.LastUpdateTime = &latestStartTime
		}
		resp = append(resp, descendantResp)
	}

	return resp, nil
}

func (s *service) queryDescendantStatus(ctx core.Context, filter prom.PQLFilter, startTime, endTime int64) (*DescendantStatusMap, error) {
	avgDepLatency, err := s.promRepo.QueryMetricsWithPQLFilter(ctx,
		prom.WithDefaultForPolarisActiveSeries(prom.PQLAvgDepLatencyWithPQLFilter, prom.DefaultDepLatency),
		startTime, endTime,
		prom.EndpointGranularity,
		filter)
	if err != nil {
		return nil, err
	}

	avgLatency, err := s.promRepo.QueryMetricsWithPQLFilter(ctx,
		prom.PQLAvgLatencyWithPQLFilter,
		startTime, endTime,
		prom.EndpointGranularity,
		filter)
	if err != nil {
		return nil, err
	}

	avgLatencyDoD, err := s.promRepo.QueryMetricsWithPQLFilter(ctx,
		prom.DayOnDayTemplate(prom.PQLAvgLatencyWithPQLFilter),
		startTime, endTime,
		prom.EndpointGranularity,
		filter)
	if err != nil {
		return nil, err
	}

	avgErrorRateDoD, err := s.promRepo.QueryMetricsWithPQLFilter(ctx,
		prom.DayOnDayTemplate(prom.PQLAvgErrorRateWithPQLFilter),
		startTime, endTime,
		prom.EndpointGranularity,
		filter)
	if err != nil {
		return nil, err
	}
	avgRequestPerSecondDoD, err := s.promRepo.QueryMetricsWithPQLFilter(ctx,
		prom.DayOnDayTemplate(prom.PQLAvgTPSWithPQLFilter),
		startTime, endTime,
		prom.EndpointGranularity,
		filter)

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

func (s *DescendantStatus) SetValues(metricGroup prom.MGroupName, metricName prom.MName, points []prom.Points) {
	// Do nothing
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

	LatencyDoD          *float64 // Delay Day-over-Day Growth Rate
	ErrorRateDoD        *float64 // Error Rate Day-over-Day Growth Rate
	RequestPerSecondDoD *float64 // Request Day-over-Day Growth Rate
}

func fillServiceDelaySourceAndREDAlarm(descendantResp *response.GetDescendantRelevanceResponse, descendantStatus *DescendantStatusMap, threshold database.Threshold) {
	ts := time.Now()
	descendantKey := prom.EndpointKey{
		ContentKey: descendantResp.EndPoint,
		SvcName:    descendantResp.ServiceName,
	}

	if descendantStatus == nil {
		descendantResp.AlertReason.Add(model.REDMetricsAlert, model.AlertDetail{
			Timestamp:    ts.UnixMicro(),
			AlertObject:  descendantResp.ServiceName,
			AlertReason:  "时间段内未统计到应用延时,应用无请求或未监控,忽略RED告警;",
			AlertMessage: "",
		})
		return
	}

	status, find := descendantStatus.MetricGroupMap[descendantKey]
	if !find {
		descendantResp.AlertReason.Add(model.REDMetricsAlert, model.AlertDetail{
			Timestamp:    ts.UnixMicro(),
			AlertObject:  descendantResp.ServiceName,
			AlertReason:  "时间段内未统计到应用延时,应用无请求或未监控,忽略RED告警;",
			AlertMessage: "",
		})
		return
	}

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
