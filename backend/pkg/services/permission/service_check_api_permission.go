// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package permission

import (
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
)

func (s *service) CheckApiPermission(userID int64, method string, path string) (ok bool, err error) {
	exists, err := s.dbRepo.UserExists(userID)
	if err != nil {
		return false, err
	}

	if !exists {
		return false, core.Error(code.UserNotExistsError, "user does not exist")
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
		return false, core.Error(code.APINotExist, "api does not exist")
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
