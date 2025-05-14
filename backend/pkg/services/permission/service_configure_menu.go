// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package permission

import (
	"context"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) ConfigureMenu(ctx_core core.Context, req *request.ConfigureMenuRequest) error {
	filter := model.RoleFilter{
		Names: []string{model.ROLE_ADMIN, model.ROLE_VIEWER, model.ROLE_MANAGER},
	}
	roles, err := s.dbRepo.GetRoles(ctx_core, filter)
	if err != nil {
		return err
	}

	addPermissions, deletePermissions := make([][]int, len(roles)), make([][]int, len(roles))
	menuPermissionID, err := s.dbRepo.GetFeatureByName(ctx_core, "菜单管理")
	if err != nil {
		return err
	}
	for i, role := range roles {
		var err error
		addPermissions[i], deletePermissions[i], err =
			s.dbRepo.GetAddAndDeletePermissions(
				ctx_core,
				int64(role.RoleID),
				model.PERMISSION_SUB_TYP_ROLE,
				model.PERMISSION_TYP_FEATURE,
				req.PermissionList)
		if err != nil {
			return err
		}
	}

	for i := 0; i < len(deletePermissions[0]); {
		if deletePermissions[0][i] == menuPermissionID {
			deletePermissions[0] = append(deletePermissions[0][:i], deletePermissions[0][i+1:]...)
		} else {
			i++
		}
	}

	grantFunc := func(ctx context.Context) error {
		for i, role := range roles {
			if len(addPermissions[i]) > 0 {
				err := s.dbRepo.GrantPermission(ctx_core, ctx,
					int64(role.RoleID),
					model.PERMISSION_SUB_TYP_ROLE,
					model.PERMISSION_TYP_FEATURE,
					addPermissions[i])
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	revokeFunc := func(ctx context.Context) error {
		for i, role := range roles {
			if len(deletePermissions[i]) > 0 {
				err := s.dbRepo.RevokePermission(ctx_core, ctx,
					int64(role.RoleID),
					model.PERMISSION_SUB_TYP_ROLE,
					model.PERMISSION_TYP_FEATURE,
					deletePermissions[i])
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	return s.dbRepo.Transaction(ctx_core, context.Background(), grantFunc, revokeFunc)
}
