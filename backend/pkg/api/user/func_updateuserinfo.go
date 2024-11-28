package user

import (
	"errors"
	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// UpdateUserInfo 更新个人信息
// @Summary 更新个人信息
// @Description 更新个人信息
// @Tags API.user
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param username formData string true "用户名"
// @Param corporation formData string false "组织"
// @Param phone formData string false "手机号"
// @Param email formData string false "邮箱"
// @Param Authorization header string true "Bearer accessToken"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/user/update/info [post]
func (h *handler) UpdateUserInfo() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.UpdateUserInfoRequest)
		if err := c.ShouldBindPostForm(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		if req.Username == config.Get().User.AnonymousUser.Username {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.UserNoPermissionError,
				code.Text(code.UserNoPermissionError)))
			return
		}

		err := h.userService.UpdateUserInfo(req)
		if err != nil {
			var vErr model.ErrWithMessage
			if errors.As(err, &vErr) {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					vErr.Code,
					code.Text(vErr.Code)).WithError(err))
			} else {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					code.UserUpdateError,
					code.Text(code.UserUpdateError)).WithError(err))
			}
			return
		}
		c.Payload("ok")
	}
}
