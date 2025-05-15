// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetUserTeam(req *request.GetUserTeamRequest) (response.GetUserTeamResponse, error) {
	exists, err := s.dbRepo.UserExists(req.UserID)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, core.Error(code.UserNotExistsError, "user does not exist")
	}

	return s.dbRepo.GetAssignedTeam(req.UserID)
}
