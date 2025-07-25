// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/code"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/common"
	"github.com/google/uuid"
)

func (s *service) AlertEventList(ctx core.Context, req *request.AlertEventSearchRequest) (*response.AlertEventSearchResponse, error) {
	if req.GroupID > 0 {
		subGroupIDs := common.DataGroupStorage.GetFullSubGroupIDs(req.GroupID)
		for _, id := range subGroupIDs {
			req.SubGroupIDs = append(req.SubGroupIDs, strconv.FormatInt(id, 10))
		}
		req.SubGroupIDs = append(req.SubGroupIDs, strconv.FormatInt(req.GroupID, 10))
	}

	events, count, err := s.chRepo.GetAlertEventWithWorkflowRecord(ctx, req, s.difyRepo.GetCacheMinutes())
	if err != nil {
		return nil, err
	}

	counts, err := s.chRepo.GetAlertEventCounts(ctx, req, s.difyRepo.GetCacheMinutes())
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(events); i++ {
		events[i].Alert.EnrichTags["source"] = events[i].Alert.Source

		s.fillWorkflowParams(ctx, &events[i])
		if events[i].IsValid == "unknown" && len(events[i].Output) > 0 {
			events[i].IsValid = "failed"
		} else if events[i].IsValid == "unknown" && events[i].Status == alert.StatusResolved {
			events[i].IsValid = "skipped"
		}
	}

	// TODO show display error log
	_ = s.fillDisplays(ctx, events)

	req.Pagination.Total = count
	return &response.AlertEventSearchResponse{
		EventList:                   events,
		Pagination:                  req.Pagination,
		AlertEventAnalyzeWorkflowID: s.difyRepo.GetAlertAnalyzeFlowID(),
		AlertCheckID:                s.difyRepo.GetAlertCheckFlowID(),
		Counts:                      counts,
	}, nil
}

func (s *service) fillDisplays(ctx core.Context, records []alert.AEventWithWRecord) error {
	tags, err := s.dbRepo.ListAlertTargetTags(ctx)
	if err != nil {
		return err
	}

	lang := ctx.LANG()
	for i := 0; i < len(records); i++ {
		tagDisplays := make([]alert.TagDisplay, 0)
		for key, value := range records[i].EnrichTags {
			tagName := getTagName(tags, key, lang)
			tagDisplays = append(tagDisplays, alert.TagDisplay{
				Key:   key,
				Name:  tagName,
				Value: value,
			})
		}
		sort.Slice(tagDisplays, func(i, j int) bool {
			return tagDisplays[i].Key < tagDisplays[j].Key
		})
		records[i].EnrichTagsDisplay = tagDisplays
	}
	return nil
}

func (s *service) fillDisplay(ctx core.Context, record *alert.AEventWithWRecord) error {
	tags, err := s.dbRepo.ListAlertTargetTags(ctx)
	if err != nil {
		return err
	}

	lang := ctx.LANG()
	tagDisplays := make([]alert.TagDisplay, 0)
	for key, value := range record.EnrichTags {
		tagName := getTagName(tags, key, lang)
		tagDisplays = append(tagDisplays, alert.TagDisplay{
			Key:   key,
			Name:  tagName,
			Value: value,
		})
	}
	sort.Slice(tagDisplays, func(i, j int) bool {
		return tagDisplays[i].Key < tagDisplays[j].Key
	})
	record.EnrichTagsDisplay = tagDisplays
	return nil
}

func getTagName(tags []alert.TargetTag, key string, lang string) string {
	if key == "source" {
		if lang == code.LANG_EN {
			return "Alert Source"
		} else {
			return "告警源"
		}
	} else if key == "status" {
		if lang == code.LANG_EN {
			return "Status"
		} else {
			return "告警状态"
		}
	}
	for _, tag := range tags {
		if key == tag.Field {
			return tag.TagName
		}
	}
	return key
}

func (s *service) fillWorkflowParams(ctx core.Context, record *alert.AEventWithWRecord) {
	var startTime, endTime time.Time
	if record.Status == alert.StatusResolved {
		startTime = record.EndTime.Add(-15 * time.Minute)
		endTime = record.EndTime
		record.Duration = formatDuration(record.EndTime.Sub(record.CreateTime))
	} else {
		if record.Validity != "unknown" && record.Validity != "skipped" {
			startTime = record.LastCheckAt.Add(-15 * time.Minute)
			endTime = record.LastCheckAt
		} else {
			startTime = record.UpdateTime.Add(-15 * time.Minute)
			endTime = record.UpdateTime
		}
		record.Duration = formatDuration(time.Since(record.CreateTime))
	}

	if len(record.Input) > 0 {
		// replace EventID by checkedEvent
		record.ID, _ = uuid.Parse(record.Input)
	}

	record.WorkflowParams = alert.WorkflowParams{
		StartTime: startTime.UnixMicro(),
		EndTime:   endTime.UnixMicro(),
		NodeName:  record.AlertEvent.GetInfraNodeTag(),
		Edition:   "ce",
	}

	alertServices, _ := tryGetAlertService(ctx, s.promRepo, &record.AlertEvent, startTime, endTime)

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
		AlertName:    record.AlertEvent.Name,
		Node:         record.AlertEvent.GetInfraNodeTag(),
		Namespace:    record.AlertEvent.GetK8sNamespaceTag(),
		Pod:          record.AlertEvent.GetK8sPodTag(),
		Pid:          record.AlertEvent.GetPidTag(),
		Detail:       record.Detail,
		ContainerID:  record.AlertEvent.GetContainerIDTag(),
		Tags:         record.Alert.EnrichTags,
		RawTags:      record.Alert.Tags,
		AlertEventId: record.Input,
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

func formatDuration(d time.Duration) string {
	day := 0
	hour := int(d.Hours())

	if hour > 24 {
		day = hour / 24
		hour = hour % 24
	}

	minute := int(d.Minutes()) % 60

	if day > 0 {
		return fmt.Sprintf("%dd %02dh %02dm", day, hour, minute)
	} else if hour > 0 {
		return fmt.Sprintf("%dh %02dm", hour, minute)
	} else {
		return fmt.Sprintf("%dm", minute)
	}
}
