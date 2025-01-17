// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) Logout(req *request.LogoutRequest) error {
	err := s.cacheRepo.AddToken(req.AccessToken)
	if err != nil {
		return err
	}

	err = s.cacheRepo.AddToken(req.RefreshToken)
	if err != nil {
		return err
	}
	return nil
}
