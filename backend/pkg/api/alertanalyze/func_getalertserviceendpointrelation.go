package alertanalyze

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetAlertServiceEndpointRelation 获取告警分析中的服务相关性
// @Summary 获取告警分析中的服务相关性
// @Description 获取告警分析中的服务相关性
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query uint64 true "查询开始时间"
// @Param endTime query uint64 true "查询结束时间"
// @Param service query string true "查询服务名"
// @Param endpoint query string true "查询Endpoint"
// @Param entryService query string false "入口服务名"
// @Param entryEndpoint query string false "入口Endpoint"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetServiceEndpointRelationResponse
// @Failure 400 {object} code.Failure
// @Router /api/alerts/service/relation [get]
func (h *handler) GetAlertServiceEndpointRelation() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetAlertServiceEndpointRelationRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.alertanalyzeService.GetAnalyzeServiceEndpointRelation(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetServiceUrlRelationError,
				code.Text(code.GetServiceUrlRelationError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
