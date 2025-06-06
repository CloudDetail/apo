// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package log

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

func (s *service) DeleteOtherTable(ctx core.Context, req *request.DeleteOtherTableRequest) (*response.DeleteOtherTableResponse, error) {
	res := &response.DeleteOtherTableResponse{}
	model := &database.OtherLogTable{
		DataBase: req.DataBase,
		Instance: req.Instance,
		Table:    req.TableName,
	}
	err := s.dbRepo.OperatorOtherLogTable(ctx, model, database.DELETE)
	if err != nil {
		res.Err = err.Error()
	}
	return res, nil
}
