// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package permission

import (
	"encoding/json"
	"sort"
	"strings"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

// GetUserConfig Gets menus and routes that users can view.
func (s *service) GetUserConfig(ctx core.Context, req *request.GetUserConfigRequest) (response.GetUserConfigResponse, error) {
	var resp response.GetUserConfigResponse
	featureIDs, err := s.getUserFeatureIDs(ctx, req.UserID)
	if err != nil {
		return resp, err
	}

	res, err := s.dbRepo.GetFeatureMappingByFeature(ctx, featureIDs, model.MAPPED_TYP_MENU)
	itemIDs := make([]int, len(res))
	allowedItemIDs := make(map[int]struct{}, len(res))
	for i := range res {
		itemIDs[i] = res[i].MappedID
		allowedItemIDs[res[i].MappedID] = struct{}{}
	}
	if err != nil {
		return resp, err
	}

	items, err := s.dbRepo.GetMenuItems(ctx)
	if err != nil {
		return resp, err
	}

	err = s.dbRepo.FillItemRouter(ctx, &items)
	if err != nil {
		return resp, err
	}
	items = s.filterCustomDashboardMenuItems(ctx, items, req.UserID)
	itemIDs = visibleAllowedMenuItemIDs(items, allowedItemIDs)
	var routers []*database.Router
	allowSharedCustomDashboards := hasCustomDashboardRootMenuPermission(items, allowedItemIDs)
	for i := range items {
		if visibleAllowedMenuItem(items[i], allowedItemIDs, allowSharedCustomDashboards) && items[i].Router != nil {
			routers = append(routers, items[i].Router)
		}
	}

	err = s.dbRepo.GetRouterInsertedPage(ctx, routers, req.Language)
	if err != nil {
		return resp, err
	}

	err = s.dbRepo.GetMenuItemTans(ctx, &items, req.Language)
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
		if item == nil {
			continue
		}
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

	// Get routers.
	resp.Routes = make([]string, len(routers))
	for i, router := range routers {
		resp.Routes[i] = router.RouterTo
	}

	sort.Slice(rootMenuItems, func(i, j int) bool {
		return rootMenuItems[i].Order < rootMenuItems[j].Order
	})

	resp.MenuItem = rootMenuItems
	return resp, nil
}

func visibleAllowedMenuItemIDs(items []database.MenuItem, allowedItemIDs map[int]struct{}) []int {
	ids := make([]int, 0, len(items))
	allowSharedCustomDashboards := hasCustomDashboardRootMenuPermission(items, allowedItemIDs)
	for _, item := range items {
		if visibleAllowedMenuItem(item, allowedItemIDs, allowSharedCustomDashboards) {
			ids = append(ids, item.ItemID)
		}
	}
	return ids
}

func visibleAllowedMenuItem(item database.MenuItem, allowedItemIDs map[int]struct{}, allowSharedCustomDashboards bool) bool {
	if _, allowed := allowedItemIDs[item.ItemID]; allowed {
		return true
	}
	return allowSharedCustomDashboards && item.Router != nil && isConcreteCustomDashboardRoute(item.Router.RouterTo)
}

func hasCustomDashboardRootMenuPermission(items []database.MenuItem, allowedItemIDs map[int]struct{}) bool {
	for _, item := range items {
		if item.Key != "customDashboard" {
			continue
		}
		_, allowed := allowedItemIDs[item.ItemID]
		return allowed
	}
	return false
}

func (s *service) filterCustomDashboardMenuItems(ctx core.Context, items []database.MenuItem, userID int64) []database.MenuItem {
	filtered := make([]database.MenuItem, 0, len(items))
	for _, item := range items {
		if item.Router == nil || !isConcreteCustomDashboardRoute(item.Router.RouterTo) {
			filtered = append(filtered, item)
			continue
		}
		id := strings.TrimPrefix(item.Router.RouterTo, "/custom-dashboard/")
		row, err := s.dbRepo.GetCustomDashboard(ctx, id)
		if err != nil || row == nil {
			continue
		}
		if customDashboardVisibleToUser(ctx, s.dbRepo, row, userID) {
			if item.Label == "" {
				item.Label = row.Name
			}
			if item.Abbreviation == "" {
				item.Abbreviation = row.Name
			}
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func isConcreteCustomDashboardRoute(route string) bool {
	return strings.HasPrefix(route, "/custom-dashboard/") &&
		!strings.Contains(route, ":") &&
		!strings.HasSuffix(route, "/edit") &&
		strings.TrimPrefix(route, "/custom-dashboard/") != ""
}

func isConcreteCustomDashboardEditRoute(route string) bool {
	if !strings.HasPrefix(route, "/custom-dashboard/") ||
		strings.Contains(route, ":") ||
		!strings.HasSuffix(route, "/edit") {
		return false
	}
	id := strings.TrimSuffix(strings.TrimPrefix(route, "/custom-dashboard/"), "/edit")
	return id != ""
}

func customDashboardVisibleToUser(ctx core.Context, repo database.Repo, row *database.CustomDashboard, userID int64) bool {
	if row.OwnerID == userID || row.Visibility == "everyone" {
		return true
	}
	if row.Visibility != "team" {
		return false
	}
	teamIDs, err := decodeDashboardMenuTeamIDs(row.TeamIDs)
	if err != nil {
		return false
	}
	userTeams, err := repo.GetUserTeams(ctx, userID)
	if err != nil {
		return false
	}
	teamSet := map[int64]struct{}{}
	for _, teamID := range userTeams {
		teamSet[teamID] = struct{}{}
	}
	for _, teamID := range teamIDs {
		if _, ok := teamSet[teamID]; ok {
			return true
		}
	}
	return false
}

func decodeDashboardMenuTeamIDs(raw string) ([]int64, error) {
	if raw == "" {
		return nil, nil
	}
	var ids []int64
	err := json.Unmarshal([]byte(raw), &ids)
	return ids, err
}
