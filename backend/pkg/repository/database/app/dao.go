// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"gorm.io/gorm"
)

type AppRepo interface {
	CreateLogo(logo *model.Logo) error
	GetLogo() (*model.Logo, error)
}

type appRepo struct {
	db *gorm.DB
}

func NewAppRepo(db *gorm.DB) (AppRepo, error) {
	err := db.AutoMigrate(&model.Logo{})
	if err != nil {
		return nil, err
	}

	return &appRepo{
		db: db,
	}, nil
}
