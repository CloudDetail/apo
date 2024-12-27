// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"github.com/CloudDetail/apo/backend/internal/model/request"
	"github.com/CloudDetail/apo/backend/internal/model/response"
	"github.com/CloudDetail/apo/backend/internal/repository/database"
)

var _ Service = (*service)(nil)

type Service interface {
	Create(req *request.CreateRequest) (resp *response.CreateResponse, err error)
	PageList(req *request.ListRequest) (resp *response.ListResponse, err error)
	Detail(req *request.DetailRequest) (info *response.DetailResponse, err error)
	Delete(req *request.DeleteRequest) error
}

type service struct {
	dbRepo database.Repo
}

func New(dbRepo database.Repo) Service {
	return &service{
		dbRepo: dbRepo,
	}
}
