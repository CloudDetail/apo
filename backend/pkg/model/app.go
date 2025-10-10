// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package model

type Logo struct {
	Id        int    `gorm:"column:id;primary_key" json:"-"`
	LogoImage []byte `gorm:"column:logo_image;" json:"logoImage"`
	LogoText  string `gorm:"column:logo_text;type:varchar(15)" json:"logoText"`
}

func (Logo) TableName() string {
	return "app_logo"
}
