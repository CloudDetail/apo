// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/cache"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/services/user"
	"go.uber.org/zap"
)

type Handler interface {
	// Login
	// @Tags API.user
	// @Router /api/user/login [post]
	Login() core.HandlerFunc
	// Logout Logout
	// @Tags API.user
	// @Router /api/user/logout [post]
	Logout() core.HandlerFunc
	// Create user CreateUser
	// @Tags API.user
	// @Router /api/user/create [post]
	CreateUser() core.HandlerFunc
	// RefreshToken refresh accessToken
	// @Tags API.user
	// @Router /api/user/refresh [get]
	RefreshToken() core.HandlerFunc
	// UpdateUserInfo update personal information
	// @Tags API.user
	// @Router /api/user/update/info [post]
	UpdateUserInfo() core.HandlerFunc
	// UpdateUserPassword update password
	// @Tags API.user
	// @Router /api/user/update/password [post]
	UpdateUserPassword() core.HandlerFunc
	// UpdateUserPhone update/bind phone number
	// @Tags API.user
	// @Router /api/user/update/phone [post]
	UpdateUserPhone() core.HandlerFunc
	// UpdateUserEmail update/bind mailbox
	// @Tags API.user
	// @Router /api/user/update/email [post]
	UpdateUserEmail() core.HandlerFunc
	// GetUserInfo access to personal information
	// @Tags API.user
	// @Router /api/user/info [get]
	GetUserInfo() core.HandlerFunc

	// GetUserList get user list
	// @Tags API.user
	// @Router /api/user/list [get]
	GetUserList() core.HandlerFunc

	// Remove user RemoveUser
	// @Tags API.user
	// @Router /api/user/remove [post]
	RemoveUser() core.HandlerFunc

	// ResetPassword reset password
	// @Tags API.user
	// @Router /api/user/reset [post]
	ResetPassword() core.HandlerFunc

	// RoleOperation Grant or revoke user's role.
	// @Tags API.permission
	// @Router /api/permission/role/operation [post]
	RoleOperation() core.HandlerFunc

	// GetRole Gets all roles.
	// @Tags API.permission
	// @Router /api/permission/roles [get]
	GetRole() core.HandlerFunc

	// GetUserRole Get user's role.
	// @Tags API.permission
	// @Router /api/permission/role [get]
	GetUserRole() core.HandlerFunc

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
	logger      *zap.Logger
	userService user.Service
}

func New(logger *zap.Logger, dbRepo database.Repo, cacheRepo cache.Repo) Handler {
	return &handler{
		logger:      logger,
		userService: user.New(dbRepo, cacheRepo),
	}
}
