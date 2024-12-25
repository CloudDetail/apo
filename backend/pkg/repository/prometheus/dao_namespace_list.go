// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"context"
	"fmt"
	prometheus_model "github.com/prometheus/common/model"
	"time"
)

func (repo *promRepo) GetNamespaceList(startTime int64, endTime int64) ([]string, error) {
	query := fmt.Sprintf(TEMPLATE_GET_NAMESPACES, VecFromS2E(startTime, endTime))
	value, _, err := repo.GetApi().Query(context.Background(), query, time.UnixMicro(endTime))

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
