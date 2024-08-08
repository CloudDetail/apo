package clickhouse

import (
	"context"
	"time"
)

// InfrastructureAlert 查询基础设施告警，按节点名称区分，如果有数据返回true，没有数据返回false
func (ch *chRepo) InfrastructureAlert(startTime time.Time, endTime time.Time, nodeNames []string) (bool, error) {
	// 构建查询语句
	query := `
		SELECT 1
		FROM alert_event
		WHERE received_time BETWEEN $1 AND $2 AND tags['nodename'] IN $3 
		  AND group='infra' AND status='firing'
		LIMIT 1
	`

	// 执行查询
	rows, err := ch.conn.Query(context.Background(), query, startTime.Unix(), endTime.Unix(), nodeNames)
	if err != nil {
		return false, err
	}
	// 检查是否有查询结果
	if rows.Next() {
		return true, nil
	}

	return false, nil
}

// NetworkAlert   查网络告警
func (ch *chRepo) NetworkAlert(startTime time.Time, endTime time.Time, pods []string, nodeNames []string, pids []string) (bool, error) {
	// 构建查询语句
	query := `    SELECT 1
    FROM alert_event
    WHERE received_time BETWEEN toDateTime($1) AND toDateTime($2)
      AND (
          tags['src_pod'] IN $3 OR (
          tags['src_node'] IN $4 AND
          arrayExists(pid -> has(splitByChar(',', tags['pid']), toString(pid)), $5)
      ))
      AND group = 'network' AND status = 'firing'
    LIMIT 1`

	// 执行查询
	rows, err := ch.conn.Query(context.Background(), query, startTime.Unix(), endTime.Unix(), pods, nodeNames, pids)
	if err != nil {
		return false, err
	}
	// 检查是否有查询结果
	if rows.Next() {
		return true, nil
	}

	return false, nil
}

// K8sAlert   查询K8S告警
func (ch *chRepo) K8sAlert(startTime time.Time, endTime time.Time, pods []string) (bool, error) {
	// 构建查询语句
	query := `
		SELECT 1
		FROM k8s_events
		WHERE Timestamp BETWEEN toDateTime($1) AND toDateTime($2) AND ResourceAttributes['k8s.object.name'] IN $3 AND SeverityNumber>9
		LIMIT 1
	`

	// 执行查询
	rows, err := ch.conn.Query(context.Background(), query, startTime.Unix(), endTime.Unix(), pods)
	if err != nil {
		return false, err
	}
	// 检查是否有查询结果
	if rows.Next() {
		return true, nil
	}

	return false, nil
}

var startTimeLayout = "2006-01-02 15:04:05 -0700 MST"

// RebootTime 查询基础设施告警，按节点名称区分，返回最新的重启时间和错误
func (ch *chRepo) RebootTime(endTime int64, podsOrNodeNames []string) (*time.Time, error) {
	// 构建查询语句
	query := `
        SELECT LogAttributes['k8s.event.start_time'] AS start_time
        FROM k8s_events
        WHERE Timestamp <= $1
            AND LogAttributes['k8s.event.reason'] = 'Started'
            AND ResourceAttributes['k8s.object.name'] IN $2
        ORDER BY start_time DESC
        LIMIT 1
    `

	// 执行查询
	rows, err := ch.conn.Query(context.Background(), query, endTime/1e6, podsOrNodeNames)
	if err != nil {
		return nil, err
	}

	// 检查是否有查询结果
	if rows.Next() {
		var rebootTimeStr string
		if err := rows.Scan(&rebootTimeStr); err != nil {
			return nil, err
		}
		// 解析时间字符串为 time.Time 类型
		rebootTime, err := time.Parse(startTimeLayout, rebootTimeStr)
		if err != nil {
			return nil, err
		}
		return &rebootTime, nil
	}

	return nil, nil
}
