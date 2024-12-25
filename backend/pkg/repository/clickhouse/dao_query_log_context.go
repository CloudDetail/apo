// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"fmt"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

var keys = []string{"source", "container_id", "pid", "container_name", "host_ip", "host_name", "k8s_namespace_name", "k8s_pod_name"}

func isKey(key string) bool {
	for _, k := range keys {
		if k == key {
			return true
		}
	}
	return false
}

func tagsCondition(tags map[string]string) string {
	var res string
	for k, v := range tags {
		if isKey(k) {
			res += fmt.Sprintf(`AND %s='%s'`, k, v)
		}
	}
	if res == "" {
		res = "AND (1='1')"
	}

	return res
}

func reverseSlice(s []map[string]any) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
func (ch *chRepo) QueryLogContext(req *request.LogQueryContextRequest) ([]map[string]any, []map[string]any, error) {
	//condition := NewQueryCondition(req.StartTime, req.EndTime, req.TimeField, req.Query)
	logtime := req.Time / 1000000
	timefront := fmt.Sprintf("toUnixTimestamp(timestamp) < %d AND  toUnixTimestamp(timestamp) > %d ", logtime, logtime-60)
	tags := tagsCondition(req.Tags)
	//查前面50条，反转
	bySqlfront := NewByLimitBuilder().
		OrderBy("timestamp", false).
		Limit(50).
		String()
	frontSql := fmt.Sprintf(querySQl, req.DataBase, req.TableName, timefront+tags, bySqlfront)
	front, err := ch.queryRowsData(frontSql)
	if err != nil {
		front = []map[string]any{}
	}
	reverseSlice(front)
	timeend := fmt.Sprintf("toUnixTimestamp(timestamp) >= %d AND toUnixTimestamp(timestamp) < %d ", logtime, logtime+60)
	bySqlend := NewByLimitBuilder().
		OrderBy("timestamp", true).
		Limit(50).
		String()
	endSql := fmt.Sprintf(querySQl, req.DataBase, req.TableName, timeend+tags, bySqlend)
	end, err := ch.queryRowsData(endSql)
	if err != nil {
		end = []map[string]any{}
	}

	return front, end, nil

}
