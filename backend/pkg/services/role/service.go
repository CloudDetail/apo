// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package role

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

var _ Service = (*service)(nil)

type Service interface {
	RoleOperation(ctx_core core.Context, req *request.RoleOperationRequest) error
	GetRoles(ctx_core core.Context,) (response.GetRoleResponse, error)
	GetUserRole(ctx_core core.Context, req *request.GetUserRoleRequest) (response.GetUserRoleResponse, error)
	CreateRole(ctx_core core.Context, req *request.CreateRoleRequest) error
	UpdateRole(ctx_core core.Context, req *request.UpdateRoleRequest) error
	DeleteRole(ctx_core core.Context, req *request.DeleteRoleRequest) error
}

type service struct {
	dbRepo database.Repo
}

func New(dbRepo database.Repo) Service {
	return &service{
		dbRepo: dbRepo,
	}
}
