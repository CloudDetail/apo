package model

import (
	"fmt"
	"strings"
)

type Modifier struct {
	RangeVector string `json:"rangeVector,omitempty"` // 1m/1h/1d
	Scale       string `json:"scale,omitempty"`       // 倍率
	Quantile    string `json:"quantile,omitempty"`    // 百分位数
	DayOnDay    bool   `json:"dayOnDay"`              // 使用昨日同时段数据
	WeekOnWeek  bool   `json:"weekOnWeek"`            // 使用上周同时段数据
}

// PredefinedMetrics 预定义指标
type PredefinedMetric struct {
	Name   string `json:"name"`
	Metric string `json:"metric"`
}

// DetectExpr 检测表达式
type DetectExprPart struct {
	PredefinedMetric
	Modifier Modifier `json:"modifier"`

	// 用于记录指标所属的告警事件组
	Group string `json:"group"`

	// 用于记录可以相互比较的检测表达式
	CompareGroup string `json:"compareGroup"`

	// 常量表达式
	CustomMetric string `json:"customMetric"`

	// 用来记录指标的临时排序
	MetricIdx int `json:"-"`
}

type PredefinedMetricExpr struct {
	DetectExprPart

	Describe string `json:"describe"`
}

func (p *DetectExprPart) String() string {
	if p == nil {
		return ""
	}

	if len(p.CustomMetric) > 0 {
		return p.CustomMetric
	}

	if len(p.PredefinedMetric.Name) == 0 {
		// 非预定义指标,直接用原始值
		return p.PredefinedMetric.Metric
	}

	rawMetric := p.PredefinedMetric.Metric
	if len(p.Modifier.RangeVector) > 0 {
		rawMetric = strings.ReplaceAll(rawMetric, "%RangeVector%", p.Modifier.RangeVector)
	}
	if len(p.Modifier.Quantile) > 0 {
		rawMetric = strings.ReplaceAll(rawMetric, "%Quantile%", p.Modifier.Quantile)
	}

	if p.Modifier.DayOnDay {
		rawMetric = strings.ReplaceAll(rawMetric, "%OFFSET%", "offset 24h")
	} else if p.Modifier.WeekOnWeek {
		rawMetric = strings.ReplaceAll(rawMetric, "%OFFSET%", "offset 7d")
	} else {
		rawMetric = strings.ReplaceAll(rawMetric, "%OFFSET%", "")
	}

	if len(p.Modifier.Scale) > 0 {
		rawMetric = fmt.Sprintf("(%s) * %s", rawMetric, p.Modifier.Scale)
	}

	return rawMetric
}

// DetectMutation 执行的异常检测记录
// 新增异常检测记录时存入
type DetectMutation struct {
	DetectName              string `json:"name" ch:"name"`                                          // 告警名
	MutationCheck           string ` json:"mutationCheckPQL" ch:"mutation_check_pql"`               // 需要执行故障检测的语句
	For                     string `json:"for" ch:"for_duration"`                                   // 突变告警等待时间
	StartTime               int64  `json:"startTime" ch:"start_time"`                               // 查询开始时间
	EndTime                 int64  `json:"endTime" ch:"end_time"`                                   // 查询结束时间
	Timestamp               int64  `json:"timestamp" ch:"timestamp"`                                // 执行时间
	Step                    int64  `json:"step" ch:"step"`                                          // 查询步长(us)
	SynchronizeToAlertRules bool   `ch:"synchronize_to_alert_rules" json:"synchronizeToAlertRules"` // 是否同步到告警规则
	Group                   string `json:"group" ch:"group"`                                        // 告警组名
}
