// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/services/common"
)

func (s *service) CheckDatasourcePermission(ctx core.Context, userID, groupID int64, namespaces, services interface{}, fillCategory string) (err error) {
	// TODO 调整到CheckScopePermission 和 CheckGroupPermission
	return nil
}

// 只有未使用GroupID过滤的接口需要检查入参权限
// WARN 尽量减少未使用数据组过滤的接口
func (s *service) CheckScopePermission(ctx core.Context, cluster, namespace, service string) (bool, error) {
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

// 对于使用了GroupID 过滤的接口，仅检查用户对数据组的权限即可
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
