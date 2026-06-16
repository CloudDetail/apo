// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package systemhealth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
	alertmodel "github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/common"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prommodel "github.com/prometheus/common/model"
	corev1 "k8s.io/api/core/v1"
)

var expectedSources = []string{"alerts", "impactEntries", "relations", "redMetrics", "descendantContribution"}

var sourceLabels = map[string]string{
	"alerts":                 "真实告警事件",
	"impactEntries":          "影响入口",
	"relations":              "服务调用关系",
	"redMetrics":             "APM/RED指标",
	"descendantContribution": "下游贡献度",
	"descendantRelevance":    "下游相关性",
	"trace":                  "Trace链路",
	"logs":                   "日志语义",
	"ebpf":                   "eBPF明细",
	"infra":                  "基础设施指标",
	"alertReport":            "原告警分析报告",
}

const (
	issueKindActiveAlert              = "activeAlert"
	issueKindInspection               = "inspection"
	issueKindRecoveredAlert           = "recoveredAlert"
	issueKindInvalidAlert             = "invalidAlert"
	maxEvidenceSamples                = 3
	maxAlertReportLookups             = 6
	maxMetricEvidenceSeries           = 3
	maxMetricEvidencePoints           = 240
	enableSystemHealthReportCache     = false
	systemHealthReportCacheTTL        = 45 * time.Second
	maxSystemHealthReportCacheEntries = 64
)

type systemHealthReportCache struct {
	mutex   sync.RWMutex
	entries map[string]systemHealthReportCacheEntry
}

type systemHealthReportCacheEntry struct {
	report    response.SystemHealthReportResp
	expiresAt time.Time
}

type systemHealthAISummaryRequest struct {
	Window string                     `json:"window"`
	Report systemHealthAIReportInput  `json:"report"`
	Issues []systemHealthAIIssueInput `json:"issues"`
}

type systemHealthAIReportInput struct {
	Status           string   `json:"status"`
	Score            int      `json:"score"`
	ServiceRiskCount int      `json:"serviceRiskCount"`
	ActiveAlertCount int      `json:"activeAlertCount"`
	AvailableSources []string `json:"availableSources"`
	Summary          string   `json:"summary"`
}

type systemHealthAIIssueInput struct {
	Key                   string                             `json:"key"`
	Priority              string                             `json:"priority"`
	Title                 string                             `json:"title"`
	Status                string                             `json:"status"`
	Scope                 string                             `json:"scope"`
	Impact                string                             `json:"impact"`
	ActiveAlerts          int                                `json:"activeAlerts"`
	StartedAt             string                             `json:"startedAt"`
	DataSources           []string                           `json:"dataSources"`
	Evidence              []systemHealthAIEvidenceInput      `json:"evidence"`
	ExistingAlertAnalyses []systemHealthAIAlertAnalysisInput `json:"existingAlertAnalyses"`
	ExistingReportActions []systemHealthAIReportActionInput  `json:"existingReportActions"`
}

type systemHealthAIEvidenceInput struct {
	Type     string `json:"type"`
	Source   string `json:"source,omitempty"`
	Severity string `json:"severity,omitempty"`
	Result   string `json:"result"`
}

type systemHealthAIAlertAnalysisInput struct {
	EventID        string `json:"eventId,omitempty"`
	AlertID        string `json:"alertId,omitempty"`
	AlertName      string `json:"alertName,omitempty"`
	AlertDirection string `json:"alertDirection,omitempty"`
	AnalyzeRunID   string `json:"analyzeRunId,omitempty"`
	AnalyzeError   string `json:"analyzeError,omitempty"`
	CheckOutput    string `json:"checkOutput,omitempty"`
	Validity       string `json:"validity,omitempty"`
}

type systemHealthAIReportActionInput struct {
	AlertName string `json:"alertName,omitempty"`
	Level     string `json:"level,omitempty"`
	Action    string `json:"action"`
}

type systemHealthAISummaryResponse struct {
	Summary       string                      `json:"summary"`
	WorkflowRunID string                      `json:"workflowRunId"`
	Issues        []systemHealthAIIssueOutput `json:"issues"`
}

type systemHealthAIIssueOutput struct {
	Key        string                                `json:"key"`
	Analysis   string                                `json:"analysis"`
	RootCause  string                                `json:"rootCause"`
	Actions    []string                              `json:"actions"`
	ActionPlan []response.SystemHealthActionPlanStep `json:"actionPlan"`
}

type openAIResponsesRequest struct {
	Model        string                    `json:"model"`
	Instructions string                    `json:"instructions,omitempty"`
	Input        string                    `json:"input"`
	Text         openAIResponsesTextConfig `json:"text,omitempty"`
}

type openAIResponsesTextConfig struct {
	Format openAIResponsesTextFormat `json:"format"`
}

type openAIResponsesTextFormat struct {
	Type   string         `json:"type"`
	Name   string         `json:"name"`
	Strict bool           `json:"strict"`
	Schema map[string]any `json:"schema"`
}

type openAIResponsesResponse struct {
	ID         string                   `json:"id"`
	OutputText string                   `json:"output_text"`
	Output     []openAIResponsesOutput  `json:"output"`
	Error      *openAIResponsesAPIError `json:"error,omitempty"`
}

type openAIResponsesOutput struct {
	Type    string                         `json:"type"`
	Content []openAIResponsesOutputContent `json:"content"`
}

type openAIResponsesOutputContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type openAIResponsesAPIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

func newSystemHealthReportCache() *systemHealthReportCache {
	return &systemHealthReportCache{
		entries: make(map[string]systemHealthReportCacheEntry),
	}
}

func (c *systemHealthReportCache) get(key string, now time.Time) (response.SystemHealthReportResp, bool) {
	if c == nil || key == "" {
		return response.SystemHealthReportResp{}, false
	}
	c.mutex.RLock()
	entry, ok := c.entries[key]
	c.mutex.RUnlock()
	if !ok || now.After(entry.expiresAt) {
		return response.SystemHealthReportResp{}, false
	}
	return entry.report, true
}

func (c *systemHealthReportCache) set(key string, report response.SystemHealthReportResp, now time.Time) {
	if c == nil || key == "" {
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if len(c.entries) >= maxSystemHealthReportCacheEntries {
		for existingKey, entry := range c.entries {
			if now.After(entry.expiresAt) {
				delete(c.entries, existingKey)
			}
		}
	}
	if len(c.entries) >= maxSystemHealthReportCacheEntries {
		var oldestKey string
		var oldestExpiry time.Time
		for existingKey, entry := range c.entries {
			if oldestKey == "" || entry.expiresAt.Before(oldestExpiry) {
				oldestKey = existingKey
				oldestExpiry = entry.expiresAt
			}
		}
		delete(c.entries, oldestKey)
	}
	c.entries[key] = systemHealthReportCacheEntry{
		report:    report,
		expiresAt: now.Add(systemHealthReportCacheTTL),
	}
}

type reportInput struct {
	key                   string
	aggregationKey        string
	kind                  string
	title                 string
	scope                 string
	scopeKind             string
	startedAt             string
	updatedAt             string
	activeAlerts          int
	deduplicatedAlerts    int
	rawEventCount         int
	repeatCount           int
	durationSeconds       float64
	objectCriticality     float64
	alertSeverities       []string
	alertNames            []string
	events                []alertmodel.AEventWithWRecord
	normalizedEvents      []normalizedAlertEvent
	deduplicatedInstances []deduplicatedAlertInstance
	availableCalcSources  []string
	displaySources        []string
	affectedObjects       []response.SystemHealthAffectedObject
	mergeReasons          []string
	alertReportActions    []systemHealthAlertReportAction
	evidence              []response.SystemHealthEvidence
	impact                string
}

type systemHealthAlertReportAction struct {
	AlertName string
	Level     string
	Action    string
}

type scoredIssue struct {
	input                   reportInput
	entryImpactScore        float64
	alertSeverityScore      float64
	redDegradationScore     float64
	propagationScore        float64
	dataCompletenessPenalty float64
	pathRisk                float64
	dataQualityScore        int
	scoreImpact             int
	priority                string
	status                  string
}

type alertMetricSeries struct {
	Labels map[string]string
	Values []prometheus.Points
}

type alertMetricTrendStats struct {
	SeriesCount   int
	PointCount    int
	FirstValue    float64
	LastValue     float64
	MinValue      float64
	MaxValue      float64
	Delta         float64
	ChangePercent *float64
}

type alertMetricTrendQueryResult struct {
	Rule           *request.AlertRule
	Expression     string
	Step           time.Duration
	SeriesReturned int
	MatchedSeries  []alertMetricSeries
	Stats          alertMetricTrendStats
	MatchLabels    map[string]string
	EventScoped    bool
	Message        string
}

type relatedObjectStatusCandidate struct {
	Namespace string
	Pod       string
	Container string
	Node      string
}

type alertManagerAPIAlert struct {
	Status struct {
		State string `json:"state"`
	} `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     string            `json:"startsAt"`
	EndsAt       string            `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}

type vmAlertAPIAlertsResp struct {
	Status string `json:"status"`
	Data   struct {
		Alerts []vmAlertAPIAlert `json:"alerts"`
	} `json:"data"`
}

type vmAlertAPIAlert struct {
	State      string            `json:"state"`
	Name       string            `json:"name"`
	Expression string            `json:"expression"`
	Query      string            `json:"query"`
	ID         string            `json:"id"`
	GroupID    string            `json:"group_id"`
	Labels     map[string]string `json:"labels"`
}

type vmAlertAPIRulesResp struct {
	Status string `json:"status"`
	Data   struct {
		Groups []vmAlertAPIGroup `json:"groups"`
	} `json:"data"`
}

type vmAlertAPIGroup struct {
	Name  string           `json:"name"`
	ID    string           `json:"id"`
	File  string           `json:"file"`
	Rules []vmAlertAPIRule `json:"rules"`
}

type vmAlertAPIRule struct {
	State     string            `json:"state"`
	Name      string            `json:"name"`
	Query     string            `json:"query"`
	Expr      string            `json:"expr"`
	ID        string            `json:"id"`
	GroupID   string            `json:"group_id"`
	GroupName string            `json:"group_name"`
	Labels    map[string]string `json:"labels"`
	Alerts    []vmAlertAPIAlert `json:"alerts"`
}

type telemetryCapabilityStatus struct {
	Enabled  []string
	Disabled []string
	Sources  []string
	Err      string
}

type normalizedAlertEvent struct {
	Source         string
	EventID        string
	AlertName      string
	Status         string
	Severity       string
	ServiceName    string
	Endpoint       string
	ObjectType     string
	ObjectName     string
	Group          string
	StartedAt      time.Time
	UpdatedAt      time.Time
	RecoveredAt    time.Time
	DedupKey       string
	AggregationKey string
}

type deduplicatedAlertInstance struct {
	Key             string
	AlertName       string
	ObjectType      string
	ObjectName      string
	Status          string
	MaxSeverity     string
	FirstSeenAt     time.Time
	LastSeenAt      time.Time
	RepeatCount     int
	DurationSeconds float64
}

// GetReport returns the read-only system health report used by the AI health center.
// @Summary get AI system health report
// @Description get AI system health report
// @Tags API.systemHealth
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param groupId query int false "data group id"
// @Param relativeTime query string false "relative time window"
// @Param startTime query int false "query start time"
// @Param endTime query int false "query end time"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.SystemHealthReportResp
// @Failure 400 {object} code.Failure
// @Router /api/system-health/report [get]
func (h *handler) GetReport() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetSystemHealthReportRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(http.StatusBadRequest, code.ParamBindError, err)
			return
		}

		resp, err := h.buildReport(c, req)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, code.ServerError, err)
			return
		}
		c.Payload(resp)
	}
}

func (h *handler) buildReport(ctx core.Context, req *request.GetSystemHealthReportRequest) (response.SystemHealthReportResp, error) {
	startTime, endTime, window := resolveTimeRange(req)
	now := time.Now()
	if enableSystemHealthReportCache {
		cacheKey := systemHealthReportCacheKey(req.GroupID, startTime, endTime)
		if cached, ok := h.reportCache.get(cacheKey, now); ok {
			return cached, nil
		}
	}
	events, err := h.queryRealAlertEvents(ctx, req.GroupID, startTime, endTime)
	if err != nil {
		return response.SystemHealthReportResp{}, err
	}
	if alertManagerEvents, err := h.queryAlertManagerActiveEvents(ctx, req.GroupID); err == nil {
		events = mergeAlertEvents(events, alertManagerEvents)
	}

	capabilities := h.queryTelemetryCapabilities(ctx)
	inputs := h.buildInputsFromRealEvents(ctx, req.GroupID, events, startTime, endTime, capabilities)
	scored := scoreIssues(inputs)
	healthScore := calculateHealthScore(scored)
	status := healthStatus(healthScore)
	scoreFactors := buildScoreFactors(scored)
	issues := buildIssues(scored)

	resp := response.SystemHealthReportResp{
		Report: response.SystemHealthReport{
			Status:           status,
			Score:            healthScore,
			ResourceUsage:    0,
			InfraAlerts:      countInfraAlerts(scored),
			ServiceRiskCount: len(issues),
			ActiveAlertCount: uniqueActiveAlertCount(scored),
			LogRiskCount:     0,
			Summary:          buildSummary(scored, healthScore),
			Window:           window,
			GeneratedAt:      time.Now().Format("2006-01-02 15:04:05"),
			AvailableSources: reportAvailableSources(scored),
			AIAnalysis: response.SystemHealthAIAnalysis{
				Status: "disabled",
				Source: "structured-analysis",
			},
		},
		ScoreFactors: scoreFactors,
		Issues:       issues,
	}
	h.enrichReportWithAI(ctx, &resp, scored)
	if enableSystemHealthReportCache {
		h.reportCache.set(systemHealthReportCacheKey(req.GroupID, startTime, endTime), resp, now)
	}
	return resp, nil
}

func systemHealthReportCacheKey(groupID int64, startTime time.Time, endTime time.Time) string {
	return fmt.Sprintf(
		"group=%d:start=%d:end=%d",
		groupID,
		startTime.UnixMicro(),
		endTime.UnixMicro(),
	)
}

func (h *handler) enrichReportWithAI(ctx core.Context, resp *response.SystemHealthReportResp, scored []scoredIssue) {
	if resp == nil {
		return
	}
	aiConfig := config.Get().AI
	endpoint := openAIResponsesEndpoint(aiConfig.BaseURL)
	resp.Report.AIAnalysis = response.SystemHealthAIAnalysis{
		Status:   "not_configured",
		Source:   "structured-analysis",
		Endpoint: endpoint,
	}
	if !aiConfig.Enable || strings.TrimSpace(aiConfig.BaseURL) == "" || strings.TrimSpace(aiConfig.Model) == "" {
		return
	}
	if strings.TrimSpace(aiConfig.WireAPI) != "" && strings.TrimSpace(aiConfig.WireAPI) != "responses" {
		resp.Report.AIAnalysis.Status = "failed"
		resp.Report.AIAnalysis.Error = "unsupported AI wire API: " + aiConfig.WireAPI
		return
	}
	if len(resp.Issues) == 0 {
		resp.Report.AIAnalysis.Status = "skipped"
		return
	}

	aiResp, err := callSystemHealthAI(ctx, aiConfig.BaseURL, resolveSystemHealthAIAPIKey(aiConfig.APIKey, aiConfig.EnvKey), aiConfig.Model, aiConfig.TimeoutSecond, buildSystemHealthAIRequest(resp, scored))
	if err != nil {
		resp.Report.AIAnalysis.Status = "failed"
		resp.Report.AIAnalysis.Error = truncateText(err.Error(), 220)
		return
	}
	resp.Report.AIAnalysis.Status = "succeeded"
	resp.Report.AIAnalysis.Source = "ai-api"
	resp.Report.AIAnalysis.WorkflowRunID = strings.TrimSpace(aiResp.WorkflowRunID)
	if strings.TrimSpace(aiResp.Summary) != "" {
		resp.Report.Summary = strings.TrimSpace(aiResp.Summary)
	}

	issueOutputByKey := make(map[string]systemHealthAIIssueOutput, len(aiResp.Issues))
	for _, issue := range aiResp.Issues {
		if strings.TrimSpace(issue.Key) == "" {
			continue
		}
		issueOutputByKey[issue.Key] = issue
	}
	for index := range resp.Issues {
		issueOutput, ok := issueOutputByKey[resp.Issues[index].Key]
		if !ok {
			continue
		}
		var reportActionPlan []response.SystemHealthActionPlanStep
		if index < len(scored) {
			reportActionPlan = alertReportActionPlanSteps(scored[index].input.alertReportActions, resp.Issues[index].Scope)
		}
		if strings.TrimSpace(issueOutput.Analysis) != "" {
			resp.Issues[index].AIConclusion = strings.TrimSpace(issueOutput.Analysis)
		}
		if strings.TrimSpace(issueOutput.RootCause) != "" {
			resp.Issues[index].RootCause = strings.TrimSpace(issueOutput.RootCause)
		}
		if len(issueOutput.ActionPlan) > 0 {
			resp.Issues[index].ActionPlan = mergeActionPlanSteps(issueOutput.ActionPlan, reportActionPlan)
			resp.Issues[index].Actions = actionTexts(resp.Issues[index].ActionPlan)
		} else if len(issueOutput.Actions) > 0 {
			resp.Issues[index].ActionPlan = mergeActionPlanSteps(actionsToActionPlanSteps(issueOutput.Actions), reportActionPlan)
			resp.Issues[index].Actions = actionTexts(resp.Issues[index].ActionPlan)
		}
	}
}

func callSystemHealthAI(ctx core.Context, baseURL string, apiKey string, model string, timeoutSecond int, payload systemHealthAISummaryRequest) (systemHealthAISummaryResponse, error) {
	if timeoutSecond <= 0 {
		timeoutSecond = 10
	}
	if strings.TrimSpace(apiKey) == "" {
		return systemHealthAISummaryResponse{}, fmt.Errorf("missing AI API key")
	}
	inputPayload, err := json.Marshal(payload)
	if err != nil {
		return systemHealthAISummaryResponse{}, err
	}
	requestBody, err := json.Marshal(openAIResponsesRequest{
		Model: strings.TrimSpace(model),
		Instructions: strings.Join([]string{
			"你是系统健康中心的告警汇总助手。",
			"只基于输入中的真实监控数据、已有告警分析产出和证据生成总结。",
			"不要解释告警为何被聚合，不要输出置信度，不要编造输入中不存在的根因。",
			"输出必须严格符合请求中的 json_schema。",
			"issues 只返回输入中已有的 issue.key，不能新增、改写或翻译 key。",
			"summary 必须面向普通用户说明系统情况，按“系统当前处于什么状态 + 发生了什么告警事件 + 有哪些需要关注的问题”组织。",
			"summary 控制在 120 个中文字符以内，最多 3 句话；不要使用“整体判断”“风险模式”等内部分析口径。",
			"summary 推荐句式：系统当前处于{状态}状态。窗口内发生过{范围模式}{症状}告警，{当前是否仍触发/是否恢复观察}。需要关注{复发/RED退化/共同根因/规则口径}。",
			"summary 使用“多入口/多服务/单节点/多实例”等范围模式，不要出现具体服务名、接口名、实例名或对象清单。",
			"summary 禁止使用“主要问题集中在 A、B、C”这类服务/接口列表式表达。",
			"每个 issue.analysis 避免重复 evidence 中已有的数字罗列。",
			"existingAlertAnalyses 是告警发生时的分析结果，只作为辅助信息；issue.rootCause 要从这些分析结果、事件规律和指标/APM证据中整理根因判断，不要原样罗列辅助分析。",
			"当多个 existingAlertAnalyses 指向一致时，rootCause 优先归纳共同指向的服务、入口、下游或指标问题；当分析方向冲突时，rootCause 要说明仍为可能根因并指出需要用指标/APM继续确认。",
			"P4 invalidAlert 表示告警持续触发但已有分析反复判定无效，应分析告警指标、表达式、阈值或标签口径是否存在问题；不能当作服务故障处置。",
			"P4 invalidAlert 的 actionPlan 重点融合规则诊断、RED 指标对比和告警规则修正，不要输出扩容、重启等业务止血动作，除非输入证据明确支持。",
			"existingReportActions 来自原告警分析报告的建议行动；生成 actionPlan 时必须融合这些动作，不能无理由丢弃。",
			"actionPlan 只给可执行的排查、止血、恢复和验证步骤；字段为空时使用空字符串，不要省略字段。",
		}, "\n"),
		Input: string(inputPayload),
		Text: openAIResponsesTextConfig{
			Format: openAIResponsesTextFormat{
				Type:   "json_schema",
				Name:   "system_health_summary",
				Strict: true,
				Schema: systemHealthAIResponseSchema(),
			},
		},
	})
	if err != nil {
		return systemHealthAISummaryResponse{}, err
	}
	requestContext := context.Background()
	if ctx != nil {
		requestContext = ctx.GetContext()
	}
	requestContext, cancel := context.WithTimeout(requestContext, time.Duration(timeoutSecond)*time.Second)
	defer cancel()

	endpoint := openAIResponsesEndpoint(baseURL)
	req, err := http.NewRequestWithContext(requestContext, http.MethodPost, endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return systemHealthAISummaryResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apiKey))

	resp, err := (&http.Client{Timeout: time.Duration(timeoutSecond) * time.Second}).Do(req)
	if err != nil {
		return systemHealthAISummaryResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return systemHealthAISummaryResponse{}, fmt.Errorf("AI summary API returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var responsesResp openAIResponsesResponse
	if err := json.NewDecoder(resp.Body).Decode(&responsesResp); err != nil {
		return systemHealthAISummaryResponse{}, err
	}
	if responsesResp.Error != nil {
		return systemHealthAISummaryResponse{}, fmt.Errorf("AI response error: %s", responsesResp.Error.Message)
	}
	aiResp, err := parseSystemHealthAIResponseText(extractOpenAIResponsesText(responsesResp))
	if err != nil {
		return systemHealthAISummaryResponse{}, err
	}
	aiResp.WorkflowRunID = firstNonEmpty(aiResp.WorkflowRunID, responsesResp.ID)
	return aiResp, nil
}

func openAIResponsesEndpoint(baseURL string) string {
	return strings.TrimRight(strings.TrimSpace(baseURL), "/") + "/responses"
}

func systemHealthAIResponseSchema() map[string]any {
	stringSchema := map[string]any{"type": "string"}
	actionPlanItemSchema := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"phase", "target", "action", "validation", "risk"},
		"properties": map[string]any{
			"phase":      stringSchema,
			"target":     stringSchema,
			"action":     stringSchema,
			"validation": stringSchema,
			"risk":       stringSchema,
		},
	}
	issueSchema := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"key", "analysis", "rootCause", "actionPlan"},
		"properties": map[string]any{
			"key":        stringSchema,
			"analysis":   stringSchema,
			"rootCause":  stringSchema,
			"actionPlan": map[string]any{"type": "array", "items": actionPlanItemSchema},
		},
	}
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"summary", "issues"},
		"properties": map[string]any{
			"summary": stringSchema,
			"issues":  map[string]any{"type": "array", "items": issueSchema},
		},
	}
}

func resolveSystemHealthAIAPIKey(apiKey string, envKey string) string {
	if strings.TrimSpace(apiKey) != "" {
		return strings.TrimSpace(apiKey)
	}
	if strings.TrimSpace(envKey) == "" {
		return ""
	}
	return strings.TrimSpace(os.Getenv(strings.TrimSpace(envKey)))
}

func extractOpenAIResponsesText(resp openAIResponsesResponse) string {
	if strings.TrimSpace(resp.OutputText) != "" {
		return strings.TrimSpace(resp.OutputText)
	}
	for _, output := range resp.Output {
		for _, content := range output.Content {
			if strings.TrimSpace(content.Text) != "" {
				return strings.TrimSpace(content.Text)
			}
		}
	}
	return ""
}

func parseSystemHealthAIResponseText(text string) (systemHealthAISummaryResponse, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return systemHealthAISummaryResponse{}, fmt.Errorf("AI response text is empty")
	}
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)
	var aiResp systemHealthAISummaryResponse
	if err := json.Unmarshal([]byte(text), &aiResp); err != nil {
		return systemHealthAISummaryResponse{}, fmt.Errorf("failed to parse AI response JSON: %w", err)
	}
	return aiResp, nil
}

func buildSystemHealthAIRequest(resp *response.SystemHealthReportResp, scored []scoredIssue) systemHealthAISummaryRequest {
	if resp == nil {
		return systemHealthAISummaryRequest{}
	}
	request := systemHealthAISummaryRequest{
		Window: resp.Report.Window,
		Report: systemHealthAIReportInput{
			Status:           resp.Report.Status,
			Score:            resp.Report.Score,
			ServiceRiskCount: resp.Report.ServiceRiskCount,
			ActiveAlertCount: resp.Report.ActiveAlertCount,
			AvailableSources: resp.Report.AvailableSources,
			Summary:          resp.Report.Summary,
		},
		Issues: make([]systemHealthAIIssueInput, 0, minInt(len(resp.Issues), len(scored))),
	}
	for index, issue := range resp.Issues {
		var input reportInput
		if index < len(scored) {
			input = scored[index].input
		}
		request.Issues = append(request.Issues, systemHealthAIIssueInput{
			Key:                   issue.Key,
			Priority:              issue.Priority,
			Title:                 issue.Title,
			Status:                issue.Status,
			Scope:                 issue.Scope,
			Impact:                issue.Impact,
			ActiveAlerts:          issue.ActiveAlerts,
			StartedAt:             issue.StartedAt,
			DataSources:           issue.DataSources,
			Evidence:              buildSystemHealthAIEvidenceInput(issue.Evidence),
			ExistingAlertAnalyses: buildSystemHealthAIAlertAnalysisInput(input.events),
			ExistingReportActions: buildSystemHealthAIReportActionInput(input.alertReportActions),
		})
	}
	return request
}

func buildSystemHealthAIEvidenceInput(evidence []response.SystemHealthEvidence) []systemHealthAIEvidenceInput {
	items := make([]systemHealthAIEvidenceInput, 0, minInt(len(evidence), 4))
	for _, item := range evidence {
		if len(items) >= 4 {
			break
		}
		if strings.TrimSpace(item.Result) == "" {
			continue
		}
		items = append(items, systemHealthAIEvidenceInput{
			Type:     item.Type,
			Source:   item.Source,
			Severity: item.Severity,
			Result:   truncateText(item.Result, 360),
		})
	}
	return items
}

func buildSystemHealthAIAlertAnalysisInput(events []alertmodel.AEventWithWRecord) []systemHealthAIAlertAnalysisInput {
	items := make([]systemHealthAIAlertAnalysisInput, 0, minInt(len(events), maxEvidenceSamples))
	for _, event := range events {
		if len(items) >= maxEvidenceSamples {
			break
		}
		if strings.TrimSpace(event.AlertDirection) == "" && strings.TrimSpace(event.AnalyzeRunID) == "" && strings.TrimSpace(event.AnalyzeErr) == "" && strings.TrimSpace(event.Output) == "" {
			continue
		}
		items = append(items, systemHealthAIAlertAnalysisInput{
			EventID:        event.EventID,
			AlertID:        event.AlertID,
			AlertName:      event.Name,
			AlertDirection: truncateText(event.AlertDirection, 220),
			AnalyzeRunID:   event.AnalyzeRunID,
			AnalyzeError:   truncateText(event.AnalyzeErr, 220),
			CheckOutput:    truncateText(event.Output, 260),
			Validity:       firstNonEmpty(event.Validity, event.IsValid),
		})
	}
	return items
}

func buildSystemHealthAIReportActionInput(actions []systemHealthAlertReportAction) []systemHealthAIReportActionInput {
	items := make([]systemHealthAIReportActionInput, 0, minInt(len(actions), maxEvidenceSamples))
	for _, action := range actions {
		if len(items) >= maxEvidenceSamples {
			break
		}
		if strings.TrimSpace(action.Action) == "" {
			continue
		}
		items = append(items, systemHealthAIReportActionInput{
			AlertName: action.AlertName,
			Level:     action.Level,
			Action:    truncateText(action.Action, 220),
		})
	}
	return items
}

func buildIssueExistingAlertAnalyses(events []alertmodel.AEventWithWRecord) []response.SystemHealthAlertAnalysis {
	items := make([]response.SystemHealthAlertAnalysis, 0, minInt(len(events), maxEvidenceSamples))
	seen := make(map[string]struct{})
	for _, event := range events {
		if len(items) >= maxEvidenceSamples {
			break
		}
		if strings.TrimSpace(event.AlertDirection) == "" && strings.TrimSpace(event.AnalyzeRunID) == "" && strings.TrimSpace(event.AnalyzeErr) == "" && strings.TrimSpace(event.Output) == "" {
			continue
		}
		key := strings.Join([]string{
			event.EventID,
			event.AlertID,
			event.Name,
			event.AlertDirection,
			event.AnalyzeRunID,
		}, "\x00")
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		items = append(items, response.SystemHealthAlertAnalysis{
			EventID:        event.EventID,
			AlertID:        event.AlertID,
			AlertName:      event.Name,
			AlertDirection: truncateText(event.AlertDirection, 300),
			AnalyzeRunID:   event.AnalyzeRunID,
			AnalyzeError:   truncateText(event.AnalyzeErr, 220),
			CheckOutput:    truncateText(event.Output, 260),
			Validity:       firstNonEmpty(event.Validity, event.IsValid),
		})
	}
	return items
}

func (h *handler) queryAlertReportActions(ctx core.Context, input reportInput, startTime time.Time, endTime time.Time) []systemHealthAlertReportAction {
	if h == nil || h.chRepo == nil || ctx == nil || len(input.events) == 0 {
		return nil
	}
	actions := make([]systemHealthAlertReportAction, 0, maxEvidenceSamples)
	seenEvents := make(map[string]struct{}, len(input.events))
	reportStartTime := startTime.Add(-time.Hour)
	reportEndTime := endTime.Add(time.Hour)
	checked := 0
	for _, event := range input.events {
		if len(actions) >= maxEvidenceSamples {
			break
		}
		if checked >= maxAlertReportLookups {
			break
		}
		eventID := strings.TrimSpace(event.EventID)
		if eventID == "" {
			continue
		}
		if _, ok := seenEvents[eventID]; ok {
			continue
		}
		seenEvents[eventID] = struct{}{}
		checked++
		report, err := h.chRepo.GetAlertEventReport(ctx, reportStartTime, reportEndTime, map[string]string{"alertEventId": eventID})
		if err != nil || report == nil {
			continue
		}
		actions = append(actions, parseAlertReportSuggestActions(report.Suggest, event.Name)...)
		actions = uniqueAlertReportActions(actions)
	}
	if len(actions) > maxEvidenceSamples {
		actions = actions[:maxEvidenceSamples]
	}
	return actions
}

func parseAlertReportSuggestActions(suggests []string, alertName string) []systemHealthAlertReportAction {
	actions := make([]systemHealthAlertReportAction, 0, len(suggests))
	for _, raw := range suggests {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		var item map[string]any
		if err := json.Unmarshal([]byte(raw), &item); err != nil {
			actions = append(actions, systemHealthAlertReportAction{
				AlertName: alertName,
				Action:    truncateText(raw, 220),
			})
			continue
		}
		actionText := firstNonEmpty(
			stringValueFromMap(item, "action"),
			stringValueFromMap(item, "title"),
			stringValueFromMap(item, "content"),
			stringValueFromMap(item, "description"),
			stringValueFromMap(item, "suggestion"),
		)
		if actionText == "" {
			continue
		}
		actions = append(actions, systemHealthAlertReportAction{
			AlertName: alertName,
			Level:     strings.ToLower(stringValueFromMap(item, "level")),
			Action:    truncateText(actionText, 220),
		})
	}
	return uniqueAlertReportActions(actions)
}

func stringValueFromMap(item map[string]any, key string) string {
	if item == nil {
		return ""
	}
	value, ok := item[key]
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case fmt.Stringer:
		return strings.TrimSpace(typed.String())
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}

func uniqueAlertReportActions(actions []systemHealthAlertReportAction) []systemHealthAlertReportAction {
	seen := make(map[string]struct{}, len(actions))
	result := make([]systemHealthAlertReportAction, 0, len(actions))
	for _, action := range actions {
		action.Action = strings.TrimSpace(action.Action)
		if action.Action == "" {
			continue
		}
		key := strings.ToLower(action.Action)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, action)
	}
	return result
}

func trimNonEmptyStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		result = append(result, value)
	}
	return result
}

func (h *handler) queryRealAlertEvents(ctx core.Context, groupID int64, startTime time.Time, endTime time.Time) ([]alertmodel.AEventWithWRecord, error) {
	req := &request.AlertEventSearchRequest{
		StartTime: startTime.UnixMicro(),
		EndTime:   endTime.UnixMicro(),
		GroupID:   groupID,
		Pagination: &model.Pagination{
			CurrentPage: 1,
			PageSize:    500,
		},
		SortBy: "isValid desc,receivedTime desc",
	}
	if groupID > 0 {
		for _, id := range common.DataGroupStorage.GetFullSubGroupIDs(groupID) {
			req.SubGroupIDs = append(req.SubGroupIDs, strconv.FormatInt(id, 10))
		}
		req.SubGroupIDs = append(req.SubGroupIDs, strconv.FormatInt(groupID, 10))
	}

	events, _, err := h.chRepo.GetAlertEventWithWorkflowRecord(ctx, req, 0)
	return events, err
}

func (h *handler) queryAlertManagerActiveEvents(ctx core.Context, groupID int64) ([]alertmodel.AEventWithWRecord, error) {
	if strings.TrimSpace(h.alertManagerURL) == "" {
		return nil, nil
	}
	endpoint := strings.TrimRight(h.alertManagerURL, "/") + "/api/v2/alerts"
	req, err := http.NewRequestWithContext(ctx.GetContext(), http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := (&http.Client{Timeout: 8 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("alertmanager returned status %d", resp.StatusCode)
	}

	var alerts []alertManagerAPIAlert
	if err := json.NewDecoder(resp.Body).Decode(&alerts); err != nil {
		return nil, err
	}

	allowedGroupIDs := allowedAlertGroupIDs(groupID)
	events := make([]alertmodel.AEventWithWRecord, 0, len(alerts))
	for _, alert := range alerts {
		event, ok := alertManagerAlertToEvent(alert, allowedGroupIDs)
		if ok {
			events = append(events, event)
		}
	}
	return events, nil
}

func allowedAlertGroupIDs(groupID int64) map[string]struct{} {
	if groupID <= 0 {
		return nil
	}
	allowed := map[string]struct{}{
		strconv.FormatInt(groupID, 10): {},
	}
	if common.DataGroupStorage != nil {
		for _, id := range common.DataGroupStorage.GetFullSubGroupIDs(groupID) {
			allowed[strconv.FormatInt(id, 10)] = struct{}{}
		}
	}
	return allowed
}

func alertManagerAlertToEvent(alert alertManagerAPIAlert, allowedGroupIDs map[string]struct{}) (alertmodel.AEventWithWRecord, bool) {
	if alert.Status.State != "" && alert.Status.State != "active" && alert.Status.State != alertmodel.StatusFiring {
		return alertmodel.AEventWithWRecord{}, false
	}
	if len(allowedGroupIDs) > 0 {
		if _, ok := allowedGroupIDs[alertGroupIDLabel(alert.Labels)]; !ok {
			return alertmodel.AEventWithWRecord{}, false
		}
	}
	name := strings.TrimSpace(alert.Labels["alertname"])
	if name == "" {
		return alertmodel.AEventWithWRecord{}, false
	}
	startTime := parseAlertManagerTime(alert.StartsAt, time.Now())
	endTime := parseAlertManagerTime(alert.EndsAt, time.Time{})
	annotations := make(map[string]string, len(alert.Annotations)+2)
	for key, value := range alert.Annotations {
		annotations[key] = value
	}
	if alert.GeneratorURL != "" {
		annotations["generatorURL"] = alert.GeneratorURL
	}
	if alert.Fingerprint != "" {
		annotations["fingerprint"] = alert.Fingerprint
	}
	detail, _ := json.Marshal(annotations)

	rawTags := make(alertmodel.RawTags, len(alert.Labels))
	enrichTags := make(map[string]string, len(alert.Labels))
	for key, value := range alert.Labels {
		rawTags[key] = value
		enrichTags[key] = value
	}
	group := firstNonEmpty(alert.Labels["group"], alert.Labels["alertgroup"])
	event := alertmodel.AEventWithWRecord{
		AlertEvent: alertmodel.AlertEvent{
			Alert: alertmodel.Alert{
				Source:     "alertManager",
				SourceID:   "alertmanager-api",
				AlertID:    alertmodel.FastAlertIDByStringMap(name, alert.Labels),
				Group:      group,
				Name:       name,
				EnrichTags: enrichTags,
				Tags:       rawTags,
			},
			Detail:       string(detail),
			CreateTime:   startTime,
			UpdateTime:   time.Now(),
			EndTime:      endTime,
			ReceivedTime: time.Now(),
			Severity:     firstNonEmpty(alert.Labels["severity"], alertmodel.SeverityUnknownLevel),
			Status:       alertmodel.StatusFiring,
		},
		WorkflowDetail: alertmodel.WorkflowDetail{
			Validity: "valid",
			IsValid:  "true",
		},
	}
	return event, true
}

func alertGroupIDLabel(labels map[string]string) string {
	for _, key := range []string{"groupId", "groupID", "group_id", "alertgroup", "alertGroup"} {
		if value := strings.TrimSpace(labels[key]); value != "" {
			return value
		}
	}
	return ""
}

func parseAlertManagerTime(value string, fallback time.Time) time.Time {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err == nil {
		return parsed
	}
	return fallback
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func mergeAlertEvents(base []alertmodel.AEventWithWRecord, extra []alertmodel.AEventWithWRecord) []alertmodel.AEventWithWRecord {
	if len(extra) == 0 {
		return base
	}
	seen := make(map[string]struct{}, len(base))
	for _, event := range base {
		seen[event.AlertID+"::"+event.Status] = struct{}{}
	}
	for _, event := range extra {
		key := event.AlertID + "::" + event.Status
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		base = append(base, event)
	}
	return base
}

func (h *handler) buildInputsFromRealEvents(ctx core.Context, groupID int64, events []alertmodel.AEventWithWRecord, startTime time.Time, endTime time.Time, capabilities telemetryCapabilityStatus) []reportInput {
	grouped := make(map[string]*reportInput)
	for _, event := range events {
		kind := classifyIssueKind(event)
		if kind == "" {
			continue
		}
		normalized := normalizeAlertEvent(event)
		scope, scopeKind := normalizedAggregationScope(normalized, event)
		aggregationKey := normalized.AggregationKey
		if aggregationKey == "" {
			aggregationKey = scopeKind + "::" + scope
		}
		key := kind + "::" + aggregationKey
		input, exists := grouped[key]
		if !exists {
			input = &reportInput{
				key:                  stableKey(key),
				aggregationKey:       aggregationKey,
				kind:                 kind,
				title:                eventTitle(event, kind),
				scope:                scope,
				scopeKind:            scopeKind,
				startedAt:            formatEventTime(event.CreateTime),
				updatedAt:            formatEventTime(event.UpdateTime),
				availableCalcSources: []string{"alerts"},
				displaySources:       []string{sourceLabel("alerts")},
				affectedObjects:      affectedObjectsFromEvent(event),
				objectCriticality:    objectCriticality(normalized.ObjectType, normalized.ObjectName, normalized.ServiceName, normalized.Endpoint),
			}
			grouped[key] = input
		}

		input.events = append(input.events, event)
		input.normalizedEvents = append(input.normalizedEvents, normalized)
		input.alertNames = append(input.alertNames, event.Name)
		input.alertSeverities = append(input.alertSeverities, normalized.Severity)
		input.affectedObjects = uniqueAffectedObjects(append(input.affectedObjects, affectedObjectsFromEvent(event)...))
		input.objectCriticality = math.Max(input.objectCriticality, objectCriticality(normalized.ObjectType, normalized.ObjectName, normalized.ServiceName, normalized.Endpoint))
		if event.CreateTime.Before(parseDisplayTime(input.startedAt)) {
			input.startedAt = formatEventTime(event.CreateTime)
		}
		if event.UpdateTime.After(parseDisplayTime(input.updatedAt)) {
			input.updatedAt = formatEventTime(event.UpdateTime)
		}
	}

	inputs := make([]reportInput, 0, len(grouped))
	for _, input := range grouped {
		finalizeReportInput(input, endTime)
		input.alertNames = uniqueStrings(input.alertNames)
		input.title = buildAggregatedIssueTitle(*input)
		input.alertReportActions = h.queryAlertReportActions(ctx, *input, startTime, endTime)
		if len(input.alertReportActions) > 0 {
			input.availableCalcSources = append(input.availableCalcSources, "alertReport")
			input.displaySources = append(input.displaySources, sourceLabel("alertReport"))
		}
		input.evidence = h.buildRealEvidence(ctx, *input, groupID, startTime, endTime, capabilities)
		input.displaySources = uniqueStrings(append(input.displaySources, displaySourcesFromEvidence(input.evidence)...))
		input.impact = buildRealImpact(*input)
		inputs = append(inputs, *input)
	}
	return inputs
}

func displaySourcesFromEvidence(evidence []response.SystemHealthEvidence) []string {
	sources := make([]string, 0, 1)
	for _, item := range evidence {
		if item.Type == "Trace链路样本" {
			sources = append(sources, sourceLabel("trace"))
		}
	}
	return sources
}

func normalizeAlertEvent(event alertmodel.AEventWithWRecord) normalizedAlertEvent {
	serviceName := alertServiceName(event)
	endpoint := alertEndpoint(event)
	objectType, objectName := primaryAlertObject(event, serviceName, endpoint)
	updatedAt := event.UpdateTime
	if updatedAt.IsZero() {
		updatedAt = event.ReceivedTime
	}
	if updatedAt.IsZero() {
		updatedAt = event.CreateTime
	}
	normalized := normalizedAlertEvent{
		Source:      firstNonEmpty(event.Source, event.SourceID, "alert_event"),
		EventID:     firstNonEmpty(event.EventID, event.AlertID),
		AlertName:   strings.TrimSpace(event.Name),
		Status:      normalizeStatus(event.Status),
		Severity:    normalizeSeverity(event.Severity),
		ServiceName: serviceName,
		Endpoint:    endpoint,
		ObjectType:  objectType,
		ObjectName:  objectName,
		Group:       strings.TrimSpace(event.Group),
		StartedAt:   event.CreateTime,
		UpdatedAt:   updatedAt,
		RecoveredAt: event.EndTime,
	}
	if normalized.StartedAt.IsZero() {
		normalized.StartedAt = updatedAt
	}
	normalized.DedupKey = buildDedupKey(normalized)
	normalized.AggregationKey = buildAggregationKey(normalized)
	return normalized
}

func normalizeStatus(status string) string {
	normalized := strings.ToLower(strings.TrimSpace(status))
	if normalized == alertmodel.StatusFiring || normalized == "active" {
		return alertmodel.StatusFiring
	}
	if normalized == alertmodel.StatusResolved || normalized == "recovered" || normalized == "recover" {
		return alertmodel.StatusResolved
	}
	return "unknown"
}

func normalizeSeverity(severity string) string {
	switch strings.ToLower(strings.TrimSpace(severity)) {
	case "critical", "disaster", "p0":
		return alertmodel.SeverityCriticalLevel
	case "error", "high", "average", "major", "p1":
		return alertmodel.SeverityErrorLevel
	case "warning", "warn", "minor", "p2":
		return alertmodel.SeverityWarnLevel
	case "info", "information", "notice", "p3":
		return alertmodel.SeverityInfoLevel
	case "":
		return alertmodel.SeverityUnknownLevel
	default:
		return strings.ToLower(strings.TrimSpace(severity))
	}
}

func primaryAlertObject(event alertmodel.AEventWithWRecord, serviceName string, endpoint string) (string, string) {
	candidates := []response.SystemHealthAffectedObject{
		{Type: "endpoint", Name: endpoint},
		{Type: "service", Name: serviceName},
		{Type: "database", Name: event.GetDatabaseURL()},
		{Type: "pod", Name: event.GetK8sPodTag()},
		{Type: "node", Name: event.GetInfraNodeTag()},
		{Type: "namespace", Name: event.GetK8sNamespaceTag()},
		{Type: "target", Name: event.GetTargetObj()},
		{Type: "group", Name: event.Group},
	}
	for _, candidate := range candidates {
		if strings.TrimSpace(candidate.Name) != "" {
			return candidate.Type, strings.TrimSpace(candidate.Name)
		}
	}
	return "unknown", firstNonEmpty(event.AlertID, event.EventID, event.Name, "unknown")
}

func buildDedupKey(event normalizedAlertEvent) string {
	alertName := firstNonEmpty(event.AlertName, "unknown-alert")
	objectType := firstNonEmpty(event.ObjectType, "unknown")
	objectName := firstNonEmpty(event.ObjectName, event.ServiceName, event.Group, "unknown")
	return strings.Join([]string{
		strings.ToLower(firstNonEmpty(event.Source, "alert_event")),
		strings.ToLower(alertName),
		strings.ToLower(objectType),
		strings.ToLower(objectName),
	}, "::")
}

func buildAggregationKey(event normalizedAlertEvent) string {
	if event.ServiceName != "" && event.Endpoint != "" {
		return "entry::" + event.ServiceName + "::" + event.Endpoint
	}
	if event.ServiceName != "" {
		return "service::" + event.ServiceName
	}
	if event.ObjectName != "" {
		return "object::" + event.ObjectType + "::" + event.ObjectName
	}
	return "group::" + firstNonEmpty(event.Group, "unknown")
}

func normalizedAggregationScope(event normalizedAlertEvent, rawEvent alertmodel.AEventWithWRecord) (string, string) {
	if event.ServiceName != "" && event.Endpoint != "" {
		return event.ServiceName + " / " + event.Endpoint, "entry"
	}
	if event.ServiceName != "" {
		return event.ServiceName, "service"
	}
	scope := eventScope(rawEvent)
	if scope == "" {
		scope = firstNonEmpty(event.ObjectName, event.Group, "unknown")
	}
	return scope, firstNonEmpty(event.ObjectType, "object")
}

func objectCriticality(objectType string, objectName string, serviceName string, endpoint string) float64 {
	objectType = strings.ToLower(strings.TrimSpace(objectType))
	objectName = strings.ToLower(strings.TrimSpace(objectName))
	switch {
	case serviceName != "" && endpoint != "":
		return 90
	case serviceName != "":
		return 80
	case objectType == "database" || strings.Contains(objectName, "db") || strings.Contains(objectName, "mysql") || strings.Contains(objectName, "postgres"):
		return 85
	case strings.Contains(objectName, "redis") || strings.Contains(objectName, "cache"):
		return 75
	case strings.Contains(objectName, "mq") || strings.Contains(objectName, "kafka") || strings.Contains(objectName, "queue"):
		return 75
	case objectType == "pod":
		return 65
	case objectType == "node":
		return 60
	case objectType == "namespace":
		return 45
	default:
		return 35
	}
}

func finalizeReportInput(input *reportInput, endTime time.Time) {
	input.rawEventCount = len(input.events)
	input.deduplicatedInstances = deduplicateAlertInstances(input.normalizedEvents, endTime)
	input.deduplicatedAlerts = len(input.deduplicatedInstances)
	input.activeAlerts = 0
	input.repeatCount = 0
	input.durationSeconds = 0
	input.alertSeverities = input.alertSeverities[:0]
	input.mergeReasons = buildMergeReasons(*input)
	for _, instance := range input.deduplicatedInstances {
		input.repeatCount += maxInt(instance.RepeatCount-1, 0)
		input.durationSeconds = math.Max(input.durationSeconds, instance.DurationSeconds)
		input.alertSeverities = append(input.alertSeverities, instance.MaxSeverity)
		if instance.Status == alertmodel.StatusFiring && input.kind == issueKindActiveAlert {
			input.activeAlerts++
		}
	}
	if input.deduplicatedAlerts == 0 {
		input.deduplicatedAlerts = len(input.events)
		input.activeAlerts = len(input.events)
	}
}

func deduplicateAlertInstances(events []normalizedAlertEvent, endTime time.Time) []deduplicatedAlertInstance {
	grouped := make(map[string]*deduplicatedAlertInstance, len(events))
	for _, event := range events {
		key := event.DedupKey
		if key == "" {
			key = buildDedupKey(event)
		}
		instance, exists := grouped[key]
		if !exists {
			instance = &deduplicatedAlertInstance{
				Key:         key,
				AlertName:   event.AlertName,
				ObjectType:  event.ObjectType,
				ObjectName:  event.ObjectName,
				Status:      event.Status,
				MaxSeverity: event.Severity,
				FirstSeenAt: event.StartedAt,
				LastSeenAt:  event.UpdatedAt,
				RepeatCount: 0,
			}
			grouped[key] = instance
		}
		instance.RepeatCount++
		if event.StartedAt.Before(instance.FirstSeenAt) || instance.FirstSeenAt.IsZero() {
			instance.FirstSeenAt = event.StartedAt
		}
		if event.UpdatedAt.After(instance.LastSeenAt) {
			instance.LastSeenAt = event.UpdatedAt
		}
		if severityRisk(event.Severity) > severityRisk(instance.MaxSeverity) {
			instance.MaxSeverity = event.Severity
		}
		if event.Status == alertmodel.StatusFiring {
			instance.Status = alertmodel.StatusFiring
		}
	}
	instances := make([]deduplicatedAlertInstance, 0, len(grouped))
	for _, instance := range grouped {
		referenceTime := endTime
		if referenceTime.IsZero() {
			referenceTime = time.Now()
		}
		if instance.Status != alertmodel.StatusFiring && !instance.LastSeenAt.IsZero() {
			referenceTime = instance.LastSeenAt
		}
		if !instance.FirstSeenAt.IsZero() && referenceTime.After(instance.FirstSeenAt) {
			instance.DurationSeconds = referenceTime.Sub(instance.FirstSeenAt).Seconds()
		}
		instances = append(instances, *instance)
	}
	sort.Slice(instances, func(i, j int) bool {
		if severityRisk(instances[i].MaxSeverity) != severityRisk(instances[j].MaxSeverity) {
			return severityRisk(instances[i].MaxSeverity) > severityRisk(instances[j].MaxSeverity)
		}
		return instances[i].LastSeenAt.After(instances[j].LastSeenAt)
	})
	return instances
}

func buildMergeReasons(input reportInput) []string {
	reasons := make([]string, 0, 4)
	if input.scopeKind == "entry" {
		reasons = append(reasons, "同一业务入口")
	} else if input.scopeKind == "service" {
		reasons = append(reasons, "同一服务")
	} else {
		reasons = append(reasons, "同一告警对象")
	}
	if len(input.normalizedEvents) > 1 {
		reasons = append(reasons, "同一时间窗口")
	}
	if len(uniqueStrings(input.alertNames)) > 1 {
		reasons = append(reasons, "多类告警共同指向")
	}
	if input.deduplicatedAlerts > 0 && input.rawEventCount > input.deduplicatedAlerts {
		reasons = append(reasons, "重复触发已去重")
	}
	return uniqueStrings(reasons)
}

func classifyIssueKind(event alertmodel.AEventWithWRecord) string {
	if event.Status == alertmodel.StatusFiring && isInvalidAlert(event) {
		return issueKindInvalidAlert
	}
	if event.Status == alertmodel.StatusResolved && isValidAlert(event) {
		return issueKindRecoveredAlert
	}
	if isInspectionEvent(event) && event.Status == alertmodel.StatusFiring && !isInvalidAlert(event) {
		return issueKindInspection
	}
	if event.Status == alertmodel.StatusFiring && isValidAlert(event) {
		return issueKindActiveAlert
	}
	return ""
}

func isValidAlert(event alertmodel.AEventWithWRecord) bool {
	return event.Validity == "valid" || event.IsValid == "true"
}

func isInvalidAlert(event alertmodel.AEventWithWRecord) bool {
	return event.Validity == "invalid" || event.IsValid == "false"
}

func isInspectionEvent(event alertmodel.AEventWithWRecord) bool {
	text := strings.ToLower(strings.Join([]string{event.Group, event.Source, event.Name, event.Detail}, " "))
	return strings.Contains(text, "mutation") ||
		strings.Contains(text, "detect") ||
		strings.Contains(text, "inspection") ||
		strings.Contains(text, "巡检")
}

func eventScope(event alertmodel.AEventWithWRecord) string {
	candidates := []string{
		alertServiceName(event),
		event.GetTargetObj(),
		event.GetInfraNodeTag(),
		event.GetDatabaseURL(),
		event.GetK8sNamespaceTag(),
		event.GetK8sPodTag(),
	}
	for _, candidate := range candidates {
		if strings.TrimSpace(candidate) != "" {
			return candidate
		}
	}
	return strings.TrimSpace(event.Group)
}

func eventAggregationScope(event alertmodel.AEventWithWRecord) (string, string) {
	if service := alertServiceName(event); service != "" {
		return service, "service"
	}
	scope := eventScope(event)
	if scope == "" {
		scope = strings.TrimSpace(event.Group)
	}
	return scope, "object"
}

func alertServiceName(event alertmodel.AEventWithWRecord) string {
	rawTags := rawTagsFromEvent(event)
	return firstNonEmpty(
		event.GetServiceNameTag(),
		rawTags["serviceName"],
		rawTags["svc_name"],
		rawTags["service_name"],
	)
}

func eventTitle(event alertmodel.AEventWithWRecord, kind string) string {
	if kind == issueKindInspection {
		return "巡检指标异常：" + event.Name
	}
	if kind == issueKindRecoveredAlert {
		return "已恢复告警：" + event.Name
	}
	if kind == issueKindInvalidAlert {
		return "疑似告警规则问题：" + event.Name
	}
	return event.Name
}

func buildAggregatedIssueTitle(input reportInput) string {
	alertNames := uniqueStrings(input.alertNames)
	baseTitle := firstNonEmpty(firstString(alertNames), input.title)
	if len(alertNames) > 1 {
		baseTitle = fmt.Sprintf("%s等 %d 类告警", baseTitle, len(alertNames))
	}
	if input.kind == issueKindInspection {
		return "巡检指标异常：" + baseTitle
	}
	if input.kind == issueKindRecoveredAlert {
		return "已恢复告警：" + baseTitle
	}
	if input.kind == issueKindInvalidAlert {
		return "疑似告警规则问题：" + baseTitle
	}
	return baseTitle
}

func firstString(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func buildAlertEventEvidenceResult(input reportInput) string {
	alertNames := uniqueStrings(input.alertNames)
	directions := alertDirections(input.events)
	patternSummary := buildAlertEventPatternSummary(input, alertNames, directions)
	parts := []string{
		fmt.Sprintf("%d 条真实告警事件", len(input.events)),
		fmt.Sprintf("%d 个去重告警实例", input.deduplicatedAlerts),
		fmt.Sprintf("%d 类告警", len(alertNames)),
		"状态=" + statusText(input),
		"聚合对象=" + input.scope,
	}
	if patternSummary != "" {
		parts = append(parts, "事件规律="+patternSummary)
	}
	if input.repeatCount > 0 {
		parts = append(parts, fmt.Sprintf("重复触发=%d次", input.repeatCount))
	}
	if input.durationSeconds > 0 {
		parts = append(parts, "最长持续="+formatDurationCN(input.durationSeconds))
	}
	if len(alertNames) > 0 {
		parts = append(parts, "告警类型="+strings.Join(alertNames, "、"))
	}
	if len(directions) > 0 {
		parts = append(parts, "告警方向="+strings.Join(directions, "、"))
	}
	return truncateText(strings.Join(parts, "，"), 360)
}

func buildAlertEventPatternSummary(input reportInput, alertNames []string, directions []string) string {
	parts := make([]string, 0, 5)
	eventCount := len(input.events)
	if eventCount == 0 {
		return ""
	}
	if input.repeatCount > 0 {
		parts = append(parts, fmt.Sprintf("同一告警实例反复出现%d次", input.repeatCount))
	} else if eventCount > 1 && input.deduplicatedAlerts <= 1 {
		parts = append(parts, "同一对象出现多次同类事件")
	}
	if len(alertNames) > 1 {
		parts = append(parts, fmt.Sprintf("%d类告警在同一对象聚合", len(alertNames)))
	}
	if input.deduplicatedAlerts > 1 {
		parts = append(parts, fmt.Sprintf("涉及%d个去重实例", input.deduplicatedAlerts))
	}
	if timePattern := alertEventTimePattern(input.events); timePattern != "" {
		parts = append(parts, timePattern)
	}
	if len(directions) == 1 && eventCount > 1 {
		parts = append(parts, "已有分析方向一致："+truncateText(directions[0], 80))
	}
	return truncateText(strings.Join(parts, "；"), 260)
}

func alertEventTimePattern(events []alertmodel.AEventWithWRecord) string {
	if len(events) <= 1 {
		return ""
	}
	var first time.Time
	var last time.Time
	for _, event := range events {
		eventTime := event.UpdateTime
		if eventTime.IsZero() {
			eventTime = event.CreateTime
		}
		if eventTime.IsZero() {
			continue
		}
		if first.IsZero() || eventTime.Before(first) {
			first = eventTime
		}
		if last.IsZero() || eventTime.After(last) {
			last = eventTime
		}
	}
	if first.IsZero() || last.IsZero() || !last.After(first) {
		return ""
	}
	spread := last.Sub(first)
	switch {
	case spread <= 30*time.Minute:
		return "事件在30分钟内集中出现"
	case spread <= 2*time.Hour:
		return "事件在2小时内连续出现"
	default:
		return "事件跨" + formatDurationCN(spread.Seconds()) + "持续出现"
	}
}

func buildAlertEventEvidenceSamples(input reportInput) []response.SystemHealthAlertEventEvidence {
	events := make([]response.SystemHealthAlertEventEvidence, 0, minInt(maxEvidenceSamples, len(input.events)))
	for _, event := range input.events {
		if len(events) >= maxEvidenceSamples {
			break
		}
		summary := alertDetailSummary(event)
		if summary == "" {
			summary = alertMetricSummary(event)
		}
		events = append(events, response.SystemHealthAlertEventEvidence{
			EventID:   event.EventID,
			AlertID:   event.AlertID,
			Name:      event.Name,
			Group:     event.Group,
			Source:    displaySource(event),
			Status:    event.Status,
			Severity:  event.Severity,
			StartedAt: formatEventTime(event.CreateTime),
			UpdatedAt: formatEventTime(event.UpdateTime),
			EndedAt:   formatEventTime(event.EndTime),
			Summary:   truncateText(summary, 220),
			Labels:    selectedAlertLabelMap(rawTagsFromEvent(event), 8),
		})
	}
	return events
}

func (h *handler) relatedObjectStatusEvidence(ctx core.Context, input reportInput) response.SystemHealthEvidence {
	evidence := response.SystemHealthEvidence{
		Type:     "相关对象状态",
		Source:   "Kubernetes",
		Severity: maxSeverity(input.alertSeverities),
		Query:    "kubernetes.namespace + kubernetes.pod.status",
	}
	if len(input.events) == 0 {
		return evidence
	}
	statuses, warnings := h.queryRelatedObjectStatuses(ctx, input.events[:minInt(maxEvidenceSamples, len(input.events))])
	if len(statuses) == 0 && len(warnings) == 0 {
		return evidence
	}
	evidence.ObjectStatuses = statuses
	evidence.Result = buildRelatedObjectStatusResult(statuses, warnings)
	return evidence
}

func (h *handler) queryRelatedObjectStatuses(ctx core.Context, events []alertmodel.AEventWithWRecord) ([]response.SystemHealthObjectStatusEvidence, []string) {
	if len(events) == 0 {
		return nil, nil
	}
	if h.k8sRepo == nil {
		return nil, []string{"Kubernetes 仓库未初始化，无法查询相关对象状态"}
	}

	candidates := relatedObjectStatusCandidates(events)
	statuses := make([]response.SystemHealthObjectStatusEvidence, 0, len(candidates)*2)
	warnings := make([]string, 0)
	seenNamespaces := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		if candidate.Namespace == "" && candidate.Pod != "" {
			warnings = append(warnings, fmt.Sprintf("Pod %s 缺少 namespace，无法查询 Kubernetes 状态", candidate.Pod))
			continue
		}
		if candidate.Namespace != "" {
			if _, exists := seenNamespaces[candidate.Namespace]; !exists {
				seenNamespaces[candidate.Namespace] = struct{}{}
				namespace, err := h.k8sRepo.GetNamespaceInfo(candidate.Namespace)
				if err != nil {
					warnings = append(warnings, fmt.Sprintf("namespace %s 查询失败：%s", candidate.Namespace, err.Error()))
				} else if namespace != nil {
					statuses = append(statuses, namespaceObjectStatus(namespace))
				}
			}
		}
		if candidate.Pod != "" && candidate.Namespace != "" {
			pod, err := h.k8sRepo.GetPodInfo(candidate.Namespace, candidate.Pod)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("pod %s/%s 查询失败：%s", candidate.Namespace, candidate.Pod, err.Error()))
				continue
			}
			if pod != nil {
				statuses = append(statuses, podObjectStatus(pod, candidate.Container))
			}
			continue
		}
		if candidate.Node != "" {
			statuses = append(statuses, response.SystemHealthObjectStatusEvidence{
				Type:    "node",
				Name:    candidate.Node,
				Status:  "unknown",
				Message: "当前 Kubernetes 仓库未提供 Node 状态查询接口，仅记录告警事件中的节点标签。",
			})
		}
	}
	return statuses, uniqueStrings(warnings)
}

func relatedObjectStatusCandidates(events []alertmodel.AEventWithWRecord) []relatedObjectStatusCandidate {
	seen := make(map[string]struct{}, len(events))
	candidates := make([]relatedObjectStatusCandidate, 0, len(events))
	for _, event := range events {
		rawTags := rawTagsFromEvent(event)
		candidate := relatedObjectStatusCandidate{
			Namespace: firstNonEmpty(event.GetK8sNamespaceTag(), rawTags["namespace"], rawTags["src_namespace"]),
			Pod:       firstNonEmpty(event.GetK8sPodTag(), rawTags["pod"], rawTags["pod_name"], rawTags["k8s_pod"]),
			Container: firstNonEmpty(event.GetContainerTag(), rawTags["container"], rawTags["container_name"]),
			Node:      firstNonEmpty(event.GetInfraNodeTag(), rawTags["node_name"], rawTags["node"], rawTags["instance_name"]),
		}
		key := strings.Join([]string{candidate.Namespace, candidate.Pod, candidate.Container, candidate.Node}, "\xff")
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		candidates = append(candidates, candidate)
	}
	return candidates
}

func namespaceObjectStatus(namespace *corev1.Namespace) response.SystemHealthObjectStatusEvidence {
	return response.SystemHealthObjectStatusEvidence{
		Type:       "namespace",
		Name:       namespace.Name,
		Status:     string(namespace.Status.Phase),
		AgeSeconds: objectAgeSeconds(namespace.CreationTimestamp.Time),
		Labels:     selectedStatusLabels(namespace.Labels),
		Conditions: namespaceConditions(namespace.Status.Conditions),
	}
}

func podObjectStatus(pod *corev1.Pod, targetContainer string) response.SystemHealthObjectStatusEvidence {
	ready := podReady(pod)
	containers := containerStatuses(pod.Status.InitContainerStatuses, pod.Status.ContainerStatuses, targetContainer)
	return response.SystemHealthObjectStatusEvidence{
		Type:       "pod",
		Name:       pod.Name,
		Namespace:  pod.Namespace,
		Status:     string(pod.Status.Phase),
		Ready:      &ready,
		Restarts:   podRestartCount(pod.Status.InitContainerStatuses, pod.Status.ContainerStatuses),
		NodeName:   pod.Spec.NodeName,
		AgeSeconds: objectAgeSeconds(pod.CreationTimestamp.Time),
		Message:    podStatusMessage(pod),
		Labels:     selectedStatusLabels(pod.Labels),
		Conditions: podConditions(pod.Status.Conditions),
		Containers: containers,
	}
}

func podReady(pod *corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}

func podRestartCount(initContainers []corev1.ContainerStatus, containers []corev1.ContainerStatus) int {
	total := 0
	for _, container := range initContainers {
		total += int(container.RestartCount)
	}
	for _, container := range containers {
		total += int(container.RestartCount)
	}
	return total
}

func podStatusMessage(pod *corev1.Pod) string {
	parts := make([]string, 0, 2)
	if pod.Status.Reason != "" {
		parts = append(parts, pod.Status.Reason)
	}
	if pod.Status.Message != "" {
		parts = append(parts, pod.Status.Message)
	}
	for _, condition := range pod.Status.Conditions {
		if condition.Status == corev1.ConditionFalse && (condition.Reason != "" || condition.Message != "") {
			parts = append(parts, strings.TrimSpace(condition.Reason+" "+condition.Message))
			break
		}
	}
	return truncateText(strings.Join(parts, "；"), 220)
}

func namespaceConditions(conditions []corev1.NamespaceCondition) []response.SystemHealthObjectConditionEvidence {
	result := make([]response.SystemHealthObjectConditionEvidence, 0, len(conditions))
	for _, condition := range conditions {
		result = append(result, response.SystemHealthObjectConditionEvidence{
			Type:    string(condition.Type),
			Status:  string(condition.Status),
			Reason:  condition.Reason,
			Message: truncateText(condition.Message, 160),
		})
	}
	return result
}

func podConditions(conditions []corev1.PodCondition) []response.SystemHealthObjectConditionEvidence {
	result := make([]response.SystemHealthObjectConditionEvidence, 0, len(conditions))
	for _, condition := range conditions {
		result = append(result, response.SystemHealthObjectConditionEvidence{
			Type:    string(condition.Type),
			Status:  string(condition.Status),
			Reason:  condition.Reason,
			Message: truncateText(condition.Message, 160),
		})
	}
	return result
}

func containerStatuses(initContainers []corev1.ContainerStatus, containers []corev1.ContainerStatus, targetContainer string) []response.SystemHealthContainerStatusEvidence {
	all := append([]corev1.ContainerStatus{}, initContainers...)
	all = append(all, containers...)
	if len(all) == 0 {
		return nil
	}
	result := make([]response.SystemHealthContainerStatusEvidence, 0, len(all))
	for _, status := range all {
		if targetContainer != "" && status.Name != targetContainer && !strings.Contains(status.ContainerID, targetContainer) {
			continue
		}
		state, reason := containerState(status)
		result = append(result, response.SystemHealthContainerStatusEvidence{
			Name:         status.Name,
			Ready:        status.Ready,
			RestartCount: status.RestartCount,
			State:        state,
			Reason:       reason,
			Image:        status.Image,
		})
	}
	if len(result) == 0 && targetContainer != "" {
		return containerStatuses(nil, all, "")
	}
	return result
}

func containerState(status corev1.ContainerStatus) (string, string) {
	switch {
	case status.State.Running != nil:
		return "running", ""
	case status.State.Waiting != nil:
		return "waiting", status.State.Waiting.Reason
	case status.State.Terminated != nil:
		return "terminated", status.State.Terminated.Reason
	default:
		return "unknown", ""
	}
}

func objectAgeSeconds(createdAt time.Time) int64 {
	if createdAt.IsZero() {
		return 0
	}
	return int64(time.Since(createdAt).Seconds())
}

func selectedStatusLabels(labels map[string]string) map[string]string {
	if len(labels) == 0 {
		return nil
	}
	selected := make(map[string]string, 8)
	for _, key := range []string{
		"app", "app.kubernetes.io/name", "app.kubernetes.io/instance", "component",
		"pod-template-hash", "version", "release", "heritage",
	} {
		if value := strings.TrimSpace(labels[key]); value != "" {
			selected[key] = value
		}
	}
	if len(selected) == 0 {
		return nil
	}
	return selected
}

func buildRelatedObjectStatusResult(statuses []response.SystemHealthObjectStatusEvidence, warnings []string) string {
	parts := make([]string, 0, len(statuses)+len(warnings))
	for _, status := range statuses {
		switch status.Type {
		case "pod":
			ready := "unknown"
			if status.Ready != nil {
				ready = strconv.FormatBool(*status.Ready)
			}
			parts = append(parts, fmt.Sprintf("Pod %s/%s phase=%s ready=%s restarts=%d node=%s", status.Namespace, status.Name, status.Status, ready, status.Restarts, emptyAsAny(status.NodeName)))
		case "namespace":
			parts = append(parts, fmt.Sprintf("Namespace %s phase=%s", status.Name, status.Status))
		default:
			parts = append(parts, fmt.Sprintf("%s %s status=%s", status.Type, status.Name, status.Status))
		}
	}
	if len(warnings) > 0 {
		parts = append(parts, "查询边界："+strings.Join(warnings, "；"))
	}
	return truncateText(strings.Join(parts, "；"), 520)
}

func alertDirections(events []alertmodel.AEventWithWRecord) []string {
	values := make([]string, 0, len(events))
	for _, event := range events {
		values = append(values, event.AlertDirection)
	}
	return uniqueStrings(values)
}

func displaySource(event alertmodel.AEventWithWRecord) string {
	if event.Source != "" {
		return event.Source
	}
	if event.Group != "" {
		return event.Group
	}
	return "实时告警"
}

func (h *handler) buildRealEvidence(ctx core.Context, input reportInput, groupID int64, startTime time.Time, endTime time.Time, capabilities telemetryCapabilityStatus) []response.SystemHealthEvidence {
	evidence := make([]response.SystemHealthEvidence, 0, 4)
	evidence = append(evidence, response.SystemHealthEvidence{
		Type:        "告警事件",
		Source:      sourceLabel("alerts"),
		Severity:    maxSeverity(input.alertSeverities),
		Query:       "alert_event",
		Result:      buildAlertEventEvidenceResult(input),
		AlertEvents: buildAlertEventEvidenceSamples(input),
	})

	if trendEvidence := h.alertRuleMetricTrendEvidence(ctx, input, groupID, startTime, endTime); trendEvidence.Result != "" {
		evidence = append(evidence, trendEvidence)
	}
	if objectStatusEvidence := h.relatedObjectStatusEvidence(ctx, input); objectStatusEvidence.Result != "" {
		evidence = append(evidence, objectStatusEvidence)
	}
	if traceEvidence := h.traceSampleEvidence(ctx, input, groupID, startTime, endTime); traceEvidence.Result != "" {
		evidence = append(evidence, traceEvidence)
	}

	detailSamples := make([]string, 0, minInt(maxEvidenceSamples, len(input.events)))
	for _, event := range input.events {
		result := alertDetailSummary(event)
		if result == "" {
			result = fmt.Sprintf("告警级别 %s，状态 %s，来源 %s", event.Severity, event.Status, displaySource(event))
		}
		if len(detailSamples) < maxEvidenceSamples {
			detailSamples = append(detailSamples, truncateText(result, 90))
		}
	}
	if len(detailSamples) > 0 {
		detailResult := strings.Join(detailSamples, "；")
		if len(input.events) > 1 {
			detailResult = fmt.Sprintf("%d 条同服务/同对象告警明细已合并，最高级别 %s；代表明细：%s", len(input.events), maxSeverity(input.alertSeverities), detailResult)
		}
		evidence = append(evidence, response.SystemHealthEvidence{
			Type:     "告警明细",
			Source:   sourceLabel("alerts"),
			Severity: maxSeverity(input.alertSeverities),
			Query:    "alert_event.detail",
			Result:   truncateText(detailResult, 240),
		})
	}

	missing := missingSourceLabels(input.availableCalcSources)
	if len(missing) > 0 {
		evidence = append(evidence, response.SystemHealthEvidence{
			Type:   "数据边界",
			Source: "system-health",
			Query:  "collector_status",
			Result: "尚未接入：" + strings.Join(missing, "、") + "。本结论未使用缺失数据做推断。",
		})
	}
	if capabilityEvidence := buildTelemetryCapabilityEvidence(capabilities); capabilityEvidence.Result != "" {
		evidence = append(evidence, capabilityEvidence)
	}
	return evidence
}

func (h *handler) traceSampleEvidence(ctx core.Context, input reportInput, groupID int64, startTime time.Time, endTime time.Time) response.SystemHealthEvidence {
	evidence := response.SystemHealthEvidence{
		Type:     "Trace链路样本",
		Source:   sourceLabel("trace"),
		Severity: maxSeverity(input.alertSeverities),
		Query:    "span_trace.labels['service_name'] + labels['content_key']",
	}
	if h.chRepo == nil || len(input.events) == 0 {
		return evidence
	}

	for _, event := range input.events[:minInt(maxEvidenceSamples, len(input.events))] {
		rawTags := rawTagsFromEvent(event)
		serviceName := firstNonEmpty(
			rawTags["service_name"],
			rawTags["serviceName"],
			rawTags["svc_name"],
			event.GetServiceNameTag(),
		)
		contentKey := firstNonEmpty(rawTags["content_key"], rawTags["endpoint"])
		if serviceName == "" {
			continue
		}

		req := &request.GetTracePageListRequest{
			GroupID:   groupID,
			StartTime: startTime.UnixMicro(),
			EndTime:   endTime.UnixMicro(),
			Service:   []string{serviceName},
			EndPoint:  contentKey,
			PageNum:   1,
			PageSize:  20,
		}
		if clusterID := firstNonEmpty(rawTags["cluster_id"], rawTags["clusterId"]); clusterID != "" {
			req.ClusterIDs = []string{clusterID}
		}

		traces, total, err := h.chRepo.GetTracePageList(ctx, req)
		if err != nil || total == 0 {
			continue
		}
		evidence.Result = truncateText(fmt.Sprintf(
			"按告警 raw_tags 反查 span_trace：service=%s，endpoint=%s；当前窗口匹配 span 样本=%d；%s。该证据只说明同窗口链路样本存在，不作为独立根因结论。",
			serviceName,
			emptyAsAny(contentKey),
			total,
			summarizeTraceSamples(traces),
		), 460)
		return evidence
	}
	return evidence
}

func summarizeTraceSamples(traces []clickhouse.QueryTraceResult) string {
	if len(traces) == 0 {
		return "未返回可展示样本"
	}
	errorCount := 0
	slowCount := 0
	var maxDuration uint64
	var sample clickhouse.QueryTraceResult
	for _, trace := range traces {
		if trace.IsError {
			errorCount++
		}
		if trace.IsSlow {
			slowCount++
		}
		if trace.Duration >= maxDuration {
			maxDuration = trace.Duration
			sample = trace
		}
	}
	sampleID := firstNonEmpty(sample.TraceId, "无 traceId")
	return fmt.Sprintf("本页%d条样本中 error=%d、slow=%d、最大耗时=%.2fms、代表trace=%s、span=%s",
		len(traces),
		errorCount,
		slowCount,
		float64(maxDuration)/1000,
		sampleID,
		emptyAsAny(sample.SpanId),
	)
}

func emptyAsAny(value string) string {
	if strings.TrimSpace(value) == "" {
		return "任意"
	}
	return strings.TrimSpace(value)
}

func (h *handler) queryTelemetryCapabilities(ctx core.Context) telemetryCapabilityStatus {
	status := telemetryCapabilityStatus{}
	if h.dbRepo == nil {
		status.Err = "database repository is not configured"
		return status
	}
	dataPlanes, err := h.dbRepo.ListDataPlane(ctx)
	if err != nil {
		status.Err = err.Error()
		return status
	}

	enabledSources := make(map[integration.TelemetryType][]string)
	for _, dataPlane := range dataPlanes {
		source := dataPlane.Name
		if dataPlane.TypeDisplayName != "" {
			source = fmt.Sprintf("%s(%s)", dataPlane.Name, dataPlane.TypeDisplayName)
		} else if dataPlane.Typ != "" {
			source = fmt.Sprintf("%s(%s)", dataPlane.Name, dataPlane.Typ)
		}
		for _, capability := range dataPlane.Capability.Obj {
			telemetryType := integration.TelemetryType(capability)
			enabledSources[telemetryType] = append(enabledSources[telemetryType], source)
		}
	}

	for _, telemetryType := range integration.AllTelemetryTypes {
		label := telemetryTypeLabel(telemetryType)
		sources := uniqueStrings(enabledSources[telemetryType])
		if len(sources) == 0 {
			status.Disabled = append(status.Disabled, label)
			continue
		}
		status.Enabled = append(status.Enabled, label)
		status.Sources = append(status.Sources, fmt.Sprintf("%s=%s", label, strings.Join(sources, ",")))
	}
	return status
}

func buildTelemetryCapabilityEvidence(status telemetryCapabilityStatus) response.SystemHealthEvidence {
	evidence := response.SystemHealthEvidence{
		Type:   "接入能力边界",
		Source: "dataplane",
		Query:  "data_plane.capability",
	}
	if status.Err != "" {
		evidence.Result = "接入能力查询失败：" + truncateText(status.Err, 180) + "。本报告不使用未确认能力做推断。"
		return evidence
	}
	parts := make([]string, 0, 3)
	if len(status.Enabled) > 0 {
		parts = append(parts, "已接入能力："+strings.Join(status.Enabled, "、"))
	} else {
		parts = append(parts, "未发现已接入能力")
	}
	if len(status.Disabled) > 0 {
		parts = append(parts, "未接入能力："+strings.Join(status.Disabled, "、"))
	}
	if len(status.Sources) > 0 {
		parts = append(parts, "来源："+strings.Join(status.Sources, "；"))
	}
	parts = append(parts, "已接入能力仅表示当前可查询，不代表本报告已将其用于根因推断")
	evidence.Result = truncateText(strings.Join(parts, "。"), 420)
	return evidence
}

func telemetryTypeLabel(telemetryType integration.TelemetryType) string {
	switch telemetryType {
	case integration.TelemetryMetrics:
		return "基础设施指标"
	case integration.TelemetryLogs:
		return "日志语义"
	case integration.TelemetryTraces:
		return "Trace链路"
	case integration.TelemetryRedMetric:
		return "APM/RED指标"
	case integration.TelemetryTopology:
		return "服务调用关系"
	case integration.TelemetryServiceInstance:
		return "服务实例"
	default:
		return string(telemetryType)
	}
}

func (h *handler) alertRuleMetricTrendEvidence(ctx core.Context, input reportInput, groupID int64, startTime time.Time, endTime time.Time) response.SystemHealthEvidence {
	evidence := response.SystemHealthEvidence{
		Type:     "告警指标趋势",
		Source:   "真实告警事件/Prometheus",
		Severity: maxSeverity(input.alertSeverities),
		Query:    "alert_event.raw_tags + alert_rule.expr",
	}
	if len(input.events) == 0 {
		return evidence
	}
	trend, err := h.queryAlertMetricTrend(ctx, input.events[:minInt(maxEvidenceSamples, len(input.events))], groupID, startTime, endTime)
	if err != nil {
		evidence.Result = err.Error()
		return evidence
	}
	evidence.Query = trend.Expression
	evidence.Metric = buildMetricEvidence(trend)
	if trend.EventScoped {
		summary := summarizeMetricTrend(trend.Stats)
		evidence.Result = truncateText(fmt.Sprintf(
			"告警=%s，通过 alertName + 告警 labels 定位当前触发 series；规则=%s/%s；匹配当前 series=%d/%d；匹配标签=%s；%s",
			trend.Rule.Alert,
			trend.Rule.Group,
			trend.Rule.Alert,
			trend.Stats.SeriesCount,
			trend.SeriesReturned,
			selectedAlertLabels(trend.MatchLabels),
			summary,
		), 520)
		return evidence
	}
	evidence.Result = truncateText(fmt.Sprintf(
		"已匹配告警规则 %s/%s；规则查询返回 %d 条 series，但当前窗口内没有命中告警事件对应的 series，未展示规则整体聚合值。",
		trend.Rule.Group,
		trend.Rule.Alert,
		trend.SeriesReturned,
	), 420)
	return evidence
}

func (h *handler) queryAlertMetricTrend(ctx core.Context, events []alertmodel.AEventWithWRecord, groupID int64, startTime time.Time, endTime time.Time) (alertMetricTrendQueryResult, error) {
	if len(events) == 0 {
		return alertMetricTrendQueryResult{}, fmt.Errorf("未提供告警事件，无法反查 PQL 和指标趋势。")
	}
	event := events[0]
	rule, found := h.findAlertRuleForEvent(ctx, event, groupID)
	if !found || rule == nil {
		return alertMetricTrendQueryResult{}, fmt.Errorf("未找到告警规则：告警事件=%s，无法反查 PQL 和指标趋势。", event.Name)
	}
	expr := strings.TrimSpace(rule.Expr)
	if expr == "" {
		return alertMetricTrendQueryResult{}, fmt.Errorf("已匹配告警规则 %s/%s，但规则未配置 PQL。", rule.Group, rule.Alert)
	}
	if h.promRepo == nil {
		return alertMetricTrendQueryResult{}, fmt.Errorf("已匹配告警规则 %s/%s；Prometheus 仓库未初始化，无法查询窗口趋势。", rule.Group, rule.Alert)
	}

	step := trendQueryStep(startTime, endTime)
	results, err := h.queryRawMetricSeries(ctx, startTime, endTime, expr, step)
	if err != nil {
		return alertMetricTrendQueryResult{}, fmt.Errorf("已匹配告警规则 %s/%s；查询窗口趋势失败：%w", rule.Group, rule.Alert, err)
	}
	for _, candidate := range events {
		matched := filterSeriesByAlertEvent(results, candidate)
		stats, ok := calculateMetricTrendStats(matched)
		if !ok {
			continue
		}
		return alertMetricTrendQueryResult{
			Rule:           rule,
			Expression:     expr,
			Step:           step,
			SeriesReturned: len(results),
			MatchedSeries:  matched,
			Stats:          stats,
			MatchLabels:    selectedAlertLabelMap(rawTagsFromEvent(candidate), 8),
			EventScoped:    true,
			Message:        "已按告警事件标签命中对应指标序列。",
		}, nil
	}
	stats, ok := calculateMetricTrendStats(results)
	if !ok {
		return alertMetricTrendQueryResult{}, fmt.Errorf("已匹配告警规则 %s/%s；规则查询未返回可用时序点。", rule.Group, rule.Alert)
	}
	return alertMetricTrendQueryResult{
		Rule:           rule,
		Expression:     expr,
		Step:           step,
		SeriesReturned: len(results),
		Stats:          stats,
		EventScoped:    false,
		Message:        "规则查询只返回全局或不可匹配的聚合序列，未能用当前告警事件标签定位到指定 series。",
	}, nil
}

func buildMetricEvidence(trend alertMetricTrendQueryResult) *response.SystemHealthMetricEvidence {
	series, pointsTruncated := buildMetricSeriesEvidence(trend.MatchedSeries, maxMetricEvidenceSeries, maxMetricEvidencePoints)
	metric := &response.SystemHealthMetricEvidence{
		RuleGroup:       trend.Rule.Group,
		RuleName:        trend.Rule.Alert,
		Expression:      trend.Expression,
		Step:            trend.Step.String(),
		SeriesReturned:  trend.SeriesReturned,
		SeriesMatched:   trend.Stats.SeriesCount,
		PointCount:      trend.Stats.PointCount,
		FirstValue:      trend.Stats.FirstValue,
		LastValue:       trend.Stats.LastValue,
		MinValue:        trend.Stats.MinValue,
		MaxValue:        trend.Stats.MaxValue,
		Delta:           trend.Stats.Delta,
		ChangePercent:   trend.Stats.ChangePercent,
		MatchedLabels:   trend.MatchLabels,
		MatchExpression: metricMatchExpression(trend.MatchLabels),
		EventScoped:     trend.EventScoped,
		Message:         trend.Message,
		Series:          series,
		PointLimit:      maxMetricEvidencePoints,
		PointsTruncated: pointsTruncated,
	}
	return metric
}

func buildMetricSeriesEvidence(series []alertMetricSeries, seriesLimit int, pointLimit int) ([]response.SystemHealthMetricSeriesEvidence, bool) {
	if len(series) == 0 || seriesLimit <= 0 || pointLimit <= 0 {
		return nil, false
	}
	truncated := len(series) > seriesLimit
	limit := minInt(len(series), seriesLimit)
	items := make([]response.SystemHealthMetricSeriesEvidence, 0, limit)
	for _, item := range series[:limit] {
		points, pointsTruncated := downsampleMetricPoints(item.Values, pointLimit)
		truncated = truncated || pointsTruncated
		items = append(items, response.SystemHealthMetricSeriesEvidence{
			Labels: copyStringMap(item.Labels),
			Points: points,
		})
	}
	return items, truncated
}

func downsampleMetricPoints(points []prometheus.Points, limit int) ([]response.SystemHealthMetricDataPoint, bool) {
	validPoints := make([]prometheus.Points, 0, len(points))
	for _, point := range points {
		if math.IsNaN(point.Value) || math.IsInf(point.Value, 0) {
			continue
		}
		validPoints = append(validPoints, point)
	}
	if len(validPoints) == 0 || limit <= 0 {
		return nil, false
	}
	if len(validPoints) <= limit {
		return metricDataPointsFromPromPoints(validPoints), false
	}
	if limit == 1 {
		return metricDataPointsFromPromPoints(validPoints[:1]), true
	}

	sampled := make([]prometheus.Points, 0, limit)
	lastIndex := len(validPoints) - 1
	for i := 0; i < limit; i++ {
		index := int(math.Round(float64(i) * float64(lastIndex) / float64(limit-1)))
		if len(sampled) > 0 && sampled[len(sampled)-1].TimeStamp == validPoints[index].TimeStamp {
			continue
		}
		sampled = append(sampled, validPoints[index])
	}
	return metricDataPointsFromPromPoints(sampled), true
}

func metricDataPointsFromPromPoints(points []prometheus.Points) []response.SystemHealthMetricDataPoint {
	dataPoints := make([]response.SystemHealthMetricDataPoint, 0, len(points))
	for _, point := range points {
		dataPoints = append(dataPoints, response.SystemHealthMetricDataPoint{
			Timestamp: point.TimeStamp,
			Value:     point.Value,
		})
	}
	return dataPoints
}

func copyStringMap(source map[string]string) map[string]string {
	if len(source) == 0 {
		return nil
	}
	copied := make(map[string]string, len(source))
	for key, value := range source {
		copied[key] = value
	}
	return copied
}

func metricMatchExpression(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	keys := orderedSelectedAlertLabelKeys(labels, labels)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		escapedValue := strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(labels[key])
		parts = append(parts, fmt.Sprintf(`%s="%s"`, key, escapedValue))
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func (h *handler) findAlertRuleForEvent(ctx core.Context, event alertmodel.AEventWithWRecord, groupID int64) (*request.AlertRule, bool) {
	if rule, found := h.findVMAlertRuleForEvent(ctx, event); found {
		return rule, true
	}
	if h.k8sRepo == nil || strings.TrimSpace(event.Name) == "" {
		return nil, false
	}
	groupIDs := []int64{groupID}
	if rawGroupID := event.GetStringTagWithRaw("groupId"); rawGroupID != "" {
		if parsed, err := strconv.ParseInt(rawGroupID, 10, 64); err == nil {
			groupIDs = append(groupIDs, parsed)
		}
	}
	groupIDs = append(groupIDs, 0)

	for index, id := range uniqueInt64(groupIDs) {
		rule, found := h.findAlertRuleByName(event, id, index == 0)
		if found {
			return rule, true
		}
	}
	return nil, false
}

func (h *handler) findVMAlertRuleForEvent(ctx core.Context, event alertmodel.AEventWithWRecord) (*request.AlertRule, bool) {
	if strings.TrimSpace(h.vmAlertURL) == "" || strings.TrimSpace(event.Name) == "" {
		return nil, false
	}
	if rule, found := h.findVMAlertActiveRuleForEvent(ctx, event); found {
		return rule, true
	}
	return h.findVMAlertRuleDefinitionForEvent(ctx, event)
}

func (h *handler) findVMAlertActiveRuleForEvent(ctx core.Context, event alertmodel.AEventWithWRecord) (*request.AlertRule, bool) {
	endpoint := strings.TrimRight(h.vmAlertURL, "/") + "/api/v1/alerts"
	req, err := http.NewRequestWithContext(ctx.GetContext(), http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, false
	}
	resp, err := (&http.Client{Timeout: 8 * time.Second}).Do(req)
	if err != nil {
		return nil, false
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, false
	}

	var apiResp vmAlertAPIAlertsResp
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, false
	}
	for _, alert := range apiResp.Data.Alerts {
		if strings.TrimSpace(alert.Name) != strings.TrimSpace(event.Name) {
			continue
		}
		if alert.State != "" && alert.State != alertmodel.StatusFiring {
			continue
		}
		if !labelsMatchAlertEvent(alert.Labels, event) {
			continue
		}
		expr := strings.TrimSpace(firstNonEmpty(alert.Expression, alert.Query))
		if expr == "" {
			continue
		}
		return &request.AlertRule{
			Group: firstNonEmpty(alert.Labels["alertgroup"], alert.Labels["group"], alert.GroupID),
			Alert: event.Name,
			Expr:  expr,
			Labels: map[string]string{
				"group_id": alert.GroupID,
				"rule_id":  alert.ID,
				"source":   "vmalert-api",
			},
		}, true
	}
	return nil, false
}

func (h *handler) findVMAlertRuleDefinitionForEvent(ctx core.Context, event alertmodel.AEventWithWRecord) (*request.AlertRule, bool) {
	endpoint := strings.TrimRight(h.vmAlertURL, "/") + "/api/v1/rules"
	req, err := http.NewRequestWithContext(ctx.GetContext(), http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, false
	}
	resp, err := (&http.Client{Timeout: 8 * time.Second}).Do(req)
	if err != nil {
		return nil, false
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, false
	}

	var apiResp vmAlertAPIRulesResp
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, false
	}

	rawTags := rawTagsFromEvent(event)
	var fallback *request.AlertRule
	for _, group := range apiResp.Data.Groups {
		for _, rule := range group.Rules {
			if strings.TrimSpace(rule.Name) != strings.TrimSpace(event.Name) {
				continue
			}
			expr := strings.TrimSpace(firstNonEmpty(rule.Query, rule.Expr))
			if expr == "" {
				continue
			}
			candidate := &request.AlertRule{
				Group: firstNonEmpty(rawTags["alertgroup"], rawTags["groupId"], rule.GroupName, group.Name, rule.GroupID, group.ID),
				Alert: event.Name,
				Expr:  expr,
				Labels: map[string]string{
					"group_id": firstNonEmpty(rule.GroupID, group.ID),
					"rule_id":  rule.ID,
					"source":   "vmalert-rules-api",
				},
			}
			if fallback == nil {
				fallback = candidate
			}
			if vmAlertRuleMatchesEvent(rule, group, event) {
				return candidate, true
			}
		}
	}
	if fallback != nil {
		return fallback, true
	}
	return nil, false
}

func vmAlertRuleMatchesEvent(rule vmAlertAPIRule, group vmAlertAPIGroup, event alertmodel.AEventWithWRecord) bool {
	rawTags := rawTagsFromEvent(event)
	candidates := []map[string]string{rule.Labels}
	for _, alert := range rule.Alerts {
		candidates = append(candidates, alert.Labels)
	}
	for _, labels := range candidates {
		if len(labels) == 0 {
			continue
		}
		if labelsMatchAlertEvent(labels, event) {
			return true
		}
	}

	rawGroup := firstNonEmpty(rawTags["alertgroup"], rawTags["groupId"])
	if rawGroup == "" {
		return true
	}
	for _, candidate := range []string{rule.GroupName, rule.GroupID, group.Name, group.ID, rule.Labels["groupId"]} {
		if strings.TrimSpace(candidate) == strings.TrimSpace(rawGroup) {
			return true
		}
	}
	return false
}

func (h *handler) findAlertRuleByName(event alertmodel.AEventWithWRecord, groupID int64, syncNow bool) (*request.AlertRule, bool) {
	filter := &request.AlertRuleFilter{Alert: event.Name}
	if groupID > 0 {
		filter.GroupID = groupID
	}
	rules, _ := h.k8sRepo.GetAlertRules("", filter, nil, syncNow)
	if len(rules) == 0 && !syncNow {
		rules, _ = h.k8sRepo.GetAlertRules("", filter, nil, true)
	}
	return selectAlertRule(event, rules)
}

func selectAlertRule(event alertmodel.AEventWithWRecord, rules []*request.AlertRule) (*request.AlertRule, bool) {
	groupCandidates := uniqueStrings([]string{
		event.GetStringTagWithRaw("alertgroup"),
		event.Group,
	})
	var fallback *request.AlertRule
	for _, rule := range rules {
		if rule == nil || rule.Alert != event.Name {
			continue
		}
		if fallback == nil {
			fallback = rule
		}
		for _, group := range groupCandidates {
			if group != "" && rule.Group == group {
				return rule, true
			}
		}
	}
	if fallback != nil {
		return fallback, true
	}
	return nil, false
}

func trendQueryStep(startTime time.Time, endTime time.Time) time.Duration {
	window := endTime.Sub(startTime)
	if window <= 0 {
		return time.Minute
	}
	step := window / 30
	if step < 15*time.Second {
		return 15 * time.Second
	}
	if step > 5*time.Minute {
		return 5 * time.Minute
	}
	return step
}

func (h *handler) queryRawMetricSeries(ctx core.Context, startTime time.Time, endTime time.Time, expr string, step time.Duration) ([]alertMetricSeries, error) {
	value, _, err := h.promRepo.GetApi().QueryRange(ctx.GetContext(), expr, v1.Range{
		Start: startTime,
		End:   endTime,
		Step:  step,
	})
	if err != nil {
		return nil, err
	}
	matrix, ok := value.(prommodel.Matrix)
	if !ok {
		return nil, fmt.Errorf("unexpected type %T, expected prometheus matrix", value)
	}
	series := make([]alertMetricSeries, 0, len(matrix))
	for _, stream := range matrix {
		labels := make(map[string]string, len(stream.Metric))
		for key, value := range stream.Metric {
			labels[string(key)] = string(value)
		}
		points := make([]prometheus.Points, 0, len(stream.Values))
		for _, point := range stream.Values {
			points = append(points, prometheus.Points{
				TimeStamp: point.Timestamp.UnixNano() / 1e3,
				Value:     float64(point.Value),
			})
		}
		series = append(series, alertMetricSeries{Labels: labels, Values: points})
	}
	return series, nil
}

func filterSeriesByAlertEvent(results []alertMetricSeries, event alertmodel.AEventWithWRecord) []alertMetricSeries {
	selectors := alertEventSeriesSelectors(event)
	if len(selectors) == 0 {
		return nil
	}
	matched := make([]alertMetricSeries, 0, len(results))
	for _, result := range results {
		if len(result.Labels) == 0 {
			continue
		}
		if labelsMatchSelectors(result.Labels, selectors) {
			matched = append(matched, result)
		}
	}
	return matched
}

func labelsMatchAlertEvent(labels map[string]string, event alertmodel.AEventWithWRecord) bool {
	return labelsMatchSelectors(labels, alertEventSeriesSelectors(event))
}

type alertSeriesSelector struct {
	Keys  []string
	Value string
}

func alertEventSeriesSelectors(event alertmodel.AEventWithWRecord) []alertSeriesSelector {
	rawTags := rawTagsFromEvent(event)
	selectors := make([]alertSeriesSelector, 0, 8)
	add := func(keys []string, value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		for _, selector := range selectors {
			if selector.Value == value && stringSlicesOverlap(selector.Keys, keys) {
				return
			}
		}
		selectors = append(selectors, alertSeriesSelector{Keys: keys, Value: value})
	}

	add([]string{"svc_name", "serviceName", "service_name", "service"}, alertServiceName(event))
	add([]string{"content_key", "endpoint", "uri", "route"}, alertEndpoint(event))
	for _, key := range []string{
		"cluster_id", "namespace", "pod", "pod_name", "k8s_pod", "container", "container_name",
		"pid", "comm", "instance", "node", "node_name", "nodename", "device", "mountpoint",
		"database", "db", "db_url", "job",
	} {
		if value := rawTags[key]; value != "" {
			add([]string{key}, value)
		}
	}
	return selectors
}

func labelsMatchSelectors(labels map[string]string, selectors []alertSeriesSelector) bool {
	matchedLabels := 0
	for _, selector := range selectors {
		seenComparableLabel := false
		selectorMatched := false
		for _, key := range selector.Keys {
			labelValue := strings.TrimSpace(labels[key])
			if labelValue == "" {
				continue
			}
			seenComparableLabel = true
			if labelValue == selector.Value {
				selectorMatched = true
				break
			}
		}
		if selectorMatched {
			matchedLabels++
			continue
		}
		if seenComparableLabel {
			return false
		}
	}
	return matchedLabels > 0
}

func stringSlicesOverlap(left []string, right []string) bool {
	for _, leftValue := range left {
		for _, rightValue := range right {
			if leftValue == rightValue {
				return true
			}
		}
	}
	return false
}

func rawTagsFromEvent(event alertmodel.AEventWithWRecord) map[string]string {
	rawTags := make(map[string]string, len(event.Tags)+len(event.EnrichTags))
	for key, value := range event.Tags {
		switch typed := value.(type) {
		case string:
			rawTags[key] = typed
		case fmt.Stringer:
			rawTags[key] = typed.String()
		default:
			rawTags[key] = fmt.Sprint(typed)
		}
	}
	for key, value := range event.EnrichTags {
		if _, exists := rawTags[key]; !exists {
			rawTags[key] = value
		}
	}
	return rawTags
}

func calculateMetricTrendStats(results []alertMetricSeries) (alertMetricTrendStats, bool) {
	aggregated := make(map[int64]float64)
	seriesCount := 0
	pointCount := 0
	for _, result := range results {
		if len(result.Values) == 0 {
			continue
		}
		seriesHasValidPoint := false
		for _, point := range result.Values {
			if math.IsNaN(point.Value) || math.IsInf(point.Value, 0) {
				continue
			}
			aggregated[point.TimeStamp] += point.Value
			pointCount++
			seriesHasValidPoint = true
		}
		if seriesHasValidPoint {
			seriesCount++
		}
	}
	if len(aggregated) == 0 {
		return alertMetricTrendStats{}, false
	}
	timestamps := make([]int64, 0, len(aggregated))
	for timestamp := range aggregated {
		timestamps = append(timestamps, timestamp)
	}
	sort.Slice(timestamps, func(i, j int) bool { return timestamps[i] < timestamps[j] })

	first := aggregated[timestamps[0]]
	last := aggregated[timestamps[len(timestamps)-1]]
	minValue := first
	maxValue := first
	for _, timestamp := range timestamps {
		value := aggregated[timestamp]
		minValue = math.Min(minValue, value)
		maxValue = math.Max(maxValue, value)
	}
	delta := last - first
	var changePercent *float64
	if first != 0 {
		change := delta / first * 100
		changePercent = &change
	}
	return alertMetricTrendStats{
		SeriesCount:   seriesCount,
		PointCount:    pointCount,
		FirstValue:    first,
		LastValue:     last,
		MinValue:      minValue,
		MaxValue:      maxValue,
		Delta:         delta,
		ChangePercent: changePercent,
	}, true
}

func summarizeMetricTrend(stats alertMetricTrendStats) string {
	change := "无法计算百分比"
	if stats.ChangePercent != nil {
		change = fmt.Sprintf("%s%%", formatMetricFloat(*stats.ChangePercent))
	}
	return fmt.Sprintf("窗口趋势：序列=%d，采样点=%d，聚合方式=按时间戳求和，首值=%s，末值=%s，最小=%s，最大=%s，变化=%s（%s）",
		stats.SeriesCount,
		stats.PointCount,
		formatMetricFloat(stats.FirstValue),
		formatMetricFloat(stats.LastValue),
		formatMetricFloat(stats.MinValue),
		formatMetricFloat(stats.MaxValue),
		formatSignedMetricFloat(stats.Delta),
		change,
	)
}

func alertDetailDescription(event alertmodel.AEventWithWRecord) string {
	detail := strings.TrimSpace(event.Detail)
	if detail == "" {
		return ""
	}
	fields := make(map[string]any)
	if err := json.Unmarshal([]byte(detail), &fields); err != nil {
		return detail
	}
	for _, key := range []string{"description", "summary", "message"} {
		if value, ok := fields[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return detail
}

func alertMetricSummary(event alertmodel.AEventWithWRecord) string {
	description := alertDetailDescription(event)
	parts := make([]string, 0, 6)
	if event.Name != "" {
		parts = append(parts, "告警规则="+event.Name)
	}
	if object := eventScope(event); object != "" {
		parts = append(parts, "对象="+object)
	}
	if endpoint := alertEndpoint(event); endpoint != "" {
		parts = append(parts, "接口="+endpoint)
	}
	if value := metricValueFromDescription(description); value != "" {
		parts = append(parts, "触发值="+value)
	}
	if labels := selectedAlertLabels(labelsFromDescription(description)); labels != "" {
		parts = append(parts, "关键标签="+labels)
	}
	if event.Severity != "" {
		parts = append(parts, "级别="+event.Severity)
	}
	if event.Status != "" {
		parts = append(parts, "状态="+event.Status)
	}
	return strings.Join(parts, "，")
}

func alertDetailSummary(event alertmodel.AEventWithWRecord) string {
	description := alertDetailDescription(event)
	if description == "" {
		return ""
	}
	parts := make([]string, 0, 3)
	if message := metricMessageFromDescription(description); message != "" {
		parts = append(parts, message)
	}
	if value := metricValueFromDescription(description); value != "" {
		parts = append(parts, "触发值="+value)
	}
	if labels := selectedAlertLabels(labelsFromDescription(description)); labels != "" {
		parts = append(parts, "关键标签="+labels)
	}
	if len(parts) > 0 {
		return strings.Join(parts, "，")
	}
	return description
}

func alertEndpoint(event alertmodel.AEventWithWRecord) string {
	candidates := []string{
		event.GetEndpointTag(),
		event.GetStringTagWithRaw("content_key"),
		event.GetStringTagWithRaw("endpoint"),
	}
	for _, candidate := range candidates {
		if strings.TrimSpace(candidate) != "" {
			return strings.TrimSpace(candidate)
		}
	}
	return ""
}

func metricMessageFromDescription(description string) string {
	description = strings.TrimSpace(description)
	if description == "" {
		return ""
	}
	if index := strings.Index(description, "VALUE ="); index >= 0 {
		return strings.TrimSpace(description[:index])
	}
	if index := strings.Index(description, "LABELS ="); index >= 0 {
		return strings.TrimSpace(description[:index])
	}
	return description
}

func metricValueFromDescription(description string) string {
	const marker = "VALUE ="
	index := strings.Index(description, marker)
	if index < 0 {
		return ""
	}
	value := strings.TrimSpace(description[index+len(marker):])
	if lineEnd := strings.IndexAny(value, "\r\n"); lineEnd >= 0 {
		value = strings.TrimSpace(value[:lineEnd])
	}
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

func labelsFromDescription(description string) map[string]string {
	const marker = "LABELS ="
	index := strings.Index(description, marker)
	if index < 0 {
		return nil
	}
	rawLabels := strings.TrimSpace(description[index+len(marker):])
	if rawLabels == "" {
		return nil
	}
	if lineEnd := strings.IndexAny(rawLabels, "\r\n"); lineEnd >= 0 {
		rawLabels = strings.TrimSpace(rawLabels[:lineEnd])
	}
	rawLabels = strings.Trim(rawLabels, "`")
	if strings.HasPrefix(rawLabels, "map[") && strings.HasSuffix(rawLabels, "]") {
		return parsePrometheusMapLabels(strings.TrimSuffix(strings.TrimPrefix(rawLabels, "map["), "]"))
	}
	if strings.HasPrefix(rawLabels, "{") {
		fields := make(map[string]any)
		if err := json.Unmarshal([]byte(rawLabels), &fields); err == nil {
			labels := make(map[string]string, len(fields))
			for key, value := range fields {
				labels[key] = fmt.Sprint(value)
			}
			return labels
		}
	}
	return nil
}

func parsePrometheusMapLabels(rawLabels string) map[string]string {
	labels := make(map[string]string)
	for _, pair := range strings.Fields(rawLabels) {
		key, value, ok := strings.Cut(pair, ":")
		if !ok || strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			continue
		}
		labels[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	if len(labels) == 0 {
		return nil
	}
	return labels
}

func selectedAlertLabels(labels map[string]string) string {
	selected := selectedAlertLabelMap(labels, 6)
	if len(selected) == 0 {
		return ""
	}
	keys := orderedSelectedAlertLabelKeys(labels, selected)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+selected[key])
	}
	return strings.Join(parts, "，")
}

func selectedAlertLabelMap(labels map[string]string, limit int) map[string]string {
	if len(labels) == 0 {
		return nil
	}
	if limit <= 0 {
		limit = 6
	}
	priorityKeys := []string{
		"content_key", "endpoint", "serviceName", "svc_name",
		"node", "nodename", "instance", "namespace", "pod", "pod_name",
		"container", "container_name", "pid", "comm", "job", "mountpoint",
		"device", "alertgroup", "alertname",
	}
	selected := make(map[string]string, limit)
	used := make(map[string]struct{})
	for _, key := range priorityKeys {
		if value := strings.TrimSpace(labels[key]); value != "" {
			selected[key] = value
			used[key] = struct{}{}
			if len(selected) >= limit {
				return selected
			}
		}
	}
	keys := make([]string, 0, len(labels))
	for key := range labels {
		if _, ok := used[key]; !ok {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	for _, key := range keys {
		if value := strings.TrimSpace(labels[key]); value != "" {
			selected[key] = value
			if len(selected) >= limit {
				break
			}
		}
	}
	if len(selected) == 0 {
		return nil
	}
	return selected
}

func orderedSelectedAlertLabelKeys(source map[string]string, selected map[string]string) []string {
	priorityKeys := []string{
		"content_key", "endpoint", "serviceName", "svc_name",
		"node", "nodename", "instance", "namespace", "pod", "pod_name",
		"container", "container_name", "pid", "comm", "job", "mountpoint",
		"device", "alertgroup", "alertname",
	}
	keys := make([]string, 0, len(selected))
	seen := make(map[string]struct{}, len(selected))
	for _, key := range priorityKeys {
		if _, ok := selected[key]; ok {
			keys = append(keys, key)
			seen[key] = struct{}{}
		}
	}
	remaining := make([]string, 0, len(selected))
	for key := range selected {
		if _, ok := seen[key]; ok {
			continue
		}
		remaining = append(remaining, key)
	}
	sort.SliceStable(remaining, func(i, j int) bool {
		left := source[remaining[i]]
		right := source[remaining[j]]
		if left != right {
			return left < right
		}
		return remaining[i] < remaining[j]
	})
	keys = append(keys, remaining...)
	return keys
}

func buildRealImpact(input reportInput) string {
	switch input.kind {
	case issueKindActiveAlert:
		return fmt.Sprintf("%d 个去重有效告警实例仍在触发，影响对象为 %s；最长持续 %s。", input.activeAlerts, input.scope, formatDurationCN(input.durationSeconds))
	case issueKindInspection:
		return fmt.Sprintf("真实巡检/异常检测事件仍在触发，影响对象为 %s。", input.scope)
	case issueKindRecoveredAlert:
		return fmt.Sprintf("%d 个去重有效告警实例已恢复，影响对象 %s 进入恢复观察。", input.deduplicatedAlerts, input.scope)
	case issueKindInvalidAlert:
		return fmt.Sprintf("%d 个告警实例仍在触发但被反复识别为无效，影响对象为 %s；优先判断告警指标、表达式或阈值是否与服务 RED 指标背离。", input.deduplicatedAlerts, input.scope)
	default:
		return fmt.Sprintf("影响对象为 %s。", input.scope)
	}
}

func scoreIssues(inputs []reportInput) []scoredIssue {
	scored := make([]scoredIssue, 0, len(inputs))
	for _, input := range inputs {
		entryImpact := entryImpactScore(input)
		alertSeverity := alertSeverityScore(input.alertSeverities, input.activeAlerts)
		if input.kind == issueKindInvalidAlert {
			alertSeverity = 0
		}
		redDegradation := redDegradationScore(input)
		propagation := propagationScore(input)
		completenessPenalty := dataCompletenessPenalty(input.availableCalcSources)
		pathRisk := 0.35*entryImpact + 0.25*alertSeverity + 0.20*redDegradation + 0.15*propagation + 0.05*completenessPenalty
		priority := priority(pathRisk, input.alertSeverities, input.kind)
		if priority == "" {
			continue
		}
		status := issueStatus(pathRisk)
		if input.kind == issueKindRecoveredAlert {
			status = "recovering"
		}
		if input.kind == issueKindInvalidAlert {
			status = "minor"
		}
		scored = append(scored, scoredIssue{
			input:                   input,
			entryImpactScore:        entryImpact,
			alertSeverityScore:      alertSeverity,
			redDegradationScore:     redDegradation,
			propagationScore:        propagation,
			dataCompletenessPenalty: completenessPenalty,
			pathRisk:                pathRisk,
			dataQualityScore:        issueDataQualityScore(input),
			priority:                priority,
			status:                  status,
		})
	}

	sort.Slice(scored, func(i, j int) bool {
		if scored[i].priority != scored[j].priority {
			return priorityRank(scored[i].priority) < priorityRank(scored[j].priority)
		}
		if scored[i].pathRisk != scored[j].pathRisk {
			return scored[i].pathRisk > scored[j].pathRisk
		}
		return scored[i].dataQualityScore > scored[j].dataQualityScore
	})

	for i := range scored {
		scored[i].scoreImpact = -int(math.Round(scored[i].pathRisk * issueWeightInSystemRisk(scored, i)))
	}
	return scored
}

func buildIssues(scored []scoredIssue) []response.SystemHealthIssue {
	issues := make([]response.SystemHealthIssue, 0, len(scored))
	for _, item := range scored {
		actionPlan := buildActionPlan(item)
		issues = append(issues, response.SystemHealthIssue{
			Key:                   item.input.key,
			AggregationKey:        item.input.aggregationKey,
			Priority:              item.priority,
			Title:                 item.input.title,
			Status:                item.status,
			Scope:                 item.input.scope,
			AffectedObjects:       item.input.affectedObjects,
			ActiveAlerts:          item.input.activeAlerts,
			DeduplicatedAlerts:    item.input.deduplicatedAlerts,
			RawEventCount:         item.input.rawEventCount,
			Owner:                 "",
			SLA:                   defaultSLA(item.priority),
			WorkflowStatus:        defaultWorkflowStatus(item.priority, item.status),
			Progress:              defaultProgress(item),
			ScoreImpact:           item.scoreImpact,
			RiskBreakdown:         riskBreakdown(item),
			UpdatedAt:             item.input.updatedAt,
			StartedAt:             item.input.startedAt,
			MergeBasis:            buildMergeBasis(item),
			MergeReasons:          item.input.mergeReasons,
			DataQuality:           buildDataQuality(item),
			MissingData:           missingSourceLabels(item.input.availableCalcSources),
			EBPFFinding:           buildEBPFFinding(item),
			SystemProblem:         buildSystemProblem(item),
			AIConclusion:          buildAIConclusion(item),
			RootCause:             buildRootCause(item),
			EBPFRole:              buildEBPFRole(item),
			Impact:                item.input.impact,
			DataSources:           item.input.displaySources,
			Actions:               actionTexts(actionPlan),
			ActionPlan:            actionPlan,
			ExistingAlertAnalyses: buildIssueExistingAlertAnalyses(item.input.events),
			Evidence:              item.input.evidence,
		})
	}
	return issues
}

func buildScoreFactors(scored []scoredIssue) []response.SystemHealthScoreFactor {
	factors := make([]response.SystemHealthScoreFactor, 0, minInt(3, len(scored)))
	for i, item := range scored {
		if i >= 3 {
			break
		}
		factors = append(factors, response.SystemHealthScoreFactor{
			Key:    item.input.key,
			Label:  item.priority + " " + item.input.title,
			Value:  item.scoreImpact,
			Tone:   scoreTone(item.status),
			Reason: scoreReason(item),
		})
	}
	return factors
}

func buildActionPlan(item scoredIssue) []response.SystemHealthActionPlanStep {
	basePlan := buildDefaultActionPlan(item)
	reportPlan := alertReportActionPlanSteps(item.input.alertReportActions, item.input.scope)
	return mergeActionPlanSteps(reportPlan, basePlan)
}

func buildDefaultActionPlan(item scoredIssue) []response.SystemHealthActionPlanStep {
	switch item.input.kind {
	case issueKindRecoveredAlert:
		return []response.SystemHealthActionPlanStep{
			{Phase: "恢复验证", Target: "确认有效告警恢复后不再复发。", Action: "查看该对象的历史告警和最新指标趋势。", Validation: "恢复观察窗口内无新的 firing 事件。", Risk: "恢复期仍可能出现同类告警复发。"},
		}
	case issueKindInspection:
		return []response.SystemHealthActionPlanStep{
			{Phase: "异常确认", Target: "确认巡检异常是否影响业务入口。", Action: "查看该对象相关指标趋势和关联服务状态。", Validation: "指标回到阈值内或确认无业务影响。", Risk: "巡检异常可能提前于正式告警出现。"},
		}
	case issueKindInvalidAlert:
		return []response.SystemHealthActionPlanStep{
			{Phase: "规则诊断", Target: "判断告警指标是否误报。", Action: "对比告警触发指标与同窗口服务 RED 指标，确认请求量、错误率和延迟是否同步异常。", Validation: "RED 指标无明显变化且告警持续被识别为无效时，进入告警规则修正流程。", Risk: "直接屏蔽告警可能掩盖真实异常，应先确认指标口径。"},
			{Phase: "规则修正", Target: "降低无效告警反复触发。", Action: "检查告警表达式、标签匹配、阈值、同比基线和聚合维度，必要时调整规则或静默条件。", Validation: "规则调整后无效告警数量下降，真实业务异常仍能触发有效告警。", Risk: "阈值放宽过度会降低告警灵敏度。"},
		}
	default:
		return []response.SystemHealthActionPlanStep{
			{Phase: "影响控制", Target: "优先处理仍在触发的有效告警。", Action: "进入关联告警，确认告警对象、级别、最近触发时间和影响范围。", Validation: "有效 firing 告警数量下降，关键对象状态恢复。", Risk: "只关闭告警不代表业务影响已经恢复。"},
			{Phase: "恢复验证", Target: "确认问题不再扩散或复发。", Action: "持续观察该对象的告警状态和相关指标。", Validation: "观察窗口内无新增有效告警。", Risk: "下游影响可能滞后于告警恢复。"},
		}
	}
}

func alertReportActionPlanSteps(actions []systemHealthAlertReportAction, scope string) []response.SystemHealthActionPlanStep {
	steps := make([]response.SystemHealthActionPlanStep, 0, len(actions))
	for _, action := range actions {
		actionText := strings.TrimSpace(action.Action)
		if actionText == "" {
			continue
		}
		target := firstNonEmpty(action.AlertName, scope)
		steps = append(steps, response.SystemHealthActionPlanStep{
			Phase:      alertReportActionPhase(action.Level),
			Target:     target,
			Action:     actionText,
			Validation: "执行后回看关联告警、关键指标和报告证据，确认风险已收敛。",
			Risk:       "来自原告警分析报告，执行前需结合当前系统健康状态确认适用性。",
		})
	}
	return steps
}

func alertReportActionPhase(level string) string {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "high", "critical", "p0", "p1":
		return "原报告高优先级建议"
	case "medium", "warning", "p2":
		return "原报告建议"
	case "low", "info", "p3":
		return "原报告观察建议"
	default:
		return "原报告建议"
	}
}

func actionsToActionPlanSteps(actions []string) []response.SystemHealthActionPlanStep {
	steps := make([]response.SystemHealthActionPlanStep, 0, len(actions))
	for index, action := range trimNonEmptyStrings(actions) {
		steps = append(steps, response.SystemHealthActionPlanStep{
			Phase:  fmt.Sprintf("建议 %d", index+1),
			Action: action,
		})
	}
	return steps
}

func mergeActionPlanSteps(primary []response.SystemHealthActionPlanStep, secondary []response.SystemHealthActionPlanStep) []response.SystemHealthActionPlanStep {
	merged := make([]response.SystemHealthActionPlanStep, 0, len(primary)+len(secondary))
	seen := make(map[string]struct{}, len(primary)+len(secondary))
	appendStep := func(step response.SystemHealthActionPlanStep) {
		step.Action = strings.TrimSpace(step.Action)
		if step.Action == "" {
			return
		}
		key := strings.ToLower(step.Action)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		merged = append(merged, step)
	}
	for _, step := range primary {
		appendStep(step)
	}
	for _, step := range secondary {
		appendStep(step)
	}
	return merged
}

func calculateHealthScore(scored []scoredIssue) int {
	if len(scored) == 0 {
		return 100
	}
	risks := make([]float64, 0, len(scored))
	for _, item := range scored {
		risks = append(risks, item.pathRisk)
	}
	sort.Sort(sort.Reverse(sort.Float64Slice(risks)))
	systemRisk := 0.55*risks[0] + 0.30*avgFloat(risks[:minInt(3, len(risks))]) + 0.15*avgFloat(risks)
	return int(math.Round(100 - clamp(systemRisk, 0, 100)))
}

func entryImpactScore(input reportInput) float64 {
	duration := durationRisk(input.durationSeconds)
	activePressure := math.Min(float64(maxInt(input.activeAlerts, input.deduplicatedAlerts))/4, 1) * 100
	switch input.kind {
	case issueKindActiveAlert:
		return clamp(0.45*activePressure+0.35*input.objectCriticality+0.20*duration, 0, 100)
	case issueKindInspection:
		return clamp(0.45*input.objectCriticality+0.35*duration+20, 0, 70)
	case issueKindRecoveredAlert:
		return clamp(0.25*input.objectCriticality+0.15*duration, 0, 35)
	case issueKindInvalidAlert:
		return clamp(0.15*activePressure+0.10*duration, 0, 18)
	default:
		return 0
	}
}

func alertSeverityScore(severities []string, activeAlerts int) float64 {
	if len(severities) == 0 {
		return 0
	}
	maxSeverity := 0.0
	for _, severity := range severities {
		maxSeverity = math.Max(maxSeverity, severityRisk(severity))
	}
	alertCountPressure := math.Min(float64(maxInt(activeAlerts, len(severities)))/5, 1) * 100
	return 0.70*maxSeverity + 0.30*alertCountPressure
}

func redDegradationScore(input reportInput) float64 {
	if input.kind == issueKindInspection {
		return 45
	}
	return 0
}

func propagationScore(input reportInput) float64 {
	if input.deduplicatedAlerts <= 1 && input.repeatCount == 0 {
		return 0
	}
	instanceSpread := math.Min(float64(maxInt(input.deduplicatedAlerts-1, 0))/3, 1) * 100
	repeatSpread := repeatPressure(input.repeatCount)
	scopeBoost := 0.0
	if input.scopeKind == "entry" || input.scopeKind == "service" {
		scopeBoost = 15
	}
	return clamp(0.55*instanceSpread+0.30*repeatSpread+scopeBoost, 0, 65)
}

func durationRisk(seconds float64) float64 {
	switch {
	case seconds <= 0:
		return 0
	case seconds < 5*60:
		return 15
	case seconds < 15*60:
		return 45
	case seconds < 60*60:
		return 75
	default:
		return 100
	}
}

func repeatPressure(repeatCount int) float64 {
	if repeatCount <= 0 {
		return 0
	}
	return math.Min(float64(repeatCount)/5, 1) * 100
}

func evidenceConsistency(input reportInput) float64 {
	if len(input.normalizedEvents) <= 1 {
		return 1
	}
	aggregationKeys := make([]string, 0, len(input.normalizedEvents))
	objectKeys := make([]string, 0, len(input.normalizedEvents))
	for _, event := range input.normalizedEvents {
		aggregationKeys = append(aggregationKeys, event.AggregationKey)
		objectKeys = append(objectKeys, event.ObjectType+"::"+event.ObjectName)
	}
	aggregationConsistency := 0.0
	if len(uniqueStrings(aggregationKeys)) == 1 {
		aggregationConsistency = 1
	}
	objectConsistency := 0.0
	if len(uniqueStrings(objectKeys)) == 1 {
		objectConsistency = 1
	}
	timeConsistency := 1.0
	if input.durationSeconds == 0 && input.rawEventCount > 1 {
		timeConsistency = 0.5
	}
	return (aggregationConsistency + objectConsistency + timeConsistency) / 3
}

func topologyCompleteness(input reportInput) float64 {
	switch input.scopeKind {
	case "entry":
		return 1
	case "service":
		return 0.75
	case "database", "pod", "node", "target":
		return 0.45
	default:
		return 0.30
	}
}

func dataCompletenessPenalty(available []string) float64 {
	missing := len(missingSourceLabels(available))
	if len(expectedSources) == 0 {
		return 0
	}
	return math.Min(float64(missing)*100/float64(len(expectedSources)), 100)
}

func issueDataQualityScore(input reportInput) int {
	sourceCoverage := expectedSourceCoverage(input.availableCalcSources)
	evidenceConsistency := evidenceConsistency(input)
	topologyCompleteness := topologyCompleteness(input)
	freshness := 1.0
	if len(input.events) == 0 {
		freshness = 0
	} else {
		latest := input.events[0].UpdateTime
		for _, event := range input.events {
			if event.UpdateTime.After(latest) {
				latest = event.UpdateTime
			}
		}
		freshness = clamp(1-time.Since(latest).Seconds()/(15*60), 0, 1)
		if input.kind == issueKindRecoveredAlert {
			freshness = 1
		}
	}
	dataQualityScore := 100 * (0.45*sourceCoverage + 0.25*evidenceConsistency + 0.20*topologyCompleteness + 0.10*freshness)
	return int(math.Round(clamp(dataQualityScore, 0, 100)))
}

func sourceLabel(source string) string {
	if label, ok := sourceLabels[source]; ok {
		return label
	}
	return source
}

func expectedSourceCoverage(available []string) float64 {
	if len(expectedSources) == 0 {
		return 1
	}
	availableSet := make(map[string]struct{}, len(available))
	for _, source := range available {
		availableSet[source] = struct{}{}
	}
	covered := 0
	for _, source := range expectedSources {
		if _, ok := availableSet[source]; ok {
			covered++
		}
	}
	return float64(covered) / float64(len(expectedSources))
}

func missingSourceLabels(available []string) []string {
	availableSet := make(map[string]struct{}, len(available))
	for _, source := range available {
		availableSet[source] = struct{}{}
	}
	missing := make([]string, 0, len(expectedSources))
	for _, source := range expectedSources {
		if _, ok := availableSet[source]; ok {
			continue
		}
		missing = append(missing, sourceLabel(source))
	}
	return missing
}

func reportAvailableSources(scored []scoredIssue) []string {
	sources := make([]string, 0)
	for _, item := range scored {
		sources = append(sources, item.input.displaySources...)
	}
	if len(sources) == 0 {
		sources = append(sources, sourceLabel("alerts"))
	}
	return uniqueStrings(sources)
}

func riskBreakdown(item scoredIssue) response.SystemHealthRiskBreakdown {
	return response.SystemHealthRiskBreakdown{
		EntryImpactScore:        int(math.Round(item.entryImpactScore)),
		AlertSeverityScore:      int(math.Round(item.alertSeverityScore)),
		RedDegradationScore:     int(math.Round(item.redDegradationScore)),
		PropagationScore:        int(math.Round(item.propagationScore)),
		DataCompletenessPenalty: int(math.Round(item.dataCompletenessPenalty)),
		PathRisk:                int(math.Round(item.pathRisk)),
	}
}

func maxSeverity(severities []string) string {
	if len(severities) == 0 {
		return ""
	}
	result := strings.TrimSpace(severities[0])
	maxRisk := severityRisk(result)
	for _, severity := range severities[1:] {
		if risk := severityRisk(severity); risk > maxRisk {
			result = strings.TrimSpace(severity)
			maxRisk = risk
		}
	}
	return result
}

func affectedObjectsFromEvent(event alertmodel.AEventWithWRecord) []response.SystemHealthAffectedObject {
	objects := []response.SystemHealthAffectedObject{
		{Type: "endpoint", Name: alertEndpoint(event)},
		{Type: "target", Name: event.GetTargetObj()},
		{Type: "service", Name: alertServiceName(event)},
		{Type: "node", Name: event.GetInfraNodeTag()},
		{Type: "database", Name: event.GetDatabaseURL()},
		{Type: "namespace", Name: event.GetK8sNamespaceTag()},
		{Type: "pod", Name: event.GetK8sPodTag()},
		{Type: "group", Name: event.Group},
	}
	return uniqueAffectedObjects(objects)
}

func uniqueAffectedObjects(objects []response.SystemHealthAffectedObject) []response.SystemHealthAffectedObject {
	seen := make(map[string]struct{}, len(objects))
	result := make([]response.SystemHealthAffectedObject, 0, len(objects))
	for _, object := range objects {
		object.Type = strings.TrimSpace(object.Type)
		object.Name = strings.TrimSpace(object.Name)
		if object.Type == "" || object.Name == "" {
			continue
		}
		key := object.Type + "::" + object.Name
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, object)
	}
	return result
}

func issueWeightInSystemRisk(scored []scoredIssue, index int) float64 {
	if len(scored) == 0 {
		return 0
	}
	maxRisk := scored[0].pathRisk
	maxPart := 0.0
	if scored[index].pathRisk == maxRisk {
		maxPart = 0.55
	}
	top3Part := 0.0
	if index < minInt(3, len(scored)) {
		top3Part = 0.30 / float64(minInt(3, len(scored)))
	}
	avgPart := 0.15 / float64(len(scored))
	return maxPart + top3Part + avgPart
}

func severityRisk(severity string) float64 {
	switch strings.ToLower(severity) {
	case "critical", "disaster", "high":
		return 100
	case "error", "average":
		return 80
	case "warning", "warn":
		return 45
	case "info", "information":
		return 20
	default:
		return 30
	}
}

func priority(pathRisk float64, severities []string, issueKind string) string {
	switch issueKind {
	case issueKindInvalidAlert:
		return "P4"
	case issueKindRecoveredAlert:
		return "P3"
	case issueKindInspection:
		return "P2"
	}
	for _, severity := range severities {
		if severityRisk(severity) == 100 {
			return "P0"
		}
	}
	if pathRisk >= 75 {
		return "P0"
	}
	if pathRisk >= 50 {
		return "P1"
	}
	return "P2"
}

func issueStatus(pathRisk float64) string {
	switch {
	case pathRisk >= 75:
		return "severe"
	case pathRisk >= 50:
		return "warning"
	case pathRisk >= 25:
		return "minor"
	default:
		return "healthy"
	}
}

func healthStatus(score int) string {
	switch {
	case score >= 90:
		return "healthy"
	case score >= 75:
		return "minor"
	case score >= 60:
		return "warning"
	case score >= 40:
		return "severe"
	default:
		return "outage"
	}
}

func priorityRank(priority string) int {
	switch priority {
	case "P0":
		return 0
	case "P1":
		return 1
	case "P2":
		return 2
	case "P3":
		return 3
	case "P4":
		return 4
	default:
		return 9
	}
}

func scoreTone(status string) string {
	switch status {
	case "severe", "outage":
		return "severe"
	case "warning":
		return "warning"
	default:
		return "minor"
	}
}

func scoreReason(item scoredIssue) string {
	return strings.Join([]string{
		"入口/对象影响 " + strconvRound(item.entryImpactScore),
		"告警严重度 " + strconvRound(item.alertSeverityScore),
		"RED退化 " + strconvRound(item.redDegradationScore),
		"传播风险 " + strconvRound(item.propagationScore),
		"数据缺口 " + strconvRound(item.dataCompletenessPenalty),
	}, " / ")
}

func buildSummary(scored []scoredIssue, healthScore int) string {
	if len(scored) == 0 {
		return "当前时间窗口内已接入的告警事件中，未发现可归纳的有效告警、巡检异常或已恢复有效告警。"
	}
	activeCount := 0
	recoveringCount := 0
	ruleDiagnosticCount := 0
	for _, item := range scored {
		activeCount += item.input.activeAlerts
		if item.input.activeAlerts == 0 || item.input.kind == issueKindRecoveredAlert {
			recoveringCount++
		}
		if item.input.kind == issueKindInvalidAlert {
			ruleDiagnosticCount++
		}
	}
	focus := "需要关注告警是否复发、服务 RED 指标是否退化，以及告警分析是否指向共同根因"
	if ruleDiagnosticCount > 0 {
		focus = "需要关注告警规则口径是否与服务 RED 指标背离，避免把无效告警当作服务故障处理"
	} else if activeCount > 0 {
		focus = "需要关注告警影响是否继续扩大，并优先确认入口、下游和关键指标的恢复情况"
	}
	return fmt.Sprintf("系统当前处于%s状态。%s。%s。",
		healthSummaryState(healthScore),
		reportEventDescriptionText(scored, activeCount, recoveringCount),
		focus,
	)
}

func healthSummaryState(score int) string {
	switch {
	case score >= 90:
		return "正常"
	case score >= 75:
		return "轻微异常"
	case score >= 60:
		return "警告"
	case score >= 40:
		return "严重"
	default:
		return "故障"
	}
}

func reportEventPatternText(scored []scoredIssue, activeCount int, recoveringCount int) string {
	totalEvents := 0
	totalInstances := 0
	for _, item := range scored {
		totalEvents += item.input.rawEventCount
		totalInstances += item.input.deduplicatedAlerts
	}
	repeatSeen := totalEvents >= 5 || (totalInstances > 0 && totalEvents > totalInstances)
	if activeCount > 0 && repeatSeen {
		return "活动告警仍在触发，且窗口内出现反复事件"
	}
	if activeCount > 0 {
		return "活动告警尚未清零"
	}
	if recoveringCount > 0 && repeatSeen {
		return "告警已清零，但窗口内反复出现后进入恢复观察"
	}
	if recoveringCount > 0 {
		return "告警已清零，目前处于恢复观察"
	}
	return "当前未见持续告警"
}

func reportEventDescriptionText(scored []scoredIssue, activeCount int, recoveringCount int) string {
	totalEvents := 0
	totalInstances := 0
	for _, item := range scored {
		totalEvents += item.input.rawEventCount
		totalInstances += item.input.deduplicatedAlerts
	}
	repeatSeen := totalEvents >= 5 || (totalInstances > 0 && totalEvents > totalInstances)
	eventPrefix := "当前窗口识别到"
	if totalEvents > 0 {
		eventPrefix = "窗口内发生过"
	}
	repeatText := ""
	if repeatSeen {
		repeatText = "，并出现反复触发"
	}
	if activeCount > 0 {
		return fmt.Sprintf("%s%s%s告警%s，目前仍有告警未恢复", eventPrefix, reportScopePatternText(scored), reportSymptomText(scored), repeatText)
	}
	if recoveringCount > 0 {
		return fmt.Sprintf("%s%s%s告警%s，目前告警已清零，相关对象处于恢复观察", eventPrefix, reportScopePatternText(scored), reportSymptomText(scored), repeatText)
	}
	return "当前窗口未发现持续触发的核心告警，系统处于观察状态"
}

func reportScopePatternText(scored []scoredIssue) string {
	countByKind := map[string]int{}
	for _, item := range scored {
		countByKind[item.input.scopeKind]++
	}
	if countByKind["entry"] > 0 && countByKind["service"] > 0 && len(scored) > 3 {
		return "多入口/多服务"
	}
	if countByKind["entry"] > 0 {
		if len(scored) > 1 {
			return "多入口"
		}
		return "单入口"
	}
	if countByKind["service"] > 0 {
		if len(scored) > 1 {
			return "多服务"
		}
		return "单服务"
	}
	if countByKind["node"] > 0 {
		if len(scored) > 1 {
			return "多节点"
		}
		return "单节点"
	}
	if countByKind["pod"] > 0 {
		if len(scored) > 1 {
			return "多实例"
		}
		return "单实例"
	}
	if len(scored) > 1 {
		return "多对象"
	}
	return "单对象"
}

func reportSymptomText(scored []scoredIssue) string {
	counts := map[string]int{}
	for _, item := range scored {
		counts[issueSymptomText(item)]++
	}
	topLabel := "健康风险"
	topCount := 0
	for label, count := range counts {
		if count > topCount {
			topLabel = label
			topCount = count
		}
	}
	return topLabel
}

func issueSymptomText(item scoredIssue) string {
	if item.input.kind == issueKindInvalidAlert {
		return "告警规则"
	}
	text := strings.ToLower(strings.Join(append([]string{
		item.input.title,
		item.input.scope,
		item.input.impact,
	}, item.input.alertNames...), " "))
	switch {
	case strings.Contains(text, "latency"), strings.Contains(text, "duration"), strings.Contains(text, "timeout"), strings.Contains(text, "slow"),
		strings.Contains(text, "p95"), strings.Contains(text, "p99"), strings.Contains(text, "耗时"), strings.Contains(text, "延时"), strings.Contains(text, "延迟"), strings.Contains(text, "慢调用"), strings.Contains(text, "超时"):
		return "请求延时"
	case strings.Contains(text, "error"), strings.Contains(text, "exception"), strings.Contains(text, "fail"),
		strings.Contains(text, "失败"), strings.Contains(text, "错误"), strings.Contains(text, "异常"), strings.Contains(text, "5xx"), strings.Contains(text, "4xx"):
		return "错误异常"
	case strings.Contains(text, "cpu"), strings.Contains(text, "memory"), strings.Contains(text, "oom"), strings.Contains(text, "磁盘"), strings.Contains(text, "内存"), strings.Contains(text, "负载"), strings.Contains(text, "资源"):
		return "资源压力"
	case strings.Contains(text, "unavailable"), strings.Contains(text, "down"), strings.Contains(text, "不可用"), strings.Contains(text, "宕机"), strings.Contains(text, "断连"):
		return "可用性"
	case strings.Contains(text, "规则"), strings.Contains(text, "阈值"), strings.Contains(text, "表达式"), strings.Contains(text, "无效告警"), strings.Contains(text, "invalid"):
		return "告警规则"
	default:
		return "健康风险"
	}
}

func buildMergeBasis(item scoredIssue) string {
	reasons := item.input.mergeReasons
	if len(reasons) == 0 {
		reasons = []string{"同一告警对象"}
	}
	return fmt.Sprintf("%d 条真实事件去重为 %d 个告警实例，按%s归并到 %s。", item.input.rawEventCount, item.input.deduplicatedAlerts, strings.Join(reasons, "、"), item.input.scope)
}

func buildDataQuality(item scoredIssue) []string {
	missing := missingSourceLabels(item.input.availableCalcSources)
	if item.dataCompletenessPenalty == 0 {
		return []string{"真实告警事件已覆盖", "未使用本地内置样例数据"}
	}
	return []string{
		"真实数据源覆盖不完整，已计入数据完整性惩罚",
		"未用内置样例数据补齐缺失字段",
		"待接入：" + strings.Join(missing, "、"),
	}
}

func buildEBPFFinding(item scoredIssue) response.SystemHealthEBPFFinding {
	return response.SystemHealthEBPFFinding{
		Observed: "当前只接入真实告警事件。",
		Excluded: "未使用 APM、Trace、日志、eBPF 或基础设施指标生成额外推断。",
		Narrowed: "根因范围仅来自真实告警事件、标签和告警详情。",
	}
}

func buildSystemProblem(item scoredIssue) string {
	return item.input.impact
}

func buildAIConclusion(item scoredIssue) string {
	if item.input.kind == issueKindInvalidAlert {
		return "该对象存在持续触发但被识别为无效的告警，优先按 P4 规则诊断处理：对比告警触发指标与服务 RED 指标，确认是否为告警表达式、阈值、标签匹配或同比基线问题。"
	}
	missing := missingSourceLabels(item.input.availableCalcSources)
	conclusion := "AI 基于 " + strconv.Itoa(item.input.rawEventCount) + " 条真实事件去重后的 " + strconv.Itoa(item.input.deduplicatedAlerts) +
		" 个告警实例归纳，风险对象为 " + item.input.scope + "，计算路径风险 " + strconvRound(item.pathRisk) + "，优先级为 " + item.priority + "。"
	if len(missing) > 0 {
		conclusion += " 未接入的 " + strings.Join(missing, "、") + " 不参与推断。"
	}
	return conclusion
}

func buildRootCause(item scoredIssue) string {
	if item.input.kind == issueKindInvalidAlert {
		return "告警本身持续触发但已有分析判定无效，根因优先考虑告警表达式、阈值、标签匹配或指标口径问题。"
	}
	directions := alertDirections(item.input.events)
	if len(directions) == 0 {
		return ""
	}
	roots := make([]string, 0, len(directions))
	for _, direction := range directions {
		if root := rootCauseFromAlertDirection(direction); root != "" {
			roots = append(roots, root)
		}
	}
	roots = uniqueStrings(roots)
	if len(roots) == 0 {
		return ""
	}
	if len(roots) == 1 {
		return "告警发生时分析指向：" + truncateText(roots[0], 180)
	}
	return "告警发生时分析给出多个可能根因：" + truncateText(strings.Join(roots[:minInt(len(roots), 3)], "；"), 220)
}

func rootCauseFromAlertDirection(direction string) string {
	root := strings.TrimSpace(direction)
	if root == "" {
		return ""
	}
	for _, prefix := range []string{"疑似", "可能是", "可能为", "可能", "初步判断", "分析认为", "根因可能是", "根因可能为"} {
		root = strings.TrimSpace(strings.TrimPrefix(root, prefix))
	}
	for _, separator := range []string{"导致", "引发", "造成", "使得"} {
		if left, _, ok := strings.Cut(root, separator); ok && strings.TrimSpace(left) != "" {
			root = strings.TrimSpace(left)
			break
		}
	}
	return truncateText(root, 220)
}

func buildEBPFRole(item scoredIssue) string {
	return "当前只接入真实告警事件，因此不展示 APM、Trace、日志、eBPF 或基础设施根因结论。"
}

func actionTexts(actionPlan []response.SystemHealthActionPlanStep) []string {
	actions := make([]string, 0, len(actionPlan))
	for _, step := range actionPlan {
		actions = append(actions, step.Action)
	}
	return actions
}

func defaultSLA(priority string) string {
	switch priority {
	case "P0":
		return "30 分钟内恢复核心影响"
	case "P1":
		return "60 分钟内控制风险升级"
	case "P2":
		return "巡检窗口内完成确认"
	case "P4":
		return "规则维护窗口内完成诊断"
	default:
		return "恢复观察窗口内无复发后关闭"
	}
}

func defaultWorkflowStatus(priority string, status string) string {
	if priority == "P4" {
		return "规则诊断"
	}
	if priority == "P3" {
		return "已恢复"
	}
	if status == "minor" {
		return "恢复观察"
	}
	if priority == "P0" {
		return "处理中"
	}
	return "待处理"
}

func defaultProgress(item scoredIssue) string {
	switch item.input.kind {
	case issueKindRecoveredAlert:
		return "恢复验证"
	case issueKindInspection:
		return "异常确认"
	case issueKindInvalidAlert:
		return "规则诊断"
	default:
		return "影响确认"
	}
}

func countInfraAlerts(scored []scoredIssue) int {
	count := 0
	for _, item := range scored {
		for _, event := range item.input.events {
			if event.Group == "infra" {
				count++
			}
		}
	}
	return count
}

func uniqueActiveAlertCount(scored []scoredIssue) int {
	total := 0
	for _, item := range scored {
		total += item.input.activeAlerts
	}
	return total
}

func resolveTimeRange(req *request.GetSystemHealthReportRequest) (time.Time, time.Time, string) {
	if req.StartTime > 0 && req.EndTime > req.StartTime {
		return time.UnixMicro(req.StartTime), time.UnixMicro(req.EndTime), "custom"
	}
	duration := parseRelativeTime(req.RelativeTime)
	window := strings.TrimSpace(req.RelativeTime)
	if duration <= 0 {
		duration = 15 * time.Minute
		window = "15m"
	}
	endTime := time.Now()
	if enableSystemHealthReportCache {
		endTime = endTime.Truncate(systemHealthReportCacheTTL)
	}
	return endTime.Add(-duration), endTime, window
}

func parseRelativeTime(value string) time.Duration {
	if value == "" {
		return 15 * time.Minute
	}
	duration, err := time.ParseDuration(value)
	if err == nil {
		return duration
	}
	if strings.HasSuffix(value, "d") {
		days, err := strconv.Atoi(strings.TrimSuffix(value, "d"))
		if err == nil {
			return time.Duration(days) * 24 * time.Hour
		}
	}
	return 0
}

func statusText(input reportInput) string {
	statuses := make([]string, 0, len(input.events))
	for _, event := range input.events {
		statuses = append(statuses, event.Status)
	}
	return strings.Join(uniqueStrings(statuses), "/")
}

func stableKey(value string) string {
	value = strings.ToLower(value)
	replacer := strings.NewReplacer("::", "-", "/", "-", " ", "-", "_", "-", ".", "-")
	return strings.Trim(replacer.Replace(value), "-")
}

func formatEventTime(value time.Time) string {
	if value.IsZero() {
		return "-"
	}
	return value.Format("2006-01-02 15:04")
}

func parseDisplayTime(value string) time.Time {
	parsed, err := time.ParseInLocation("2006-01-02 15:04", value, time.Local)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func formatDurationCN(seconds float64) string {
	if seconds <= 0 {
		return "不足 1 分钟"
	}
	duration := time.Duration(seconds) * time.Second
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	if hours > 0 {
		return fmt.Sprintf("%d 小时 %d 分钟", hours, minutes)
	}
	if minutes > 0 {
		return fmt.Sprintf("%d 分钟", minutes)
	}
	return fmt.Sprintf("%d 秒", int(duration.Seconds()))
}

func truncateText(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit]) + "..."
}

func sourceContains(values []string, keyword string) bool {
	for _, value := range values {
		if strings.Contains(value, keyword) {
			return true
		}
	}
	return false
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func uniqueInt64(values []int64) []int64 {
	seen := make(map[int64]struct{}, len(values))
	result := make([]int64, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func avgFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}

func clamp(value, minValue, maxValue float64) float64 {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func strconvRound(value float64) string {
	return strconv.Itoa(int(math.Round(value)))
}

func formatMetricFloat(value float64) string {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return "NaN"
	}
	if math.Abs(value) >= 100 {
		return strconv.FormatFloat(value, 'f', 2, 64)
	}
	if math.Abs(value) >= 1 {
		return strconv.FormatFloat(value, 'f', 4, 64)
	}
	return strconv.FormatFloat(value, 'f', 6, 64)
}

func formatSignedMetricFloat(value float64) string {
	formatted := formatMetricFloat(value)
	if value > 0 {
		return "+" + formatted
	}
	return formatted
}
