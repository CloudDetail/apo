// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/cache"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/services/user"
)

var _ Middleware = (*middleware)(nil)

type Middleware interface {
	AuthMiddleware() core.HandlerFunc
}

type middleware struct {
	userService user.Service
}

func New(cacheRepo cache.Repo, dbRepo database.Repo) Middleware {
	return &middleware{
		userService: user.New(dbRepo, cacheRepo),
	}
}
