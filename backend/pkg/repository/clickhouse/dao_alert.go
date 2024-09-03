package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

const (
	// SQL_GET_SAMPLE_ALERT_EVENT 按alarm_event的name分组,每组取发生事件最晚的记录,并在返回结果中记录同name的告警次数数量
	SQL_GET_SAMPLE_ALERT_EVENT = `WITH grouped_alarm AS (
		SELECT source,group,id,create_time,update_time,end_time,received_time,severity,name,detail,tags,status,
        	ROW_NUMBER() OVER (PARTITION BY name ORDER BY received_time) AS rn,
			COUNT(*) OVER (PARTITION BY name) AS alarm_count
    	FROM alert_event
		%s
	)
	SELECT *
	FROM grouped_alarm
	WHERE rn <= %d %s`

	// SQL_GET_PAGED_ALERT_EVENT 分页取出所有满足条件的告警事件
	SQL_GET_PAGED_ALERT_EVENT = `WITH paginatedEvent AS (
		SELECT
			source,group,id,create_time,update_time,end_time,received_time,severity,name,detail,tags,status,
			COUNT(*) OVER () AS total_count,
			ROW_NUMBER() OVER (%s) AS rn
		FROM alert_event
		%s
	)
	SELECT *
	FROM paginatedEvent
	%s ORDER BY rn`

	SQL_GET_ALERT_EVENT = `SELECT source,group,id,create_time,update_time,end_time,received_time,severity,name,detail,tags,status
	FROM alert_event
	%s,%s`
)

func (ch *chRepo) InfrastructureAlert(startTime time.Time, endTime time.Time, nodeNames []string) (*model.AlertEvent, bool, error) {
	builder := NewQueryBuilder().
		Between("received_time", startTime.Unix(), endTime.Unix()).
		InStrings("tags['nodename']", nodeNames).
		Equals("group", "infra").
		Equals("status", "firing")

	byLimit := NewByLimitBuilder().
		Limit(1).
		OrderBy("received_time", false)

	sql := fmt.Sprintf(SQL_GET_ALERT_EVENT, builder.String(), byLimit.String())

	var events []model.AlertEvent
	err := ch.conn.Select(context.Background(), &events, sql, builder.values...)
	if err != nil || len(events) == 0 {
		return nil, false, err
	}

	return &events[0], true, nil
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

// GetAlarmsEvents 获取实例所有的告警事件
func (ch *chRepo) GetAlertEventsSample(sampleCount int, startTime time.Time, endTime time.Time, filter request.AlertFilter, instances []*model.ServiceInstance) ([]AlertEventSample, error) {
	// 组合生成:
	//  1. group = 'app' AND svc = svc_name
	//  2. group = 'container' AND ((namespace,pod) in (...))
	//  3. group = 'network' AND ((src_namespace,pod) in (...) OR (src_node,pid) in (...))
	//  4. group = 'infra' AND ((instance_name) in (...))
	whereInstance := extractFilter(filter, instances)

	builder := NewQueryBuilder().
		Between("received_time", startTime.Unix(), endTime.Unix()).
		EqualsNotEmpty("source", filter.Source).
		EqualsNotEmpty("group", filter.Group).
		EqualsNotEmpty("name", filter.Name).
		EqualsNotEmpty("id", filter.ID).
		EqualsNotEmpty("severity", filter.Severity).
		EqualsNotEmpty("status", filter.Status).
		And(whereInstance)

	byBuilder := NewByLimitBuilder().
		OrderBy("group", true).
		OrderBy("name", true)

	sql := fmt.Sprintf(SQL_GET_SAMPLE_ALERT_EVENT, builder.String(), sampleCount, byBuilder.String())

	var events []AlertEventSample
	err := ch.conn.Select(context.Background(), &events, sql, builder.values...)
	return events, err
}

func (ch *chRepo) GetAlertEvents(startTime time.Time, endTime time.Time, filter request.AlertFilter, instances []*model.ServiceInstance, pageParam *request.PageParam) ([]PagedAlertEvent, int, error) {
	whereInstance := extractFilter(filter, instances)

	builder := NewQueryBuilder().
		Between("received_time", startTime.Unix(), endTime.Unix()).
		EqualsNotEmpty("source", filter.Source).
		EqualsNotEmpty("group", filter.Group).
		EqualsNotEmpty("name", filter.Name).
		EqualsNotEmpty("id", filter.ID).
		EqualsNotEmpty("severity", filter.Severity).
		EqualsNotEmpty("status", filter.Status).
		And(whereInstance)

	// HACK 基于窗口函数实现数据分页,和不同查询语句不同
	// !!! 该位置不得添加Limit / Group 参数
	orderBuilder := NewByLimitBuilder().
		OrderBy("group", true).
		OrderBy("name", true).
		OrderBy("received_time", false)
	orders := orderBuilder.String()

	sql := fmt.Sprintf(SQL_GET_PAGED_ALERT_EVENT, orders, builder.String(), RnLimit(pageParam))
	var events []PagedAlertEvent
	err := ch.conn.Select(context.Background(), &events, sql, builder.values...)
	var total_count = 0
	if len(events) > 0 {
		total_count = int(events[0].TotalCount)
	}
	return events, total_count, err
}

func extractFilter(filter request.AlertFilter, instances []*model.ServiceInstance) *whereSQL {
	var whereInstance []*whereSQL
	if len(filter.Group) == 0 || filter.Group == "app" {
		whereGroup := EqualsIfNotEmpty("group", "app")
		whereInstance = append(whereInstance, MergeWheres(
			AndSep,
			whereGroup,
			Equals("tags['svc_name']", filter.Service),
		))
	}

	if len(filter.Group) == 0 || filter.Group == "container" {
		whereGroup := EqualsIfNotEmpty("group", "container")
		var k8sPods ValueInGroups = ValueInGroups{
			Keys: []string{"tags['namespace']", "tags['pod']"},
		}
		for _, instance := range instances {
			if instance == nil {
				continue
			}
			if len(instance.PodName) > 0 {
				k8sPods.ValueGroups = append(k8sPods.ValueGroups, clickhouse.GroupSet{
					Value: []any{instance.Namespace, instance.PodName},
				})
			}
		}

		whereInstance = append(whereInstance, MergeWheres(
			AndSep,
			whereGroup,
			InGroup(k8sPods),
		))
	}

	if len(filter.Group) == 0 || filter.Group == "network" {
		whereGroup := EqualsIfNotEmpty("group", "network")
		var k8sPods ValueInGroups = ValueInGroups{
			Keys: []string{"tags['src_namespace']", "tags['src_pod']"},
		}
		var vmPods ValueInGroups = ValueInGroups{
			Keys: []string{"tags['src_node']", "tags['pid']"},
		}

		for _, instance := range instances {
			if instance == nil {
				continue
			}
			if len(instance.PodName) > 0 {
				k8sPods.ValueGroups = append(k8sPods.ValueGroups, clickhouse.GroupSet{
					Value: []any{instance.Namespace, instance.PodName},
				})
			} else {
				vmPods.ValueGroups = append(vmPods.ValueGroups, clickhouse.GroupSet{
					Value: []any{instance.NodeName, instance.Pid},
				})
			}
		}

		k8sOrVm := MergeWheres(OrSep, InGroup(k8sPods), InGroup(vmPods))
		whereInstance = append(whereInstance, MergeWheres(
			AndSep,
			whereGroup,
			k8sOrVm,
		))
	}

	if len(filter.Group) == 0 || filter.Group == "infra" {
		whereGroup := EqualsIfNotEmpty("group", "infra")
		var tmpSet = map[string]struct{}{}
		var nodes clickhouse.ArraySet
		for _, instance := range instances {
			if instance == nil {
				continue
			}
			_, find := tmpSet[instance.NodeName]
			if !find {
				nodes = append(nodes, instance.NodeName)
				tmpSet[instance.NodeName] = struct{}{}
			}
		}

		whereInstance = append(whereInstance, MergeWheres(
			AndSep,
			whereGroup,
			In("tags['instance_name']", nodes),
		))
	}

	return MergeWheres(OrSep, whereInstance...)
}

type AlertEventSample struct {
	model.AlertEvent

	// 记录行号
	Rn         uint64 `ch:"rn" json:"-"`
	AlarmCount uint64 `ch:"alarm_count" json:"alarm_count"`
}

type PagedAlertEvent struct {
	model.AlertEvent

	// 记录行号
	Rn         uint64 `ch:"rn" json:"-"`
	TotalCount uint64 `ch:"total_count" json:"-"`
}

func RnLimit(p *request.PageParam) string {
	if p == nil {
		return ""
	}
	startIdx := 1 + (p.CurrentPage-1)*p.PageSize
	endIdx := p.CurrentPage * p.PageSize
	return fmt.Sprintf(" WHERE rn BETWEEN %d AND %d ", startIdx, endIdx)
}
