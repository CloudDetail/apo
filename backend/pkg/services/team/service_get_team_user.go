// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package team

import (
	"github.com/CloudDetail/apo/backend/pkg/code"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetTeamUser(ctx_core core.Context, req *request.GetTeamUserRequest) (resp response.GetTeamUserResponse, err error) {
	filter := model.TeamFilter{
		ID: req.TeamID,
	}
	exists, err := s.dbRepo.TeamExist(ctx_core, filter)
	if err != nil {
		return
	}

	if !exists {
		err = core.Error(code.TeamNotExistError, "team does not exist")
		return
	}

	users, err := s.dbRepo.GetTeamUserList(ctx_core, req.TeamID)
	if err != nil {
		return
	}
	resp = users
	return
}
