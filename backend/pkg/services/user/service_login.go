// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/util/jwt"
)

func (s *service) Login(ctx core.Context, req *request.LoginRequest) (response.LoginResponse, error) {
	user, err := s.dbRepo.Login(req.Username, req.Password)
	if err != nil {
		return response.LoginResponse{}, err
	}
	accessToken, refreshToken, err := jwt.GenerateTokens(user.Username, user.UserID)
	if err != nil {
		return response.LoginResponse{}, err
	}
	resp := response.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
	}
	return resp, nil
}
