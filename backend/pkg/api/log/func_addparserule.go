package log

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// AddLogParseRule 新增日志表解析规则
// @Summary 新增日志表解析规则
// @Description 新增日志表解析规则
// @Tags API.log
// @Accept json
// @Produce json
// @Param Request body request.AddLogParseRequest true "请求信息"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.LogParseResponse
// @Failure 400 {object} code.Failure
// @Router /api/log/rule/add [post]
func (h *handler) AddLogParseRule() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.AddLogParseRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		resp, err := h.logService.AddLogParseRule(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.AddLogParseRuleError,
				code.Text(code.AddLogParseRuleError)+err.Error()).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
