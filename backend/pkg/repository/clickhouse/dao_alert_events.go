// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/google/uuid"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

type AlertGroup string

const (
	APP_GROUP        AlertGroup = "app"
	NETWORK_GROUP    AlertGroup = "network"
	CONTAINER_GROUP  AlertGroup = "container"
	INFRA_GROUP      AlertGroup = "infra"
	MIDDLEWARE_GROUP AlertGroup = "middleware"
)

func (g AlertGroup) GetAlertType() string {
	switch g {
	case INFRA_GROUP:
		return model.InfrastructureAlert
	case NETWORK_GROUP:
		return model.NetAlert
	case APP_GROUP:
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

	GET_ALERT_EVENTS_COUNT = `SELECT count(1) as count FROM alert_event %s`

	SQL_GET_PAGED_ALERT_EVENT = `SELECT
		source,group,id,create_time,update_time,end_time,received_time,severity,name,detail,tags,raw_tags,status,alert_id
	FROM alert_event
	%s %s`
)

// GetAlertEventCountGroupByInstance to quickly query the number of alarms associated with each Instance (counted separately by alarm level)
func (ch *chRepo) GetAlertEventCountGroupByInstance(startTime time.Time, endTime time.Time, filter request.AlertFilter, instances *model.RelatedInstances) ([]model.AlertEventCount, error) {
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
	err := ch.conn.Select(context.Background(), &events, sql, builder.values...)
	return events, err
}

// Obtain all alarm events of the instance GetAlarmsEvents
func (ch *chRepo) GetAlertEventsSample(sampleCount int, startTime time.Time, endTime time.Time, filter request.AlertFilter, instances *model.RelatedInstances) ([]AlertEventSample, error) {
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

	byBuilder := NewByLimitBuilder().
		OrderBy("group", true).
		OrderBy("name", true)

	sql := fmt.Sprintf(SQL_GET_SAMPLE_ALERT_EVENT, builder.String(), sampleCount, byBuilder.String())

	var events []AlertEventSample
	err := ch.conn.Select(context.Background(), &events, sql, builder.values...)
	return events, err
}

func (ch *chRepo) GetAlertEvents(startTime time.Time, endTime time.Time, filter request.AlertFilter, instances *model.RelatedInstances, pageParam *request.PageParam) ([]alert.AlertEvent, uint64, error) {
	var whereInstance *whereSQL = ALWAYS_TRUE
	if len(filter.Services) > 0 ||
		len(filter.Endpoint) > 0 ||
		(instances != nil && (len(instances.SIs) > 0 || len(instances.MIs) > 0)) {
		whereInstance = extractFilter(filter, instances)
	}

	source, valid := validateInputStr(filter.Source)
	if !valid {
		return nil, 0, errors.New("invalid source")
	}
	group, valid := validateInputStr(filter.Group)
	if !valid {
		return nil, 0, errors.New("invalid group")
	}
	name, valid := validateInputStr(filter.Name)
	if !valid {
		return nil, 0, errors.New("invalid name")
	}
	id, valid := validateInputStr(filter.ID)
	if !valid {
		return nil, 0, errors.New("invalid id")
	}
	severity, valid := validateInputStr(filter.Severity)
	if !valid {
		return nil, 0, errors.New("invalid severity")
	}
	status, valid := validateInputStr(filter.Status)
	if !valid {
		return nil, 0, errors.New("invalid status")
	}

	builder := NewQueryBuilder().
		Between("received_time", startTime.Unix(), endTime.Unix()).
		EqualsNotEmpty("source", source).
		EqualsNotEmpty("group", group).
		EqualsNotEmpty("name", name).
		EqualsNotEmpty("id", id).
		EqualsNotEmpty("severity", severity).
		EqualsNotEmpty("status", status).
		And(whereInstance)

	var count uint64
	countSql := buildAlertEventsCountQuery(GET_ALERT_EVENTS_COUNT, builder)
	err := ch.conn.QueryRow(context.Background(), countSql, builder.values...).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	// HACK implements data paging based on window functions, which is different from different query statements.
	// !!! Limit / Group parameter must not be added at this location
	orderBuilder := NewByLimitBuilder().
		OrderBy("group", true).
		OrderBy("name", true).
		OrderBy("received_time", false).
		Offset((pageParam.CurrentPage - 1) * pageParam.PageSize).
		Limit(pageParam.PageSize)

	sql := buildGetPagedAlertEventQuery(SQL_GET_PAGED_ALERT_EVENT, builder, orderBuilder)
	var events []alert.AlertEvent
	err = ch.conn.Select(context.Background(), &events, sql, builder.values...)
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

func (ch *chRepo) InsertBatchAlertEvents(ctx context.Context, events []*model.AlertEvent) error {
	batch, err := ch.conn.PrepareBatch(ctx, `
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
func (ch *chRepo) ReadAlertEvent(ctx context.Context, id uuid.UUID) (*model.AlertEvent, error) {
	var event model.AlertEvent
	query := `
		SELECT source, id, create_time, update_time, end_time, received_time, severity
		       ,group, name, detail, tags, status
		FROM alert_event
		WHERE id = ?
	`
	err := ch.conn.QueryRow(ctx, query, id).Scan(
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

func extractFilter(filter request.AlertFilter, instances *model.RelatedInstances) *whereSQL {
	var whereInstance []*whereSQL
	whereGroup := equalsIfNotEmpty("group", filter.Group)

	if len(filter.Group) == 0 || filter.Group == "app" {
		serviceCondition := []*whereSQL{}
		if len(filter.Services) > 0 {
			arr := make(clickhouse.ArraySet, 0, len(filter.Services))
			for _, s := range filter.Services {
				arr = append(arr, s)
			}
			serviceCondition = append(serviceCondition,
				mergeWheres(
					OrSep,
					mergeWheres(
						AndSep,
						in("tags['svc_name']", arr),
						equalsIfNotEmpty("tags['content_key']", filter.Endpoint),
					), // Compatible with older versions
					mergeWheres(
						AndSep,
						in("tags['serviceName']", arr),
						equalsIfNotEmpty("tags['endpoint']", filter.Endpoint),
					),
				),
			)
		}

		whereInstance = append(whereInstance, serviceCondition...)
	}

	if len(filter.Group) == 0 || filter.Group == "container" || filter.Group == "network" {
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

type AlertEventSample struct {
	model.AlertEvent

	// Record line number
	Rn         uint64 `ch:"rn" json:"-"`
	AlarmCount uint64 `ch:"alarm_count" json:"alarmCount"`

	AlertKey string `ch:"alert_key" json:"alertKey"`
}
