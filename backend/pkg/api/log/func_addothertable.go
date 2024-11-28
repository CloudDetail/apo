package log

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// AddOtherTable 添加外部日志表
// @Summary 添加外部日志表
// @Description 添加外部日志表
// @Tags API.log
// @Accept json
// @Produce json
// @Param Request body request.AddOtherTableRequest true "请求信息"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.AddOtherTableResponse
// @Failure 400 {object} code.Failure
// @Router /api/log/other/add [post]
func (h *handler) AddOtherTable() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.AddOtherTableRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		req.FillerValue()
		resp, err := h.logService.AddOtherTable(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.AddOtherLogTableError,
				code.Text(code.AddOtherLogTableError)+err.Error()).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
