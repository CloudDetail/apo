// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package team

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

func (s *service) GetTeamList(ctx_core core.Context, req *request.GetTeamRequest) (resp response.GetTeamResponse, err error) {
	teams, count, err := s.dbRepo.GetTeamList(ctx_core, req)
	if err != nil {
		return
	}

	resp = response.GetTeamResponse{
		TeamList:	teams,
		Pagination: model.Pagination{
			CurrentPage:	req.CurrentPage,
			Total:		count,
			PageSize:	req.PageSize,
		},
	}
	return
}
