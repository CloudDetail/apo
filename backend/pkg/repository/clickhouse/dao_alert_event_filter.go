// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

const SQL_SEARCH_ALERT_FILTER_KEYS = `WITH lastEvent AS (
  SELECT alert_id,status,raw_tags
  FROM alert_event ae
  %s
  ORDER BY received_time DESC LIMIT 1 BY alert_id
),
filtered_workflows AS (
  SELECT ref,output,
  CASE
    WHEN output = 'false' THEN 2
    WHEN output = 'true' THEN 1
    ELSE 0
  END as importance
  FROM workflow_records
  %s
  ORDER BY created_at DESC LIMIT 1 BY ref
),
combined_data AS (
  SELECT arrayJoin(mapKeys(raw_tags)) AS filter_key,
  CASE
    WHEN fw.importance = 0 and fw.output != '' THEN 'failed'
    WHEN fw.importance = 0 and ae.status = 'firing'  THEN 'unknown'
    WHEN fw.importance = 0 and ae.status = 'resolved' THEN 'skipped'
    WHEN fw.importance = 1 THEN 'invalid'
    WHEN fw.importance = 2 THEN 'valid'
  END as validity
  FROM lastEvent ae
  LEFT JOIN filtered_workflows fw ON ae.alert_id = fw.ref
  %s
)
SELECT
 DISTINCT filter_key
FROM combined_data
ORDER BY filter_key desc
`

const SQL_SEARCH_ALERT_FILTER_VALUES = `WITH lastEvent AS (
  SELECT alert_id,status, %s as filterValue
  FROM alert_event ae
  %s
  ORDER BY received_time DESC LIMIT 1 BY alert_id
),
filtered_workflows AS (
  SELECT ref,output,
  CASE
    WHEN output = 'false' THEN 2
    WHEN output = 'true' THEN 1
    ELSE 0
  END as importance
  FROM workflow_records
  %s
  ORDER BY created_at DESC LIMIT 1 BY ref
),
combined_data AS (
  SELECT filterValue,
  CASE
    WHEN fw.importance = 0 and fw.output != '' THEN 'failed'
    WHEN fw.importance = 0 and ae.status = 'firing'  THEN 'unknown'
    WHEN fw.importance = 0 and ae.status = 'resolved' THEN 'skipped'
    WHEN fw.importance = 1 THEN 'invalid'
    WHEN fw.importance = 2 THEN 'valid'
  END as validity
  FROM lastEvent ae
  LEFT JOIN filtered_workflows fw ON ae.alert_id = fw.ref
  %s
)
SELECT
  count(1) as count,
  filterValue
FROM combined_data
GROUP BY filterValue
ORDER BY count desc`

var (
	alertFilters    []request.AlertEventFilter
	alertFilters_en []request.AlertEventFilter
)

func init() {
	for key, filter := range staticFilters {
		filter.Key = key
		alertFilters = append(alertFilters, filter.AlertEventFilter)

		filter_en := filter.AlertEventFilter
		filter_en.Name = filter.Name_EN
		filter_en.Options = make([]request.AlertEventFilterOption, 0)
		for _, option := range filter.Options {
			filter_en.Options = append(filter_en.Options, request.AlertEventFilterOption{
				Value:   option.Value,
				Display: option.Value,
			})
		}
		alertFilters_en = append(alertFilters_en, filter_en)
	}
}

func (ch *chRepo) GetStaticFilterKeys(ctx core.Context) []request.AlertEventFilter {
	switch ctx.LANG() {
	case "en":
		return alertFilters_en
	default:
		return alertFilters
	}
}

func (ch *chRepo) GetAlertEventFilterLabelKeys(
	ctx core.Context,
	req *request.SearchAlertEventFilterValuesRequest,
) ([]string, error) {
	alertFilter := NewQueryBuilder().Between("received_time", req.StartTime/1e6, req.EndTime/1e6)
	recordFilter := NewQueryBuilder().Between("created_at", req.StartTime/1e6, req.EndTime/1e6)
	resultFilter := NewQueryBuilder()

	err := applyFilter(req.Filters, resultFilter, alertFilter)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(SQL_SEARCH_ALERT_FILTER_KEYS,
		alertFilter.String(),
		recordFilter.String(),
		resultFilter.String(),
	)

	values := make([]any, 0)
	values = append(values, alertFilter.values...)
	values = append(values, recordFilter.values...)
	values = append(values, resultFilter.values...)

	rows, err := ch.GetContextDB(ctx).Query(ctx.GetContext(), sql, values...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var result []string
	for rows.Next() {
		var filterKey string
		if err := rows.Scan(&filterKey); err != nil {
			return nil, err
		}
		result = append(result, fmt.Sprintf("labels.%s", filterKey))
	}
	return result, nil
}

func (ch *chRepo) GetAlertEventFilterValues(ctx core.Context, req *request.SearchAlertEventFilterValuesRequest) (*request.AlertEventFilter, error) {
	var staticFilters []request.AlertEventFilter
	lang := ctx.LANG()
	if lang == "en" {
		staticFilters = alertFilters_en
	} else {
		staticFilters = alertFilters
	}

	switch req.SearchKey {
	case "severity", "status", "validity":
		for _, filter := range staticFilters {
			if filter.Key == req.SearchKey {
				return &filter, nil
			}
		}
	}

	alertFilter := NewQueryBuilder().Between("received_time", req.StartTime/1e6, req.EndTime/1e6)
	recordFilter := NewQueryBuilder().Between("created_at", req.StartTime/1e6, req.EndTime/1e6)
	resultFilter := NewQueryBuilder()

	err := applyFilter(req.Filters, resultFilter, alertFilter)
	if err != nil {
		return nil, err
	}

	var targetKey string = req.SearchKey
	if strings.HasPrefix(req.SearchKey, "labels.") {
		tagKey := req.SearchKey[7:]
		if !allowFilterKey.MatchString(tagKey) {
			return nil, fmt.Errorf("filter key %s not allowed", req.SearchKey)
		}
		targetKey = fmt.Sprintf(`raw_tags['%s']`, tagKey)
	} else if strings.HasPrefix(req.SearchKey, "tags.") {
		tagKey := req.SearchKey[5:]
		if !allowFilterKey.MatchString(tagKey) {
			return nil, fmt.Errorf("filter key %s not allowed", req.SearchKey)
		}
		targetKey = fmt.Sprintf(`tags['%s']`, tagKey)
	}

	sql := fmt.Sprintf(SQL_SEARCH_ALERT_FILTER_VALUES,
		targetKey,
		alertFilter.String(),
		recordFilter.String(),
		resultFilter.String(),
	)

	values := make([]any, 0)
	values = append(values, alertFilter.values...)
	values = append(values, recordFilter.values...)
	values = append(values, resultFilter.values...)

	var filterValues []filterValue
	err = ch.GetContextDB(ctx).Select(ctx.GetContext(), &filterValues, sql, values...)
	if err != nil {
		return nil, err
	}

	var res request.AlertEventFilter
	res.Key = req.SearchKey
	res.Options = make([]request.AlertEventFilterOption, 0)

	for _, filterValue := range filterValues {
		if filterValue.Value == "" {
			continue
		}
		res.Options = append(res.Options, request.AlertEventFilterOption{
			Value:   filterValue.Value,
			Display: filterValue.Value,
		})
	}
	return &res, nil
}

func applyFilter(filters []request.AlertEventFilter, resultFilter *QueryBuilder, alertFilter *QueryBuilder) error {
	for _, filter := range filters {
		if filter.Key == "" {
			continue
		}
		if filter.Key == "validity" {
			for _, v := range filter.Selected {
				if v == "other" {
					filter.Selected = append(filter.Selected, "unknown", "failed", "skipped")
					break
				}
			}

			resultFilter.InStrings("validity", filter.Selected)
			continue
		} else if filter.Key == "status" {
			resultFilter.InStrings("status", filter.Selected)
			continue
		} else if strings.HasPrefix(filter.Key, "workflow.") {
			rawFieldKey := filter.Key[9:]
			if allowFilterKey.MatchString(rawFieldKey) {
				if len(filter.MatchExpr) > 0 {
					if filter.MatchExpr == "*" {
						resultFilter.NotEquals(rawFieldKey, "").
							NotEquals(rawFieldKey, "failed: status: failed, output: null")
					} else {
						resultFilter.Like(rawFieldKey, strings.Replace(filter.MatchExpr, "*", "%", -1))
					}
				} else {
					resultFilter.InStrings(rawFieldKey, filter.Selected)
				}
			}
			continue
		}
		subSql, err := extractAlertEventFilter(&filter)
		if err != nil {
			return fmt.Errorf("illegal filter: %s", filter.Key)
		}
		alertFilter.And(subSql)
	}
	return nil
}

type filterValue struct {
	Value string `ch:"filterValue"`
	Count uint64 `ch:"count"`
}

var allowFilterKey = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func extractAlertEventFilter(filter *request.AlertEventFilter) (*whereSQL, error) {
	if strings.HasPrefix(filter.Key, "labels.") {
		rawTagKey := filter.Key[7:]
		if allowFilterKey.MatchString(rawTagKey) {
			if len(filter.MatchExpr) > 0 && len(filter.Selected) == 0 {
				return like(fmt.Sprintf("raw_tags['%s']", rawTagKey), strings.Replace(filter.MatchExpr, "*", "%", -1)), nil
			}
			return inStrings(fmt.Sprintf("raw_tags['%s']", rawTagKey), filter.Selected), nil
		}
		return nil, fmt.Errorf("filter key %s not allowed", filter.Key)
	}

	if _, find := staticFilters[filter.Key]; !find {
		return nil, fmt.Errorf("filter key %s not found", filter.Key)
	}

	if strings.HasPrefix(filter.Key, "tags.") {
		return inStrings(fmt.Sprintf("tags['%s']", filter.Key[5:]), filter.Selected), nil
	}
	if len(filter.MatchExpr) > 0 && len(filter.Selected) == 0 {
		return like(filter.Key, strings.Replace(filter.MatchExpr, "*", "%", -1)), nil
	}
	return inStrings(filter.Key, filter.Selected), nil
}

type _staticFilters struct {
	request.AlertEventFilter
	Name_EN string
}

var staticFilters map[string]_staticFilters = map[string]_staticFilters{
	"name": {
		AlertEventFilter: request.AlertEventFilter{Name: "告警事件名", Wildcard: true},
		Name_EN:          "Alert Name",
	},
	"group": {
		AlertEventFilter: request.AlertEventFilter{Name: "告警类型"},
		Name_EN:          "Alert Type",
	},
	"severity": {
		AlertEventFilter: request.AlertEventFilter{
			Name: "告警级别",
			Options: []request.AlertEventFilterOption{
				{Value: "unknown", Display: "未知"},
				{Value: "info", Display: "信息"},
				{Value: "warning", Display: "警告"},
				{Value: "error", Display: "错误"},
				{Value: "critical", Display: "严重"},
			},
		},
		Name_EN: "Severity",
	},
	"status": {
		AlertEventFilter: request.AlertEventFilter{
			Name: "告警状态",
			Options: []request.AlertEventFilterOption{
				{Value: "firing", Display: "告警中"},
				{Value: "resolved", Display: "已恢复"},
			},
		},
		Name_EN: "Status",
	},
	"validity": {
		AlertEventFilter: request.AlertEventFilter{
			Name: "告警有效性",
			Options: []request.AlertEventFilterOption{
				{Value: "valid", Display: "有效"},
				{Value: "invalid", Display: "无效"},
				{Value: "other", Display: "其他"},
				// {Value: "unknown", Display: "未知"},
				// {Value: "failed", Display: "失败"},
				// {Value: "skipped", Display: "跳过检查"},
			},
		},
		Name_EN: "Validity",
	},
	"source": {
		AlertEventFilter: request.AlertEventFilter{Name: "告警源"},
		Name_EN:          "Alert Source",
	},
	"tags.serviceName": {
		AlertEventFilter: request.AlertEventFilter{Name: "服务名"},
		Name_EN:          "Service Name",
	},
	"tags.endpoint": {
		AlertEventFilter: request.AlertEventFilter{Name: "服务端点"},
		Name_EN:          "Service Endpoint",
	},
	"tags.namespace": {
		AlertEventFilter: request.AlertEventFilter{Name: "命名空间"},
		Name_EN:          "Namespace",
	},
	"tags.pod": {
		AlertEventFilter: request.AlertEventFilter{Name: "POD名"},
		Name_EN:          "POD Name",
	},
	"tags.node": {
		AlertEventFilter: request.AlertEventFilter{Name: "主机名"},
		Name_EN:          "Hostname",
	},
	"tags.pid": {
		AlertEventFilter: request.AlertEventFilter{Name: "进程PID"},
		Name_EN:          "Process PID",
	},
}

func _extractAlertEventFilter(filter *alert.AlertEventFilter) *whereSQL {
	if filter == nil {
		return ALWAYS_TRUE
	}

	var basicFilters []*whereSQL
	basicFilters = append(basicFilters,
		equalsIfNotEmpty("source", filter.Source),
		// EqualsIfNotEmpty("group", filter.Group),
		equalsIfNotEmpty("name", filter.Name),
		equalsIfNotEmpty("id", filter.EventID),
		equalsIfNotEmpty("severity", filter.Severity),
		equalsIfNotEmpty("status", filter.Status),
	)

	if len(filter.GroupIDs) > 0 {
		basicFilters = append(basicFilters, inStrings("raw_tags['groupId']", filter.GroupIDs))
	}

	if len(filter.Group) > 0 && filter.WithMutation {
		basicFilters = append(basicFilters,
			in("group", clickhouse.ArraySet{
				filter.Group,
				"mutation-" + filter.Group,
			}))
	} else if len(filter.Group) > 0 {
		basicFilters = append(basicFilters, equals("group", filter.Group))
	}

	if !filter.WithMutation {
		basicFilters = append(basicFilters, notLike("group", "mutation%"))
	}

	basicSQL := mergeWheres(AndSep, basicFilters...)

	if filter.AlertTagsFilter == nil {
		return basicSQL
	}

	// TODO use tagFilter to decrease events
	// return MergeWheres(AndSep, basicSQL, extractAlertTagsFilter(filter.AlertTagsFilter))
	return basicSQL
}
