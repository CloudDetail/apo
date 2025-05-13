// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"errors"

	"gorm.io/gorm"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

// User-configured threshold table
const GLOBAL = "global"
const LATENCY = 5.0
const ERROR_RATE = 5.0
const LOG = 5.0
const TPS = 5.0

type Threshold struct {
	ID		uint	`gorm:"primaryKey;autoIncrement"`
	ServiceName	string	`gorm:"type:varchar(255)"`
	Level		string	`gorm:"type:varchar(255)"`
	EndPoint	string	`gorm:"type:varchar(255)"`
	Latency		float64	`gorm:"type:decimal(10,2)"`
	Tps		float64	`gorm:"type:decimal(10,2)"`
	ErrorRate	float64	`gorm:"type:decimal(10,2)"`
	Log		float64	`gorm:"type:decimal(10,2)"`
}

func (Threshold) TableName() string {
	// TableName() must be implemented, otherwise it will become a mocks table.
	return "threshold"
}

func (repo *daoRepo) CreateOrUpdateThreshold(ctx_core core.Context, model *Threshold) error {
	var existing Threshold
	result := repo.db.Where("service_name = ? AND end_point = ? AND Level = ?", model.ServiceName, model.EndPoint, model.Level).First(&existing)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// If no record is found, insert a new record
		return repo.db.Create(model).Error
	} else if result.Error != nil {
		// return error if other errors occur
		return result.Error
	} else {
		// Update record if found
		return repo.db.Model(&existing).Updates(model).Error
	}
}

func (repo *daoRepo) GetOrCreateThreshold(ctx_core core.Context, serviceName string, endPoint string, level string) (Threshold, error) {
	var threshold Threshold
	// Query based on serviceName and endPoint provided
	result := repo.db.Where("service_name = ? AND end_point = ? AND Level = ?", serviceName, endPoint, level).First(&threshold)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// If no record is found, query again based on serviceName = "global" and endPoint = "global"
		result = repo.db.Where("Level = ?", GLOBAL).First(&threshold)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// If the record is still not found, create a new threshold data
			newThreshold := Threshold{
				Level:		GLOBAL,
				Latency:	LATENCY,
				Tps:		TPS,
				ErrorRate:	ERROR_RATE,
				Log:		LOG,
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
func (repo *daoRepo) DeleteThreshold(ctx_core core.Context, serviceName string, endPoint string) error {
	// Delete records based on serviceName and endPoint
	result := repo.db.Where("service_name = ? AND end_point = ?", serviceName, endPoint).Delete(&Threshold{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no record found to delete")
	}
	return nil
}
