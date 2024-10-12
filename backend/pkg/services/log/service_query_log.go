package log

import (
	"encoding/json"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

func (s *service) QueryLog(req *request.LogQueryRequest) (*response.LogQueryResponse, error) {
	logs, sql, err := s.chRepo.QueryAllLogs(req)
	res := &response.LogQueryResponse{Query: sql}
	if err != nil {
		res.Err = err.Error()
		return res, nil
	}
	if len(logs) == 0 {
		res.Err = "未查询到任何日志数据"
		return res, nil
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
	_ = json.Unmarshal([]byte(model.Fields), &fields)
	// if err != nil {
	// 	return nil, err
	// }

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
			if item == "timestamp" || item == "content" {
				continue
			}
			defaultFields = append(defaultFields, item)
		}
	}
	logitems := make([]response.LogItem, len(logs))
	for i, log := range logs {
		content := log["content"]
		delete(log, "content")
		logitems[i] = response.LogItem{
			Content: content,
			Tags:    log,
		}
	}

	res.Limited = req.PageSize
	res.HiddenFields = hiddenFields
	res.DefaultFields = defaultFields
	res.Logs = logitems
	res.Query = sql
	return res, nil
}
