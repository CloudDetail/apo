package log

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// OtherTable 获取外部日志表信息
// @Summary 获取外部日志表信息
// @Description 获取外部日志表信息
// @Tags API.log
// @Accept json
// @Produce json
// @Param Request body request.OtherTableInfoRequest true "请求信息"
// @Success 200 {object} response.OtherTableInfoResponse
// @Failure 400 {object} code.Failure
// @Router /api/log/other/table [post]
func (h *handler) OtherTableInfo() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.OtherTableInfoRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		resp, err := h.logService.OtherTableInfo(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetOtherLogTableError,
				code.Text(code.GetOtherLogTableError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
