// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package permission

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
)

func (s *service) getUserFeatureIDs(ctx core.Context, userID int64) ([]int, error) {
	// 1. Get user's role
	roles, err := s.dbRepo.GetUserRole(userID)
	if err != nil {
		return nil, err
	}
	// 2. Get user's team
	teamIDs, err := s.dbRepo.GetUserTeams(userID)
	if err != nil {
		return nil, err
	}

	// 3. Get user's feature permission
	roleIDs := make([]int64, len(roles))
	for i := range roleIDs {
		roleIDs[i] = int64(roles[i].RoleID)
	}
	rolesFeatures, err := s.dbRepo.GetSubjectsPermission(roleIDs, model.PERMISSION_SUB_TYP_ROLE, model.PERMISSION_TYP_FEATURE)
	if err != nil {
		return nil, err
	}

	uFeatureIDs, err := s.dbRepo.GetSubjectPermission(userID, model.PERMISSION_SUB_TYP_USER, model.PERMISSION_TYP_FEATURE)
	if err != nil {
		return nil, err
	}

	teamFeatures, err := s.dbRepo.GetSubjectsPermission(teamIDs, model.PERMISSION_SUB_TYP_TEAM, model.PERMISSION_TYP_FEATURE)
	if err != nil {
		return nil, err
	}

	featureIDs := make([]int, 0, len(rolesFeatures)+len(teamFeatures)+len(uFeatureIDs))

	for _, rolePermission := range rolesFeatures {
		featureIDs = append(featureIDs, rolePermission.PermissionID)
	}

	for _, teamPermission := range teamFeatures {
		featureIDs = append(featureIDs, teamPermission.PermissionID)
	}

	featureIDs = append(featureIDs, uFeatureIDs...)

	return featureIDs, nil
}
