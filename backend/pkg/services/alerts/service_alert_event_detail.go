package alerts

import (
	"encoding/json"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) AlertDetail(req *request.GetAlertDetailRequest) (*response.GetAlertDetailResponse, error) {
	eventDetail, err := s.chRepo.GetAlertDetail(req, s.alertWorkflow.AlertCheck.CacheMinutes)
	if err != nil {
		// TODO
	}

	s.fillWorkflowParams(eventDetail)

	releatedEvents, total, err := s.chRepo.GetRelatedAlertEvents(req, s.alertWorkflow.AlertCheck.CacheMinutes)
	if err != nil {
		// TODO
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
		AlertEventAnalyzeWorkflowID: s.alertWorkflow.AnalyzeFlowId,
		AlertCheckID:                s.alertWorkflow.AlertCheck.FlowId,
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
			records[i].Duration = record.EndTime.Sub(records[i].CreateTime).Round(time.Minute).String()
		} else {
			startTime = records[i].UpdateTime.Add(-15 * time.Minute)
			endTime = records[i].UpdateTime

			duration := records[i].UpdateTime.Sub(records[i].CreateTime)
			duration += time.Minute
			records[i].Duration = duration.Round(time.Minute).String()
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
			Node:      records[i].AlertEvent.GetInfraNodeTag(),
			Namespace: records[i].AlertEvent.GetK8sNamespaceTag(),
			Pod:       records[i].AlertEvent.GetK8sPodTag(),
			Pid:       records[i].AlertEvent.GetPidTag(),
			AlertName: records[i].AlertEvent.Name,
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
