// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"fmt"
	"strconv"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prometheus_model "github.com/prometheus/common/model"
)

func (repo *promRepo) GetDataplaneServiceList(ctx core.Context, startTime int64, endTime int64, filter string) ([]*model.Service, error) {
	query := `group by (cluster_id, source, service_id, service_name) (apo_svc_instance_uptime{` + filter + `}[` + VecFromS2E(startTime, endTime) + `])`
	value, _, err := repo.GetApi().QueryRange(ctx.GetContext(), query, v1.Range{
		Start: time.UnixMicro(startTime),
		End:   time.UnixMicro(endTime),
		Step:  time.Duration(endTime-startTime) * time.Microsecond,
	})
	if err != nil {
		return nil, err
	}
	vector, ok := value.(prometheus_model.Matrix)
	if !ok {
		return nil, fmt.Errorf("unexpected type %T, expected model.Vector", value)
	}

	result := make([]*model.Service, 0)
	for _, sample := range vector {
		clusterId := string(sample.Metric["cluster_id"])
		source := string(sample.Metric["source"])
		serviceId := string(sample.Metric["service_id"])
		serviceName := string(sample.Metric["service_name"])

		result = append(result, &model.Service{
			ClusterId: clusterId,
			Source:    source,
			Id:        serviceId,
			Name:      serviceName,
		})
	}
	return result, nil
}

func (repo *promRepo) GetDataplaneServiceInstances(ctx core.Context, startTime int64, endTime int64, serviceName string, filter string) ([]*model.ServiceInstance, error) {
	query := `group by (node_ip, node_name, pid, container_id, pod_name) (apo_svc_instance_uptime{` + filter + `}[` + VecFromS2E(startTime, endTime) + `])`
	value, _, err := repo.GetApi().QueryRange(ctx.GetContext(), query, v1.Range{
		Start: time.UnixMicro(startTime),
		End:   time.UnixMicro(endTime),
		Step:  time.Duration(endTime-startTime) * time.Microsecond,
	})
	if err != nil {
		return nil, err
	}
	vector, ok := value.(prometheus_model.Matrix)
	if !ok {
		return nil, fmt.Errorf("unexpected type %T, expected model.Vector", value)
	}

	result := make([]*model.ServiceInstance, 0)
	for _, sample := range vector {
		nodeIp := string(sample.Metric["node_ip"])
		nodeName := string(sample.Metric["node_name"])
		pid := string(sample.Metric["pid"])
		containerId := string(sample.Metric["container_id"])
		podName := string(sample.Metric["pod_name"])

		var pidVal int64 = 0
		if pid != "" {
			pidVal, err = strconv.ParseInt(pid, 10, 64)
			if err != nil {
				return nil, err
			}
		}
		result = append(result, &model.ServiceInstance{
			ServiceName: serviceName,
			NodeIP:      nodeIp,
			NodeName:    nodeName,
			Pid:         pidVal,
			ContainerId: containerId,
			PodName:     podName,
		})
	}
	return result, nil
}
