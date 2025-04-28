// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetGroupSubs(ctx core.Context, req *request.GetGroupSubsRequest) (response.GetGroupSubsResponse, error) {
	var (
		resp response.GetGroupSubsResponse
		err  error
	)

	switch req.SubjectType {
	case model.DATA_GROUP_SUB_TYP_USER:
		resp, err = s.dbRepo.GetDataGroupUsers(ctx, req.DataGroupID)
	case model.DATA_GROUP_SUB_TYP_TEAM:
		resp, err = s.dbRepo.GetDataGroupTeams(ctx, req.DataGroupID)
	case "":
		resp, err = s.dbRepo.GetDataGroupUsers(ctx, req.DataGroupID)
		if err != nil {
			return nil, err
		}
		authTeam, err := s.dbRepo.GetDataGroupTeams(ctx, req.DataGroupID)
		if err != nil {
			return nil, err
		}
		resp = append(resp, authTeam...)
	}

	return resp, err
}
