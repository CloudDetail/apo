package log

import (
	"encoding/json"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

const (
	defaultParseInfo = "默认Java日志解析, 从日志字段中解析出level、thread、method信息"
	defaultParseName = "默认JAVA日志解析"
	defaultRouteRule = `!starts_with(string!(."k8s.pod.name"), "apo")`
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

func (s *service) CreateLogTable(req *request.LogTableRequest) (*response.LogTableResponse, error) {
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
	// 不存在才去插入logtableinfo
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

func (s *service) DropLogTable(req *request.LogTableRequest) (*response.LogTableResponse, error) {

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

func (s *service) UpdateLogTable(req *request.LogTableRequest) (*response.LogTableResponse, error) {
	res := &response.LogTableResponse{}
	logtable := &database.LogTableInfo{
		Cluster:  req.Cluster,
		DataBase: req.DataBase,
		Table:    req.TableName,
	}
	err := s.dbRepo.OperateLogTableInfo(logtable, database.QUERY)
	if err != nil {
		res.Err = err.Error()
		return res, nil
	}
	var fields []request.Field
	err = json.Unmarshal([]byte(logtable.Fields), &fields)
	if err != nil {
		res.Err = err.Error()
		return res, nil
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
	res.Sqls = sqls
	if err != nil {
		res.Err = err.Error()
		return res, nil
	}
	fieldsJSON, err := json.Marshal(req.Fields)
	if err != nil {
		res.Err = err.Error()
		return res, nil
	}
	logtable.Fields = string(fieldsJSON)
	err = s.dbRepo.OperateLogTableInfo(logtable, database.UPDATE)
	if err != nil {
		res.Err = err.Error()
		return res, nil
	}
	return res, nil
}
