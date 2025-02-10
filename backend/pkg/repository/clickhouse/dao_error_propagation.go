// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

const (
	SQL_GET_INSTANCE_ERROR_PROPAGATION = `
		WITH found_trace_ids AS
		(
			SELECT error_propagation.timestamp as timestamp, error_propagation.trace_id as trace_id, error_propagation.entry_span_id as entry_span_id,
				nodes.service as service, nodes.instance as instance_id, nodes.path as path, nodes.depth as depth, nodes.error_types as error_types, nodes.error_msgs as error_msgs, nodes.is_error as is_error
			FROM %s.error_propagation
			ARRAY JOIN nodes
			%s %s
		)
		SELECT found_trace_ids.timestamp as timestamp, found_trace_ids.service as service, found_trace_ids.instance_id as instance_id, found_trace_ids.trace_id as trace_id, found_trace_ids.error_types as error_types, found_trace_ids.error_msgs as error_msgs,
			   parent_node.parent_services as parent_services, parent_node.parent_instances as parent_instances, parent_node.parent_traced as parent_traced,
			   child_node.child_services as child_services, child_node.child_instances as child_instances, child_node.child_traced as child_traced
		FROM found_trace_ids
		LEFT JOIN(
			SELECT error_propagation.trace_id as trace_id, groupArray(nodes.service) as parent_services, groupArray(nodes.instance) as parent_instances, groupArray(nodes.is_traced) as parent_traced
			FROM error_propagation
			ARRAY JOIN nodes
			GLOBAL JOIN found_trace_ids ON error_propagation.trace_id = found_trace_ids.trace_id AND error_propagation.entry_span_id = found_trace_ids.entry_span_id
			WHERE timestamp BETWEEN %d AND %d AND startsWith(found_trace_ids.path, nodes.path) AND nodes.depth=found_trace_ids.depth - 1 AND nodes.is_error = found_trace_ids.is_error
			GROUP BY trace_id
		) AS parent_node ON parent_node.trace_id = found_trace_ids.trace_id
		LEFT JOIN(
			SELECT error_propagation.trace_id as trace_id, groupArray(nodes.service) as child_services, groupArray(nodes.instance) as child_instances, groupArray(nodes.is_traced) as child_traced
			FROM error_propagation
			ARRAY JOIN nodes
			GLOBAL JOIN found_trace_ids ON error_propagation.trace_id = found_trace_ids.trace_id AND error_propagation.entry_span_id = found_trace_ids.entry_span_id
			WHERE timestamp BETWEEN %d AND %d AND startsWith(nodes.path, found_trace_ids.path) AND nodes.depth=found_trace_ids.depth+1 AND nodes.is_error = found_trace_ids.is_error
			GROUP BY trace_id
		) AS child_node on child_node.trace_id = found_trace_ids.trace_id
	`

	SQL_GET_ERROR_PROPAGATION = `SELECT timestamp,entry_service,entry_url,entry_span_id,trace_id
		, nodes.service,nodes.instance,nodes.url,nodes.is_traced,nodes.is_error
		, nodes.error_types,nodes.error_msgs,nodes.depth,nodes.path
		FROM %s.error_propagation
		%s %s
	`
)

func (ch *chRepo) ListErrorByEntryService(startTime, endTime int64, entryService, entryEndpoint string, endpoints []model.EndpointKey) ([]ErrorPropation, error) {
	startTime = startTime / 1000000
	endTime = endTime / 1000000

	whereEntry := MergeWheres(
		AndSep,
		Equals("entry_service", entryService),
		Equals("entry_url", entryEndpoint),
	)
	// 如果APM发生了断链
	// 可以通过endpoints直接去除对应服务的Exception
	if len(endpoints) > 0 {
		whereEndpoints := ValueInGroups{
			Keys: []string{"entry_service", "entry_url"},
		}
		for _, endpoint := range endpoints {
			whereEndpoints.ValueGroups = append(whereEndpoints.ValueGroups, clickhouse.GroupSet{
				Value: []any{endpoint.ServiceName, endpoint.Endpoint},
			})
		}
		whereEntry = MergeWheres(
			OrSep,
			whereEntry,
			InGroup(whereEndpoints),
		)
	}

	queryBuilder := NewQueryBuilder().
		Between("timestamp", startTime, endTime).
		Statement("LENGTH(nodes.error_types) > 0"). // 返回的数据必须有ErrorTypes
		And(whereEntry)

	bySql := NewByLimitBuilder().
		OrderBy("timestamp", false).
		Limit(2000).String()

	var results []ErrorPropation
	sql := fmt.Sprintf(SQL_GET_ERROR_PROPAGATION, ch.database, queryBuilder.String(), bySql)
	if err := ch.conn.Select(context.Background(), &results, sql, queryBuilder.values...); err != nil {
		return nil, err
	}
	return results, nil
}

func (ch *chRepo) ListErrorPropagation(req *request.GetErrorInstanceRequest) ([]ErrorInstancePropagation, error) {
	startTime := req.StartTime / 1000000
	endTime := req.EndTime / 1000000
	queryBuilder := NewQueryBuilder().
		Between("timestamp", startTime, endTime).
		Equals("nodes.service", req.Service).
		Equals("nodes.url", req.Endpoint).
		Equals("nodes.is_traced", true).
		Equals("nodes.is_error", true).
		EqualsNotEmpty("entry_service", req.EntryService).
		EqualsNotEmpty("entry_url", req.EntryEndpoint).
		Statement("LENGTH(nodes.error_types) > 0") // The data returned must be ErrorTypes
	bySql := NewByLimitBuilder().
		OrderBy("timestamp", false).
		Limit(2000).String()
	var results []ErrorInstancePropagation
	sql := fmt.Sprintf(SQL_GET_INSTANCE_ERROR_PROPAGATION, ch.database, queryBuilder.String(), bySql, startTime, endTime, startTime, endTime)
	if err := ch.conn.Select(context.Background(), &results, sql, queryBuilder.values...); err != nil {
		return nil, err
	}
	return results, nil
}

type ErrorInstancePropagation struct {
	Timestamp       time.Time `ch:"timestamp"`
	Service         string    `ch:"service"`
	InstanceId      string    `ch:"instance_id"`
	TraceId         string    `ch:"trace_id"`
	ErrorTypes      []string  `ch:"error_types"`
	ErrorMsgs       []string  `ch:"error_msgs"`
	ParentServices  []string  `ch:"parent_services"`
	ParentInstances []string  `ch:"parent_instances"`
	ParentTraced    []bool    `ch:"parent_traced"`
	ChildServices   []string  `ch:"child_services"`
	ChildInstances  []string  `ch:"child_instances"`
	ChildTraced     []bool    `ch:"child_traced"`
}

type ErrorPropation struct {
	Timestamp       time.Time  `ch:"timestamp"`
	EntryService    string     `ch:"entry_service"`
	EntryUrl        string     `ch:"entry_url"`
	EntrySpanId     string     `ch:"entry_span_id"`
	TraceId         string     `ch:"trace_id"`
	NodesService    []string   `ch:"nodes.service"`
	NodesInstance   []string   `ch:"nodes.instance"`
	NodesUrl        []string   `ch:"nodes.url"`
	NodesIsTraced   []bool     `ch:"nodes.is_traced"`
	NodesIsError    []bool     `ch:"nodes.is_error"`
	NodesErrorTypes [][]string `ch:"nodes.error_types"`
	NodesErrorMsgs  [][]string `ch:"nodes.error_msgs"`
	NodesDepth      []int      `ch:"nodes.depth"`
	NodesPath       []string   `ch:"nodes.path"`
}
