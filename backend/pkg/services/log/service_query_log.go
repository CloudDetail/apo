package log

import (
	"encoding/json"
	"errors"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

func (s *service) QueryLog(req *request.LogQueryRequest) (*response.LogQueryResponse, error) {
	logs, sql, err := s.chRepo.QueryAllLogs(req)
	if err != nil {
		return nil, err
	}
	if len(logs) == 0 {
		return nil, errors.New("no found logs")
	}
	allFileds := []string{}
	if len(logs) > 0 {
		for k := range logs[0] {
			allFileds = append(allFileds, k)
		}
	}

	hiddenFields := []string{}
	model := &database.LogTableInfo{
		DataBase: req.DataBase,
		Table:    req.TableName,
	}
	s.dbRepo.OperateLogTableInfo(model, database.QUERY)
	var fields []request.Field
	err = json.Unmarshal([]byte(model.Fields), &fields)
	if err != nil {
		return nil, err
	}

	for _, field := range fields {
		hiddenFields = append(hiddenFields, field.Name)
	}

	hMap := make(map[string]struct{})
	for _, item := range hiddenFields {
		hMap[item] = struct{}{}
	}

	var defaultFields []string
	for _, item := range allFileds {
		if _, exists := hMap[item]; !exists {
			if item == "timestamp" {
				continue
			}
			defaultFields = append(defaultFields, item)
		}
	}

	res := &response.LogQueryResponse{
		Limited:       req.PageSize,
		HiddenFields:  hiddenFields,
		DefaultFields: defaultFields,
		Logs:          logs,
		Query:         sql,
	}

	return res, nil
}
