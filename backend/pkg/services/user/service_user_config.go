package user

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

// GetUserConfig Gets menus and routes that users can view.
func (s *service) GetUserConfig(req *request.GetUserConfigRequest) (response.GetUserConfigResponse, error) {
	var resp response.GetUserConfigResponse
	if req.UserID == 0 {
		anonymousUser, err := s.dbRepo.GetAnonymousUser()
		if err != nil {
			return resp, err
		}

		req.UserID = anonymousUser.UserID
	}

	// 1. Get user's role
	roles, err := s.dbRepo.GetUserRole(req.UserID)
	if err != nil {
		return resp, err
	}
	// 2. TODO Get user's team
	// 3. Get user's feature permission
	subIDs := make([]int64, len(roles)+1)
	var i int
	for ; i < len(roles); i++ {
		subIDs[i] = int64(roles[i].RoleID)
	}
	subIDs[i] = req.UserID
	features, err := s.dbRepo.GetSubjectsPermission(subIDs, model.PERMISSION_TYP_FEATURE)
	if err != nil {
		return resp, err
	}
	featureIDs := make([]int, len(features))
	for i, feature := range features {
		featureIDs[i] = feature.PermissionID
	}
	// 4. Get menu item ids
	res, err := s.dbRepo.GetMappedMenuItem(featureIDs)
	itemIDs := make([]int, len(res))
	for i := range res {
		itemIDs[i] = res[i].MenuItemID
	}
	if err != nil {
		return resp, err
	}
	// 5. Get menu item
	items, err := s.dbRepo.GetMenuItems(itemIDs)
	if err != nil {
		return resp, err
	}
	err = s.dbRepo.GetItemRouter(&items)
	if err != nil {
		return resp, err
	}
	err = s.dbRepo.GetItemInsertPage(&items)
	if err != nil {
		return resp, err
	}
	menuItemMap := make(map[int]*database.MenuItem)
	var rootMenuItems []*database.MenuItem

	for _, item := range items {
		m := item
		menuItemMap[m.ItemID] = &m
	}

	for _, item := range items {
		if item.ParentID == nil {
			rootMenuItems = append(rootMenuItems, menuItemMap[item.ItemID])
		} else {
			if parent, exists := menuItemMap[*item.ParentID]; exists {
				parent.Children = append(parent.Children, *menuItemMap[item.ItemID])
			}
		}
	}

	resp.MenuItem = rootMenuItems
	return resp, nil
}
