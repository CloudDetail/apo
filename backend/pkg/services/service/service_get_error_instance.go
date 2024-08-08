package service

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetErrorInstance(req *request.GetErrorInstanceRequest) (*response.GetErrorInstanceResponse, error) {
	serviceInstances, err := s.promRepo.GetInstanceList(req.StartTime, req.EndTime, req.Service, req.Endpoint)
	if err != nil {
		return nil, err
	}

	// 遍历服务实例，查询对应的日志告警数据
	instanceList := make([]*response.ErrorInstance, 0)
	for _, instance := range serviceInstances.GetInstances() {
		logs, err := s.promRepo.QueryLogCountByInstanceId(instance, req.StartTime, req.EndTime, req.Step)
		if err != nil {
			return nil, err
		}

		instanceList = append(instanceList, &response.ErrorInstance{
			Name:        instance.GetInstanceId(),
			ContainerId: instance.ContainerId,
			NodeName:    instance.NodeName,
			Pid:         instance.Pid,
			Propations:  make([]*response.ErrorPropation, 0),
			Logs:        logs,
		})
	}

	// 获取错误传播链路
	propagations, err := s.chRepo.ListErrorPropagation(req)
	if err != nil {
		return nil, err
	}

	// 根据InstanceId进行分组
	errorPropationsMap := make(map[string][]*response.ErrorPropation)
	newInstanceList := make(map[string]bool)
	status := model.STATUS_NORMAL
	if len(propagations) > 0 {
		status = model.STATUS_CRITICAL
		for _, propagation := range propagations {
			parents := make([]*response.InstanceNode, 0)
			for i, size := 0, len(propagation.ParentInstances); i < size; i++ {
				parents = append(parents, &response.InstanceNode{
					Service:  propagation.ParentServices[i],
					Instance: propagation.ParentInstances[i],
					IsTraced: propagation.ParentTraced[i],
				})
			}
			children := make([]*response.InstanceNode, 0)
			for i, size := 0, len(propagation.ChildInstances); i < size; i++ {
				children = append(children, &response.InstanceNode{
					Service:  propagation.ChildServices[i],
					Instance: propagation.ChildInstances[i],
					IsTraced: propagation.ChildTraced[i],
				})
			}

			// 此处需考虑替换为Pod
			instanceId := propagation.InstanceId
			if matchInstance, exist := serviceInstances.InstanceMap[propagation.InstanceId]; exist {
				instanceId = matchInstance.GetInstanceId()
			} else {
				newInstanceList[instanceId] = true
			}

			errorPropations, exist := errorPropationsMap[instanceId]
			if !exist {
				errorPropations = make([]*response.ErrorPropation, 0)
			}
			errorInfos := make([]*response.ErrorInfo, 0)
			for i := 0; i < len(propagation.ErrorTypes); i++ {
				errorInfos = append(errorInfos, &response.ErrorInfo{
					Type:    propagation.ErrorTypes[i],
					Message: propagation.ErrorMsgs[i],
				})
			}

			errorPropations = append(errorPropations, &response.ErrorPropation{
				Timestamp:  propagation.Timestamp.UnixMicro(),
				TraceId:    propagation.TraceId,
				ErrorInfos: errorInfos,
				Parents:    parents,
				Current: &response.InstanceNode{
					Service:  propagation.Service,
					Instance: instanceId,
					IsTraced: true,
				},
				Children: children,
			})
			errorPropationsMap[instanceId] = errorPropations
		}
	}

	for _, errorInstance := range instanceList {
		if errorPropations, exist := errorPropationsMap[errorInstance.Name]; exist {
			errorInstance.Propations = errorPropations
		}
	}

	// 存在未匹配的InstanceId
	for instanceId := range newInstanceList {
		instanceList = append(instanceList, &response.ErrorInstance{
			Name:       instanceId,
			Propations: errorPropationsMap[instanceId],
			Logs:       make(map[int64]float64),
		})
	}

	// 只显示有数据的实例列表
	filteredInstanceList := make([]*response.ErrorInstance, 0)
	for _, instance := range instanceList {
		if exist_metrics(instance.Logs) || len(instance.Propations) > 0 {
			filteredInstanceList = append(filteredInstanceList, instance)
		}
	}
	return &response.GetErrorInstanceResponse{
		Status:    status,
		Instances: filteredInstanceList,
	}, nil
}
