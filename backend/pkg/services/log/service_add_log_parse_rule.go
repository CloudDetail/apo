package log

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/services/log/vector"
	"gopkg.in/yaml.v3"
)

func getRouteRule(routeMap map[string]string) string {
	var res []string
	for k, v := range routeMap {
		if k == "k8s.pod.name" {
			strValues := strings.Split(v, ",")
			for _, vv := range strValues {
				res = append(res, fmt.Sprintf(`starts_with(string!(."%s"), "%s")`, k, vv))
			}
		} else {
			res = append(res, fmt.Sprintf(`starts_with(string!(."%s"), "%s")`, k, v))
		}
	}
	return strings.Join(res, " || ")
}

var fieldsRegexp = regexp.MustCompile(`\?P<(?P<name>\w+)>`)

func (s *service) AddLogParseRule(req *request.AddLogParseRequest) (*response.LogParseResponse, error) {
	// 先去建表
	logReq := &request.LogTableRequest{
		TableName: "logs_" + req.ParseName,
	}
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

	logReq.TTL = req.LogTable.TTL
	logReq.Fields = fields
	logReq.Buffer = req.LogTable.Buffer
	logReq.FillerValue()

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
		TableName: "logs_" + req.ParseName,
		ParseRule: req.ParseRule,
		RouteRule: getRouteRule(req.RouteRule),
	}
	newData, err := p.AddParseRule(vectorCfg)
	if err != nil {
		return nil, err
	}
	err = s.k8sApi.UpdateVectorConfigFile(newData)
	if err != nil {
		return nil, err
	}
	_, err = s.chRepo.CreateLogTable(logReq)
	if err != nil {
		return nil, err
	}
	fieldsJSON, err := json.Marshal(logReq.Fields)
	if err != nil {
		return nil, err
	}

	// 更新sqlite表信息
	log := database.LogTableInfo{
		ParseInfo: req.ParseInfo,
		ParseName: req.ParseName,
		ParseRule: req.ParseRule,
		RouteRule: getRouteRule(req.RouteRule),
		Table:     "logs_" + req.ParseName,
		DataBase:  logReq.DataBase,
		Cluster:   logReq.Cluster,
		Fields:    string(fieldsJSON),
		Service:   strings.Join(req.Service, ","),
	}
	err = s.dbRepo.OperateLogTableInfo(&log, database.INSERT)
	if err != nil {
		return nil, err
	}

	return res, nil
}
