// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package log

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) OtherTable(ctx core.Context, req *request.OtherTableRequest) (*response.OtherTableResponse, error) {
	res := &response.OtherTableResponse{}
	rows, err := s.chRepo.OtherLogTable(ctx)
	if err != nil {
		return nil, err
	}
	resMap := make(map[string][]string)
	for _, row := range rows {
		resMap[row["database"].(string)] = append(resMap[row["database"].(string)], row["name"].(string))
	}
	others := make([]response.OtherDB, 0)
	for db, tables := range resMap {
		othertables := make([]response.OtherTable, 0)
		for _, table := range tables {
			othertables = append(othertables, response.OtherTable{
				TableName: table,
			})
		}
		others = append(others, response.OtherDB{
			DataBase: db,
			Tables:   othertables,
		})
	}
	res.OtherTables = others

	return res, nil
}
