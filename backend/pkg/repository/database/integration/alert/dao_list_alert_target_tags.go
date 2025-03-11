// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"strings"

	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

func (repo *subRepo) ListAlertTargetTags(lang string) ([]alert.TargetTag, error) {
	var targetTags []alert.TargetTag
	err := repo.db.Model(&alert.TargetTag{}).
		Select("id", "field", getTargetTag(lang), getTargetTagDescribe(lang)).
		Order("id ASC").
		Scan(&targetTags).Error
	return targetTags, err
}

func getTargetTag(lang string) string {
	if strings.HasPrefix(lang, "en") { // en_US,en
		return "tag_name_en AS tag_name"
	}
	// if strings.HasPrefix(lang, "zh") { // zh_CN,zh
	// 	return "name"
	// }
	return "tag_name"
}

func getTargetTagDescribe(lang string) string {
	if strings.HasPrefix(lang, "en") { // en_US,en
		return "describe_en AS describe"
	}
	// if strings.HasPrefix(lang, "zh") { // zh_CN,zh
	// 	return "name"
	// }
	return "describe"
}
