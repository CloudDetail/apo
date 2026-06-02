// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"os"
	"testing"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/profile"
	"github.com/CloudDetail/apo/backend/pkg/repository/database/driver"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type menuI18nConfig struct {
	Key  string `yaml:"key"`
	I18n []struct {
		Language    string `yaml:"language"`
		Translation string `yaml:"translation"`
		FieldName   string `yaml:"field_name"`
		EntityType  string `yaml:"entity_type"`
	} `yaml:"i18n"`
}

func TestI18nConfigIncludesCustomDashboardMenuTranslations(t *testing.T) {
	raw, err := os.ReadFile("../../../sqlscripts/i18n.yml")
	if err != nil {
		t.Fatalf("read i18n config: %v", err)
	}

	var config map[string]menuI18nConfig
	if err := yaml.Unmarshal(raw, &config); err != nil {
		t.Fatalf("parse i18n config: %v", err)
	}

	item, ok := config["customDashboard"]
	if !ok {
		t.Fatalf("customDashboard menu translations are missing from i18n.yml")
	}
	if item.Key != "customDashboard" {
		t.Fatalf("customDashboard i18n key mismatch: %q", item.Key)
	}

	required := map[string]string{
		"en:label":        "My Dashboards",
		"zh:label":        "我的大盘",
		"en:abbreviation": "Dashboards",
		"zh:abbreviation": "我的大盘",
	}
	for _, translation := range item.I18n {
		if translation.EntityType != model.TRANSLATION_TYP_MENU {
			continue
		}
		key := translation.Language + ":" + translation.FieldName
		if want, exists := required[key]; exists && translation.Translation == want {
			delete(required, key)
		}
	}
	if len(required) > 0 {
		t.Fatalf("customDashboard menu translations are incomplete: missing %#v", required)
	}
}

func TestI18nConfigIncludesCustomDashboardFeatureTranslations(t *testing.T) {
	raw, err := os.ReadFile("../../../sqlscripts/i18n.yml")
	if err != nil {
		t.Fatalf("read i18n config: %v", err)
	}

	var config map[string]menuI18nConfig
	if err := yaml.Unmarshal(raw, &config); err != nil {
		t.Fatalf("parse i18n config: %v", err)
	}

	requiredItems := map[string]map[string]string{
		"我的大盘": {
			"en:featureName": "My Dashboards",
			"zh:featureName": "我的大盘",
		},
	}

	for itemKey, required := range requiredItems {
		item, ok := config[itemKey]
		if !ok {
			t.Fatalf("%s feature translations are missing from i18n.yml", itemKey)
		}
		if item.Key != itemKey {
			t.Fatalf("%s i18n key mismatch: %q", itemKey, item.Key)
		}
		for _, translation := range item.I18n {
			if translation.EntityType != model.TRANSLATION_TYP_FEATURE {
				continue
			}
			key := translation.Language + ":" + translation.FieldName
			if want, exists := required[key]; exists && translation.Translation == want {
				delete(required, key)
			}
		}
		if len(required) > 0 {
			t.Fatalf("%s feature translations are incomplete: missing %#v", itemKey, required)
		}
	}
}

func TestI18nConfigDoesNotExposeCustomDashboardCreateFeature(t *testing.T) {
	raw, err := os.ReadFile("../../../sqlscripts/i18n.yml")
	if err != nil {
		t.Fatalf("read i18n config: %v", err)
	}

	var config map[string]menuI18nConfig
	if err := yaml.Unmarshal(raw, &config); err != nil {
		t.Fatalf("parse i18n config: %v", err)
	}

	if _, ok := config["创建自定义大盘"]; ok {
		t.Fatalf("custom dashboard creation should not have a standalone feature translation")
	}
}

func TestInitI18nTranslationPreservesDynamicCustomDashboardTranslations(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:init-i18n-dynamic-dashboard?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&I18nTranslation{}, &MenuItem{}, &profile.Feature{}); err != nil {
		t.Fatalf("migrate i18n tables: %v", err)
	}
	repo := &daoRepo{DB: &driver.DB{DB: db}}

	const dashboardMenuID = 900
	customTranslation := I18nTranslation{
		Language:    "zh",
		Translation: "用户新建大盘",
		FieldName:   "label",
		EntityID:    dashboardMenuID,
		EntityType:  model.TRANSLATION_TYP_MENU,
	}
	if err := db.Create(&customTranslation).Error; err != nil {
		t.Fatalf("seed custom dashboard translation: %v", err)
	}

	t.Chdir("../../..")
	if err := repo.initI18nTranslation(nil); err != nil {
		t.Fatalf("init i18n translations: %v", err)
	}

	var count int64
	if err := db.Model(&I18nTranslation{}).
		Where("entity_id = ? AND entity_type = ? AND language = ? AND field_name = ? AND translation = ?",
			dashboardMenuID,
			model.TRANSLATION_TYP_MENU,
			"zh",
			"label",
			"用户新建大盘",
		).
		Count(&count).Error; err != nil {
		t.Fatalf("query custom dashboard translation: %v", err)
	}
	if count != 1 {
		t.Fatalf("dynamic custom dashboard translation should survive i18n init, got %d records", count)
	}
}
