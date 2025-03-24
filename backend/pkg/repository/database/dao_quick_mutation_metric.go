package database

import "strings"

type QuickMutationMetric struct {
	Id int `json:"-" gorm:"primaryKey;autoIncrement"`

	Name           string `json:"name" gorm:"not null;type:varchar(255);column:name"` // 基础指标名
	NameEN         string `json:"-" gorm:"type:varchar(255);column:name_en"`
	Describe       string `json:"describe" gorm:"not null;type:varchar(255);column:describe"` // 应用Modifier后的指标描述
	DescribeEN     string `json:"-" gorm:"type:varchar(255);column:describe_en"`
	CustomMetric   string `json:"customMetric" gorm:"not null;type:varchar(5000);column:custom_metric"` // 自定义表达式,原始为基于模版和motify生成的数据
	MetricTemplate string `json:"metricTemplate" gorm:"type:varchar(5000);column:metric_template"`      // 快速指标模版,用于配合Modifier修改指标

	// Modifier 用于记录快速指标的更改项
	RangeVector string `json:"rangeVector,omitempty" gorm:"type:varchar(255);column:range_vector"` // 1m/1h/1d
	Scale       string `json:"scale,omitempty" gorm:"type:varchar(255);column:scale"`              // 倍率
	Quantile    string `json:"quantile,omitempty" gorm:"type:varchar(255);column:quantile"`        // 百分位数
	DayOnDay    bool   `json:"dayOnDay" gorm:"column:day_on_day"`                                  // 使用昨日同时段数据
	WeekOnWeek  bool   `json:"weekOnWeek" gorm:"column:week_on_week"`                              // 使用上周同时段数据

	// 指标的额外信息
	Group        string `json:"group" gorm:"not null;type:varchar(255);column:group"` // 用于记录指标所属的告警事件组
	GroupEN      string `json:"-" gorm:"type:varchar(255);column:group_en"`
	CompareGroup string `json:"compareGroup" gorm:"type:varchar(255);column:compare_group"` // 用于记录指标类型, 相同类型的指标才有比较的意义
}

func (QuickMutationMetric) TableName() string {
	return "quick_mutation_metric"
}

// ListQuickMutationMetric 列出所有的快速指标
func (repo *daoRepo) ListQuickMutationMetric(lang string) ([]QuickMutationMetric, error) {
	var quickMutationMetrics []QuickMutationMetric
	err := repo.db.Model(&QuickMutationMetric{}).
		Select(getQuickMutationName(lang), getQuickMutationDescribe(lang), getQuickMutationGroup(lang),
			"custom_metric", "metric_template", "range_vector", "scale", "quantile", "day_on_day", "week_on_week", "compare_group",
		).
		Scan(&quickMutationMetrics).Error
	return quickMutationMetrics, err
}

func getQuickMutationName(lang string) string {
	if strings.HasPrefix(lang, "en") { // en_US,en
		return `name_en AS "name"`
	}
	// if strings.HasPrefix(lang, "zh") { // zh_CN,zh
	// 	return "name"
	// }
	return `"name"`
}

func getQuickMutationDescribe(lang string) string {
	if strings.HasPrefix(lang, "en") { // en_US,en
		return `describe_en AS "describe"`
	}
	// if strings.HasPrefix(lang, "zh") { // zh_CN,zh
	// 	return "name"
	// }
	return `"describe"`
}

func getQuickMutationGroup(lang string) string {
	if strings.HasPrefix(lang, "en") { // en_US,en
		return `group_en AS "group"`
	}
	// if strings.HasPrefix(lang, "zh") { // zh_CN,zh
	// 	return "name"
	// }
	return `"group"`
}
