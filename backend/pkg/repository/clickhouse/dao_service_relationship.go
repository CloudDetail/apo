package clickhouse

import (
	"context"
	"fmt"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

const (
	GROUP_MQ      = "mq"
	GROUP_SERVICE = "service"

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

	SQL_GET_DESCENDANT_NODES = `
		WITH found_trace_ids AS
		(
			SELECT trace_id, path
			FROM %s.service_relationship
			%s
			GROUP BY trace_id, path
			LIMIT 10000
		)
		SELECT service, url, sum(case when flags['is_traced'] then 1 else 0 end) > 0 as traced
		FROM service_relationship
		GLOBAL JOIN found_trace_ids ON service_relationship.trace_id = found_trace_ids.trace_id
		WHERE startsWith(service_relationship.path, found_trace_ids.path)
		AND service_relationship.path != found_trace_ids.path
		GROUP BY service, url
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
func (ch *chRepo) ListParentNodes(req *request.GetServiceEndpointTopologyRequest) ([]*TopologyNode, error) {
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
func (ch *chRepo) ListChildNodes(req *request.GetServiceEndpointTopologyRequest) ([]*TopologyNode, error) {
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
func (ch *chRepo) ListDescendantNodes(req *request.GetDescendantMetricsRequest) ([]ServiceNode, error) {
	startTime := req.StartTime / 1000000
	endTime := req.EndTime / 1000000
	queryBuilder := NewQueryBuilder().
		Between("timestamp", startTime, endTime).
		Equals("service", req.Service).
		Equals("url", req.Endpoint).
		EqualsNotEmpty("entry_service", req.EntryService).
		EqualsNotEmpty("entry_url", req.EntryEndpoint)
	sql := fmt.Sprintf(SQL_GET_DESCENDANT_NODES, ch.database, queryBuilder.String())
	results := []ServiceNode{}
	if err := ch.conn.Select(context.Background(), &results, sql, queryBuilder.values...); err != nil {
		return nil, err
	}
	return results, nil
}

// 查询所有子孙的拓扑关系
func (ch *chRepo) ListDescendantRelations(req *request.GetServiceEndpointTopologyRequest) ([]*ToplogyRelation, error) {
	startTime := req.StartTime / 1000000
	endTime := req.EndTime / 1000000
	queryBuilder := NewQueryBuilder().
		Between("timestamp", startTime, endTime).
		Equals("service", req.Service).
		Equals("url", req.Endpoint).
		EqualsNotEmpty("entry_service", req.EntryService).
		EqualsNotEmpty("entry_url", req.EntryEndpoint)
	sql := fmt.Sprintf(SQL_GET_DESCENDANT_TOPOLOGY, ch.database, queryBuilder.String())
	results := []ChildRealtion{}
	if err := ch.conn.Select(context.Background(), &results, sql, queryBuilder.values...); err != nil {
		return nil, err
	}
	return getChildRelations(results), nil
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
func getParentNodes(parentNodes []ParentNode) []*TopologyNode {
	result := make([]*TopologyNode, 0)
	if len(parentNodes) == 0 {
		return result
	}

	parentMap := make(map[string]*TopologyNode)
	for _, parentNode := range parentNodes {
		if parentNode.ClientGroup == GROUP_MQ {
			key := fmt.Sprintf("%s.%s", parentNode.ClientPeer, parentNode.ClientKey)
			if _, exist := parentMap[key]; !exist {
				parentMap[key] = &TopologyNode{
					Service:  parentNode.ClientPeer,
					Endpoint: parentNode.ClientKey,
					IsTraced: false,
					Group:    parentNode.ClientGroup,
					System:   parentNode.ClientType,
				}
			}
		} else if parentNode.ParentService != "" && parentNode.ParentUrl != "" {
			// 存在 A -> B多次，但边不同
			key := fmt.Sprintf("%s.%s", parentNode.ParentService, parentNode.ParentUrl)
			if _, exist := parentMap[key]; !exist {
				parentMap[key] = &TopologyNode{
					Service:  parentNode.ParentService,
					Endpoint: parentNode.ParentUrl,
					IsTraced: parentNode.ParentTraced,
					Group:    GROUP_SERVICE,
					System:   parentNode.ClientType,
				}
			}
		}
	}

	for _, topologyNode := range parentMap {
		result = append(result, topologyNode)
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
func getChildNodes(childNodes []ChildNode) []*TopologyNode {
	result := make([]*TopologyNode, 0)
	if len(childNodes) == 0 {
		return result
	}

	childMap := make(map[string]*TopologyNode)
	for _, childNode := range childNodes {
		key := fmt.Sprintf("%s.%s", childNode.ClientPeer, childNode.ClientKey)
		if childNode.ClientGroup == GROUP_MQ {
			// MQ数据
			if _, exist := childMap[key]; !exist {
				childMap[key] = &TopologyNode{
					Service:  childNode.ClientPeer,
					Endpoint: childNode.ClientKey,
					IsTraced: childNode.IsTraced,
					Group:    childNode.ClientGroup,
					System:   childNode.ClientType,
				}
			}
		} else if childNode.Service != "" {
			// 已监控服务数据
			childMap[key] = &TopologyNode{
				Service:  childNode.Service,
				Endpoint: childNode.Url,
				IsTraced: childNode.IsTraced,
				Group:    GROUP_SERVICE,
				System:   childNode.ClientType,
			}
		}
	}

	// 未监控服务
	// 此处存在下游节点丢失可能，需尽可能补全该数据
	for _, childNode := range childNodes {
		if childNode.Service == "" && childNode.ClientGroup != GROUP_MQ {
			key := fmt.Sprintf("%s.%s", childNode.ClientPeer, childNode.ClientKey)
			if _, exist := childMap[key]; !exist {
				// 未监控服务
				childMap[key] = &TopologyNode{
					Service:  childNode.ClientPeer,
					Endpoint: childNode.ClientKey,
					IsTraced: childNode.IsTraced,
					Group:    childNode.ClientGroup,
					System:   childNode.ClientType,
				}
			}
		}
	}
	for _, topologyNode := range childMap {
		result = append(result, topologyNode)
	}
	return result
}

type ChildRealtion struct {
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

func getChildRelations(relations []ChildRealtion) []*ToplogyRelation {
	result := make([]*ToplogyRelation, 0)
	if len(relations) == 0 {
		return result
	}

	childMap := make(map[string]*ToplogyRelation)
	for _, relation := range relations {
		if relation.ClientGroup == GROUP_MQ {
			// MQ数据 需补充一条调用链路
			// A -> MQ
			// MQ -> B
			if relation.ParentService != "" {
				clientKey := fmt.Sprintf("%s.%s.%s.%s", relation.ParentService, relation.ParentUrl, relation.ClientPeer, relation.ClientKey)
				if _, exist := childMap[clientKey]; !exist {
					childMap[clientKey] = &ToplogyRelation{
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
				if _, exist := childMap[serverKey]; !exist {
					childMap[serverKey] = &ToplogyRelation{
						ParentService:  relation.ClientPeer,
						ParentEndpoint: relation.ClientKey,
						Service:        relation.Service,
						Endpoint:       relation.Url,
						IsTraced:       relation.IsTraced,
						Group:          GROUP_SERVICE,
						System:         relation.ClientType,
					}
				}
			}
		} else if relation.ParentService != "" && relation.Service != "" {
			// 已监控服务数据
			key := fmt.Sprintf("%s.%s.%s.%s", relation.ParentService, relation.ParentUrl, relation.ClientPeer, relation.ClientKey)
			childMap[key] = &ToplogyRelation{
				ParentService:  relation.ParentService,
				ParentEndpoint: relation.ParentUrl,
				Service:        relation.Service,
				Endpoint:       relation.Url,
				IsTraced:       relation.IsTraced,
				Group:          GROUP_SERVICE,
				System:         relation.ClientType,
			}
		}
	}

	// 未监控服务
	// 此处存在下游节点丢失可能，需尽可能补全该数据
	for _, relation := range relations {
		if relation.ParentService != "" && relation.Service == "" && relation.ClientGroup != GROUP_MQ {
			key := fmt.Sprintf("%s.%s.%s.%s", relation.ParentService, relation.ParentUrl, relation.ClientPeer, relation.ClientKey)
			if _, exist := childMap[key]; !exist {
				// 未监控服务
				childMap[key] = &ToplogyRelation{
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
	for _, topologyRelation := range childMap {
		result = append(result, topologyRelation)
	}
	return result
}

type TopologyNode struct {
	Service  string `json:"service"`
	Endpoint string `json:"endpoint"`
	IsTraced bool   `json:"isTraced"`
	Group    string `json:"group"`
	System   string `json:"system"`
}

func NewTopologyNode(service string, url string) *TopologyNode {
	return &TopologyNode{
		Service:  service,
		Endpoint: url,
		IsTraced: true,
		Group:    GROUP_SERVICE,
		System:   "",
	}
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

type ToplogyRelation struct {
	ParentService  string `json:"parentService"`
	ParentEndpoint string `json:"parentEndpoint"`
	Service        string `json:"service"`
	Endpoint       string `json:"endpoint"`
	IsTraced       bool   `json:"isTraced"`
	Group          string `json:"group"`
	System         string `json:"system"`
}
