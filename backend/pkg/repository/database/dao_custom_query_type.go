package database

import (
	"errors"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"gorm.io/gorm"
)

type QueryServiceType struct {
	ID uint `gorm:"primaryKey;autoIncrement"`
	//UserName   string `gorm:"type:varchar(255);column:username"`
	CustomName string `gorm:"type:varchar(255);column:customname"`
	//GroupName string `gorm:"type:varchar(255);column:groupname"`
	QueryType string `gorm:"type:varchar(255);column:querytype"`
}

func (QueryServiceType) TableName() string {
	return "custom_query_service"
}

func (repo *daoRepo) CreateCustomQuery(ctx core.Context, serviceType *QueryServiceType) error {
	var count int64
	repo.GetContextDB(ctx).Model(&QueryServiceType{}).Where("customname = ?", serviceType.CustomName).Count(&count)
	if count > 0 {
		return errors.New("已经存在自定义查询名称")
	}
	return repo.GetContextDB(ctx).Create(serviceType).Error
}

func (repo *daoRepo) GetCustomQuery(ctx core.Context) ([]QueryServiceType, error) {
	var serviceType []QueryServiceType
	err := repo.GetContextDB(ctx).Find(&serviceType).Error
	return serviceType, err
}

func (repo *daoRepo) UpdateCustomQuery(ctx core.Context, customName string, customType string) error {
	var serviceType QueryServiceType
	err := repo.GetContextDB(ctx).Where("customname = ?", customName).First(&serviceType).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("不存在该自定义查询")
	} else if err != nil {
		return err
	}
	// serviceType.QueryType = customType
	return repo.GetContextDB(ctx).Model(&QueryServiceType{}).Where("customname = ?", customName).Update("querytype", customType).Error
}

func (repo *daoRepo) DeleteCustomQuery(ctx core.Context, customName string) error {
	var serviceType QueryServiceType
	err := repo.GetContextDB(ctx).Where("customname = ?", customName).First(&serviceType).Error
	if err != nil {
		return err
	}
	return repo.GetContextDB(ctx).Where("customname = ?", customName).Delete(&QueryServiceType{}).Error
}
