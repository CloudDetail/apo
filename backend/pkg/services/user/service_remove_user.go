// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"context"
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/model"
)

func (s *service) RemoveUser(userID int64) error {
	exists, err := s.dbRepo.UserExists(userID)
	if err != nil {
		return err
	}

	if !exists {
		return model.NewErrWithMessage(errors.New("user does not exist"), code.UserNotExistsError)
	}

	// 1. Get roles.
	roles, err := s.dbRepo.GetUserRole(userID)
	if err != nil {
		return err
	}

	roleIDs := make([]int, 0, len(roles))
	for _, role := range roles {
		roleIDs = append(roleIDs, role.RoleID)
	}

	var deleteAuthDataGroupFunc = func(ctx context.Context) error {
		return s.dbRepo.DeleteAuthDataGroup(ctx, userID, model.DATA_GROUP_SUB_TYP_USER)
	}

	var removeFromTeam = func(ctx context.Context) error {
		return s.dbRepo.DeleteAllUserTeam(ctx, userID, "user")
	}

	var removeUserFunc = func(ctx context.Context) error {
		return s.dbRepo.RemoveUser(ctx, userID)
	}

	var revokeFeaturePermFunc = func(ctx context.Context) error {
		return s.dbRepo.RevokePermission(ctx, userID, model.PERMISSION_SUB_TYP_USER, model.PERMISSION_TYP_FEATURE, nil)
	}

	return s.dbRepo.Transaction(context.Background(), revokeFeaturePermFunc, removeUserFunc, removeFromTeam, deleteAuthDataGroupFunc)
}
