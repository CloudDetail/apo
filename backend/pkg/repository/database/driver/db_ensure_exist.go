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

func ensureDBExist(driverName string, sysDSN string, dbDSN string, database string) error {
	db, err := sql.Open(driverName, dbDSN)
	if err == nil {
		defer db.Close()
		if err = db.Ping(); err == nil {
			return nil
		}
	}

	database, err = sanitizeDatabaseName(database)
	if err != nil {
		return err
	}

	db, err = sql.Open(driverName, sysDSN)
	if err != nil {
		return fmt.Errorf("database not exist, and connect to system failed, err:%v", err)
	}

	_, err = db.Exec(fmt.Sprintf(`CREATE DATABASE "%s"`, database))
	if err != nil {
		return fmt.Errorf("create database failed, err:%v", err)
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
