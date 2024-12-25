// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package request

type QueryLogParseRequest struct {
	DataBase  string `form:"dataBase" json:"dataBase"`
	TableName string `form:"tableName" json:"tableName"`
}

type UpdateLogParseRequest struct {
	DataBase     string            `json:"dataBase"`
	TableName    string            `json:"tableName"`
	ParseInfo    string            `json:"parseInfo"`
	ParseName    string            `json:"parseName"`
	Service      []string          `json:"serviceName"`
	RouteRule    map[string]string `json:"routeRule"`
	ParseRule    string            `json:"parseRule"`
	TableFields  []Field           `json:"tableFields"`
	IsStructured bool              `json:"isStructured"`
}

type AddLogParseRequest struct {
	ParseName    string            `json:"parseName"`
	Service      []string          `json:"serviceName"`
	ParseInfo    string            `json:"parseInfo"`
	RouteRule    map[string]string `json:"routeRule"`
	ParseRule    string            `json:"parseRule"`
	LogTable     LogTable          `json:"logTable"`
	Fields       []Field           `json:"tableFields"` // 自定义表字段
	IsStructured bool              `json:"isStructured"`
}

type GetServiceRouteRequest struct {
	Service []string `form:"serviceName"`
}

type LogTable struct {
	TTL    uint               `json:"ttl"`
	Fields []Field            `json:"fields"`
	Buffer BufferEngineConfig `json:"buffer"`
}

type DeleteLogParseRequest struct {
	DataBase  string `json:"dataBase"`
	TableName string `json:"tableName"`
	ParseName string `json:"parseName"`
}
