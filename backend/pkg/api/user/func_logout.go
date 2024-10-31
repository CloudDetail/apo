package user

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

// Logout 退出登录
// @Summary 退出登录
// @Description 退出登录
// @Tags API.user
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.LogoutRequest true "请求信息"
// @Success 200 {object} response.LogoutResponse
// @Failure 400 {object} code.Failure
// @Router /api/user/logout [get]
func (h *handler) Logout() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.LogoutRequest)
		// TODO 根据请求参数类型调整API
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		// TODO 替换为Service调用
		resp := new(response.LogoutResponse)
		c.Payload(resp)
	}
}
