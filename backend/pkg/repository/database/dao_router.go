// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

// Router front end router.
type Router struct {
	RouterID         int    `gorm:"column:router_id;primary_key" json:"routerId"`
	RouterTo         string `gorm:"column:router_to;uniqueIndex;type:varchar(200)" json:"to"`
	Custom           bool   `gorm:"column:custom;default:false" json:"-"`
	HideTimeSelector bool   `gorm:"column:hide_time_selector" json:"hideTimeSelector"`

	Page *InsertPage `gorm:"-" json:"page,omitempty"`
}

func (t *Router) TableName() string {
	return "router"
}

func (repo *daoRepo) FillItemRouter(items *[]MenuItem) error {
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

func (repo *daoRepo) GetItemsRouter(itemIDs []int) ([]Router, error) {
	var routers []Router

	if err := repo.db.Where("router_id IN ?", itemIDs).Find(&routers).Error; err != nil {
		return nil, err
	}

	return routers, nil
}