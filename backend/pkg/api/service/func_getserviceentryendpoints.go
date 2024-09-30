package service

import (
	"net/http"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/services/serviceoverview"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

// GetServiceEntryEndpoints 获取服务入口Endpoint列表
// @Summary 获取服务入口Endpoint列表
// @Description 获取服务入口Endpoint列表
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query uint64 true "查询开始时间"
// @Param endTime query uint64 true "查询结束时间"
// @Param service query string true "查询服务名"
// @Param endpoint query string true "查询Endpoint"
// @Param step query int64 true "查询步长(us)"
// @Success 200 {object} []response.GetServiceEntryEndpointsResponse
// @Failure 400 {object} code.Failure
// @Router /api/service/entry/endpoints [get]
func (h *handler) GetServiceEntryEndpoints() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetServiceEntryEndpointsRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		var (
			err           error
			threshold     response.GetThresholdResponse
			endpointResps []response.ServiceEndPointsRes
			alertResps    []response.ServiceAlertRes
		)

		result := make(map[string]*response.EntryInstanceData, 0)
		resp := response.GetServiceEntryEndpointsResponse{
			Status: model.STATUS_NORMAL,
			Data:   make([]*response.EntryInstanceData, 0),
		}
		entryNodes, err := h.serviceInfoService.GetServiceEntryEndpoints(req)
		if err == nil {
			// TODO 先默认全局Threshold，后续调整为具体服务的Threshold
			threshold, err = h.serviceoverviewService.GetThreshold(database.GLOBAL, "", "")
		}
		if err == nil {
			startTime := time.UnixMicro(req.StartTime)
			endTime := time.UnixMicro(req.EndTime)
			sortRule := serviceoverview.DODThreshold
			step := time.Duration(req.Step * 1000)

			for _, entryNode := range entryNodes {
				filter := serviceoverview.EndpointsFilter{
					ContainsSvcName:      entryNode.Service,
					ContainsEndpointName: entryNode.Endpoint,
					Namespace:            "",
				}
				endpointResps, err = h.serviceoverviewService.GetServicesEndPointData(startTime, endTime, step, filter, sortRule)
				if err == nil {
					for _, endpointResp := range endpointResps {
						if serviceResp, found := result[endpointResp.ServiceName]; found {
							serviceResp.Namespaces = endpointResp.Namespaces
							serviceResp.EndpointCount += endpointResp.EndpointCount
							serviceResp.AddNamespaces(endpointResp.Namespaces)
						} else {
							result[endpointResp.ServiceName] = &response.EntryInstanceData{
								ServiceName:    endpointResp.ServiceName,
								Namespaces:     endpointResp.Namespaces,
								EndpointCount:  endpointResp.EndpointCount,
								ServiceDetails: endpointResp.ServiceDetails,
							}
						}

						for _, detail := range endpointResp.ServiceDetails {
							if detail.Latency.Ratio.DayOverDay != nil && *detail.Latency.Ratio.DayOverDay > threshold.Latency {
								resp.Status = model.STATUS_CRITICAL
							}
							if detail.Latency.Ratio.WeekOverDay != nil && *detail.Latency.Ratio.WeekOverDay > threshold.Latency {
								resp.Status = model.STATUS_CRITICAL
							}
							if detail.ErrorRate.Ratio.DayOverDay != nil && *detail.ErrorRate.Ratio.DayOverDay > threshold.ErrorRate {
								resp.Status = model.STATUS_CRITICAL
							}
							if detail.ErrorRate.Ratio.WeekOverDay != nil && *detail.ErrorRate.Ratio.WeekOverDay > threshold.ErrorRate {
								resp.Status = model.STATUS_CRITICAL
							}
						}
					}
				} else {
					break
				}
			}
		}
		if err == nil {
			// 补全日志错误数等信息
			startTime := time.Unix(req.StartTime/1000000, 0)
			endTime := time.Unix(req.EndTime/1000000, 0)
			step := time.Duration(req.Step * 1000)
			serviceNames := make([]string, 0)
			for serviceName := range result {
				serviceNames = append(serviceNames, serviceName)
			}
			alertResps, err = h.serviceoverviewService.GetServicesAlert(startTime, endTime, step, serviceNames, nil)
			if err == nil {
				for _, alertResp := range alertResps {
					if serviceResp, found := result[alertResp.ServiceName]; found {
						serviceResp.Logs = alertResp.Logs
						serviceResp.Timestamp = alertResp.Timestamp
						serviceResp.AlertStatus = alertResp.AlertStatus
						serviceResp.AlertReason = alertResp.AlertReason
					}

					if alertResp.Logs.Ratio.DayOverDay != nil && *alertResp.Logs.Ratio.DayOverDay > threshold.Log {
						resp.Status = model.STATUS_CRITICAL
					}
					if alertResp.Logs.Ratio.WeekOverDay != nil && *alertResp.Logs.Ratio.WeekOverDay > threshold.Log {
						resp.Status = model.STATUS_CRITICAL
					}
					if alertResp.AlertStatusCH.InfrastructureStatus == model.STATUS_CRITICAL ||
						alertResp.AlertStatusCH.NetStatus == model.STATUS_CRITICAL ||
						alertResp.AlertStatusCH.K8sStatus == model.STATUS_CRITICAL {
						resp.Status = model.STATUS_CRITICAL
					}
				}
			}
		}

		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetServiceEntryEndpointsError,
				code.Text(code.GetServiceEntryEndpointsError)).WithError(err),
			)
			return
		}

		for _, endpointsResp := range result {
			resp.Data = append(resp.Data, endpointsResp)
		}
		c.Payload(resp)
	}
}
