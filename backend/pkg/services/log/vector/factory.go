// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package vector

import (
	"encoding/json"
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

func (p *ParseInfo) AddParseRule(config VectorConfig) ([]byte, error) {
	routeLogs, ok := config.Transforms["route_logs"].(map[string]interface{})
	if ok {
		route := routeLogs["route"].(map[string]interface{})
		route[p.ParseName+"_route"] = p.RouteRule
	}

	_, ok = config.Transforms["parse_"+p.ParseName].(map[string]interface{})
	if ok {
		return nil, errors.New("规则解析名已存在，请确保唯一")
	} else {
		new_transform := map[string]interface{}{
			"type":   "remap",
			"inputs": []string{"route_logs." + p.ParseName + "_route"},
			"source": p.ParseRule,
		}
		config.Transforms["parse_"+p.ParseName] = new_transform
	}

	sinkOriginal := config.Sinks["to_default_java"].(map[string]interface{})
	sinkTempBytes, _ := json.Marshal(sinkOriginal)
	var sinkTemp map[string]interface{}
	json.Unmarshal(sinkTempBytes, &sinkTemp)

	sinkTemp["inputs"] = []string{"parse_" + p.ParseName}
	sinkTemp["table"] = p.TableName + "_buffer"
	config.Sinks["to_"+p.ParseName] = sinkTemp

	updatedData, err := yaml.Marshal(&config)
	if err != nil {
		return nil, errors.New("marshal failed")
	}
	return updatedData, nil
}

func (p *ParseInfo) UpdateParseRule(config VectorConfig) ([]byte, error) {
	routeLogs, ok := config.Transforms["route_logs"].(map[string]interface{})
	if ok {
		route := routeLogs["route"].(map[string]interface{})
		route[p.ParseName+"_route"] = p.RouteRule
	} else {
		return nil, errors.New("配置文件更新出错")
	}

	// 更新 parse_test 的 source 字段
	parseTest, ok := config.Transforms["parse_"+p.ParseName].(map[string]interface{})
	if ok {
		parseTest["source"] = p.ParseRule
	} else {
		return nil, errors.New("解析规则" + p.ParseName + "不存在")
	}
	updatedData, err := yaml.Marshal(&config)
	if err != nil {
		return nil, errors.New("配置文件更新出错")
	}
	return updatedData, nil
}

func (p *ParseInfo) DeleteParseRule(config VectorConfig) ([]byte, error) {
	routeLogs, ok := config.Transforms["route_logs"].(map[string]interface{})
	if ok {
		route := routeLogs["route"].(map[string]interface{})
		delete(route, p.ParseName+"_route")
		routeLogs["route"] = route
		config.Transforms["route_logs"] = routeLogs
	}
	delete(config.Transforms, "parse_"+p.ParseName)
	delete(config.Sinks, "to_"+p.ParseName)
	updatedData, err := yaml.Marshal(&config)
	if err != nil {
		return nil, errors.New("配置文件更新出错")
	}
	return updatedData, nil
}
