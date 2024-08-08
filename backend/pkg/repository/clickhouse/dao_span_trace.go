package clickhouse

import (
	"context"
	"fmt"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

const (
	TEMPLATE_COUNT_SPAN_TRACE = "SELECT count(1) as total FROM span_trace %s"
	TEMPLATE_QUERY_SPAN_TRACE = "SELECT %s FROM span_trace %s %s"
)

func (ch *chRepo) GetFaultLogPageList(query *FaultLogQuery) ([]FaultLogResult, int64, error) {
	queryBuilder := NewQueryBuilder().
		Between("timestamp", query.StartTime/1000000, query.EndTime/1000000).
		EqualsNotEmpty("labels['service_name']", query.Service).
		EqualsNotEmpty("labels['instance_id']", query.Instance).
		EqualsNotEmpty("labels['content_key']", query.EndPoint).
		EqualsNotEmpty("labels['node_name']", query.NodeName).
		EqualsNotEmpty("trace_id", query.TraceId)
	if len(query.ContainerId) > 0 {
		queryBuilder.Equals("labels['container_id']", query.ContainerId)
	} else if query.Pid > 0 {
		queryBuilder.Equals("pid", query.Pid)
	}
	if query.Type == 1 {
		queryBuilder.Statement("flags['is_error'] = true")
	} else if query.Type == 2 {
		queryBuilder.Statement("(flags['is_error'] = true or flags['is_profiled'] = true or flags['is_slow'] = true)")
	} else {
		queryBuilder.Statement("(flags['is_error'] = true or (flags['is_profiled'] = true AND flags['is_slow'] = true))")
	}
	whereClause := queryBuilder.String()
	var countResults []QueryCount
	// 查询记录数
	err := ch.conn.Select(context.Background(), &countResults, fmt.Sprintf(TEMPLATE_COUNT_SPAN_TRACE, whereClause), queryBuilder.values...)
	if err != nil {
		return nil, 0, err
	}

	count := int64(countResults[0].Total)

	fieldSql := NewFieldBuilder().
		Fields("trace_id", "pid").
		Alias("intDiv(start_time, 1000)", "start_time_us").
		Alias("intDiv(end_time, 1000)", "end_time_us").
		Alias("labels['content_key']", "endpoint").
		Alias("labels['pod_name']", "pod_name").
		Alias("labels['node_name']", "node_name").
		Alias("labels['container_id']", "container_id").
		Alias("labels['instance_id']", "instance_id").
		Alias("labels['service_name']", "service_name").
		String()
	bySql := NewByLimitBuilder().
		OrderBy("timestamp", false).
		Limit(query.PageSize).
		Offset((query.PageNum - 1) * query.PageSize).
		String()
	var result []FaultLogResult
	// 查询列表数据
	sql := fmt.Sprintf(TEMPLATE_QUERY_SPAN_TRACE, fieldSql, whereClause, bySql)
	err = ch.conn.Select(context.Background(), &result, sql, queryBuilder.values...)
	if err != nil {
		return nil, count, err
	}
	return result, count, nil
}

func (ch *chRepo) GetTracePageList(req *request.GetTracePageListRequest) ([]QueryTraceResult, int64, error) {
	queryBuilder := NewQueryBuilder().
		Between("timestamp", req.StartTime/1000000, req.EndTime/1000000).
		EqualsNotEmpty("labels['service_name']", req.Service).
		EqualsNotEmpty("labels['content_key']", req.EndPoint).
		EqualsNotEmpty("labels['instance_id']", req.Instance).
		EqualsNotEmpty("labels['node_name']", req.NodeName).
		EqualsNotEmpty("trace_id", req.TraceId)
	if len(req.ContainerId) > 0 {
		queryBuilder.Equals("labels['container_id']", req.ContainerId)
	} else if req.Pid > 0 {
		queryBuilder.Equals("pid", req.Pid)
	}
	whereClause := queryBuilder.String()
	var countResults []QueryCount
	// 查询记录数
	err := ch.conn.Select(context.Background(), &countResults, fmt.Sprintf(TEMPLATE_COUNT_SPAN_TRACE, whereClause), queryBuilder.values...)
	if err != nil {
		return nil, 0, err
	}

	count := int64(countResults[0].Total)

	fieldSql := NewFieldBuilder().
		Fields("trace_id").
		Alias("toUnixTimestamp64Micro(timestamp)", "ts").
		Alias("intDiv(duration, 1000)", "duration_us").
		Alias("labels['content_key']", "endpoint").
		Alias("labels['service_name']", "service_name").
		Alias("labels['instance_id']", "instance_id").
		Alias("flags['is_error']", "is_error").
		String()
	bySql := NewByLimitBuilder().
		OrderBy("timestamp", false).
		Limit(req.PageSize).
		Offset((req.PageNum - 1) * req.PageSize).
		String()
	var result []QueryTraceResult
	// 查询列表数据
	sql := fmt.Sprintf(TEMPLATE_QUERY_SPAN_TRACE, fieldSql, whereClause, bySql)
	err = ch.conn.Select(context.Background(), &result, sql, queryBuilder.values...)
	if err != nil {
		return nil, count, err
	}
	return result, count, nil
}

type FaultLogQuery struct {
	StartTime   int64
	EndTime     int64
	Service     string
	Instance    string
	NodeName    string
	ContainerId string
	Pid         uint32
	EndPoint    string
	TraceId     string
	PageNum     int
	PageSize    int
	Type        int // 0 - slow & error, 1 - error
}

type QueryCount struct {
	Total uint64 `ch:"total"`
}

type FaultLogResult struct {
	ServiceName string `ch:"service_name" json:"serviceName"`
	InstanceId  string `ch:"instance_id" json:"instanceId"`
	TraceId     string `ch:"trace_id" json:"traceId"`
	StartTime   uint64 `ch:"start_time_us" json:"startTime"`
	EndTime     uint64 `ch:"end_time_us" json:"endTime"`
	EndPoint    string `ch:"endpoint" json:"endpoint"`
	PodName     string `ch:"pod_name" json:"podName"`
	ContainerId string `ch:"container_id" json:"containerId"`
	NodeName    string `ch:"node_name" json:"nodeName"`
	Pid         uint32 `ch:"pid" json:"pid"`
}

type QueryTraceResult struct {
	Timestamp   int64  `ch:"ts" json:"timestamp"`
	Duration    uint64 `ch:"duration_us" json:"duration"`
	ServiceName string `ch:"service_name" json:"serviceName"`
	TraceId     string `ch:"trace_id" json:"traceId"`
	EndPoint    string `ch:"endpoint" json:"endpoint"`
	InstanceId  string `ch:"instance_id" json:"instanceId"`
	IsError     bool   `ch:"is_error" json:"isError"`
}
