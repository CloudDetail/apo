// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

func (s *service) CheckDatasourcePermission(ctx core.Context, userID, groupID int64, namespaces, services interface{}, fillCategory string) (err error) {
	// TODO
	return nil
}
