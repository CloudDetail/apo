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
func (repo *promRepo) QueryExternalRangePercentile(ctx core.Context, startTime int64, endTime int64, step int64, nodes *model.TopologyNodes) ([]DescendantMetrics, error) {
	svcs, endpoints, systems := nodes.GetLabels(model.GROUP_EXTERNAL)
	if len(svcs) == 0 {
		return nil, nil
	}
	tRange := v1.Range{
		Start: time.UnixMicro(startTime),
		End:   time.UnixMicro(endTime),
		Step:  time.Duration(step * 1000),
	}

	query := getExternalP9xSql(repo.promRange, tRange.Step, svcs, endpoints, systems)
	res, _, err := repo.GetApi().QueryRange(ctx.GetContext(), query, tRange)
	if err != nil {
		return nil, err
	}
	return getDescendantMetrics("address", "name", tRange, res), nil
}

func getExternalP9xSql(promRange string, step time.Duration, svcs []string, endpoints []string, systems []string) string {
	builder := NewUnionP9xBuilder(
		"0.9",
		"kindling_external_duration_nanoseconds_bucket",
		[]string{promRange, "address", "name"},
		step,
	)
	builder.AddCondition("address", svcs)
	builder.AddCondition("name", endpoints)
	builder.AddCondition("system", systems)
	return builder.ToString()
}
