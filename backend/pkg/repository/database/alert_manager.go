package database

import (
	"errors"
	"gorm.io/gorm"
)

type DingTalkConfig struct {
	ConfigFile string `json:"-" gorm:"column:config_file;type:varchar(50)"`
	AlertName  string `json:"-" gorm:"column:alert_name;type:varchar(50)"`
	UUID       string `json:"-" gorm:"column:uuid;unique;type:varchar(50)"`
	URL        string `json:"url,omitempty" gorm:"column:url;type:varchar(150)"`
	Secret     string `json:"secret,omitempty" gorm:"secret"`
}

func (t DingTalkConfig) TableName() string {
	return "ding_talk_config"
}

func (repo *daoRepo) CreateDingTalkReceiver(dingTalkConfig *DingTalkConfig) error {
	return repo.db.Create(dingTalkConfig).Error
}

func (repo *daoRepo) GetDingTalkReceiver(uuid string) (DingTalkConfig, error) {
	config := DingTalkConfig{}
	err := repo.db.Select("url, secret").Where("uuid = ?", uuid).First(&config).Error
	return config, err
}

func (repo *daoRepo) GetDingTalkReceiverByAlertName(configFile string, alertName string, page, pageSize int) ([]*DingTalkConfig, int64, error) {
	var dingTalkConfigs []*DingTalkConfig
	offset := (page - 1) * pageSize

	query := repo.db.Select("url, secret").Where("config_file = ?", configFile)
	countQuery := repo.db.Model(&DingTalkConfig{}).Select("*").Where("config_file = ?", configFile)

	if len(alertName) > 0 {
		query = query.Where("alert_name = ?", alertName)
		countQuery = countQuery.Where("alert_name = ?", alertName)
	}

	err := query.Offset(offset).Limit(pageSize).Find(&dingTalkConfigs).Error
	var count int64
	countQuery.Count(&count)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, count, err
	}

	return dingTalkConfigs, count, nil
}

func (repo *daoRepo) UpdateDingTalkReceiver(dingTalkConfig *DingTalkConfig, oldName string) error {
	return repo.db.Where("config_file = ? AND alert_name = ?", dingTalkConfig.ConfigFile, oldName).Updates(dingTalkConfig).Error
}

func (repo *daoRepo) DeleteDingTalkReceiver(configFile, alertName string) error {
	return repo.db.Where("config_file = ? AND alert_name = ?", configFile, alertName).Delete(&DingTalkConfig{}).Error
}
