package alertanalyze

import (
	"errors"
	"net/http"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/serviceoverview"
)

// GetAlertImpact 获取告警数据的影响面
// @Summary 获取告警数据的影响面
// @Description 获取告警数据的影响面
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param eventId query string false "查询告警事件ID"
// @Param startTime query uint64 true "查询开始时间"
// @Param endTime query uint64 true "查询结束时间"
// @Param step query int64 true "查询步长(us)"
// @Success 200 {object} []response.GetServiceEntryEndpointsResponse
// @Failure 400 {object} code.Failure
// @Router /api/alerts/event/impact [get]
func (h *handler) GetAlertImpact() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.AlertImpactRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		entryRelations, events, relatedEventCounts, pagation, err := h.alertanalyzeService.AlertImpact(req)
		if err != nil {
			var vErr model.ErrAlertImpactMissingTag
			var vErr2 model.ErrAlertImpactNoMatchedService
			if errors.As(err, &vErr) {
				// 告警所需的关联信息不足
				// 提示告警中需要包含下列任意label组合
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					code.AlertEventImpactMissingTag,
					code.Text(code.AlertEventImpactMissingTag)+vErr.CheckedTagGroups()).WithError(err),
				)
				return
			} else if errors.As(err, &vErr2) {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					code.AlertEventImpactNoMatchedService,
					code.Text(code.AlertEventImpactNoMatchedService)+vErr2.CheckedTagGroup()).WithError(err),
				)
				return
			} else {
				// 查询失败
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					code.AlertEventImpactError,
					code.Text(code.AlertEventImpactError)).WithError(err),
				)
				return
			}
		}

		// 填充入口节点的RED/告警等信息
		resp, err := fillEntryNodeDetail(h.serviceoverviewService, req, entryRelations)
		if err != nil {
			// 查询失败
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.AlertEventImpactError,
				code.Text(code.AlertEventImpactError)).WithError(err),
			)
			return
		}

		// 填充入口节点的相关告警相关率
		fillRelatedAlertRate(resp, relatedEventCounts, pagation)

		resp.AlertEvents = events
		resp.Pagination = pagation
		c.Payload(resp)
	}
}

func fillRelatedAlertRate(resp *response.GetAlertImpactResponse, relatedEventCounts map[model.EndpointKey]int, pagation *model.Pagination) {
	for _, entryEndpoint := range resp.Data {
		key := model.EndpointKey{
			ServiceName: entryEndpoint.ServiceName,
			Endpoint:    entryEndpoint.Endpoint,
		}
		relatedAlertCount, find := relatedEventCounts[key]
		if find {
			entryEndpoint.RelatedAlertRate = float64(relatedAlertCount) * 100 / float64(pagation.Total)
		} else {
			entryEndpoint.RelatedAlertRate = 0
		}
	}
}

// fillEntryNodeDetail 填充EntryEndpoint信息
// 复制于 backend/pkg/api/service/func_getserviceentryendpoints.go
func fillEntryNodeDetail(s serviceoverview.Service, req *request.AlertImpactRequest, entryRelationship []clickhouse.EntryRelationship) (*response.GetAlertImpactResponse, error) {
	resp := response.GetAlertImpactResponse{
		Data: make([]*response.EndpointData, 0),
	}

	startTime := time.UnixMicro(req.StartTime)
	endTime := time.UnixMicro(req.EndTime)
	sortRule := serviceoverview.DODThreshold
	step := time.Duration(req.Step * 1000)

	endpointsSet := map[prometheus.EndpointKey]struct{}{}
	entryEndpoints := make([]prometheus.EndpointKey, 0)
	for _, relation := range entryRelationship {
		entryEndpointKey := prometheus.EndpointKey{
			SvcName:    relation.GetEntryService(),
			ContentKey: relation.GetEntryEndpoint(),
		}
		if _, find := endpointsSet[entryEndpointKey]; find {
			continue
		}

		endpointsSet[entryEndpointKey] = struct{}{}
		entryEndpoints = append(entryEndpoints, entryEndpointKey)
	}

	endpointResps, err := s.GetServicesEndpointDataByEndpoints(startTime, endTime, step, entryEndpoints, sortRule)
	if err != nil {
		return nil, err
	}

	for _, endpointResp := range endpointResps {
		if len(endpointResp.ServiceDetails) == 0 {
			continue
		}
		resp.Data = append(resp.Data, &response.EndpointData{
			ServiceName:   endpointResp.ServiceName,
			Namespaces:    []string{},
			ServiceDetail: endpointResp.ServiceDetails[0],
		})
	}

	return &resp, nil
}
