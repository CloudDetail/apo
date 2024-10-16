package log

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetLogParseRule 获取日志表解析规则
// @Summary 获取日志表解析规则
// @Description 获取日志表解析规则
// @Tags API.log
// @Accept json
// @Produce json
// @Param Request body request.QueryLogParseRequest true "请求信息"
// @Success 200 {object} response.LogParseResponse
// @Failure 400 {object} code.Failure
// @Router /api/log/rule/get [get]
func (h *handler) GetLogParseRule() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.QueryLogParseRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		resp, err := h.logService.GetLogParseRule(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetLogParseRuleError,
				code.Text(code.GetLogParseRuleError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
