package alertanalyze

import (
	"net/http"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

// GetAlertImpact 获取告警数据的影响面
// @Summary 获取告警数据的影响面
// @Description 获取告警数据的影响面
// @Tags API.alerts
// @Accept json
// @Produce json
// @Param Request body request.AlertImpactRequest true "请求信息"
// @Success 200 {object} response.GetAlertImpactResponse
// @Failure 400 {object} code.Failure
// @Router /api/alerts/event/impact [post]
func (h *handler) GetAlertImpact() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.AlertImpactRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		entryServices, err := h.alertanalyzeService.GetAlertsImpact(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.AlertEventImpactError,
				err,
			)
			return
		}

		startTime := time.UnixMicro(req.StartTime)
		endTime := time.UnixMicro(req.EndTime)

		// fill RED metric and namespace list
		// filter by namespace
		entries, _ := h.serviceoverviewService.GetAlertRelatedEntryData(
			c,
			startTime, endTime, req.Namespaces,
			entryServices)

		c.Payload(response.GetAlertImpactResponse{
			Data: entries,
		})
	}
}
