// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package team

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/services/team"
	"go.uber.org/zap"
)

type Handler interface {
	// CreateTeam Creates a team.
	// @Tags API.team
	// @Router /api/team/create [post]
	CreateTeam() core.HandlerFunc

	// UpdateTeam Update team's information.
	// @Tags API.team
	// @Router /api/team/update [post]
	UpdateTeam() core.HandlerFunc

	// GetTeam Get teams.
	// @Tags API.team
	// @Router /api/team [get]
	GetTeam() core.HandlerFunc

	// DeleteTeam Delete a team.
	// @Tags API.team
	// @Router /api/team/delete [post]
	DeleteTeam() core.HandlerFunc

	// TeamOperation Assigns a user to teams or removes a user from teams.
	// @Tags API.team
	// @Router /api/team/operation [post]
	TeamOperation() core.HandlerFunc

	// TeamUserOperation Assigns users to a team or remove user from a team.
	// @Tags API.team
	// @Router /api/team/user/operation [post]
	TeamUserOperation() core.HandlerFunc

	// GetTeamUser Get team's users.
	// @Tags API.team
	// @Router /api/team/user [get]
	GetTeamUser() core.HandlerFunc
}

type handler struct {
	logger      *zap.Logger
	teamService team.Service
}

func New(logger *zap.Logger, dbRepo database.Repo) Handler {
	return &handler{
		logger:      logger,
		teamService: team.New(dbRepo),
	}
}
