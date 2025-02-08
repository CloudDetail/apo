// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package permission

import (
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/model"
)

func (s *service) CheckApiPermission(userID int64, method string, path string) (ok bool, err error) {
	exists, err := s.dbRepo.UserExists(userID)
	if err != nil {
		return false, err
	}

	if !exists {
		return false, model.NewErrWithMessage(errors.New("user does not exist"), code.UserNotExistsError)
	}

	featureIDs, err := s.getUserFeatureIDs(userID)
	if err != nil {
		return false, err
	}
	api, err := s.dbRepo.GetAPIByPath(path, method)
	if err != nil {
		return false, err
	}

	if api == nil {
		return false, model.NewErrWithMessage(errors.New("api does not exist"), code.APINotExist)
	}

	fm, err := s.dbRepo.GetFeatureMappingByMapped(api.ID, model.MAPPED_TYP_API)
	if err != nil {
		return false, err
	}

	for _, id := range featureIDs {
		if id == fm.FeatureID {
			return true, nil
		}
	}

	return false, nil
}
