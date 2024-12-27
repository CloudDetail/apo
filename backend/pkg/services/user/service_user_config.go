// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"sort"
)

// GetUserConfig Gets menus and routes that users can view.
func (s *service) GetUserConfig(req *request.GetUserConfigRequest) (response.GetUserConfigResponse, error) {
	var resp response.GetUserConfigResponse
	// 1. Get user's role
	roles, err := s.dbRepo.GetUserRole(req.UserID)
	if err != nil {
		return resp, err
	}
	// 2. TODO Get user's team
	// 3. Get user's feature permission
	subIDs := make([]int64, len(roles))
	var i int
	for ; i < len(roles); i++ {
		subIDs[i] = int64(roles[i].RoleID)
	}
	features, err := s.dbRepo.GetSubjectsPermission(subIDs, model.PERMISSION_SUB_TYP_ROLE, model.PERMISSION_TYP_FEATURE)
	if err != nil {
		return resp, err
	}

	uFeatureIDs, err := s.dbRepo.GetSubjectPermission(req.UserID, model.PERMISSION_SUB_TYP_USER, model.PERMISSION_TYP_FEATURE)
	if err != nil {
		return resp, err
	}

	featureIDs := make([]int, len(features))
	for i, feature := range features {
		featureIDs[i] = feature.PermissionID
	}

	featureIDs = append(featureIDs, uFeatureIDs...)
	// 4. Get menu item ids
	res, err := s.dbRepo.GetFeatureMapping(featureIDs, model.MAPPED_TYP_MENU)
	itemIDs := make([]int, len(res))
	for i := range res {
		itemIDs[i] = res[i].MappedID
	}
	if err != nil {
		return resp, err
	}

	// 5. Get menu item
	items, err := s.dbRepo.GetMenuItems()
	if err != nil {
		return resp, err
	}

	err = s.dbRepo.GetItemRouter(&items)
	if err != nil {
		return resp, err
	}
	var routers []*database.Router
	for i := range items {
		if items[i].Router != nil {
			routers = append(routers, items[i].Router)
		}
	}

	err = s.dbRepo.GetRouterInsertedPage(routers)
	if err != nil {
		return resp, err
	}

	err = s.dbRepo.GetMenuItemTans(&items, req.Language)
	if err != nil {
		return resp, err
	}
	menuItemMap := make(map[int]*database.MenuItem)
	var rootMenuItems []*database.MenuItem

	for _, item := range items {
		m := item
		menuItemMap[m.ItemID] = &m
	}

	addedToRoot := make(map[int]bool)

	for _, id := range itemIDs {
		item := menuItemMap[id]
		if item.ParentID == nil {
			if !addedToRoot[item.ItemID] {
				rootMenuItems = append(rootMenuItems, menuItemMap[item.ItemID])
				addedToRoot[item.ItemID] = true
			}
		} else {
			if parent, exists := menuItemMap[*item.ParentID]; exists {
				parent.Children = append(parent.Children, *menuItemMap[item.ItemID])

				if !addedToRoot[parent.ItemID] {
					rootMenuItems = append(rootMenuItems, parent)
					addedToRoot[parent.ItemID] = true
				}
			}
		}
	}

	sort.Slice(rootMenuItems, func(i, j int) bool {
		return rootMenuItems[i].Order < rootMenuItems[j].Order
	})

	resp.MenuItem = rootMenuItems
	return resp, nil
}
