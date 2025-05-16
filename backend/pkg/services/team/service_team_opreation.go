// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package team

import (
	"context"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) TeamOperation(req *request.TeamOperationRequest) error {
	teamIDs, err := s.dbRepo.GetUserTeams(req.UserID)
	if err != nil {
		return err
	}

	teams, _, err := s.dbRepo.GetTeamList(&request.GetTeamRequest{})
	if err != nil {
		return err
	}

	teamMap := make(map[int64]struct{})
	for _, team := range teams {
		teamMap[team.TeamID] = struct{}{}
	}

	uTeamMap := make(map[int64]struct{})
	for _, id := range teamIDs {
		uTeamMap[id] = struct{}{}
	}

	var toAdd, toDelete []int64
	for _, id := range req.TeamList {
		if _, ok := teamMap[id]; !ok {
			return core.Error(code.TeamNotExistError, "team does not exist")
		}

		if _, ok := uTeamMap[id]; !ok {
			toAdd = append(toAdd, id)
		} else {
			delete(uTeamMap, id)
		}
	}

	for id := range uTeamMap {
		toDelete = append(toDelete, id)
	}

	var assignFunc = func(ctx context.Context) error {
		return s.dbRepo.AssignUserToTeam(ctx, req.UserID, toAdd)
	}

	var removeFunc = func(ctx context.Context) error {
		return s.dbRepo.RemoveFromTeamByUser(ctx, req.UserID, toDelete)
	}

	return s.dbRepo.Transaction(context.Background(), assignFunc, removeFunc)
}
