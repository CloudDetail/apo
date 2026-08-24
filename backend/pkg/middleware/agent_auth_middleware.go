package middleware

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

func (m *middleware) AgentAuthMiddleware() core.HandlerFunc {
	return func(c core.Context) {
		expected := os.Getenv("AGENT_CONFIG_API_KEY")
		provided := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		if expected == "" || len(expected) != len(provided) || subtle.ConstantTimeCompare([]byte(expected), []byte(provided)) != 1 {
			c.AbortWithError(http.StatusUnauthorized, code.UnAuth, errors.New("invalid agent credential"))
			return
		}
		nodeID := strings.TrimSpace(c.GetHeader("X-Agent-Node-ID"))
		if nodeID == "" {
			c.AbortWithError(http.StatusUnauthorized, code.UnAuth, errors.New("missing agent node identity"))
			return
		}
		c.Set("agentNodeID", nodeID)
		c.Next()
	}
}
