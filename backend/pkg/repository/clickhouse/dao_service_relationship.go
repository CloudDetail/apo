package clickhouse

import (
	"context"
	"fmt"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

const (
	SQL_GET_PARENT_NODES = `SELECT parent_service as parentService, parent_url as parentUrl, sum(case when flags['parent_traced'] then 1 else 0 end) > 0 as parentTraced,
		labels['client_group'] as clientGroup, labels['client_type'] as clientType, labels['client_peer'] as clientPeer, labels['client_key'] as clientKey
		FROM service_relationship
		%s
		GROUP BY parentService, parentUrl, clientGroup, clientType, clientPeer, clientKey
	`

	SQL_GET_CHILD_NODES = `SELECT service, url, sum(case when flags['is_traced'] then 1 else 0 end) > 0 as traced,
		labels['client_group'] as clientGroup, labels['client_type'] as clientType, labels['client_peer'] as clientPeer, labels['client_key'] as clientKey
		FROM service_relationship
		%s
		GROUP BY service, url, clientGroup, clientType, clientPeer, clientKey
	`

	SQL_GET_DESCENDANT_TOPOLOGY = `
		WITH found_trace_ids AS
		(
			SELECT trace_id, path , '' as empty_path
			FROM %s.service_relationship
			%s
			GROUP BY trace_id, path
			LIMIT 10000
		)
		SELECT parent_service as parentService, parent_url as parentUrl, service, url, sum(case when flags['is_traced'] then 1 else 0 end) > 0 as traced,
		labels['client_group'] as clientGroup, labels['client_type'] as clientType, labels['client_peer'] as clientPeer, labels['client_key'] as clientKey
		FROM service_relationship
		GLOBAL JOIN found_trace_ids ON service_relationship.trace_id = found_trace_ids.trace_id
		WHERE startsWith(service_relationship.path, found_trace_ids.path)
		AND service_relationship.path != found_trace_ids.path
		AND service_relationship.parent_service != found_trace_ids.empty_path
		GROUP BY parentService, parentUrl, service, url, clientGroup, clientType, clientPeer, clientKey
	`

	SQL_GET_ENTRY_NODES = `
		SELECT entry_service as service, entry_url as endpoint
		FROM service_relationship
		%s
		GROUP BY entry_service, entry_url
	`
)

// 查询上游节点列表
func (ch *chRepo) ListParentNodes(req *request.GetServiceEndpointTopologyRequest) (*model.TopologyNodes, error) {
	queryBuilder := NewQueryBuilder().
		Between("timestamp", req.StartTime/1000000, req.EndTime/1000000).
		Equals("service", req.Service).
		Equals("url", req.Endpoint).
		NotEquals("parentService", ""). // 过滤入口节点为空的数据
		NotEquals("clientGroup", "").   // 此处需保证能查询到 MQ -> A的数据
		EqualsNotEmpty("entry_service", req.EntryService).
		EqualsNotEmpty("entry_url", req.EntryEndpoint)

	results := []ParentNode{}
	sql := fmt.Sprintf(SQL_GET_PARENT_NODES, queryBuilder.String())
	if err := ch.conn.Select(context.Background(), &results, sql, queryBuilder.values...); err != nil {
		return nil, err
	}

	return getParentNodes(results), nil
}

// 查询下游对外调用列表
func (ch *chRepo) ListChildNodes(req *request.GetServiceEndpointTopologyRequest) (*model.TopologyNodes, error) {
	queryBuilder := NewQueryBuilder().
		Between("timestamp", req.StartTime/1000000, req.EndTime/1000000).
		Equals("parent_service", req.Service).
		Equals("parent_url", req.Endpoint).
		EqualsNotEmpty("entry_service", req.EntryService).
		EqualsNotEmpty("entry_url", req.EntryEndpoint)

	results := []ChildNode{}
	sql := fmt.Sprintf(SQL_GET_CHILD_NODES, queryBuilder.String())
	if err := ch.conn.Select(context.Background(), &results, sql, queryBuilder.values...); err != nil {
		return nil, err
	}

	return getChildNodes(results), nil
}

// 查询所有子孙节点列表
func (ch *chRepo) ListDescendantNodes(req *request.GetDescendantMetricsRequest) (*model.TopologyNodes, error) {
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
	return getDescendantNodes(results), nil
}

// 查询所有子孙的拓扑关系
func (ch *chRepo) ListDescendantRelations(req *request.GetServiceEndpointTopologyRequest) ([]*model.ToplogyRelation, error) {
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
	return getDescendantRelations(results), nil
}

// 查询相关入口节点列表
func (ch *chRepo) ListEntryEndpoints(req *request.GetServiceEntryEndpointsRequest) ([]EntryNode, error) {
	startTime := req.StartTime / 1000000
	endTime := req.EndTime / 1000000
	queryBuilder := NewQueryBuilder().
		Between("timestamp", startTime, endTime).
		Equals("service", req.Service).
		Equals("url", req.Endpoint)
	if !req.ShowMissTop {
		queryBuilder.Equals("miss_top", false)
	}
	results := []EntryNode{}
	sql := fmt.Sprintf(SQL_GET_ENTRY_NODES, queryBuilder.String())
	if err := ch.conn.Select(context.Background(), &results, sql, queryBuilder.values...); err != nil {
		return nil, err
	}
	return results, nil
}

type ParentNode struct {
	ParentService string `ch:"parentService"`
	ParentUrl     string `ch:"parentUrl"`
	ParentTraced  bool   `ch:"parentTraced"`
	ClientGroup   string `ch:"clientGroup"`
	ClientType    string `ch:"clientType"`
	ClientPeer    string `ch:"clientPeer"`
	ClientKey     string `ch:"clientKey"`
}

// 考虑2种场景
// MQ -> B
// A -> B
func getParentNodes(parentNodes []ParentNode) *model.TopologyNodes {
	result := model.NewTopologyNodes()
	if len(parentNodes) == 0 {
		return result
	}

	for _, parentNode := range parentNodes {
		if parentNode.ClientGroup == model.GROUP_MQ {
			result.AddTopologyNode(
				fmt.Sprintf("%s.%s", parentNode.ClientPeer, parentNode.ClientKey),
				parentNode.ClientPeer,
				parentNode.ClientKey,
				false,
				parentNode.ClientGroup,
				parentNode.ClientType,
			)
		} else if parentNode.ParentService != "" && parentNode.ParentUrl != "" {
			result.AddServerNode(
				fmt.Sprintf("%s.%s", parentNode.ParentService, parentNode.ParentUrl),
				parentNode.ParentService,
				parentNode.ParentUrl,
				parentNode.ParentTraced,
			)
		}
	}
	return result
}

type ChildNode struct {
	Service     string `ch:"service"`
	Url         string `ch:"url"`
	IsTraced    bool   `ch:"traced"`
	ClientGroup string `ch:"clientGroup"`
	ClientType  string `ch:"clientType"`
	ClientPeer  string `ch:"clientPeer"`
	ClientKey   string `ch:"clientKey"`
}

// 考虑2种场景
// A -> MQ
// A -> External -> B 存在B部分缺失，此时需补全为 A -> B
func getChildNodes(childNodes []ChildNode) *model.TopologyNodes {
	result := model.NewTopologyNodes()
	if len(childNodes) == 0 {
		return result
	}

	childMap := make(map[string]struct{})
	for _, childNode := range childNodes {
		if childNode.ClientGroup == model.GROUP_MQ {
			// MQ数据
			result.AddTopologyNode(
				fmt.Sprintf("%s.%s", childNode.ClientPeer, childNode.ClientKey),
				childNode.ClientPeer,
				childNode.ClientKey,
				childNode.IsTraced,
				childNode.ClientGroup,
				childNode.ClientType,
			)
		} else if childNode.Service != "" {
			// 已监控服务数据
			result.AddServerNode(
				fmt.Sprintf("%s.%s", childNode.Service, childNode.Url),
				childNode.Service,
				childNode.Url,
				childNode.IsTraced,
			)
			// 边 缓存已关联标识
			// 此处存在下游节点丢失可能，需尽可能补全该数据
			if childNode.ClientKey != "" {
				childMap[fmt.Sprintf("%s.%s", childNode.ClientPeer, childNode.ClientKey)] = struct{}{}
			}
		}
	}

	// 未监控服务
	for _, childNode := range childNodes {
		if childNode.Service == "" && childNode.ClientGroup != model.GROUP_MQ {
			key := fmt.Sprintf("%s.%s", childNode.ClientPeer, childNode.ClientKey)
			if _, exist := childMap[key]; !exist {
				result.AddTopologyNode(
					key,
					childNode.ClientPeer,
					childNode.ClientKey,
					childNode.IsTraced,
					childNode.ClientGroup,
					childNode.ClientType,
				)
			}
		}
	}
	return result
}

func getDescendantNodes(relations []ChildRelation) *model.TopologyNodes {
	result := model.NewTopologyNodes()
	if len(relations) == 0 {
		return result
	}

	childMap := make(map[string]struct{})
	for _, relation := range relations {
		if relation.ClientGroup == model.GROUP_MQ {
			// MQ数据 A -> MQ -> B, 需生成2个节点 MQ 和 B
			if relation.ClientKey != "" {
				result.AddTopologyNode(
					fmt.Sprintf("%s.%s", relation.ClientPeer, relation.ClientKey),
					relation.ClientPeer,
					relation.ClientKey,
					false,
					relation.ClientGroup,
					relation.ClientType,
				)
			}
			if relation.Service != "" {
				result.AddServerNode(
					fmt.Sprintf("%s.%s", relation.Service, relation.Url),
					relation.Service,
					relation.Url,
					relation.IsTraced,
				)
			}
		} else if relation.Service != "" {
			// 已监控服务数据
			// A -> B
			result.AddServerNode(
				fmt.Sprintf("%s.%s", relation.Service, relation.Url),
				relation.Service,
				relation.Url,
				relation.IsTraced,
			)
			// 存在 A -> B 有多个不同边的场景
			// A -> 边 缓存关系
			if relation.ClientKey != "" {
				childMap[relation.getParentClientKey()] = struct{}{}
			}
		}
	}

	// 未监控服务
	// 此处存在下游节点丢失可能，需尽可能补全该数据
	for _, relation := range relations {
		if relation.ClientKey != "" && relation.Service == "" && relation.ClientGroup != model.GROUP_MQ {
			// 读取 A -> 边缓存关系，剔除可能存在的脏数据
			key := relation.getParentClientKey()
			if _, exist := childMap[key]; !exist {
				// 未监控服务
				result.AddTopologyNode(
					key,
					relation.ClientPeer,
					relation.ClientKey,
					false,
					relation.ClientGroup,
					relation.ClientType,
				)
			}
		}
	}
	return result
}

type ChildRelation struct {
	ParentService string `ch:"parentService"`
	ParentUrl     string `ch:"parentUrl"`
	Service       string `ch:"service"`
	Url           string `ch:"url"`
	IsTraced      bool   `ch:"traced"`
	ClientGroup   string `ch:"clientGroup"`
	ClientType    string `ch:"clientType"`
	ClientPeer    string `ch:"clientPeer"`
	ClientKey     string `ch:"clientKey"`
}

func (relation *ChildRelation) getParentClientKey() string {
	return fmt.Sprintf("%s.%s.%s.%s", relation.ParentService, relation.ParentUrl, relation.ClientPeer, relation.ClientKey)
}

func (relation *ChildRelation) getParentCurrentKey() string {
	return fmt.Sprintf("%s.%s.%s.%s", relation.ParentService, relation.ParentUrl, relation.Service, relation.Url)
}

func getDescendantRelations(relations []ChildRelation) []*model.ToplogyRelation {
	result := make([]*model.ToplogyRelation, 0)
	if len(relations) == 0 {
		return result
	}

	relationMap := make(map[string]*model.ToplogyRelation)
	childMap := make(map[string]struct{}) // 剔除脏数据
	for _, relation := range relations {
		if relation.ClientGroup == model.GROUP_MQ {
			// MQ数据 需补充一条调用链路
			// A -> MQ
			// MQ -> B
			if relation.ParentService != "" {
				clientKey := relation.getParentClientKey()
				if _, exist := relationMap[clientKey]; !exist {
					relationMap[clientKey] = &model.ToplogyRelation{
						ParentService:  relation.ParentService,
						ParentEndpoint: relation.ParentUrl,
						Service:        relation.ClientPeer,
						Endpoint:       relation.ClientKey,
						IsTraced:       false,
						Group:          relation.ClientGroup,
						System:         relation.ClientType,
					}
				}
			}
			if relation.Service != "" {
				serverKey := fmt.Sprintf("%s.%s.%s.%s", relation.ClientPeer, relation.ClientKey, relation.Service, relation.Url)
				if _, exist := relationMap[serverKey]; !exist {
					relationMap[serverKey] = model.NewServerRelation(
						relation.ClientPeer,
						relation.ClientKey,
						relation.Service,
						relation.Url,
						relation.IsTraced,
					)
				}
			}
		} else if relation.ParentService != "" && relation.Service != "" {
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
			// 存在 A -> B 有多个不同边的场景
			// A -> 边 缓存关系
			childMap[relation.getParentClientKey()] = struct{}{}
		}
	}

	// 未监控服务
	// 此处存在下游节点丢失可能，需尽可能补全该数据
	for _, relation := range relations {
		if relation.ParentService != "" && relation.Service == "" && relation.ClientGroup != model.GROUP_MQ {
			// 读取 A -> 边缓存关系，剔除可能存在的脏数据
			key := relation.getParentClientKey()
			if _, exist := childMap[key]; !exist {
				// 未监控服务
				relationMap[key] = &model.ToplogyRelation{
					ParentService:  relation.ParentService,
					ParentEndpoint: relation.ParentUrl,
					Service:        relation.ClientPeer,
					Endpoint:       relation.ClientKey,
					IsTraced:       relation.IsTraced,
					Group:          relation.ClientGroup,
					System:         relation.ClientType,
				}
			}
		}
	}
	for _, topologyRelation := range relationMap {
		result = append(result, topologyRelation)
	}
	return result
}

type ServiceNode struct {
	Service  string `ch:"service" json:"service"`
	Endpoint string `ch:"url" json:"endpoint"`
	IsTraced bool   `ch:"traced" json:"isTraced"`
}

type EntryNode struct {
	Service  string `ch:"service" json:"service"`
	Endpoint string `ch:"endpoint" json:"endpoint"`
}
