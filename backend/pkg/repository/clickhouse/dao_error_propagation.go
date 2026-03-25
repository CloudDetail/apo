// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

const (
	SQL_GET_INSTANCE_ERROR_PROPAGATION = `SELECT
    sample.1 AS timestamp,
    sample.2 AS service,
    sample.3 AS instance_id,
    sample.4 AS trace_id,
    sample.5 AS error_types,
    sample.6 AS error_msgs,
    arrayMap(x -> x.1, arrayFilter(x -> (x.3 = sample.7 - 1 AND startsWith(sample.8, x.2)), all_nodes)) AS parent_services,
    arrayMap(x -> x.4, arrayFilter(x -> (x.3 = sample.7 - 1 AND startsWith(sample.8, x.2)), all_nodes)) AS parent_instances,
    arrayMap(x -> x.5, arrayFilter(x -> (x.3 = sample.7 - 1 AND startsWith(sample.8, x.2)), all_nodes)) AS parent_traced,
    arrayMap(x -> x.1, arrayFilter(x -> (x.3 = sample.7 + 1 AND startsWith(x.2, sample.8)), all_nodes)) AS child_services,
    arrayMap(x -> x.4, arrayFilter(x -> (x.3 = sample.7 + 1 AND startsWith(x.2, sample.8)), all_nodes)) AS child_instances,
    arrayMap(x -> x.5, arrayFilter(x -> (x.3 = sample.7 + 1 AND startsWith(x.2, sample.8)), all_nodes)) AS child_traced
FROM (
    SELECT
        trace_id,
        groupArray((s, p, d, i, t)) AS all_nodes,
        groupArrayIf((timestamp, s, i, trace_id, err_t, err_m, d, p), is_err = 1) AS target_samples
    FROM (
        SELECT
            trace_id,
            timestamp,
            nodes.service AS s,
            nodes.path AS p,
            nodes.depth AS d,
            nodes.instance AS i,
            nodes.is_traced AS t,
            nodes.is_error AS is_err,
            nodes.error_types AS err_t,
            nodes.error_msgs AS err_m
        FROM error_propagation
        ARRAY JOIN nodes
        WHERE trace_id GLOBAL IN (
            SELECT trace_id FROM (
                SELECT trace_id, timestamp, nodes.service, nodes.url,nodes.is_traced, nodes.is_error, entry_service, entry_url , nodes.error_types
                FROM error_propagation
                ARRAY JOIN nodes
            )
            %s
            ORDER BY timestamp DESC
            LIMIT 1000
        )
    )
    GROUP BY trace_id
)
ARRAY JOIN target_samples AS sample
ORDER BY timestamp DESC
LIMIT 1000`

	SQL_GET_ERROR_PROPAGATION = `SELECT timestamp,entry_service,entry_url,entry_span_id,trace_id
		, nodes.service,nodes.instance,nodes.url,nodes.is_traced,nodes.is_error
		, nodes.error_types,nodes.error_msgs,nodes.depth,nodes.path
		FROM %s.error_propagation
		%s %s
	`
)

func (ch *chRepo) ListErrorByEntryService(ctx core.Context, startTime, endTime int64, entryService, entryEndpoint string, endpoints []model.EndpointKey) ([]ErrorPropation, error) {
	startTime = startTime / 1000000
	endTime = endTime / 1000000

	whereEntry := mergeWheres(
		AndSep,
		equals("entry_service", entryService),
		equals("entry_url", entryEndpoint),
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
		whereEntry = mergeWheres(
			OrSep,
			whereEntry,
			inGroup(whereEndpoints),
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
	if err := ch.GetContextDB(ctx).Select(ctx.GetContext(), &results, sql, queryBuilder.values...); err != nil {
		return nil, err
	}
	return results, nil
}

func (ch *chRepo) ListErrorPropagation(ctx core.Context, req *request.GetErrorInstanceRequest) ([]ErrorInstancePropagation, error) {
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
	var results []ErrorInstancePropagation
	sql := fmt.Sprintf(SQL_GET_INSTANCE_ERROR_PROPAGATION, queryBuilder.String())
	if err := ch.GetContextDB(ctx).Select(ctx.GetContext(), &results, sql, queryBuilder.values...); err != nil {
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
