// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

func (s *service) GetAlertEventsSample(ctx core.Context, req *request.GetAlertEventsSampleRequest) (resp *response.GetAlertEventsSampleResponse, err error) {
	// Query the instance to which the Service belongs.
	filter := prometheus.NewFilter()
	filter.RegexMatch(prometheus.ServiceNameKey, prometheus.RegexMultipleValue(req.Services...))
	if len(req.ClusterIDs) > 0 {
		filter.RegexMatch("cluster_id", prometheus.RegexMultipleValue(req.ClusterIDs...))
	}

	instances, err := s.promRepo.GetInstanceListByPQLFilter(ctx, req.StartTime, req.EndTime, filter)
	if err != nil || instances == nil {
		return nil, err
	}

	if req.SampleCount <= 0 {
		req.SampleCount = 1
	}
	startTime := time.UnixMicro(req.StartTime)
	endTime := time.UnixMicro(req.EndTime)

	var dbInstances []model.MiddlewareInstance
	if len(req.AlertFilter.Group) == 0 || req.AlertFilter.Group == "middleware" {
		dbInstances, err = s.promRepo.GetDescendantDatabase(ctx, req.StartTime, req.EndTime, filter)
		if err != nil {
			return nil, err
		}
	}

	// Query the AlertEvent of the instance
	events, err := s.chRepo.GetAlertEventsSample(
		ctx,
		req.SampleCount,
		startTime, endTime,
		req.AlertFilter,
		&model.RelatedInstances{
			SIs: instances.GetInstances(),
			MIs: dbInstances,
		},
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

func (s *service) GetAlertEvents(ctx core.Context, req *request.GetAlertEventsRequest) (*response.GetAlertEventsResponse, error) {
	// Query the instance to which the Service belongs.
	filter := prometheus.NewFilter()
	filter.RegexMatch(prometheus.ServiceNameKey, prometheus.RegexMultipleValue(req.Services...))
	if len(req.ClusterIDs) > 0 {
		filter.RegexMatch("cluster_id", prometheus.RegexMultipleValue(req.ClusterIDs...))
	}

	instances, err := s.promRepo.GetInstanceListByPQLFilter(ctx, req.StartTime, req.EndTime, filter)
	if err != nil {
		return nil, err
	}

	startTime := time.UnixMicro(req.StartTime)
	endTime := time.UnixMicro(req.EndTime)

	var dbInstances []model.MiddlewareInstance
	if len(req.AlertFilter.Group) == 0 || req.AlertFilter.Group == "middleware" {
		dbInstances, err = s.promRepo.GetDescendantDatabase(ctx, req.StartTime, req.EndTime, filter)
		if err != nil {
			return nil, err
		}
	}

	// Query the AlertEvent of the instance
	events, totalCount, err := s.chRepo.GetAlertEvents(
		ctx,
		startTime, endTime,
		req.AlertFilter,
		&model.RelatedInstances{
			SIs: instances.GetInstances(),
			MIs: dbInstances,
		},
		req.PageParam,
	)
	if err != nil {
		return nil, err
	}

	// HACK returns data directly as a list
	return &response.GetAlertEventsResponse{
		TotalCount: totalCount,
		EventList:  events,
	}, nil
}

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
