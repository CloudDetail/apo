// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"errors"
	"gorm.io/gorm"
)

// 用户配置阈值表
const GLOBAL = "global"
const LATENCY = 5.0
const ERROR_RATE = 5.0
const LOG = 5.0
const TPS = 5.0

type Threshold struct {
	ID          uint    `gorm:"primaryKey;autoIncrement"`
	ServiceName string  `gorm:"type:varchar(100)"`
	Level       string  `gorm:"type:varchar(100)"`
	EndPoint    string  `gorm:"type:varchar(100)"`
	Latency     float64 `gorm:"type:decimal(10,2)"`
	Tps         float64 `gorm:"type:decimal(10,2)"`
	ErrorRate   float64 `gorm:"type:decimal(10,2)"`
	Log         float64 `gorm:"type:decimal(10,2)"`
}

func (Threshold) TableName() string {
	// 必须实现TableName()，不然会变为mocks表
	return "threshold"
}

func (repo *daoRepo) CreateOrUpdateThreshold(model *Threshold) error {
	var existing Threshold
	result := repo.db.Where("service_name = ? AND end_point = ? AND Level = ?", model.ServiceName, model.EndPoint, model.Level).First(&existing)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// 如果没有找到记录，则插入新记录
		return repo.db.Create(model).Error
	} else if result.Error != nil {
		// 如果出现其他错误，返回错误
		return result.Error
	} else {
		// 如果找到记录，则更新记录
		return repo.db.Model(&existing).Updates(model).Error
	}
}

func (repo *daoRepo) GetOrCreateThreshold(serviceName string, endPoint string, level string) (Threshold, error) {
	var threshold Threshold
	// 根据提供的 serviceName 和 endPoint 查询
	result := repo.db.Where("service_name = ? AND end_point = ? AND Level = ?", serviceName, endPoint, level).First(&threshold)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// 如果没有找到记录，再次根据 serviceName = "global" 和 endPoint = "global" 查询
		result = repo.db.Where("Level = ?", GLOBAL).First(&threshold)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// 如果仍然没有找到记录，创建一个新的阈值数据
			newThreshold := Threshold{
				Level:     GLOBAL,
				Latency:   LATENCY,
				Tps:       TPS,
				ErrorRate: ERROR_RATE,
				Log:       LOG,
			}
			if createErr := repo.db.Create(&newThreshold).Error; createErr != nil {
				return newThreshold, createErr
			}
			return newThreshold, nil
		} else if result.Error != nil {
			return threshold, result.Error
		}
	}
	return threshold, nil
}
func (repo *daoRepo) DeleteThreshold(serviceName string, endPoint string) error {
	// 根据 serviceName 和 endPoint 删除记录
	result := repo.db.Where("service_name = ? AND end_point = ?", serviceName, endPoint).Delete(&Threshold{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no record found to delete")
	}
	return nil
}
