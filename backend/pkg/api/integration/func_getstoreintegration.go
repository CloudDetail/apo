package integration

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetStoreIntegration
// @Summary
// @Description
// @Tags API.integration
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Success 200 {object} response.getStoreIntegrationResponse
// @Failure 400 {object} code.Failure
// @Router /api/integration/configuration [get]
func (h *handler) GetStoreIntegration() core.HandlerFunc {
	return func(c core.Context) {
		storeIntegration := h.integrationService.GetDatasourceAndDatabase()
		c.Payload(storeIntegration)
	}
}
