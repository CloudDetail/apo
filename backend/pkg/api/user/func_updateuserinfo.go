package user

import (
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
// @Success 200 {object} response.UpdateUserInfoResponse
// @Failure 400 {object} code.Failure
// @Router /api/user/update/info [post]
func (h *handler) UpdateUserInfo() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.UpdateUserInfoRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

	}
}
