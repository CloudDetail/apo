package custom

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// DeleteCustomQuery 删除用户自定义查询数据类型
// @Summary 删除用户自定义查询数据类型
// @Description 删除用户自定义查询数据类型
// @Tags API.custom
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.DeleteCustomRequest true "请求信息"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/custom/delete [delete]
func (h *handler) DeleteCustomQuery() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.DeleteCustomRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}
		err := h.service.DeleteCustomQuery(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.DeleteCustomTypeError,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
