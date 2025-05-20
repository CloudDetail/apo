// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) Logout(ctx core.Context, req *request.LogoutRequest) error {
	err := s.cacheRepo.AddToken(ctx, req.AccessToken)
	if err != nil {
		return err
	}

	err = s.cacheRepo.AddToken(ctx, req.RefreshToken)
	if err != nil {
		return err
	}
	return nil
}
