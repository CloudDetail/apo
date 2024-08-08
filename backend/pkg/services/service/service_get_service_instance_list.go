package service

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) GetServiceInstanceList(req *request.GetServiceInstanceListRequest) ([]string, error) {
	// 获取活跃服务实例列表
	instances, err := s.promRepo.GetActiveInstanceList(req.StartTime, req.EndTime, req.ServiceName)
	if err != nil {
		return nil, err
	}

	return instances.GetInstanceIds(), nil
}
