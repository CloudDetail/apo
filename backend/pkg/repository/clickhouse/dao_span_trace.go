// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

const (
	TEMPLATE_COUNT_SPAN_TRACE = "SELECT count(1) as total FROM span_trace %s"
	TEMPLATE_QUERY_SPAN_TRACE = "SELECT %s FROM span_trace %s %s"

	SQL_GET_LABEL_FILTER_KEYS = `SELECT DISTINCT
    key, 'string' as data_type , 'labels' as parent_field
	FROM span_trace st
	ARRAY JOIN mapKeys(labels) AS key
	%s %s`

	SQL_GET_FLAGS_FILTER_KEYS = `SELECT DISTINCT
    key, 'bool' as data_type , 'flags' as parent_field
	FROM span_trace st
	ARRAY JOIN mapKeys(flags) AS key
	%s %s`

	SQL_GET_FILTER_VALUES = `SELECT DISTINCT
	%s as label_value
	FROM span_trace st
	%s %s`
)

func (ch *chRepo) GetFaultLogPageList(query *FaultLogQuery) ([]FaultLogResult, int64, error) {
	queryBuilder := NewQueryBuilder().
		Between("timestamp", query.StartTime/1000000, query.EndTime/1000000).
		EqualsNotEmpty("labels['service_name']", query.Service).
		EqualsNotEmpty("labels['instance_id']", query.Instance).
		EqualsNotEmpty("labels['content_key']", query.EndPoint).
		EqualsNotEmpty("labels['node_name']", query.NodeName).
		EqualsNotEmpty("trace_id", query.TraceId).
		EqualsNotEmpty("labels['pod_name']", query.Pod)

	queryBuilder.baseQuery = "SELECT count(1) as total FROM span_trace "
	if len(query.MultiServices) > 0 {
		queryBuilder.In("labels['service_name']", query.MultiServices)
	}
	if len(query.MultiNamespace) > 0 {
		queryBuilder.In("labels['namespace']", query.MultiNamespace)
	}
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

	var count uint64
	// Number of query records
	err := ch.conn.QueryRow(context.Background(), queryBuilder.String(), queryBuilder.values...).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	result := []FaultLogResult{}
	if count == 0 {
		return result, 0, nil
	}

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
	// Query list data
	sql := buildSpanTraceQuery(TEMPLATE_QUERY_SPAN_TRACE, fieldSql, bySql, queryBuilder)
	err = ch.conn.Select(context.Background(), &result, sql, queryBuilder.values...)
	if err != nil {
		return nil, int64(count), err
	}
	return result, int64(count), nil
}

func (ch *chRepo) GetAvailableFilterKey(startTime, endTime time.Time, needUpdate bool) ([]request.SpanTraceFilter, error) {
	if needUpdate {
		filers, err := ch.UpdateFilterKey(startTime, endTime)
		if err != nil {
			return []request.SpanTraceFilter{}, err
		}
		return filers, nil
	}

	now := time.Now()
	if len(ch.Filters) == 0 || now.Sub(ch.FilterUpdateTime) > 48*time.Hour {
		filters, err := ch.UpdateFilterKey(now.Add(-48*time.Hour), now)
		if err != nil {
			return []request.SpanTraceFilter{}, err
		}
		ch.SetAvailableFilters(filters, now)
		return ch.Filters, nil
	}

	return ch.Filters, nil
}

func buildSpanTraceQuery(baseQuery string, fieldSql string, bySql string, builder *QueryBuilder) string {
	sql := fmt.Sprintf(baseQuery, fieldSql, builder.String(), bySql)
	return sql
}

func (ch *chRepo) GetTracePageList(req *request.GetTracePageListRequest) ([]QueryTraceResult, int64, error) {
	queryBuilder := NewQueryBuilder().
		Between("timestamp", req.StartTime/1000000, req.EndTime/1000000).
		EqualsNotEmpty("labels['content_key']", req.EndPoint).
		EqualsNotEmpty("labels['instance_id']", req.Instance).
		EqualsNotEmpty("labels['node_name']", req.NodeName).
		EqualsNotEmpty("trace_id", req.TraceId).
		EqualsNotEmpty("labels['container_id']", req.ContainerId)

	if len(req.Service) > 0 {
		queryBuilder.In("labels['service_name']", req.Service)
	}

	if len(req.Namespace) > 0 {
		queryBuilder.In("labels['namespace']", req.Namespace)
	}

	if req.Pid > 0 {
		queryBuilder.Equals("pid", req.Pid)
	}

	for _, filter := range req.Filters {
		queryBuilder.And(ch.extractSpanFilter(filter))
	}

	queryBuilder.baseQuery = "SELECT count(1) as total FROM span_trace "
	var count uint64
	// Number of query records
	err := ch.conn.QueryRow(context.Background(), queryBuilder.String(), queryBuilder.values...).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	result := []QueryTraceResult{}
	if count == 0 {
		return result, 0, nil
	}

	fieldSql := NewFieldBuilder().
		Fields("trace_id").
		Fields("pid").
		Fields("tid").
		Alias("toUnixTimestamp64Micro(timestamp)", "ts").
		Alias("intDiv(duration, 1000)", "duration_us").
		Alias("labels['content_key']", "endpoint").
		Alias("labels['service_name']", "service_name").
		Alias("labels['instance_id']", "instance_id").
		Alias("flags['is_error']", "is_error").
		Alias("flags", "flags").
		Alias("labels", "labels").
		Alias("apm_span_id", "span_id").
		String()
	bySql := NewByLimitBuilder().
		OrderBy("timestamp", false).
		Limit(req.PageSize).
		Offset((req.PageNum - 1) * req.PageSize).
		String()
	// Query list data
	sql := buildSpanTraceQuery(TEMPLATE_QUERY_SPAN_TRACE, fieldSql, bySql, queryBuilder)
	err = ch.conn.Select(context.Background(), &result, sql, queryBuilder.values...)
	if err != nil {
		return nil, int64(count), err
	}
	return result, int64(count), nil
}

type FaultLogQuery struct {
	StartTime      int64
	EndTime        int64
	Service        string
	Instance       string
	NodeName       string
	ContainerId    string
	Pid            uint32
	EndPoint       string
	TraceId        string
	PageNum        int
	PageSize       int
	Type           int      // 0 - slow & error, 1 - error
	MultiServices  []string // Match multiple service
	MultiNamespace []string // Match multiple namespace
	Pod            string   // Pod name
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
	Timestamp         int64   `ch:"ts" json:"timestamp"`
	Duration          uint64  `ch:"duration_us" json:"duration"`
	ServiceName       string  `ch:"service_name" json:"serviceName"`
	Pid               uint32  `ch:"pid" json:"pid"`
	Tid               uint32  `ch:"tid" json:"tid"`
	TraceId           string  `ch:"trace_id" json:"traceId"`
	EndPoint          string  `ch:"endpoint" json:"endpoint"`
	InstanceId        string  `ch:"instance_id" json:"instanceId"`
	SpanId            string  `ch:"span_id" json:"spanId"`
	ApmType           string  `ch:"apm_type" json:"apmType"`
	Reason            string  `ch:"reason" json:"reason"`
	IsError           bool    `ch:"is_error" json:"isError"`
	IsSlow            bool    `ch:"is_slow" json:"isSlow"`
	ThresholdValue    float64 `ch:"threshold_value" json:"thresholdValue"`
	ThresholdMultiple float64 `ch:"threshold_multiple" json:"thresholdMultiple"`

	Labels  map[string]string `ch:"labels" json:"labels"`
	Flags   map[string]bool   `ch:"flags"  json:"flags"`
	Metrics map[string]uint64 `ch:"metrics" json:"metrics"`

	MutatedValue uint64 `ch:"mutated_value" json:"mutatedValue"`
	IsMutated    uint8  `ch:"is_mutated" json:"isMutated"` // whether the delay changes abruptly
}

func (af *availableFilters) extractSpanFilter(f *request.ComplexSpanTraceFilter) *whereSQL {
	if !af.ValidCheckAndAdjust(f) {
		return ALWAYS_FALSE
	}

	if len(f.SpanTraceFilters) > 0 {
		whereSQLs := make([]*whereSQL, 0, len(f.SpanTraceFilters))
		for _, subFilter := range f.SpanTraceFilters {
			whereSQLs = append(whereSQLs, af.extractSpanFilter(subFilter))
		}
		return mergeWheres(getMergeSep(f.MergeSep), whereSQLs...)
	}

	if f.SpanTraceFilter == nil {
		return ALWAYS_TRUE
	}

	key := formatFieldName(f.SpanTraceFilter)
	param, success := extractFilterParams(f.SpanTraceFilter)

	if !success {
		return ALWAYS_FALSE
	}

	switch f.Operation {
	case request.OpEqual:
		return equals(key, param[0])
	case request.OpNotEqual:
		return notEquals(key, param[0])
	case request.OpIn:
		return in(key, param)
	case request.OpNotIn:
		return notIn(key, param)
	case request.OpLike:
		return like(key, param[0])
	case request.OpNotLike:
		return notLike(key, param[0])
	case request.OpExists:
		return exists(key)
	case request.OpNotExists:
		return notExists(key)
	case request.OpContains:
		return contains(key, param[0])
	case request.OpNotContains:
		return notContains(key, param[0])
	case request.OpGreaterThan:
		return greaterThan(key, param[0])
	case request.OpLessThan:
		return lessThan(key, param[0])
	}

	return ALWAYS_FALSE
}

func formatFieldName(f *request.SpanTraceFilter) string {
	if f.ParentField == request.PF_Flags {
		return fmt.Sprintf("flags['%s']", f.Key)
	} else if f.ParentField == request.PF_Labels {
		return fmt.Sprintf("labels['%s']", f.Key)
	}
	return f.Key
}

func extractFilterParams(f *request.SpanTraceFilter) ([]any, bool) {
	var param []any

	if len(f.Value) == 0 &&
		f.Operation != request.OpExists &&
		f.Operation != request.OpNotExists {
		return nil, false
	}

	switch f.DataType {
	case request.U32Column, request.U64Column:
		for _, v := range f.Value {
			if i, err := strconv.ParseUint(v, 10, 64); err == nil {
				param = append(param, i)
			} else {
				return nil, false
			}
		}
	case request.I64Column:
		for _, v := range f.Value {
			if i, err := strconv.ParseInt(v, 10, 64); err == nil {
				param = append(param, i)
			} else {
				return nil, false
			}
		}
	case request.StringColumn:
		for _, v := range f.Value {
			param = append(param, v)
		}
	case request.BoolColumn:
		for _, v := range f.Value {
			if b, err := strconv.ParseBool(v); err == nil {
				param = append(param, b)
			} else {
				return nil, false
			}
		}
	}
	return param, true
}

type SpanTraceOptions struct {
	request.SpanTraceFilter

	Options any `json:"options"`
}

var const_span_filter = []request.SpanTraceFilter{
	{
		Key:      "pid",
		DataType: request.U32Column,
	},
	{
		Key:      "tid",
		DataType: request.U32Column,
	},
	{
		Key:      "duration",
		DataType: request.U64Column,
	},
	{
		Key:      "end_time",
		DataType: request.U64Column,
	},
	{
		Key:      "start_time",
		DataType: request.U64Column,
	},
}

func (ch *chRepo) UpdateFilterKey(startTime, endTime time.Time) ([]request.SpanTraceFilter, error) {
	builder := NewQueryBuilder().
		Between("timestamp", startTime.Unix(), endTime.Unix())

	byLimits := NewByLimitBuilder().
		Limit(1000).
		OrderBy("timestamp", false)

	sql := fmt.Sprintf(SQL_GET_LABEL_FILTER_KEYS, builder.String(), byLimits.String())
	var labelRes []request.SpanTraceFilter
	err := ch.GetConn().Select(context.Background(), &labelRes, sql, builder.values...)
	if err != nil {
		return nil, err
	}

	sql = fmt.Sprintf(SQL_GET_FLAGS_FILTER_KEYS, builder.String(), byLimits.String())
	var flagRes []request.SpanTraceFilter
	err = ch.GetConn().Select(context.Background(), &flagRes, sql, builder.values...)
	if err != nil {
		return nil, err
	}

	filters := append(const_span_filter, labelRes...)
	filters = append(filters, flagRes...)
	return filters, nil
}

func (ch *chRepo) GetFieldValues(searchText string, filter *request.SpanTraceFilter, startTime, endTime time.Time) (*SpanTraceOptions, error) {
	if filter.DataType == request.BoolColumn {
		return &SpanTraceOptions{SpanTraceFilter: *filter, Options: []bool{true, false}}, nil
	}

	var field string
	if len(filter.ParentField) > 0 {
		field = fmt.Sprintf("%s['%s']", filter.ParentField, filter.Key)
	} else {
		field = filter.Key
	}

	if !ch.CheckField(field) {
		return nil, fmt.Errorf("field '%s' is invalid", field)
	}

	builder := NewQueryBuilder().
		Between("timestamp", startTime.Unix(), endTime.Unix())

	if filter.DataType == request.StringColumn && len(searchText) > 0 {
		builder.Like(field, searchText+"%")
	}

	byLimits := NewByLimitBuilder().
		Limit(100).
		OrderBy("label_value", false)

	sql := fmt.Sprintf(SQL_GET_FILTER_VALUES, field, builder.String(), byLimits.String())

	rows, err := ch.GetConn().Query(context.Background(), sql, builder.values...)
	if err != nil {
		return nil, err
	}

	var res any
	switch filter.DataType {
	case request.U32Column:
		var numOptions []uint32
		for rows.Next() {
			var value uint32
			if err := rows.Scan(&value); err != nil {
				log.Println(err)
			}
			numOptions = append(numOptions, value)
			res = numOptions
		}
	case request.U64Column:
		var numOptions []uint64
		for rows.Next() {
			var value uint64
			if err := rows.Scan(&value); err != nil {
				log.Println(err)
			}
			numOptions = append(numOptions, value)
			res = numOptions
		}
	case request.I64Column:
		var numOptions []int64
		for rows.Next() {
			var value int64
			if err := rows.Scan(&value); err != nil {
				log.Println(err)
			}
			numOptions = append(numOptions, value)
			res = numOptions
		}
	case request.StringColumn:
		var strOptions []string
		for rows.Next() {
			var value string
			if err := rows.Scan(&value); err != nil {
				log.Println(err)
			}
			strOptions = append(strOptions, value)
			res = strOptions
		}
	}

	return &SpanTraceOptions{
		SpanTraceFilter: *filter,
		Options:         res,
	}, nil
}

func (af *availableFilters) ValidCheckAndAdjust(f *request.ComplexSpanTraceFilter) bool {
	if len(f.SpanTraceFilters) > 0 {
		for _, filter := range f.SpanTraceFilters {
			if !af.ValidCheckAndAdjust(filter) {
				return false
			}
		}
		return true
	}

	if f.SpanTraceFilter == nil {
		return true
	}

	field := formatFieldName(f.SpanTraceFilter)
	if !af.CheckField(field) {
		return false
	}

	switch f.Key {
	case "duration":
		for i := 0; i < len(f.Value); i++ {
			f.Value[i] += "000"
		}
	}
	return true
}

type availableFilters struct {
	Filters          []request.SpanTraceFilter
	Keys             []string
	FilterUpdateTime time.Time
}

func (f *availableFilters) SetAvailableFilters(filters []request.SpanTraceFilter, updateTime time.Time) {
	f.Filters = filters
	f.Keys = make([]string, 0)
	f.Keys = append(f.Keys, "flags['is_error']", "flags['is_slow']")
	for _, filter := range filters {
		var field string
		if len(filter.ParentField) > 0 {
			field = fmt.Sprintf("%s['%s']", filter.ParentField, filter.Key)
		} else {
			field = filter.Key
		}
		f.Keys = append(f.Keys, field)
	}
	f.FilterUpdateTime = updateTime
}

func (f *availableFilters) CheckField(field string) bool {
	for _, key := range f.Keys {
		if key == field {
			return true
		}
	}
	return false
}
