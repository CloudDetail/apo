// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"context"
	"fmt"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prometheus_model "github.com/prometheus/common/model"
)

const (
	TEMPLATE_LOG_SUM = `
	(
		sum by (%s) (increase(originx_logparser_level_count_total{%s, level=~"error|critical"}[%s]))
	  + (sum by (%s) (increase(originx_logparser_exception_count_total{%s}[%s])) or 0)
	) or sum by (%s) (increase(originx_logparser_exception_count_total{%s}[%s]))
	`
)

// Query the log alarm distribution curve
func (repo *promRepo) QueryLogCountByInstanceId(instance *model.ServiceInstance, startTime int64, endTime int64, step int64) (map[int64]float64, error) {
	tRange := v1.Range{
		Start: time.UnixMicro(startTime),
		End:   time.UnixMicro(endTime),
		Step:  time.Duration(step * 1000),
	}
	var (
		key            string
		queryCondition string
	)
	if instance.PodName != "" {
		key = "pod_name"
		queryCondition = fmt.Sprintf("pod_name='%s'", instance.PodName)
	} else if instance.ContainerId != "" {
		key = "host_name"
		queryCondition = fmt.Sprintf("host_name='%s', container_id='%s'", instance.NodeName, instance.ContainerId)
	} else {
		// VM scenario
		key = "host_name"
		queryCondition = fmt.Sprintf("host_name='%s', pid='%d'", instance.NodeName, instance.Pid)
	}
	queryStep := getDurationFromStep(tRange.Step)
	query := fmt.Sprintf(TEMPLATE_LOG_SUM,
		key, queryCondition, queryStep,
		key, queryCondition, queryStep,
		key, queryCondition, queryStep,
	)
	res, _, err := repo.GetApi().QueryRange(context.Background(), query, tRange)
	if err != nil {
		return nil, err
	}
	result := make(map[int64]float64, 0)
	values, ok := res.(prometheus_model.Matrix)
	if !ok {
		return result, nil
	}
	if len(values) == 1 {
		val := values[0]
		for _, pair := range val.Values {
			result[int64(pair.Timestamp)*1000] = float64(pair.Value)
		}
	}
	return result, nil
}
