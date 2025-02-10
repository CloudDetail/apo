package database

import "os"

type QuickMutationMetric struct {
	Id int `json:"-" gorm:"primaryKey;autoIncrement"`

	Name           string `json:"name" gorm:"not null;type:varchar(100);column:name"`                   // 基础指标名
	Describe       string `json:"describe" gorm:"not null;type:varchar(100);column:describe"`           // 应用Modifier后的指标描述
	CustomMetric   string `json:"customMetric" gorm:"not null;type:varchar(5000);column:custom_metric"` // 自定义表达式,原始为基于模版和motify生成的数据
	MetricTemplate string `json:"metricTemplate" gorm:"type:varchar(5000);column:metric_template"`      // 快速指标模版,用于配合Modifier修改指标

	// Modifier 用于记录快速指标的更改项
	RangeVector string `json:"rangeVector,omitempty" gorm:"type:varchar(100);column:range_vector"` // 1m/1h/1d
	Scale       string `json:"scale,omitempty" gorm:"type:varchar(100);column:scale"`              // 倍率
	Quantile    string `json:"quantile,omitempty" gorm:"type:varchar(100);column:quantile"`        // 百分位数
	DayOnDay    bool   `json:"dayOnDay" gorm:"type:tinyint(1);column:day_on_day"`                  // 使用昨日同时段数据
	WeekOnWeek  bool   `json:"weekOnWeek" gorm:"type:tinyint(1);column:week_on_week"`              // 使用上周同时段数据

	// 指标的额外信息
	Group        string `json:"group" gorm:"not null;type:varchar(100);column:group"`       // 用于记录指标所属的告警事件组
	CompareGroup string `json:"compareGroup" gorm:"type:varchar(100);column:compare_group"` // 用于记录指标类型, 相同类型的指标才有比较的意义
}

func (QuickMutationMetric) TableName() string {
	return "quick_mutation_metric"
}

// ListQuickMutationMetric 列出所有的快速指标
func (repo *daoRepo) ListQuickMutationMetric() ([]QuickMutationMetric, error) {
	var quickMutationMetrics []QuickMutationMetric
	err := repo.db.Find(&quickMutationMetrics).Error
	return quickMutationMetrics, err
}

// InitPredefinedQuickMutationMetric 初始化并加载预定义的快速指标
func (repo *daoRepo) InitPredefinedQuickMutationMetric(sqlScript string) error {
	if err := repo.db.AutoMigrate(&QuickMutationMetric{}); err != nil {
		return err
	}
	var count int64
	if err := repo.db.Model(&QuickMutationMetric{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	if sqlScript == "" {
		// 默认的初始化脚本
		sqlScript = "./sqlscripts/default_quick_mutation_metric.sql"
	}

	// 检查初始化脚本是否存在
	if _, err := os.Stat(sqlScript); err == nil {
		// 读取文件并执行初始化脚本
		sql, err := os.ReadFile(sqlScript)
		if err != nil {
			return err
		}
		if err := repo.db.Exec(string(sql)).Error; err != nil {
			return err
		}
	}
	return nil
}
