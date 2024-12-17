package database

// MenuItem is a menu item on the left menu bar.
type MenuItem struct {
	ItemID       int    `gorm:"column:item_id;primary_key" json:"itemId"`
	Key          string `gorm:"column:key" json:"key"`
	Label        string `gorm:"column:label" json:"label"` // AKA item name.
	RouterID     int    `gorm:"column:router_id" json:"-"` // Router id.
	InsertPageID int    `gorm:"column:insert_page_id" json:"-"`
	Icon         string `gorm:"column:icon" json:"icon"`
	ParentID     *int   `gorm:"column:parent_id" json:"-"`
	Abbreviation string `gorm:"column:abbreviation" json:"abbreviation"`

	Children []MenuItem  `gorm:"-" json:"children" swaggerignore:"true"`
	Router   *Router     `gorm:"-" json:"to"` // Frontend router.
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