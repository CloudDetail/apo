// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package role

import (
	"context"

	"github.com/CloudDetail/apo/backend/pkg/code"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

func (s *service) CreateRole(ctx_core core.Context, req *request.CreateRoleRequest) error {
	filter := model.RoleFilter{
		Name: req.RoleName,
	}
	roles, err := s.dbRepo.GetRoles(ctx_core, filter)
	if err != nil {
		return err
	}

	if len(roles) > 0 {
		return core.Error(code.RoleExistsError, "role already exists")
	}

	if len(req.PermissionList) > 0 {
		f, err := s.dbRepo.GetFeature(ctx_core, req.PermissionList)
		if err != nil {
			return err
		}

		if len(f) != len(req.PermissionList) {
			return core.Error(code.PermissionNotExistError, "permission does not exist")
		}
	}

	exist, err := s.dbRepo.UserExists(ctx_core, req.UserList...)
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
	var createRoleFunc = func(ctx context.Context) error {
		return s.dbRepo.CreateRole(ctx_core, ctx, role)
	}

	var grantPermissionFunc = func(ctx context.Context) error {
		return s.dbRepo.GrantPermission(ctx_core, ctx, int64(role.RoleID), model.PERMISSION_SUB_TYP_ROLE, model.PERMISSION_TYP_FEATURE, req.PermissionList)
	}

	var grantRoleFunc = func(ctx context.Context) error {
		return s.dbRepo.GrantRoleWithRole(ctx_core, ctx, role.RoleID, req.UserList)
	}

	return s.dbRepo.Transaction(ctx_core, context.Background(), createRoleFunc, grantPermissionFunc, grantRoleFunc)
}
