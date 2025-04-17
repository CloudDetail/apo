// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package role

import (
	"context"
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

func (s *service) CreateRole(req *request.CreateRoleRequest) error {
	filter := model.RoleFilter{
		Name: req.RoleName,
	}
	roles, err := s.dbRepo.GetRoles(filter)
	if err != nil {
		return err
	}

	if len(roles) > 0 {
		return model.NewErrWithMessage(errors.New("role already exists"), code.RoleExistsError)
	}

	if len(req.PermissionList) > 0 {
		f, err := s.dbRepo.GetFeature(req.PermissionList)
		if err != nil {
			return err
		}
	
		if len(f) != len(req.PermissionList) {
			return model.NewErrWithMessage(errors.New("permission does not exist"), code.PermissionNotExistError)
		}
	}

	exist, err := s.dbRepo.UserExists(req.UserList...)
	if err != nil {
		return err
	}

	if !exist {
		return model.NewErrWithMessage(errors.New("user does not exist"), code.UserNotExistsError)
	}

	role := &database.Role{
		RoleName:    req.RoleName,
		Description: req.Description,
	}
	var createRoleFunc = func(ctx context.Context) error {
		return s.dbRepo.CreateRole(ctx, role)
	}

	var grantPermissionFunc = func(ctx context.Context) error {
		return s.dbRepo.GrantPermission(ctx, int64(role.RoleID), model.PERMISSION_SUB_TYP_ROLE, model.PERMISSION_TYP_FEATURE, req.PermissionList)
	}

	var grantRoleFunc = func(ctx context.Context) error {
		return s.dbRepo.GrantRoleWithRole(ctx, role.RoleID, req.UserList)
	}

	return s.dbRepo.Transaction(context.Background(), createRoleFunc, grantPermissionFunc, grantRoleFunc)
}
