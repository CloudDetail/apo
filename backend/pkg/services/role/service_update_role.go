// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package role

import (
	"context"
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) UpdateRole(req *request.UpdateRoleRequest) error {
	exists, err := s.dbRepo.RoleExists(req.RoleID)
	if err != nil {
		return err
	}

	filter := model.RoleFilter{
		Name: req.RoleName,
	}

	r, err := s.dbRepo.GetRoles(filter)
	if err != nil {
		return err
	}
	if len(r) > 0 {
		return model.NewErrWithMessage(errors.New("role already exist"), code.RoleExistsError)
	}

	if !exists {
		return model.NewErrWithMessage(errors.New("role does not exist"), code.RoleNotExistsError)
	}

	toAdd, toDelete, err := s.dbRepo.GetAddAndDeletePermissions(int64(req.RoleID), model.PERMISSION_SUB_TYP_ROLE, model.PERMISSION_TYP_FEATURE, req.PermissionList)
	if err != nil {
		return err
	}

	var updateRoleFunc = func(ctx context.Context) error {
		return s.dbRepo.UpdateRole(ctx, req.RoleID, req.RoleName, req.Description)
	}

	var grantPermissionFunc = func(ctx context.Context) error {
		return s.dbRepo.GrantPermission(ctx, int64(req.RoleID), model.PERMISSION_SUB_TYP_ROLE, model.PERMISSION_TYP_FEATURE, toAdd)
	}

	var revokePermissionFunc = func(ctx context.Context) error {
		return s.dbRepo.RevokePermission(ctx, int64(req.RoleID), model.PERMISSION_SUB_TYP_ROLE, model.PERMISSION_TYP_FEATURE, toDelete)
	}

	return s.dbRepo.Transaction(context.Background(), updateRoleFunc, grantPermissionFunc, revokePermissionFunc)
}
