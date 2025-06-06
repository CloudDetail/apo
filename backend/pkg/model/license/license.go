// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package license

import "gorm.io/gorm"

const ActivateLicenseCheck = true

type Activation struct {
	ID   int    `gorm:"column:id;primary_key;"`
	Mode string `gorm:"column:mode;type:varchar(10)"`
}

func (Activation) TableName() string {
	return "activation"
}

type License struct {
	gorm.Model
	Token             string `gorm:"column:token;type:varchar(700)"`
	SignPublicKey     string `gorm:"column:sign_public_key;type:varchar(600)"`
	EncryptPublicKey  string `gorm:"column:encrypt_public_key;type:varchar(700)"`
	DecryptPrivateKey string `gorm:"column:decrypt_private_key;type:varchar(2000)"`
	Activated         bool   `gorm:"column:activated"`
}

func (License) TableName() string {
	return "license"
}
