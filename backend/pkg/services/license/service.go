// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package license

import (
	mylicense "github.com/CloudDetail/apo/backend/pkg/model/license"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"go.uber.org/zap"
)

type Service interface {
	UploadLicense(req *request.UploadLicenseRequest) error
	GetLicenses() (response.GetLicenseResp, error)
	GetCurrentLicense() (response.GetCurrentLicenseResp, error)
	CheckFeature(features ...string) (bool, error)
	CheckLicenseNodes() (exceed bool, err error)
	CheckLicenseDate() (expire bool, err error)
	GetActivation() (mylicense.Activation, error)
}

type service struct {
	promRepo prometheus.Repo
	logger   *zap.Logger
	dbRepo   database.Repo
}

func New(l *zap.Logger, promRepo prometheus.Repo, dbRepo database.Repo) Service {
	s := &service{
		logger:   l,
		promRepo: promRepo,
		dbRepo:   dbRepo,
	}

	go s.checkLoop()

	return s
}
