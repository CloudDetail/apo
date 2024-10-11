package log

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetLogParseRule 更新日志表解析规则
// @Summary 更新日志表解析规则
// @Description 更新日志表解析规则
// @Tags API.log
// @Accept json
// @Produce json
// @Param Request body request.UpdateLogParseRequest true "请求信息"
// @Success 200 {object} response.LogParseResponse
// @Failure 400 {object} code.Failure
// @Router /api/log/rule/update [post]
func (h *handler) UpdateLogParseRule() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.UpdateLogParseRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		resp, err := h.logService.UpdateLogParseRule(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.UpdateLogParseRuleError,
				code.Text(code.UpdateLogParseRuleError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
