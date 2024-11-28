package alertanalyze

import (
	"strings"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	ck "github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
)

func (s *service) GetAnormalTrendByEntry(req *request.GetDescendantAnormalTrendRequest) (*response.GetDescendantDeltaAnormalEventResponse, error) {
	startTime := time.UnixMicro(req.StartTime)
	endTime := time.UnixMicro(req.EndTime)

	descendants, err := s.chRepo.ListDescendantNodes(&request.GetDescendantMetricsRequest{
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Service:   req.Service,
		Endpoint:  req.Endpoint,
	})
	if err != nil {
		return nil, err
	}

	descendants = append(descendants, ck.ServiceNode{
		Service:  req.Service,
		Endpoint: req.Endpoint,
	})

	var selectedEvents []string
	if len(req.AnormalTypes) > 0 {
		selectedEvents = strings.Split(req.AnormalTypes, ",")
	}

	// 便于后续查询受告警影响的服务
	var instances []*model.ServiceInstance
	var endpoints []model.EndpointKey
	instanceMap := newInstanceMap()
	// TODO 优化,减少查询次数

	for _, descendant := range descendants {
		endpoint := model.EndpointKey{
			ServiceName: descendant.Service,
			Endpoint:    descendant.Endpoint,
		}
		instanceMap.AddEndpoint(endpoint)
		// TODO 优化,减少查询次数,获取每个endpoint下的所有实例
		instanceList, err := s.promRepo.GetActiveRequestInstanceList(req.StartTime, req.EndTime, descendant.Service, descendant.Endpoint)
		if err != nil {
			continue
		}

		instancesForDescendant := instanceList.GetInstances()
		instanceMap.AddInstances(endpoint, instancesForDescendant)

		instances = append(instances, instancesForDescendant...)
		endpoints = append(endpoints, endpoint)
	}

	// 返回结果列表
	var anormalEventList []model.AnormalEvent

	// 获取匹配的error
	if len(selectedEvents) == 0 || isSelect(selectedEvents, errorEvent) {
		propagations, err := s.chRepo.ListErrorByEntryService(req.StartTime, req.EndTime, req.Service, req.Endpoint, endpoints)
		if err == nil {
			errorEvents := s.parseErrorEvent(propagations, instanceMap, req.Step)
			// 存放error事件
			anormalEventList = append(anormalEventList, errorEvents...)
		} else {
			return nil, err
		}
	}

	if len(selectedEvents) == 0 || isSelect(selectedEvents, alertEvent) {
		// 获取匹配的alertEvents
		alertEvents, err := s.chRepo.GetAlertEventsWithKeyByInstanceAndEndpoints(
			startTime, endTime,
			// 同时获取Firing和Resolved故障
			request.AlertFilter{WithMutation: true},
			instances, endpoints,
		)

		if err == nil {
			anormalAlerts := s.parseAlertEvents(alertEvents, instanceMap, selectedEvents, req.StartTime)
			// 存放告警事件
			anormalEventList = append(anormalEventList, anormalAlerts...)
		} else {
			return nil, err
		}
	}

	// 按Step分组并生成Chart, 统计每个Step时间段内未解决的异常数量
	anormalCount := response.TempChartObject{
		ChartData: map[int64]float64{},
	}
	// 初始化
	for i := req.StartTime; i <= req.EndTime; i += req.Step {
		anormalCount.ChartData[i] = 0
	}

	// anormalEventTimeGroup := make(map[int64][]model.AnormalEvent)
	for _, event := range anormalEventList {
		// 填充Chart
		startFiring := int64(-1)
		for updateIdx, updateTS := range event.UpdateTSs {
			if updateIdx == len(event.UpdateTSs)-1 && updateTS.AnormalStatus == model.StatusFiring {
				if startFiring == -1 {
					startFiring = updateTS.Timestamp
				}
				for ts, count := range anormalCount.ChartData {
					if ts >= startFiring {
						anormalCount.ChartData[ts] = count + float64(len(event.ImpactEndpoints))
					}
				}
			} else if startFiring == -1 && updateTS.AnormalStatus == model.StatusFiring {
				// 新的一次告警
				startFiring = updateTS.Timestamp
			} else if startFiring != -1 && updateTS.AnormalStatus == model.StatusResolved {
				// 结算之前的告警
				for ts, count := range anormalCount.ChartData {
					if ts >= startFiring && ts < updateTS.Timestamp {
						anormalCount.ChartData[ts] = count + float64(len(event.ImpactEndpoints))
					}
				}
				startFiring = -1
			}
		}
	}
	return &response.GetDescendantDeltaAnormalEventResponse{
		AnormalCount: anormalCount,
	}, nil
}
