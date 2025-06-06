package alertanalyze

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

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
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		err := h.alertanalyzeService.DetectDefects(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.DetectDefectsError,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
