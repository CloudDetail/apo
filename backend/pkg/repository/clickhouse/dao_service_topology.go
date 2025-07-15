// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"fmt"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
)

const (
	TEMPLATE_QUERY_REALTIME_TOPOLOGY = "SELECT parent_service, parent_type, child_service, child_type FROM service_topology %s GROUP BY parent_service, parent_type, child_service, child_type"
)

func (ch *chRepo) QueryRealtimeServiceTopology(ctx core.Context, startTime int64, endTime int64, clusterId string) ([]model.ServiceToplogy, error) {
	queryBuilder := NewQueryBuilder().
		Between("toUnixTimestamp(timestamp)", startTime/1000000-3600, endTime/1000000+3600).
		EqualsNotEmpty("cluster_id", clusterId)
	query := fmt.Sprintf(TEMPLATE_QUERY_REALTIME_TOPOLOGY, queryBuilder.String())

	result := []model.ServiceToplogy{}
	// Number of query records
	err := ch.GetContextDB(ctx).Select(ctx.GetContext(), &result, query, queryBuilder.values...)
	if err != nil {
		return nil, err
	}
	return result, nil
}
