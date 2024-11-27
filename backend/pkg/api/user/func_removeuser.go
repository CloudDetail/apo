package user

import (
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/middleware"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// RemoveUser 移除用户
// @Summary 移除用户
// @Description 移除用户
// @Tags API.user
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Authorization header string true "Bearer accessToken"
// @Param username formData string true "请求信息"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/user/remove [post]
func (h *handler) RemoveUser() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.RemoveUserRequest)
		if err := c.ShouldBindPostForm(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		username, _ := c.Get(middleware.UserKey)
		if username == req.Username {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.UserRemoveSelfError,
				code.Text(code.UserRemoveSelfError)))
			return
		}
		err := h.userService.RemoveUser(req.Username, username.(string))
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
					code.RemoveUserError,
					code.Text(code.RemoveUserError),
				).WithError(err))
			}
			return
		}
		c.Payload("ok")
	}
}
