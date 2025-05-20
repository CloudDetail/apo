// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package team

import (
	"github.com/CloudDetail/apo/backend/pkg/code"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) TeamUserOperation(ctx core.Context, req *request.AssignToTeamRequest) error {
	filter := model.TeamFilter{
		ID: req.TeamID,
	}
	exists, err := s.dbRepo.TeamExist(ctx, filter)
	if err != nil {
		return err
	}

	if !exists {
		return core.Error(code.TeamNotExistError, "team does not exist")
	}

	exists, err = s.dbRepo.UserExists(ctx, req.UserList...)
	if err != nil {
		return err
	}

	if !exists {
		return core.Error(code.UserNotExistsError, "user does not exist")
	}

	hasUsers, err := s.dbRepo.GetTeamUsers(ctx, req.TeamID)
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

	var inviteFunc = func(ctx core.Context) error {
		return s.dbRepo.InviteUserToTeam(ctx, req.TeamID, toAdd)
	}

	var removeFunc = func(ctx core.Context) error {
		return s.dbRepo.RemoveFromTeamByTeam(ctx, req.TeamID, toDelete)
	}

	return s.dbRepo.Transaction(ctx, inviteFunc, removeFunc)
}
