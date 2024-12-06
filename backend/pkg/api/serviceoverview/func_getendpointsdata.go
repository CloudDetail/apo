package serviceoverview

import (
	"net/http"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/services/serviceoverview"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

// GetEndPointsData 获取endpoints服务列表
// @Summary 获取endpoints服务列表
// @Description 获取endpoints服务列表
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "查询开始时间"
// @Param endTime query int64 true "查询结束时间"
// @Param step query int64 true "步长"
// @Param serviceName query []string false "服务名称" collectionFormat(multi)
// @Param namespace query []string true "命名空间" collectionFormat(multi)
// @Param endpointName query []string false "服务端点" collectionFormat(multi)
// @Param sortRule query int true "排序逻辑"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.ServiceEndPointsRes
// @Failure 400 {object} code.Failure
// @Router /api/service/endpoints [get]
func (h *handler) GetEndPointsData() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetEndPointsDataRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		var startTime time.Time
		var endTime time.Time
		startTime = time.UnixMicro(req.StartTime)
		endTime = time.UnixMicro(req.EndTime)
		step := time.Duration(req.Step * 1000)
		sortRule := serviceoverview.SortType(req.SortRule)

		filter := serviceoverview.EndpointsFilter{
			MultiService:   req.ServiceName,
			MultiEndpoint:  req.EndpointName,
			MultiNamespace: req.Namespace,
		}

		data, err := h.serviceoverview.GetServicesEndPointData(startTime, endTime, step, filter, sortRule)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetTop3UrlListError,
				code.Text(code.GetTop3UrlListError)).WithError(err),
			)
			return
		}

		if data == nil {
			data = []response.ServiceEndPointsRes{}
		}

		c.Payload(data)
	}
}
