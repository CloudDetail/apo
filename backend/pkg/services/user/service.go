// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/cache"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/dify"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

var _ Service = (*service)(nil)

type Service interface {
	Login(ctx_core core.Context, req *request.LoginRequest) (response.LoginResponse, error)
	Logout(ctx_core core.Context, req *request.LogoutRequest) error
	CreateUser(ctx_core core.Context, req *request.CreateUserRequest) error
	RefreshToken(ctx_core core.Context, token string) (response.RefreshTokenResponse, error)
	UpdateUserInfo(ctx_core core.Context, req *request.UpdateUserInfoRequest) error
	UpdateSelfInfo(ctx_core core.Context, req *request.UpdateSelfInfoRequest) error
	UpdateUserPhone(ctx_core core.Context, req *request.UpdateUserPhoneRequest) error
	UpdateUserEmail(ctx_core core.Context, req *request.UpdateUserEmailRequest) error
	UpdateUserPassword(ctx_core core.Context, req *request.UpdateUserPasswordRequest) error
	GetUserInfo(ctx_core core.Context, userID int64) (response.GetUserInfoResponse, error)
	GetUserList(ctx_core core.Context, req *request.GetUserListRequest) (response.GetUserListResponse, error)
	RemoveUser(ctx_core core.Context, userID int64) error
	RestPassword(ctx_core core.Context, req *request.ResetPasswordRequest) error

	GetUserTeam(ctx_core core.Context, req *request.GetUserTeamRequest) (response.GetUserTeamResponse, error)

	IsInBlacklist(ctx_core core.Context, token string) (bool, error)
}

type service struct {
	dbRepo		database.Repo
	cacheRepo	cache.Repo
	difyRepo	dify.DifyRepo
}

func New(dbRepo database.Repo, cacheRepo cache.Repo, difyRepo dify.DifyRepo) Service {
	return &service{
		dbRepo:		dbRepo,
		cacheRepo:	cacheRepo,
		difyRepo:	difyRepo,
	}
}
