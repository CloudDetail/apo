package custom

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// UpdateCustomQuery 更新用户自定义查询数据类型
// @Summary 更新用户自定义查询数据类型
// @Description 更新用户自定义查询数据类型
// @Tags API.custom
// @Accept json
// @Produce json
// @Param Request body request.UpdateCustomRequest true "请求信息"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/custom/update [post]
func (h *handler) UpdateCustomQuery() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.UpdateCustomRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}
		err := h.service.UpdateCustomQuery(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetCustomTypeError,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
