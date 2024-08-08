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
			FROM service_topology
			ARRAY JOIN nodes
			%s
			GROUP BY trace_id, path
			LIMIT 10000
		)
		SELECT nodes.service as service, nodes.url as endpoint, sum(case when nodes.is_traced then 1 else 0 end) > 0 as traced
		FROM service_topology
		ARRAY JOIN nodes
		JOIN found_trace_ids ON service_topology.trace_id = found_trace_ids.trace_id
		WHERE startsWith(nodes.path, found_trace_ids.path)
		AND nodes.path != found_trace_ids.path
		GROUP BY nodes.service, nodes.url
	`
)

// 查询所有子孙节点列表
func (ch *chRepo) ListDescendantNodes(req *request.GetDescendantMetricsRequest) ([]TopologyNode, error) {
	queryBuilder := NewQueryBuilder().
		Between("timestamp", req.StartTime/1000000, req.EndTime/1000000).
		Equals("nodes.service", req.Service).
		Equals("nodes.url", req.Endpoint).
		EqualsNotEmpty("entry_service", req.EntryService).
		EqualsNotEmpty("entry_url", req.EntryEndpoint)
	sql := fmt.Sprintf(SQL_GET_DESCENDANT_NODES, queryBuilder.String())
	var results []TopologyNode
	if err := ch.conn.Select(context.Background(), &results, sql, queryBuilder.values...); err != nil {
		return nil, err
	}
	return results, nil
}
