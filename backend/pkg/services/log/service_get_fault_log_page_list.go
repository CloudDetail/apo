package log

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
)

func (s *service) GetFaultLogPageList(req *request.GetFaultLogPageListRequest) (*response.GetFaultLogPageListResponse, error) {
	// 分页查询故障现场日志
	query := &clickhouse.FaultLogQuery{
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Service:     req.Service,
		NodeName:    req.NodeName,
		ContainerId: req.ContainerId,
		Pid:         req.Pid,
		Instance:    req.Instance,
		TraceId:     req.TraceId,
		Type:        2, // Slow && Error && Normal
		PageNum:     req.PageNum,
		PageSize:    req.PageSize,
	}
	list, total, err := s.chRepo.GetFaultLogPageList(query)
	if err != nil {
		return nil, err
	}
	return &response.GetFaultLogPageListResponse{
		Pagination: &model.Pagination{
			Total:       total,
			CurrentPage: req.PageNum,
			PageSize:    req.PageSize,
		},
		List: list,
	}, nil
}
