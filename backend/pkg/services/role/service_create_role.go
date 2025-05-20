// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package role

import (
	"github.com/CloudDetail/apo/backend/pkg/code"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

func (s *service) CreateRole(ctx core.Context, req *request.CreateRoleRequest) error {
	filter := model.RoleFilter{
		Name: req.RoleName,
	}
	roles, err := s.dbRepo.GetRoles(ctx, filter)
	if err != nil {
		return err
	}

	if len(roles) > 0 {
		return core.Error(code.RoleExistsError, "role already exists")
	}

	if len(req.PermissionList) > 0 {
		f, err := s.dbRepo.GetFeature(ctx, req.PermissionList)
		if err != nil {
			return err
		}

		if len(f) != len(req.PermissionList) {
			return core.Error(code.PermissionNotExistError, "permission does not exist")
		}
	}

	exist, err := s.dbRepo.UserExists(ctx, req.UserList...)
	if err != nil {
		return err
	}

	if !exist {
		return core.Error(code.UserNotExistsError, "user does not exist")
	}

	role := &database.Role{
		RoleName:    req.RoleName,
		Description: req.Description,
	}
	var createRoleFunc = func(ctx core.Context) error {
		return s.dbRepo.CreateRole(ctx, role)
	}

	var grantPermissionFunc = func(ctx core.Context) error {
		return s.dbRepo.GrantPermission(ctx, int64(role.RoleID), model.PERMISSION_SUB_TYP_ROLE, model.PERMISSION_TYP_FEATURE, req.PermissionList)
	}

	var grantRoleFunc = func(ctx core.Context) error {
		return s.dbRepo.GrantRoleWithRole(ctx, role.RoleID, req.UserList)
	}

	return s.dbRepo.Transaction(ctx, createRoleFunc, grantPermissionFunc, grantRoleFunc)
}
