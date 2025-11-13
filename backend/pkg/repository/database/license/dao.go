// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package license

import (
	"github.com/CloudDetail/apo/backend/pkg/model/license"
	"gorm.io/gorm"
)

type LicenseRepo interface {
	CreateLicense(lic *license.License) error
	CreateActivation(activation *license.Activation) error
	GetActivations() ([]license.Activation, error)
	GetLicense() ([]license.License, error)
	GetCurrentLicense() (license.License, error)
}

type licenseRepo struct {
	db *gorm.DB
}

func NewLicenseRepo(db *gorm.DB) (LicenseRepo, error) {
	err := db.AutoMigrate(&license.License{}, &license.Activation{})
	if err != nil {
		return nil, err
	}

	return &licenseRepo{
		db: db,
	}, nil
}
