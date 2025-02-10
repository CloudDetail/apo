package alertanalyze

import (
	"errors"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// AddDefectsDetect 执行异常检测
// @Summary 执行异常检测
// @Description 执行异常检测
// @Tags API.alerts
// @Accept json
// @Produce json
// @Param Request body request.DetectMutationRequest true "请求信息"
// @Success 200
// @Failure 400 {object} code.Failure
// @Router /api/alerts/detect/mutation/add [post]
func (h *handler) AddDefectsDetect() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.DetectMutationRequest)
		if err := c.ShouldBindJSON(req); err != nil || len(req.MutationCheck.CustomMetric) == 0 {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		err := h.alertanalyzeService.DetectDefects(req)
		if err != nil {
			var vErr model.ErrWithMessage
			if errors.As(err, &vErr) {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					vErr.Code,
					code.Text(vErr.Code)).WithError(err),
				)
				return
			} else {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					code.DetectDefectsError,
					code.Text(code.DetectDefectsError)).WithError(err),
				)
				return
			}
		}
		c.Payload("ok")
	}
}
