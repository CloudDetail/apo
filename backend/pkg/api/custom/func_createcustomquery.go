package custom

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// CreateCustomQuery 创建一个用户自定义查询数据类型
// @Summary 创建一个用户自定义查询数据类型
// @Description 创建一个用户自定义查询数据类型
// @Tags API.custom
// @Accept json
// @Produce json
// @Param Request body request.CreateCustomRequest true "请求信息"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/custom/create [post]
func (h *handler) CreateCustomQuery() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.CreateCustomRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}
		err := h.service.CreateCustomQuery(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.CreateCustomTypeError,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
