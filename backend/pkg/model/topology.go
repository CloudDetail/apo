// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package model

const (
	GROUP_SERVICE  = "service"
	GROUP_MQ       = "mq"
	GROUP_DB       = "db"
	GROUP_EXTERNAL = "external"
)

type TopologyNodes struct {
	Nodes map[string]*TopologyNode
}

func NewTopologyNodes() *TopologyNodes {
	return &TopologyNodes{
		Nodes: make(map[string]*TopologyNode),
	}
}

func (nodes *TopologyNodes) AddServerNode(key string, service string, url string, isTraced bool) {
	if _, exist := nodes.Nodes[key]; exist {
		return
	}
	nodes.Nodes[key] = NewServerNode(service, url, isTraced)
}

func (nodes *TopologyNodes) AddTopologyNode(key string, service string, url string, isTraced bool, group string, system string) {
	if _, exist := nodes.Nodes[key]; exist {
		return
	}

	nodes.Nodes[key] = &TopologyNode{
		Service:  service,
		Endpoint: url,
		IsTraced: isTraced,
		Group:    group,
		System:   system,
	}
}

func (nodes *TopologyNodes) GetNodes() []*TopologyNode {
	result := make([]*TopologyNode, 0)
	for _, node := range nodes.Nodes {
		result = append(result, node)
	}
	return result
}

func (nodes *TopologyNodes) GetLabels(group string) ([]string, []string, []string) {
	services := make([]string, 0)
	endpoints := make([]string, 0)
	systems := make([]string, 0)
	for _, node := range nodes.Nodes {
		if node.Group == group {
			services = append(services, node.Service)
			endpoints = append(endpoints, node.Endpoint)
			systems = append(systems, node.System)
		}
	}
	return services, endpoints, systems
}

type TopologyNode struct {
	Service  string `json:"service"`
	Endpoint string `json:"endpoint"`
	IsTraced bool   `json:"isTraced"`
	Group    string `json:"group"`
	System   string `json:"system"`
}

func NewServerNode(service string, url string, isTraced bool) *TopologyNode {
	return &TopologyNode{
		Service:  service,
		Endpoint: url,
		IsTraced: isTraced,
		Group:    GROUP_SERVICE,
		System:   "",
	}
}

type ToplogyRelation struct {
	ParentService  string `json:"parentService"`
	ParentEndpoint string `json:"parentEndpoint"`
	Service        string `json:"service"`
	Endpoint       string `json:"endpoint"`
	IsTraced       bool   `json:"isTraced"`
	Group          string `json:"group"`
	System         string `json:"system"`
}

func NewServerRelation(parentService, parentEndPoint, service, endpoint string, isTraced bool) *ToplogyRelation {
	return &ToplogyRelation{
		ParentService:  parentService,
		ParentEndpoint: parentEndPoint,
		Service:        service,
		Endpoint:       endpoint,
		IsTraced:       isTraced,
		Group:          GROUP_SERVICE,
		System:         "",
	}
}
