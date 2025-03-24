// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

type TargetTag struct {
	ID         uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	TagName    string `gorm:"type:varchar(255);column:tag_name" json:"tagName"`
	TagNameEN  string `gorm:"type:varchar(255);column:tag_name_en" json:"-"`
	Describe   string `gorm:"type:varchar(255);column:describe" json:"describe"`
	DescribeEN string `gorm:"type:varchar(255);column:describe_en" json:"-"`
	Field      string `gorm:"type:varchar(255);column:field" json:"targetTag"`
}

func (TargetTag) TableName() string {
	return "target_tags"
}
