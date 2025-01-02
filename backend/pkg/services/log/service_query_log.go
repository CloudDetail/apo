// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

func (s *service) QueryLog(req *request.LogQueryRequest) (*response.LogQueryResponse, error) {
	offset := (req.PageNum - 1) * req.PageSize
	if offset > 10000 {
		logcharts, _ := s.GetLogChart(req)
		var count = 0
		for _, chart := range logcharts.Histograms {
			count += int(chart.Count)
			if count > offset {
				offset = count - int(chart.Count)
				req.StartTime = chart.From
				break
			}
		}
	}

	req.PageNum = offset
	logs, sql, err := s.chRepo.QueryAllLogs(req)
	res := &response.LogQueryResponse{Query: sql}
	if err != nil {
		res.Err = err.Error()
		return res, nil
	}

	// query column name and type
	rows, err := s.chRepo.OtherLogTableInfo(&request.OtherTableInfoRequest{
		DataBase:  req.DataBase,
		TableName: req.TableName,
	})
	if err != nil {
		res.Err = err.Error()
		return res, nil
	}
	allFields := []string{}
	for _, row := range rows {
		allFields = append(allFields, row["name"].(string))
	}

	hMap := make(map[string]struct{})

	if req.IsExternal {
		res.HiddenFields = allFields
	} else {
		hiddenFields := []string{}
		model := &database.LogTableInfo{
			DataBase: req.DataBase,
			Table:    req.TableName,
		}
		// query log field json
		s.dbRepo.OperateLogTableInfo(model, database.QUERY)
		var fields []request.Field
		_ = json.Unmarshal([]byte(model.Fields), &fields)

		for _, field := range fields {
			hiddenFields = append(hiddenFields, field.Name)
		}

		for _, item := range hiddenFields {
			hMap[item] = struct{}{}
		}

		var defaultFields []string
		for _, item := range allFields {
			if _, exists := hMap[item]; !exists {
				if item == req.TimeField || item == req.LogField {
					continue
				}
				defaultFields = append(defaultFields, item)
			}
		}
		res.HiddenFields = hiddenFields
		res.DefaultFields = defaultFields
	}

	res.Limited = req.PageSize

	if len(logs) == 0 {
		res.Err = "No logs found"
		return res, nil
	}

	var timestamp int64
	logitems := make([]response.LogItem, len(logs))
	for i, log := range logs {
		content := log[req.LogField]
		delete(log, req.LogField)

		logFields := map[string]interface{}{}
		for k, v := range log {
			if k == req.TimeField {
				ts, ok := v.(time.Time)
				if ok {
					timestamp = ts.UnixMicro()
				} else {
					return nil, errors.New("timestamp type error")
				}
				delete(log, k)
			}
			vMap, ok := v.(map[string]string)
			if ok {
				for k2, v2 := range vMap {
					log[k+"."+k2] = v2
				}
				delete(log, k)
			}

			// this pair of kv is log field
			if _, exists := hMap[k]; exists {
				logFields[k] = v
				delete(log, k)
			}
		}
		item := response.LogItem{
			Time: timestamp,
		}
		if req.IsExternal {
			item.LogFields = log
		} else {
			item.Tags = log
			item.Content = content
			item.LogFields = logFields
		}

		logitems[i] = item
	}

	res.Logs = logitems
	res.Query = sql
	return res, nil
}
