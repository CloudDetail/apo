// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"context"
	"errors"

	"github.com/CloudDetail/apo/backend/pkg/code"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
)

func (s *service) RemoveUser(ctx_core core.Context, userID int64) error {
	exists, err := s.dbRepo.UserExists(userID)
	if err != nil {
		return err
	}

	if !exists {
		return core.Error(code.UserNotExistsError, "user does not exist")
	}

	roles, err := s.dbRepo.GetUserRole(userID)
	if err != nil {
		return err
	}

	roleIDs := make([]int, 0, len(roles))
	for _, role := range roles {
		roleIDs = append(roleIDs, role.RoleID)
	}

	user, err := s.dbRepo.GetUserInfo(userID)
	if err != nil {
		return err
	}

	var revokeRoleFunc = func(ctx context.Context) error {
		return s.dbRepo.RevokeRole(ctx, userID, roleIDs)
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

	var removeDifyUserFunc = func(ctx context.Context) error {
		resp, err := s.difyRepo.RemoveUser(user.Username)
		if err != nil || resp.Result != "success" {
			return errors.New("failed to remove user in dify")
		}
		return nil
	}

	var revokeFeaturePermFunc = func(ctx context.Context) error {
		return s.dbRepo.RevokePermission(ctx, userID, model.PERMISSION_SUB_TYP_USER, model.PERMISSION_TYP_FEATURE, nil)
	}

	return s.dbRepo.Transaction(context.Background(), revokeFeaturePermFunc, deleteAuthDataGroupFunc, revokeRoleFunc, removeFromTeam, removeUserFunc, removeDifyUserFunc)
}
