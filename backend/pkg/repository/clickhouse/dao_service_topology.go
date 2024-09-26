package clickhouse

import (
	"context"
	"fmt"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

const (
	SQL_GET_DESCENDANT_NODES = `
		WITH found_trace_ids AS
		(
			SELECT trace_id, nodes.path as path
			FROM %s.service_topology
			ARRAY JOIN nodes
			%s
			GROUP BY trace_id, path
			LIMIT 10000
		)
		SELECT nodes.service as service, nodes.url as endpoint, sum(case when nodes.is_traced then 1 else 0 end) > 0 as traced
		FROM service_topology
		ARRAY JOIN nodes
		GLOBAL JOIN found_trace_ids ON service_topology.trace_id = found_trace_ids.trace_id
		WHERE timestamp BETWEEN %d AND %d AND startsWith(nodes.path, found_trace_ids.path)
		AND nodes.path != found_trace_ids.path
		GROUP BY nodes.service, nodes.url
	`

	SQL_GET_DESCENDANT_TOPOLOGY = `
		WITH found_trace_ids AS
		(
			SELECT trace_id, nodes.path as path
			FROM %s.service_topology
			ARRAY JOIN nodes
			%s
			GROUP BY trace_id, path
			LIMIT 10000
		)
		SELECT nodes.service as service, nodes.url as endpoint, nodes.parent_service as p_service, nodes.parent_url as p_endpoint, sum(case when nodes.is_traced then 1 else 0 end) > 0 as traced
		FROM service_topology
		ARRAY JOIN nodes
		GLOBAL JOIN found_trace_ids ON service_topology.trace_id = found_trace_ids.trace_id
		WHERE timestamp BETWEEN %d AND %d AND startsWith(nodes.path, found_trace_ids.path)
		AND nodes.path != found_trace_ids.path
		AND nodes.parent_service != ''
		GROUP BY nodes.service, nodes.url, nodes.parent_service, nodes.parent_url
	`

	SQL_GET_ENTRY_NODES = `
		SELECT entry_service as service, entry_url as endpoint
			FROM service_topology
			ARRAY JOIN nodes
			%s
			GROUP BY entry_service, entry_url
	`
)

// 查询所有子孙节点列表
func (ch *chRepo) ListDescendantNodes(req *request.GetDescendantMetricsRequest) ([]TopologyNode, error) {
	startTime := req.StartTime / 1000000
	endTime := req.EndTime / 1000000
	queryBuilder := NewQueryBuilder().
		Between("timestamp", startTime, endTime).
		Equals("nodes.service", req.Service).
		Equals("nodes.url", req.Endpoint).
		EqualsNotEmpty("entry_service", req.EntryService).
		EqualsNotEmpty("entry_url", req.EntryEndpoint)
	sql := fmt.Sprintf(SQL_GET_DESCENDANT_NODES, ch.database, queryBuilder.String(), startTime, endTime)
	results := []TopologyNode{}
	if err := ch.conn.Select(context.Background(), &results, sql, queryBuilder.values...); err != nil {
		return nil, err
	}
	return results, nil
}

// 查询所有子孙的拓扑关系
func (ch *chRepo) ListDescendantRelations(req *request.GetServiceEndpointTopologyRequest) ([]ToplogyRelation, error) {
	startTime := req.StartTime / 1000000
	endTime := req.EndTime / 1000000
	queryBuilder := NewQueryBuilder().
		Between("timestamp", startTime, endTime).
		Equals("nodes.service", req.Service).
		Equals("nodes.url", req.Endpoint).
		EqualsNotEmpty("entry_service", req.EntryService).
		EqualsNotEmpty("entry_url", req.EntryEndpoint)
	sql := fmt.Sprintf(SQL_GET_DESCENDANT_TOPOLOGY, ch.database, queryBuilder.String(), startTime, endTime)
	results := []ToplogyRelation{}
	if err := ch.conn.Select(context.Background(), &results, sql, queryBuilder.values...); err != nil {
		return nil, err
	}
	return results, nil
}

// 查询相关入口节点列表
func (ch *chRepo) ListEntryEndpoints(req *request.GetServiceEntryEndpointsRequest) ([]EntryNode, error) {
	startTime := req.StartTime / 1000000
	endTime := req.EndTime / 1000000
	queryBuilder := NewQueryBuilder().
		Between("timestamp", startTime, endTime).
		Equals("nodes.service", req.Service).
		Equals("nodes.url", req.Endpoint)
	results := []EntryNode{}
	sql := fmt.Sprintf(SQL_GET_ENTRY_NODES, queryBuilder.String())
	if err := ch.conn.Select(context.Background(), &results, sql, queryBuilder.values...); err != nil {
		return nil, err
	}
	return results, nil
}

type ToplogyRelation struct {
	ParentService  string `ch:"p_service" json:"parentService"`
	ParentEndpoint string `ch:"p_endpoint" json:"parentEndpoint"`
	Service        string `ch:"service" json:"service"`
	Endpoint       string `ch:"endpoint" json:"endpoint"`
	IsTraced       bool   `ch:"traced" json:"isTraced"`
}

type EntryNode struct {
	Service  string `ch:"service" json:"service"`
	Endpoint string `ch:"endpoint" json:"endpoint"`
}
