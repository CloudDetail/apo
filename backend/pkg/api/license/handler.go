// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package license

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/license"
	"go.uber.org/zap"
)

type Handler interface {
	// UploadLicense uploads your license, sign public key, decrypt key.
	// @Tags API.license
	// @Router /api/license/upload [post]
	UploadLicense() core.HandlerFunc

	// GetLicenses get all license.
	// @Tags API.license
	// @Router /api/license [get]
	GetLicenses() core.HandlerFunc

	// GetCurrentLicense get current license.
	// @Tags API.license
	// @Router /api/license/current [get]
	GetCurrentLicense() core.HandlerFunc
}

type handler struct {
	service license.Service
}

func New(l *zap.Logger, promRepo prometheus.Repo, dbRepo database.Repo) Handler {
	return &handler{
		service: license.New(l, promRepo, dbRepo),
	}
}
