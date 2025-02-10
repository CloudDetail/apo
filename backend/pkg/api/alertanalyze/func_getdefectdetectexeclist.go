package alertanalyze

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetDefectDetectExecList 获取异常检测执行记录
// @Summary 获取异常检测执行记录
// @Description 获取异常检测执行记录
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param currentPage query int false "当前页"
// @Param pageSize query int false "页大小"
// @Success 200 {object} response.GetDefectDetectExecListResponse
// @Failure 400 {object} code.Failure
// @Router /api/alerts/detect/mutation/exec-list [get]
func (h *handler) GetDefectDetectExecList() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetDefectDetectExecListRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		if req.PageParam == nil {
			req.PageSize = 10
			req.CurrentPage = 1
		}

		resp, err := h.alertanalyzeService.GetDefectDetectExecList(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetDetectMutationExecListError,
				code.Text(code.GetDetectMutationExecListError)).WithError(err))
			return
		}
		c.Payload(resp)
	}
}
