// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"fmt"

	"github.com/CloudDetail/apo/backend/config"
	"gorm.io/gorm"

	"gorm.io/driver/mysql"
)

func NewMySqlDialector() gorm.Dialector {
	// Build DSN information
	mysqlCfg := config.Get().Database.MySql
	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=%v&parseTime=True&multiStatements=true&loc=Local",
		mysqlCfg.UserName,
		mysqlCfg.Password,
		mysqlCfg.Host,
		mysqlCfg.Port,
		mysqlCfg.Database,
		mysqlCfg.Charset,
	)
	return mysql.New(mysql.Config{
		DSN: dsn,
	})
}
