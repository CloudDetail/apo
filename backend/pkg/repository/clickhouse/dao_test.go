package clickhouse

import (
	"os"
	"testing"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/logger"
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

	testListParentNodes(t, repo)     // 查询上游节点列表
	testListChildNodes(t, repo)      // 查询下游节点列表
	testDescendantNodes(t, repo)     // 查询所有子孙节点列表
	testDescendantRelations(t, repo) // 查询下游节点调用关系列表

	// testCountK8sEvents(t, repo) // 计算K8s事件数量
}

func testListParentNodes(t *testing.T, repo Repo) {
	req := &request.GetServiceEndpointTopologyRequest{
		StartTime: 1722244803099000,
		EndTime:   1722245703099000,
		Service:   "ts-seat-service-ts-seat-service",
		Endpoint:  "POST /api/v1/seatservice/seats/left_tickets",
	}
	resp, err := repo.ListParentNodes(req)
	if err != nil {
		t.Errorf("Error to list parent nodes: %v", err)
	}
	expect := []TopologyNode{
		{
			Service:  "ts-travel-service-ts-travel-service",
			Endpoint: "POST /api/v1/travelservice/trips/left",
			IsTraced: false,
		},
		{
			Service:  "ts-travel2-service-ts-travel2-service",
			Endpoint: "POST /api/v1/travel2service/trips/left",
			IsTraced: false,
		},
	}
	checkTopologyNodes(util.NewValidator(t, "ListParentNodes"), expect, resp)
}

func testListChildNodes(t *testing.T, repo Repo) {
	req := &request.GetServiceEndpointTopologyRequest{
		StartTime: 1722244803099000,
		EndTime:   1722245703099000,
		Service:   "ts-seat-service-ts-seat-service",
		Endpoint:  "POST /api/v1/seatservice/seats/left_tickets",
	}
	resp, err := repo.ListChildNodes(req)
	if err != nil {
		t.Errorf("Error to list child nodes: %v", err)
	}
	expect := []TopologyNode{
		{
			Service:  "ts-config-service-ts-config-service",
			Endpoint: "GET /api/v1/configservice/configs/{configName}",
			IsTraced: false,
		},
		{
			Service:  "ts-order-other-service-ts-order-other-service",
			Endpoint: "POST /api/v1/orderOtherService/orderOther/tickets",
			IsTraced: false,
		},
		{
			Service:  "ts-order-service-ts-order-service",
			Endpoint: "POST /api/v1/orderservice/order/tickets",
			IsTraced: false,
		},
	}
	checkTopologyNodes(util.NewValidator(t, "ListChildNodes"), expect, resp)
}

func testDescendantNodes(t *testing.T, repo Repo) {
	req := &request.GetDescendantMetricsRequest{
		StartTime: 1722244803099000,
		EndTime:   1722245703099000,
		Service:   "ts-seat-service-ts-seat-service",
		Endpoint:  "POST /api/v1/seatservice/seats/left_tickets",
	}
	resp, err := repo.ListDescendantNodes(req)
	if err != nil {
		t.Errorf("Error to list child nodes: %v", err)
	}
	expect := []TopologyNode{
		{
			Service:  "ts-config-service-ts-config-service",
			Endpoint: "GET /api/v1/configservice/configs/{configName}",
			IsTraced: false,
		},
		{
			Service:  "ts-order-other-service-ts-order-other-service",
			Endpoint: "POST /api/v1/orderOtherService/orderOther/tickets",
			IsTraced: false,
		},
		{
			Service:  "ts-order-service-ts-order-service",
			Endpoint: "POST /api/v1/orderservice/order/tickets",
			IsTraced: false,
		},
	}
	checkTopologyNodes(util.NewValidator(t, "ListDescendantNodes"), expect, resp)
}

func testDescendantRelations(t *testing.T, repo Repo) {
	req := &request.GetServiceEndpointRelationRequest{
		StartTime: 1723514677000000,
		EndTime:   1723515577000000,
		Service:   "ts-seat-service",
		Endpoint:  "POST /api/v1/seatservice/seats/left_tickets",
	}
	resp, err := repo.ListDescendantRelations(req)
	if err != nil {
		t.Errorf("Error to list descendant relation: %v", err)
	}
	expect := []ToplogyRelation{
		{
			ParentService:  "ts-seat-service",
			ParentEndpoint: "POST /api/v1/seatservice/seats/left_tickets",
			Service:        "ts-order-service",
			Endpoint:       "POST /api/v1/orderservice/order/tickets",
			IsTraced:       true,
		},

		{
			ParentService:  "ts-seat-service",
			ParentEndpoint: "POST /api/v1/seatservice/seats/left_tickets",
			Service:        "ts-order-other-service",
			Endpoint:       "POST /api/v1/orderOtherService/orderOther/tickets",
			IsTraced:       true,
		},
		{
			ParentService:  "ts-seat-service",
			ParentEndpoint: "POST /api/v1/seatservice/seats/left_tickets",
			Service:        "ts-config-service",
			Endpoint:       "GET /api/v1/configservice/configs/{configName}",
			IsTraced:       true,
		},
	}
	validator := util.NewValidator(t, "ListDescendantTopology").
		CheckIntValue("Response Size", len(expect), len(resp))
	for i, gotTopology := range resp {
		expectTopology := expect[i]
		validator.
			CheckStringValue("parentService", expectTopology.ParentService, gotTopology.ParentService).
			CheckStringValue("parentEndpoint", expectTopology.ParentEndpoint, gotTopology.ParentEndpoint).
			CheckStringValue("service", expectTopology.Service, gotTopology.Service).
			CheckStringValue("endpoint", expectTopology.Endpoint, gotTopology.Endpoint).
			CheckBoolValue("isTraced", expectTopology.IsTraced, gotTopology.IsTraced)
	}
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

func checkTopologyNodes(validator *util.Validator, expect []TopologyNode, got []TopologyNode) {
	validator.CheckIntValue("Response Size", len(expect), len(got))
	for i, gotNode := range got {
		expectNode := expect[i]
		validator.
			CheckStringValue("service", expectNode.Service, gotNode.Service).
			CheckStringValue("endpoint", expectNode.Endpoint, gotNode.Endpoint).
			CheckBoolValue("isTraced", expectNode.IsTraced, gotNode.IsTraced)
	}
}
