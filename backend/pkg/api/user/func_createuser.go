package user

import (
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// CreateUser 创建用户
// @Summary 创建用户
// @Description 创建用户
// @Tags API.user
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param username formData string true "用户名"
// @Param password formData string true "密码"
// @Param confirmPassword formData string true "确认密码"
// @Param roleList formData []int false "角色id" collectionFormat(multi)
// @Param email formData string false "邮箱"
// @Param phone formData string false "手机号"
// @Param corporation formData string false "组织"
// @Param Authorization header string false "Bearer 令牌"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/user/create [post]
func (h *handler) CreateUser() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.CreateUserRequest)
		if err := c.ShouldBindPostForm(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		if req.ConfirmPassword != req.Password {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.UserConfirmPasswordError,
				code.Text(code.UserConfirmPasswordError)),
			)
			return
		}

		if len(req.Phone) > 0 && !phoneRegexp.MatchString(req.Phone) {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.UserPhoneFormatError,
				code.Text(code.UserPhoneFormatError)))
			return
		}

		err := h.userService.CreateUser(req)
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
					code.UserCreateError,
					code.Text(code.UserCreateError),
				).WithError(err))
			}
			return
		}
		c.Payload("ok")
	}
}
