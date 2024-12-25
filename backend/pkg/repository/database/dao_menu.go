package database

// MenuItem is a menu item on the left or top menu bar.
type MenuItem struct {
	ItemID       int    `gorm:"column:item_id;primary_key" json:"itemId"`
	Key          string `gorm:"column:key;uniqueIndex" json:"key"`
	Label        string `gorm:"-" json:"label"` // AKA item name.
	Icon         string `gorm:"column:icon" json:"icon"`
	ParentID     *int   `gorm:"column:parent_id" json:"-"`
	Abbreviation string `gorm:"-" json:"abbreviation,omitempty"`
	RouterID     int    `gorm:"column:router_id" json:"-"`

	Children []MenuItem `gorm:"-" json:"children,omitempty" swaggerignore:"true"`
	Router   *Router    `gorm:"-" json:"router,omitempty"` // Frontend router.
}

func (t *MenuItem) TableName() string {
	return "menu_item"
}

func (repo *daoRepo) GetMenuItems() ([]MenuItem, error) {
	var menuItems []MenuItem

	err := repo.db.Find(&menuItems).Error
	return menuItems, err
}
