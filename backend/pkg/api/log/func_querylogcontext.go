package log

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// QueryLogContext 获取日志上下文
// @Summary 获取日志上下文
// @Description 获取日志上下文
// @Tags API.log
// @Accept json
// @Produce json
// @Param Request body request.LogQueryContextRequest true "请求信息"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.LogQueryContextResponse
// @Failure 400 {object} code.Failure
// @Router /api/log/context [post]
func (h *handler) QueryLogContext() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.LogQueryContextRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		resp, err := h.logService.QueryLogContext(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.QueryLogContextError,
				code.Text(code.QueryLogContextError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
