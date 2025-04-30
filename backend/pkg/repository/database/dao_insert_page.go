// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"errors"

	"gorm.io/gorm"
)

// InsertPage saves embedded interface.
type InsertPage struct {
	PageID int    `gorm:"column:page_id;primary_key;auto_increment" json:"pageId"`
	Url    string `gorm:"column:url;type:varchar(150)" json:"url"`
	Type   string `gorm:"column:type;type:varchar(20)" json:"type"` // For now it's grafana or jaeger.
	Custom bool   `gorm:"column:custom" json:"-"`
}

// RouterInsertPage maps router to inserted page.
type RouterInsertPage struct {
	ID       int    `gorm:"column:id"`
	RouterID int    `gorm:"column:router_id"`
	PageID   int    `gorm:"column:page_id"`
	Language string `gorm:"column:language;type:varchar(20);default:NULL" json:"language"`
}

func (RouterInsertPage) TableName() string {
	return "router_insert_page"
}

func (t *InsertPage) TableName() string {
	return "insert_page"
}

func (repo *daoRepo) GetRouterInsertedPage(routers []*Router, language string) error {
	if len(routers) == 0 {
		return nil
	}

	for i := range routers {
		var page InsertPage
		err := repo.db.Table("router_insert_page").
			Select("insert_page.*").
			Joins("JOIN insert_page ON router_insert_page.page_id = insert_page.page_id").
			Where(`router_insert_page.router_id = ? AND router_insert_page."language" = ? OR (router_insert_page."language" IS NULL AND insert_page.custom = true)`, routers[i].RouterID, language).
			First(&page).Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return err
		}
		if page.PageID != 0 {
			routers[i].Page = &page
		}
	}

	return nil
}
