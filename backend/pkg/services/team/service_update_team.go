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

func (s *service) UpdateTeam(ctx_core core.Context, req *request.UpdateTeamRequest) error {
	team, err := s.dbRepo.GetTeam(ctx_core, req.TeamID)
	if err != nil {
		return err
	}

	if team.TeamID == 0 {
		return core.Error(code.TeamAlreadyExistError, "team does not exist")
	}

	if team.TeamName != req.TeamName {
		filter := model.TeamFilter{
			Name: req.TeamName,
		}

		exists, err := s.dbRepo.TeamExist(ctx_core, filter)
		if err != nil {
			return err
		}

		if exists {
			return core.Error(code.TeamAlreadyExistError, "team already existed")
		}
	}

	team.TeamName = req.TeamName
	team.Description = req.Description

	// toAddFeature, toDeleteFeature, err := s.dbRepo.GetAddAndDeletePermissions(req.TeamID, model.PERMISSION_SUB_TYP_TEAM, model.PERMISSION_TYP_FEATURE, req.FeatureList)
	// if err != nil {
	// 	return err
	// }

	// toModifyDg, toDeleteDg, err := s.dbRepo.GetModifyAndDeleteDataGroup(req.TeamID, model.DATA_GROUP_SUB_TYP_TEAM, req.DataGroupPermissions)
	// if err != nil {
	// 	return err
	// }

	// determine added or removed users
	hasUsers, err := s.dbRepo.GetTeamUsers(ctx_core, req.TeamID)
	if err != nil {
		return err
	}

	hasUserMap := make(map[int64]struct{})
	for _, id := range hasUsers {
		hasUserMap[id] = struct{}{}
	}

	var toAdd, toDelete []int64
	for _, id := range req.UserList {
		if _, ok := hasUserMap[id]; !ok {
			toAdd = append(toAdd, id)
		} else {
			delete(hasUserMap, id)
		}
	}

	for id := range hasUserMap {
		toDelete = append(toDelete, id)
	}

	var inviteFunc = func(ctx context.Context) error {
		return s.dbRepo.InviteUserToTeam(ctx_core, ctx, req.TeamID, toAdd)
	}

	var removeFunc = func(ctx context.Context) error {
		return s.dbRepo.RemoveFromTeamByTeam(ctx_core, ctx, req.TeamID, toDelete)
	}

	var updateTeamFunc = func(ctx context.Context) error {
		return s.dbRepo.UpdateTeam(ctx_core, ctx, team)
	}

	// var grantPermissionFunc = func(ctx context.Context) error {
	// 	return s.dbRepo.GrantPermission(ctx, req.TeamID, model.PERMISSION_SUB_TYP_TEAM, model.PERMISSION_TYP_FEATURE, toAddFeature)
	// }

	// var revokePermissionFunc = func(ctx context.Context) error {
	// 	return s.dbRepo.RevokePermission(ctx, req.TeamID, model.PERMISSION_SUB_TYP_TEAM, model.PERMISSION_TYP_FEATURE, toDeleteFeature)
	// }

	// var assignDataGroupFunc = func(ctx context.Context) error {
	// 	return s.dbRepo.AssignDataGroup(ctx, toModifyDg)
	// }

	// var removeDataGroupFunc = func(ctx context.Context) error {
	// 	return s.dbRepo.RevokeDataGroupByGroup(ctx, toDeleteDg, req.TeamID)
	// }

	return s.dbRepo.Transaction(ctx_core, context.Background(), updateTeamFunc, inviteFunc, removeFunc)
}
