// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package response

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/profile"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

type LoginResponse struct {
	AccessToken  string `json:"accessToken"`  // accessToken is used to call the interface get resources
	RefreshToken string `json:"refreshToken"` // refreshToken for refreshing accessToken
	profile.User
}

type RefreshTokenResponse struct {
	AccessToken string `json:"accessToken"` // accessToken is used to call the interface get resources
}

type GetUserInfoResponse struct {
	profile.User
}

type GetUserListResponse struct {
	Users []profile.User `json:"users"`
	model.Pagination
}

type GetRoleResponse []profile.Role

type GetUserConfigResponse struct {
	MenuItem []*database.MenuItem `json:"menuItem"`
	Routes   []string             `json:"routes"`
}

type GetFeatureResponse []*profile.Feature
type GetSubjectFeatureResponse []profile.Feature
type GetUserRoleResponse []profile.Role

type GetUserTeamResponse []profile.Team

type CheckRouterPermissionResponse bool
