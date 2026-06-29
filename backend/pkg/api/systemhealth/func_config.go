// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package systemhealth

import (
	"fmt"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

func (h *handler) GetConfig() core.HandlerFunc {
	return func(c core.Context) {
		resp, err := h.systemHealthConfig(c)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, code.GetDataGroupError, err)
			return
		}
		c.Payload(resp)
	}
}

func (h *handler) UpdateConfig() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.UpdateSystemHealthConfigRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(http.StatusBadRequest, code.ParamBindError, err)
			return
		}
		if req.SevereAbnormalScoreThreshold >= req.GeneralAbnormalScoreThreshold {
			c.AbortWithError(http.StatusBadRequest, code.ParamBindError, fmt.Errorf("severeAbnormalScoreThreshold must be lower than generalAbnormalScoreThreshold"))
			return
		}
		cfg, err := h.dbRepo.UpdateSystemHealthConfig(c, req.SevereAbnormalScoreThreshold, req.GeneralAbnormalScoreThreshold)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, code.GetDataGroupError, err)
			return
		}
		c.Payload(systemHealthConfigResponse(cfg))
	}
}

func (h *handler) systemHealthConfig(ctx core.Context) (response.SystemHealthConfigResp, error) {
	cfg, err := h.dbRepo.GetOrCreateSystemHealthConfig(ctx)
	if err != nil {
		return defaultSystemHealthScoreThresholds(), err
	}
	return systemHealthConfigResponse(cfg), nil
}

func systemHealthConfigResponse(cfg database.SystemHealthConfig) response.SystemHealthConfigResp {
	return normalizeSystemHealthScoreThresholds(response.SystemHealthConfigResp{
		AbnormalScoreThreshold:        cfg.AbnormalScoreThreshold,
		SevereAbnormalScoreThreshold:  cfg.SevereAbnormalScoreThreshold,
		GeneralAbnormalScoreThreshold: cfg.GeneralAbnormalScoreThreshold,
	})
}

func defaultSystemHealthScoreThresholds() response.SystemHealthConfigResp {
	return response.SystemHealthConfigResp{
		AbnormalScoreThreshold:        database.DefaultSystemHealthGeneralAbnormalScoreThreshold,
		SevereAbnormalScoreThreshold:  database.DefaultSystemHealthSevereAbnormalScoreThreshold,
		GeneralAbnormalScoreThreshold: database.DefaultSystemHealthGeneralAbnormalScoreThreshold,
	}
}

func normalizeSystemHealthScoreThresholds(config response.SystemHealthConfigResp) response.SystemHealthConfigResp {
	generalThreshold := config.GeneralAbnormalScoreThreshold
	if generalThreshold < 1 || generalThreshold > 100 {
		generalThreshold = config.AbnormalScoreThreshold
	}
	if generalThreshold < 1 || generalThreshold > 100 {
		generalThreshold = database.DefaultSystemHealthGeneralAbnormalScoreThreshold
	}

	severeThreshold := config.SevereAbnormalScoreThreshold
	if severeThreshold < 1 || severeThreshold >= generalThreshold {
		severeThreshold = database.DefaultSystemHealthSevereAbnormalScoreThreshold
	}
	if severeThreshold < 1 || severeThreshold >= generalThreshold {
		return defaultSystemHealthScoreThresholds()
	}

	return response.SystemHealthConfigResp{
		AbnormalScoreThreshold:        generalThreshold,
		SevereAbnormalScoreThreshold:  severeThreshold,
		GeneralAbnormalScoreThreshold: generalThreshold,
	}
}
