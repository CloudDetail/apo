package service

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
)

func (s *service) GetTraceLogs(req *request.GetTraceLogsRequest) ([]clickhouse.FaultLogResult, error) {
	// 获取Trace故障现场日志
	query := &clickhouse.FaultLogQuery{
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Service:     req.Service,
		Instance:    req.Instance,
		NodeName:    req.NodeName,
		ContainerId: req.ContainerId,
		Pid:         req.Pid,
		EndPoint:    req.Endpoint,
		Type:        0, // Slow && Error
		PageNum:     1,
		PageSize:    5,
	}
	list, _, err := s.chRepo.GetFaultLogPageList(query)
	if err != nil {
		return nil, err
	}
	return list, nil
}
