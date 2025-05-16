// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package log

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetLogTableInfo(ctx core.Context, req *request.LogTableInfoRequest) (*response.LogTableInfoResponse, error) {
	rows, err := s.dbRepo.GetAllLogTable()
	res := &response.LogTableInfoResponse{}
	if err != nil {
		return nil, err
	}
	parses := make([]response.Parse, 0)
	for _, row := range rows {
		parses = append(parses, response.Parse{
			DataBase:  row.DataBase,
			ParseName: row.ParseName,
			TableName: row.Table,
			ParseInfo: row.ParseInfo,
		})
	}

	others, err := s.dbRepo.GetAllOtherLogTable()
	if err != nil {
		return nil, err
	}
	instances := make([]response.Instance, 0)
	instanceMap := make(map[string]map[string][]response.LogTableInfo)
	for _, other := range others {
		instance, ok := instanceMap[other.Instance]
		if !ok {
			instance = make(map[string][]response.LogTableInfo)
			instanceMap[other.Instance] = instance
		}
		instance[other.DataBase] = append(instance[other.DataBase], response.LogTableInfo{
			LogField:  other.LogField,
			TableName: other.Table,
			TimeField: other.TimeField,
			Cluster:   other.Cluster,
		})
		instanceMap[other.Instance] = instance
	}
	for instance, DataBases := range instanceMap {
		databases := make([]response.DBInfo, 0)
		for dataBase, tables := range DataBases {
			databases = append(databases, response.DBInfo{
				DataBase: dataBase,
				Tables:   tables,
			})
		}
		instances = append(instances, response.Instance{
			InstanceName: instance,
			DataBases:    databases,
		})
	}

	res.Parses = parses
	res.Instances = instances
	return res, nil
}
