// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	prometheus_model "github.com/prometheus/common/model"
)

func (repo *promRepo) GetNamespaceList(ctx core.Context, startTime int64, endTime int64, filter PQLFilter) ([]string, error) {
	pqlTemplate := PQLMetricSeries(SPAN_TRACE_COUNT)
	query := pqlTemplate(VecFromS2E(startTime, endTime), "namespace", filter, "")
	value, _, err := repo.GetApi().Query(ctx.GetContext(), query, time.UnixMicro(endTime))

	if err != nil {
		return nil, err
	}
	result := make([]string, 0)
	vector, ok := value.(prometheus_model.Vector)
	if !ok {
		return result, nil
	}
	for _, sample := range vector {
		result = append(result, string(sample.Metric["namespace"]))
	}
	return result, nil
}
