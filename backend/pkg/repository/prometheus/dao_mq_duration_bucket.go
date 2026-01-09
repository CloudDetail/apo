// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

// Query the P90 curve based on the service list, URL list, time period and step size.
func (repo *promRepo) QueryMqRangePercentile(ctx core.Context, startTime int64, endTime int64, step int64, nodes *model.TopologyNodes) ([]DescendantMetrics, error) {
	svcs, endpoints, systems := nodes.GetLabels(model.GROUP_MQ)
	if len(svcs) == 0 {
		return nil, nil
	}
	tRange := v1.Range{
		Start: time.UnixMicro(startTime),
		End:   time.UnixMicro(endTime),
		Step:  time.Duration(step * 1000),
	}
	qb := getMqP9xSql(repo.promRange, tRange.Step, svcs, endpoints, systems)
	res, _, err := repo.QueryRangeWithP9xBuilder(ctx, qb, tRange)
	if err != nil {
		return nil, err
	}
	return getDescendantMetrics("address", "name", tRange, res), nil
}

func getMqP9xSql(promRange string, step time.Duration, svcs []string, endpoints []string, systems []string) *UnionP9xBuilder {
	builder := NewUnionP9xBuilder(
		"0.9",
		"kindling_mq_duration_nanoseconds_bucket",
		[]string{promRange, "address", "name"},
		step,
	)
	builder.AddCondition("address", svcs)
	builder.AddCondition("name", endpoints)
	builder.AddCondition("system", systems)
	builder.AddExtraCondition("role!='consumer'")
	return builder
}
