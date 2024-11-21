package database

import "os"

// AlertMetricsData 提供用户可选择的指标对应的PQL
type AlertMetricsData struct {
	Id int `json:"-" gorm:"primaryKey;autoIncrement"`

	Name  string `json:"name" gorm:"not null;type:varchar(100);column:name"`
	PQL   string `json:"pql" gorm:"not null;type:varchar(5000);column:pql"`
	Unit  string `json:"unit" gorm:"not null;type:varchar(100);column:unit"`
	Group string `json:"group" gorm:"not null;type:varchar(100);column:group"`
}

func (a *AlertMetricsData) TableName() string {
	return "quick_alert_rule_metric"
}

// ListQuickMutationMetric 列出所有的快速指标
func (repo *daoRepo) ListQuickAlertRuleMetric() ([]AlertMetricsData, error) {
	var quickAlertMetrics []AlertMetricsData
	err := repo.db.Find(&quickAlertMetrics).Error
	return quickAlertMetrics, err
}

func (repo *daoRepo) InitPredefinedQuickAlertRuleMetric(sqlScript string) error {
	if err := repo.db.AutoMigrate(&AlertMetricsData{}); err != nil {
		return err
	}
	var count int64
	if err := repo.db.Model(&AlertMetricsData{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	if sqlScript == "" {
		// 默认的初始化脚本
		sqlScript = "./sqlscripts/default_quick_alert_rule_metric.sql"
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
