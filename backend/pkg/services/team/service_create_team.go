// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package team

import (
	"github.com/CloudDetail/apo/backend/pkg/code"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/profile"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/util"
)

func (s *service) CreateTeam(ctx core.Context, req *request.CreateTeamRequest) error {
	team := profile.Team{
		TeamID:      util.Generator.GenerateID(),
		TeamName:    req.TeamName,
		Description: req.Description,
	}

	filter := model.TeamFilter{
		Name: req.TeamName,
	}
	exists, err := s.dbRepo.TeamExist(ctx, filter)
	if err != nil {
		return err
	}

	if exists {
		return core.Error(code.TeamAlreadyExistError, "team already existed")
	}
	if len(req.FeatureList) > 0 {
		features, err := s.dbRepo.GetFeature(ctx, req.FeatureList)
		if err != nil {
			return err
		}

		if len(features) != len(req.FeatureList) {
			return core.Error(code.PermissionNotExistError, "permission does not exist")
		}
	}

	if len(req.DataGroupPermissions) > 0 {
		filter := model.DataGroupFilter{}
		for _, dgPermission := range req.DataGroupPermissions {
			filter.IDs = append(filter.IDs, dgPermission.DataGroupID)
		}

		exist, err := s.dbRepo.DataGroupExist(ctx, filter)
		if err != nil {
			return err
		}

		if !exist {
			return core.Error(code.DataGroupNotExistError, "data group does not exist")
		}
	}

	exist, err := s.dbRepo.UserExists(ctx, req.UserList...)
	if err != nil {
		return err
	}

	if !exist {
		return core.Error(code.UserNotExistsError, "user does not exist")
	}

	authDataGroup := make([]database.AuthDataGroup, len(req.DataGroupPermissions))
	for i, dgPermission := range req.DataGroupPermissions {
		authDataGroup[i] = database.AuthDataGroup{
			GroupID:     dgPermission.DataGroupID,
			SubjectType: model.DATA_GROUP_SUB_TYP_TEAM,
			SubjectID:   team.TeamID,
			Type:        dgPermission.PermissionType,
		}
	}

	var assignDataGroupFunc = func(ctx core.Context) error {
		return s.dbRepo.AssignDataGroup(ctx, authDataGroup)
	}

	var createTeamFunc = func(ctx core.Context) error {
		return s.dbRepo.CreateTeam(ctx, team)
	}

	var grantPermissionFunc = func(ctx core.Context) error {
		return s.dbRepo.GrantPermission(ctx, team.TeamID, model.PERMISSION_SUB_TYP_TEAM, model.PERMISSION_TYP_FEATURE, req.FeatureList)
	}

	var inviteUserFunc = func(ctx core.Context) error {
		return s.dbRepo.InviteUserToTeam(ctx, team.TeamID, req.UserList)
	}

	return s.dbRepo.Transaction(ctx, createTeamFunc, grantPermissionFunc, assignDataGroupFunc, inviteUserFunc)
}
