package custom

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// CetCustomQuery 获取用户自定义查询数据类型
// @Summary 获取用户自定义查询数据类型
// @Description 获取用户自定义查询数据类型
// @Tags API.custom
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.GetCustomRequest true "请求信息"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetCustomResponse
// @Failure 400 {object} code.Failure
// @Router /api/custom/get [get]
func (h *handler) GetCustomQuery() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetCustomRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				c.ErrMessage(code.ParamBindError)).WithError(err),
			)
			return
		}
		resp, err := h.service.GetCustomQuery(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetCustomTypeError,
				err.Error()).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
