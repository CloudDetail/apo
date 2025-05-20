// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package role

import (
	"github.com/CloudDetail/apo/backend/pkg/code"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

func (s *service) GetRoles(ctx core.Context) (response.GetRoleResponse, error) {
	roles, err := s.dbRepo.GetRoles(ctx, model.RoleFilter{})
	var resp response.GetRoleResponse
	if err != nil {
		return resp, err
	}

	resp = roles
	return resp, nil
}

func (s *service) GetUserRole(ctx core.Context, req *request.GetUserRoleRequest) (response.GetUserRoleResponse, error) {
	userRole, err := s.dbRepo.GetUserRole(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	roleIDs := make([]int, len(userRole))
	for i, roleID := range userRole {
		roleIDs[i] = roleID.RoleID
	}
	filter := model.RoleFilter{IDs: roleIDs}
	roles, err := s.dbRepo.GetRoles(ctx, filter)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (s *service) RoleOperation(ctx core.Context, req *request.RoleOperationRequest) error {
	// 1. get user's role
	userRole, err := s.dbRepo.GetUserRole(ctx, req.UserID)
	if err != nil {
		return err
	}

	// 2. get all roles
	roles, err := s.dbRepo.GetRoles(ctx, model.RoleFilter{})
	if err != nil {
		return err
	}

	roleMap := make(map[int]struct{})
	for _, role := range roles {
		roleMap[role.RoleID] = struct{}{}
	}

	addRoles, deleteRoles, err := GetAddDeleteRoles(userRole, req.RoleList, roles)
	if err != nil {
		return err
	}

	var grantFunc = func(txCtx core.Context) error {
		return s.dbRepo.GrantRoleWithUser(ctx, req.UserID, addRoles)
	}

	var revokeFunc = func(txCtx core.Context) error {
		return s.dbRepo.RevokeRole(ctx, req.UserID, deleteRoles)
	}

	return s.dbRepo.Transaction(ctx, grantFunc, revokeFunc)
}

// GetAddDeleteRoles Determine grant and revoke roles.
func GetAddDeleteRoles(userRoles []database.UserRole, want []int, all []database.Role) (addRoles []int, deleteRoles []int, err error) {
	roleMap := make(map[int]struct{})
	for _, role := range all {
		roleMap[role.RoleID] = struct{}{}
	}

	userRoleMap := make(map[int]struct{})
	for _, ur := range userRoles {
		userRoleMap[ur.RoleID] = struct{}{}
	}

	for _, role := range want {
		if _, exists := roleMap[role]; !exists {
			return nil, nil, core.Error(code.RoleNotExistsError, "role does not exist")
		}
		if _, hasRole := userRoleMap[role]; !hasRole {
			addRoles = append(addRoles, role)
		} else {
			delete(userRoleMap, role)
		}
	}

	for roleID := range userRoleMap {
		deleteRoles = append(deleteRoles, roleID)
	}

	return
}
