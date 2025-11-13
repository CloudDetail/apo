// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package license

import "github.com/CloudDetail/apo/backend/pkg/model/license"

func (repo *licenseRepo) CreateLicense(lic *license.License) error {
	err := repo.db.Model(&license.License{}).Where("activated = ?", true).Update("activated", false).Error
	if err != nil {
		return err
	}
	return repo.db.Create(lic).Error
}

func (repo *licenseRepo) DeactivateLicense(id int) error {
	return repo.db.Model(&license.License{}).Where("id = ?", id).Update("activated", false).Error
}

func (repo *licenseRepo) GetLicense() ([]license.License, error) {
	var licenses []license.License
	if err := repo.db.Find(&licenses).Error; err != nil {
		return nil, err
	}

	return licenses, nil
}

func (repo *licenseRepo) GetCurrentLicense() (license.License, error) {
	var lic license.License
	if err := repo.db.Where("activated = ?", true).Find(&lic).Error; err != nil {
		return license.License{}, nil
	}

	return lic, nil
}
