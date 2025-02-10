package alertanalyze

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetDescendantContribution 获取下游故障贡献度
// @Summary 获取下游故障贡献度
// @Description 获取下游故障贡献度
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "查询开始时间"
// @Param endTime query int64 true "查询结束时间"
// @Param service query string true "查询服务名"
// @Param endpoint query string true "查询Endpoint"
// @Param step query int64 true "查询步长(us)"
// @Param entryService query string false "入口服务名"
// @Param entryEndpoint query string false "入口Endpoint"
// @Param sortBy query string false "排序字段,支持latency和errorRate,多个使用,链接"
// @Success 200 {object} response.GetDescendantAlertContributationResponse
// @Failure 400 {object} code.Failure
// @Router /api/alerts/descendant/anormal/contribution [get]
func (h *handler) GetDescendantContribution() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetDescendantAlertContributaionRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.alertanalyzeService.GetDescendantContribution(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.AlertAnalyzeDescendantAnormalContribution,
				code.Text(code.AlertAnalyzeDescendantAnormalContribution),
			).WithError(err))
			return
		}
		c.Payload(resp)
	}
}
