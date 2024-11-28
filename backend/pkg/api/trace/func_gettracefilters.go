package trace

import (
	"net/http"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetTraceFilters 查询Trace列表可用的过滤器
// @Summary 查询Trace列表可用的过滤器
// @Description 查询Trace列表可用的过滤器
// @Tags API.trace
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "查询开始时间"
// @Param endTime query int64 true "查询结束时间"
// @Param needUpdate query bool false "是否根据用户输入的时间立即更新可用过滤器"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetTraceFiltersResponse
// @Failure 400 {object} code.Failure
// @Router /api/trace/pagelist/filters [get]
func (h *handler) GetTraceFilters() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetTraceFiltersRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		startTime := time.UnixMicro(req.StartTime)
		endTime := time.UnixMicro(req.EndTime)
		resp, err := h.traceService.GetTraceFilters(startTime, endTime, req.NeedUpdate)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetTraceFiltersError,
				code.Text(code.GetTraceFiltersError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
