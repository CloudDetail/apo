// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/util/jwt"
)

func (s *service) RefreshToken(token string) (response.RefreshTokenResponse, error) {
	accessToken, err := jwt.RefreshToken(token)
	var resp response.RefreshTokenResponse
	if err != nil {
		return resp, err
	}
	resp.AccessToken = accessToken
	return resp, nil
}

func (s *service) IsInBlacklist(token string) (bool, error) {
	return s.cacheRepo.IsInBlacklist(token)
}
