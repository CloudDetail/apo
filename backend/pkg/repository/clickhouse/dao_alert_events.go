// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/google/uuid"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/datagroup"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

type AlertGroup string

const (
	APP_GROUP          AlertGroup = "app"
	NETWORK_GROUP      AlertGroup = "network"
	CONTAINER_GROUP    AlertGroup = "container"
	INFRA_GROUP        AlertGroup = "infra"
	MIDDLEWARE_GROUP   AlertGroup = "middleware"
	APP_INSTANCE_GROUP AlertGroup = "app_instance"
)

const (
	MUTATION_APP_GROUP          AlertGroup = "mutation-app"
	MUTATION_NETWORK_GROUP      AlertGroup = "mutation-network"
	MUTATION_CONTAINER_GROUP    AlertGroup = "mutation-container"
	MUTATION_INFRA_GROUP        AlertGroup = "mutation-infra"
	MUTATION_MIDDLEWARE_GROUP   AlertGroup = "mutation-middleware"
	MUTATION_APP_INSTANCE_GROUP AlertGroup = "mutation-app_instance"
)

func (g AlertGroup) GetAlertType() string {
	switch g {
	case INFRA_GROUP:
		return model.InfrastructureAlert
	case NETWORK_GROUP:
		return model.NetAlert
	case APP_GROUP:
		return model.AppAlert
	case APP_INSTANCE_GROUP:
		return model.AppAlert
	case CONTAINER_GROUP:
		return model.ContainerAlert
	}
	return model.UndefinedAlert
}

func GetAlertType(g string) string {
	group := AlertGroup(g)
	return group.GetAlertType()
}

const (
	// The SQL _GET_SAMPLE_ALERT_EVENT are grouped by the alarm_event name. Each group takes the record with the latest event and records the number of alarms with the same name in the returned result.
	SQL_GET_SAMPLE_ALERT_EVENT = `WITH grouped_alarm AS (
		SELECT source,group,id,create_time,update_time,end_time,received_time,severity,name,detail,tags,raw_tags,status,
        	if(alert_id != '', alert_id, arrayStringConcat(arrayMap(x -> x.2, arraySort(arrayZip(mapKeys(tags), mapValues(tags)))), ', ')) AS alert_key,
			ROW_NUMBER() OVER (PARTITION BY name, alert_key ORDER BY received_time desc) AS rn,
			COUNT(*) OVER (PARTITION BY name, alert_key) AS alarm_count
    	FROM alert_event
		%s
	)
	SELECT *
	FROM grouped_alarm
	WHERE rn <= %d %s`

	SQL_GET_ALERT_EVENT_DATA_SCOPE = `SELECT
		DISTINCT tags['cluster_id'] as cluster_id,
		tags['namespace'] as namespace,
		tags['serviceName'] as service
	FROM alert_event
	%s`

	SQL_GET_GROUP_COUNTS_ALERT_EVENT = `WITH grouped_alarm AS (
	SELECT group,severity,tags,
		ROW_NUMBER() OVER (PARTITION BY %s) AS rn,
		COUNT(*) OVER (PARTITION BY %s) AS alarm_count
	FROM alert_event
	%s
	)
	SELECT *
	FROM grouped_alarm
	WHERE rn <= 1`

	SQL_GET_ALERT_EVENT = `SELECT source,group,id,create_time,update_time,end_time,received_time,severity,name,detail,tags,raw_tags,status
	FROM alert_event %s %s`

	SQL_GET_ALERT_EVENT_WITH_LABEL_KEY = `SELECT source,group,id,create_time,update_time,end_time,received_time,severity,name,detail,tags,raw_tags,status
        ,if(alert_id != '', alert_id, arrayStringConcat(arrayMap(x -> x.2, arraySort(arrayZip(mapKeys(tags), mapValues(tags)))), ', ')) AS alert_key
    FROM alert_event
	%s %s`

	GET_ALERT_EVENTS_COUNT = `SELECT count(1) as count FROM alert_event %s`

	SQL_GET_PAGED_ALERT_EVENT = `SELECT
		source,group,id,create_time,update_time,end_time,received_time,severity,name,detail,tags,raw_tags,status,alert_id
	FROM alert_event
	%s %s`
)

// GetAlertEventCountGroupByInstance to quickly query the number of alarms associated with each Instance (counted separately by alarm level)
func (ch *chRepo) GetAlertEventCountGroupByInstance(ctx core.Context, startTime time.Time, endTime time.Time, filter request.AlertFilter, instances *model.RelatedInstances) ([]model.AlertEventCount, error) {
	builder := NewQueryBuilder().
		Between("received_time", startTime.Unix(), endTime.Unix()).
		EqualsNotEmpty("source", filter.Source).
		EqualsNotEmpty("group", filter.Group).
		EqualsNotEmpty("name", filter.Name).
		EqualsNotEmpty("id", filter.ID).
		EqualsNotEmpty("severity", filter.Severity).
		EqualsNotEmpty("status", filter.Status)

	if instances != nil {
		// Combined generation:
		//  1. group = 'app' AND svc = svc_name
		//  2. group = 'container' AND ((namespace,pod) in (...))
		//  3. group = 'network' AND ((src_namespace,pod) in (...) OR (src_node,pid) in (...))
		//  4. group = 'infra' AND ((instance_name) in (...))
		whereInstance := extractFilter(filter, instances)
		builder.And(whereInstance)
	}

	groupByInstance := `group,severity,tags['svc_name'],tags['content_key'],tags['namespace'],tags['pod'],tags['src_namespace'], tags['src_pod'],tags['src_node'],tags['pid'],tags['instance_name']`

	sql := fmt.Sprintf(SQL_GET_GROUP_COUNTS_ALERT_EVENT, groupByInstance, groupByInstance, builder.String())

	var events []model.AlertEventCount
	err := ch.GetContextDB(ctx).Select(ctx.GetContext(), &events, sql, builder.values...)
	return events, err
}

// Obtain all alarm events of the instance GetAlarmsEvents
func (ch *chRepo) GetAlertEventsSample(ctx core.Context, sampleCount int, startTime time.Time, endTime time.Time, filter request.AlertFilter, instances *model.RelatedInstances) ([]AlertEventSample, error) {
	// Combined generation:
	//  1. group = 'app' AND svc = svc_name
	//  2. group = 'container' AND ((namespace,pod) in (...))
	//  3. group = 'network' AND ((src_namespace,pod) in (...) OR (src_node,pid) in (...))
	//  4. group = 'infra' AND ((instance_name) in (...))
	whereInstance := extractFilter(filter, instances)

	builder := NewQueryBuilder().
		Between("received_time", startTime.Unix(), endTime.Unix()).
		EqualsNotEmpty("source", filter.Source).
		EqualsNotEmpty("group", filter.Group).
		EqualsNotEmpty("name", filter.Name).
		EqualsNotEmpty("id", filter.ID).
		EqualsNotEmpty("severity", filter.Severity).
		EqualsNotEmpty("status", filter.Status).
		And(whereInstance)

	if len(filter.ClusterIDs) > 0 {
		for i := 0; i < len(filter.ClusterIDs); i++ {
			if filter.ClusterIDs[i] == "unknown" {
				filter.ClusterIDs[i] = ""
			}
		}
		builder.InStrings("cluster_id", filter.ClusterIDs)
	}

	byBuilder := NewByLimitBuilder().
		OrderBy("group", true).
		OrderBy("name", true)

	sql := fmt.Sprintf(SQL_GET_SAMPLE_ALERT_EVENT, builder.String(), sampleCount, byBuilder.String())

	var events []AlertEventSample
	err := ch.GetContextDB(ctx).Select(ctx.GetContext(), &events, sql, builder.values...)
	return events, err
}

func (ch *chRepo) GetAlertDataScope(ctx core.Context, startTime time.Time, endTime time.Time) ([]datagroup.DataScope, error) {
	builder := NewQueryBuilder().
		Between("update_time", startTime.Unix(), endTime.Unix()).
		NotGreaterThan("end_time", endTime.Unix())

	sql := fmt.Sprintf(SQL_GET_ALERT_EVENT_DATA_SCOPE, builder.String())
	var datasources []datagroup.DataScope
	err := ch.GetContextDB(ctx).Select(ctx.GetContext(), &datasources, sql, builder.values...)

	for i := 0; i < len(datasources); i++ {
		datasources[i].Category = datagroup.DATASOURCE_CATEGORY_ALERT
		if len(datasources[i].Service) > 0 {
			datasources[i].Type = datagroup.DATASOURCE_TYP_SERVICE
			datasources[i].Name = datasources[i].Service
		} else if len(datasources[i].Namespace) > 0 {
			datasources[i].Type = datagroup.DATASOURCE_TYP_NAMESPACE
			datasources[i].Name = datasources[i].Namespace
		} else {
			datasources[i].Type = datagroup.DATASOURCE_TYP_CLUSTER
			if datasources[i].ClusterID == "" {
				datasources[i].ClusterID = "unknown"
			}
			datasources[i].Name = datasources[i].ClusterID
		}
	}

	return datasources, err
}

func (ch *chRepo) GetAlertEvents(ctx core.Context, startTime time.Time, endTime time.Time, filter request.AlertFilter, instances *model.RelatedInstances, pageParam *request.PageParam, sortBy string) ([]alert.AlertEvent, uint64, error) {
	var whereInstance *whereSQL = ALWAYS_TRUE
	if len(filter.Services) > 0 ||
		len(filter.Endpoints) > 0 ||
		(instances != nil && (len(instances.SIs) > 0 || len(instances.MIs) > 0)) {
		whereInstance = extractFilter(filter, instances)
	}

	return ch.getAlertEventsByWhere(ctx, startTime, endTime, filter, whereInstance, sortBy, pageParam)
}

// GetAlertEvents 取出时间范围内的输入实例的所有告警事件
// instances为空时,不返回任何告警
func (ch *chRepo) GetAlertEventsByInstanceAndEndpoints(ctx core.Context, startTime time.Time, endTime time.Time, filter request.AlertFilter, instances *model.RelatedInstances, pageParam *request.PageParam, sortBy string) ([]alert.AlertEvent, uint64, error) {
	whereInstance := extractFilter(filter, instances)
	return ch.getAlertEventsByWhere(ctx, startTime, endTime, filter, whereInstance, sortBy, pageParam)
}

func (ch *chRepo) GetAlertEventsWithKeyByInstanceAndEndpoints(ctx core.Context, startTime time.Time, endTime time.Time, filter request.AlertFilter, instances *model.RelatedInstances) ([]AlertEventWithKey, error) {
	whereInstance := extractFilter(filter, instances)
	return ch.getGroupAlertEvents(ctx, startTime, endTime, filter, whereInstance)
}

func (ch *chRepo) getGroupAlertEvents(ctx core.Context, startTime time.Time, endTime time.Time, filter request.AlertFilter, whereInstance *whereSQL) ([]AlertEventWithKey, error) {
	builder := NewQueryBuilder().
		Between("received_time", startTime.Unix(), endTime.Unix()).
		EqualsNotEmpty("source", filter.Source).
		EqualsNotEmpty("group", filter.Group).
		EqualsNotEmpty("name", filter.Name).
		EqualsNotEmpty("id", filter.ID).
		EqualsNotEmpty("severity", filter.Severity).
		EqualsNotEmpty("status", filter.Status).
		And(whereInstance)

	// HACK implements data paging based on window functions, which is different from different query statements.
	// !!! Limit / Group parameter must not be added at this location
	orderBuilder := NewByLimitBuilder().
		OrderBy("alert_key", true).
		OrderBy("received_time", true).
		Limit(5000)

	sql := fmt.Sprintf(SQL_GET_ALERT_EVENT_WITH_LABEL_KEY, builder.String(), orderBuilder.String())
	var events []AlertEventWithKey
	err := ch.GetContextDB(ctx).Select(ctx.GetContext(), &events, sql, builder.values...)
	return events, err
}

func (ch *chRepo) getAlertEventsByWhere(ctx core.Context, startTime time.Time, endTime time.Time, filter request.AlertFilter, whereInstance *whereSQL, sortBy string, pageParam *request.PageParam) ([]alert.AlertEvent, uint64, error) {
	builder := NewQueryBuilder().
		Between("update_time", startTime.Unix(), endTime.Unix()).
		NotGreaterThan("end_time", endTime.Unix()).
		EqualsNotEmpty("source", filter.Source).
		EqualsNotEmpty("group", filter.Group).
		EqualsNotEmpty("name", filter.Name).
		EqualsNotEmpty("id", filter.ID).
		EqualsNotEmpty("severity", filter.Severity).
		EqualsNotEmpty("status", filter.Status).
		And(whereInstance)

	if filter.GroupID > 0 {
		builder.Equals("raw_tags['groupId']", strconv.FormatInt(filter.GroupID, 10))
	}

	if len(filter.ClusterIDs) > 0 {
		for i := 0; i < len(filter.ClusterIDs); i++ {
			if filter.ClusterIDs[i] == "unknown" {
				filter.ClusterIDs[i] = ""
			}
		}
		builder.InStrings("cluster_id", filter.ClusterIDs)
	}

	var count uint64
	countSql := buildAlertEventsCountQuery(GET_ALERT_EVENTS_COUNT, builder)
	err := ch.GetContextDB(ctx).QueryRow(ctx.GetContext(), countSql, builder.values...).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	orderBuilder := extractOrderParam(sortBy).
		OrderBy("group", true).
		OrderBy("name", true).
		OrderBy("received_time", false).
		Offset((pageParam.CurrentPage - 1) * pageParam.PageSize).
		Limit(pageParam.PageSize)

	sql := buildGetPagedAlertEventQuery(SQL_GET_PAGED_ALERT_EVENT, builder, orderBuilder)
	var events []alert.AlertEvent
	err = ch.GetContextDB(ctx).Select(ctx.GetContext(), &events, sql, builder.values...)
	if err != nil {
		return nil, 0, err
	}

	return events, count, err
}

func buildGetPagedAlertEventQuery(baseQuery string, builder *QueryBuilder, orderBuilder *ByLimitBuilder) string {
	sql := fmt.Sprintf(baseQuery, builder.String(), orderBuilder.String())
	return sql
}

func buildAlertEventsCountQuery(baseQuery string, builder *QueryBuilder) string {
	countSql := fmt.Sprintf(baseQuery, builder.String())
	return countSql
}

func (ch *chRepo) InsertBatchAlertEvents(ctx core.Context, events []*model.AlertEvent) error {
	batch, err := ch.GetContextDB(ctx).PrepareBatch(ctx.GetContext(), `
		INSERT INTO alert_event (source, id, alert_id, create_time, update_time, end_time, received_time, severity, group,
		                         name, detail, tags, status)
		VALUES
	`)
	if err != nil {
		return err
	}
	for _, event := range events {
		alertId := alert.FastAlertIDByStringMap(event.Name, event.Tags)
		if err := batch.Append(event.Source, event.ID, alertId, event.CreateTime, event.UpdateTime, event.EndTime,
			event.ReceivedTime, int8(event.Severity), event.Group, event.Name, event.Detail, event.Tags, int8(event.Status)); err != nil {
			log.Println("Failed to send data:", err)
			continue
		}
	}

	if err := batch.Send(); err != nil {
		return err
	}
	return nil
}

// ReadAlertEvent implement the Read method of the AlertEventDAO interface
func (ch *chRepo) ReadAlertEvent(ctx core.Context, id uuid.UUID) (*model.AlertEvent, error) {
	var event model.AlertEvent
	query := `
		SELECT source, id, create_time, update_time, end_time, received_time, severity
		       ,group, name, detail, tags, status
		FROM alert_event
		WHERE id = ?
	`
	err := ch.GetContextDB(ctx).QueryRow(ctx.GetContext(), query, id).Scan(
		&event.Source, &event.ID, &event.CreateTime, &event.UpdateTime, &event.EndTime,
		&event.ReceivedTime, &event.Severity, &event.Group, &event.Name, &event.Detail, &event.Tags, &event.Status,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("event with ID %s not found", id)
		}
		return nil, fmt.Errorf("failed to read event: %w", err)
	}
	return &event, nil
}

func (ch *chRepo) GetAlertEventById(ctx core.Context, eventId string, startTime, endTime time.Time) (*model.AlertEvent, error) {
	builder := NewQueryBuilder().
		Between("received_time", startTime.Unix(), endTime.Unix()).
		Equals("id", eventId)

	byLimit := NewByLimitBuilder().Limit(1)

	sql := fmt.Sprintf(SQL_GET_ALERT_EVENT, builder.String(), byLimit.String())
	var event []model.AlertEvent
	err := ch.GetContextDB(ctx).Select(ctx.GetContext(), &event, sql, builder.values...)
	if err != nil {
		return nil, err
	} else if len(event) == 0 {
		return nil, fmt.Errorf("no event found with ID %s", eventId)
	}
	return &event[0], err
}

func extractFilter(filter request.AlertFilter, instances *model.RelatedInstances) *whereSQL {
	var whereInstance []*whereSQL
	whereGroup := equalsIfNotEmpty("group", filter.Group)
	if filter.WithMutation {
		whereGroup = mergeWheres(OrSep, whereGroup, equalsIfNotEmpty("group", "mutation-"+filter.Group))
	}

	if len(filter.Group) == 0 || filter.Group == "app" {
		serviceCondition := []*whereSQL{}
		if len(filter.Services) > 0 {
			arr := make(clickhouse.ArraySet, 0, len(filter.Services))
			for _, s := range filter.Services {
				arr = append(arr, s)
			}

			endpoints := make(clickhouse.ArraySet, 0, len(filter.Endpoints))
			for _, e := range filter.Endpoints {
				endpoints = append(endpoints, e)
			}

			serviceCondition = append(serviceCondition,
				mergeWheres(
					AndSep,
					in("tags['serviceName']", arr),
					in("tags['endpoint']", endpoints),
				),
			)
			whereInstance = append(whereInstance, serviceCondition...)
		} else if len(filter.EndpointKeys) > 0 {
			whereInstance = append(whereInstance, extractEndpointKeys(filter.EndpointKeys, whereGroup)...)
		}
	}

	if len(filter.Group) == 0 || filter.Group == "container" {
		var k8sPods ValueInGroups = ValueInGroups{
			Keys: []string{"tags['namespace']", "tags['pod']"},
		}

		if instances != nil {
			for _, instance := range instances.SIs {
				if instance == nil {
					continue
				}
				if len(instance.PodName) > 0 {
					k8sPods.ValueGroups = append(k8sPods.ValueGroups, clickhouse.GroupSet{
						Value: []any{instance.Namespace, instance.PodName},
					})
				}
			}

			whereInstance = append(whereInstance, mergeWheres(
				AndSep,
				whereGroup,
				inGroup(k8sPods),
			))
		}
	}

	// filter by VMProcess
	// filter by VMProcess
	if len(filter.Group) == 0 || filter.Group == "network" {
		var k8sPods ValueInGroups = ValueInGroups{
			Keys: []string{"tags['src_namespace']", "tags['src_pod']"},
		}
		var vmPods ValueInGroups = ValueInGroups{
			Keys: []string{"tags['node_name']", "tags['pid']"},
		}
		var vmPodsNew ValueInGroups = ValueInGroups{
			Keys: []string{"tags['node']", "tags['pid']"},
		}

		if instances != nil {
			for _, instance := range instances.SIs {
				if instance == nil {
					continue
				}
				if len(instance.PodName) > 0 {
					k8sPods.ValueGroups = append(k8sPods.ValueGroups, clickhouse.GroupSet{
						Value: []any{instance.Namespace, instance.PodName},
					})
				} else {
					vmPods.ValueGroups = append(vmPods.ValueGroups, clickhouse.GroupSet{
						Value: []any{instance.NodeName, instance.Pid},
					})
					vmPodsNew.ValueGroups = append(vmPodsNew.ValueGroups, clickhouse.GroupSet{
						Value: []any{instance.NodeName, instance.Pid},
					})
					vmPodsNew.ValueGroups = append(vmPodsNew.ValueGroups, clickhouse.GroupSet{
						Value: []any{instance.NodeName, instance.Pid},
					})
				}
			}

			k8sOrVm := mergeWheres(OrSep,
				inGroup(k8sPods), // Compatible with older versions
				inGroup(vmPods),  // Compatible with older versions
				inGroup(vmPodsNew),
			)
			whereInstance = append(whereInstance, mergeWheres(
				AndSep,
				whereGroup,
				k8sOrVm,
			))
		}
	}

	if len(filter.Group) == 0 || filter.Group == "infra" {
		infraGroup := equals("group", "infra")
		var tmpSet = map[string]struct{}{}
		var nodes clickhouse.ArraySet

		if instances != nil {
			for _, instance := range instances.SIs {
				if instance == nil {
					continue
				}
				_, find := tmpSet[instance.NodeName]
				if !find {
					nodes = append(nodes, instance.NodeName)
					tmpSet[instance.NodeName] = struct{}{}
				}
			}

			whereInstance = append(whereInstance, mergeWheres(
				AndSep,
				infraGroup,
				mergeWheres(OrSep,
					in("tags['instance_name']", nodes),
					in("tags['node']", nodes),
				),
			))
		}
	}

	if len(filter.Group) == 0 || filter.Group == "middleware" {
		var dbUrls clickhouse.ArraySet
		var ipPorts ValueInGroups = ValueInGroups{
			Keys: []string{"tags['dbIP']", "tags['dbPort']"},
		}

		if instances != nil {
			for _, middleware := range instances.MIs {
				if len(middleware.DatabaseURL) > 0 {
					dbUrls = append(dbUrls, middleware.DatabaseURL)
				}
				if len(middleware.DatabaseIP) > 0 {
					ipPorts.ValueGroups = append(ipPorts.ValueGroups, clickhouse.GroupSet{
						Value: []any{middleware.DatabaseIP, middleware.DatabasePort},
					})
				}
			}
			whereInstance = append(whereInstance, mergeWheres(
				OrSep,
				in("tags['dbURL']", dbUrls),
				inGroup(ipPorts),
			))
		}
	}

	return mergeWheres(OrSep, whereInstance...)
}

func extractEndpointKeys(endpoints []model.EndpointKey, whereGroup *whereSQL) []*whereSQL {
	var whereInstance []*whereSQL
	whereEndpoints := ValueInGroups{
		Keys: []string{"tags['svc_name']", "tags['content_key']"},
	}

	whereEndpointsNew := ValueInGroups{
		Keys: []string{"tags['serviceName']", "tags['endpoint']"},
	}
	for _, endpoint := range endpoints {
		whereEndpoints.ValueGroups = append(whereEndpoints.ValueGroups, clickhouse.GroupSet{
			Value: []any{endpoint.ServiceName, endpoint.Endpoint},
		})
		whereEndpointsNew.ValueGroups = append(whereEndpointsNew.ValueGroups, clickhouse.GroupSet{
			Value: []any{endpoint.ServiceName, endpoint.Endpoint},
		})
	}
	whereInstance = append(whereInstance, mergeWheres(
		AndSep,
		whereGroup,
		mergeWheres(
			OrSep,
			inGroup(whereEndpoints),
			inGroup(whereEndpointsNew),
		),
	))
	return whereInstance
}

type AlertEventSample struct {
	model.AlertEvent

	// Record line number
	Rn         uint64 `ch:"rn" json:"-"`
	AlarmCount uint64 `ch:"alarm_count" json:"alarmCount"`

	AlertKey string `ch:"alert_key" json:"alertKey"`
}

type AlertEventWithKey struct {
	model.AlertEvent

	AlertKey string `ch:"alert_key" json:"alertKey"`
}

const (
	OrderAlertByGroupName    = "group,name,-receivedTime"
	OrderAlertByReceivedTime = "receivedTime,group,name"
)

// 常用排序
// orderAlertByGroupName
// orderAlertByReceivedTime
var (
	orderAlertByGroupName = NewByLimitBuilder().
				OrderBy("group", true).
				OrderBy("name", true).
				OrderBy("received_time", false)
	orderAlertByReceivedTime = NewByLimitBuilder().
					OrderBy("received_time", false).
					OrderBy("group", true).
					OrderBy("name", true)
)

// extractOrderParam 解析排序参数
func extractOrderParam(sortBy string) *ByLimitBuilder {
	if sortBy == OrderAlertByGroupName {
		return orderAlertByGroupName
	} else if sortBy == OrderAlertByReceivedTime {
		return orderAlertByReceivedTime
	} else if len(sortBy) == 0 {
		return NewByLimitBuilder()
	}

	builder := NewByLimitBuilder()
	parts := strings.Split(sortBy, ",")
	for _, part := range parts {
		key := strings.TrimSpace(part)
		if strings.HasPrefix(key, "-") {
			builder.OrderBy(camelToSnake(strings.TrimPrefix(key, "-")), false)
		} else if strings.HasPrefix(key, "+") {
			builder.OrderBy(camelToSnake(strings.TrimPrefix(key, "+")), true)
		} else {
			builder.OrderBy(camelToSnake(key), true)
		}
	}
	return builder
}

func camelToSnake(s string) string {
	var result []rune
	for i, r := range s {
		// 如果是大写字母，且不是第一个字符，添加下划线
		if unicode.IsUpper(r) {
			// 非首字母加下划线
			if i > 0 {
				result = append(result, '_')
			}
			// 转换为小写字母
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}
