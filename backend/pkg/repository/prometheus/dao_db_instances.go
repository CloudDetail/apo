// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"context"
	"fmt"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	prometheus_model "github.com/prometheus/common/model"
)

const (
	TEMPLATE_GET_DESCENDANT_DATABASE_BY_SERVICE = `group by (db_url,container_id,node_name,pid)
	(last_over_time(kindling_db_duration_nanoseconds_count{%s}[%s])) * on (container_id,node_name,pid) group_left (peer_ip,peer_port) (group by (container_id,node_name,pid,peer_ip,peer_port) (last_over_time(apo_network_middleware_connect[%s])))`
)

// GetDescendantDatabase query database which the service connected to.
func (repo *promRepo) GetDescendantDatabase(startTime int64, endTime int64, serviceName string, endpoint string) ([]model.MiddlewareInstance, error) {
	vec := VecFromS2E(startTime, endTime)

	pql := fmt.Sprintf(
		TEMPLATE_GET_DESCENDANT_DATABASE_BY_SERVICE,
		fmt.Sprintf("svc_name=\"%s\",content_key=\"%s\"", serviceName, endpoint),
		vec, vec,
	)

	res, _, err := repo.GetApi().Query(context.Background(), pql, time.UnixMicro(endTime))
	if err != nil {
		return nil, err
	}
	dbInstances := make([]model.MiddlewareInstance, 0)
	vector, ok := res.(prometheus_model.Vector)
	if !ok {
		return dbInstances, nil
	}

	for _, sample := range vector {
		dbInstances = append(dbInstances, model.MiddlewareInstance{
			DatabaseURL:  string(sample.Metric["db_url"]),
			DatabaseIP:   string(sample.Metric["peer_ip"]),
			DatabasePort: string(sample.Metric["peer_port"]),
		})
	}
	return dbInstances, nil
}
