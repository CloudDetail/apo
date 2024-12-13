package response

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

type LoginResponse struct {
	AccessToken  string `json:"accessToken"`  // accessToken用于调用接口获取资源
	RefreshToken string `json:"refreshToken"` // refreshToken用于刷新accessToken
	database.User
}

type RefreshTokenResponse struct {
	AccessToken string `json:"accessToken"` // accessToken用于调用接口获取资源
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
