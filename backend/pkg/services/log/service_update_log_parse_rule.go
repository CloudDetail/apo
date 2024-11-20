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
	fields := make([]request.Field, 0)
	if len(req.TableFields) > 0 {
		fields = req.TableFields
	} else {
		matchesFields := fieldsRegexp.FindAllStringSubmatch(req.ParseRule, -1)
		for _, match := range matchesFields {
			if match[1] == "msg" || match[1] == "ts" {
				continue
			}
			fields = append(fields, request.Field{
				Name: match[1],
				Type: "String",
			})
		}
	}
	logReq := &request.LogTableRequest{
		DataBase:     req.DataBase,
		TableName:    req.TableName,
		Fields:       fields,
		IsStructured: req.IsStructured,
	}
	logReq.FillerValue()
	_, err := s.UpdateLogTable(logReq)
	if err != nil {
		return nil, err
	}

	// 更新k8s configmap
	data, err := s.k8sApi.GetVectorConfigFile()
	if err != nil {
		return nil, err
	}
	var vectorCfg vector.VectorConfig
	err = yaml.Unmarshal([]byte(data["aggregator.yaml"]), &vectorCfg)
	if err != nil {
		return nil, err
	}
	// 结构化日志，不需要parse rule
	if req.IsStructured {
		req.ParseRule = ""
	}
	p := vector.ParseInfo{
		ParseName: req.ParseName,
		RouteRule: getRouteRule(req.RouteRule),
		ParseRule: req.ParseRule,
	}

	newData, err := p.UpdateParseRule(vectorCfg)
	if err != nil {
		return nil, err
	}
	err = s.k8sApi.UpdateVectorConfigFile(newData)
	if err != nil {
		return nil, err
	}

	// 修改表记录
	fieldsJSON, err := json.Marshal(logReq.Fields)
	if err != nil {
		return nil, err
	}

	log := database.LogTableInfo{
		Service:      strings.Join(req.Service, ","),
		ParseInfo:    req.ParseInfo,
		RouteRule:    getRouteRule(req.RouteRule),
		Fields:       string(fieldsJSON),
		Table:        req.TableName,
		DataBase:     req.DataBase,
		IsStructured: req.IsStructured,
		ParseRule:    req.ParseRule,
	}

	err = s.dbRepo.UpdateLogParseRule(&log)
	if err != nil {
		return nil, err
	}

	res := &response.LogParseResponse{
		ParseName:    req.ParseName,
		ParseRule:    req.ParseRule,
		RouteRule:    req.RouteRule,
		IsStructured: req.IsStructured,
	}
	return res, nil
}
