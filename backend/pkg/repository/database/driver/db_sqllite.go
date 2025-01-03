// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"github.com/CloudDetail/apo/backend/config"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewSqlliteDialector() gorm.Dialector {
	return sqlite.Open(config.Get().Database.Sqllite.Database)
}
