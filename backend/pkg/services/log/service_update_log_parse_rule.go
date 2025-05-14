// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"encoding/json"
	"strings"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/services/log/vector"
	"gopkg.in/yaml.v3"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

func (s *service) UpdateLogParseRule(ctx_core core.Context, req *request.UpdateLogParseRequest) (*response.LogParseResponse, error) {
	// Update the log table
	fields := make([]request.Field, 0)
	if req.IsStructured {
		fields = req.TableFields
	} else {
		matchesFields := fieldsRegexp.FindAllStringSubmatch(req.ParseRule, -1)
		for _, match := range matchesFields {
			if match[1] == "msg" || match[1] == "ts" {
				continue
			}

			parsedField := request.Field{
				Name:	match[1],
				Type:	"String",
			}

			for _, customizedFiled := range req.TableFields {
				if parsedField.Name == customizedFiled.Name {
					parsedField.Type = customizedFiled.Type
				}
			}
			fields = append(fields, parsedField)
		}
	}

	logReq := &request.LogTableRequest{
		DataBase:	req.DataBase,
		TableName:	req.TableName,
		Fields:		fields,
		IsStructured:	req.IsStructured,
	}
	logReq.FillerValue()
	_, err := s.UpdateLogTable(ctx_core, logReq)
	if err != nil {
		return nil, err
	}

	// update k8s configmap
	data, err := s.k8sApi.GetVectorConfigFile()
	if err != nil {
		return nil, err
	}
	var vectorCfg vector.VectorConfig
	err = yaml.Unmarshal([]byte(data["aggregator.yaml"]), &vectorCfg)
	if err != nil {
		return nil, err
	}
	// Structured log, no parse rule is required
	if req.IsStructured {
		req.ParseRule = ""
	}
	p := vector.ParseInfo{
		ParseName:	req.ParseName,
		RouteRule:	getRouteRule(req.RouteRule),
		ParseRule:	req.ParseRule,
	}

	newData, err := p.UpdateParseRule(vectorCfg)
	if err != nil {
		return nil, err
	}
	err = s.k8sApi.UpdateVectorConfigFile(newData)
	if err != nil {
		return nil, err
	}

	// modify table records
	fieldsJSON, err := json.Marshal(logReq.Fields)
	if err != nil {
		return nil, err
	}

	log := database.LogTableInfo{
		Service:	strings.Join(req.Service, ","),
		ParseInfo:	req.ParseInfo,
		RouteRule:	getRouteRule(req.RouteRule),
		Fields:		string(fieldsJSON),
		Table:		req.TableName,
		DataBase:	req.DataBase,
		IsStructured:	req.IsStructured,
		ParseRule:	req.ParseRule,
	}

	err = s.dbRepo.UpdateLogParseRule(ctx_core, &log)
	if err != nil {
		return nil, err
	}

	res := &response.LogParseResponse{
		ParseName:	req.ParseName,
		ParseRule:	req.ParseRule,
		RouteRule:	req.RouteRule,
		IsStructured:	req.IsStructured,
	}
	return res, nil
}
