// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	core "github.com/CloudDetail/apo/backend/pkg/core"
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

	SQL_GET_UPSTREAM_RELATIONSHIP = `WITH found_trace_ids AS
  (
    SELECT trace_id, path , service, url, '' as empty_path
    FROM service_relationship
    %s
    GROUP BY trace_id, path, service, url
    LIMIT 1000
  )
  SELECT sr.service as parent_service,sr.url as parent_url, ft.service as service, ft.url as url,
  	min(length(splitByChar('.', service_relationship.path))) AS depth
  FROM service_relationship sr
  GLOBAL JOIN found_trace_ids ft ON sr.trace_id = ft.trace_id
  WHERE startsWith(ft.path, sr.path)
  AND sr.parent_service != ft.empty_path
  group by parent_service,parent_url,service,url`
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
	// TODO 后续支持异步入口
	// if len(r.AsyncEntryService) > 0 {
	// 	return r.AsyncEntryEndpoint
	// }

	return r.EntryService
}

func (r *EntryRelationship) GetEntryEndpoint() string {
	// TODO 后续支持异步入口
	// if len(r.AsyncEntryService) > 0 {
	// 	return r.AsyncEntryEndpoint
	// }
	return r.EntryEndpoint
}

// 和ServiceNode结构完全相同
// 作为查询条件时,允许Endpoint为空
// 表示对应Service下的所有Endpoint
type AlertService struct {
	Service  string `json:"serviceName" ch:"service_name"`
	Endpoint string `json:"endpoint" ch:"endpoint"`

	DatabaseURL  string `json:"dbURL"`
	DatabaseIP   string `json:"dbIP"`
	DatabasePort string `json:"dbPort"`
}

func (ch *chRepo) SearchEntryEndpointsByAlertServiceInGroup(
	ctx core.Context,
	alertServices []AlertService,
	startTime, endTime int64,
	groupSvcList []string,
) ([]EntryRelationship, error) {
	if len(groupSvcList) == 0 {
		return ch.SearchEntryEndpointsByAlertService(ctx, alertServices, startTime, endTime)
	}

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
		if len(endpoint.DatabaseURL) > 0 {
			// 跳过数据库节点
			continue
		}
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
		And(mergeWheres(OrSep, inGroup(endpoints), inGroup(services)))

	relations := []ServiceNodeRelationWithDepth{}
	sql := fmt.Sprintf(SQL_GET_UPSTREAM_RELATIONSHIP, queryBuilder.String())
	if err := ch.GetContextDB(ctx).Select(ctx.GetContext(), &relations, sql, queryBuilder.values...); err != nil {
		return nil, err
	}

	var tmpGroupSvcMap = map[string]struct{}{}
	for _, svc := range groupSvcList {
		tmpGroupSvcMap[svc] = struct{}{}
	}

	type entryInGroup struct {
		relations []EntryRelationship
		depth     uint64
	}

	var entryInGroups = make(map[string]*entryInGroup, 0)
	for i := 0; i < len(relations); i++ {
		key := fmt.Sprintf("%s-%s", relations[i].Service, relations[i].Endpoint)
		entries, find := entryInGroups[key]
		if !find {
			entries = &entryInGroup{
				relations: []EntryRelationship{},
				depth:     999,
			}
			entryInGroups[key] = entries
		}

		if relations[i].Depth > entries.depth {
			continue
		}
		if _, ok := tmpGroupSvcMap[relations[i].ParentService]; !ok {
			continue
		}
		if relations[i].Depth < entries.depth {
			entries.depth = relations[i].Depth
			entries.relations = make([]EntryRelationship, 0)
		}
		entries.relations = append(entries.relations, EntryRelationship{
			Service:            relations[i].Service,
			Endpoint:           relations[i].Endpoint,
			EntryService:       relations[i].ParentService,
			EntryEndpoint:      relations[i].ParentEndpoint,
			AsyncEntryService:  "",
			AsyncEntryEndpoint: "",
		})
	}

	var res []EntryRelationship
	for _, entries := range entryInGroups {
		res = append(res, entries.relations...)
	}

	return res, nil
}

// SearchEntryEndpointsByAlertService 根据下游节点查询入口节点
// 链路中存在异步节点时,不保留真正的入口节点
func (ch *chRepo) SearchEntryEndpointsByAlertService(ctx core.Context, alertServices []AlertService, startTime, endTime int64) ([]EntryRelationship, error) {
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
		if len(endpoint.DatabaseURL) > 0 {
			// 跳过数据库节点
			continue
		}
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
		And(mergeWheres(OrSep, inGroup(endpoints), inGroup(services)))

	parentTopologys := []EntryRelationship{}
	sql := fmt.Sprintf(SQL_GET_ENTRY_RELATIONSHIP,
		ch.database,
		queryBuilder.String(),
	)
	if err := ch.GetContextDB(ctx).Select(ctx.GetContext(), &parentTopologys, sql, queryBuilder.values...); err != nil {
		return nil, err
	}

	return parentTopologys, nil
}

func (ch *chRepo) ListDescendantRelationsWithoutEdge(ctx core.Context, req *request.GetServiceEndpointTopologyRequest) ([]*model.TopologyRelation, error) {
	startTime := req.StartTime / 1000000
	endTime := req.EndTime / 1000000
	queryBuilder := NewQueryBuilder().
		Between("timestamp", startTime, endTime).
		Equals("service", req.Service).
		Equals("url", req.Endpoint).
		EqualsNotEmpty("entry_service", req.EntryService).
		EqualsNotEmpty("entry_url", req.EntryEndpoint)
	sql := fmt.Sprintf(SQL_GET_DESCENDANT_TOPOLOGY, ch.database, queryBuilder.String(), startTime, endTime)
	results := []ChildRelation{}
	if err := ch.GetContextDB(ctx).Select(ctx.GetContext(), &results, sql, queryBuilder.values...); err != nil {
		return nil, err
	}
	return getDescendantServiceNode(results), nil
}

func getDescendantServiceNode(relations []ChildRelation) []*model.TopologyRelation {
	result := make([]*model.TopologyRelation, 0)
	if len(relations) == 0 {
		return result
	}

	relationMap := make(map[string]*model.TopologyRelation)
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
