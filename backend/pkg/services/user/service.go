// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/cache"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/dify"
)

var _ Service = (*service)(nil)

type Service interface {
	Login(req *request.LoginRequest) (response.LoginResponse, error)
	Logout(req *request.LogoutRequest) error
	CreateUser(req *request.CreateUserRequest) error
	RefreshToken(token string) (response.RefreshTokenResponse, error)
	UpdateUserInfo(req *request.UpdateUserInfoRequest) error
	UpdateSelfInfo(req *request.UpdateSelfInfoRequest) error
	UpdateUserPhone(req *request.UpdateUserPhoneRequest) error
	UpdateUserEmail(req *request.UpdateUserEmailRequest) error
	UpdateUserPassword(req *request.UpdateUserPasswordRequest) error
	GetUserInfo(userID int64) (response.GetUserInfoResponse, error)
	GetUserList(req *request.GetUserListRequest) (response.GetUserListResponse, error)
	RemoveUser(userID int64) error
	RestPassword(req *request.ResetPasswordRequest) error

	GetUserTeam(req *request.GetUserTeamRequest) (response.GetUserTeamResponse, error)

	IsInBlacklist(token string) (bool, error)
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
