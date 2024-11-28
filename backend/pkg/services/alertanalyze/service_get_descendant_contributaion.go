package alertanalyze

import (
	"strings"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/polarisanalyzer"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

// GetDescendantContribution 基于入口获取下游的故障贡献度排序
func (s *service) GetDescendantContribution(req *request.GetDescendantAlertContributaionRequest) (resp *response.GetDescendantAlertContributationResponse, err error) {
	// 查询所有子孙节点
	nodes, err := s.chRepo.ListDescendantNodes(&req.GetDescendantMetricsRequest)
	if err != nil {
		return nil, err
	}

	if len(nodes.Nodes) == 0 {
		return &response.GetDescendantAlertContributationResponse{
			LatencyContributationList: make([]model.EndpointKey, 0),
		}, nil
	}

	unsortedDescendant := make([]polarisanalyzer.Relevance, 0, len(nodes.Nodes))
	for _, node := range nodes.Nodes {
		unsortedDescendant = append(unsortedDescendant, polarisanalyzer.Relevance{
			Service:  node.Service,
			Endpoint: node.Endpoint,
		})
	}

	if len(req.SortBy) == 0 {
		req.SortBy = "latency,errorRate"
	}

	sortBys := strings.Split(req.SortBy, ",")
	var latencyContributationList []model.EndpointKey
	var contributationMap map[string][]model.EndpointKey = make(map[string][]model.EndpointKey)
	for _, sortBy := range sortBys {
		// 按延时相似度排序
		sortResp, err := s.polRepo.SortDescendantByRelevance(
			req.StartTime, req.EndTime, prometheus.VecFromDuration(time.Duration(req.Step)*time.Microsecond),
			req.Service, req.Endpoint,
			unsortedDescendant, sortBy,
		)

		if err != nil {
			return &response.GetDescendantAlertContributationResponse{}, err
		}

		var contributationList []model.EndpointKey
		for i := 0; i < len(sortResp.SortedDescendant) && i < 3; i++ {
			contributationList = append(contributationList, model.EndpointKey{
				ServiceName: sortResp.SortedDescendant[i].Service,
				Endpoint:    sortResp.SortedDescendant[i].Endpoint,
			})
		}
		contributationMap[sortBy] = contributationList

		// TODO 兼容
		if sortBy == "latency" {
			latencyContributationList = contributationList
		}
	}

	return &response.GetDescendantAlertContributationResponse{
		LatencyContributationList: latencyContributationList,
		ContributationMap:         contributationMap,
	}, nil
}
