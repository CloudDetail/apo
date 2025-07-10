// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package request

type QueryServicesRequest struct {
	Cluster   string `form:"cluster"`                                      // query Cluster
	StartTime int64  `form:"startTime" binding:"min=0"`                    // query start time
	EndTime   int64  `form:"endTime" binding:"required,gtfield=StartTime"` // query end time
}

type QueryServiceRedChartsRequest struct {
	Cluster     string `form:"cluster"`                                      // query Cluster
	StartTime   int64  `form:"startTime" binding:"min=0"`                    // query start time
	EndTime     int64  `form:"endTime" binding:"required,gtfield=StartTime"` // query end time
	ServiceName string `form:"service" binding:"required"`                   // query service name
	Endpoint    string `form:"endpoint"`                                     // query Endpoint
}

type QueryServiceEndpointsRequest struct {
	Cluster     string `form:"cluster"`                                      // query Cluster
	StartTime   int64  `form:"startTime" binding:"min=0"`                    // query start time
	EndTime     int64  `form:"endTime" binding:"required,gtfield=StartTime"` // query end time
	ServiceName string `form:"service" binding:"required"`                   // query service name
}

type QueryServiceInstancesRequest struct {
	Cluster     string `form:"cluster"`                                      // query Cluster
	StartTime   int64  `form:"startTime" binding:"min=0"`                    // query start time
	EndTime     int64  `form:"endTime" binding:"required,gtfield=StartTime"` // query end time
	ServiceName string `form:"service" binding:"required"`                   // query service name
}

type QueryServiceNameRequest struct {
	Cluster   string              `json:"cluster"`                                      // query Cluster
	StartTime int64               `json:"startTime" binding:"required"`                 // query start time
	EndTime   int64               `json:"endTime" binding:"required,gtfield=StartTime"` // query end time
	Tags      QueryServiceNameTag `json:"tags"`
}

type QueryTopologyRequest struct {
	Cluster     string `form:"cluster"`                                      // query Cluster
	StartTime   int64  `form:"startTime" binding:"min=0"`                    // query start time
	EndTime     int64  `form:"endTime" binding:"required,gtfield=StartTime"` // query end time
	ServiceName string `form:"service" binding:"required"`                   // query service name
}

type QueryServiceNameTag struct {
	PodName     string `json:"pod"`
	ContainerId string `json:"containerId"`
	Pid         string `json:"pid"`
	NodeName    string `json:"nodeName"`
}

type CreateCustomTopologyRequest struct {
	ClusterId string `form:"clusterId" binding:"required"`
	LeftNode  string `form:"leftNode" binding:"required"`
	LeftType  string `form:"leftType" binding:"required"`
	RightNode string `form:"rightNode" binding:"required"`
	RightType string `form:"rightType" binding:"required"`
}

type ListCustomTopologyRequest struct {
	StartTime int64 `form:"startTime" binding:"required"`
	EndTime   int64 `form:"endTime" binding:"required,gtfield=StartTime"`
}

type DeleteCustomTopologyRequest struct {
	ID        int   `form:"id" binding:"required"`
	StartTime int64 `form:"startTime" binding:"required"`
	EndTime   int64 `form:"endTime" binding:"required,gtfield=StartTime"`
}

type SetServiceNameRuleRequest struct {
	RuleId      int                                  `json:"ruleId"`
	ServiceName string                               `json:"service"`
	ClusterId   string                               `json:"clusterId"`
	Conditions  []SetServiceNameRuleConditionRequest `json:"conditions"`
}

type SetServiceNameRuleConditionRequest struct {
	CondtiondId int    `json:"conditionId"`
	Key         string `json:"key"`
	MatchType   string `json:"matchType"`
	Value       string `json:"value"`
}
