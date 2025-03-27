// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	_ "gorm.io/driver/mysql"
	_ "gorm.io/driver/postgres"
)

func ensureDBExist(driverName, sysDSN, dbDSN, database string) error {
	if err := checkDBConnection(driverName, dbDSN); err == nil {
		return nil
	}

	sanitizedDB, err := sanitizeDatabaseName(database)
	if err != nil {
		return err
	}
	return createDatabase(driverName, sysDSN, sanitizedDB)
}

func checkDBConnection(driver, dsn string) error {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	return db.Ping()
}

func createDatabase(driver, sysDSN, database string) error {
	sysDB, err := sql.Open(driver, sysDSN)
	if err != nil {
		return fmt.Errorf("failed to connect to system DB: %v", err)
	}
	defer sysDB.Close()

	if _, err = sysDB.Exec(fmt.Sprintf(`CREATE DATABASE "%s"`, database)); err != nil {
		return fmt.Errorf("failed to create database: %v", err)
	}
	return nil
}

func sanitizeDatabaseName(input string) (string, error) {
	input = strings.TrimSpace(input)

	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]{1,64}$`, input); !matched {
		return "", fmt.Errorf("the database name is invalid, it can only contain letters, numbers, and underscores, and the length is 1-64")
	}

	return input, nil
}
