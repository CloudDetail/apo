// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/CloudDetail/apo/backend/pkg/util"
	"gorm.io/gorm"
)

type StructureDefine interface {
	TableName() string
}

func InitSQL(db *gorm.DB, model StructureDefine) error {
	if err := db.AutoMigrate(&model); err != nil {
		return err
	}

	var count int64
	db.Model(&model).Count(&count)

	if count == 0 {
		initSqlScript := fmt.Sprintf("./sqlscripts/%s.sql", model.TableName())
		if _, err := os.Stat(initSqlScript); err == nil {
			sql, err := os.ReadFile(initSqlScript)
			if err != nil {
				return err
			}

			validSql, err := util.ValidateSQL(string(sql))
			if err != nil {
				return err
			}

			if err := db.Exec(validSql).Error; err != nil {
				return err
			}
		}
		return nil
	}

	updateScripts, err := walkMatch(fmt.Sprintf("./sqlscripts/upgrade/%s/", model.TableName()), "*.sql")
	if err != nil {
		return nil
	}
	sort.Strings(updateScripts)
	for _, script := range updateScripts {
		sql, err := os.ReadFile(script)
		if err != nil {
			continue
		}

		validSql, err := util.ValidateSQL(string(sql))
		if err != nil {
			return err
		}

		if err := db.Exec(validSql).Error; err != nil {
			log.Printf("update [%s]:%s failed: err: %v", model.TableName(), script, err)
			continue
		}
	}
	return nil
}

func walkMatch(root, pattern string) ([]string, error) {
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if matched {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}
