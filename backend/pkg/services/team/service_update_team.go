// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package team

import (
	"context"
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

func (s *service) UpdateTeam(req *request.UpdateTeamRequest) error {
	exists, err := s.dbRepo.TeamExist(req.TeamID)
	if err != nil {
		return err
	}

	if !exists {
		return model.NewErrWithMessage(errors.New("team does not exist"), code.TeamAlreadyExistError)
	}

	toAddFeature, toDeleteFeature, err := s.dbRepo.GetAddAndDeletePermissions(req.TeamID, model.PERMISSION_SUB_TYP_TEAM, model.PERMISSION_TYP_FEATURE, req.FeatureList)
	if err != nil {
		return err
	}

	toModifyDg, toDeleteDg, err := s.dbRepo.GetModifyAndDeleteDataGroup(req.TeamID, model.DATA_GROUP_SUB_TYP_TEAM, req.DataGroupPermissions)
	if err != nil {
		return err
	}

	team := database.Team{
		TeamID:      req.TeamID,
		TeamName:    req.TeamName,
		Description: req.Description,
	}
	var updateTeamFunc = func(ctx context.Context) error {
		return s.dbRepo.UpdateTeam(ctx, team)
	}

	var grantPermissionFunc = func(ctx context.Context) error {
		return s.dbRepo.GrantPermission(ctx, req.TeamID, model.PERMISSION_SUB_TYP_TEAM, model.PERMISSION_TYP_FEATURE, toAddFeature)
	}

	var revokePermissionFunc = func(ctx context.Context) error {
		return s.dbRepo.RevokePermission(ctx, req.TeamID, model.PERMISSION_SUB_TYP_TEAM, model.PERMISSION_TYP_FEATURE, toDeleteFeature)
	}

	var assignDataGroupFunc = func(ctx context.Context) error {
		return s.dbRepo.AssignDataGroup(ctx, toModifyDg)
	}

	var removeDataGroupFunc = func(ctx context.Context) error {
		return s.dbRepo.RevokeDataGroup(ctx, toDeleteDg)
	}

	return s.dbRepo.Transaction(context.Background(), updateTeamFunc, grantPermissionFunc, revokePermissionFunc, assignDataGroupFunc, removeDataGroupFunc)
}
