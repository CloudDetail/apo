// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"context"
	"errors"
	"fmt"
	"time"

	prometheus_model "github.com/prometheus/common/model"
)

const (
	TEMPLATE_GET_NODES = `count(count by (node_ip) (originx_agent_ready[%s]))`
)

func (repo *promRepo) GetNodesCount(startTime int64, endTime int64) (int, error) {
	query := fmt.Sprintf(TEMPLATE_GET_NODES, VecFromS2E(startTime, endTime))
	value, _, err := repo.GetApi().Query(context.Background(), query, time.UnixMicro(endTime))

	if err != nil {
		return 0, err
	}

	vector, ok := value.(prometheus_model.Vector)
	if !ok {
		return 0, errors.New("failed to get node count")
	}

	if len(vector) == 0 {
		return 0, nil
	}

	return int(vector[0].Value), nil
}
