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
		req.FillerValue()
		resp, err := h.logService.CreateLogTable(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.CreateLogTableError,
				code.Text(code.CreateLogTableError)).WithError(err),
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
				code.DropLogTableError,
				code.Text(code.DropLogTableError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}

// UpdateLogTable 更新日志表
// @Summary 更新日志表
// @Description 更新日志表
// @Tags API.log
// @Accept json
// @Produce json
// @Param Request body request.LogTableRequest true "请求信息"
// @Success 200 {object} response.LogTableResponse
// @Failure 400 {object} code.Failure
// @Router /api/log/update [post]
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
				code.UpateLogTableError,
				code.Text(code.UpateLogTableError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
