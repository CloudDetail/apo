// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/ClickHouse/clickhouse-go/v2"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

type AlertGroup string

const (
	APP_GROUP          AlertGroup = "app"
	APP_INSTANCE_GROUP AlertGroup = "appInstance"
	NETWORK_GROUP      AlertGroup = "network"
	CONTAINER_GROUP    AlertGroup = "container"
	INFRA_GROUP        AlertGroup = "infra"

	MUTATION_APP_GROUP          AlertGroup = "mutation-app"
	MUTATION_APP_INSTANCE_GROUP AlertGroup = "mutation-appInstance"
	MUTATION_NETWORK_GROUP      AlertGroup = "mutation-network"
	MUTATION_CONTAINER_GROUP    AlertGroup = "mutation-container"
	MUTATION_INFRA_GROUP        AlertGroup = "mutation-infra"
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
		SELECT source,group,id,create_time,update_time,end_time,received_time,severity,name,detail,tags,status,
        	arrayStringConcat(arrayMap(x -> x.2, arraySort(arrayZip(mapKeys(tags), mapValues(tags)))), ', ') AS alert_key,
			ROW_NUMBER() OVER (PARTITION BY name, alert_key ORDER BY received_time) AS rn,
			COUNT(*) OVER (PARTITION BY name, alert_key) AS alarm_count
    	FROM alert_event
		%s
	)
	SELECT *
	FROM grouped_alarm
	WHERE rn <= %d %s`

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

	// SQL _GET_PAGED_ALERT_EVENT paging out all alarm events that meet the conditions
	SQL_GET_PAGED_ALERT_EVENT = `WITH paginatedEvent AS (
		SELECT
			source,group,id,create_time,update_time,end_time,received_time,severity,name,detail,tags,status,
			COUNT(*) OVER () AS total_count,
			ROW_NUMBER() OVER (%s) AS rn
		FROM alert_event
		%s
	)
	SELECT *
	FROM paginatedEvent
	%s ORDER BY rn`

	SQL_GET_ALERT_EVENT = `SELECT source,group,id,create_time,update_time,end_time,received_time,severity,name,detail,tags,status
	FROM alert_event %s %s`

	SQL_GET_ALERT_EVENT_WITH_LABEL_KEY = `SELECT source,group,id,create_time,update_time,end_time,received_time,severity,name,detail,tags,status
        ,arrayStringConcat(arrayMap(x -> x.2, arraySort(arrayZip(mapKeys(tags), mapValues(tags)))), ', ') AS alert_key
    FROM alert_event
	%s %s`
)

// GetAlertEventCountGroupByInstance to quickly query the number of alarms associated with each Instance (counted separately by alarm level)
func (ch *chRepo) GetAlertEventCountGroupByInstance(startTime time.Time, endTime time.Time, filter request.AlertFilter, instances []*model.ServiceInstance) ([]model.AlertEventCount, error) {
	builder := NewQueryBuilder().
		Between("received_time", startTime.Unix(), endTime.Unix()).
		EqualsNotEmpty("source", filter.Source).
		EqualsNotEmpty("group", filter.Group).
		EqualsNotEmpty("name", filter.Name).
		EqualsNotEmpty("id", filter.ID).
		EqualsNotEmpty("severity", filter.Severity).
		EqualsNotEmpty("status", filter.Status)

	if len(instances) > 0 {
		// Combined generation:
		//  1. group = 'app' AND svc = svc_name
		//  2. group = 'container' AND ((namespace,pod) in (...))
		//  3. group = 'network' AND ((src_namespace,pod) in (...) OR (src_node,pid) in (...))
		//  4. group = 'infra' AND ((instance_name) in (...))
		whereInstance := extractFilter(filter, nil, instances)
		builder.And(whereInstance)
	}

	groupByInstance := `group,severity,tags['svc_name'],tags['content_key'],tags['namespace'],tags['pod'],tags['src_namespace'], tags['src_pod'],tags['src_node'],tags['pid'],tags['instance_name']`

	sql := fmt.Sprintf(SQL_GET_GROUP_COUNTS_ALERT_EVENT, groupByInstance, groupByInstance, builder.String())

	var events []model.AlertEventCount
	err := ch.conn.Select(context.Background(), &events, sql, builder.values...)
	return events, err
}

// Obtain all alarm events of the instance GetAlarmsEvents
func (ch *chRepo) GetAlertEventsSample(sampleCount int, startTime time.Time, endTime time.Time, filter request.AlertFilter, instances []*model.ServiceInstance) ([]AlertEventSample, error) {
	// Combined generation:
	//  1. group = 'app' AND svc = svc_name
	//  2. group = 'container' AND ((namespace,pod) in (...))
	//  3. group = 'network' AND ((src_namespace,pod) in (...) OR (src_node,pid) in (...))
	//  4. group = 'infra' AND ((instance_name) in (...))
	whereInstance := extractFilter(filter, nil, instances)

	builder := NewQueryBuilder().
		Between("received_time", startTime.Unix(), endTime.Unix()).
		EqualsNotEmpty("source", filter.Source).
		EqualsNotEmpty("group", filter.Group).
		EqualsNotEmpty("name", filter.Name).
		EqualsNotEmpty("id", filter.ID).
		EqualsNotEmpty("severity", filter.Severity).
		EqualsNotEmpty("status", filter.Status).
		And(whereInstance)

	byBuilder := NewByLimitBuilder().
		OrderBy("group", true).
		OrderBy("name", true)

	sql := fmt.Sprintf(SQL_GET_SAMPLE_ALERT_EVENT, builder.String(), sampleCount, byBuilder.String())

	var events []AlertEventSample
	err := ch.conn.Select(context.Background(), &events, sql, builder.values...)
	return events, err
}

var (
	GroupOrders = NewByLimitBuilder().
			OrderBy("group", true).
			OrderBy("name", true).
			OrderBy("received_time", false).String()

	ReceivedTimeOrders = NewByLimitBuilder().
				OrderBy("received_time", false).String()
)

// GetAlertEvents 取出时间范围内的输入实例的所有告警事件
// instances为空时,不返回任何告警
func (ch *chRepo) GetAlertEvents(startTime time.Time, endTime time.Time, filter request.AlertFilter, instances []*model.ServiceInstance, pageParam *request.PageParam, sortBy string) ([]PagedAlertEvent, int, error) {
	var whereInstance *whereSQL
	if len(filter.Service) > 0 || len(filter.Endpoint) > 0 || len(instances) > 0 {
		whereInstance = extractFilter(filter, nil, instances)
	} else {
		whereInstance = ALWAYS_TRUE
	}
	return ch.getAlertEventsByWhere(startTime, endTime, filter, whereInstance, sortBy, pageParam)
}

// GetAlertEvents 取出时间范围内的输入实例的所有告警事件
// instances为空时,不返回任何告警
func (ch *chRepo) GetAlertEventsByInstanceAndEndpoints(startTime time.Time, endTime time.Time, filter request.AlertFilter, instances []*model.ServiceInstance, endpoints []model.EndpointKey, pageParam *request.PageParam, sortBy string) ([]PagedAlertEvent, int, error) {
	whereInstance := extractFilter(filter, endpoints, instances)
	return ch.getAlertEventsByWhere(startTime, endTime, filter, whereInstance, sortBy, pageParam)
}

func (ch *chRepo) GetAlertEventsWithKeyByInstanceAndEndpoints(startTime time.Time, endTime time.Time, filter request.AlertFilter, instances []*model.ServiceInstance, endpoints []model.EndpointKey) ([]AlertEventWithKey, error) {
	whereInstance := extractFilter(filter, endpoints, instances)
	return ch.getGroupAlertEvents(startTime, endTime, filter, whereInstance)
}

func (ch *chRepo) getGroupAlertEvents(startTime time.Time, endTime time.Time, filter request.AlertFilter, whereInstance *whereSQL) ([]AlertEventWithKey, error) {
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
	err := ch.conn.Select(context.Background(), &events, sql, builder.values...)
	return events, err
}

func (ch *chRepo) getAlertEventsByWhere(startTime time.Time, endTime time.Time, filter request.AlertFilter, whereInstance *whereSQL, sortBy string, pageParam *request.PageParam) ([]PagedAlertEvent, int, error) {
	builder := NewQueryBuilder().
		Between("received_time", startTime.Unix(), endTime.Unix()).
		EqualsNotEmpty("source", filter.Source).
		EqualsNotEmpty("group", filter.Group).
		EqualsNotEmpty("name", filter.Name).
		EqualsNotEmpty("id", filter.ID).
		EqualsNotEmpty("severity", filter.Severity).
		EqualsNotEmpty("status", filter.Status).
		And(whereInstance)

	orderBuilder := extractOrderParam(sortBy)
	orders := orderBuilder.String()

	sql := fmt.Sprintf(SQL_GET_PAGED_ALERT_EVENT, orders, builder.String(), RnLimit(pageParam))
	var events []PagedAlertEvent
	err := ch.conn.Select(context.Background(), &events, sql, builder.values...)
	var total_count = 0
	if len(events) > 0 {
		total_count = int(events[0].TotalCount)
	}
	return events, total_count, err
}

func (ch *chRepo) GetAlertEventById(eventId string, startTime, endTime time.Time) (*model.AlertEvent, error) {
	builder := NewQueryBuilder().
		Between("received_time", startTime.Unix(), endTime.Unix()).
		Equals("id", eventId)

	byLimit := NewByLimitBuilder().Limit(1)

	sql := fmt.Sprintf(SQL_GET_ALERT_EVENT, builder.String(), byLimit.String())
	var event []model.AlertEvent
	err := ch.conn.Select(context.Background(), &event, sql, builder.values...)
	if err != nil {
		return nil, err
	} else if len(event) == 0 {
		return nil, fmt.Errorf("no event found with ID %s", eventId)
	}
	return &event[0], err
}

func extractFilter(filter request.AlertFilter, endpoints []model.EndpointKey, instances []*model.ServiceInstance) *whereSQL {
	var whereInstance []*whereSQL
	if len(filter.Group) == 0 || filter.Group == "app" {
		whereGroup := EqualsIfNotEmpty("group", "app")
		if filter.WithMutation {
			whereGroup = MergeWheres(OrSep, whereGroup, Equals("group", "mutation-app"))
		}
		if len(endpoints) == 0 {
			whereInstance = append(whereInstance, MergeWheres(
				AndSep,
				whereGroup,
				Equals("tags['svc_name']", filter.Service),
				EqualsIfNotEmpty("tags['content_key']", filter.Endpoint),
			))
		} else {
			whereEndpoints := ValueInGroups{
				Keys: []string{"tags['svc_name']", "tags['content_key']"},
			}
			for _, endpoint := range endpoints {
				whereEndpoints.ValueGroups = append(whereEndpoints.ValueGroups, clickhouse.GroupSet{
					Value: []any{endpoint.ServiceName, endpoint.Endpoint},
				})
			}
			whereInstance = append(whereInstance, MergeWheres(
				AndSep,
				whereGroup,
				InGroup(whereEndpoints),
			))
		}
	}

	if len(filter.Group) == 0 || filter.Group == "container" {
		whereGroup := EqualsIfNotEmpty("group", "container")
		if filter.WithMutation {
			whereGroup = MergeWheres(OrSep, whereGroup, Equals("group", "mutation-container"))
		}
		var k8sPods ValueInGroups = ValueInGroups{
			Keys: []string{"tags['namespace']", "tags['pod']"},
		}
		for _, instance := range instances {
			if instance == nil {
				continue
			}
			if len(instance.PodName) > 0 {
				k8sPods.ValueGroups = append(k8sPods.ValueGroups, clickhouse.GroupSet{
					Value: []any{instance.Namespace, instance.PodName},
				})
			}
		}

		whereInstance = append(whereInstance, MergeWheres(
			AndSep,
			whereGroup,
			InGroup(k8sPods),
		))
	}

	if len(filter.Group) == 0 || filter.Group == "network" {
		whereGroup := EqualsIfNotEmpty("group", "network")
		if filter.WithMutation {
			whereGroup = MergeWheres(OrSep, whereGroup, Equals("group", "mutation-network"))
		}
		var k8sPods ValueInGroups = ValueInGroups{
			Keys: []string{"tags['src_namespace']", "tags['src_pod']"},
		}
		var vmPods ValueInGroups = ValueInGroups{
			Keys: []string{"tags['node_name']", "tags['pid']"},
		}

		for _, instance := range instances {
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
			}
		}

		k8sOrVm := MergeWheres(OrSep, InGroup(k8sPods), InGroup(vmPods))
		whereInstance = append(whereInstance, MergeWheres(
			AndSep,
			whereGroup,
			k8sOrVm,
		))
	}

	if len(filter.Group) == 0 || filter.Group == "infra" {
		whereGroup := EqualsIfNotEmpty("group", "infra")
		if filter.WithMutation {
			whereGroup = MergeWheres(OrSep, whereGroup, Equals("group", "mutation-infra"))
		}
		var tmpSet = map[string]struct{}{}
		var nodes clickhouse.ArraySet
		for _, instance := range instances {
			if instance == nil {
				continue
			}
			_, find := tmpSet[instance.NodeName]
			if !find {
				nodes = append(nodes, instance.NodeName)
				tmpSet[instance.NodeName] = struct{}{}
			}
		}

		whereInstance = append(whereInstance, MergeWheres(
			AndSep,
			whereGroup,
			In("tags['instance_name']", nodes),
		))
	}

	return MergeWheres(OrSep, whereInstance...)
}

type AlertEventSample struct {
	model.AlertEvent

	// Record line number
	Rn         uint64 `ch:"rn" json:"-"`
	AlarmCount uint64 `ch:"alarm_count" json:"alarmCount"`

	AlertKey string `ch:"alert_key" json:"alertKey"`
}

type PagedAlertEvent struct {
	model.AlertEvent

	// Record line number
	Rn         uint64 `ch:"rn" json:"-"`
	TotalCount uint64 `ch:"total_count" json:"-"`
}

type AlertEventWithKey struct {
	model.AlertEvent

	AlertKey string `ch:"alert_key" json:"alertKey"`
}

func RnLimit(p *request.PageParam) string {
	if p == nil {
		return ""
	}
	startIdx := 1 + (p.CurrentPage-1)*p.PageSize
	endIdx := p.CurrentPage * p.PageSize
	return fmt.Sprintf(" WHERE rn BETWEEN %d AND %d ", startIdx, endIdx)
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
