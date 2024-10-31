package user

import (
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/middleware"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// UpdateUserPassword 更新密码
// @Summary 更新密码
// @Description 更新密码
// @Tags API.user
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.UpdateUserPasswordRequest true "请求信息"
// @Param Authorization header string true "Bearer 令牌"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/user/update/password [post]
func (h *handler) UpdateUserPassword() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.UpdateUserPasswordRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		username, _ := c.Get(middleware.UserKey)
		err := h.userService.UpdateUserPassword(username.(string), req)
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
					code.UserUpdateError,
					code.Text(code.UserUpdateError),
				).WithError(err))
			}
			return
		}
		c.Payload("ok")
	}
}
