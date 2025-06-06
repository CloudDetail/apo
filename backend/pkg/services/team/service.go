// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package team

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

type Service interface {
	CreateTeam(ctx core.Context, req *request.CreateTeamRequest) error
	UpdateTeam(ctx core.Context, req *request.UpdateTeamRequest) error
	GetTeamList(ctx core.Context, req *request.GetTeamRequest) (resp response.GetTeamResponse, err error)
	DeleteTeam(ctx core.Context, req *request.DeleteTeamRequest) error
	TeamOperation(ctx core.Context, req *request.TeamOperationRequest) error
	TeamUserOperation(ctx core.Context, req *request.AssignToTeamRequest) error
	GetTeamUser(ctx core.Context, req *request.GetTeamUserRequest) (response.GetTeamUserResponse, error)
}

type service struct {
	dbRepo database.Repo
}

func New(dbRepo database.Repo) Service {
	return &service{
		dbRepo: dbRepo,
	}
}
