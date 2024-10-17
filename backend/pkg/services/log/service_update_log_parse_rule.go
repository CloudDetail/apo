package log

import (
	"encoding/json"
	"strings"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/services/log/vector"
	"gopkg.in/yaml.v3"
)

func (s *service) UpdateLogParseRule(req *request.UpdateLogParseRequest) (*response.LogParseResponse, error) {
	//更新日志表
	matchesFields := fieldsRegexp.FindAllStringSubmatch(req.ParseRule, -1)

	fields := make([]request.Field, 0)
	for _, match := range matchesFields {
		if match[1] == "msg" || match[1] == "ts" {
			continue
		}
		fields = append(fields, request.Field{
			Name: match[1],
			Type: "String",
		})
	}
	logReq := &request.LogTableRequest{
		DataBase:  req.DataBase,
		TableName: req.TableName,
		Fields:    fields,
	}
	logReq.FillerValue()
	_, err := s.UpdateLogTable(logReq)
	if err != nil {
		return nil, err
	}

	// 更新k8s configmap
	res := &response.LogParseResponse{
		ParseName: req.ParseName,
		ParseRule: req.ParseRule,
		RouteRule: req.RouteRule,
	}
	data, err := s.k8sApi.GetVectorConfigFile()
	if err != nil {
		return nil, err
	}
	var vectorCfg vector.VectorConfig
	err = yaml.Unmarshal([]byte(data["aggregator.yaml"]), &vectorCfg)
	if err != nil {
		return nil, err
	}
	p := vector.ParseInfo{
		ParseName: req.ParseName,
		ParseRule: req.ParseRule,
		RouteRule: getRouteRule(req.RouteRule),
	}
	newData, err := p.UpdateParseRule(vectorCfg)
	if err != nil {
		return nil, err
	}
	err = s.k8sApi.UpdateVectorConfigFile(newData)
	if err != nil {
		return nil, err
	}

	// 调整整个表结构

	fieldsJSON, err := json.Marshal(logReq.Fields)
	if err != nil {
		return nil, err
	}

	log := database.LogTableInfo{
		Service:   strings.Join(req.Service, ","),
		ParseRule: req.ParseRule,
		ParseInfo: req.ParseInfo,
		RouteRule: getRouteRule(req.RouteRule),
		Fields:    string(fieldsJSON),
		Table:     req.TableName,
		DataBase:  req.DataBase,
	}
	err = s.dbRepo.UpdateLogPaseRule(&log)
	if err != nil {
		return nil, err
	}

	return res, nil
}
