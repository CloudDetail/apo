package log

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// DeleteLogParseRule 删除日志表解析规则
// @Summary 删除日志表解析规则
// @Description 删除日志表解析规则
// @Tags API.log
// @Accept json
// @Produce json
// @Param Request body request.DeleteLogParseRequest true "请求信息"
// @Success 200 {object} response.LogParseResponse
// @Failure 400 {object} code.Failure
// @Router /api/log/rule/delete [delete]
func (h *handler) DeleteLogParseRule() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.DeleteLogParseRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		resp, err := h.logService.DeleteLogParseRule(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.DeleteLogParseRuleError,
				code.Text(code.DeleteLogParseRuleError)+err.Error()).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
