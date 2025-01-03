// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package response

import "github.com/CloudDetail/apo/backend/pkg/model/request"

type LogParseResponse struct {
	ParseInfo    string            `json:"parseInfo"`
	Service      []string          `json:"serviceName"`
	ParseName    string            `json:"parseName"`
	RouteRule    map[string]string `json:"routeRule"`
	ParseRule    string            `json:"parseRule"`
	LogFields    []request.Field   `json:"tableFields"`
	IsStructured bool              `json:"isStructured"`
}

type GetServiceRouteResponse struct {
	RouteRule map[string]string `json:"routeRule"`
}
