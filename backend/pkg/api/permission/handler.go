// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package permission

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/services/permission"
	"go.uber.org/zap"
)

type Handler interface {
	// GetUserConfig Gets user's menu config and which route can access.
	// @Tags API.permission
	// @Router /api/permission/config [get]
	GetUserConfig() core.HandlerFunc

	// GetFeature Gets all feature permission.
	// @Tags API.permission
	// @Router /api/permission/feature [get]
	GetFeature() core.HandlerFunc

	// GetSubjectFeature Gets subject's feature permission.
	// @Tags API.permission
	// @Router /api/permission/sub/feature [get]
	GetSubjectFeature() core.HandlerFunc

	// PermissionOperation Grant or revoke user's permission(feature).
	// @Tags API.permission
	// @Router /api/permission/operation [post]
	PermissionOperation() core.HandlerFunc

	// ConfigureMenu Configure global menu.
	// @Tags API.permission
	// @Router /api/permission/menu/configure [post]
	ConfigureMenu() core.HandlerFunc
}

type handler struct {
	logger            *zap.Logger
	permissionService permission.Service
}

func New(logger *zap.Logger, dbRepo database.Repo) Handler {
	return &handler{
		logger:            logger,
		permissionService: permission.New(dbRepo),
	}
}
