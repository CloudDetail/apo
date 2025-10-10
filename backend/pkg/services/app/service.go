// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

type Service interface {
	UploadLogo(req *request.UploadLogoRequest) error
	GetLogo() (response.GetLogoResp, error)
}

type service struct {
	dbRepo database.Repo
}

func New(dbRepo database.Repo) Service {
	s := &service{
		dbRepo: dbRepo,
	}

	return s
}
