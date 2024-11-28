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

	// 查询所有下游节点的调用关系列表
	relations, err := s.chRepo.ListDescendantRelations(&req.GetServiceEndpointTopologyRequest)
	if err != nil {
		return nil, err
	}

	res := &response.GetServiceEndpointRelationResponse{
		Parents:       parents.GetNodes(),
		Current:       model.NewServerNode(req.Service, req.Endpoint, true),
		ChildRelation: relations,
	}

	if !req.WithNotTracedLeafNode || req.WithTopologyLevel {
		if req.EntryService == "" || req.EntryEndpoint == "" {
			req.EntryService = req.Service
			req.EntryEndpoint = req.Endpoint
		}

		// 构造邻接矩阵
		endpointMatrix := newEndpointDepthMatrix(req.EntryService, req.EntryEndpoint, relations)

		// 如果要剪除未监控的叶子节点,则将相关Relation进行过滤
		if !req.WithNotTracedLeafNode {
			for idx, info := range endpointMatrix.EndpointInfos {
				if info.IsTraced {
					continue
				}
				childCount := endpointMatrix.GetChildCounts(idx)
				if childCount > 0 {
					continue
				}
				// 从Relation中移除相关关联
				for i := 0; i < len(relations); {
					relation := relations[i]
					if relation.Service == info.ServiceName && relation.Endpoint == info.Endpoint {
						parentIdx, find := endpointMatrix.EndpointSet[model.EndpointKey{
							ServiceName: relation.ParentService,
							Endpoint:    relation.ParentEndpoint,
						}]
						if find {
							// 在父节点中添加下游附属信息
							endpointMatrix.EndpointInfos[parentIdx].ExternalClientCalls = append(
								endpointMatrix.EndpointInfos[parentIdx].ExternalClientCalls,
								clickhouse.TopologyNode{
									Service:  relation.Service,
									Endpoint: relation.Endpoint,
								},
							)
						}
						// 移除当前Relation
						relations = append(relations[:i], relations[i+1:]...)
					} else {
						i++
					}
				}

				// 标记后续深度分析时跳过此节点
				endpointMatrix.NotTracedEndpointsIdx = append(endpointMatrix.NotTracedEndpointsIdx, idx)
			}

			// 剪枝完成,把剪下的Relation添加到返回结构中
			for _, info := range endpointMatrix.EndpointInfos {
				if len(info.ExternalClientCalls) <= 0 {
					continue
				}
				res.ExternalClientCalls = append(res.ExternalClientCalls, response.TopologyExternalClientCall{
					TopologyNode: clickhouse.TopologyNode{
						Service:  info.ServiceName,
						Endpoint: info.Endpoint,
						IsTraced: info.IsTraced,
					},
					Childs: info.ExternalClientCalls,
				})
			}

			res.ChildRelation = relations
		}

		if req.WithTopologyLevel {
			// 下游接口评估非循环下的最大深度
			res.TopologyLevels = endpointMatrix.EndpointsLevel()
		}
	}
	return res, nil
}

type EndpointDepthMatrix struct {
	EndpointSet   map[model.EndpointKey]int
	EndpointKeys  []model.EndpointKey
	EndpointInfos []EndpointInfo // 记录端点的特殊信息,是否Traced

	NotTracedEndpointsIdx []int

	Matrix [][]int
}

type EndpointInfo struct {
	model.EndpointKey
	IsTraced bool

	ExternalClientCalls []clickhouse.TopologyNode
}

func newEndpointDepthMatrix(entryService, entryEndpoint string, relations []*clickhouse.ToplogyRelation) *EndpointDepthMatrix {
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
	matrix.EndpointInfos = append(matrix.EndpointInfos, EndpointInfo{
		EndpointKey:         root,
		IsTraced:            true,
		ExternalClientCalls: []clickhouse.TopologyNode{},
	})
	for _, relation := range relations {
		childEndpoint := model.EndpointKey{
			ServiceName: relation.Service,
			Endpoint:    relation.Endpoint,
		}
		idx, find := matrix.EndpointSet[childEndpoint]
		if !find {
			childIdx := len(matrix.EndpointKeys)
			matrix.EndpointSet[childEndpoint] = childIdx
			matrix.EndpointKeys = append(matrix.EndpointKeys, childEndpoint)
			matrix.EndpointInfos = append(matrix.EndpointInfos, EndpointInfo{
				EndpointKey:         childEndpoint,
				IsTraced:            relation.IsTraced,
				ExternalClientCalls: []clickhouse.TopologyNode{},
			})
		} else if relation.IsTraced && !matrix.EndpointInfos[idx].IsTraced {
			// 更新IsTraced状态
			matrix.EndpointInfos[idx].IsTraced = true
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
			matrix.EndpointInfos = append(matrix.EndpointInfos, EndpointInfo{
				EndpointKey:         parentEndpoint,
				IsTraced:            true, // 父节点一定被监控
				ExternalClientCalls: []clickhouse.TopologyNode{},
			})
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

func (m *EndpointDepthMatrix) AddRelation(relation *clickhouse.ToplogyRelation) {
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
