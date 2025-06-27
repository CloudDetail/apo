// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

func (s *service) GetErrorInstance(ctx core.Context, req *request.GetErrorInstanceRequest) (*response.GetErrorInstanceResponse, error) {
	filter := prometheus.NewFilter()
	filter.EqualIfNotEmpty(prometheus.ContentKeyKey, req.Endpoint)
	filter.EqualIfNotEmpty(prometheus.ServiceNameKey, req.Service)

	if len(req.ClusterIDs) > 0 {
		filter.RegexMatch("cluster_id", prometheus.RegexMultipleValue(req.ClusterIDs...))
	}

	serviceInstances, err := s.promRepo.GetInstanceListByPQLFilter(ctx, req.StartTime, req.EndTime, filter)
	if err != nil {
		return nil, err
	}

	// Traverses the service instance and queries the corresponding log alarm data.
	instanceList := make([]*response.ErrorInstance, 0)
	for _, instance := range serviceInstances.GetInstances() {
		logs, err := s.promRepo.QueryLogCountByInstanceId(ctx, instance, req.StartTime, req.EndTime, req.Step)
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

	// Get error propagation link
	propagations, err := s.chRepo.ListErrorPropagation(ctx, req)
	if err != nil {
		return nil, err
	}

	// Group according to InstanceId
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

			// Consider replacing with Pod here
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

	// Unmatched InstanceId exists
	for instanceId := range newInstanceList {
		instanceList = append(instanceList, &response.ErrorInstance{
			Name:       instanceId,
			Propations: errorPropationsMap[instanceId],
			Logs:       make(map[int64]float64),
		})
	}

	// Display only the list of instances with data
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
