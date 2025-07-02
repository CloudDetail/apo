// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

func GetGroupPQLFilter(ctx core.Context, dbRepo database.Repo, category string, groupID int64) (*prometheus.OrFilter, error) {
	// TODO
	return nil, nil
}
