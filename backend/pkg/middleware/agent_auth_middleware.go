package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

const (
	agentClusterIDHeader = "X-Agent-Cluster-ID"
	agentNodeIDHeader    = "X-Agent-Node-ID"
)

func (m *middleware) AgentAuthMiddleware() core.HandlerFunc {
	return func(c core.Context) {
		clusterID := strings.TrimSpace(c.GetHeader(agentClusterIDHeader))
		if clusterID == "" {
			c.AbortWithError(http.StatusUnauthorized, code.UnAuth, errors.New("missing agent cluster identity"))
			return
		}
		nodeID := strings.TrimSpace(c.GetHeader(agentNodeIDHeader))
		c.Set("agentClusterID", clusterID)
		c.Set("agentNodeID", nodeID)
		c.Next()
	}
}
