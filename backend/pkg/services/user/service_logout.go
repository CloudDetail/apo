// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/util"
)

func (s *service) Logout(req *request.LogoutRequest) error {
	_, err := util.ParseAccessToken(req.AccessToken)
	if errors.Is(err, util.TokenExpired) {
		return nil
	} else if err != nil {
		return err
	} else {
		err := s.cacheRepo.AddToken(req.AccessToken)
		if err != nil {
			return err
		}
	}

	_, err = util.ParseRefreshToken(req.RefreshToken)
	if errors.Is(err, util.TokenExpired) {
		return nil
	} else if err != nil {
		return err
	} else {
		err := s.cacheRepo.AddToken(req.RefreshToken)
		if err != nil {
			return err
		}
	}
	return nil
}
