package vector

import (
	"errors"

	"gopkg.in/yaml.v2"
)

type VectorConfig struct {
	Sources    map[string]interface{} `yaml:"sources"`
	Transforms map[string]interface{} `yaml:"transforms"`
	Sinks      map[string]interface{} `yaml:"sinks"`
}

type ParseInfo struct {
	ParseName string
	TableName string
	RouteRule string
	ParseRule string
}

func (p *ParseInfo) UpdateParseRule(config VectorConfig) ([]byte, error) {
	// 更新 route_logs 的 route 字段
	routeLogs, ok := config.Transforms["route_logs"].(map[string]interface{})
	if ok {
		route := routeLogs["route"].(map[string]interface{})
		route[p.ParseName+"_route"] = p.RouteRule
	} else {
		return nil, errors.New("route_logs not found")
	}

	// 更新 parse_test 的 source 字段
	parseTest, ok := config.Transforms["parse_"+p.ParseName].(map[string]interface{})
	if ok {
		parseTest["source"] = p.ParseRule
	} else {
		return nil, errors.New("parse_" + p.ParseName + " not found")
	}
	updatedData, err := yaml.Marshal(&config)
	if err != nil {
		return nil, errors.New("marshal failed")
	}
	return updatedData, nil
}
