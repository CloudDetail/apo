// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
)

func (repo *promRepo) GetMinimalAgentExisted(ctx core.Context, endTime time.Time, duration time.Duration) ([]string, error) {
	res, err := repo.QueryData(ctx,
		endTime,
		groupBy("cluster_id",
			lastOverTime(rangeVec("originx_agent_ready", nil, VecFromDuration(duration), "")),
		))
	if err != nil {
		return nil, err
	}

	var clusterIDs []string
	for _, metric := range res {
		clusterIDs = append(clusterIDs, metric.Metric.ClusterID)
	}

	return clusterIDs, nil
}
