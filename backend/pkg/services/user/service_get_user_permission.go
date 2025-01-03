// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

func (s *service) getUserFeature(userID int64, language string) (resp response.GetSubjectFeatureResponse, err error) {
	// 1. Get user's role
	userRoles, err := s.dbRepo.GetUserRole(userID)
	if err != nil {
		return nil, err
	}

	roleIDs := make([]int64, len(userRoles))
	for i := range userRoles {
		roleIDs[i] = int64(userRoles[i].RoleID)
	}

	rolePermission, err := s.dbRepo.GetSubjectsPermission(roleIDs, model.PERMISSION_SUB_TYP_ROLE, model.PERMISSION_TYP_FEATURE)
	rolePermIDs := make([]int, len(rolePermission))
	for i, rolePerm := range rolePermission {
		rolePermIDs[i] = rolePerm.PermissionID
	}

	roleFeatures, err := s.dbRepo.GetFeature(rolePermIDs)
	for _, feature := range roleFeatures {
		feature.Source = model.PERMISSION_SUB_TYP_ROLE
		resp = append(resp, feature)
	}
	// 2. TODO Get user's team

	// 3. Get user's permission.
	userPermissions, err := s.dbRepo.GetSubjectPermission(userID, model.PERMISSION_SUB_TYP_USER, model.PERMISSION_TYP_FEATURE)
	if err != nil {
		return nil, err
	}

	// 4. Get features.
	features, err := s.dbRepo.GetFeature(userPermissions)
	if err != nil {
		return nil, err
	}

	for _, feature := range features {
		feature.Source = model.PERMISSION_SUB_TYP_USER
		resp = append(resp, feature)
	}

	// 5. Translation
	err = s.dbRepo.GetFeatureTans((*[]database.Feature)(&resp), language)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
