package alertanalyze

import (
	"sort"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	ck "github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

type GetAlertsImpactResp struct {
	Data []*response.AlertRelatedEntry `json:"data"`
}

func (s *service) GetAlertsImpact(req *request.AlertImpactRequest) ([]response.AlertRelatedEntry, error) {
	startTime := time.UnixMicro(req.StartTime)
	endTime := time.UnixMicro(req.EndTime)

	// 获取Alert列表,并统计之间段内告警(Firing)的次数
	alerts, total, err := s.chRepo.GetAlertsWithEventCount(
		startTime, endTime, nil, 5000,
	)
	if err != nil {
		return nil, err
	}

	// 获取Alert相关的所有服务;
	// 服务和其相关的告警事件下标列表
	var service2AlertMap = map[ck.AlertService][]int{}

	endpointsMap := newEndpointsMap()
	for i := 0; i < len(alerts); i++ {
		endpoints, associated := endpointsMap.SearchAndAddEndpointsForAlert(
			s.promRepo, &alerts[i].Alert, startTime, endTime)
		if !associated || len(endpoints) == 0 {
			continue
		}

		for _, service := range endpoints {
			service2AlertMap[service] = append(service2AlertMap[service], i)
		}
	}

	endpoints := endpointsMap.EndpointsList()
	// 查询入口和下游服务节点的关联
	entryRelation, err := s.chRepo.SearchEntryEndpointsByAlertService(endpoints, startTime.UnixMicro(), endTime.UnixMicro())
	if err != nil {
		return nil, err
	}
	// entryKey -> map[alertIdx]{}
	entry2AlertsMap := connectEntryAndAlerts(entryRelation, service2AlertMap, req)

	var entryEndpoints []response.AlertRelatedEntry
	for entry, alertIdx := range entry2AlertsMap {
		var count int64 = 0
		for idx := range alertIdx {
			count += int64(alerts[idx].Count)
		}

		entryEndpoints = append(entryEndpoints, response.AlertRelatedEntry{
			ServiceName: entry.SvcName,
			ServiceDetail: response.ServiceDetail{
				Endpoint: entry.ContentKey,
			},
			RelatedAlertRate: float64(count) * 100 / float64(total),
		})
	}

	sort.Slice(entryEndpoints, func(i, j int) bool {
		return entryEndpoints[i].RelatedAlertRate > entryEndpoints[j].RelatedAlertRate
	})

	return entryEndpoints, nil
}

func connectEntryAndAlerts(
	entryRelation []ck.EntryRelationship,
	service2AlertMap map[ck.AlertService][]int,
	req *request.AlertImpactRequest,
) map[prometheus.EndpointKey]map[int]struct{} {

	filter := newFilter(req.Services, req.Endpoints)

	endpointsSet := map[prometheus.EndpointKey]map[int]struct{}{}
	for _, relation := range entryRelation {
		svc := relation.GetEntryService()
		endpoint := relation.GetEntryEndpoint()

		if filter.IsFiltered(svc, endpoint) {
			continue
		}

		entryEndpointKey := prometheus.EndpointKey{
			SvcName:    svc,
			ContentKey: endpoint,
		}

		relatedAlerts, find := endpointsSet[entryEndpointKey]
		if !find {
			relatedAlerts = make(map[int]struct{}, 0)
		}

		svcKey := ck.AlertService{
			Service:  relation.Service,
			Endpoint: relation.Endpoint,
		}

		if alertIdx, find := service2AlertMap[svcKey]; find {
			for _, idx := range alertIdx {
				relatedAlerts[idx] = struct{}{}
			}
		}

		svcKey.Endpoint = ""
		if alertIdx, find := service2AlertMap[svcKey]; find {
			for _, idx := range alertIdx {
				relatedAlerts[idx] = struct{}{}
			}
		}

		endpointsSet[entryEndpointKey] = relatedAlerts
	}
	return endpointsSet
}

type entryFilter struct {
	Services  map[string]struct{}
	Endpoints map[string]struct{}
}

func newFilter(services, endpoints []string) *entryFilter {
	filter := &entryFilter{
		Services:  map[string]struct{}{},
		Endpoints: map[string]struct{}{},
	}
	for _, service := range services {
		filter.Services[service] = struct{}{}
	}

	for _, endpoint := range endpoints {
		filter.Endpoints[endpoint] = struct{}{}
	}
	return filter
}

func (e *entryFilter) IsFiltered(svc, endpoint string) bool {
	if len(e.Services) > 0 {
		if _, find := e.Services[svc]; !find {
			return true
		}
	}

	if len(e.Endpoints) > 0 {
		if _, find := e.Services[endpoint]; !find {
			return true
		}
	}
	return false
}
