// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package role

import (
	"context"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) DeleteRole(req *request.DeleteRoleRequest) error {
	exists, err := s.dbRepo.RoleExists(req.RoleID)
	if err != nil {
		return err
	}

	if !exists {
		return core.Error(code.RoleNotExistsError, "role does not exist")
	}

	granted, err := s.dbRepo.RoleGranted(req.RoleID)
	if err != nil {
		return err
	}

	if granted {
		return core.Error(code.RoleGrantedError, "role has been granted")
	}

	var revokeRoleFunc = func(ctx context.Context) error {
		return s.dbRepo.RevokeRoleWithRole(ctx, req.RoleID)
	}

	var deleteFunc = func(ctx context.Context) error {
		return s.dbRepo.DeleteRole(ctx, req.RoleID)
	}

	return s.dbRepo.Transaction(context.Background(), revokeRoleFunc, deleteFunc)
}
