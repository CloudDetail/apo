package agentconfig

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	ac "github.com/CloudDetail/apo/backend/pkg/model/agentconfig"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	svc "github.com/CloudDetail/apo/backend/pkg/services/agentconfig"
	"github.com/google/uuid"
)

type Handler struct{ s *svc.Service }

func New(repo database.Repo) *Handler   { return &Handler{s: svc.New(repo)} }
func fail(c core.Context, e error)      { c.AbortWithError(http.StatusBadRequest, code.ParamBindError, e) }
func agentNodeID(c core.Context) string { v, _ := c.Get("agentNodeID"); s, _ := v.(string); return s }
func (h *Handler) Inspect() core.HandlerFunc {
	return func(c core.Context) {
		var r struct {
			Content string `json:"content"`
		}
		if e := c.ShouldBindJSON(&r); e != nil {
			fail(c, e)
			return
		}
		c.Payload(map[string]any{"variables": h.s.Inspect(r.Content)})
	}
}
func (h *Handler) ListTemplates() core.HandlerFunc {
	return func(c core.Context) {
		v, e := h.s.ListTemplates(c, c.Request().URL.Query().Get("probeType"))
		if e != nil {
			fail(c, e)
			return
		}
		c.Payload(map[string]any{"templates": v})
	}
}
func (h *Handler) GetTemplate() core.HandlerFunc {
	return func(c core.Context) {
		id, e := uuid.Parse(c.Param("id"))
		if e != nil {
			fail(c, e)
			return
		}
		v, e := h.s.GetTemplate(c, id)
		if e != nil {
			fail(c, e)
			return
		}
		c.Payload(v)
	}
}
func (h *Handler) CreateTemplate() core.HandlerFunc {
	return func(c core.Context) {
		var r svc.TemplateInput
		if e := c.ShouldBindJSON(&r); e != nil {
			fail(c, e)
			return
		}
		v, e := h.s.CreateTemplate(c, r, c.UserID())
		if e != nil {
			fail(c, e)
			return
		}
		c.Payload(v)
	}
}
func (h *Handler) UpdateTemplate() core.HandlerFunc {
	return func(c core.Context) {
		id, e := uuid.Parse(c.Param("id"))
		if e != nil {
			fail(c, e)
			return
		}
		var r svc.TemplateInput
		if e = c.ShouldBindJSON(&r); e != nil {
			fail(c, e)
			return
		}
		v, e := h.s.UpdateTemplate(c, id, r, c.UserID())
		if e != nil {
			fail(c, e)
			return
		}
		c.Payload(v)
	}
}
func (h *Handler) DeleteTemplate() core.HandlerFunc {
	return func(c core.Context) {
		id, e := uuid.Parse(c.Param("id"))
		if e == nil {
			e = h.s.DeleteTemplate(c, id)
		}
		if e != nil {
			fail(c, e)
			return
		}
		c.Payload("ok")
	}
}
func (h *Handler) ListNodes() core.HandlerFunc {
	return func(c core.Context) {
		v, e := h.s.ListNodes(c)
		if e != nil {
			fail(c, e)
			return
		}
		c.Payload(map[string]any{"nodes": v})
	}
}
func (h *Handler) ListStates() core.HandlerFunc {
	return func(c core.Context) {
		v, e := h.s.ListStates(c)
		if e != nil {
			fail(c, e)
			return
		}
		c.Payload(map[string]interface{}{"states": v})
	}
}
func (h *Handler) GetGlobalCircuit() core.HandlerFunc {
	return func(c core.Context) {
		settings, err := h.s.GetGlobalCircuit(c)
		if err != nil {
			fail(c, err)
			return
		}
		c.Payload(settings)
	}
}
func (h *Handler) SaveGlobalCircuit() core.HandlerFunc {
	return func(c core.Context) {
		var settings ac.CircuitBreakerSettings
		if err := c.ShouldBindJSON(&settings); err != nil {
			fail(c, err)
			return
		}
		saved, err := h.s.SaveGlobalCircuit(c, settings, c.UserID())
		if err != nil {
			fail(c, err)
			return
		}
		c.Payload(saved)
	}
}
func (h *Handler) GetNodeCircuit() core.HandlerFunc {
	return func(c core.Context) {
		settings, source, err := h.s.GetNodeCircuit(c, c.Param("nodeId"))
		if err != nil {
			fail(c, err)
			return
		}
		c.Payload(map[string]interface{}{"circuitBreaker": settings, "source": source})
	}
}
func (h *Handler) SaveNodeCircuit() core.HandlerFunc {
	return func(c core.Context) {
		var settings ac.CircuitBreakerSettings
		if err := c.ShouldBindJSON(&settings); err != nil {
			fail(c, err)
			return
		}
		saved, err := h.s.SaveNodeCircuit(c, c.Param("nodeId"), settings, c.UserID())
		if err != nil {
			fail(c, err)
			return
		}
		c.Payload(saved)
	}
}
func (h *Handler) BatchSaveCircuit() core.HandlerFunc {
	return func(c core.Context) {
		var input svc.CircuitBatchInput
		if err := c.ShouldBindJSON(&input); err != nil {
			fail(c, err)
			return
		}
		result, err := h.s.BatchSaveCircuit(c, input, c.UserID())
		if err != nil {
			fail(c, err)
			return
		}
		c.Payload(result)
	}
}
func (h *Handler) GetCurrentNodeConfigs() core.HandlerFunc {
	return func(c core.Context) {
		config, err := h.s.CurrentNodeConfigs(c, c.Param("nodeId"))
		if err != nil {
			fail(c, err)
			return
		}
		c.Payload(config)
	}
}
func (h *Handler) UpdateNodeConfigs() core.HandlerFunc {
	return func(c core.Context) {
		var input svc.NodeConfigsInput
		if err := c.ShouldBindJSON(&input); err != nil {
			fail(c, err)
			return
		}
		configs, err := h.s.UpdateNodeConfigs(c, c.Param("nodeId"), input)
		if err != nil {
			fail(c, err)
			return
		}
		c.Payload(configs)
	}
}
func (h *Handler) ListNodeConfigHistory() core.HandlerFunc {
	return func(c core.Context) {
		limit, _ := strconv.Atoi(c.Request().URL.Query().Get("limit"))
		configs, err := h.s.ListNodeConfigHistory(c, c.Param("nodeId"), c.Param("probeType"), limit)
		if err != nil {
			fail(c, err)
			return
		}
		c.Payload(map[string]any{"configs": configs})
	}
}
func (h *Handler) BatchAssign() core.HandlerFunc {
	return func(c core.Context) {
		var input svc.BatchAssignmentInput
		if err := c.ShouldBindJSON(&input); err != nil {
			fail(c, err)
			return
		}
		result, err := h.s.BatchAssign(c, input, c.UserID())
		if err != nil {
			fail(c, err)
			return
		}
		c.Payload(result)
	}
}
func (h *Handler) CreateCommand() core.HandlerFunc {
	return func(c core.Context) {
		var r svc.CommandInput
		if e := c.ShouldBindJSON(&r); e != nil {
			fail(c, e)
			return
		}
		v, e := h.s.CreateCommand(c, c.Param("nodeId"), c.Param("probeType"), r, c.UserID())
		if e != nil {
			fail(c, e)
			return
		}
		c.Payload(v)
	}
}
func (h *Handler) ListCommands() core.HandlerFunc {
	return func(c core.Context) {
		limit, _ := strconv.Atoi(c.Request().URL.Query().Get("limit"))
		v, e := h.s.ListCommands(c, c.Param("nodeId"), c.Param("probeType"), limit)
		if e != nil {
			fail(c, e)
			return
		}
		c.Payload(map[string]interface{}{"commands": v})
	}
}
func (h *Handler) Register() core.HandlerFunc {
	return func(c core.Context) {
		var n ac.Node
		if e := c.ShouldBindJSON(&n); e != nil {
			fail(c, e)
			return
		}
		if n.ID != "" && n.ID != agentNodeID(c) {
			fail(c, errors.New("node identity mismatch"))
			return
		}
		n.ID = agentNodeID(c)
		if e := h.s.Register(c, &n); e != nil {
			fail(c, e)
			return
		}
		c.Payload("ok")
	}
}
func (h *Handler) SaveInitialConfig() core.HandlerFunc {
	return func(c core.Context) {
		var input svc.InitialConfigInput
		if err := c.ShouldBindJSON(&input); err != nil {
			fail(c, err)
			return
		}
		baseline, saved, err := h.s.SaveInitialConfig(c, agentNodeID(c), input)
		if err != nil {
			fail(c, err)
			return
		}
		c.Payload(map[string]any{"saved": saved, "baseline": baseline})
	}
}
func (h *Handler) GetInitialConfig() core.HandlerFunc {
	return func(c core.Context) {
		baseline, err := h.s.GetInitialConfig(c, c.Param("nodeId"), c.Param("probeType"))
		if err != nil {
			fail(c, err)
			return
		}
		c.Payload(baseline)
	}
}
func (h *Handler) GetConfig() core.HandlerFunc {
	return func(c core.Context) {
		known, _ := strconv.ParseInt(c.Request().URL.Query().Get("knownGeneration"), 10, 64)
		v, e := h.s.GetConfig(c, agentNodeID(c), c.Request().URL.Query().Get("probeType"), known)
		if e != nil {
			fail(c, e)
			return
		}
		if v == nil {
			c.SetHeader("X-Agent-No-Change", "true")
			c.Payload(nil)
			return
		}
		c.Payload(map[string]any{"nodeId": v.NodeID, "probeType": v.ProbeType, "generation": v.Generation, "contentId": v.ContentSHA256, "contentSha256": v.ContentSHA256, "content": v.RenderedContent, "renderedContent": v.RenderedContent, "files": v.Files, "logSources": v.LogSources})
	}
}
func (h *Handler) Settings() core.HandlerFunc {
	return func(c core.Context) {
		settings, err := h.s.SettingsForProbe(c, agentNodeID(c), c.Request().URL.Query().Get("probeType"))
		if err != nil {
			fail(c, err)
			return
		}
		c.Payload(settings)
	}
}
func (h *Handler) Report() core.HandlerFunc {
	return func(c core.Context) {
		var v ac.ApplyState
		if e := c.ShouldBindJSON(&v); e != nil {
			fail(c, e)
			return
		}
		if v.NodeID != "" && v.NodeID != agentNodeID(c) {
			fail(c, errors.New("node identity mismatch"))
			return
		}
		v.NodeID = agentNodeID(c)
		if e := h.s.Report(c, &v); e != nil {
			fail(c, e)
			return
		}
		c.Payload("ok")
	}
}
func (h *Handler) Commands() core.HandlerFunc {
	return func(c core.Context) {
		v, e := h.s.Commands(c, agentNodeID(c), c.Request().URL.Query().Get("probeType"), c.GetHeader("X-Agent-Instance-ID"))
		if e != nil {
			fail(c, e)
			return
		}
		c.Payload(map[string]any{"commands": v})
	}
}
func (h *Handler) FinishCommand() core.HandlerFunc {
	return func(c core.Context) {
		id, e := uuid.Parse(c.Param("id"))
		var r struct {
			Status     string `json:"status" binding:"required"`
			LeaseToken string `json:"leaseToken" binding:"required"`
			Result     string `json:"result"`
		}
		if e == nil {
			e = c.ShouldBindJSON(&r)
		}
		if e == nil {
			_, e = h.s.TransitionCommand(c, id, agentNodeID(c), c.GetHeader("X-Agent-Instance-ID"), r.LeaseToken, r.Status, r.Result)
		}
		if e != nil {
			fail(c, e)
			return
		}
		c.Payload("ok")
	}
}
