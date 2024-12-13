package user

import (
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// PermissionOperation Grant or revoke user's permission(feature).
// @Summary Grant or revoke user's permission(feature).
// @Description Grant or revoke user's permission(feature).
// @Tags API.user
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param subjectId formData int64 true "授权主体id"
// @Param subjectType formData string true "授权主体类型: 'role','user','team'"
// @Param type formData string true "授权类型: 'feature','data'"
// @Param permissionList formData []int false "权限id列表" collectionFormat(multi)
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/user/permission/operation [post]
func (h *handler) PermissionOperation() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.PermissionOperationRequest)
		if err := c.ShouldBindPostForm(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		err := h.userService.PermissionOperation(req)
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
