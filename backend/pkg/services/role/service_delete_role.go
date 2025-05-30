// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package role

import (
	"github.com/CloudDetail/apo/backend/pkg/code"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) DeleteRole(ctx core.Context, req *request.DeleteRoleRequest) error {
	exists, err := s.dbRepo.RoleExists(ctx, req.RoleID)
	if err != nil {
		return err
	}

	if !exists {
		return core.Error(code.RoleNotExistsError, "role does not exist")
	}

	granted, err := s.dbRepo.RoleGranted(ctx, req.RoleID)
	if err != nil {
		return err
	}

	if granted {
		return core.Error(code.RoleGrantedError, "role has been granted")
	}

	var revokeRoleFunc = func(ctx core.Context) error {
		return s.dbRepo.RevokeRoleWithRole(ctx, req.RoleID)
	}

	var deleteFunc = func(ctx core.Context) error {
		return s.dbRepo.DeleteRole(ctx, req.RoleID)
	}

	return s.dbRepo.Transaction(ctx, revokeRoleFunc, deleteFunc)
}
