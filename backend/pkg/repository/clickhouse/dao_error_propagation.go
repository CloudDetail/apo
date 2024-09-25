package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

const (
	SQL_GET_INSTANCE_ERROR_PROPAGATION = `
		WITH found_trace_ids AS
		(
			SELECT error_propagation.timestamp as timestamp, error_propagation.trace_id as trace_id, error_propagation.entry_span_id as entry_span_id,
				nodes.service as service, nodes.instance as instance_id, nodes.path as path, nodes.depth as depth, nodes.error_types as error_types, nodes.error_msgs as error_msgs
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
			WHERE timestamp BETWEEN %d AND %d AND startsWith(found_trace_ids.path, nodes.path) AND nodes.depth=found_trace_ids.depth - 1 AND nodes.is_error = true
			GROUP BY trace_id
		) AS parent_node ON parent_node.trace_id = found_trace_ids.trace_id
		LEFT JOIN(
			SELECT error_propagation.trace_id as trace_id, groupArray(nodes.service) as child_services, groupArray(nodes.instance) as child_instances, groupArray(nodes.is_traced) as child_traced
			FROM error_propagation
			ARRAY JOIN nodes
			GLOBAL JOIN found_trace_ids ON error_propagation.trace_id = found_trace_ids.trace_id AND error_propagation.entry_span_id = found_trace_ids.entry_span_id
			WHERE timestamp BETWEEN %d AND %d AND startsWith(nodes.path, found_trace_ids.path) AND nodes.depth=found_trace_ids.depth+1 AND nodes.is_error = true
			GROUP BY trace_id
		) AS child_node on child_node.trace_id = found_trace_ids.trace_id
	`
)

// 查询实例相关的错误传播链
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
		Statement("LENGTH(nodes.error_types) > 0") // 返回的数据必须有ErrorTypes
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
