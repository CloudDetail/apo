// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"github.com/CloudDetail/apo/backend/internal/model/request"
	"github.com/CloudDetail/apo/backend/internal/model/response"
	"github.com/CloudDetail/apo/backend/internal/repository/database"
)

func (s *service) Create(req *request.CreateRequest) (resp *response.CreateResponse, err error) {
	mock := &database.Mock{
		Name: req.Name,
	}
	id, err := s.dbRepo.CreateMock(mock)
	if err != nil {
		return nil, err
	}

	return &response.CreateResponse{
		Id: id,
	}, nil
}
