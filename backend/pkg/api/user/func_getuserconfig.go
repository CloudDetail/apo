package user

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetUserConfig Gets user's menu config and which route can access.
// @Summary Gets user's menu config and which route can access.
// @Description Get user's menu config.
// @Tags API.user
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param userId query int64 true "用户id"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetUserConfigResponse
// @Failure 400 {object} code.Failure
// @Router /api/user/config [get]
func (h *handler) GetUserConfig() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetUserConfigRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.userService.GetUserConfig(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetMenuConfigError,
				code.Text(code.GetMenuConfigError)).WithError(err))
			return
		}
		c.Payload(resp)
	}
}
