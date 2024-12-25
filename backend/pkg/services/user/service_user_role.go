// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"context"
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetRoles() (response.GetRoleResponse, error) {
	roles, err := s.dbRepo.GetRoles(model.RoleFilter{})
	var resp response.GetRoleResponse
	if err != nil {
		return resp, err
	}

	resp = roles
	return resp, nil
}

func (s *service) GetUserRole(req *request.GetUserRoleRequest) (response.GetUserRoleResponse, error) {
	userRole, err := s.dbRepo.GetUserRole(req.UserID)
	if err != nil {
		return nil, err
	}

	roleIDs := make([]int, len(userRole))
	for i, roleID := range userRole {
		roleIDs[i] = roleID.RoleID
	}
	filter := model.RoleFilter{IDs: roleIDs}
	roles, err := s.dbRepo.GetRoles(filter)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (s *service) RoleOperation(req *request.RoleOperationRequest) error {
	// 1. get user's role
	userRole, err := s.dbRepo.GetUserRole(req.UserID)
	if err != nil {
		return err
	}

	// 2. get all roles
	roles, err := s.dbRepo.GetRoles(model.RoleFilter{})
	if err != nil {
		return err
	}

	roleMap := make(map[int]struct{})
	for _, role := range roles {
		roleMap[role.RoleID] = struct{}{}
	}

	// 3. determine grant and revoke
	var addRoles, deleteRoles []int

	userRoleMap := make(map[int]struct{})
	for _, ur := range userRole {
		userRoleMap[ur.RoleID] = struct{}{}
	}

	for _, role := range req.RoleList {
		if _, exists := roleMap[role]; !exists {
			return model.NewErrWithMessage(errors.New("role does not exist"), code.RoleNotExistsError)
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

	var grantFunc = func(txCtx context.Context) error {
		return s.dbRepo.GrantRole(txCtx, req.UserID, addRoles)
	}

	var revokeFunc = func(txCtx context.Context) error {
		return s.dbRepo.RevokeRole(txCtx, req.UserID, deleteRoles)
	}

	return s.dbRepo.Transaction(context.Background(), grantFunc, revokeFunc)
}
