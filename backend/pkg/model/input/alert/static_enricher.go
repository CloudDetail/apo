// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import "github.com/google/uuid"

type StaticEnrichConfig struct {
	SourceID   uuid.UUID
	EnricherID uint

	ConditionGroupID string

	FromTag      string
	CompareType  string // equal;exist;not_exist // TODO regex
	CompareValue string

	TargetTagID        uint
	TargetTagValue     string
	TargetTagValueType string // constant; TODO regex

	CustomField string `gorm:"type:varchar(100);column:custom_field"`
}

type StaticEnrichConfigVO struct {
	StaticEnrichConfig

	TagName string `gorm:"type:varchar(100);column:tag_name"`
}
