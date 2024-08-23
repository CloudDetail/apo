package clickhouse

import (
	"context"
	"fmt"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

const (
	SQL_GET_PARENT_NODES = `SELECT left_service as service, left_url as endpoint, sum(case when left_traced then 1 else 0 end) > 0 as traced
		FROM service_relation
		%s
		GROUP BY left_service, left_url
	`
	SQL_GET_CHILD_NODES = `SELECT right_service as service, right_url as endpoint, sum(case when right_traced then 1 else 0 end) > 0 as traced
		FROM service_relation
		%s
		GROUP BY right_service, right_url
	`
)

// 查询上游节点列表
func (ch *chRepo) ListParentNodes(req *request.GetServiceEndpointTopologyRequest) ([]TopologyNode, error) {
	queryBuilder := NewQueryBuilder().
		Between("timestamp", req.StartTime/1000000, req.EndTime/1000000).
		Equals("right_service", req.Service).
		Equals("right_url", req.Endpoint).
		EqualsNotEmpty("entry_service", req.EntryService).
		EqualsNotEmpty("entry_url", req.EntryEndpoint)

	results := []TopologyNode{}
	sql := fmt.Sprintf(SQL_GET_PARENT_NODES, queryBuilder.String())
	if err := ch.conn.Select(context.Background(), &results, sql, queryBuilder.values...); err != nil {
		return nil, err
	}
	return results, nil
}

// 查询下游节点列表
func (ch *chRepo) ListChildNodes(req *request.GetServiceEndpointTopologyRequest) ([]TopologyNode, error) {
	queryBuilder := NewQueryBuilder().
		Between("timestamp", req.StartTime/1000000, req.EndTime/1000000).
		Equals("left_service", req.Service).
		Equals("left_url", req.Endpoint).
		EqualsNotEmpty("entry_service", req.EntryService).
		EqualsNotEmpty("entry_url", req.EntryEndpoint)

	results := []TopologyNode{}
	sql := fmt.Sprintf(SQL_GET_CHILD_NODES, queryBuilder.String())
	if err := ch.conn.Select(context.Background(), &results, sql, queryBuilder.values...); err != nil {
		return nil, err
	}
	return results, nil
}

type TopologyNode struct {
	Service  string `ch:"service" json:"service"`
	Endpoint string `ch:"endpoint" json:"endpoint"`
	IsTraced bool   `ch:"traced" json:"isTraced"`
}
