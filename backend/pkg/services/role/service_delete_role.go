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

func (s *service) DeleteRole(req *request.DeleteRoleRequest) error {
	exists, err := s.dbRepo.RoleExists(req.RoleID)
	if err != nil {
		return err
	}

	if !exists {
		return model.NewErrWithMessage(errors.New("role does not exist"), code.RoleNotExistsError)
	}

	var deleteFunc = func(ctx context.Context) error {
		return s.dbRepo.DeleteRole(ctx, req.RoleID)
	}

	return s.dbRepo.Transaction(context.Background(), deleteFunc)
}
