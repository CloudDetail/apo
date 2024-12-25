package database

// Router front end router.
type Router struct {
	RouterID         int    `gorm:"column:router_id;primary_key" json:"routerId"`
	RouterTo         string `gorm:"column:router_to;uniqueIndex" json:"to"`
	Custom           bool   `gorm:"column:custom;default:false" json:"-"`
	HideTimeSelector bool   `gorm:"column:hide_time_selector" json:"hideTimeSelector"`

	Page *InsertPage `gorm:"-" json:"page,omitempty"`
}

func (t *Router) TableName() string {
	return "router"
}

func (repo *daoRepo) GetItemRouter(items *[]MenuItem) error {
	if items == nil {
		return nil
	}
	routerIDs := make([]int, 0, len(*items))
	for _, item := range *items {
		routerIDs = append(routerIDs, item.RouterID)
	}

	var routers []Router
	if err := repo.db.Where("router_id IN ?", routerIDs).Find(&routers).Error; err != nil {
		return err
	}

	routerMap := make(map[int]Router)
	for _, router := range routers {
		routerMap[router.RouterID] = router
	}

	for i := range *items {
		if router, ok := routerMap[(*items)[i].RouterID]; ok {
			(*items)[i].Router = &router
		}
	}

	return nil
}
