// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"fmt"

	"github.com/CloudDetail/apo/backend/config"
	"gorm.io/gorm"

	"gorm.io/driver/mysql"
)

func NewMySqlDialector() (gorm.Dialector, error) {
	// Build DSN information
	mysqlCfg := config.Get().Database.MySql

	sysDSN, dsn := mysqlDSNs(
		mysqlCfg.UserName,
		mysqlCfg.Password,
		mysqlCfg.Host,
		mysqlCfg.Port,
		mysqlCfg.Database,
		mysqlCfg.Charset)

	err := ensureDBExist("mysql", sysDSN, dsn, mysqlCfg.Database)
	return mysql.New(mysql.Config{
		DSN: dsn,
	}), err
}

func mysqlDSNs(
	username, password, host string, port int, database string, charset string,
) (sysDSN string, dbDSN string) {
	// set sql_mode=ANSI_QUOTES to maintain the universality of SQL across different database
	// using '"' as the delimiter for identifiers.
	sysDSN = fmt.Sprintf("%v:%v@tcp(%v:%v)/?charset=%v&parseTime=True&multiStatements=true&loc=Local&&sql_mode=ANSI_QUOTES",
		username, password, host, port, charset)
	dbDSN = fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=%v&parseTime=True&multiStatements=true&loc=Local&&sql_mode=ANSI_QUOTES",
		username, password, host, port, database, charset)
	return sysDSN, dbDSN
}
