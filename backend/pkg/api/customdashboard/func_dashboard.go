// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package customdashboard

import (
	"errors"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	customdashboardservice "github.com/CloudDetail/apo/backend/pkg/services/customdashboard"
)

func (h *handler) List() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.DashboardListRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(http.StatusBadRequest, code.ParamBindError, err)
			return
		}
		resp, err := h.service.List(c, req)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, code.GetCustomDashboardError, err)
			return
		}
		c.Payload(resp)
	}
}

func (h *handler) Get() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.DashboardIDRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(http.StatusBadRequest, code.ParamBindError, err)
			return
		}
		resp, err := h.service.Get(c, req.ID)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, code.GetCustomDashboardError, err)
			return
		}
		c.Payload(resp)
	}
}

func (h *handler) Create() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.CreateDashboardRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(http.StatusBadRequest, code.ParamBindError, err)
			return
		}
		resp, err := h.service.Create(c, req)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, code.CreateCustomDashboardError, err)
			return
		}
		c.Payload(resp)
	}
}

func (h *handler) Update() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.UpdateDashboardRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(http.StatusBadRequest, code.ParamBindError, err)
			return
		}
		resp, err := h.service.Update(c, req)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, code.UpdateCustomDashboardError, err)
			return
		}
		c.Payload(resp)
	}
}

func (h *handler) Copy() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.DashboardIDRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(http.StatusBadRequest, code.ParamBindError, err)
			return
		}
		resp, err := h.service.Copy(c, req.ID)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, code.CreateCustomDashboardError, err)
			return
		}
		c.Payload(resp)
	}
}

func (h *handler) Delete() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.DashboardIDRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(http.StatusBadRequest, code.ParamBindError, err)
			return
		}
		if err := h.service.Delete(c, req.ID); err != nil {
			c.AbortWithError(http.StatusBadRequest, code.DeleteCustomDashboardError, err)
			return
		}
		c.Payload("ok")
	}
}

func (h *handler) Templates() core.HandlerFunc {
	return func(c core.Context) {
		c.Payload(h.service.Templates(c))
	}
}

func (h *handler) PresetQueries() core.HandlerFunc {
	return func(c core.Context) {
		c.Payload(h.service.PresetQueries(c))
	}
}

func (h *handler) VariableCandidates() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.DashboardVariableRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(http.StatusBadRequest, code.ParamBindError, err)
			return
		}
		resp, err := h.service.VariableCandidates(c, req)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, variableCandidatesErrorCode(err), err)
			return
		}
		c.Payload(resp)
	}
}

func (h *handler) VariableCandidatesQuery() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.DashboardVariableRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(http.StatusBadRequest, code.ParamBindError, err)
			return
		}
		resp, err := h.service.VariableCandidates(c, req)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, variableCandidatesErrorCode(err), err)
			return
		}
		c.Payload(resp)
	}
}

func variableCandidatesErrorCode(err error) string {
	if errors.Is(err, customdashboardservice.ErrNoDataGroupPermission) {
		return code.UserNoPermissionError
	}
	return code.GetCustomDashboardError
}

func (h *handler) ValidateQuery() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.DashboardValidateQueryRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(http.StatusBadRequest, code.ParamBindError, err)
			return
		}
		c.Payload(h.service.ValidateQuery(c, req.Expr))
	}
}

func (h *handler) PreviewQuery() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.DashboardQueryRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(http.StatusBadRequest, code.ParamBindError, err)
			return
		}
		resp, err := h.service.PreviewQuery(c, req)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, code.QueryCustomDashboardError, err)
			return
		}
		c.Payload(resp)
	}
}
