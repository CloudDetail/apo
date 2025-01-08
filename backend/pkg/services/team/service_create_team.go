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
	"github.com/CloudDetail/apo/backend/pkg/util"
)

func (s *service) CreateTeam(req *request.CreateTeamRequest) error {
	team := database.Team{
		TeamID:      util.Generator.GenerateID(),
		TeamName:    req.TeamName,
		Description: req.Description,
	}

	if len(req.FeatureList) > 0 {
		features, err := s.dbRepo.GetFeature(req.FeatureList)
		if err != nil {
			return err
		}

		if len(features) != len(req.FeatureList) {
			return model.NewErrWithMessage(errors.New("permission does not exist"), code.PermissionNotExistError)
		}
	}

	filter := model.DataGroupFilter{}
	for _, dgPermission := range req.DataGroupPermissions {
		filter.IDs = append(filter.IDs, dgPermission.DataGroupID)
	}

	exist, err := s.dbRepo.DataGroupExist(filter)
	if err != nil {
		return err
	}

	if !exist {
		return model.NewErrWithMessage(errors.New("data group does not exist"), code.DataGroupNotExistError)
	}

	exist, err = s.dbRepo.UserExists(req.UserList...)
	if err != nil {
		return err
	}

	if !exist {
		return model.NewErrWithMessage(errors.New("user does not exist"), code.UserNotExistsError)
	}

	authDataGroup := make([]database.AuthDataGroup, len(req.DataGroupPermissions))
	for i, dgPermission := range req.DataGroupPermissions {
		authDataGroup[i] = database.AuthDataGroup{
			DataGroupID: dgPermission.DataGroupID,
			SubjectType: model.DATA_GROUP_SUB_TYP_TEAM,
			SubjectID:   team.TeamID,
			Type:        dgPermission.PermissionType,
		}
	}

	var assignDataGroupFunc = func(ctx context.Context) error {
		return s.dbRepo.AssignDataGroup(ctx, authDataGroup)
	}

	var createTeamFunc = func(ctx context.Context) error {
		return s.dbRepo.CreateTeam(ctx, team)
	}

	var grantPermissionFunc = func(ctx context.Context) error {
		return s.dbRepo.GrantPermission(ctx, team.TeamID, model.PERMISSION_SUB_TYP_TEAM, model.PERMISSION_TYP_FEATURE, req.FeatureList)
	}

	var inviteUserFunc = func(ctx context.Context) error {
		return s.dbRepo.InviteUserToTeam(ctx, team.TeamID, req.UserList)
	}

	return s.dbRepo.Transaction(context.Background(), createTeamFunc, grantPermissionFunc, assignDataGroupFunc, inviteUserFunc)
}
