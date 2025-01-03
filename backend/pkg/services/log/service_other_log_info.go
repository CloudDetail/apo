// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) OtherTableInfo(req *request.OtherTableInfoRequest) (*response.OtherTableInfoResponse, error) {
	res := &response.OtherTableInfoResponse{}
	rows, err := s.chRepo.OtherLogTableInfo(req)
	if err != nil {
		res.Err = err.Error()
		return res, nil
	}
	columns := make([]response.Column, 0, len(rows))
	for _, row := range rows {
		columns = append(columns, response.Column{
			Name: row["name"].(string),
			Type: row["type"].(string),
		})
	}
	res.Columns = columns
	return res, nil
}
