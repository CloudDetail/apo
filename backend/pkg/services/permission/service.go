// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package permission

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

type Service interface {
	GetFeature(ctx_core core.Context, req *request.GetFeatureRequest) (response.GetFeatureResponse, error)
	PermissionOperation(ctx_core core.Context, req *request.PermissionOperationRequest) error
	ConfigureMenu(ctx_core core.Context, req *request.ConfigureMenuRequest) error
	GetUserConfig(ctx_core core.Context, req *request.GetUserConfigRequest) (response.GetUserConfigResponse, error)
	GetSubjectFeature(ctx_core core.Context, req *request.GetSubjectFeatureRequest) (resp response.GetSubjectFeatureResponse, err error)
	CheckApiPermission(ctx_core core.Context, userID int64, method string, path string) (ok bool, err error)
	CheckRouterPermission(ctx_core core.Context, userID int64, router string) (bool, error)
}

type service struct {
	dbRepo database.Repo
}

func New(dbRepo database.Repo) Service {
	return &service{
		dbRepo: dbRepo,
	}
}
