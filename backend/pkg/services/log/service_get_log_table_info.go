package log

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetLogTableInfo(req *request.LogTableInfoRequest) (*response.LogTableInfoResponse, error) {
	rows, err := s.dbRepo.GetAllLogTable()
	res := &response.LogTableInfoResponse{}
	if err != nil {
		res.Err = err.Error()
		return res, nil
	}
	parses := make([]response.Parse, 0)
	for _, row := range rows {
		parses = append(parses, response.Parse{
			DataBase:  row.DataBase,
			ParseName: row.ParseName,
			TableName: row.Table,
		})
	}

	others, err := s.dbRepo.GetAllOtherLogTable()
	if err != nil {
		res.Err = err.Error()
		return res, nil
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
		for dataBase, tables := range DataBases {
			instances = append(instances, response.Instance{
				InstanceName: instance,
				DataBases: []response.DBInfo{
					{
						DataBase: dataBase,
						Tables:   tables,
					},
				},
			})
		}
	}

	res.Parses = parses
	res.Instances = instances
	return res, nil
}
