// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package role

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

var _ Service = (*service)(nil)

type Service interface {
	RoleOperation(ctx core.Context, req *request.RoleOperationRequest) error
	GetRoles(ctx core.Context) (response.GetRoleResponse, error)
	GetUserRole(ctx core.Context, req *request.GetUserRoleRequest) (response.GetUserRoleResponse, error)
	CreateRole(ctx core.Context, req *request.CreateRoleRequest) error
	UpdateRole(ctx core.Context, req *request.UpdateRoleRequest) error
	DeleteRole(ctx core.Context, req *request.DeleteRoleRequest) error
}

type service struct {
	dbRepo database.Repo
}

func New(dbRepo database.Repo) Service {
	return &service{
		dbRepo: dbRepo,
	}
}
