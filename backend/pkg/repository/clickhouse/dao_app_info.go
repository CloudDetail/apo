// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"fmt"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

const (
	TEMPLATE_QUERY_APP_INFO = "SELECT host_pid, container_id, labels FROM originx_app_info %s"
)

func (ch *chRepo) GetToResolveApps(ctx core.Context) ([]model.AppInfo, error) {
	queryBuilder := NewQueryBuilder().
		LessThan("heart_flag", 2)
	sql := fmt.Sprintf(TEMPLATE_QUERY_APP_INFO, queryBuilder.String())
	// Query list data
	apps := []model.AppInfo{}
	err := ch.GetContextDB(ctx).Select(ctx.GetContext(), &apps, sql, queryBuilder.values...)
	if err != nil {
		return nil, err
	}

	return apps, nil
}

func (ch *chRepo) GetResolvedApps(ctx core.Context, ruleIds []string, startTime int64, endTime int64) ([]model.AppInfo, error) {
	queryBuilder := NewQueryBuilder().
		NotEquals("labels['service_name']", "").
		NotGreaterThan("start_time", endTime/1000000).
		NotLessThan("heart_time", startTime/1000000).
		In("labels['rule_id']", ruleIds)
	sql := fmt.Sprintf(TEMPLATE_QUERY_APP_INFO, queryBuilder.String())
	apps := []model.AppInfo{}
	err := ch.GetContextDB(ctx).Select(ctx.GetContext(), &apps, sql, queryBuilder.values...)
	if err != nil {
		return nil, err
	}

	return apps, nil
}

func (ch *chRepo) GetDataplaneServiceList(ctx core.Context, req *request.QueryServiceNameRequest) ([]*model.Service, error) {
	queryBuilder := NewQueryBuilder().
		NotEquals("labels['service_name']", "").
		EqualsNotEmpty("labels['cluster_id']", req.Cluster).
		EqualsNotEmpty("labels['node_name']", req.Tags.NodeName).
		EqualsNotEmpty("labels['pod_name']", req.Tags.PodName).
		EqualsNotEmpty("container_id", req.Tags.ContainerId).
		NotGreaterThan("start_time", req.EndTime/1000000).
		NotLessThan("heart_time", req.StartTime/1000000)
	sql := fmt.Sprintf(TEMPLATE_QUERY_APP_INFO, queryBuilder.String())
	apps := make([]model.AppInfo, 0)
	if err := ch.GetContextDB(ctx).Select(ctx.GetContext(), &apps, sql, queryBuilder.values...); err != nil {
		return nil, err
	}

	result := make([]*model.Service, 0)
	for _, app := range apps {
		result = append(result, &model.Service{
			ClusterId: app.Labels["cluster_id"],
			Source:    app.Labels["source"],
			Id:        app.Labels["service_id"],
			Name:      app.Labels["service_name"],
		})
	}
	return result, nil
}

func (ch *chRepo) GetDataplaneServiceInstances(ctx core.Context, startTime int64, endTime int64, cluster string, serviceNames []string) ([]*model.ServiceInstance, error) {
	queryBuilder := NewQueryBuilder().
		EqualsNotEmpty("labels['cluster_id']", cluster).
		NotGreaterThan("start_time", endTime/1000000).
		NotLessThan("heart_time", startTime/1000000)
	if len(serviceNames) == 1 {
		queryBuilder.Equals("labels['service_name']", serviceNames[0])
	} else {
		queryBuilder.In("labels['service_name']", serviceNames)
	}
	sql := fmt.Sprintf(TEMPLATE_QUERY_APP_INFO, queryBuilder.String())
	apps := make([]model.AppInfo, 0)
	if err := ch.GetContextDB(ctx).Select(ctx.GetContext(), &apps, sql, queryBuilder.values...); err != nil {
		return nil, err
	}
	result := make([]*model.ServiceInstance, 0)
	for _, app := range apps {
		result = append(result, &model.ServiceInstance{
			ServiceName: app.Labels["service_name"],
			ContainerId: app.ContainerId,
			PodName:     app.Labels["pod_name"],
			Namespace:   app.Labels["namespace"],
			NodeName:    app.Labels["node_name"],
			Pid:         int64(app.HostPid),
			NodeIP:      app.Labels["node_ip"],
			ClusterID:   app.Labels["cluster_id"],
		})
	}
	return result, nil
}
