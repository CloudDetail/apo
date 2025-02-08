// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetGroupSubs(req *request.GetGroupSubsRequest) (response.GetGroupSubsResponse, error) {
	var (
		resp response.GetGroupSubsResponse
		err  error
	)

	switch req.SubjectType {
	case model.DATA_GROUP_SUB_TYP_USER:
		resp, err = s.dbRepo.GetDataGroupUsers(req.DataGroupID)
	case model.DATA_GROUP_SUB_TYP_TEAM:
		resp, err = s.dbRepo.GetDataGroupTeams(req.DataGroupID)
	case "":
		resp, err = s.dbRepo.GetDataGroupUsers(req.DataGroupID)
		if err != nil {
			return nil, err
		}
		authTeam, err := s.dbRepo.GetDataGroupTeams(req.DataGroupID)
		if err != nil {
			return nil, err
		}
		resp = append(resp, authTeam...)
	}

	return resp, err
}
