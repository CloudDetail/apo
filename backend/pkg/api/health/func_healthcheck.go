package health

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// HealthCheck 用于k8s检查后端健康状态
// @Summary 用于k8s检查后端健康状态
// @Description 用于k8s检查后端健康状态
// @Tag API.health
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/health [get]
func (h *handler) HealthCheck() core.HandlerFunc {
	return func(c core.Context) {
		c.Payload("ok")
	}
}
