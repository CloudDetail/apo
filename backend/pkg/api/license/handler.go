// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package license

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/cache"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/license"
	"go.uber.org/zap"
)

type Handler interface {
	// UploadLicense uploads your license, sign public key, decrypt key.
	// @Tag API.license
	// @Router /api/license/upload [post]
	UploadLicense() core.HandlerFunc
}

type handler struct {
	service license.Service
}

func New(l *zap.Logger, cacheRepo cache.Repo, promRepo prometheus.Repo) Handler {
	return &handler{
		service: license.New(l, cacheRepo, promRepo),
	}
}
