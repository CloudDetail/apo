// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package team

import (
	"context"

	"github.com/CloudDetail/apo/backend/pkg/code"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) DeleteTeam(ctx_core core.Context, req *request.DeleteTeamRequest) error {
	filter := model.TeamFilter{
		ID: req.TeamID,
	}
	exists, err := s.dbRepo.TeamExist(ctx_core, filter)
	if err != nil {
		return err
	}

	if !exists {
		return core.Error(code.TeamNotExistError, "team does not exist")
	}

	permissionIDs, err := s.dbRepo.GetSubjectPermission(ctx_core, req.TeamID, model.PERMISSION_SUB_TYP_TEAM, model.PERMISSION_TYP_FEATURE)
	if err != nil {
		return err
	}

	var deleteTeamFunc = func(ctx context.Context) error {
		return s.dbRepo.DeleteTeam(ctx_core, ctx, req.TeamID)
	}

	var revokePermissionFunc = func(ctx context.Context) error {
		return s.dbRepo.RevokePermission(ctx_core, ctx, req.TeamID, model.PERMISSION_SUB_TYP_TEAM, model.PERMISSION_TYP_FEATURE, permissionIDs)
	}

	var deleteTeamUserFunc = func(ctx context.Context) error {
		return s.dbRepo.DeleteAllUserTeam(ctx_core, ctx, req.TeamID, "team")
	}

	var deleteAuthDataGroupFunc = func(ctx context.Context) error {
		return s.dbRepo.DeleteAuthDataGroup(ctx_core, ctx, req.TeamID, model.DATA_GROUP_SUB_TYP_TEAM)
	}
	return s.dbRepo.Transaction(ctx_core, context.Background(), deleteTeamFunc, revokePermissionFunc, deleteTeamUserFunc, deleteAuthDataGroupFunc)
}
