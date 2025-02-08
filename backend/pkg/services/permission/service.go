// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package permission

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

type Service interface {
	GetFeature(req *request.GetFeatureRequest) (response.GetFeatureResponse, error)
	PermissionOperation(req *request.PermissionOperationRequest) error
	ConfigureMenu(req *request.ConfigureMenuRequest) error
	GetUserConfig(req *request.GetUserConfigRequest) (response.GetUserConfigResponse, error)
	GetSubjectFeature(req *request.GetSubjectFeatureRequest) (resp response.GetSubjectFeatureResponse, err error)
	CheckApiPermission(userID int64, method string, path string) (ok bool, err error)
}

type service struct {
	dbRepo database.Repo
}

func New(dbRepo database.Repo) Service {
	return &service{
		dbRepo: dbRepo,
	}
}
