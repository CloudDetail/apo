// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/dify"
)

func (s *service) LoginDify(ctx core.Context, userID int64) (*dify.DifyLoginResponse, error) {
	user, err := s.GetUserInfo(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.difyRepo.LoginByAPOSession(user.Username)
}
