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
	parsesMap := make(map[string][]response.ParseInfo)
	for _, row := range rows {
		parsesMap[row.DataBase] = append(parsesMap[row.DataBase], response.ParseInfo{
			ParseName: row.ParseName,
			TableName: row.Table,
		})
	}
	for dataBase, parseInfos := range parsesMap {
		parses = append(parses, response.Parse{
			DataBase:   dataBase,
			ParseInfos: parseInfos,
		})
	}

	others, err := s.dbRepo.GetAllOtherLogTable()
	if err != nil {
		res.Err = err.Error()
		return res, nil
	}
	logTables := make([]response.LogTable, 0)
	for _, other := range others {
		logTables = append(logTables, response.LogTable{
			Cluster:  other.Cluster,
			DataBase: other.DataBase,
			Tables:   []response.LogTableInfo{{LogField: other.LogField, TableName: other.Table, TimeField: other.TimeField}},
		})
	}
	res.Parses = parses
	res.LogTables = logTables
	return res, nil
}
