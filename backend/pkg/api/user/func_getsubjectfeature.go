package user

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetSubjectFeature Gets subject's feature permission.
// @Summary Gets subject's permission.
// @Description Gets subject's permission.
// @Tags API.permission
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param subjectId query int64 true "授权主体id"
// @Param subjectType query string true "授权主体类型"
// @Success 200 {object} response.GetSubjectFeatureResponse
// @Failure 400 {object} code.Failure
// @Router /api/permission/sub/feature [get]
func (h *handler) GetSubjectFeature() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetSubjectFeatureRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.userService.GetSubjectFeature(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetFeatureError,
				code.Text(code.GetFeatureError)).WithError(err))
		}
		c.Payload(resp)
	}
}
