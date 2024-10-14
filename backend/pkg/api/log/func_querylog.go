package log

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// QueryLog 查询全量日志
// @Summary 查询全量日志日志
// @Description 查询全量日志
// @Tags API.log
// @Accept json
// @Produce json
// @Param Request body request.LogQueryRequest true "请求信息"
// @Success 200 {object} response.LogQueryResponse
// @Failure 400 {object} code.Failure
// @Router /api/log/query [post]
func (h *handler) QueryLog() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.LogQueryRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		if req.PageNum == 0 {
			req.PageNum = 1
		}
		if req.PageSize == 0 {
			req.PageSize = 10
		}
		if req.Query == "" {
			req.Query = "(1='1')"
		}
		if req.TimeField == "" {
			req.TimeField = "timestamp"
		}
		if req.LogField == "" {
			req.LogField = "content"
		}
		resp, err := h.logService.QueryLog(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.QueryLogError,
				code.Text(code.QueryLogError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
