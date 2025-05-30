// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

import (
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

const UP = 1.0

func (s *service) GetMonitorStatus(ctx core.Context, startTime time.Time, endTime time.Time) (response.GetMonitorStatusResponse, error) {
	resp := response.GetMonitorStatusResponse{}
	startMicroTS := startTime.UnixMicro()
	endMicroTs := endTime.UnixMicro()

	status, err := s.promRepo.QueryAggMetricsWithFilter(ctx, prometheus.PQLMonitorStatus, startMicroTS, endMicroTs, "")
	if err != nil {
		return resp, nil
	}
	for _, st := range status {
		monitor := response.MonitorStatus{
			MonitorName: st.Metric.MonitorName,
		}
		if st.Values[0].Value == UP {
			monitor.IsAlive = true
		} else {
			// down pending maintenance
			monitor.IsAlive = false
		}
		resp.MonitorList = append(resp.MonitorList, monitor)
	}

	return resp, nil
}
