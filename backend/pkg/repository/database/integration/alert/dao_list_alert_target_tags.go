// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"strings"

	"github.com/CloudDetail/apo/backend/pkg/code"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

func (repo *subRepo) ListAlertTargetTags(ctx core.Context) ([]alert.TargetTag, error) {
	var targetTags []alert.TargetTag
	err := repo.GetContextDB(ctx).Model(&alert.TargetTag{}).
		Select("id", "field", getTargetTag(ctx), getTargetTagDescribe(ctx)).
		Order("id ASC").
		Scan(&targetTags).Error
	return targetTags, err
}

func getTargetTag(ctx core.Context) string {
	var lang = code.LANG_EN
	if ctx != nil {
		lang = ctx.LANG()
	}
	if strings.HasPrefix(lang, "en") { // en_US,en
		return `tag_name_en AS "tag_name"`
	}
	return `"tag_name"`
}

func getTargetTagDescribe(ctx core.Context) string {
	var lang = code.LANG_EN
	if ctx != nil {
		lang = ctx.LANG()
	}
	if strings.HasPrefix(lang, "en") { // en_US,en
		return `describe_en AS "describe"`
	}
	return `"describe"`
}
