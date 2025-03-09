// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

func (s *service) AlertEventList(req *request.AlertEventSearchRequest) (*response.AlertEventSearchResponse, error) {
	events, count, err := s.chRepo.GetAlertEventWithWorkflowRecord(req)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(events); i++ {
		s.fillWorkflowParmas(&events[i])
	}

	req.Pagination.Total = count
	return &response.AlertEventSearchResponse{
		EventList:                   events,
		Pagination:                  req.Pagination,
		AlertEventAnalyzeWorkflowID: s.alertWorkflow.EventAnalyzeFlowId,
	}, nil
}

func (s *service) fillWorkflowParmas(record *alert.AEventWithWRecord) {

	startTime := record.ReceivedTime.Add(-15 * time.Minute)
	endTime := record.ReceivedTime

	record.WorkflowParams = alert.WorkflowParams{
		StartTime: startTime.UnixMicro(),
		EndTime:   endTime.UnixMicro(),
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

	if len(services) > 0 {
		record.WorkflowParams.Service = strings.Join(services, "|")
		record.WorkflowParams.Endpoint = strings.Join(endpoints, "|")
	}

	jsonStr, err := json.Marshal(alert.AlertAnalyzeWorkflowParams{
		Node:      record.AlertEvent.GetInfraNodeTag(),
		Namespace: record.AlertEvent.GetK8sNamespaceTag(),
		Pod:       record.AlertEvent.GetK8sPodTag(),
	})
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
