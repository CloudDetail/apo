// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"encoding/json"
	"errors"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

func log2item(logs []map[string]any, logFields map[string]interface{}) ([]response.LogItem, error) {
	var timestamp int64
	logitems := make([]response.LogItem, len(logs))
	for i, tags := range logs {
		content := tags["content"]
		delete(tags, "content")
		fields := make(map[string]interface{})
		for k, v := range tags {
			if k == "timestamp" {
				ts, ok := v.(time.Time)
				if ok {
					timestamp = ts.UnixMicro()
				} else {
					return nil, errors.New("timestamp type error")
				}
				delete(tags, k)
			}
			vMap, ok := v.(map[string]string)
			if ok {
				for k2, v2 := range vMap {
					tags[k+"."+k2] = v2
				}
				delete(tags, k)
			}

			if _, exists := logFields[k]; exists {
				fields[k] = v
				delete(tags, k)
			}
		}

		logitems[i] = response.LogItem{
			Content:   content,
			Tags:      tags,
			Time:      timestamp,
			LogFields: fields,
		}
	}
	return logitems, nil
}

func (s *service) QueryLogContext(ctx core.Context, req *request.LogQueryContextRequest) (*response.LogQueryContextResponse, error) {

	logFields := map[string]interface{}{}
	model := &database.LogTableInfo{
		DataBase: req.DataBase,
		Table:    req.TableName,
	}
	// query log field json
	s.dbRepo.OperateLogTableInfo(model, database.QUERY)
	var fields []request.Field
	_ = json.Unmarshal([]byte(model.Fields), &fields)

	for _, field := range fields {
		logFields[field.Name] = struct{}{}
	}

	front, end, _ := s.chRepo.QueryLogContext(req)

	frontItem, err := log2item(front, logFields)
	if err != nil {
		return nil, err
	}
	endItem, err := log2item(end, logFields)
	if err != nil {
		return nil, err
	}

	return &response.LogQueryContextResponse{
		Front: frontItem,
		Back:  endItem,
	}, nil
}
