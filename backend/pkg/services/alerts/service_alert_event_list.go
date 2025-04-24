// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"encoding/json"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

func (s *service) AlertEventList(req *request.AlertEventSearchRequest) (*response.AlertEventSearchResponse, error) {
	events, count, err := s.chRepo.GetAlertEventWithWorkflowRecord(req, s.difyRepo.GetCacheMinutes())
	if err != nil {
		return nil, err
	}

	counts, err := s.chRepo.GetAlertEventCounts(req, s.difyRepo.GetCacheMinutes())
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(events); i++ {
		events[i].Alert.EnrichTags["source"] = events[i].Alert.Source

		s.fillWorkflowParams(&events[i])
		if events[i].IsValid == "unknown" && len(events[i].Output) > 0 {
			events[i].IsValid = "failed"
		} else if events[i].IsValid == "unknown" && events[i].Status == alert.StatusResolved {
			events[i].IsValid = "skipped"
		}
	}

	req.Pagination.Total = count
	return &response.AlertEventSearchResponse{
		EventList:                   events,
		Pagination:                  req.Pagination,
		AlertEventAnalyzeWorkflowID: s.difyRepo.GetAlertAnalyzeFlowID(),
		AlertCheckID:                s.difyRepo.GetAlertCheckFlowID(),
		Counts:                      counts,
	}, nil
}

func (s *service) fillWorkflowParams(record *alert.AEventWithWRecord) {
	var startTime, endTime time.Time
	if record.Status == alert.StatusResolved {
		startTime = record.EndTime.Add(-15 * time.Minute)
		endTime = record.EndTime
		record.Duration = record.EndTime.Sub(record.CreateTime).Round(time.Minute).String()
	} else {
		if record.Validity != "unknown" && record.Validity != "skipped" {
			startTime = record.LastCheckAt.Add(-15 * time.Minute)
		} else {
			startTime = record.UpdateTime.Add(-15 * time.Minute)
		}
		endTime = record.UpdateTime
		record.Duration = time.Since(record.CreateTime).Round(time.Minute).String()
	}

	record.WorkflowParams = alert.WorkflowParams{
		StartTime: startTime.UnixMicro(),
		EndTime:   endTime.UnixMicro(),
		NodeName:  record.AlertEvent.GetInfraNodeTag(),
	}

	alertServices, _ := tryGetAlertService(s.promRepo, &record.AlertEvent, startTime, endTime)

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
		Node:      record.AlertEvent.GetInfraNodeTag(),
		Namespace: record.AlertEvent.GetK8sNamespaceTag(),
		Pod:       record.AlertEvent.GetK8sPodTag(),
		Pid:       record.AlertEvent.GetPidTag(),
		AlertName: record.AlertEvent.Name,
	}

	if len(services) == 1 {
		parmas.Service = services[0]
		parmas.Endpoint = endpoints[0]
	}

	jsonStr, err := json.Marshal(parmas)
	if err != nil {
		record.WorkflowParams.Params = "{}"
	} else {
		record.WorkflowParams.Params = string(jsonStr)
	}
}

func tryGetAlertService(repo prometheus.Repo, event *alert.AlertEvent, startTime time.Time, endTime time.Time) ([]clickhouse.AlertService, error) {
	var tryMethods = []func(prometheus.Repo, *alert.AlertEvent, time.Time, time.Time) ([]clickhouse.AlertService, error){
		tryGetAlertServiceByService,
		tryGetAlertServiceByDB,
		tryGetAlertServiceByK8sPod,
		tryGetAlertServiceByVMProcess,
		tryGetAlertServiceByInfraNode,
	}
	var endpoints []clickhouse.AlertService
	for _, tryGetService := range tryMethods {
		var err error
		endpoints, err = tryGetService(repo, event, startTime, endTime)
		if err == nil && len(endpoints) > 0 {
			return endpoints, nil
		}
	}

	return endpoints, nil
}

func tryGetAlertServiceByService(_ prometheus.Repo, event *alert.AlertEvent, _ time.Time, _ time.Time) ([]clickhouse.AlertService, error) {
	serviceName := event.GetServiceNameTag()
	if len(serviceName) == 0 {
		return nil, nil
	}

	alertServices := []clickhouse.AlertService{
		{
			Service:  serviceName,
			Endpoint: event.GetEndpointTag(),
		},
	}

	return alertServices, nil
}

func tryGetAlertServiceByK8sPod(repo prometheus.Repo, event *alert.AlertEvent, startTime time.Time, endTime time.Time) ([]clickhouse.AlertService, error) {
	podName := event.GetNetSrcPodTag()
	namespace := event.GetK8sNamespaceTag()
	if len(podName) == 0 || len(namespace) == 0 {
		return nil, nil
	}

	// 通常也只会有一个Service
	services, err := repo.GetServiceListByFilter(
		startTime, endTime,
		prometheus.NamespacePQLFilter, namespace,
		prometheus.PodPQLFilter, podName,
	)
	if err != nil {
		return nil, err
	}
	var endpoints []clickhouse.AlertService
	// 通常只有一个service
	for _, service := range services {
		// 不关系ContentKey
		endpoints = append(endpoints, clickhouse.AlertService{
			Service: service,
		})
	}

	return endpoints, nil
}

func tryGetAlertServiceByVMProcess(repo prometheus.Repo, event *alert.AlertEvent, startTime time.Time, endTime time.Time) ([]clickhouse.AlertService, error) {
	nodeName := event.GetNetSrcNodeTag()
	pid := event.GetNetSrcPidTag()
	if len(nodeName) == 0 || len(pid) == 0 {
		return nil, nil
	}

	services, err := repo.GetServiceListByFilter(
		startTime, endTime,
		prometheus.NodeNamePQLFilter, nodeName,
		prometheus.PidPQLFilter, pid,
	)

	if err != nil {
		return nil, err
	}

	var endpoints []clickhouse.AlertService
	for _, service := range services {
		endpoints = append(endpoints, clickhouse.AlertService{
			Service: service,
		})
	}

	return endpoints, nil
}

func tryGetAlertServiceByInfraNode(repo prometheus.Repo, event *alert.AlertEvent, startTime time.Time, endTime time.Time) ([]clickhouse.AlertService, error) {
	if event.Group != string(clickhouse.INFRA_GROUP) {
		return nil, nil
	}

	nodeName := event.GetInfraNodeTag()
	if len(nodeName) == 0 {
		return nil, nil
	}

	services, err := repo.GetServiceListByFilter(
		startTime, endTime,
		prometheus.NodeNamePQLFilter, nodeName,
	)

	if err != nil {
		return nil, err
	}

	var endpoints []clickhouse.AlertService
	for _, service := range services {
		endpoints = append(endpoints, clickhouse.AlertService{
			Service: service,
		})
	}

	return endpoints, nil
}

func tryGetAlertServiceByDB(repo prometheus.Repo, event *alert.AlertEvent, startTime time.Time, endTime time.Time) ([]clickhouse.AlertService, error) {
	// 尝试获取数据库URL
	dbURL := event.GetDatabaseURL()
	dbIP := event.GetDatabaseIP()
	dbPort := event.GetDatabasePort()
	if len(dbURL) == 0 && len(dbIP) == 0 {
		return nil, nil
	}

	// 查询受此数据库影响的服务
	services, err := repo.GetServiceListByDatabase(
		startTime, endTime, dbURL, dbIP, dbPort)

	if err != nil {
		return nil, err
	}
	var endpoints []clickhouse.AlertService
	endpoints = append(endpoints, clickhouse.AlertService{
		DatabaseURL:  dbURL,
		DatabaseIP:   dbIP,
		DatabasePort: dbPort,
	})
	for _, service := range services {
		endpoints = append(endpoints, clickhouse.AlertService{
			Service: service,
		})
	}

	return endpoints, nil
}
