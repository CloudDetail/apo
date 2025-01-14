// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import "github.com/CloudDetail/apo/backend/pkg/model/input/alert"

func (repo *subRepo) ListAlertTargetTags() ([]alert.TargetTag, error) {
	var targetTags []alert.TargetTag
	err := repo.db.Find(&targetTags).Order("id ASC").Error
	return targetTags, err
}
