// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"os"
	"testing"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/logger"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/util"
)

func init() {
	os.Setenv("APO_CONFIG", "../../../config/apo.yml")
}

func TestRepo(t *testing.T) {
	cfg := config.Get().ClickHouse
	zapLog := logger.NewLogger(logger.WithLevel("debug"))
	repo, err := New(zapLog, []string{cfg.Address}, cfg.Database, cfg.Username, cfg.Password)
	if err != nil {
		t.Fatalf("Error to connect clickhouse: %v", err)
	}

	testListParentNodes(t, repo) // Query the list of upstream nodes
	testKafkaListParentNodes(t, repo)
	testListChildNodes(t, repo)           // Query the list of downstream nodes
	testDescendantNodes(t, repo)          // query the list of all descendant nodes
	testDescendantRelations(t, repo)      // Query the call relationship list of downstream nodes
	testKafkaDescendantRelations(t, repo) // Query the call relationship list of downstream nodes

	// testCountK8sEvents(t, repo) // count the number of K8s events
}

func testListParentNodes(t *testing.T, repo Repo) {
	req := &request.GetServiceEndpointTopologyRequest{
		StartTime: 1732550400000000,
		EndTime:   1732636800000000,
		Service:   "stuck-demo-undertow",
		Endpoint:  "GET /db/query",
	}
	resp, err := repo.ListParentNodes(req)
	if err != nil {
		t.Errorf("Error to list parent nodes: %v", err)
	}
	expect := map[string]*model.TopologyNode{
		"": {
			Service:  "stuck-demo-tomcat",
			Endpoint: "GET /wait/callOthers",
			IsTraced: true,
			Group:    model.GROUP_SERVICE,
			System:   "",
		},
	}
	checkTopologyNodes(util.NewValidator(t, "ListParentNodes"), expect, resp.Nodes)
}

func testKafkaListParentNodes(t *testing.T, repo Repo) {
	req := &request.GetServiceEndpointTopologyRequest{
		StartTime: 1732550400000000,
		EndTime:   1732636800000000,
		Service:   "kafka-consumer",
		Endpoint:  "topic_login process",
	}
	resp, err := repo.ListParentNodes(req)
	if err != nil {
		t.Errorf("Error to list parent nodes: %v", err)
	}
	expect := map[string]*model.TopologyNode{
		"": {
			Service:  "",
			Endpoint: "topic_login",
			IsTraced: false,
			Group:    model.GROUP_MQ,
			System:   "kafka",
		},
	}
	checkTopologyNodes(util.NewValidator(t, "ListParentNodes"), expect, resp.Nodes)
}

func testListChildNodes(t *testing.T, repo Repo) {
	req := &request.GetServiceEndpointTopologyRequest{
		StartTime: 1732550400000000,
		EndTime:   1732636800000000,
		Service:   "stuck-demo-tomcat",
		Endpoint:  "GET /redis/query",
	}
	resp, err := repo.ListChildNodes(req)
	if err != nil {
		t.Errorf("Error to list child nodes: %v", err)
	}
	expect := map[string]*model.TopologyNode{
		"": {
			Service:  "10.0.2.15:6379",
			Endpoint: "GET",
			IsTraced: false,
			Group:    "db",
			System:   "redis",
		},
		"1": {
			Service:  "10.0.2.15:6379",
			Endpoint: "SET",
			IsTraced: false,
			Group:    "db",
			System:   "redis",
		},
		"2": {
			Service:  "10.0.2.15:6379",
			Endpoint: "EXISTS",
			IsTraced: false,
			Group:    "db",
			System:   "redis",
		},
	}
	checkTopologyNodes(util.NewValidator(t, "ListChildNodes"), expect, resp.Nodes)
}

func testDescendantNodes(t *testing.T, repo Repo) {
	req := &request.GetDescendantMetricsRequest{
		StartTime: 1732550400000000,
		EndTime:   1732636800000000,
		Service:   "kafka-provider",
		Endpoint:  "GET /send",
	}
	resp, err := repo.ListDescendantNodes(req)
	if err != nil {
		t.Errorf("Error to list child nodes: %v", err)
	}
	expect := map[string]*model.TopologyNode{
		"": {
			Service:  "",
			Endpoint: "topic_login",
			IsTraced: false,
			Group:    "mq",
			System:   "kafka",
		},
		"2": {
			Service:  "kafka-consumer",
			Endpoint: "topic_login process",
			IsTraced: true,
			Group:    model.GROUP_SERVICE,
			System:   "",
		},
	}
	checkTopologyNodes(util.NewValidator(t, "ListDescendantNodes"), expect, resp.Nodes)
}

func testDescendantRelations(t *testing.T, repo Repo) {
	req := &request.GetServiceEndpointRelationRequest{
		StartTime: 1732550400000000,
		EndTime:   1732636800000000,
		Service:   "stuck-demo-tomcat",
		Endpoint:  "GET /wait/callOthers",
	}
	resp, err := repo.ListDescendantRelations(req)
	if err != nil {
		t.Errorf("Error to list descendant relation: %v", err)
	}
	expect := []*model.ToplogyRelation{
		{
			ParentService:  "stuck-demo-tomcat",
			ParentEndpoint: "GET /wait/callOthers",
			Service:        "stuck-demo-undertow",
			Endpoint:       "GET /db/query",
			IsTraced:       true,
			Group:          model.GROUP_SERVICE,
			System:         "",
		},
		{
			ParentService:  "stuck-demo-tomcat",
			ParentEndpoint: "GET /wait/callOthers",
			Service:        "stuck-demo-undertow",
			Endpoint:       "GET /cpu/loop/{times}",
			IsTraced:       true,
			Group:          model.GROUP_SERVICE,
			System:         "",
		},
		{
			ParentService:  "stuck-demo-undertow",
			ParentEndpoint: "GET /db/query",
			Service:        "10.0.2.4:3306",
			Endpoint:       "SELECT test.weather",
			IsTraced:       false,
			Group:          "db",
			System:         "mysql",
		},
	}
	checkToplogyRelation(util.NewValidator(t, "ListDescendantTopology"), expect, resp)
}

func testKafkaDescendantRelations(t *testing.T, repo Repo) {
	req := &request.GetServiceEndpointRelationRequest{
		StartTime: 1732550400000000,
		EndTime:   1732636800000000,
		Service:   "kafka-provider",
		Endpoint:  "GET /send",
	}
	resp, err := repo.ListDescendantRelations(req)
	if err != nil {
		t.Errorf("Error to list descendant relation: %v", err)
	}
	expect := []*model.ToplogyRelation{
		{
			ParentService:  "kafka-provider",
			ParentEndpoint: "GET /send",
			Service:        "",
			Endpoint:       "topic_login",
			IsTraced:       false,
			Group:          model.GROUP_MQ,
			System:         "kafka",
		},
		{
			ParentService:  "",
			ParentEndpoint: "topic_login",
			Service:        "kafka-consumer",
			Endpoint:       "topic_login process",
			IsTraced:       true,
			Group:          model.GROUP_SERVICE,
			System:         "",
		},
	}
	checkToplogyRelation(util.NewValidator(t, "ListDescendantTopology"), expect, resp)
}

func testCountK8sEvents(t *testing.T, repo Repo) {
	pods := []string{"ts-travel2-service-fdbbd5946-l4h2r"}
	count, err := repo.CountK8sEvents(1722244803099000, 1722245703099000, pods)
	if err != nil {
		t.Errorf("Error to get k8s events: %v", err)
	}
	util.NewValidator(t, "CountK8sEvents").
		CheckIntValue("Count Size", 10, len(count))
}

func checkTopologyNodes(validator *util.Validator, expect map[string]*model.TopologyNode, got map[string]*model.TopologyNode) {
	validator.CheckIntValue("Response Size", len(expect), len(got))
	for key, gotNode := range got {
		expectNode := expect[key]
		validator.
			CheckStringValue("service", expectNode.Service, gotNode.Service).
			CheckStringValue("endpoint", expectNode.Endpoint, gotNode.Endpoint).
			CheckStringValue("group", expectNode.Group, gotNode.Group).
			CheckStringValue("system", expectNode.System, gotNode.System).
			CheckBoolValue("isTraced", expectNode.IsTraced, gotNode.IsTraced)
	}
}

func checkToplogyRelation(validator *util.Validator, expect []*model.ToplogyRelation, got []*model.ToplogyRelation) {
	validator.CheckIntValue("Response Size", len(expect), len(got))
	for i, gotTopology := range got {
		expectTopology := expect[i]
		validator.
			CheckStringValue("parentService", expectTopology.ParentService, gotTopology.ParentService).
			CheckStringValue("parentEndpoint", expectTopology.ParentEndpoint, gotTopology.ParentEndpoint).
			CheckStringValue("service", expectTopology.Service, gotTopology.Service).
			CheckStringValue("endpoint", expectTopology.Endpoint, gotTopology.Endpoint).
			CheckStringValue("group", expectTopology.Group, gotTopology.Group).
			CheckStringValue("system", expectTopology.System, gotTopology.System).
			CheckBoolValue("isTraced", expectTopology.IsTraced, gotTopology.IsTraced)
	}
}
