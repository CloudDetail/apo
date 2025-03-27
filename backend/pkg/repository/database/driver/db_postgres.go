// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"fmt"

	"github.com/CloudDetail/apo/backend/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgresDialector() (gorm.Dialector, error) {
	// Build DSN information
	postgresCfg := config.Get().Database.Postgres

	if len(postgresCfg.SSLMode) == 0 {
		postgresCfg.SSLMode = "disable"
	}
	if len(postgresCfg.Timezone) == 0 {
		postgresCfg.Timezone = "Asia/Shanghai"
	}

	sysDSN, dbDSN := postgresDSNs(
		postgresCfg.Host,
		postgresCfg.UserName,
		postgresCfg.Password,
		postgresCfg.Database,
		postgresCfg.Port,
		postgresCfg.SSLMode,
		postgresCfg.Timezone)

	err := ensureDBExist("pgx", sysDSN, dbDSN, postgresCfg.Database)
	return postgres.New(postgres.Config{
		DSN:                  dbDSN,
		PreferSimpleProtocol: true,
	}), err
}

func postgresDSNs(host, username, password, database string, port int, sslmode, timezone string) (sysDSN string, dbDSN string) {
	sysDSN = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		host,
		username,
		password,
		"postgres",
		port,
		sslmode,
		timezone,
	)

	dbDSN = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		host,
		username,
		password,
		database,
		port,
		sslmode,
		timezone,
	)
	return sysDSN, dbDSN
}
