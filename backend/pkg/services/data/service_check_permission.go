// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/services/common"
)

func (s *service) CheckScopePermission(ctx core.Context, cluster, namespace, service string) (bool, error) {
	if cluster == "" && namespace == "" && service == "" {
		return true, nil
	}

	userID := ctx.UserID()
	premGroups, err := s.dbRepo.GetDataGroupIDsByUserId(ctx, userID)
	if err != nil {
		return false, err
	}
	if containsInInt(premGroups, 0) {
		return true, nil
	}
	return s.dbRepo.CheckScopePermission(ctx, premGroups, cluster, namespace, service)
}

func (s *service) CheckGroupPermission(ctx core.Context, groupID int64) (bool, error) {
	userID := ctx.UserID()
	premGroups, err := s.dbRepo.GetDataGroupIDsByUserId(ctx, userID)

	if containsInInt(premGroups, 0) {
		return true, nil
	}

	if err != nil {
		return false, err
	}
	return common.DataGroupStorage.CheckGroupPermission(groupID, premGroups), nil
}
