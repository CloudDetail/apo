// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"net/http"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/middleware"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/services/serviceoverview"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

// GetServiceEntryEndpoints get the list of service portal Endpoint
// @Summary get the service entry Endpoint list
// @Description get the service entry Endpoint list
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query uint64 true "query start time"
// @Param endTime query uint64 true "query end time"
// @Param service query string true "Query service name"
// @Param endpoint query string true "Query Endpoint"
// @Param step query int64 true "query step (us)"
// @Param showMissTop query bool false "Show missing entry"
// @Param Authorization header string false "Bearer accessToken"
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
				c.ErrMessage(code.ParamBindError)).WithError(err),
			)
			return
		}
		var (
			err           error
			threshold     response.GetThresholdResponse
			endpointResps []response.ServiceEndPointsRes
			alertResps    []response.ServiceAlertRes
		)

		userID := middleware.GetContextUserID(c)
		err = h.dataService.CheckDatasourcePermission(userID, 0, nil, &req.Service, model.DATASOURCE_CATEGORY_APM)
		if err != nil {
			c.HandleError(err, code.AuthError, &response.GetServiceEntryEndpointsResponse{
				Status: model.STATUS_NORMAL,
				Data:   []*response.EntryInstanceData{},
			})
			return
		}

		result := make(map[string]*response.EntryInstanceData, 0)
		resp := response.GetServiceEntryEndpointsResponse{
			Status: model.STATUS_NORMAL,
			Data:   make([]*response.EntryInstanceData, 0),
		}
		entryNodes, err := h.serviceInfoService.GetServiceEntryEndpoints(req)
		if err == nil {
			// TODO defaults to global Threshold first, and then adjusts to the Threshold of specific services.
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
			// Complete information such as the number of log errors
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
				c.ErrMessage(code.GetServiceEntryEndpointsError)).WithError(err),
			)
			return
		}

		for _, endpointsResp := range result {
			resp.Data = append(resp.Data, endpointsResp)
		}
		c.Payload(resp)
	}
}
