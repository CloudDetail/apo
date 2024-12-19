package service

import (
	"strconv"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

func (s *service) GetServiceInstanceList(req *request.GetServiceInstanceListRequest) ([]string, error) {
	// 获取活跃服务实例列表
	instances, err := s.promRepo.GetActiveInstanceList(req.StartTime, req.EndTime, req.ServiceName)
	if err != nil {
		return nil, err
	}

	return instances.GetInstanceIds(), nil
}

func (s *service) GetServiceInstanceInfoList(req *request.GetServiceInstanceListRequest) ([]prometheus.InstanceKey, error) {
	var ins []prometheus.InstanceKey
	// 获取实例
	instanceList, err := s.promRepo.GetInstanceList(req.StartTime, req.EndTime, req.ServiceName, "")
	if err != nil {
		return ins, err
	}

	// 填充实例
	for _, instance := range instanceList.GetInstanceIdMap() {
		key := prometheus.InstanceKey{
			PID:         strconv.FormatInt(instance.Pid, 10),
			ContainerId: instance.ContainerId,
			Pod:         instance.PodName,
			Namespace:   instance.Namespace,
			NodeName:    instance.NodeName,
			NodeIP:      instance.NodeIP,
		}
		ins = append(ins, key)
	}
	return ins, nil
}
