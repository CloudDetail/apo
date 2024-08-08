package service

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
)

func (s *service) GetErrorInstanceLogs(req *request.GetErrorInstanceLogsRequest) ([]clickhouse.FaultLogResult, error) {
	// 获取错误实例故障现场日志
	query := &clickhouse.FaultLogQuery{
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Service:   req.Service,
		Instance:  req.Instance,
		EndPoint:  req.Endpoint,
		Type:      1, // Error Only
		PageNum:   1,
		PageSize:  5,
	}
	list, _, err := s.chRepo.GetFaultLogPageList(query)
	if err != nil {
		return nil, err
	}
	return list, nil
}
