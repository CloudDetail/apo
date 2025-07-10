// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dify

import (
	"time"

	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

func tryGetAlertService(ctx core.Context, repo prometheus.Repo, event *alert.AlertEvent, startTime time.Time, endTime time.Time) ([]clickhouse.AlertService, error) {
	var tryMethods = []func(core.Context, prometheus.Repo, *alert.AlertEvent, time.Time, time.Time) ([]clickhouse.AlertService, error){
		tryGetAlertServiceByService,
		tryGetAlertServiceByDB,
		tryGetAlertServiceByK8sPod,
		tryGetAlertServiceByVMProcess,
		tryGetAlertServiceByInfraNode,
	}
	var endpoints []clickhouse.AlertService
	for _, tryGetService := range tryMethods {
		var err error
		endpoints, err = tryGetService(ctx, repo, event, startTime, endTime)
		if err == nil && len(endpoints) > 0 {
			return endpoints, nil
		}
	}

	return endpoints, nil
}

func tryGetAlertServiceByService(ctx core.Context, _ prometheus.Repo, event *alert.AlertEvent, _ time.Time, _ time.Time) ([]clickhouse.AlertService, error) {
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

func tryGetAlertServiceByK8sPod(ctx core.Context, repo prometheus.Repo, event *alert.AlertEvent, startTime time.Time, endTime time.Time) ([]clickhouse.AlertService, error) {
	podName := event.GetNetSrcPodTag()
	namespace := event.GetK8sNamespaceTag()
	if len(podName) == 0 || len(namespace) == 0 {
		return nil, nil
	}

	// 通常也只会有一个Service
	services, err := repo.GetServiceListByFilter(
		ctx,
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

func tryGetAlertServiceByVMProcess(ctx core.Context, repo prometheus.Repo, event *alert.AlertEvent, startTime time.Time, endTime time.Time) ([]clickhouse.AlertService, error) {
	nodeName := event.GetNetSrcNodeTag()
	pid := event.GetNetSrcPidTag()
	if len(nodeName) == 0 || len(pid) == 0 {
		return nil, nil
	}

	services, err := repo.GetServiceListByFilter(
		ctx,
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

func tryGetAlertServiceByInfraNode(ctx core.Context, repo prometheus.Repo, event *alert.AlertEvent, startTime time.Time, endTime time.Time) ([]clickhouse.AlertService, error) {
	if event.Group != string(clickhouse.INFRA_GROUP) {
		return nil, nil
	}

	nodeName := event.GetInfraNodeTag()
	if len(nodeName) == 0 {
		return nil, nil
	}

	services, err := repo.GetServiceListByFilter(
		ctx,
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

func tryGetAlertServiceByDB(ctx core.Context, repo prometheus.Repo, event *alert.AlertEvent, startTime time.Time, endTime time.Time) ([]clickhouse.AlertService, error) {
	// 尝试获取数据库URL
	dbURL := event.GetDatabaseURL()
	dbIP := event.GetDatabaseIP()
	dbPort := event.GetDatabasePort()
	if len(dbURL) == 0 && len(dbIP) == 0 {
		return nil, nil
	}

	// 查询受此数据库影响的服务
	services, err := repo.GetServiceListByDatabase(
		ctx,
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
