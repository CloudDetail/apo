// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

// MenuItem is a menu item on the left or top menu bar.
//
// ! Handle the `Key` field carefully, it is a reserved word in MySQL, and must be delimited with the identifier '"'
type MenuItem struct {
	ItemID       int    `gorm:"column:item_id;primary_key" json:"itemId"`
	Key          string `gorm:"column:key;type:varchar(20);uniqueIndex" json:"key"`
	Label        string `gorm:"-" json:"label"` // AKA item name.
	Icon         string `gorm:"column:icon;type:varchar(150)" json:"icon"`
	ParentID     *int   `gorm:"column:parent_id" json:"-"`
	Abbreviation string `gorm:"-" json:"abbreviation,omitempty"`
	RouterID     int    `gorm:"column:router_id" json:"-"`
	Order        int    `gorm:"column:sort_order;index:sort_order_idx" json:"-"` // The order of a menu item.

	Children []MenuItem `gorm:"-" json:"children,omitempty" swaggerignore:"true"`
	Router   *Router    `gorm:"-" json:"router,omitempty"` // Frontend router.

	AccessInfo string `gorm:"access_info"`
}

func (t *MenuItem) TableName() string {
	return "menu_item"
}

func (repo *daoRepo) GetMenuItems() ([]MenuItem, error) {
	var menuItems []MenuItem

	err := repo.db.Find(&menuItems).Error
	return menuItems, err
}
