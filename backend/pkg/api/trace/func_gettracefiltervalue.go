package trace

import (
	"net/http"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetTraceFilterValue 查询指定过滤器的可用数值
// @Summary 查询指定过滤器的可用数值
// @Description 查询指定过滤器的可用数值
// @Tags API.trace
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.GetTraceFilterValueRequest true "请求信息"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetTraceFilterValueResponse
// @Failure 400 {object} code.Failure
// @Router /api/trace/pagelist/filter/value [post]
func (h *handler) GetTraceFilterValue() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetTraceFilterValueRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		startTime := time.UnixMicro(req.StartTime)
		endTime := time.UnixMicro(req.EndTime)
		resp, err := h.traceService.GetTraceFilterValues(startTime, endTime, req.SearchText, req.Filter)

		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetTraceFilterValueError,
				code.Text(code.GetTraceFilterValueError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
