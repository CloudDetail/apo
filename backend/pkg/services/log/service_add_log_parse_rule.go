package log

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/services/log/vector"
	"gopkg.in/yaml.v3"
)

func getRouteRule(routeMap map[string]string, service string) string {
	var res []string
	for k, v := range routeMap {
		res = append(res, fmt.Sprintf(`starts_with(string!(."%s"), %s)`, k, v))
	}
	if service != "" {
		res = append(res, fmt.Sprintf(`starts_with(string!(."k8s.pod.name"), %s)`, service))
	}
	return strings.Join(res, " && ")
}

func (s *service) AddLogParseRule(req *request.AddLogParseRequest) (*response.LogParseResponse, error) {
	// 先去建表
	logReq := &request.LogTableRequest{
		TableName: req.ParseName,
	}
	logReq.TTL = req.LogTable.TTL
	logReq.Fields = req.LogTable.Fields
	logReq.Buffer = req.LogTable.Buffer
	logReq.FillerValue()

	// 获取当前时间
	now := time.Now()
	currentTimestamp := now.UnixMicro()
	sevenDaysAgo := now.AddDate(0, 0, -7)
	sevenDaysAgoTimestamp := sevenDaysAgo.UnixMicro()

	instances, err := s.promRepo.GetActiveInstanceList(sevenDaysAgoTimestamp, currentTimestamp, req.Service)
	if err != nil {
		return nil, err
	}
	var deployName string
	for instanceName, _ := range instances.GetInstanceIdMap() {
		parts := strings.Split(instanceName, "-")
		if len(parts) >= 3 {
			deployName = strings.Join(parts[:len(parts)-2], "-")
			break
		}
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
		TableName: req.ParseName,
		ParseRule: req.ParseRule,
		RouteRule: getRouteRule(req.RouteRule, deployName),
	}
	newData, err := p.AddParseRule(vectorCfg)
	if err != nil {
		res.Err = err.Error()
		return res, nil
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
		res.Err = err.Error()
		return res, nil
	}

	// 更新sqlite表信息
	log := database.LogTableInfo{
		ParseInfo: req.ParseInfo,
		ParseName: req.ParseName,
		ParseRule: req.ParseRule,
		RouteRule: getRouteRule(req.RouteRule, deployName),
		Table:     req.ParseName,
		DataBase:  logReq.DataBase,
		Cluster:   logReq.Cluster,
		Fields:    string(fieldsJSON),
		Service:   req.Service,
	}
	err = s.dbRepo.OperateLogTableInfo(&log, database.INSERT)
	if err != nil {
		return nil, err
	}

	return res, nil
}
