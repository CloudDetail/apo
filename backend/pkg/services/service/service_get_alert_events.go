package service

import (
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
)

func (s *service) GetAlertEventsSample(req *request.GetAlertEventsSampleRequest) (resp *response.GetAlertEventsSampleResponse, err error) {
	// 查询Service所属实例
	instances, err := s.promRepo.GetActiveInstanceList(req.StartTime, req.EndTime, req.Service)
	if err != nil || instances == nil {
		return nil, err
	}

	if req.SampleCount <= 0 {
		req.SampleCount = 1
	}
	startTime := time.UnixMicro(req.StartTime)
	endTime := time.UnixMicro(req.EndTime)

	// 查询实例的AlertEvent
	events, err := s.chRepo.GetAlertEventsSample(
		req.SampleCount,
		startTime, endTime,
		req.AlertFilter,
		instances.GetInstances(),
	)
	if err != nil {
		return nil, err
	}

	groupedEvents := splitByGroupAndName(events)

	var status = model.STATUS_NORMAL
	if len(groupedEvents) > 0 {
		status = model.STATUS_CRITICAL
	}
	return &response.GetAlertEventsSampleResponse{
		EventMap: groupedEvents,
		Status:   status,
	}, nil
}

func (s *service) GetAlertEvents(req *request.GetAlertEventsRequest) (*response.GetAlertEventsResponse, error) {
	// 查询Service所属实例
	instances, err := s.promRepo.GetActiveInstanceList(req.StartTime, req.EndTime, req.Service)
	if err != nil {
		return nil, err
	}

	startTime := time.UnixMicro(req.StartTime)
	endTime := time.UnixMicro(req.EndTime)

	// 查询实例的AlertEvent
	events, totalCount, err := s.chRepo.GetAlertEvents(
		startTime, endTime,
		req.AlertFilter,
		instances.GetInstances(),
		req.PageParam,
	)
	if err != nil {
		return nil, err
	}

	// HACK 直接以列表形式返回数据
	return &response.GetAlertEventsResponse{
		TotalCount: totalCount,
		EventList:  events,
	}, nil
}

/*
splitByGroupAndName 将结果按Group和name分组

	app:
		alert1: [item1,item2...]
		alert2: [...]
	network:
	...
*/
func splitByGroupAndName(events []clickhouse.AlertEventSample) map[string]map[string][]clickhouse.AlertEventSample {
	var from int = 0
	var lastGroup, lastName string

	var res = make(map[string]map[string][]clickhouse.AlertEventSample)
	for i := 0; i < len(events); i++ {
		event := events[i]

		if lastGroup == event.Group && lastName == event.Name {
			continue
		}
		if lastGroup != event.Group {
			if i > 0 {
				res[lastGroup][lastName] = events[from:i]
			}
			lastGroup = event.Group
			lastName = event.Name
			res[lastGroup] = make(map[string][]clickhouse.AlertEventSample)
		} else if lastName != event.Name {
			res[lastGroup][lastName] = events[from:i]
			lastName = event.Name
		}
		from = i
	}

	if len(events) > 0 {
		res[lastGroup][lastName] = events[from:]
	}

	return res
}
