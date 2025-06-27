// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
)

const (
	MIDDLEWARE_CONNECT_COUNT = "apo_network_middleware_connect"
)

// GetDescendantDatabase query database which the service connected to.
func (repo *promRepo) GetDescendantDatabase(ctx core.Context, startTime int64, endTime int64, filter PQLFilter) ([]model.MiddlewareInstance, error) {
	vec := VecFromS2E(startTime, endTime)

	// append label (peer_ip,peer_port) into SPAN_DB_COUNT result
	pql := labelLeftOn(
		groupBy(string(DBInstanceGranularity), lastOverTime(rangeVec(SPAN_DB_COUNT, filter, vec, ""))),
		"container_id,node_name,pid",
		"peer_ip,peer_port",
		lastOverTime(rangeVec(MIDDLEWARE_CONNECT_COUNT, nil, vec, "")),
	)

	res, err := repo.QueryData(ctx, time.UnixMicro(endTime), pql)
	if err != nil {
		return nil, err
	}
	dbInstances := make([]model.MiddlewareInstance, 0)
	for _, sample := range res {
		dbInstances = append(dbInstances, model.MiddlewareInstance{
			DatabaseURL:  string(sample.Metric.DBUrl),
			DatabaseIP:   string(sample.Metric.PeerIP),
			DatabasePort: string(sample.Metric.PeerPort),
		})
	}
	return dbInstances, nil
}
