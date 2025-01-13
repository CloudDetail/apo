// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package response

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

type LoginResponse struct {
	AccessToken  string `json:"accessToken"`  // accessToken is used to call the interface get resources
	RefreshToken string `json:"refreshToken"` // refreshToken for refreshing accessToken
	database.User
}

type RefreshTokenResponse struct {
	AccessToken string `json:"accessToken"` // accessToken is used to call the interface get resources
}

type GetUserInfoResponse struct {
	database.User
}

type GetUserListResponse struct {
	Users []database.User `json:"users"`
	model.Pagination
}

type GetRoleResponse []database.Role

type GetUserConfigResponse struct {
	MenuItem []*database.MenuItem `json:"menuItem"`
	Routes   []string             `json:"routes"`
}

type GetFeatureResponse []*database.Feature
type GetSubjectFeatureResponse []database.Feature
type GetUserRoleResponse []database.Role
