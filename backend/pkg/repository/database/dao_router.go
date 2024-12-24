package database

// Router front end router.
type Router struct {
	RouterID         int    `gorm:"column:router_id;primary_key" json:"routerId"`
	RouterTo         string `gorm:"column:router_to;uniqueIndex" json:"to"`
	HideTimeSelector bool   `gorm:"column:hide_time_selector" json:"hideTimeSelector"`
	MenuItemID       int    `gorm:"column:menu_item_id" json:"-"`

	MenuItemKey string `gorm:"-" json:"-"`
}

func (t *Router) TableName() string {
	return "router"
}

func (repo *daoRepo) GetItemRouter(items *[]MenuItem) error {
	if items == nil {
		return nil
	}
	itemIDs := make([]int, 0, len(*items))
	for _, item := range *items {
		itemIDs = append(itemIDs, item.ItemID)
	}

	var routers []Router
	if err := repo.db.Table("router").
		Select("router_id, router_to, hide_time_selector, menu_item_id").
		Where("menu_item_id IN ?", itemIDs).
		Find(&routers).Error; err != nil {
		return err
	}

	routerMap := make(map[int]Router)
	for _, router := range routers {
		routerMap[router.MenuItemID] = router
	}

	for i := range *items {
		if router, ok := routerMap[(*items)[i].ItemID]; ok {
			(*items)[i].Router = &router
		}
	}

	return nil
}
