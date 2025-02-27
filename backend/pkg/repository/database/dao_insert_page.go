// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"gorm.io/gorm/clause"
)

// InsertPage saves embedded interface.
type InsertPage struct {
	PageID   int    `gorm:"column:page_id;primary_key;auto_increment" json:"pageId"`
	Url      string `gorm:"column:url;type:varchar(150)" json:"url"`
	Type     string `gorm:"column:type;type:varchar(20)" json:"type"` // For now it's grafana or jaeger.
	Custom   bool   `gorm:"column:custom" json:"-"`
	Language string `gorm:"column:language;type:varchar(20);" json:"language"` // zh, en
}

// RouterInsertPage maps router to inserted page.
type RouterInsertPage struct {
	ID       int `gorm:"column:id"`
	RouterID int `gorm:"column:router_id;uniqueIndex"`
	PageID   int `gorm:"column:page_id;"`
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
			Where("router_insert_page.router_id = ?", routers[i].RouterID).
			Order(clause.Expr{
				SQL:  "CASE WHEN insert_page.language = ? THEN 0 ELSE 1 END ASC",
				Vars: []interface{}{language}},
			).
			Find(&page).Error
		if err != nil {
			return err
		}
		if page.PageID != 0 {
			routers[i].Page = &page
		}
	}

	return nil
}
