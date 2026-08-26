package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CloudDetail/apo/backend/pkg/core"
	"go.uber.org/zap"
)

func TestAgentAuthMiddlewareRequiresClusterIdentity(t *testing.T) {
	tests := []struct {
		name      string
		clusterID string
		nodeID    string
		wantCode  int
	}{
		{name: "missing cluster", nodeID: "node-1", wantCode: http.StatusUnauthorized},
		{name: "cluster only", clusterID: "cluster-1", wantCode: http.StatusOK},
		{name: "cluster and node", clusterID: "cluster-1", nodeID: "node-1", wantCode: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux, err := core.New(zap.NewNop())
			if err != nil {
				t.Fatal(err)
			}
			var clusterID, nodeID string
			router := mux.Group("/agent").Use((&middleware{}).AgentAuthMiddleware())
			router.GET("/test", func(c core.Context) {
				cluster, _ := c.Get("agentClusterID")
				node, _ := c.Get("agentNodeID")
				clusterID, _ = cluster.(string)
				nodeID, _ = node.(string)
				c.Payload("ok")
			})

			req := httptest.NewRequest(http.MethodGet, "/agent/test", nil)
			req.Header.Set(agentClusterIDHeader, tt.clusterID)
			req.Header.Set(agentNodeIDHeader, tt.nodeID)
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, req)

			if response.Code != tt.wantCode {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
			if tt.wantCode == http.StatusOK && (clusterID != tt.clusterID || nodeID != tt.nodeID) {
				t.Fatalf("middleware identity=%q/%q", clusterID, nodeID)
			}
		})
	}
}
