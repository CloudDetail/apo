// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"encoding/json"
	"errors"
	"gorm.io/gorm"
	"regexp"
	"strings"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

var routeReg = regexp.MustCompile(`\"(.*?)\"`)

func getRouteRuleMap(routeRule string) map[string]string {
	res := make(map[string][]string)
	lines := strings.Split(routeRule, "||")
	for _, line := range lines {
		if line == "" {
			continue
		}
		matches := routeReg.FindAllStringSubmatch(line, -1)
		if len(matches) == 2 {
			key := matches[0][1]
			value := matches[1][1]
			// if key == "k8s.pod.name" {
			// 	continue
			// }
			res[key] = append(res[key], value)
		}
	}
	rc := make(map[string]string)
	for k, v := range res {
		rc[k] = strings.Join(v, ",")
	}
	return rc
}

func (s *service) GetLogParseRule(ctx_core core.Context, req *request.QueryLogParseRequest) (*response.LogParseResponse, error) {
	model := &database.LogTableInfo{
		DataBase:	req.DataBase,
		Table:		req.TableName,
	}
	err := s.dbRepo.OperateLogTableInfo(model, database.QUERY)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &response.LogParseResponse{
			ParseName:	defaultParseName,
			ParseRule:	defaultParseRule,
			RouteRule:	defaultRouteRuleMap,
		}, nil
	} else if err != nil {
		return nil, err
	}

	logFields := []request.Field{}
	json.Unmarshal([]byte(model.Fields), &logFields)
	return &response.LogParseResponse{
		Service:	strings.Split(model.Service, ","),
		ParseName:	model.ParseName,
		ParseRule:	model.ParseRule,
		ParseInfo:	model.ParseInfo,
		RouteRule:	getRouteRuleMap(model.RouteRule),
		LogFields:	logFields,
		IsStructured:	model.IsStructured,
	}, nil
}
