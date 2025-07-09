// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package response

import "github.com/CloudDetail/apo/backend/pkg/model"

type QueryServicesResponse struct {
	Msg     string                `json:"msg"`
	Results []*QueryServiceResult `json:"results,omitempty"`
}

type QueryServiceRedChartsResponse struct {
	Msg     string              `json:"msg"`
	Results []*QueryChartResult `json:"results,omitempty"`
}

type QueryServiceEndpointsResponse struct {
	Msg     string   `json:"msg"`
	Results []string `json:"results"`
}

type QueryServiceInstancesResponse struct {
	Msg     string                   `json:"msg"`
	Results []*model.ServiceInstance `json:"results"`
}

type QueryServiceNameResponse struct {
	Msg    string `json:"msg"`
	Result string `json:"result"`
}

type QueryTopologyResponse struct {
	Msg     string                      `json:"msg"`
	Results []*model.ServiceToplogyNode `json:"results"`
}

type QueryServiceResult struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Source string `json:"source"`
}

type QueryChartResult struct {
	Title      string        `json:"title"`
	Unit       string        `json:"unit"`
	Timeseries []*Timeseries `json:"timeseries"`
}

type Timeseries struct {
	Legend       string            `json:"legend"`
	LegendFormat string            `json:"legendFormat"`
	Labels       map[string]string `json:"labels"`
	Chart        TempChartObject   `json:"chart"`
}
