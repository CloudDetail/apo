package database

// MenuItem is a menu item on the left menu bar.
type MenuItem struct {
	ItemID         int    `gorm:"column:item_id;primary_key" json:"itemId"`
	Key            string `gorm:"column:key" json:"key"`
	Label          string `gorm:"column:label" json:"label"`      // AKA item name.
	EnLabel        string `gorm:"column:en_label" json:"enLabel"` // English item name.
	RouterID       int    `gorm:"column:router_id" json:"-"`      // Router id.
	InsertPageID   int    `gorm:"column:insert_page_id" json:"-"`
	Icon           string `gorm:"column:icon" json:"icon"`
	ParentID       *int   `gorm:"column:parent_id" json:"-"`
	Abbreviation   string `gorm:"column:abbreviation" json:"abbreviation"`
	EnAbbreviation string `gorm:"column:en_abbreviation" json:"enAbbreviation"` // English abbreviation.

	Children []MenuItem  `gorm:"-" json:"children" swaggerignore:"true"`
	Router   *Router     `gorm:"-" json:"router"` // Frontend router.
	Page     *InsertPage `gorm:"-" json:"page,omitempty"`
}

func (t *MenuItem) TableName() string {
	return "menu_item"
}

func (repo *daoRepo) GetMenuItems() ([]MenuItem, error) {
	var menuItems []MenuItem

	err := repo.db.Find(&menuItems).Error
	return menuItems, err
}
