// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"github.com/CloudDetail/apo/backend/config"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
)

func (s *service) queryLogCountByInstance(ctx core.Context, instance *model.ServiceInstance, startTime int64, endTime int64, step int64) (map[int64]float64, error) {
	if !config.Get().Promethues.LogparserPIDSourceQueryEnabled ||
		instance == nil ||
		instance.PodName != "" ||
		instance.ContainerId != "" ||
		instance.NodeName == "" ||
		instance.Pid <= 0 {
		return s.promRepo.QueryLogCountByInstanceId(ctx, instance, startTime, endTime, step)
	}

	pair := clickhouse.NodePID{Node: instance.NodeName, PID: instance.Pid}
	sources, err := s.chRepo.GetLatestSourceByNodePIDs(ctx, []clickhouse.NodePID{pair})
	if err != nil {
		return s.promRepo.QueryLogCountByInstanceId(ctx, instance, startTime, endTime, step)
	}
	source := sources[pair]
	if source == "" {
		return s.promRepo.QueryLogCountByInstanceId(ctx, instance, startTime, endTime, step)
	}
	return s.promRepo.QueryLogCountByInstanceSource(ctx, instance.NodeName, source, startTime, endTime, step)
}
