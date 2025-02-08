// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import "github.com/CloudDetail/apo/backend/pkg/model"

type I18nTranslation struct {
	ID          int    `gorm:"column:id;primary_key;auto_increment" json:"-"`
	Language    string `gorm:"column:language;type:varchar(20)" json:"-" mapstructure:"language"` // en, zh, etc.
	Translation string `gorm:"column:translation;type:varchar(50)" json:"translation" mapstructure:"translation"`
	FieldName   string `gorm:"column:field_name;type:varchar(20)" json:"field_name" mapstructure:"field_name"` // which field is translated
	EntityID    int    `gorm:"column:entity_id" json:"-"`
	EntityType  string `gorm:"column:entity_type;type:varchar(20)" json:"-" mapstructure:"entity_type"` // menu_item or feature
}

func (I18nTranslation) TableName() string {
	return "i18n_translation"
}

func (repo *daoRepo) GetTranslation(targetIDs []int, targetType string, language string) ([]I18nTranslation, error) {
	var translations []I18nTranslation
	err := repo.db.Where("entity_id in ? AND entity_type = ? AND language = ?", targetIDs, targetType, language).Find(&translations).Error
	return translations, err
}

func (repo *daoRepo) GetFeatureTans(features *[]Feature, language string) error {
	featureIDs := make([]int, len(*features))
	for i, f := range *features {
		featureIDs[i] = f.FeatureID
	}
	translations, err := repo.GetTranslation(featureIDs, model.TRANSLATION_TYP_FEATURE, language)
	if err != nil {
		return err
	}

	for i := range *features {
		feature := &(*features)[i]

		for _, t := range translations {
			if t.EntityID == feature.FeatureID {
				feature.FeatureName = t.Translation
			}
		}
	}
	return nil
}

func (repo *daoRepo) GetMenuItemTans(menuItems *[]MenuItem, language string) error {
	featureIDs := make([]int, len(*menuItems))
	for i, f := range *menuItems {
		featureIDs[i] = f.ItemID
	}
	translations, err := repo.GetTranslation(featureIDs, model.TRANSLATION_TYP_MENU, language)
	if err != nil {
		return err
	}

	for i := range *menuItems {
		menuItem := &(*menuItems)[i]

		for _, t := range translations {
			if t.EntityID == menuItem.ItemID {
				if t.FieldName == "label" {
					menuItem.Label = t.Translation
				} else if t.FieldName == "abbreviation" {
					menuItem.Abbreviation = t.Translation
				}
			}
		}
	}
	return nil
}
