// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/cache"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/dify"
	"github.com/CloudDetail/apo/backend/pkg/services/user"
	"go.uber.org/zap"
)

type Handler interface {
	// Login Login
	// @Tags API.user
	// @Router /api/user/login [post]
	Login() core.HandlerFunc
	// Logout Logout
	// @Tags API.user
	// @Router /api/user/logout [post]
	Logout() core.HandlerFunc
	// CreateUser Create a user.
	// @Tags API.user
	// @Router /api/user/create [post]
	CreateUser() core.HandlerFunc
	// RefreshToken Refresh accessToken
	// @Tags API.user
	// @Router /api/user/refresh [get]
	RefreshToken() core.HandlerFunc

	// UpdateUserInfo Update user's info.
	// @Tags API.user
	// @Router /api/user/update/info [post]
	UpdateUserInfo() core.HandlerFunc

	// UpdateSelfInfo Update self info.
	// @Tags API.user
	// @Router /api/user/update/self [post]
	UpdateSelfInfo() core.HandlerFunc
	// UpdateUserPassword Update password.
	// @Tags API.user
	// @Router /api/user/update/password [post]
	UpdateUserPassword() core.HandlerFunc
	// UpdateUserPhone Update phone number.
	// @Tags API.user
	// @Router /api/user/update/phone [post]
	UpdateUserPhone() core.HandlerFunc
	// UpdateUserEmail Update email.
	// @Tags API.user
	// @Router /api/user/update/email [post]
	UpdateUserEmail() core.HandlerFunc
	// GetUserInfo Get user's info.
	// @Tags API.user
	// @Router /api/user/info [get]
	GetUserInfo() core.HandlerFunc

	// GetUserList Get user list.
	// @Tags API.user
	// @Router /api/user/list [get]
	GetUserList() core.HandlerFunc

	// RemoveUser Remove a user.
	// @Tags API.user
	// @Router /api/user/remove [post]
	RemoveUser() core.HandlerFunc

	// ResetPassword Reset password.
	// @Tags API.user
	// @Router /api/user/reset [post]
	ResetPassword() core.HandlerFunc

	// GetUserTeam Get user's team.
	// @Tags API.user
	// @Router /api/user/team [post]
	GetUserTeam() core.HandlerFunc
}

type handler struct {
	logger      *zap.Logger
	userService user.Service
}

func New(logger *zap.Logger, dbRepo database.Repo, cacheRepo cache.Repo, difyRepo dify.DifyRepo) Handler {
	return &handler{
		logger:      logger,
		userService: user.New(dbRepo, cacheRepo, difyRepo),
	}
}
