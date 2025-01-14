// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

type TargetTag struct {
	ID       uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	TagName  string `gorm:"type:varchar(100);column:tag_name" json:"tagName"`
	Describe string `gorm:"type:varchar(255);column:describe" json:"describe"`
	Field    string `gorm:"type:varchar(100);column:field" json:"targetTag"`
}
