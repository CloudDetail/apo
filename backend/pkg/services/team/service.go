// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package team

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

type Service interface {
	CreateTeam(req *request.CreateTeamRequest) error
	UpdateTeam(req *request.UpdateTeamRequest) error
	GetTeamList(ctx core.Context, req *request.GetTeamRequest) (resp response.GetTeamResponse, err error)
	DeleteTeam(req *request.DeleteTeamRequest) error
	TeamOperation(ctx core.Context, req *request.TeamOperationRequest) error
	TeamUserOperation(req *request.AssignToTeamRequest) error
	GetTeamUser(req *request.GetTeamUserRequest) (response.GetTeamUserResponse, error)
}

type service struct {
	dbRepo database.Repo
}

func New(dbRepo database.Repo) Service {
	return &service{
		dbRepo: dbRepo,
	}
}
