package service

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetServiceEndpointRelation(req *request.GetServiceEndpointRelationRequest) (*response.GetServiceEndpointRelationResponse, error) {
	// 查询所有上游节点
	parents, err := s.chRepo.ListParentNodes(&req.GetServiceEndpointTopologyRequest)
	if err != nil {
		return nil, err
	}

	var relations []*model.ToplogyRelation
	if req.RemoveClientCall {
		relations, err = s.chRepo.ListDescendantRelationsWithoutEdge(&req.GetServiceEndpointTopologyRequest)
	} else {
		// 查询所有下游节点的调用关系列表
		relations, err = s.chRepo.ListDescendantRelations(&req.GetServiceEndpointTopologyRequest)
	}

	if err != nil {
		return nil, err
	}

	res := &response.GetServiceEndpointRelationResponse{
		Parents:       parents.GetNodes(),
		Current:       model.NewServerNode(req.Service, req.Endpoint, true),
		ChildRelation: relations,
	}

	if req.WithTopologyLevel {
		if req.EntryService == "" || req.EntryEndpoint == "" {
			req.EntryService = req.Service
			req.EntryEndpoint = req.Endpoint
		}

		// 构造邻接矩阵
		endpointMatrix := newEndpointDepthMatrix(req.EntryService, req.EntryEndpoint, relations)
		if req.WithTopologyLevel {
			// 下游接口评估非循环下的最大深度
			res.TopologyLevels = endpointMatrix.EndpointsLevel()
		}
	}
	return res, nil
}

type EndpointDepthMatrix struct {
	EndpointSet  map[model.EndpointKey]int
	EndpointKeys []model.EndpointKey

	NotTracedEndpointsIdx []int

	Matrix [][]int
}

type EndpointInfo struct {
	model.EndpointKey
	IsTraced bool

	ExternalClientCalls []model.TopologyNode
}

func newEndpointDepthMatrix(entryService, entryEndpoint string, relations []*model.ToplogyRelation) *EndpointDepthMatrix {
	matrix := &EndpointDepthMatrix{
		EndpointSet:  map[model.EndpointKey]int{},
		EndpointKeys: []model.EndpointKey{},
	}
	root := model.EndpointKey{
		ServiceName: entryService,
		Endpoint:    entryEndpoint,
	}
	matrix.EndpointSet[root] = 0
	matrix.EndpointKeys = append(matrix.EndpointKeys, root)

	for _, relation := range relations {
		childEndpoint := model.EndpointKey{
			ServiceName: relation.Service,
			Endpoint:    relation.Endpoint,
		}
		_, find := matrix.EndpointSet[childEndpoint]
		if !find {
			childIdx := len(matrix.EndpointKeys)
			matrix.EndpointSet[childEndpoint] = childIdx
			matrix.EndpointKeys = append(matrix.EndpointKeys, childEndpoint)
		}
		parentEndpoint := model.EndpointKey{
			ServiceName: relation.ParentService,
			Endpoint:    relation.ParentEndpoint,
		}
		_, find = matrix.EndpointSet[parentEndpoint]
		if !find {
			parentIdx := len(matrix.EndpointKeys)
			matrix.EndpointSet[parentEndpoint] = parentIdx
			matrix.EndpointKeys = append(matrix.EndpointKeys, parentEndpoint)
		}
	}

	for i := 0; i < len(matrix.EndpointKeys); i++ {
		matrix.Matrix = append(matrix.Matrix, make([]int, len(matrix.EndpointKeys)))
	}

	for _, relation := range relations {
		matrix.AddRelation(relation)
	}

	return matrix
}

func (m *EndpointDepthMatrix) AddRelation(relation *model.ToplogyRelation) {
	parentEndpoint := model.EndpointKey{
		ServiceName: relation.ParentService,
		Endpoint:    relation.ParentEndpoint,
	}
	childEndpoint := model.EndpointKey{
		ServiceName: relation.Service,
		Endpoint:    relation.Endpoint,
	}
	parentIdx := m.EndpointSet[parentEndpoint]
	childIdx := m.EndpointSet[childEndpoint]
	m.Matrix[parentIdx][childIdx] = 1
}

func (m *EndpointDepthMatrix) GetChildCounts(idx int) int {
	if idx >= len(m.Matrix) {
		return -1
	}

	connectStatus := m.Matrix[idx]
	var childCount = 0
	for _, v := range connectStatus {
		childCount += v
	}
	return childCount
}

func (m *EndpointDepthMatrix) MaxDepth(service, endpoint string) (int, bool) {
	if len(m.EndpointKeys) == 0 {
		return -1, false
	}

	if m.EndpointKeys[0].ServiceName == service && m.EndpointKeys[0].Endpoint == endpoint {
		return 0, true
	}

	endpointIdx, find := m.EndpointSet[model.EndpointKey{
		ServiceName: service,
		Endpoint:    endpoint,
	}]
	if !find {
		return -1, false
	}

	// 查询邻接矩阵中root到endpointIdx的最大不循环深度
	visited := make([]bool, len(m.Matrix))
	maxDistance := 0

	return m.dfs(0, endpointIdx, visited, 0, maxDistance)
}

func (m *EndpointDepthMatrix) dfs(node int, target int, visited []bool, currentDistance int, maxDistance int) (int, bool) {
	if node == target {
		return max(currentDistance, maxDistance), true
	}

	visited[node] = true
	var partMaxDistance = -1
	for i, edge := range m.Matrix[node] {
		if edge > 0 && !visited[i] { // 存在边且未访问
			newMaxDistance, found := m.dfs(i, target, visited, currentDistance+1, maxDistance)
			if found && newMaxDistance > partMaxDistance {
				partMaxDistance = newMaxDistance
			}
		}
	}
	visited[node] = false // 回溯

	if partMaxDistance > 0 {
		return partMaxDistance, true
	}
	return maxDistance, false
}

func (m *EndpointDepthMatrix) EndpointsLevel() []response.TopologyNodeLevel {
	var res []response.TopologyNodeLevel
	for idx, endpoint := range m.EndpointKeys {
		if isSkipped(m.NotTracedEndpointsIdx, idx) {
			continue
		}
		depth, find := m.MaxDepth(endpoint.ServiceName, endpoint.Endpoint)
		if !find {
			depth = -1
		}
		res = append(res, response.TopologyNodeLevel{
			Service:  endpoint.ServiceName,
			Endpoint: endpoint.Endpoint,
			Depth:    depth,
		})
	}
	return res
}

func isSkipped(skipped []int, idx int) bool {
	for _, v := range skipped {
		if v == idx {
			return true
		}
	}
	return false
}
