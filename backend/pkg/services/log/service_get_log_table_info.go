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
	logtables := make(map[string][]response.LogTable)
	for _, row := range rows {
		logtables[row.DataBase] = append(logtables[row.DataBase], response.LogTable{
			Cluster:   row.Cluster,
			TableName: row.Table,
		})
	}
	res.LogTables = logtables
	return res, nil
}
