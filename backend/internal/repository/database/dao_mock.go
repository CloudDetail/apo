// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"github.com/CloudDetail/apo/backend/internal/model/request"
	"github.com/CloudDetail/apo/backend/internal/model/response"
)

// BaseModel 模型基类
type BaseModel struct {
	ID uint `gorm:"AUTO_INCREMENT"`
}

type Mock struct {
	BaseModel
	Name string
}

func (Mock) TableName() string {
	// 必须实现TableName()，不然会变为mocks表
	return "mock"
}

func (repo *daoRepo) CreateMock(model *Mock) (id uint, err error) {
	if err = repo.db.Create(model).Error; err != nil {
		return
	}
	return model.ID, nil
}

func (repo *daoRepo) GetMockById(id uint) (model *Mock, err error) {
	err = repo.db.
		Where("id = ?", id).
		First(&model).Error
	return
}

func (repo *daoRepo) ListMocksByCondition(req *request.ListRequest) (r []*response.ListData, count int64, err error) {
	d := repo.db.Model(&Mock{}).
		Where("name = ?", req.Name)
	d.Count(&count) // 总数

	repo.db.
		Model(&Mock{}).
		Where("name = ?", req.Name).
		Offset((req.PageNum - 1) * req.PageSize).
		Limit(req.PageSize).
		Select("id, name").
		Find(&r)
	return
}

func (repo *daoRepo) UpdateMockById(id uint, m map[string]interface{}) error {
	return repo.db.
		Where("id = ?", id).
		Updates(m).Error
}

func (repo *daoRepo) DeleteMockById(id uint) error {
	return repo.db.
		Where("id = ?", id).
		Delete(&Mock{}).Error
}
