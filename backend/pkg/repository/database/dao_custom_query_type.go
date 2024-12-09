package database

import (
	"errors"

	"gorm.io/gorm"
)

type QueryServiceType struct {
	ID uint `gorm:"primaryKey;autoIncrement"`
	//UserName   string `gorm:"type:varchar(100);column:username"`
	CustomName string `gorm:"type:varchar(100);column:customname"`
	//GroupName string `gorm:"type:varchar(100);column:groupname"`
	QueryType string `gorm:"type:varchar(100);column:querytype"`
}

func (QueryServiceType) TableName() string {
	return "custom_query_service"
}

func (repo *daoRepo) CreateCustomQuery(serviceType *QueryServiceType) error {
	var count int64
	repo.db.Model(&QueryServiceType{}).Where("customname = ?", serviceType.CustomName).Count(&count)
	if count > 0 {
		return errors.New("已经存在自定义查询名称")
	}
	return repo.db.Create(serviceType).Error
}

func (repo *daoRepo) GetCustomQuery() ([]QueryServiceType, error) {
	var serviceType []QueryServiceType
	err := repo.db.Find(&serviceType).Error
	return serviceType, err
}

func (repo *daoRepo) UpdateCustomQuery(customName string, customType string) error {
	var serviceType QueryServiceType
	err := repo.db.Where("customname = ?", customName).First(&serviceType).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("不存在该自定义查询")
	} else if err != nil {
		return err
	}
	// serviceType.QueryType = customType
	return repo.db.Model(&QueryServiceType{}).Where("customname = ?", customName).Update("querytype", customType).Error
}

func (repo *daoRepo) DeleteCustomQuery(customName string) error {
	var serviceType QueryServiceType
	err := repo.db.Where("customname = ?", customName).First(&serviceType).Error
	if err != nil {
		return err
	}
	return repo.db.Where("customname = ?", customName).Delete(&QueryServiceType{}).Error
}
