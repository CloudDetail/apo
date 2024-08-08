package clickhouse

import (
	"context"
	"log"
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
		  GROUP BY LogAttributes['k8s.event.reason'], SeverityText
		)`
)

// CountK8sEvents 计算K8s事件数量
// 时间单位为微秒
func (ch *chRepo) CountK8sEvents(startTime int64, endTim int64, pods []string) ([]K8sEventsCount, error) {
	result := make([]K8sEventsCount, 0)
	// 执行查询
	rows, err := ch.conn.Query(context.Background(), countK8sEventsSQL, startTime/1e6, endTim/1e6, pods)
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
