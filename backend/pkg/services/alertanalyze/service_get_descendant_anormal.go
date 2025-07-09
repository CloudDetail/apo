package alertanalyze

import (
	"fmt"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/common"
)

// GetDescendantAnormal 查询入口节点的下游告警情况
func (s *service) GetDescendantAnormal(ctx core.Context, req *request.GetDescendantAnormalEventRequest) (*response.GetDescendantAnormalEventResponse, error) {
	startTime := time.UnixMicro(req.StartTime)
	endTime := time.UnixMicro(req.EndTime)

	if req.PageParam == nil {
		req.PageParam = &request.PageParam{
			CurrentPage: 1,
			PageSize:    999,
		}
	}

	// 先查询入口节点的下游节点
	nodes, err := s.chRepo.ListDescendantNodes(
		ctx,
		&request.GetDescendantMetricsRequest{
			StartTime:     req.StartTime,
			EndTime:       req.EndTime,
			Service:       req.Service,
			Endpoint:      req.Endpoint,
			EntryService:  req.EntryService,
			EntryEndpoint: req.EntryEndpoint,
		})
	if err != nil {
		return &response.GetDescendantAnormalEventResponse{}, err
	}

	nodes, err = common.CutTopologyNodeInGroup(ctx, s.dbRepo, req.GroupID, nodes)
	if err != nil {
		return nil, err
	}

	var instances []*model.ServiceInstance
	var databaseInstances []model.MiddlewareInstance
	tmpServiceSet := map[string]struct{}{}
	endpointList := make([]model.EndpointKey, 0)

	nodes.AddServerNode(fmt.Sprintf("%s.%s", req.Service, req.Endpoint),
		req.Service, req.Endpoint, true)

	for _, node := range nodes.Nodes {
		endpointList = append(endpointList, model.EndpointKey{
			ServiceName: node.Service,
			Endpoint:    node.Endpoint,
		})

		if _, find := tmpServiceSet[node.Service]; find {
			continue
		}
		tmpServiceSet[node.Service] = struct{}{}

		// 忽略URL查询Instance
		instanceList, err := s.promRepo.GetActiveRequestInstanceList(ctx, req.StartTime, req.EndTime, node.Service, "")
		if err != nil {
			continue
		}
		instances = append(instances, instanceList.GetInstances()...)

		filter := prometheus.NewFilter().
			Equal(prometheus.ServiceNameKey, node.Service).
			Equal(prometheus.ContentKeyKey, node.Endpoint)
		// 查询服务下游的数据库服务
		dbInstance, err := s.promRepo.GetDescendantDatabase(ctx, req.StartTime, req.EndTime, filter)
		if err != nil {
			continue
		}
		// instanceMap.AddDatabaseInstance(endpoint, dbInstance)
		databaseInstances = append(databaseInstances, dbInstance...)
	}

	// 通过下游节点查询告警事件
	events, count, err := s.chRepo.GetAlertEventsByInstanceAndEndpoints(ctx, startTime, endTime,
		request.AlertFilter{Status: "firing", WithMutation: true, EndpointKeys: endpointList},
		&model.RelatedInstances{
			SIs: instances,
			MIs: databaseInstances,
		}, req.PageParam,
		clickhouse.OrderAlertByReceivedTime,
	)

	// TODO 告警事件关联到影响服务
	return &response.GetDescendantAnormalEventResponse{
		AlertEvents: events,
		Pagination: &model.Pagination{
			Total:       int64(count),
			CurrentPage: req.PageParam.CurrentPage,
			PageSize:    req.PageParam.PageSize,
		},
	}, err
}
