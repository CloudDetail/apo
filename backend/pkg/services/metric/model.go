// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package metric

import (
	"encoding/json"

	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

type PreDefinedMetrics struct {
	Title     string     `json:"title"`
	Queries   []Query    `json:"queries"`
	Variables []Variable `json:"variables"`
}

type Query struct {
	ID       int    `json:"id"`
	GroupID  int    `json:"-"`
	Describe string `json:"describe"` // 描述

	Title   string   `json:"title"`   // 查询标题
	Targets []Target `json:"targets"` // 目标列表
	Params  []string `json:"params"`  // 参数列表
	Unit    string   `json:"unit"`    // 单位
}

type Target struct {
	// RefId        string   `json:"refId"`        // 子查询ID
	Expr         string   `json:"expr"`         // Prometheus 查询表达式
	LegendFormat string   `json:"legendFormat"` // 图例格式
	Variables    []string `json:"variables"`    // 使用的变量列表
}

type Variable struct {
	Name    string    `json:"name"`    // 变量名称
	Type    string    `json:"type"`    // 变量类型
	Label   string    `json:"label"`   // 变量标签
	Query   QueryDef  `json:"query"`   // 查询定义
	Options []Options `json:"options"` // 选项列表
	Current Options   `json:"current"` // 当前选中的值
	Regex   string    `json:"regex"`   // 正则表达式
}

// QueryDef 代表变量的查询定义
type QueryDef struct {
	QryType int    `json:"qryType"` // 查询类型
	Query   string `json:"query"`   // 查询字符串
}

type Options struct {
	// 根据实际需求扩展字段，例如：
	Selected bool     `json:"selected,omitempty"`
	Text     []string `json:"text,omitempty"`
	Value    []string `json:"value,omitempty"`
}

func (o *Options) UnmarshalJSON(data []byte) error {
	var options map[string]any
	if err := json.Unmarshal(data, &options); err != nil {
		return err
	}

	*o = Options{}
	for k, v := range options {
		switch k {
		case "selected":
			if selected, ok := v.(bool); ok {
				o.Selected = selected
			}
		case "text":
			if text, ok := v.([]string); ok {
				o.Text = text
			} else if text, ok := v.(string); ok {
				o.Text = []string{text}
			}
		case "value":
			if value, ok := v.([]string); ok {
				for i := 0; i < len(value); i++ {
					if value[i] == "$__all" {
						value[i] = ".*"
					}
				}
				o.Value = value
			} else if value, ok := v.(string); ok {
				if value == "$__all" {
					value = ".*"
				}
				o.Value = []string{value}
			}
		}
	}
	return nil
}

type QueryMetricsResult struct {
	Msg string `json:"msg"`

	Result  *QueryResult  `json:"result,omitempty"`
	Results []QueryResult `json:"results,omitempty"`
}

type QueryResult struct {
	// Query Query `json:"-"`

	Title      string       `json:"title"`
	Unit       string       `json:"unit"`
	Timeseries []Timeseries `json:"timeseries"`
}

type Labels = map[string]string

type Timeseries struct {
	Legend       string `json:"legend"`
	LegendFormat string `json:"legendFormat"`
	Labels       Labels `json:"labels"`

	// Values []model.SamplePair `json:"values"`

	Chart response.TempChartObject `json:"chart"`
}

type QueryMetricsRequest struct {
	MetricName  string            `json:"metricName"`
	MetricIds   []int             `json:"metricIDs,omitempty"`
	MetricNames []string          `json:"metricNames,omitempty"`
	Params      map[string]string `json:"params"`
	StartTime   int64             `json:"startTime"`
	EndTime     int64             `json:"endTime"`
	Step        int64             `json:"step"`
}
