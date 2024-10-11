package log

import (
	"sort"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetLogIndex(req *request.LogIndexRequest) (*response.LogIndexResponse, error) {
	list, sum, err := s.chRepo.GetLogIndex(req)
	if err != nil {
		return nil, err
	}
	res := make([]response.IndexItem, 0)
	var count uint64
	for k, v := range list {
		count += v
		res = append(res, response.IndexItem{
			IndexName: k,
			Count:     v,
			Percent:   float64(v) * 100 / float64(sum),
		})
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Count > res[j].Count
	})
	return &response.LogIndexResponse{Indexs: res}, nil
}