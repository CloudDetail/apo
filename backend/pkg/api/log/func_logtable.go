package log

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (h *handler) CreateLogTable() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.LogTableRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		resp, err := h.logService.CreateLogTable(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetFaultLogContentError,
				code.Text(code.GetFaultLogContentError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}

func (h *handler) DropLogTable() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.LogTableRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		resp, err := h.logService.DropLogTable(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetFaultLogContentError,
				code.Text(code.GetFaultLogContentError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}

func (h *handler) UpdateLogTable() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.LogTableRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		resp, err := h.logService.UpdateLogTable(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetFaultLogContentError,
				code.Text(code.GetFaultLogContentError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}

func (h *handler) GetLogTableInfo() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.LogTableRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		resp, err := h.logService.GetLogTableInfo(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetFaultLogContentError,
				code.Text(code.GetFaultLogContentError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
