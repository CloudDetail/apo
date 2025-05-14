// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"context"
	"fmt"

	core "github.com/CloudDetail/apo/backend/pkg/core"
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
		SELECT entry_service, entry_url
		FROM service_relationship
		%s
		GROUP BY entry_service, entry_url
	`
)

// Query the list of upstream nodes
func (ch *chRepo) ListParentNodes(ctx core.Context, req *request.GetServiceEndpointTopologyRequest) (*model.TopologyNodes, error) {
	queryBuilder := NewQueryBuilder().
		Between("timestamp", req.StartTime/1000000, req.EndTime/1000000).
		Equals("service", req.Service).
		Equals("url", req.Endpoint).
		NotEquals("parentService", ""). // Filter data with empty entry node
		NotEquals("clientGroup", "").   // Ensure that the data of MQ -> A can be queried here.
		EqualsNotEmpty("entry_service", req.EntryService).
		EqualsNotEmpty("entry_url", req.EntryEndpoint)

	results := []ParentNode{}
	sql := fmt.Sprintf(SQL_GET_PARENT_NODES, queryBuilder.String())
	if err := ch.conn.Select(context.Background(), &results, sql, queryBuilder.values...); err != nil {
		return nil, err
	}

	return getParentNodes(results), nil
}

// Query the downstream outbound call list
func (ch *chRepo) ListChildNodes(ctx core.Context, req *request.GetServiceEndpointTopologyRequest) (*model.TopologyNodes, error) {
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

// Query the list of all descendant nodes
func (ch *chRepo) ListDescendantNodes(ctx core.Context, req *request.GetDescendantMetricsRequest) (*model.TopologyNodes, error) {
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

// Query the topological relationships of all descendants
func (ch *chRepo) ListDescendantRelations(ctx core.Context, req *request.GetServiceEndpointTopologyRequest) ([]*model.ToplogyRelation, error) {
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

// Query the list of related entry nodes
func (ch *chRepo) ListEntryEndpoints(ctx core.Context, req *request.GetServiceEntryEndpointsRequest) ([]EntryNode, error) {
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

// Consider 2 scenarios
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

// Consider 2 scenarios
// A -> MQ
// A -> External -> B has part B missing, which needs to be completed as A -> B
func getChildNodes(childNodes []ChildNode) *model.TopologyNodes {
	result := model.NewTopologyNodes()
	if len(childNodes) == 0 {
		return result
	}

	childMap := make(map[string]struct{})
	for _, childNode := range childNodes {
		if childNode.ClientGroup == model.GROUP_MQ {
			// MQ data
			result.AddTopologyNode(
				fmt.Sprintf("%s.%s", childNode.ClientPeer, childNode.ClientKey),
				childNode.ClientPeer,
				childNode.ClientKey,
				childNode.IsTraced,
				childNode.ClientGroup,
				childNode.ClientType,
			)
		} else if childNode.Service != "" {
			// Monitored service data
			result.AddServerNode(
				fmt.Sprintf("%s.%s", childNode.Service, childNode.Url),
				childNode.Service,
				childNode.Url,
				childNode.IsTraced,
			)
			// Edge cache associated identifier
			// The downstream node may be lost. You need to complete the data as much as possible.
			if childNode.ClientKey != "" {
				childMap[fmt.Sprintf("%s.%s", childNode.ClientPeer, childNode.ClientKey)] = struct{}{}
			}
		}
	}

	// Service not monitored
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
			// MQ data A -> MQ -> B, two nodes MQ and B need to be generated
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
			// Monitored service data
			// A -> B
			result.AddServerNode(
				fmt.Sprintf("%s.%s", relation.Service, relation.Url),
				relation.Service,
				relation.Url,
				relation.IsTraced,
			)
			// There is a scene where A -> B has multiple different sides
			// A -> Edge Cache Relationship
			if relation.ClientKey != "" {
				childMap[relation.getParentClientKey()] = struct{}{}
			}
		}
	}

	// Service not monitored
	// The downstream node may be lost. You need to complete the data as much as possible.
	for _, relation := range relations {
		if relation.ClientKey != "" && relation.Service == "" && relation.ClientGroup != model.GROUP_MQ {
			// read A -> edge cache relationship and remove possible dirty data
			key := relation.getParentClientKey()
			if _, exist := childMap[key]; !exist {
				// Service not monitored
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
	childMap := make(map[string]struct{}) // remove dirty data
	for _, relation := range relations {
		if relation.ClientGroup == model.GROUP_MQ {
			// MQ data needs to be supplemented with a call link
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
			// Monitored service data
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
			// There is a scene where A -> B has multiple different sides
			// A -> Edge Cache Relationship
			childMap[relation.getParentClientKey()] = struct{}{}
		}
	}

	// Service not monitored
	// The downstream node may be lost. You need to complete the data as much as possible.
	for _, relation := range relations {
		if relation.ParentService != "" && relation.Service == "" && relation.ClientGroup != model.GROUP_MQ {
			// read A -> edge cache relationship and remove possible dirty data
			key := relation.getParentClientKey()
			if _, exist := childMap[key]; !exist {
				// Service not monitored
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
	Service  string `ch:"entry_service" json:"service"`
	Endpoint string `ch:"entry_url" json:"endpoint"`
}
