package data

import (
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/middleware"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"net/http"
)

// GetGroupDatasource Get group's datasource.
// @Summary Get group's datasource.
// @Description Get group's datasource.
// @Tags API.permission
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param groupId query int64 false "Data group's id"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetGroupDatasourceResponse
// @Failure 400 {object} code.Failure
// @Router /api/data/group/data [get]
func (h *handler) GetGroupDatasource() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetGroupDatasourceRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		userID := middleware.GetContextUserID(c)
		resp, err := h.dataService.GetGroupDatasource(req, userID)
		if err != nil {
			c.HandleError(err, code.GetGroupDatasourceError)
			return
		}
		c.Payload(resp)
	}
}
