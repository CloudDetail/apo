package data

import (
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"net/http"
)

// GetDatasource Gets all datasource.
// @Summary Gets all datasource.
// @Description Gets all datasource.
// @Tags API.data
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetDatasourceResponse
// @Failure 400 {object} code.Failure
// @Router /api/data/datasource [get]
func (h *handler) GetDatasource() core.HandlerFunc {
	return func(c core.Context) {
		resp, err := h.dataService.GetDataSource()
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetDatasourceError,
				code.Text(code.GetDatasourceError)).WithError(err))
			return
		}
		c.Payload(resp)
	}
}
