// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"encoding/json"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) AlertDetail(req *request.GetAlertDetailRequest) (*response.GetAlertDetailResponse, error) {
	eventDetail, err := s.chRepo.GetAlertDetail(req, s.difyRepo.GetCacheMinutes())
	if err != nil {
		return nil, err
	}

	if req.StartTime <= 0 || req.EndTime <= 0 {
		req.StartTime = eventDetail.UpdateTime.Add(-2 * time.Hour).UnixMicro()
		// req.EndTime = eventDetail.UpdateTime.Add(15 * time.Minute).UnixMicro()
		req.EndTime = time.Now().UnixMicro()
	}

	s.fillWorkflowParams(eventDetail)

	releatedEvents, total, err := s.chRepo.GetRelatedAlertEvents(req, s.difyRepo.GetCacheMinutes())
	if err != nil {
		return nil, err
	}

	req.Pagination.Total = total
	var localIndex int
	for idx, event := range releatedEvents {
		if event.ID.String() == req.EventID {
			localIndex = idx
		}
	}

	s.fillSimilarEventWorkflowParams(releatedEvents)

	return &response.GetAlertDetailResponse{
		CurrentEvent:                eventDetail,
		EventList:                   releatedEvents,
		Pagination:                  req.Pagination,
		AlertEventAnalyzeWorkflowID: s.difyRepo.GetAlertAnalyzeFlowID(),
		AlertCheckID:                s.difyRepo.GetAlertCheckFlowID(),
		LocateIdx:                   localIndex,
	}, nil
}

func (s *service) fillSimilarEventWorkflowParams(records []alert.AEventWithWRecord) {
	if len(records) == 0 {
		return
	}

	record := records[0]
	var startTime, endTime time.Time
	if record.AlertEvent.Status == alert.StatusResolved {
		startTime = record.EndTime.Add(-15 * time.Minute)
		endTime = record.EndTime
	} else {
		startTime = record.UpdateTime.Add(-15 * time.Minute)
		endTime = record.UpdateTime
	}
	alertServices, _ := tryGetAlertService(s.promRepo, &record.AlertEvent, startTime, endTime)

	for i := 0; i < len(records); i++ {
		records[i].Alert.EnrichTags["source"] = records[i].Alert.Source
		var startTime, endTime time.Time
		if records[i].AlertEvent.Status == alert.StatusResolved {
			startTime = records[i].EndTime.Add(-15 * time.Minute)
			endTime = records[i].EndTime
			records[i].Duration = formatDuration(record.EndTime.Sub(records[i].CreateTime))
		} else {
			startTime = records[i].UpdateTime.Add(-15 * time.Minute)
			endTime = records[i].UpdateTime

			duration := records[i].UpdateTime.Sub(records[i].CreateTime)
			duration += time.Minute
			records[i].Duration = formatDuration(duration)
		}
		records[i].WorkflowParams = alert.WorkflowParams{
			StartTime: startTime.UnixMicro(),
			EndTime:   endTime.UnixMicro(),
			NodeName:  records[i].AlertEvent.GetInfraNodeTag(),
		}

		var services, endpoints []string
		for _, alertService := range alertServices {
			services = append(services, alertService.Service)
			if len(alertService.Endpoint) == 0 {
				endpoints = append(endpoints, ".*")
			} else {
				endpoints = append(endpoints, alertService.Endpoint)
			}
		}

		parmas := alert.AlertAnalyzeWorkflowParams{
			AlertName:   records[i].AlertEvent.Name,
			Node:        records[i].AlertEvent.GetInfraNodeTag(),
			Namespace:   records[i].AlertEvent.GetK8sNamespaceTag(),
			Pod:         records[i].AlertEvent.GetK8sPodTag(),
			Pid:         records[i].AlertEvent.GetPidTag(),
			Detail:      records[i].AlertEvent.Detail,
			ContainerID: records[i].AlertEvent.GetContainerIDTag(),
			Tags:        records[i].AlertEvent.EnrichTags,
			RawTags:     records[i].AlertEvent.Tags,
		}

		if len(services) == 1 {
			parmas.Service = services[0]
			parmas.Endpoint = endpoints[0]
		}

		jsonStr, err := json.Marshal(parmas)
		if err != nil {
			records[i].WorkflowParams.Params = "{}"
		} else {
			records[i].WorkflowParams.Params = string(jsonStr)
		}
	}
}

func (s *service) ManualResolveLatestAlertEventByAlertID(alertID string) error {
	// TODO valid alertID
	return s.chRepo.ManualResolveLatestAlertEventByAlertID(alertID)
}
