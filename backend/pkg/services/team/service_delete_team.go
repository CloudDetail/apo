// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package team

import (
	"context"
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) DeleteTeam(req *request.DeleteTeamRequest) error {
	filter := model.TeamFilter {
		ID: req.TeamID,
	}
	exists, err := s.dbRepo.TeamExist(filter)
	if err != nil {
		return err
	}

	if !exists {
		return model.NewErrWithMessage(errors.New("team does not exist"), code.TeamNotExistError)
	}

	permissionIDs, err := s.dbRepo.GetSubjectPermission(req.TeamID, model.PERMISSION_SUB_TYP_TEAM, model.PERMISSION_TYP_FEATURE)
	if err != nil {
		return err
	}

	var deleteTeamFunc = func(ctx context.Context) error {
		return s.dbRepo.DeleteTeam(ctx, req.TeamID)
	}

	var revokePermissionFunc = func(ctx context.Context) error {
		return s.dbRepo.RevokePermission(ctx, req.TeamID, model.PERMISSION_SUB_TYP_TEAM, model.PERMISSION_TYP_FEATURE, permissionIDs)
	}

	var deleteTeamUserFunc = func(ctx context.Context) error {
		return s.dbRepo.DeleteAllUserTeam(ctx, req.TeamID, "team")
	}

	var deleteAuthDataGroupFunc = func(ctx context.Context) error {
		return s.dbRepo.DeleteAuthDataGroup(ctx, req.TeamID, model.DATA_GROUP_SUB_TYP_TEAM)
	}
	return s.dbRepo.Transaction(context.Background(), deleteTeamFunc, revokePermissionFunc, deleteTeamUserFunc, deleteAuthDataGroupFunc)
}
