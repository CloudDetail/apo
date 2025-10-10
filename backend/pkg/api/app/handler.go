// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/services/app"
)

type Handler interface {
	// UploadLogo uploads logo
	// @Tags API.app
	// @Router /api/app/upload [post]
	UploadLogo() core.HandlerFunc

	// GetLogo gets logo
	// @Tags API.app
	// @Router /api/app/logo [get]
	GetLogo() core.HandlerFunc
}

type handler struct {
	service app.Service
}

func New(dbRepo database.Repo) Handler {
	return &handler{
		service: app.New(dbRepo),
	}
}
