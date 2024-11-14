package user

import (
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// Login 登录
// @Summary 登录
// @Description 登录
// @Tags API.user
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param username query string true "用户名"
// @Param password query string true "密码"
// @Success 200 {object} response.LoginResponse
// @Failure 400 {object} code.Failure
// @Router /api/user/login [post]
func (h *handler) Login() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.LoginRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.userService.Login(req)
		if err != nil {
			var vErr model.ErrWithMessage
			if errors.As(err, &vErr) {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					vErr.Code,
					code.Text(vErr.Code),
				).WithError(err))
			} else {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					code.UserLoginError,
					code.Text(code.UserLoginError),
				).WithError(err))
			}
			return
		}
		c.Payload(resp)
	}
}