package log

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetLogTableInfo
// @Summary 获取日志表信息
// @Description 获取日志表信息
// @Tags API.log
// @Accept json
// @Produce json
// @Param Request body request.LogTableInfoRequest true "请求信息"
// @Success 200 {object} response.LogTableInfoResponse
// @Failure 400 {object} code.Failure
// @Router /api/log/table [get]
func (h *handler) GetLogTableInfo() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.LogTableInfoRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		resp, err := h.logService.GetLogTableInfo(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetLogTableInfoError,
				code.Text(code.GetLogTableInfoError)+err.Error()).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
