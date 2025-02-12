// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package license

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/cache"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"go.uber.org/zap"
)

type Service interface {
	UploadLicense(req *request.UploadLicenseRequest) error
}

type service struct {
	cacheRepo cache.Repo
	promRepo prometheus.Repo
	logger    *zap.Logger
}

func New(l *zap.Logger, cacheRepo cache.Repo, promRepo prometheus.Repo) Service {
	s := &service{
		cacheRepo: cacheRepo,
		logger:    l,
		promRepo: promRepo,
	}

	go s.checkLoop()
	return s
}
