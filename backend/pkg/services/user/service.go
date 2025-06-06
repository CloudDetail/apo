// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/cache"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/dify"
)

var _ Service = (*service)(nil)

type Service interface {
	Login(ctx core.Context, req *request.LoginRequest) (response.LoginResponse, error)
	Logout(ctx core.Context, req *request.LogoutRequest) error
	CreateUser(ctx core.Context, req *request.CreateUserRequest) error
	RefreshToken(ctx core.Context, token string) (response.RefreshTokenResponse, error)
	UpdateUserInfo(ctx core.Context, req *request.UpdateUserInfoRequest) error
	UpdateSelfInfo(ctx core.Context, req *request.UpdateSelfInfoRequest) error
	UpdateUserPhone(ctx core.Context, req *request.UpdateUserPhoneRequest) error
	UpdateUserEmail(ctx core.Context, req *request.UpdateUserEmailRequest) error
	UpdateUserPassword(ctx core.Context, req *request.UpdateUserPasswordRequest) error
	GetUserInfo(ctx core.Context, userID int64) (response.GetUserInfoResponse, error)
	GetUserList(ctx core.Context, req *request.GetUserListRequest) (response.GetUserListResponse, error)
	RemoveUser(ctx core.Context, userID int64) error
	RestPassword(ctx core.Context, req *request.ResetPasswordRequest) error

	GetUserTeam(ctx core.Context, req *request.GetUserTeamRequest) (response.GetUserTeamResponse, error)

	IsInBlacklist(ctx core.Context, token string) (bool, error)
}

type service struct {
	dbRepo    database.Repo
	cacheRepo cache.Repo
	difyRepo  dify.DifyRepo
}

func New(dbRepo database.Repo, cacheRepo cache.Repo, difyRepo dify.DifyRepo) Service {
	return &service{
		dbRepo:    dbRepo,
		cacheRepo: cacheRepo,
		difyRepo:  difyRepo,
	}
}
