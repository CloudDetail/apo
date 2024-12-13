package user

import (
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"net/http"
)

// GetRole Gets all roles.
// @Summary Gets all roles.
// @Description Gets all roles.
// @Tags API.user
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Authorization header string true "Bearer accessToken"
// @Success 200 {object} response.GetRoleResponse
// @Failure 400 {object} code.Failure
// @Router /api/user/roles [get]
func (h *handler) GetRole() core.HandlerFunc {
	return func(c core.Context) {

		resp, err := h.userService.GetRoles()
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.UserGetRolesERROR,
				code.Text(code.UserGetRolesERROR)))
		}
		c.Payload(resp)
	}
}
