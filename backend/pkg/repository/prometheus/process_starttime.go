// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	pmodel "github.com/prometheus/common/model"
)

const queryProcessStartPromQL = "max by (node_name,pid,container_id) (last_over_time(%s))"

const metricTimeseries = "originx_process_start_time{%s}[%s]"

func (repo *promRepo) QueryProcessStartTime(startTime time.Time, endTime time.Time, instances []*model.ServiceInstance) (map[model.ServiceInstance]int64, error) {
	vector := VecFromS2E(startTime.UnixMicro(), endTime.UnixMicro())

	var pids []string
	var nodeNames []string
	var containerIDs []string

	for _, instance := range instances {
		if instance == nil {
			continue
		}
		if len(instance.ContainerId) > 0 {
			containerIDs = append(containerIDs, instance.ContainerId)
		} else {
			pids = append(pids, strconv.FormatInt(instance.Pid, 10))
			nodeNames = append(nodeNames, EscapeRegexp(instance.NodeName))
		}
	}

	var timeseries []string
	if len(containerIDs) > 0 {
		filter := fmt.Sprintf("container_id=~\"%s\"", strings.Join(containerIDs, "|"))
		timeseries = append(timeseries, fmt.Sprintf(metricTimeseries, filter, vector))
	}

	if len(pids) > 0 {
		var filters []string
		if len(pids) > 0 {
			filters = append(filters, fmt.Sprintf("pid=~\"%s\"", strings.Join(pids, "|")))
		}
		if len(nodeNames) > 0 {
			filters = append(filters, fmt.Sprintf("node_name=~\"%s\"", strings.Join(nodeNames, "|")))
		}
		timeseries = append(timeseries, fmt.Sprintf(metricTimeseries, strings.Join(filters, ","), vector))
	}

	query := fmt.Sprintf(queryProcessStartPromQL, strings.Join(timeseries, " or "))
	res, _, err := repo.GetApi().Query(context.Background(), query, endTime)
	if err != nil {
		return nil, err
	}

	samples, ok := res.(pmodel.Vector)
	if !ok {
		return nil, fmt.Errorf("unexpected query result type %T, expected model.Vector", res)
	}

	var startTSmap = make(map[model.ServiceInstance]int64)
	for _, sample := range samples {
		if math.IsNaN(float64(sample.Value)) || math.IsInf(float64(sample.Value), 0) {
			continue
		}

		containerId := string(sample.Metric["container_id"])
		if len(containerId) > 0 {
			for _, instance := range instances {
				if instance == nil {
					continue
				}
				if instance.ContainerId == containerId {
					startTSmap[*instance] = int64(sample.Value)
					break
				}
			}
		} else {
			nodeName := string(sample.Metric["node_name"])
			pid := string(sample.Metric["pid"])
			if len(pid) == 0 || len(nodeName) == 0 {
				continue
			}
			pidI64, err := strconv.ParseInt(pid, 10, 64)
			if err != nil {
				continue
			}
			for _, instance := range instances {
				if instance == nil {
					continue
				}
				if nodeName == instance.NodeName && pidI64 == instance.Pid {
					startTSmap[*instance] = int64(sample.Value)
					break
				}
			}
		}
	}

	return startTSmap, nil
}
