// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"fmt"
	"log"

	core "github.com/CloudDetail/apo/backend/pkg/core"
)

const (
	countK8sEventsSQL = `
		(
		  SELECT
			'current' AS time_range,
			LogAttributes['k8s.event.reason'],
			SeverityText,
			COUNT(*)
		  FROM k8s_events
		  WHERE Timestamp BETWEEN toDateTime($1) AND toDateTime($2)
		  AND ResourceAttributes['k8s.object.name'] IN ($3)
		  %s
		  GROUP BY LogAttributes['k8s.event.reason'], SeverityText
		)
		UNION ALL
		(
		  SELECT
			'lastWeeks' AS time_range,
			LogAttributes['k8s.event.reason'],
			SeverityText,
			COUNT(*)
		  FROM k8s_events
		  WHERE Timestamp BETWEEN toDateTime($2) - toIntervalDay(7) AND toDateTime($2)
		  AND ResourceAttributes['k8s.object.name'] IN ($3)
		  %s
		  GROUP BY LogAttributes['k8s.event.reason'], SeverityText
		)
		UNION ALL
		(
		  SELECT
			'lastMonth' AS time_range,
			LogAttributes['k8s.event.reason'],
			SeverityText,
			COUNT(*)
		  FROM k8s_events
		  WHERE Timestamp BETWEEN toDateTime($2) - toIntervalDay(30) AND toDateTime($2)
		  AND ResourceAttributes['k8s.object.name'] IN ($3)
		  %s
		  GROUP BY LogAttributes['k8s.event.reason'], SeverityText
		)`
)

// CountK8sEvents count the number of K8s events
// Time in microseconds
func (ch *chRepo) CountK8sEvents(ctx core.Context, startTime int64, endTime int64, pods []string, clusterIDs []string) ([]K8sEventsCount, error) {
	result := make([]K8sEventsCount, 0)

	clusterIdSQLPart := ""
	var argsForClusterID interface{}
	if len(clusterIDs) > 0 {
		clusterIdSQLPart = "AND ResourceAttributes['cluster.id'] IN ($4)"
		argsForClusterID = clusterIDs
	}
	finalQuery := fmt.Sprintf(countK8sEventsSQL, clusterIdSQLPart, clusterIdSQLPart, clusterIdSQLPart)
	queryArgs := []interface{}{
		startTime / 1e6,
		endTime / 1e6,
		pods,
	}

	if len(clusterIDs) > 0 {
		queryArgs = append(queryArgs, argsForClusterID)
	}
	// Execute query
	rows, err := ch.GetContextDB(ctx).Query(ctx.GetContext(), finalQuery, queryArgs...)
	if err != nil {
		return result, err
	}

	for rows.Next() {
		row := K8sEventsCount{}
		err := rows.Scan(&row.TimeRange, &row.Reason, &row.Severity, &row.Count)
		if err != nil {
			log.Println("error to read the k8s count row:", err)
			continue
		}
		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error during iteration:", err)
	}
	return result, nil
}

type K8sEventsCount struct {
	TimeRange string
	Reason    string
	Severity  string
	Count     uint64
}
