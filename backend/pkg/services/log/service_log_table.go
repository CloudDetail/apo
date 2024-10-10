package log

import (
	"encoding/json"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

func (s *service) CreateLogTable(req *request.LogTableRequest) (*response.LogTableResponse, error) {
	sqls, err := s.chRepo.CreateLogTable(req)
	if err != nil {
		return nil, err
	}
	fieldsJSON, err := json.Marshal(req.Fields)
	if err != nil {
		return nil, err
	}
	logtable := &database.LogTableInfo{
		Cluster:  req.Cluster,
		DataBase: req.DataBase,
		Fields:   string(fieldsJSON),
		Table:    req.TableName,
	}
	err = s.dbRepo.OperateLogTableInfo(logtable, database.INSERT)
	if err != nil {
		return nil, err
	}
	return &response.LogTableResponse{Sqls: sqls}, nil
}

func (s *service) DropLogTable(req *request.LogTableRequest) (*response.LogTableResponse, error) {
	sqls, err := s.chRepo.DropLogTable(req)
	if err != nil {
		return nil, err
	}
	logtable := &database.LogTableInfo{
		Cluster:  req.Cluster,
		DataBase: req.DataBase,
		Table:    req.TableName,
	}
	err = s.dbRepo.OperateLogTableInfo(logtable, database.DELETE)
	if err != nil {
		return nil, err
	}
	return &response.LogTableResponse{Sqls: sqls}, nil
}

func (s *service) UpdateLogTable(req *request.LogTableRequest) (*response.LogTableResponse, error) {
	logtable := &database.LogTableInfo{
		Cluster:  req.Cluster,
		DataBase: req.DataBase,
		Table:    req.TableName,
	}
	err := s.dbRepo.OperateLogTableInfo(logtable, database.QUERY)
	if err != nil {
		return nil, err
	}
	var fields []request.Field
	err = json.Unmarshal([]byte(logtable.Fields), &fields)
	if err != nil {
		return nil, err
	}
	fieldsname := map[string]struct{}{}
	newFields := []request.Field{}
	for _, field := range fields {
		fieldsname[field.Name] = struct{}{}
	}
	for _, field := range req.Fields {
		if _, ok := fieldsname[field.Name]; !ok {
			newFields = append(newFields, field)
		}
	}
	sqls, err := s.chRepo.UpdateLogTable(req, newFields, fields)
	if err != nil {
		return nil, err
	}
	fieldsJSON, err := json.Marshal(req.Fields)
	if err != nil {
		return nil, err
	}
	logtable.Fields = string(fieldsJSON)
	err = s.dbRepo.OperateLogTableInfo(logtable, database.UPDATE)
	if err != nil {
		return nil, err
	}
	return &response.LogTableResponse{Sqls: sqls}, nil
}

func (s *service) GetLogTableInfo(req *request.LogTableRequest) (*response.LogTableResponse, error) {
	logtable := &database.LogTableInfo{
		Cluster:  req.Cluster,
		DataBase: req.DataBase,
		Table:    req.TableName,
	}
	err := s.dbRepo.OperateLogTableInfo(logtable, database.QUERY)
	if err != nil {
		return nil, err
	}
	var fields []request.Field
	err = json.Unmarshal([]byte(logtable.Fields), &fields)
	if err != nil {
		return nil, err
	}
	return &response.LogTableResponse{Sqls: []string{}}, nil
}
