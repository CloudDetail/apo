// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"slices"

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
	if slices.Contains(premGroups, 0) {
		return true, nil
	}
	return s.dbRepo.CheckScopePermission(ctx, premGroups, cluster, namespace, service)
}

func (s *service) CheckServicesPermission(ctx core.Context, services ...string) (bool, error) {
	userID := ctx.UserID()
	premGroups, err := s.dbRepo.GetDataGroupIDsByUserId(ctx, userID)
	if err != nil {
		return false, err
	}
	if slices.Contains(premGroups, 0) {
		return true, nil
	}

	selected, err := s.dbRepo.GetScopeIDsSelectedByPermGroupIDs(ctx, premGroups)
	if err != nil {
		return false, err
	}

	svcList := common.DataGroupStorage.GetFullPermissionSvcList(selected)

	for _, service := range services {
		if !slices.Contains(svcList, service) {
			return false, nil
		}
	}
	return true, nil
}

func (s *service) CheckGroupPermission(ctx core.Context, groupID int64) (bool, error) {
	userID := ctx.UserID()
	premGroups, err := s.dbRepo.GetDataGroupIDsByUserId(ctx, userID)

	if slices.Contains(premGroups, 0) {
		return true, nil
	}

	if err != nil {
		return false, err
	}
	return common.DataGroupStorage.CheckGroupPermission(groupID, premGroups), nil
}
