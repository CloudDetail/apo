// File: dao_alert_analyzer_realtion.go
// Package: clickhouse
// Description: 包含了一些告警分析中使用到的结构和方法
package clickhouse

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

const (
	// 查询指定节点的入口信息,如果链路上存在异步请求,则返回异步入口的信息
	SQL_GET_ENTRY_RELATIONSHIP = `WITH
	  direct_relation AS (
	    SELECT trace_id,path,service,url,entry_service,entry_url
	    FROM %s.service_relationship
	    %s
	  ),
	  async_path AS (
	    SELECT path,service,url,trace_id
	    FROM service_relationship
	    GLOBAL JOIN (select DISTINCT trace_id from direct_relation) t ON service_relationship.trace_id = t.trace_id
	    WHERE flags['is_async'] = TRUE
	  )
	SELECT service,url,entry_service,entry_url,async_path.service as async_entry_service,async_path.url as async_entry_endpoint FROM direct_relation
	LEFT JOIN async_path ON direct_relation.trace_id = async_path.trace_id
	WHERE startsWith(path,async_path.path)
	GROUP BY service,url,entry_service,entry_url,async_entry_service,async_entry_endpoint`

	SQL_GET_ASYNC_ENTRY_NODE_RELATIONSHIP = `WITH
	  async_entry_path AS (
	    SELECT trace_id,path,service,url,entry_service,entry_url
	    FROM %s.service_relationship
	    %s
	  )
	SELECT parent_service as parentService, parent_url as parentUrl, service, url, sum(case when flags['is_traced'] then 1 else 0 end) > 0 as traced,
	labels['client_group'] as clientGroup, labels['client_type'] as clientType, labels['client_peer'] as clientPeer, labels['client_key'] as clientKey
	FROM %s.service_relationship r
	GLOBAL JOIN async_entry_path ON r.trace_id = async_entry_path.trace_id
	WHERE startsWith(path,async_entry_path.path)
	GROUP BY parentService, parentUrl, service, url, clientGroup, clientType, clientPeer, clientKey`
)

// EntryRelationship 表示下游节点和其入口的直接对应关系
type EntryRelationship struct {
	Service  string `ch:"service"`
	Endpoint string `ch:"url"`

	EntryService  string `ch:"entry_service"`
	EntryEndpoint string `ch:"entry_url"`

	AsyncEntryService  string `ch:"async_entry_service"`
	AsyncEntryEndpoint string `ch:"async_entry_endpoint"`
}

func (r *EntryRelationship) GetEntryService() string {
	if len(r.AsyncEntryService) > 0 {
		return r.AsyncEntryEndpoint
	}

	return r.EntryService
}

func (r *EntryRelationship) GetEntryEndpoint() string {
	if len(r.AsyncEntryService) > 0 {
		return r.AsyncEntryEndpoint
	}
	return r.EntryEndpoint
}

// 和ServiceNode结构完全相同
// 作为查询条件时,允许Endpoint为空
// 表示对应Service下的所有Endpoint
type AlertService struct {
	Service  string `ch:"service" json:"serviceName"`
	Endpoint string `ch:"url" json:"endpoint"`
}

// SearchEntryEndpointsByAlertService 根据下游节点查询入口节点
// 链路中存在异步节点时,不保留真正的入口节点
func (ch *chRepo) SearchEntryEndpointsByAlertService(alertServices []AlertService, startTime, endTime int64) ([]EntryRelationship, error) {
	// microseconds -> seconds
	startTime = startTime / 1000000
	endTime = endTime / 1000000

	// services中可能包含两类数据,contentKey为空时,表示忽略contentKey
	var endpoints = ValueInGroups{
		Keys: []string{"service", "url"},
	}
	var services = ValueInGroups{
		Keys: []string{"service"},
	}

	for _, endpoint := range alertServices {
		if len(endpoint.Endpoint) > 0 {
			endpoints.ValueGroups = append(endpoints.ValueGroups, clickhouse.GroupSet{
				Value: []any{endpoint.Service, endpoint.Endpoint},
			})
		} else {
			services.ValueGroups = append(services.ValueGroups, clickhouse.GroupSet{
				Value: []any{endpoint.Service},
			})
		}
	}

	queryBuilder := NewQueryBuilder().
		Between("timestamp", startTime, endTime).
		Equals("miss_top", false).
		And(MergeWheres(OrSep, InGroup(endpoints), InGroup(services)))

	parentTopologys := []EntryRelationship{}
	sql := fmt.Sprintf(SQL_GET_ENTRY_RELATIONSHIP,
		ch.database,
		queryBuilder.String(),
	)
	if err := ch.conn.Select(context.Background(), &parentTopologys, sql, queryBuilder.values...); err != nil {
		return nil, err
	}

	return parentTopologys, nil
}

func (ch *chRepo) ListDescendantRelationsWithoutEdge(req *request.GetServiceEndpointTopologyRequest) ([]*model.ToplogyRelation, error) {
	startTime := req.StartTime / 1000000
	endTime := req.EndTime / 1000000
	queryBuilder := NewQueryBuilder().
		Between("timestamp", startTime, endTime).
		Equals("service", req.Service).
		Equals("url", req.Endpoint).
		EqualsNotEmpty("entry_service", req.EntryService).
		EqualsNotEmpty("entry_url", req.EntryEndpoint)
	sql := fmt.Sprintf(SQL_GET_DESCENDANT_TOPOLOGY, ch.database, queryBuilder.String())
	results := []ChildRelation{}
	if err := ch.conn.Select(context.Background(), &results, sql, queryBuilder.values...); err != nil {
		return nil, err
	}
	return getDescendantServiceNode(results), nil
}

func getDescendantServiceNode(relations []ChildRelation) []*model.ToplogyRelation {
	result := make([]*model.ToplogyRelation, 0)
	if len(relations) == 0 {
		return result
	}

	relationMap := make(map[string]*model.ToplogyRelation)
	for _, relation := range relations {
		if relation.ParentService != "" && relation.Service != "" {
			// 已监控服务数据
			// A -> B
			key := relation.getParentCurrentKey()
			if _, exist := relationMap[key]; !exist {
				relationMap[key] = model.NewServerRelation(
					relation.ParentService,
					relation.ParentUrl,
					relation.Service,
					relation.Url,
					relation.IsTraced,
				)
			}
		}
	}

	for _, topologyRelation := range relationMap {
		result = append(result, topologyRelation)
	}
	return result
}

// ListDescendantNodesForAsyncEntry 获取异步入口的下游依赖拓扑
func (ch *chRepo) ListDescendantNodesForAsyncEntry(req *request.GetServiceEndpointTopologyRequest) ([]*model.ToplogyRelation, error) {
	startTime := req.StartTime / 1000000
	endTime := req.EndTime / 1000000
	queryBuilder := NewQueryBuilder().
		Between("timestamp", startTime, endTime).
		Equals("service", req.EntryService).
		Equals("url", req.EntryEndpoint).
		Equals("flags['is_async']", true)
	sql := fmt.Sprintf(SQL_GET_ASYNC_ENTRY_NODE_RELATIONSHIP, ch.database, queryBuilder.String(), ch.database)
	results := []ChildRelation{}
	if err := ch.conn.Select(context.Background(), &results, sql, queryBuilder.values...); err != nil {
		return nil, err
	}
	return getDescendantServiceNode(results), nil
}
