package user

import (
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"net/http"
	"regexp"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// UpdateUserPhone 更新/绑定手机号
// @Summary 更新/绑定手机号
// @Description 更新/绑定手机号
// @Tags API.user
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param userId formData int64 true "用户id"
// @Param phone formData string true "手机号"
// @Param Authorization header string true "Bearer accessToken"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/user/update/phone [post]
func (h *handler) UpdateUserPhone() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.UpdateUserPhoneRequest)
		if err := c.ShouldBindPostForm(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		if !phoneRegexp.MatchString(req.Phone) {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)))
			return
		}

		err := h.userService.UpdateUserPhone(req)
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

var phoneRegexp = regexp.MustCompile("^1[3-9]\\d{9}$")
