// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"encoding/json"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

const (
	defaultParseInfo = ""
	defaultParseName = "all_logs"
	defaultRouteRule = `starts_with(string!(."k8s.pod.name"), "apo")`
	defaultParseRule = `.msg, err = parse_regex(.content, r' \[(?P<level>.*?)\] \[(?P<thread>.*?)\] \[(?P<method>.*?)\(.*?\)\] - (?P<msg>.*)')
if err == null {
	.content = encode_json(.msg)
}
del(.msg)
`
)

var defaultRouteRuleMap = map[string]string{
	"k8s.pod.name": "apo",
}

func (s *service) InitParseLogTable(ctx core.Context, req *request.LogTableRequest) (*response.LogTableResponse, error) {
	sqls, err := s.chRepo.CreateLogTable(req)
	res := &response.LogTableResponse{Sqls: sqls}
	if err != nil {
		res.Err = err.Error()
		return res, nil
	}
	fieldsJSON, err := json.Marshal(req.Fields)
	if err != nil {
		res.Err = err.Error()
		return res, nil
	}
	logtable := &database.LogTableInfo{
		Cluster:   req.Cluster,
		DataBase:  req.DataBase,
		Fields:    string(fieldsJSON),
		Table:     req.TableName,
		ParseName: defaultParseName,
		RouteRule: defaultRouteRule,
		ParseRule: defaultParseRule,
		ParseInfo: defaultParseInfo,
	}
	// does not exist to insert logtableinfo
	err = s.dbRepo.OperateLogTableInfo(logtable, database.QUERY)
	if err != nil {
		err = s.dbRepo.OperateLogTableInfo(logtable, database.INSERT)
		if err != nil {
			res.Err = err.Error()
			return res, nil
		}
	}

	return res, nil
}

func (s *service) DropLogTable(ctx core.Context, req *request.LogTableRequest) (*response.LogTableResponse, error) {

	sqls, err := s.chRepo.DropLogTable(req)
	res := &response.LogTableResponse{Sqls: sqls}
	if err != nil {
		res.Err = err.Error()
		return res, nil
	}
	logtable := &database.LogTableInfo{
		Cluster:  req.Cluster,
		DataBase: req.DataBase,
		Table:    req.TableName,
	}
	err = s.dbRepo.OperateLogTableInfo(logtable, database.DELETE)
	if err != nil {
		res.Err = err.Error()
	}
	return res, nil
}

func (s *service) UpdateLogTable(ctx core.Context, req *request.LogTableRequest) (*response.LogTableResponse, error) {
	res := &response.LogTableResponse{}
	logtable := &database.LogTableInfo{
		Cluster:  req.Cluster,
		DataBase: req.DataBase,
		Table:    req.TableName,
	}
	err := s.dbRepo.OperateLogTableInfo(logtable, database.QUERY)
	if err != nil {
		return nil, err
	}
	var oldFields []request.Field
	err = json.Unmarshal([]byte(logtable.Fields), &oldFields)
	if err != nil {
		return nil, err
	}

	sqls, err := s.chRepo.UpdateLogTable(req, oldFields)
	res.Sqls = sqls
	if err != nil {
		return nil, err
	}
	return res, nil
}
