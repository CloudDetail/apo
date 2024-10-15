package log

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/services/log/vector"
	"gopkg.in/yaml.v3"
)

func (s *service) UpdateLogParseRule(req *request.UpdateLogParseRequest) (*response.LogParseResponse, error) {
	// 更新k8s configmap
	res := &response.LogParseResponse{
		ParseName: req.ParseName,
		ParseRule: req.ParseRule,
		RouteRule: getRouteRule(req.RouteRule),
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
		res.Err = err.Error()
		return res, nil
	}
	err = s.k8sApi.UpdateVectorConfigFile(newData)
	if err != nil {
		return nil, err
	}
	// 更新sqlite表信息
	log := database.LogTableInfo{
		ParseRule: req.ParseRule,
		RouteRule: getRouteRule(req.RouteRule),
		Table:     req.TableName,
		DataBase:  req.DataBase,
	}
	err = s.dbRepo.UpdateLogPaseRule(&log)
	if err != nil {
		return nil, err
	}

	return res, nil
}
