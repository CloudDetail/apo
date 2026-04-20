// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/util"
	"golang.org/x/mod/semver"
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

	updateScripts, err := walkMatch(fmt.Sprintf("./sqlscripts/upgrade/%s/", model.TableName()))
	if err != nil {
		return nil
	}
	updateScripts = sortUpgradeFiles(updateScripts)
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

func walkMatch(root string) ([]string, error) {
	globalCfg := config.Get()
	connType := globalCfg.Database.Connection

	connSql := fmt.Sprintf("*-%s.sql", connType)

	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if matched, err := filepath.Match("*-upgrade.sql", filepath.Base(path)); err == nil && matched {
			matches = append(matches, path)
		} else if matched, err := filepath.Match(connSql, filepath.Base(path)); err == nil && matched {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func sortUpgradeFiles(files []string) []string {
	re := regexp.MustCompile(`\d+\.\d+\.\d+`)

	sort.SliceStable(files, func(i, j int) bool {
		v1 := re.FindString(files[i])
		v2 := re.FindString(files[j])

		// 语义版本比较
		if semver.IsValid("v"+v1) && semver.IsValid("v"+v2) {
			if semver.Compare("v"+v1, "v"+v2) != 0 {
				return semver.Compare("v"+v1, "v"+v2) < 0
			}

			// 同版本号，优先级：-upgrade.sql < -upgrade-xxx.sql
			p1 := priority(files[i])
			p2 := priority(files[j])
			if p1 != p2 {
				return p1 < p2
			}

			// 同优先级，按字典序
			return strings.Compare(files[i], files[j]) < 0
		}

		// 非语义版本，按字典序
		return strings.Compare(files[i], files[j]) < 0
	})

	return files
}

// 优先级函数
func priority(name string) int {
	if strings.HasSuffix(name, "-upgrade.sql") {
		return 1 // 放前
	}
	if strings.Contains(name, "-upgrade-") && strings.HasSuffix(name, ".sql") {
		return 2 // 放后
	}
	return 3 // 其他文件
}
