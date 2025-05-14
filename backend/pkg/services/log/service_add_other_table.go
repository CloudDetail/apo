// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

func (s *service) AddOtherTable(ctx_core core.Context, req *request.AddOtherTableRequest) (*response.AddOtherTableResponse, error) {
	res := &response.AddOtherTableResponse{}
	model := &database.OtherLogTable{
		Cluster:	req.Cluster,
		DataBase:	req.DataBase,
		Instance:	req.Instance,
		LogField:	req.LogField,
		Table:		req.Table,
		TimeField:	req.TimeField,
	}
	err := s.dbRepo.OperatorOtherLogTable(ctx_core, model, database.QUERY)
	if err == nil {
		res.Err = "table already exists"
		return res, nil
	} else {
		err = s.dbRepo.OperatorOtherLogTable(ctx_core, model, database.INSERT)
	}
	if err != nil {
		res.Err = err.Error()
	}

	return res, nil
}
