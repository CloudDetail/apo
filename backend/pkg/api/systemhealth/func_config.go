// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package systemhealth

import (
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
		cfg, err := h.dbRepo.UpdateSystemHealthConfig(c, req.AbnormalScoreThreshold)
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
		return response.SystemHealthConfigResp{
			AbnormalScoreThreshold: database.DefaultSystemHealthAbnormalScoreThreshold,
		}, err
	}
	return systemHealthConfigResponse(cfg), nil
}

func systemHealthConfigResponse(cfg database.SystemHealthConfig) response.SystemHealthConfigResp {
	threshold := normalizeSystemHealthAbnormalScoreThreshold(cfg.AbnormalScoreThreshold)
	return response.SystemHealthConfigResp{AbnormalScoreThreshold: threshold}
}

func normalizeSystemHealthAbnormalScoreThreshold(threshold int) int {
	if threshold < 1 || threshold > 100 {
		return database.DefaultSystemHealthAbnormalScoreThreshold
	}
	return threshold
}
