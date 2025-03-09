// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package metric

import "fmt"

type QueryDict struct {
	querys   []Query
	queryMap map[string][]Query

	groupedQuerys []*PreDefinedMetrics

	// variableGroup []map[string]Variable
	// groupName     []string

	listQuerys []QueryInfo
}

type QueryInfo struct {
	ID      int `json:"id"`
	GroupID int `json:"-"`

	Title    string `json:"title"`    // 查询标题
	Describe string `json:"describe"` // 描述
	// Targets []Target `json:"targets"` // 目标列表
	Params []string `json:"params"` // 参数列表
	Unit   string   `json:"unit"`   // 单位
}

func (q *QueryDict) ListMetrics() []QueryInfo {
	return q.listQuerys
}

func (q *QueryDict) ListQuerys() []Query {
	return q.querys
}

func (q *QueryDict) GetVarSpec(groupID int, variableName string) (*Variable, bool) {
	if groupID < len(q.groupedQuerys) {
		variables := q.groupedQuerys[groupID].Variables
		for i := 0; i < len(variables); i++ {
			if variables[i].Name == variableName {
				return &variables[i], true
			}
		}
	}

	return nil, false
}

func (q *QueryDict) GetQuerysByIds(queryIDs []int) []Query {
	var res []Query
	for _, id := range queryIDs {
		if id < len(q.querys) {
			res = append(res, q.querys[id])
		}
	}
	return res
}

func (q *QueryDict) GetQuerysByNames(names []string) []Query {
	var res []Query

	for _, name := range names {
		if querys, find := q.queryMap[name]; find {
			res = append(res, querys...)
		}
	}

	return res
}

func (q *QueryDict) AddPreDefinedMetrics(metrics *PreDefinedMetrics) {
	groupId := len(q.groupedQuerys)

	for i := 0; i < len(metrics.Queries); i++ {
		if len(metrics.Queries[i].Targets) == 1 {
			metrics.Queries[i].ID = len(q.querys)
			metrics.Queries[i].GroupID = groupId
			metrics.Queries[i].Title = fmt.Sprintf("%s - %s", metrics.Title, metrics.Queries[i].Title)
			q.querys = append(q.querys, metrics.Queries[i])

			if querys, find := q.queryMap[metrics.Queries[i].Title]; find {
				querys = append(querys, metrics.Queries[i])
				q.queryMap[metrics.Queries[i].Title] = querys
			} else {
				q.queryMap[metrics.Queries[i].Title] = []Query{metrics.Queries[i]}
			}

			q.listQuerys = append(q.listQuerys, QueryInfo{
				ID:       metrics.Queries[i].ID,
				GroupID:  metrics.Queries[i].GroupID,
				Title:    metrics.Queries[i].Title,
				Params:   metrics.Queries[i].Params,
				Unit:     metrics.Queries[i].Unit,
				Describe: metrics.Queries[i].Describe,
			})
		}
	}

	variableGroup := make(map[string]Variable)
	for i := 0; i < len(metrics.Variables); i++ {
		variableGroup[metrics.Variables[i].Name] = metrics.Variables[i]
	}

	q.groupedQuerys = append(q.groupedQuerys, metrics)
}

// func (q *QueryDict) AddDashboards(dashboardData []byte) {

// }

// func extractQuery(dashboardData []byte) {

// }
