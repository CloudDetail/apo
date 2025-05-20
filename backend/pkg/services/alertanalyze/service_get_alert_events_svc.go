package alertanalyze

import (
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	ck "github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"go.uber.org/zap"
)

var eventFilter = request.AlertFilter{
	Status:       "firing",
	WithMutation: true,
}

func (s *service) GetAlertEventsSvc(ctx core.Context, req *request.GetAlertEventsSvcReq) (*response.GetAlertEventsSvcResp, error) {
	startTime := time.UnixMicro(req.StartTime)
	endTime := time.UnixMicro(req.EndTime)

	if req.PageParam == nil {
		req.PageParam = &request.PageParam{
			CurrentPage: 1,
			PageSize:    999,
		}
	}

	events, count, err := s.chRepo.GetAlertEvents(
		ctx,
		startTime, endTime, eventFilter, nil,
		req.PageParam, ck.OrderAlertByReceivedTime)
	if err != nil {
		return nil, err
	}

	endpointsMap := newEndpointsMap()
	var impactEvents = make([]response.ImpactAlertEvent, len(events))
	for i := 0; i < len(events); i++ {
		endpoints, associated := endpointsMap.SearchAndAddEndpointsForAlert(s.promRepo, &events[i].Alert, startTime, endTime)
		impactEvents[i] = response.ImpactAlertEvent{
			AlertEvent:      &events[i],
			IsAnalyzed:      associated,
			ImpactEndpoints: endpoints,
			ImpactEntries:   []model.EndpointKey{},
		}
	}

	endpoints := endpointsMap.EndpointsList()
	entryRelations, err := s.chRepo.SearchEntryEndpointsByAlertService(ctx,
		endpoints, startTime.UnixMicro(), endTime.UnixMicro())

	if err != nil {
		s.logger.Error("failed to search entryRelations for alertEvents, err", zap.Error(err))
	}

	for i := 0; i < len(impactEvents); i++ {
		impactEvents[i].Alert.EnrichTags["source"] = impactEvents[i].Alert.Source
		impactEvents[i].Tags["source"] = impactEvents[i].Alert.Source
		impactEvents[i].ImpactEntries = searchEntryByServices(entryRelations, impactEvents[i].ImpactEndpoints)
	}

	return &response.GetAlertEventsSvcResp{
		Events: impactEvents,
		Pagination: &model.Pagination{
			Total:       int64(count),
			CurrentPage: req.CurrentPage,
			PageSize:    req.PageSize,
		},
	}, nil
}
