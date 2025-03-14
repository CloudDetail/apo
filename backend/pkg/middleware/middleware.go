// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/cache"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/dify"
	"github.com/CloudDetail/apo/backend/pkg/services/permission"
	"github.com/CloudDetail/apo/backend/pkg/services/user"
)

var _ Middleware = (*middleware)(nil)

type Middleware interface {
	AuthMiddleware() core.HandlerFunc
	PermissionMiddleware() core.HandlerFunc
}

type middleware struct {
	userService       user.Service
	permissionService permission.Service
}

func New(cacheRepo cache.Repo, dbRepo database.Repo, difyRepo dify.DifyRepo) Middleware {
	return &middleware{
		userService:       user.New(dbRepo, cacheRepo, difyRepo),
		permissionService: permission.New(dbRepo),
	}
}
