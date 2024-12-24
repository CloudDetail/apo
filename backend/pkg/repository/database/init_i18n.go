package database

import (
	"fmt"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

func (repo *daoRepo) initI18nTranslation() error {
	type transConfig struct {
		Key  string            `mapstructure:"key"`
		I18n []I18nTranslation `mapstructure:"i18n"`
	}
	var translationConfig map[string]transConfig
	viper.SetConfigType("yaml")
	viper.SetConfigFile("./sqlscripts/i18n.yml")
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	if err := viper.Unmarshal(&translationConfig); err != nil {
		return err
	}

	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&I18nTranslation{}); err != nil {
			return err
		}

		var existingTranslations, toInsert, toDelete []I18nTranslation
		if err := tx.Find(&existingTranslations).Error; err != nil {
			return err
		}

		existingMap := make(map[string]I18nTranslation)
		for _, record := range existingTranslations {
			key := fmt.Sprintf("%s:%s:%s", record.EntityType, record.Language, record.Translation)
			existingMap[key] = record
		}

		for _, translations := range translationConfig {
			var targetID int
			typ := translations.I18n[0].EntityType
			targetName := translations.Key

			if typ == model.TRANSLATION_TYP_FEATURE {
				if err := tx.Model(&Feature{}).Select("feature_id").Where("feature_name = ?", targetName).Find(&targetID).Error; err != nil {
					return err
				}
			} else if typ == model.TRANSLATION_TYP_MENU {
				if err := tx.Model(&MenuItem{}).Select("item_id").Where("key = ?", targetName).Find(&targetID).Error; err != nil {
					return err
				}
			}

			for i := range translations.I18n {
				translations.I18n[i].EntityID = targetID
			}

			for _, newTranslation := range translations.I18n {
				key := fmt.Sprintf("%s:%s:%s", newTranslation.EntityType, newTranslation.Language, newTranslation.Translation)
				if _, exists := existingMap[key]; !exists {
					toInsert = append(toInsert, newTranslation)
				} else {
					delete(existingMap, key)
				}
			}
		}

		for _, existingTranslation := range existingMap {
			toDelete = append(toDelete, existingTranslation)
		}

		if len(toInsert) > 0 {
			if err := tx.Create(&toInsert).Error; err != nil {
				return err
			}
		}

		if len(toDelete) > 0 {
			var ids []int
			for _, record := range toDelete {
				ids = append(ids, record.ID)
			}
			if err := tx.Where("id IN ?", ids).Delete(&I18nTranslation{}).Error; err != nil {
				return err
			}
		}

		return nil
	})

}
