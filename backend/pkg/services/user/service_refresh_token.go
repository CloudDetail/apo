// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/util/jwt"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

func (s *service) RefreshToken(ctx_core core.Context, token string) (response.RefreshTokenResponse, error) {
	accessToken, err := jwt.RefreshToken(token)
	var resp response.RefreshTokenResponse
	if err != nil {
		return resp, err
	}
	resp.AccessToken = accessToken
	return resp, nil
}

func (s *service) IsInBlacklist(ctx_core core.Context, token string) (bool, error) {
	return s.cacheRepo.IsInBlacklist(token)
}
