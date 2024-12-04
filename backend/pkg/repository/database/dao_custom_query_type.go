package database

import (
	"errors"

	"gorm.io/gorm"
)

type QueryServiceType struct {
	ID         uint   `gorm:"primaryKey;autoIncrement"`
	UserName   string `gorm:"type:varchar(100);column:username"`
	CustomName string `gorm:"type:varchar(100);column:customname"`
	GroupName  string `gorm:"type:varchar(100);column:groupname"`
	QueryType  string `gorm:"type:varchar(100);column:querytype"`
}

func (QueryServiceType) TableName() string {
	return "custom_query_service"
}

func (repo *daoRepo) CreateCustomQuery(serviceType *QueryServiceType) error {
	var count int64
	repo.db.Model(&QueryServiceType{}).Where("username = ? and customname = ? and groupname = ?", serviceType.UserName, serviceType.CustomName, serviceType.GroupName).Count(&count)
	if count > 0 {
		return errors.New("已经存在自定义查询名称")
	}
	return repo.db.Create(serviceType).Error
}

func (repo *daoRepo) GetCustomQuery(userName, groupName string) ([]QueryServiceType, error) {
	var serviceType []QueryServiceType
	err := repo.db.Where("username = ? and groupname = ?", userName, groupName).Find(&serviceType).Error
	return serviceType, err
}

func (repo *daoRepo) UpdateCustomQuery(userName, groupName, customName string, customType string) error {
	var serviceType QueryServiceType
	err := repo.db.Where("username = ? and groupname = ? and customname = ?", userName, groupName, customName).First(&serviceType).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("不存在该自定义查询")
	} else if err != nil {
		return err
	}
	// serviceType.QueryType = customType
	return repo.db.Model(&QueryServiceType{}).Where("username = ? and groupname = ? and customname = ?", userName, groupName, customName).Update("querytype", customType).Error
}

func (repo *daoRepo) DeleteCustomQuery(userName, groupName, customName string) error {
	var serviceType QueryServiceType
	err := repo.db.Where("username = ? and groupname = ? and customname = ?", userName, groupName, customName).First(&serviceType).Error
	if err != nil {
		return err
	}
	return repo.db.Where("username = ? and groupname = ? and customname = ?", userName, groupName, customName).Delete(&QueryServiceType{}).Error
}
