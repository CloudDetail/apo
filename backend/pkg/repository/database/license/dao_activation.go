// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package license

import "github.com/CloudDetail/apo/backend/pkg/model/license"

func (repo *licenseRepo) CreateActivation(activation *license.Activation) error {
	return repo.db.Create(activation).Error
}

func (repo *licenseRepo) GetActivations() ([]license.Activation, error) {
	var activations []license.Activation
	if err := repo.db.Find(&activations).Error; err != nil {
		return nil, err
	}

	return activations, nil
}
