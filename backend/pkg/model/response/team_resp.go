// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package response

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

type GetTeamResponse struct {
	TeamList         []database.Team `json:"teamList"`
	model.Pagination `json:",inline"`
}

type GetTeamUserResponse []database.User
