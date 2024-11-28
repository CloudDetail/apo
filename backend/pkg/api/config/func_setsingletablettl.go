package config

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// SetSingleTableTTL 配置单个表格的TTL
// @Summary  配置单个表格的TTL
// @Description  配置单个表格的TTL
// @Tags Api.config
// @Accept json
// @Produce json
// @Param Request body request.SetSingleTTLRequest true "请求信息"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200
// @Failure 400
// @Router /api/config/setSingleTableTTL [post]
func (h *handler) SetSingleTableTTL() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.SetSingleTTLRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		if err := h.configService.SetSingleTableTTL(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.SetSingleTableTTLError,
				code.Text(code.SetSingleTableTTLError)).WithError(err),
			)
			return
		}
		c.Payload("ok")
	}
}
