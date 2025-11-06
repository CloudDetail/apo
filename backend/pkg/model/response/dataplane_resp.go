// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package response

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

type QueryServicesResponse struct {
	Msg     string           `json:"msg"`
	Results []*model.Service `json:"results"`
}

type QueryServiceRedChartsResponse struct {
	Msg     string              `json:"msg"`
	Results []*QueryChartResult `json:"results"`
}

type ListServiceRedChartsResponse struct {
	Msg     string                         `json:"msg"`
	Results map[string][]*QueryChartResult `json:"results"` // service -> []*QueryChartResult
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

type ListCustomTopologyResponse struct {
	Topologies []*database.CustomServiceTopology `json:"topologies"`
}

type ListServiceNameRuleResponse struct {
	Rules []*ListServiceNameRule `json:"rules"`
}

type ListServiceNameRule struct {
	Id          int                                  `json:"id"`
	ServiceName string                               `json:"serviceName"`
	ClusterId   string                               `json:"clusterId"`
	Conditions  []*database.ServiceNameRuleCondition `json:"conditions"`
	Instances   []string                             `json:"instances"`
}

type CheckServiceNameRuleResponse struct {
	Instances []*model.MatchServiceInstance `json:"instances"`
}

type QueryAPPInfoTagsResponse struct {
	Labels []string `json:"labels"`
}

type QueryAPPInfoTagValuesResponse struct {
	Label  string   `json:"label"`
	Values []string `json:"values"`
}

type ListDataPlaneTypeResponse struct {
	DPTypes []integration.DataPlaneType `json:"dpTypes"`
}

type ListDataPlaneResponse struct {
	DataPlanes []integration.DataPlaneWithClusterIDs `json:"dataPlanes"`
}

type CheckMinimalDataPlaneResponse struct {
	AgentReady     bool     `json:"agentReady"`
	Clusters       []string `json:"clusters"`
	DataPlaneExist bool     `json:"dataPlaneExist"`
	DataSource     []string `json:"dataSource"`
}

type GetDatasourcesResponse struct {
	ServiceCount uint64                  `json:"serviceCount"`
	Status       string                  `json:"status"`
	Datasources  []model.DatasourceCount `json:"datasources"`
}

type GetServiceWithInstancesResponse struct {
	Services         []ServiceWithInstances `json:"services"`
	model.Pagination `json:",inline"`
}

type InstanceWithSource struct {
	model.ServiceInstance `json:",inline"`
	InstanceName          string `json:"instanceName"`
	InstanceSource        `json:"source"`
}

type InstanceSource struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type ServiceWithInstances struct {
	model.Service `json:",inline"`
	Instances     []InstanceWithSource `json:"instances"`
}
