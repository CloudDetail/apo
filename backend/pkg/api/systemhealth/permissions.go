// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package systemhealth

import (
	"slices"

	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/datagroup"
	"github.com/CloudDetail/apo/backend/pkg/model/profile"
)

func (h *handler) authorizedSystemHealthReportGroupIDs(ctx core.Context) ([]int64, error) {
	permissionGroupIDs, err := h.dbRepo.GetDataGroupIDsByUserId(ctx, ctx.UserID())
	if err != nil {
		return nil, err
	}
	groups, err := h.visibleSystemGroups(ctx)
	if err != nil {
		return nil, err
	}
	return systemHealthReportGroupIDs(groups, slices.Contains(permissionGroupIDs, int64(0))), nil
}

func systemHealthReportGroupIDs(groups []*datagroup.DataGroupTreeNode, includeRoot bool) []int64 {
	groupIDs := make([]int64, 0, len(groups)+1)
	if includeRoot {
		groupIDs = append(groupIDs, 0)
	}
	for _, group := range groups {
		if group != nil && group.GroupID > 0 {
			groupIDs = append(groupIDs, group.GroupID)
		}
	}
	return groupIDs
}

func systemHealthReportGroupAllowed(groupIDs []int64, groupID int64) bool {
	return slices.Contains(groupIDs, groupID)
}

func (h *handler) canManageSystemHealthConfig(ctx core.Context) (bool, error) {
	user, err := h.dbRepo.GetUserInfo(ctx, ctx.UserID())
	if err != nil {
		return false, err
	}
	return isSystemHealthAdminUser(user), nil
}

func isSystemHealthAdminUser(user profile.User) bool {
	for _, role := range user.RoleList {
		if role.RoleName == model.ROLE_ADMIN {
			return true
		}
	}
	return false
}
