// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package role

import (
	"github.com/CloudDetail/apo/backend/pkg/code"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) UpdateRole(ctx core.Context, req *request.UpdateRoleRequest) error {
	idFilter := model.RoleFilter{
		ID: req.RoleID,
	}

	role, err := s.dbRepo.GetRoles(ctx, idFilter)
	if err != nil {
		return err
	}

	if len(role) == 0 {
		return core.Error(code.RoleNotExistsError, "role does not exist")
	}

	if role[0].RoleName != req.RoleName {
		nameFilter := model.RoleFilter{
			Name: req.RoleName,
		}

		existRole, err := s.dbRepo.GetRoles(ctx, nameFilter)
		if err != nil {
			return err
		}

		if len(existRole) > 0 {
			return core.Error(code.RoleExistsError, "role already exist")
		}
	}

	toAdd, toDelete, err := s.dbRepo.GetAddAndDeletePermissions(ctx, int64(req.RoleID), model.PERMISSION_SUB_TYP_ROLE, model.PERMISSION_TYP_FEATURE, req.PermissionList)
	if err != nil {
		return err
	}

	var updateRoleFunc = func(ctx core.Context) error {
		return s.dbRepo.UpdateRole(ctx, req.RoleID, req.RoleName, req.Description)
	}

	var grantPermissionFunc = func(ctx core.Context) error {
		return s.dbRepo.GrantPermission(ctx, int64(req.RoleID), model.PERMISSION_SUB_TYP_ROLE, model.PERMISSION_TYP_FEATURE, toAdd)
	}

	var revokePermissionFunc = func(ctx core.Context) error {
		return s.dbRepo.RevokePermission(ctx, int64(req.RoleID), model.PERMISSION_SUB_TYP_ROLE, model.PERMISSION_TYP_FEATURE, toDelete)
	}

	return s.dbRepo.Transaction(ctx, updateRoleFunc, grantPermissionFunc, revokePermissionFunc)
}
