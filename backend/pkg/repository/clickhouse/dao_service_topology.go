// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"fmt"
	"log"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
)

const (
	TEMPLATE_QUERY_COUNT_TOPOLOGY    = "SELECT count(1) FROM service_topology %s"
	TEMPLATE_QUERY_REALTIME_TOPOLOGY = "SELECT parent_service, parent_type, child_service, child_type FROM service_topology %s GROUP BY parent_service, parent_type, child_service, child_type"
)

func (ch *chRepo) QueryServiceTopologyCount(ctx core.Context, timestamp int64, clusterId string, source string) (uint64, error) {
	queryBuilder := NewQueryBuilder().
		Equals("toUnixTimestamp(timestamp)", timestamp/1000000).
		EqualsNotEmpty("cluster_id", clusterId).
		EqualsNotEmpty("source", source)
	query := fmt.Sprintf(TEMPLATE_QUERY_COUNT_TOPOLOGY, queryBuilder.String())

	var count uint64
	// Number of query records
	err := ch.GetContextDB(ctx).QueryRow(ctx.GetContext(), query, queryBuilder.values...).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

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

func (ch *chRepo) WriteServiceTopology(ctx core.Context, timestamp int64, clusterId string, source string, topologies []*model.ServiceToplogy) error {
	batch, err := ch.GetContextDB(ctx).PrepareBatch(ctx.GetContext(), `
		INSERT INTO service_topology (timestamp, cluster_id, source, parent_service, parent_type, child_service, child_type)
		VALUES
	`)

	if err != nil {
		return err
	}
	for _, topology := range topologies {
		if err := batch.Append(
			time.Unix(timestamp/1000000, 0).UTC(),
			clusterId,
			source,
			topology.ParentService,
			topology.ParentType,
			topology.ChildService,
			topology.ChildType); err != nil {

			log.Println("Failed to send data:", err)
			continue
		}
	}
	if err := batch.Send(); err != nil {
		return err
	}
	return nil
}
