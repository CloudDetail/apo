package user

import (
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// RoleOperation Grant or revoke user's role.
// @Summary Grant or revoke user's role.
// @Description Grants permission to user
// @Tags API.permission
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param userId formData int64 ture "用户id"
// @Param roleList formData []int ture "角色id" collectionFormat(multi)
// @Param Authorization header string true "Bearer accessToken"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/permission/role/operation [post]
func (h *handler) RoleOperation() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.RoleOperationRequest)
		if err := c.ShouldBindPostForm(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		err := h.userService.RoleOperation(req)
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
					code.UserGrantRoleError,
					code.Text(code.UserGrantRoleError),
				).WithError(err))
			}
			return
		}
		c.Payload("ok")
	}
}
