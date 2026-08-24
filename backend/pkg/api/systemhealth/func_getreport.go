// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package systemhealth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
	alertmodel "github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/common"
	"github.com/google/uuid"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prommodel "github.com/prometheus/common/model"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	"requestPolaris":         "请求北极星",
	"infra":                  "基础设施指标",
	"alertReport":            "原告警分析报告",
}

var (
	htmlDetailsBlockPattern  = regexp.MustCompile(`(?is)<details\b[^>]*>.*?</details>`)
	htmlTagPattern           = regexp.MustCompile(`(?is)<[^>]+>`)
	markdownCodeFencePattern = regexp.MustCompile("(?s)```.*?```")
	cjkTextPattern           = regexp.MustCompile(`\p{Han}`)
)

const (
	issueKindActiveAlert              = "activeAlert"
	issueKindInspection               = "inspection"
	issueKindRecoveredAlert           = "recoveredAlert"
	issueKindInvalidAlert             = "invalidAlert"
	maxEvidenceSamples                = 3
	maxAlertReportLookups             = 6
	maxMetricEvidenceSeries           = 3
	maxMetricEvidencePoints           = 240
	systemHealthAIParallelism         = 4
	requestPolarisResourceThresholdMS = 50.0
	requestPolarisIncreaseThreshold   = 0.20
	requestPolarisSlowSpanThresholdMS = 500.0
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
	Key                    string                             `json:"key"`
	Priority               string                             `json:"priority"`
	Title                  string                             `json:"title"`
	AlertNames             []string                           `json:"alertNames,omitempty"`
	Status                 string                             `json:"status"`
	Scope                  string                             `json:"scope"`
	Impact                 string                             `json:"impact"`
	ActiveAlerts           int                                `json:"activeAlerts"`
	StartedAt              string                             `json:"startedAt"`
	DataSources            []string                           `json:"dataSources"`
	MetricNarrativeContext string                             `json:"metricNarrativeContext,omitempty"`
	Evidence               []systemHealthAIEvidenceInput      `json:"evidence"`
	ExistingAlertAnalyses  []systemHealthAIAlertAnalysisInput `json:"existingAlertAnalyses"`
	ExistingReportFindings []systemHealthAIReportFindingInput `json:"existingReportFindings"`
	ExistingReportActions  []systemHealthAIReportActionInput  `json:"existingReportActions"`
}

type systemHealthAIEvidenceInput struct {
	Type             string                                         `json:"type"`
	Source           string                                         `json:"source,omitempty"`
	Severity         string                                         `json:"severity,omitempty"`
	Result           string                                         `json:"result"`
	Metric           *systemHealthAIMetricInput                     `json:"metric,omitempty"`
	PolarisBreakdown *response.SystemHealthPolarisBreakdownEvidence `json:"polarisBreakdown,omitempty"`
}

type systemHealthAIMetricInput struct {
	MetricName     string   `json:"metricName,omitempty"`
	RuleName       string   `json:"ruleName,omitempty"`
	SeriesMatched  int      `json:"seriesMatched"`
	SeriesReturned int      `json:"seriesReturned"`
	PointCount     int      `json:"pointCount"`
	FirstValue     float64  `json:"firstValue"`
	LastValue      float64  `json:"lastValue"`
	AverageValue   float64  `json:"averageValue"`
	MinValue       float64  `json:"minValue"`
	MaxValue       float64  `json:"maxValue"`
	MaxDeviation   float64  `json:"maxDeviation"`
	Delta          float64  `json:"delta"`
	ChangePercent  *float64 `json:"changePercent,omitempty"`
	Trend          string   `json:"trend,omitempty"`
	StartAt        string   `json:"startAt,omitempty"`
	EndAt          string   `json:"endAt,omitempty"`
	PeakAt         string   `json:"peakAt,omitempty"`
	MinAt          string   `json:"minAt,omitempty"`
	EventScoped    bool     `json:"eventScoped"`
	Message        string   `json:"message,omitempty"`
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
	AlertName  string `json:"alertName,omitempty"`
	Level      string `json:"level,omitempty"`
	Phase      string `json:"phase,omitempty"`
	Target     string `json:"target,omitempty"`
	Action     string `json:"action"`
	Validation string `json:"validation,omitempty"`
	Risk       string `json:"risk,omitempty"`
}

type systemHealthAIReportFindingInput struct {
	AlertName      string                              `json:"alertName,omitempty"`
	RootCause      string                              `json:"rootCause,omitempty"`
	RootCauseOther []string                            `json:"rootCauseOther,omitempty"`
	AnalyzeProcess string                              `json:"analyzeProcess,omitempty"`
	OverviewReason string                              `json:"overviewReason,omitempty"`
	Detail         string                              `json:"detail,omitempty"`
	OverviewDetail string                              `json:"overviewDetail,omitempty"`
	ReportEvidence []systemHealthAIReportEvidenceInput `json:"reportEvidence,omitempty"`
}

type systemHealthAIReportEvidenceInput struct {
	Kind    string `json:"kind,omitempty"`
	Summary string `json:"summary,omitempty"`
}

type systemHealthAISummaryResponse struct {
	Summary       string                      `json:"summary"`
	WorkflowRunID string                      `json:"workflowRunId"`
	Issues        []systemHealthAIIssueOutput `json:"issues"`
}

type systemHealthAIIssueOutput struct {
	Key            string                                `json:"key"`
	Title          string                                `json:"title"`
	Analysis       string                                `json:"analysis"`
	RootCause      string                                `json:"rootCause"`
	MetricAnalysis string                                `json:"metricAnalysis"`
	Actions        []string                              `json:"actions"`
	ActionPlan     []response.SystemHealthActionPlanStep `json:"actionPlan"`
}

type systemHealthAIEnrichmentStats struct {
	Mode             string
	TotalCalls       int
	SucceededCalls   int
	FailedCalls      int
	SucceededIssues  int
	SummarySucceeded bool
	SummaryFailed    bool
	FailedIssueKeys  []string
	MissingIssueKeys []string
	ErrorMessages    []string
}

type systemHealthAICallResult struct {
	Kind         string
	IssueIndex   int
	IssueKey     string
	SubRequestID string
	Response     systemHealthAISummaryResponse
	Error        error
	Elapsed      time.Duration
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

func systemHealthInfo(message string, fields ...zap.Field) {
	zap.L().Info(message, fields...)
	systemHealthStdLog("info", message, fields...)
}

func systemHealthWarn(message string, fields ...zap.Field) {
	zap.L().Warn(message, fields...)
	systemHealthStdLog("warn", message, fields...)
}

func systemHealthInfoCtx(ctx core.Context, message string, fields ...zap.Field) {
	systemHealthInfo(message, systemHealthRequestFields(ctx, fields...)...)
}

func systemHealthWarnCtx(ctx core.Context, message string, fields ...zap.Field) {
	systemHealthWarn(message, systemHealthRequestFields(ctx, fields...)...)
}

func systemHealthRequestFields(ctx core.Context, fields ...zap.Field) []zap.Field {
	requestID := systemHealthRequestID(ctx)
	result := make([]zap.Field, 0, len(fields)+3)
	if requestID != "" {
		result = append(result, zap.String("requestId", requestID))
	}
	if ctx != nil {
		if method, path := ctx.GetMethodPath(); method != "" || path != "" {
			result = append(result, zap.String("routeMethod", method), zap.String("routePath", path))
		}
	}
	result = append(result, fields...)
	return result
}

func systemHealthStageFields(ctx core.Context, startedAt time.Time, stage string, fields ...zap.Field) []zap.Field {
	result := make([]zap.Field, 0, len(fields)+2)
	result = append(result,
		zap.String("stage", stage),
		zap.Duration("elapsed", time.Since(startedAt)),
	)
	result = append(result, fields...)
	return systemHealthRequestFields(ctx, result...)
}

func systemHealthRequestID(ctx core.Context) string {
	if ctx == nil {
		return ""
	}
	value, ok := ctx.Get(core.RequestIDKey)
	if !ok {
		return ""
	}
	requestID, _ := value.(string)
	return strings.TrimSpace(requestID)
}

type systemHealthAIRequestContext struct {
	core.Context
	requestContext context.Context
	requestID      string
	routeMethod    string
	routePath      string
}

func systemHealthAIRequestLogContext(ctx core.Context) core.Context {
	if ctx == nil {
		return nil
	}
	method, path := ctx.GetMethodPath()
	return systemHealthAIRequestContext{
		requestContext: context.Background(),
		requestID:      systemHealthRequestID(ctx),
		routeMethod:    method,
		routePath:      path,
	}
}

func (c systemHealthAIRequestContext) GetContext() context.Context {
	if c.requestContext == nil {
		return context.Background()
	}
	return c.requestContext
}

func (c systemHealthAIRequestContext) Get(key string) (any, bool) {
	if key == core.RequestIDKey && c.requestID != "" {
		return c.requestID, true
	}
	return nil, false
}

func (c systemHealthAIRequestContext) GetMethodPath() (method string, path string) {
	return c.routeMethod, c.routePath
}

func systemHealthCompactJSON(value any) string {
	if value == nil {
		return ""
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprint(value)
	}
	return string(payload)
}

func systemHealthJSONSize(value any) int {
	payload, err := json.Marshal(value)
	if err != nil {
		return 0
	}
	return len(payload)
}

func systemHealthShortHash(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(value))
	return fmt.Sprintf("%x", sum[:])[:12]
}

func systemHealthStdLog(level string, message string, fields ...zap.Field) {
	parts := make([]string, 0, len(fields)+2)
	parts = append(parts, "level="+level, "msg="+strconv.Quote(message))
	for _, field := range fields {
		if field.Key == "" {
			continue
		}
		parts = append(parts, field.Key+"="+strconv.Quote(systemHealthZapFieldValue(field)))
	}
	log.Printf("system-health-diagnostic %s", strings.Join(parts, " "))
}

func systemHealthZapFieldValue(field zap.Field) string {
	switch field.Type {
	case zapcore.StringType:
		return field.String
	case zapcore.BoolType:
		return strconv.FormatBool(field.Integer == 1)
	case zapcore.Int64Type, zapcore.Int32Type, zapcore.Int16Type, zapcore.Int8Type,
		zapcore.Uint64Type, zapcore.Uint32Type, zapcore.Uint16Type, zapcore.Uint8Type,
		zapcore.UintptrType:
		return strconv.FormatInt(field.Integer, 10)
	case zapcore.Float64Type:
		return strconv.FormatFloat(math.Float64frombits(uint64(field.Integer)), 'f', -1, 64)
	case zapcore.Float32Type:
		return strconv.FormatFloat(float64(math.Float32frombits(uint32(field.Integer))), 'f', -1, 32)
	case zapcore.DurationType:
		return time.Duration(field.Integer).String()
	case zapcore.TimeType:
		location, _ := field.Interface.(*time.Location)
		if location == nil {
			location = time.Local
		}
		return time.Unix(0, field.Integer).In(location).Format(time.RFC3339Nano)
	case zapcore.ErrorType:
		if err, ok := field.Interface.(error); ok && err != nil {
			return err.Error()
		}
		return ""
	case zapcore.StringerType:
		if stringer, ok := field.Interface.(fmt.Stringer); ok && stringer != nil {
			return stringer.String()
		}
		return fmt.Sprint(field.Interface)
	case zapcore.ReflectType, zapcore.ObjectMarshalerType, zapcore.ArrayMarshalerType:
		return fmt.Sprint(field.Interface)
	default:
		if field.Interface != nil {
			return fmt.Sprint(field.Interface)
		}
		if field.String != "" {
			return field.String
		}
		return strconv.FormatInt(field.Integer, 10)
	}
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
	key                      string
	aggregationKey           string
	kind                     string
	title                    string
	scope                    string
	scopeKind                string
	startedAt                string
	updatedAt                string
	activeAlerts             int
	deduplicatedAlerts       int
	effectiveAlerts          int
	invalidAlerts            int
	rawEventCount            int
	repeatCount              int
	durationSeconds          float64
	objectCriticality        float64
	alertSeverities          []string
	alertNames               []string
	events                   []alertmodel.AEventWithWRecord
	normalizedEvents         []normalizedAlertEvent
	deduplicatedInstances    []deduplicatedAlertInstance
	availableCalcSources     []string
	displaySources           []string
	affectedObjects          []response.SystemHealthAffectedObject
	mergeReasons             []string
	alertReportActions       []systemHealthAlertReportAction
	alertReportFindings      []systemHealthAlertReportFinding
	alertReportTopologyRoles []systemHealthAlertReportTopologyRole
	alertHistoryStats        []clickhouse.AlertEventHistoryStats
	alertHistorySegments     []clickhouse.AlertEventRawSegment
	alertHistoryStartTime    time.Time
	alertHistoryEndTime      time.Time
	evidence                 []response.SystemHealthEvidence
	impact                   string
}

type systemHealthAlertReportAction struct {
	AlertName  string
	Level      string
	Phase      string
	Target     string
	Action     string
	Validation string
	Risk       string
}

type systemHealthAlertReportFinding struct {
	AlertName      string
	RootCause      string
	RootCauseOther []string
	AnalyzeProcess string
	Detail         string
	Evidence       []systemHealthAlertReportEvidence
}

type systemHealthAlertReportEvidence struct {
	Kind    string
	Summary string
}

type systemHealthAlertReportTopologyRole struct {
	AlertName          string
	ReportAlertEventID string
	CurrentService     string
	CurrentEndpoint    string
	ParentCount        int
	ChildCount         int
	ParentNames        []string
	ChildNames         []string
	Role               string
	Reason             string
	Summary            string
	HasTopology        bool
}

type scoredIssue struct {
	input               reportInput
	entryImpactScore    float64
	alertSeverityScore  float64
	redDegradationScore float64
	propagationScore    float64
	dataCompletenessGap float64
	pathRisk            float64
	scoreImpact         int
	priority            string
	status              string
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
	AverageValue  float64
	MinValue      float64
	MaxValue      float64
	MaxDeviation  float64
	Delta         float64
	ChangePercent *float64
}

type alertMetricTrendQueryResult struct {
	Rule           *request.AlertRule
	MetricName     string
	Expression     string
	Step           time.Duration
	StartTime      time.Time
	EndTime        time.Time
	SeriesReturned int
	MatchedSeries  []alertMetricSeries
	Stats          alertMetricTrendStats
	MatchLabels    map[string]string
	EventScoped    bool
	Message        string
}

type requestPolarisMetricTrend struct {
	Key          string
	MetricName   string
	Unit         string
	Expression   string
	Stats        alertMetricTrendStats
	Series       []alertMetricSeries
	SeriesCount  int
	PointCount   int
	QueryErr     string
	Abnormal     bool
	Reason       string
	ResourceKind string
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
	IssueKind      string
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
		if req := c.Request(); req != nil && req.URL != nil {
			systemHealthInfoCtx(c, "system-health report request received",
				zap.String("rawQuery", req.URL.RawQuery),
				zap.String("clientIP", c.ClientIP()),
			)
		} else {
			systemHealthInfoCtx(c, "system-health report request received")
		}
		req := new(request.GetSystemHealthReportRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			systemHealthWarnCtx(c, "system-health report request bind failed",
				zap.Error(err),
			)
			c.AbortWithError(http.StatusBadRequest, code.ParamBindError, err)
			return
		}
		authorizedGroupIDs, err := h.authorizedSystemHealthReportGroupIDs(c)
		if err != nil {
			c.AbortWithError(http.StatusForbidden, code.AuthError, err)
			return
		}
		if !systemHealthReportGroupAllowed(authorizedGroupIDs, req.GroupID) {
			c.AbortWithError(http.StatusForbidden, code.UserNoPermissionError, nil)
			return
		}
		systemHealthInfoCtx(c, "system-health report request bound",
			zap.Int64("groupId", req.GroupID),
			zap.String("relativeTime", req.RelativeTime),
			zap.Int64("startTimeMicros", req.StartTime),
			zap.Int64("endTimeMicros", req.EndTime),
		)

		resp, err := h.buildReport(c, req)
		if err != nil {
			systemHealthWarnCtx(c, "system-health report request failed",
				zap.Int64("groupId", req.GroupID),
				zap.String("relativeTime", req.RelativeTime),
				zap.Int64("startTimeMicros", req.StartTime),
				zap.Int64("endTimeMicros", req.EndTime),
				zap.Error(err),
			)
			c.AbortWithError(http.StatusBadRequest, code.ServerError, err)
			return
		}
		systemHealthInfoCtx(c, "system-health report response ready",
			zap.Int64("groupId", req.GroupID),
			zap.String("status", resp.Report.Status),
			zap.Int("score", resp.Report.Score),
			zap.Int("issueCount", len(resp.Issues)),
			zap.String("aiStatus", resp.Report.AIAnalysis.Status),
			zap.Int("payloadBytes", systemHealthJSONSize(resp)),
		)
		c.Payload(resp)
	}
}

func (h *handler) buildReport(ctx core.Context, req *request.GetSystemHealthReportRequest) (response.SystemHealthReportResp, error) {
	buildStartedAt := time.Now()
	startTime, endTime, window := resolveTimeRange(req)
	systemHealthInfo("system-health report build started", systemHealthStageFields(ctx, buildStartedAt, "01_resolve_window",
		zap.Int64("groupId", req.GroupID),
		zap.String("window", window),
		zap.String("relativeTime", req.RelativeTime),
		zap.Time("startTime", startTime),
		zap.Time("endTime", endTime),
		zap.Int64("startTimeMicros", startTime.UnixMicro()),
		zap.Int64("endTimeMicros", endTime.UnixMicro()),
	)...)
	now := time.Now()
	if enableSystemHealthReportCache {
		cacheKey := systemHealthReportCacheKey(req.GroupID, startTime, endTime)
		if cached, ok := h.reportCache.get(cacheKey, now); ok {
			systemHealthInfo("system-health report cache hit", systemHealthStageFields(ctx, buildStartedAt, "02_cache_lookup",
				zap.Int64("groupId", req.GroupID),
				zap.String("cacheKey", cacheKey),
			)...)
			return cached, nil
		}
	}
	systemHealthInfo("system-health report stage started", systemHealthStageFields(ctx, buildStartedAt, "03_query_alert_events",
		zap.Int64("groupId", req.GroupID),
	)...)
	events, err := h.queryRealAlertEvents(ctx, req.GroupID, startTime, endTime)
	if err != nil {
		systemHealthWarn("system-health alert event query failed", systemHealthStageFields(ctx, buildStartedAt, "03_query_alert_events",
			zap.Int64("groupId", req.GroupID),
			zap.Time("startTime", startTime),
			zap.Time("endTime", endTime),
			zap.Error(err),
		)...)
		return response.SystemHealthReportResp{}, err
	}
	systemHealthInfo("system-health alert event query completed", systemHealthStageFields(ctx, buildStartedAt, "03_query_alert_events",
		zap.Int64("groupId", req.GroupID),
		zap.Int("eventCount", len(events)),
	)...)
	realEventCount := len(events)
	if strings.TrimSpace(h.alertManagerURL) == "" {
		systemHealthInfo("system-health alertmanager events query skipped", systemHealthStageFields(ctx, buildStartedAt, "04_merge_alertmanager",
			zap.Int64("groupId", req.GroupID),
			zap.String("reason", "not_configured"),
		)...)
	} else {
		systemHealthInfo("system-health report stage started", systemHealthStageFields(ctx, buildStartedAt, "04_merge_alertmanager",
			zap.Int64("groupId", req.GroupID),
		)...)
		if alertManagerEvents, err := h.queryAlertManagerActiveEvents(ctx, req.GroupID); err == nil {
			events = mergeAlertEvents(events, alertManagerEvents)
			systemHealthInfo("system-health alertmanager events merged", systemHealthStageFields(ctx, buildStartedAt, "04_merge_alertmanager",
				zap.Int64("groupId", req.GroupID),
				zap.Int("realEventCount", realEventCount),
				zap.Int("alertManagerEventCount", len(alertManagerEvents)),
				zap.Int("mergedEventCount", len(events)),
			)...)
		} else {
			systemHealthWarn("system-health alertmanager events query skipped", systemHealthStageFields(ctx, buildStartedAt, "04_merge_alertmanager",
				zap.Int64("groupId", req.GroupID),
				zap.Error(err),
			)...)
		}
	}

	systemHealthInfo("system-health report stage started", systemHealthStageFields(ctx, buildStartedAt, "05_query_telemetry_capabilities",
		zap.Int64("groupId", req.GroupID),
	)...)
	capabilities := h.queryTelemetryCapabilities(ctx)
	systemHealthInfo("system-health telemetry capability query completed", systemHealthStageFields(ctx, buildStartedAt, "05_query_telemetry_capabilities",
		zap.Int64("groupId", req.GroupID),
		zap.Strings("enabled", capabilities.Enabled),
		zap.Strings("disabled", capabilities.Disabled),
		zap.Strings("sources", capabilities.Sources),
		zap.String("error", capabilities.Err),
	)...)
	systemHealthInfo("system-health report stage started", systemHealthStageFields(ctx, buildStartedAt, "06_build_inputs",
		zap.Int64("groupId", req.GroupID),
		zap.Int("eventCount", len(events)),
	)...)
	inputs := h.buildInputsFromRealEvents(ctx, req.GroupID, events, startTime, endTime, capabilities)
	systemHealthInfo("system-health report inputs built", systemHealthStageFields(ctx, buildStartedAt, "06_build_inputs",
		zap.Int64("groupId", req.GroupID),
		zap.Int("inputCount", len(inputs)),
	)...)
	systemHealthInfo("system-health report stage started", systemHealthStageFields(ctx, buildStartedAt, "07_score_and_build_response",
		zap.Int64("groupId", req.GroupID),
		zap.Int("inputCount", len(inputs)),
	)...)
	scored := scoreIssues(inputs)
	healthScore := calculateHealthScore(scored)
	status := healthStatus(healthScore)
	scoreFactors := buildScoreFactors(scored)
	issues := buildIssues(scored)
	systemHealthInfo("system-health structured issues built", systemHealthStageFields(ctx, buildStartedAt, "07_score_and_build_response",
		zap.Int64("groupId", req.GroupID),
		zap.Int("issueCount", len(issues)),
		zap.Int("healthScore", healthScore),
		zap.String("healthStatus", status),
	)...)

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
	logSystemHealthIssueDiagnostics(ctx, scored, issues)
	systemHealthInfo("system-health report stage started", systemHealthStageFields(ctx, buildStartedAt, "08_ai_enrichment",
		zap.Int64("groupId", req.GroupID),
		zap.Int("issueCount", len(resp.Issues)),
	)...)
	h.enrichReportWithAI(ctx, &resp, scored)
	if enableSystemHealthReportCache {
		h.reportCache.set(systemHealthReportCacheKey(req.GroupID, startTime, endTime), resp, now)
	}
	systemHealthInfo("system-health report build completed", systemHealthStageFields(ctx, buildStartedAt, "09_build_completed",
		zap.Int64("groupId", req.GroupID),
		zap.String("status", resp.Report.Status),
		zap.Int("score", resp.Report.Score),
		zap.Int("eventCount", len(events)),
		zap.Int("issueCount", len(resp.Issues)),
		zap.String("aiStatus", resp.Report.AIAnalysis.Status),
		zap.Int("payloadBytes", systemHealthJSONSize(resp)),
	)...)
	systemHealthInfo("system-health report stage started", systemHealthStageFields(ctx, buildStartedAt, "10_save_report_record",
		zap.Int64("groupId", req.GroupID),
		zap.String("aiStatus", resp.Report.AIAnalysis.Status),
	)...)
	if err := h.saveReportRecord(ctx, req, startTime, endTime, resp); err != nil {
		systemHealthWarn("system-health report record save failed", systemHealthStageFields(ctx, buildStartedAt, "10_save_report_record",
			zap.Int64("groupId", req.GroupID),
			zap.Time("startTime", startTime),
			zap.Time("endTime", endTime),
			zap.Error(err),
		)...)
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

func (h *handler) saveReportRecord(ctx core.Context, req *request.GetSystemHealthReportRequest, startTime time.Time, endTime time.Time, resp response.SystemHealthReportResp) error {
	if h.dbRepo == nil {
		systemHealthInfoCtx(ctx, "system-health report record save skipped",
			zap.Int64("groupId", req.GroupID),
			zap.String("reason", "db_repository_not_configured"),
		)
		return nil
	}
	payload, err := json.Marshal(resp)
	if err != nil {
		systemHealthWarnCtx(ctx, "system-health report record marshal failed",
			zap.Int64("groupId", req.GroupID),
			zap.String("status", resp.Report.Status),
			zap.Int("issueCount", len(resp.Issues)),
			zap.String("aiStatus", resp.Report.AIAnalysis.Status),
			zap.Error(err),
		)
		return err
	}
	recordID := uuid.NewString()
	record := &database.SystemHealthReportRecord{
		RecordID:         recordID,
		GroupID:          req.GroupID,
		Status:           resp.Report.Status,
		Score:            resp.Report.Score,
		Summary:          truncateText(resp.Report.Summary, 1000),
		Window:           resp.Report.Window,
		StartTime:        startTime.UnixMicro(),
		EndTime:          endTime.UnixMicro(),
		GeneratedAt:      resp.Report.GeneratedAt,
		IssueCount:       len(resp.Issues),
		ActiveAlertCount: resp.Report.ActiveAlertCount,
		AIStatus:         resp.Report.AIAnalysis.Status,
		Payload:          string(payload),
		CreatedAt:        time.Now(),
	}
	saveStartedAt := time.Now()
	systemHealthInfoCtx(ctx, "system-health report record save started",
		zap.String("recordId", recordID),
		zap.Int64("groupId", req.GroupID),
		zap.String("status", resp.Report.Status),
		zap.Int("score", resp.Report.Score),
		zap.Int("issueCount", len(resp.Issues)),
		zap.Int("activeAlertCount", resp.Report.ActiveAlertCount),
		zap.String("aiStatus", resp.Report.AIAnalysis.Status),
		zap.Int("payloadBytes", len(payload)),
		zap.Int64("startTimeMicros", startTime.UnixMicro()),
		zap.Int64("endTimeMicros", endTime.UnixMicro()),
	)
	if err := h.dbRepo.SaveSystemHealthReportRecord(ctx, record); err != nil {
		systemHealthWarnCtx(ctx, "system-health report record save failed",
			zap.String("recordId", recordID),
			zap.Int64("groupId", req.GroupID),
			zap.Int("payloadBytes", len(payload)),
			zap.Duration("elapsed", time.Since(saveStartedAt)),
			zap.Error(err),
		)
		return err
	}
	systemHealthInfoCtx(ctx, "system-health report record save completed",
		zap.String("recordId", recordID),
		zap.Int64("groupId", req.GroupID),
		zap.Int("payloadBytes", len(payload)),
		zap.Duration("elapsed", time.Since(saveStartedAt)),
	)
	return nil
}

func systemHealthRecordSummaries(records []database.SystemHealthReportRecord) []response.SystemHealthReportRecordSummary {
	result := make([]response.SystemHealthReportRecordSummary, 0, len(records))
	for _, record := range records {
		result = append(result, systemHealthRecordSummary(record))
	}
	return result
}

func systemHealthRecordSummary(record database.SystemHealthReportRecord) response.SystemHealthReportRecordSummary {
	return response.SystemHealthReportRecordSummary{
		RecordID:         record.RecordID,
		GroupID:          record.GroupID,
		Status:           record.Status,
		Score:            record.Score,
		Summary:          record.Summary,
		Window:           record.Window,
		StartTime:        record.StartTime,
		EndTime:          record.EndTime,
		GeneratedAt:      record.GeneratedAt,
		IssueCount:       record.IssueCount,
		ActiveAlertCount: record.ActiveAlertCount,
		AIStatus:         record.AIStatus,
		CreatedAt:        record.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// ListReportRecords returns saved system health report snapshots.
// @Summary list system health report records
// @Description list saved system health report records
// @Tags API.systemHealth
// @Produce json
// @Param groupId query int false "data group id"
// @Param page query int false "page"
// @Param pageSize query int false "page size"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.SystemHealthReportRecordListResp
// @Failure 400 {object} code.Failure
// @Router /api/system-health/report/records [get]
func (h *handler) ListReportRecords() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.ListSystemHealthReportRecordsRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(http.StatusBadRequest, code.ParamBindError, err)
			return
		}
		if req.Page <= 0 {
			req.Page = 1
		}
		if req.PageSize <= 0 {
			req.PageSize = 20
		}
		if req.PageSize > 100 {
			req.PageSize = 100
		}
		authorizedGroupIDs, err := h.authorizedSystemHealthReportGroupIDs(c)
		if err != nil {
			c.AbortWithError(http.StatusForbidden, code.AuthError, err)
			return
		}
		groupIDs := authorizedGroupIDs
		if req.GroupID != 0 {
			if !systemHealthReportGroupAllowed(authorizedGroupIDs, req.GroupID) {
				c.AbortWithError(http.StatusForbidden, code.UserNoPermissionError, nil)
				return
			}
			groupIDs = []int64{req.GroupID}
		}
		records, total, err := h.dbRepo.ListSystemHealthReportRecords(c, groupIDs, req.Page, req.PageSize)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, code.ServerError, err)
			return
		}
		c.Payload(response.SystemHealthReportRecordListResp{
			Records:  systemHealthRecordSummaries(records),
			Total:    total,
			Page:     req.Page,
			PageSize: req.PageSize,
		})
	}
}

// GetReportRecord returns a saved system health report snapshot.
// @Summary get system health report record
// @Description get saved system health report record by id
// @Tags API.systemHealth
// @Produce json
// @Param recordId path string true "report record id"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.SystemHealthReportRecordResp
// @Failure 400 {object} code.Failure
// @Router /api/system-health/report/records/{recordId} [get]
func (h *handler) GetReportRecord() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetSystemHealthReportRecordRequest)
		if err := c.ShouldBindURI(req); err != nil {
			c.AbortWithError(http.StatusBadRequest, code.ParamBindError, err)
			return
		}
		record, err := h.dbRepo.GetSystemHealthReportRecord(c, req.RecordID)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, code.ServerError, err)
			return
		}
		authorizedGroupIDs, err := h.authorizedSystemHealthReportGroupIDs(c)
		if err != nil {
			c.AbortWithError(http.StatusForbidden, code.AuthError, err)
			return
		}
		if !systemHealthReportGroupAllowed(authorizedGroupIDs, record.GroupID) {
			c.AbortWithError(http.StatusForbidden, code.UserNoPermissionError, nil)
			return
		}
		var data response.SystemHealthReportResp
		if err := json.Unmarshal([]byte(record.Payload), &data); err != nil {
			c.AbortWithError(http.StatusBadRequest, code.ServerError, err)
			return
		}
		c.Payload(response.SystemHealthReportRecordResp{
			Record: systemHealthRecordSummary(*record),
			Data:   data,
		})
	}
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
		systemHealthInfoCtx(ctx, "system-health AI enrichment skipped",
			zap.Bool("enabled", aiConfig.Enable),
			zap.Bool("hasBaseURL", strings.TrimSpace(aiConfig.BaseURL) != ""),
			zap.Bool("hasModel", strings.TrimSpace(aiConfig.Model) != ""),
			zap.Int("issueCount", len(resp.Issues)),
		)
		return
	}
	if strings.TrimSpace(aiConfig.WireAPI) != "" && strings.TrimSpace(aiConfig.WireAPI) != "responses" {
		resp.Report.AIAnalysis.Status = "failed"
		resp.Report.AIAnalysis.Error = "unsupported AI wire API: " + aiConfig.WireAPI
		systemHealthWarnCtx(ctx, "system-health AI enrichment failed before request",
			zap.String("reason", "unsupported_wire_api"),
			zap.String("wireAPI", aiConfig.WireAPI),
		)
		return
	}
	if len(resp.Issues) == 0 {
		resp.Report.AIAnalysis.Status = "skipped"
		systemHealthInfoCtx(ctx, "system-health AI enrichment skipped",
			zap.String("reason", "no_issues"),
		)
		return
	}

	aiRequest := buildSystemHealthAIRequest(resp, scored)
	effectiveTimeoutSecond := effectiveSystemHealthAITimeoutSecond(aiConfig.TimeoutSecond)
	aiStartedAt := time.Now()
	systemHealthInfoCtx(ctx, "system-health AI enrichment started",
		zap.String("endpoint", endpoint),
		zap.String("model", aiConfig.Model),
		zap.Int("issueCount", len(resp.Issues)),
		zap.Int("aiIssueCount", len(aiRequest.Issues)),
		zap.Strings("requestIssueKeys", systemHealthAIIssueInputKeys(aiRequest.Issues)),
		zap.Int("requestPayloadBytes", systemHealthJSONSize(aiRequest)),
		zap.Int("configuredTimeoutSecond", aiConfig.TimeoutSecond),
		zap.Int("effectiveTimeoutSecond", effectiveTimeoutSecond),
	)
	aiResp, aiStats, err := callSystemHealthAIEnrichment(ctx, aiConfig.BaseURL, resolveSystemHealthAIAPIKey(aiConfig.APIKey, aiConfig.EnvKey), aiConfig.Model, aiConfig.TimeoutSecond, aiRequest)
	if err != nil {
		resp.Report.AIAnalysis.Status = "failed"
		resp.Report.AIAnalysis.Error = truncateText(err.Error(), 220)
		systemHealthWarnCtx(ctx, "system-health AI enrichment failed",
			zap.String("endpoint", endpoint),
			zap.String("aiMode", aiStats.Mode),
			zap.Int("issueCount", len(resp.Issues)),
			zap.Strings("requestIssueKeys", systemHealthAIIssueInputKeys(aiRequest.Issues)),
			zap.Strings("failedIssueKeys", aiStats.FailedIssueKeys),
			zap.Strings("missingIssueKeys", aiStats.MissingIssueKeys),
			zap.Int("configuredTimeoutSecond", aiConfig.TimeoutSecond),
			zap.Int("effectiveTimeoutSecond", effectiveTimeoutSecond),
			zap.Bool("timeout", isTimeoutError(err)),
			zap.Duration("elapsed", time.Since(aiStartedAt)),
			zap.Error(err),
		)
		return
	}
	resp.Report.AIAnalysis.Status = systemHealthAIAnalysisStatus(aiStats, len(aiRequest.Issues))
	resp.Report.AIAnalysis.Source = "ai-api"
	if aiStats.Mode == "parallel_per_issue" {
		resp.Report.AIAnalysis.Source = "ai-api-parallel"
	}
	resp.Report.AIAnalysis.WorkflowRunID = strings.TrimSpace(aiResp.WorkflowRunID)
	if resp.Report.AIAnalysis.Status == "partial" {
		resp.Report.AIAnalysis.Error = truncateText(systemHealthAIStatsError(aiStats), 220)
	}
	systemHealthInfoCtx(ctx, "system-health AI enrichment completed",
		zap.String("endpoint", endpoint),
		zap.String("aiMode", aiStats.Mode),
		zap.String("aiStatus", resp.Report.AIAnalysis.Status),
		zap.Int("issueCount", len(resp.Issues)),
		zap.Int("aiIssueCount", len(aiResp.Issues)),
		zap.Int("aiTotalCalls", aiStats.TotalCalls),
		zap.Int("aiSucceededCalls", aiStats.SucceededCalls),
		zap.Int("aiFailedCalls", aiStats.FailedCalls),
		zap.Int("aiSucceededIssues", aiStats.SucceededIssues),
		zap.Strings("requestIssueKeys", systemHealthAIIssueInputKeys(aiRequest.Issues)),
		zap.Strings("responseIssueKeys", systemHealthAIIssueOutputKeys(aiResp.Issues)),
		zap.Strings("failedIssueKeys", aiStats.FailedIssueKeys),
		zap.Strings("missingIssueKeys", aiStats.MissingIssueKeys),
		zap.Bool("summaryPresent", strings.TrimSpace(aiResp.Summary) != ""),
		zap.String("workflowRunID", strings.TrimSpace(aiResp.WorkflowRunID)),
		zap.Duration("elapsed", time.Since(aiStartedAt)),
	)
	if summary := cleanUserFacingReportText(aiResp.Summary); summary != "" {
		resp.Report.Summary = summary
	}

	issueOutputByKey := make(map[string]systemHealthAIIssueOutput, len(aiResp.Issues))
	for _, issue := range aiResp.Issues {
		if strings.TrimSpace(issue.Key) == "" {
			continue
		}
		issueOutputByKey[issue.Key] = issue
	}
	systemHealthInfoCtx(ctx, "system-health AI issue output mapping completed",
		zap.Int("requestIssueCount", len(aiRequest.Issues)),
		zap.Int("responseIssueCount", len(issueOutputByKey)),
		zap.Strings("missingIssueKeys", missingAIIssueKeys(aiRequest.Issues, issueOutputByKey)),
	)
	for index := range resp.Issues {
		issueOutput, ok := issueOutputByKey[resp.Issues[index].Key]
		if !ok {
			systemHealthWarnCtx(ctx, "system-health AI issue output missing, keep structured fallback",
				zap.String("issueKey", resp.Issues[index].Key),
				zap.String("aggregationKey", resp.Issues[index].AggregationKey),
				zap.String("title", resp.Issues[index].Title),
			)
			continue
		}
		var reportActionPlan []response.SystemHealthActionPlanStep
		if index < len(scored) {
			reportActionPlan = alertReportActionPlanSteps(scored[index].input.alertReportActions, resp.Issues[index].Scope)
		}
		title := cleanUserFacingReportText(issueOutput.Title)
		analysis := cleanUserFacingReportText(issueOutput.Analysis)
		rootCause := cleanUserFacingReportText(issueOutput.RootCause)
		metricAnalysis := cleanUserFacingReportText(issueOutput.MetricAnalysis)
		systemHealthInfoCtx(ctx, "system-health AI issue output received",
			zap.String("issueKey", resp.Issues[index].Key),
			zap.String("aggregationKey", resp.Issues[index].AggregationKey),
			zap.Bool("hasTitle", title != ""),
			zap.Bool("hasAnalysis", analysis != ""),
			zap.Bool("hasRootCause", rootCause != ""),
			zap.Bool("hasMetricAnalysis", metricAnalysis != ""),
			zap.Int("aiActionPlanSteps", len(issueOutput.ActionPlan)),
			zap.Int("aiActions", len(issueOutput.Actions)),
			zap.Int("reportActionPlanSteps", len(reportActionPlan)),
		)
		if title != "" && isAcceptableAIIssueTitle(title) {
			resp.Issues[index].Title = title
		} else if title != "" {
			systemHealthWarnCtx(ctx, "system-health issue title fallback kept",
				zap.String("issueKey", resp.Issues[index].Key),
				zap.String("reason", "ai_title_unacceptable"),
				zap.String("aiTitle", truncateText(title, 120)),
			)
		}
		if analysis != "" {
			resp.Issues[index].AIConclusion = analysis
		} else {
			systemHealthWarnCtx(ctx, "system-health issue analysis fallback kept",
				zap.String("issueKey", resp.Issues[index].Key),
				zap.String("reason", "ai_analysis_empty"),
			)
		}
		if index < len(scored) {
			finalRootCause, rootCauseSource, rootCauseReason := resolveAIEnrichedRootCause(rootCause, scored[index])
			if finalRootCause != "" {
				resp.Issues[index].RootCause = finalRootCause
				systemHealthInfoCtx(ctx, "system-health issue rootCause resolved",
					zap.String("issueKey", resp.Issues[index].Key),
					zap.String("aggregationKey", resp.Issues[index].AggregationKey),
					zap.String("source", rootCauseSource),
					zap.String("reason", rootCauseReason),
					zap.Bool("hasAIRootCause", rootCause != ""),
					zap.Int("alertReportOverviewDetailCount", alertReportOverviewDetailCount(scored[index].input.alertReportFindings)),
					zap.Int("alertReportFindings", len(scored[index].input.alertReportFindings)),
				)
			} else {
				systemHealthWarnCtx(ctx, "system-health issue rootCause fallback kept",
					zap.String("issueKey", resp.Issues[index].Key),
					zap.String("reason", rootCauseReason),
				)
			}
		} else if rootCause != "" {
			resp.Issues[index].RootCause = rootCause
			systemHealthInfoCtx(ctx, "system-health issue rootCause resolved",
				zap.String("issueKey", resp.Issues[index].Key),
				zap.String("aggregationKey", resp.Issues[index].AggregationKey),
				zap.String("source", "ai"),
				zap.String("reason", "scored_context_missing"),
				zap.Bool("hasAIRootCause", true),
			)
		} else {
			systemHealthWarnCtx(ctx, "system-health issue rootCause fallback kept",
				zap.String("issueKey", resp.Issues[index].Key),
				zap.String("reason", "ai_root_cause_empty_and_scored_context_missing"),
			)
		}
		if metricAnalysis != "" && !strings.ContainsRune(metricAnalysis, '\uFFFD') && !containsMetricQuerySyntax(metricAnalysis) {
			resp.Issues[index].MetricAnalysis = metricAnalysis
		} else if metricAnalysis == "" {
			systemHealthWarnCtx(ctx, "system-health issue metricAnalysis fallback kept",
				zap.String("issueKey", resp.Issues[index].Key),
				zap.String("reason", "ai_metric_analysis_empty"),
			)
		} else if containsMetricQuerySyntax(metricAnalysis) {
			systemHealthWarnCtx(ctx, "system-health issue metricAnalysis fallback kept",
				zap.String("issueKey", resp.Issues[index].Key),
				zap.String("reason", "ai_metric_analysis_contains_query_expression"),
				zap.String("metricAnalysisPreview", truncateText(metricAnalysis, 220)),
			)
		} else {
			systemHealthWarnCtx(ctx, "system-health issue metricAnalysis fallback kept",
				zap.String("issueKey", resp.Issues[index].Key),
				zap.String("reason", "ai_metric_analysis_invalid_text"),
			)
		}
		actionPlan, actionPlanMode := resolveAIEnrichedActionPlan(issueOutput, reportActionPlan, resp.Issues[index].ActionPlan, resp.Issues[index])
		if len(actionPlan) > 0 {
			resp.Issues[index].ActionPlan = actionPlan
			resp.Issues[index].Actions = actionTexts(resp.Issues[index].ActionPlan)
		}
		systemHealthInfoCtx(ctx, "system-health issue actionPlan resolved",
			zap.String("issueKey", resp.Issues[index].Key),
			zap.String("aggregationKey", resp.Issues[index].AggregationKey),
			zap.String("actionPlanMode", actionPlanMode),
			zap.Int("finalActionPlanSteps", len(resp.Issues[index].ActionPlan)),
			zap.Int("aiActionPlanSteps", len(issueOutput.ActionPlan)),
			zap.Int("aiActions", len(issueOutput.Actions)),
			zap.Int("reportActionPlanSteps", len(reportActionPlan)),
		)
	}
}

func callSystemHealthAIEnrichment(ctx core.Context, baseURL string, apiKey string, model string, timeoutSecond int, payload systemHealthAISummaryRequest) (systemHealthAISummaryResponse, systemHealthAIEnrichmentStats, error) {
	if len(payload.Issues) <= 1 {
		stats := systemHealthAIEnrichmentStats{
			Mode:       "single_request",
			TotalCalls: 1,
		}
		aiResp, err := callSystemHealthAI(ctx, baseURL, apiKey, model, timeoutSecond, payload)
		if err != nil {
			stats.FailedCalls = 1
			stats.ErrorMessages = append(stats.ErrorMessages, err.Error())
			stats.FailedIssueKeys = systemHealthAIIssueInputKeys(payload.Issues)
			return systemHealthAISummaryResponse{}, stats, err
		}
		stats.SucceededCalls = 1
		stats.SummarySucceeded = strings.TrimSpace(aiResp.Summary) != ""
		outputByKey := systemHealthAIIssueOutputMap(aiResp.Issues)
		stats.MissingIssueKeys = missingAIIssueKeys(payload.Issues, outputByKey)
		stats.SucceededIssues = len(payload.Issues) - len(stats.MissingIssueKeys)
		return aiResp, stats, nil
	}
	return callSystemHealthAIForIssuesParallel(ctx, baseURL, apiKey, model, timeoutSecond, payload)
}

func callSystemHealthAIForIssuesParallel(ctx core.Context, baseURL string, apiKey string, model string, timeoutSecond int, payload systemHealthAISummaryRequest) (systemHealthAISummaryResponse, systemHealthAIEnrichmentStats, error) {
	stats := systemHealthAIEnrichmentStats{
		Mode:       "parallel_per_issue",
		TotalCalls: len(payload.Issues) + 1,
	}
	resultCh := make(chan systemHealthAICallResult, len(payload.Issues)+1)
	sem := make(chan struct{}, systemHealthAIParallelism)
	var wg sync.WaitGroup
	logCtx := systemHealthAIRequestLogContext(ctx)
	startCall := func(kind string, issueIndex int, issueKey string, request systemHealthAISummaryRequest) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			subRequestID := systemHealthAISubRequestID(kind, issueIndex, issueKey)
			startedAt := time.Now()
			systemHealthInfoCtx(logCtx, "system-health AI subrequest started",
				zap.String("aiMode", "parallel_per_issue"),
				zap.String("subRequestKind", kind),
				zap.String("aiSubRequestID", subRequestID),
				zap.Int("issueIndex", issueIndex),
				zap.String("issueKey", issueKey),
				zap.Strings("requestIssueKeys", systemHealthAIIssueInputKeys(request.Issues)),
				zap.Int("requestPayloadBytes", systemHealthJSONSize(request)),
			)
			aiResp, err := callSystemHealthAI(logCtx, baseURL, apiKey, model, timeoutSecond, request)
			resultCh <- systemHealthAICallResult{
				Kind:         kind,
				IssueIndex:   issueIndex,
				IssueKey:     issueKey,
				SubRequestID: subRequestID,
				Response:     aiResp,
				Error:        err,
				Elapsed:      time.Since(startedAt),
			}
		}()
	}

	startCall("summary", -1, "", systemHealthAISummaryOnlyRequest(payload))
	for index, issue := range payload.Issues {
		startCall("issue", index, issue.Key, systemHealthAIIssueOnlyRequest(payload, issue))
	}
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	var merged systemHealthAISummaryResponse
	issueOutputs := make([]systemHealthAIIssueOutput, len(payload.Issues))
	issueOutputPresent := make([]bool, len(payload.Issues))
	workflowRunIDs := make([]string, 0, len(payload.Issues)+1)
	for result := range resultCh {
		if result.Error != nil {
			stats.FailedCalls++
			if result.Kind == "summary" {
				stats.SummaryFailed = true
			} else if result.IssueKey != "" {
				stats.FailedIssueKeys = append(stats.FailedIssueKeys, result.IssueKey)
			}
			stats.ErrorMessages = append(stats.ErrorMessages, systemHealthAISubRequestError(result))
			systemHealthWarnCtx(logCtx, "system-health AI subrequest failed",
				zap.String("aiMode", "parallel_per_issue"),
				zap.String("subRequestKind", result.Kind),
				zap.String("aiSubRequestID", result.SubRequestID),
				zap.Int("issueIndex", result.IssueIndex),
				zap.String("issueKey", result.IssueKey),
				zap.Bool("timeout", isTimeoutError(result.Error)),
				zap.Duration("elapsed", result.Elapsed),
				zap.Error(result.Error),
			)
			continue
		}
		stats.SucceededCalls++
		if workflowRunID := strings.TrimSpace(result.Response.WorkflowRunID); workflowRunID != "" {
			workflowRunIDs = append(workflowRunIDs, workflowRunID)
		}
		systemHealthInfoCtx(logCtx, "system-health AI subrequest completed",
			zap.String("aiMode", "parallel_per_issue"),
			zap.String("subRequestKind", result.Kind),
			zap.String("aiSubRequestID", result.SubRequestID),
			zap.Int("issueIndex", result.IssueIndex),
			zap.String("issueKey", result.IssueKey),
			zap.String("workflowRunID", strings.TrimSpace(result.Response.WorkflowRunID)),
			zap.Int("responseIssueCount", len(result.Response.Issues)),
			zap.Bool("summaryPresent", strings.TrimSpace(result.Response.Summary) != ""),
			zap.Duration("elapsed", result.Elapsed),
		)
		if result.Kind == "summary" {
			stats.SummarySucceeded = strings.TrimSpace(result.Response.Summary) != ""
			merged.Summary = result.Response.Summary
			continue
		}
		output, ok := selectSystemHealthAIIssueOutput(result.Response.Issues, result.IssueKey)
		if !ok {
			stats.MissingIssueKeys = append(stats.MissingIssueKeys, result.IssueKey)
			stats.ErrorMessages = append(stats.ErrorMessages, fmt.Sprintf("%s: AI response did not include matching issue output", result.IssueKey))
			systemHealthWarnCtx(logCtx, "system-health AI issue subrequest missing output",
				zap.String("aiMode", "parallel_per_issue"),
				zap.String("aiSubRequestID", result.SubRequestID),
				zap.Int("issueIndex", result.IssueIndex),
				zap.String("issueKey", result.IssueKey),
				zap.Strings("responseIssueKeys", systemHealthAIIssueOutputKeys(result.Response.Issues)),
			)
			continue
		}
		if strings.TrimSpace(output.Key) == "" {
			output.Key = result.IssueKey
		}
		if result.IssueIndex >= 0 && result.IssueIndex < len(issueOutputs) {
			issueOutputs[result.IssueIndex] = output
			issueOutputPresent[result.IssueIndex] = true
			stats.SucceededIssues++
		}
	}
	for index, output := range issueOutputs {
		if issueOutputPresent[index] {
			merged.Issues = append(merged.Issues, output)
		}
	}
	merged.WorkflowRunID = truncateText(strings.Join(uniqueStrings(workflowRunIDs), ","), 220)
	stats.FailedIssueKeys = uniqueStrings(stats.FailedIssueKeys)
	stats.MissingIssueKeys = uniqueStrings(stats.MissingIssueKeys)
	stats.ErrorMessages = uniqueStrings(stats.ErrorMessages)
	if stats.SucceededIssues == 0 {
		err := fmt.Errorf("all AI issue subrequests failed or returned no matching issue output: %s", systemHealthAIStatsError(stats))
		return merged, stats, err
	}
	return merged, stats, nil
}

func systemHealthAISummaryOnlyRequest(payload systemHealthAISummaryRequest) systemHealthAISummaryRequest {
	payload.Issues = []systemHealthAIIssueInput{}
	return payload
}

func systemHealthAIIssueOnlyRequest(payload systemHealthAISummaryRequest, issue systemHealthAIIssueInput) systemHealthAISummaryRequest {
	payload.Issues = []systemHealthAIIssueInput{issue}
	return payload
}

func selectSystemHealthAIIssueOutput(outputs []systemHealthAIIssueOutput, expectedKey string) (systemHealthAIIssueOutput, bool) {
	expectedKey = strings.TrimSpace(expectedKey)
	for _, output := range outputs {
		if strings.TrimSpace(output.Key) == expectedKey {
			return output, true
		}
	}
	if len(outputs) == 1 && strings.TrimSpace(outputs[0].Key) == "" && expectedKey != "" {
		output := outputs[0]
		output.Key = expectedKey
		return output, true
	}
	return systemHealthAIIssueOutput{}, false
}

func systemHealthAIIssueOutputMap(outputs []systemHealthAIIssueOutput) map[string]systemHealthAIIssueOutput {
	result := make(map[string]systemHealthAIIssueOutput, len(outputs))
	for _, output := range outputs {
		key := strings.TrimSpace(output.Key)
		if key == "" {
			continue
		}
		result[key] = output
	}
	return result
}

func systemHealthAISubRequestID(kind string, issueIndex int, issueKey string) string {
	return systemHealthShortHash(fmt.Sprintf("%s:%d:%s", strings.TrimSpace(kind), issueIndex, strings.TrimSpace(issueKey)))
}

func systemHealthAISubRequestError(result systemHealthAICallResult) string {
	if result.Error == nil {
		return ""
	}
	prefix := result.Kind
	if result.IssueKey != "" {
		prefix = result.IssueKey
	}
	return prefix + ": " + result.Error.Error()
}

func systemHealthAIAnalysisStatus(stats systemHealthAIEnrichmentStats, totalIssueCount int) string {
	if stats.SucceededIssues <= 0 {
		return "failed"
	}
	if stats.FailedCalls > 0 || stats.SummaryFailed || !stats.SummarySucceeded || len(stats.FailedIssueKeys) > 0 || len(stats.MissingIssueKeys) > 0 || stats.SucceededIssues < totalIssueCount {
		return "partial"
	}
	return "succeeded"
}

func systemHealthAIStatsError(stats systemHealthAIEnrichmentStats) string {
	parts := make([]string, 0, 4+len(stats.ErrorMessages))
	if stats.SummaryFailed {
		parts = append(parts, "全局摘要生成失败")
	}
	if len(stats.FailedIssueKeys) > 0 {
		parts = append(parts, "失败风险="+strings.Join(stats.FailedIssueKeys, ","))
	}
	if len(stats.MissingIssueKeys) > 0 {
		parts = append(parts, "缺失风险="+strings.Join(stats.MissingIssueKeys, ","))
	}
	parts = append(parts, stats.ErrorMessages...)
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(uniqueStrings(parts), "；")
}

func resolveAIEnrichedRootCause(aiRootCause string, item scoredIssue) (string, string, string) {
	aiRootCause = cleanAlertReportDirectionalText(aiRootCause)
	structuredRootCause, structuredSource, structuredReason := buildRootCauseDecision(item)
	if item.input.kind == issueKindInvalidAlert {
		if structuredRootCause != "" {
			return structuredRootCause, structuredSource, "invalid_alert_rule_diagnostic_overrides_ai"
		}
		if aiRootCause != "" {
			return aiRootCause, "ai", "invalid_alert_structured_root_cause_empty"
		}
		return "", "empty", "ai_root_cause_empty_and_structured_root_cause_empty"
	}
	if aiRootCause != "" {
		if reasonCount := alertReportOverviewReasonCount(item.input.alertReportFindings); reasonCount > 0 {
			return ensureEBPFRootCauseContext(aiRootCause), "ai_alert_report_overview_reason", fmt.Sprintf("overview_reason_findings=%d", reasonCount)
		}
		if detailCount := alertReportOverviewDetailCount(item.input.alertReportFindings); detailCount > 0 {
			return aiRootCause, "ai_alert_report_overview_detail", fmt.Sprintf("overview_detail_findings=%d", detailCount)
		}
		if findingCount := len(uniqueAlertReportFindings(item.input.alertReportFindings)); findingCount > 0 {
			return aiRootCause, "ai_alert_report", fmt.Sprintf("alert_report_findings=%d", findingCount)
		}
		return aiRootCause, "ai", "ai_root_cause_present"
	}
	if structuredRootCause != "" {
		return structuredRootCause, structuredSource, "ai_root_cause_empty_" + structuredReason
	}
	return "", "empty", "ai_root_cause_empty_and_structured_root_cause_empty"
}

func isAcceptableAIIssueTitle(title string) bool {
	title = strings.TrimSpace(title)
	if title == "" {
		return false
	}
	if strings.ContainsAny(title, "\r\n\t") {
		return false
	}
	runeCount := utf8.RuneCountInString(title)
	if runeCount < 4 || runeCount > 48 {
		return false
	}
	if containsMetricQuerySyntax(title) {
		return false
	}
	return !containsIssueTitleStateWording(title)
}

func containsIssueTitleStateWording(title string) bool {
	normalized := strings.ToLower(strings.TrimSpace(title))
	if normalized == "" {
		return false
	}
	normalized = strings.TrimSpace(strings.TrimPrefix(normalized, "已恢复："))
	normalized = strings.TrimSpace(strings.TrimPrefix(normalized, "已恢复:"))
	stateWords := []string{
		"恢复观察",
		"已恢复",
		"未恢复",
		"仍在触发",
		"触发中",
		"当前仍",
		"当前已",
		"firing",
		"resolved",
		"recovering",
	}
	for _, word := range stateWords {
		if strings.Contains(normalized, strings.ToLower(word)) {
			return true
		}
	}
	return false
}

func resolveAIEnrichedActionPlan(issueOutput systemHealthAIIssueOutput, reportActionPlan []response.SystemHealthActionPlanStep, currentPlan []response.SystemHealthActionPlanStep, issue response.SystemHealthIssue) ([]response.SystemHealthActionPlanStep, string) {
	if len(issueOutput.ActionPlan) > 0 {
		aiPlan := mergeActionPlanSteps(normalizeAIActionPlanTargets(markActionPlanSource(issueOutput.ActionPlan, "ai"), issue), nil)
		if len(aiPlan) > 0 {
			if len(reportActionPlan) > 0 {
				return aiPlan, "ai_synthesized_from_alert_report_actions"
			}
			return aiPlan, "ai_generated"
		}
	}
	if len(issueOutput.Actions) > 0 {
		aiPlan := mergeActionPlanSteps(actionsToActionPlanSteps(issueOutput.Actions), nil)
		if len(aiPlan) > 0 {
			if len(reportActionPlan) > 0 {
				return aiPlan, "ai_actions_synthesized_from_alert_report_actions"
			}
			return aiPlan, "ai_actions"
		}
	}
	if len(currentPlan) > 0 {
		if len(reportActionPlan) > 0 {
			return currentPlan, "alert_report_structured_fallback"
		}
		return currentPlan, "structured_fallback"
	}
	return nil, "empty"
}

func callSystemHealthAI(ctx core.Context, baseURL string, apiKey string, model string, timeoutSecond int, payload systemHealthAISummaryRequest) (systemHealthAISummaryResponse, error) {
	timeoutSecond = effectiveSystemHealthAITimeoutSecond(timeoutSecond)
	if strings.TrimSpace(apiKey) == "" {
		systemHealthWarnCtx(ctx, "system-health AI request skipped before HTTP",
			zap.String("reason", "missing_api_key"),
			zap.String("endpoint", openAIResponsesEndpoint(baseURL)),
			zap.String("model", model),
		)
		return systemHealthAISummaryResponse{}, fmt.Errorf("missing AI API key")
	}
	inputPayload, err := json.Marshal(payload)
	if err != nil {
		systemHealthWarnCtx(ctx, "system-health AI request payload marshal failed",
			zap.String("endpoint", openAIResponsesEndpoint(baseURL)),
			zap.String("model", model),
			zap.Error(err),
		)
		return systemHealthAISummaryResponse{}, err
	}
	requestBody, err := json.Marshal(openAIResponsesRequest{
		Model: strings.TrimSpace(model),
		Instructions: strings.Join([]string{
			"你是系统健康中心的告警汇总助手。",
			"任务：把当前窗口的系统健康状态、告警事件、已有分析结论和证据，整理成普通用户可理解、可处置的系统健康报告。",
			"全局原则：只能基于输入中的真实监控数据、已有告警分析产出和证据生成内容；不得编造输入中不存在的服务、接口、实例、根因、指标趋势或处置动作。",
			"不要输出置信度，不要解释告警为何被聚合。",
			"输出必须严格符合请求中的 json_schema。",
			"issues 只返回输入中已有的 issue.key，不能新增、改写或翻译 key。",
			"各字段职责必须清晰：summary 用于历史列表或空状态；issue.title 用于风险卡片标题；issue.analysis 用于风险卡片主结论；issue.metricAnalysis 用于指标变化依据；issue.rootCause 用于根因判断依据；issue.actionPlan 用于处置建议。",
			"避免在 analysis、metricAnalysis、rootCause 和 actionPlan 中重复同一段结论。",
			"summary 控制在 60-100 个中文字符，最多 2 句话；只说明系统状态、告警概况和主要关注方向。",
			"summary 使用“多入口/多服务/单节点/多实例/单目标/多目标”等范围模式描述影响面；不要出现具体服务名、接口名、实例名或目标清单。",
			"summary 不写证据链、根因详情或处置步骤；不要使用“整体判断”“风险模式”等内部分析口径；不要写成“主要问题集中在 A、B、C”这类列表式表达。",
			"issue.title 是风险名称，不是分析结论；使用输入中的真实服务名、接口名、scope 和 alertNames，生成 12-32 个中文字符左右的风险事件名称。",
			"issue.title 优先使用“{对象} 存在{问题类型}问题”或“{对象} {指标/请求}上升”这类风格，例如“ts-station-service 存在请求延时问题”“ts-station-service 请求延时上升”。",
			"如果风险已恢复，issue.title 可以使用状态前缀“已恢复：{对象} 存在{问题类型}问题”；不要把状态词放在标题末尾。",
			"当 alertNames 包含多条告警时，issue.title 应归纳共同症状，例如延迟、错误率、资源、可用性或规则口径；保留最关键服务名或 scope，不要逐条罗列所有告警名。",
			"issue.title 不要包含“恢复观察”“当前仍在触发”“触发中”“未恢复”等状态词；除开头的“已恢复：”前缀外，状态只能写在 issue.analysis 中。",
			"issue.title 不要返回“xxx 等 N 类告警”这种机械标题；不要根据服务名推断业务功能、行业含义或用户动作。",
			"如果无法生成比输入 title 更清晰的风险事件名称，issue.title 返回原 title。",
			"issue.analysis 只说明当前风险是什么、影响什么对象、当前仍在触发还是已恢复观察；可以简要说明故障方向指向入口、下游、资源、指标退化或规则口径。",
			"issue.analysis 不复述 metricAnalysis 的指标细节，不复述 rootCause 的完整根因判断，不罗列 evidence 中的数字，控制在 80-150 个中文字符。",
			"issue.metricAnalysis 只解释告警指标趋势和证据边界，必须使用 evidence.metric.metricName 作为指标名。",
			"issue.metricAnalysis 只能使用“告警指标趋势”证据，不要引用“请求北极星”的请求平均耗时、TPS 或资源分项；请求北极星只能用于 rootCause 或 analysis 的故障方向说明。",
			"issue.metricAnalysis 说明服务/请求路径、时间段、指标表现、趋势意义和当前风险含义；如果样本不足、未匹配到目标序列或只能作为背景参考，必须明确说明。",
			"evidence.metric.ruleName 是告警规则名，不是指标名；不要输出 PQL、PromQL、rate(、sum by(、query 表达式或其它调试查询内容；不要只复述“命中几条指标、变化百分比、区间”。",
			"如果 evidence 中没有告警指标趋势或无法形成趋势，issue.metricAnalysis 输出空字符串；否则控制在 40-100 个中文字符。",
			"issue.rootCause 来源优先级：1 existingReportFindings 中的 rootCause、overviewReason、overviewDetail、rootCauseOther、reportEvidence；2 请求北极星证据；3 existingAlertAnalyses 中一致的故障方向；4 告警事件规律和指标证据。",
			"issue.rootCause 只输出根因判断，不写处置建议；如果高优先级来源已有明确方向，不能被低优先级来源推翻。",
			"existingReportFindings 是最高优先级来源，应优先总结原告警分析报告的故障方向和关键详情；缺失数据只能作为验证边界，不能作为 rootCause 主体。",
			"overviewReason 来自基于 eBPF 数据的故障方向分析，使用时要体现“基于 eBPF 数据/方向报告分析”的来源；rootCauseOther 需要和 rootCause、overviewReason、overviewDetail 一起归纳。",
			"如果报告内容表示“无明确故障方向、未定位到明确根因、无法判断故障方向”，不要将其作为 rootCause 输出。",
			"日志分析只用于明确错误信号：只有出现错误、异常、超时、失败、慢请求、5xx/4xx、panic、exception 等信号时才能写入 analysis、rootCause 或 actionPlan。",
			"不要对正常日志、INFO 为主、无明显报错、未持续新增异常、恢复后无异常堆积等内容做评价或推测；这类日志不能作为根因、风险或恢复质量结论。",
			"不要从正常业务日志中推断问题：请求高频、批量查询、返回成功、业务字段取值、默认值或可疑业务语义，都不能作为日志异常、根因或处置依据。",
			"如果日志只显示业务请求过程、成功返回、常规 INFO 或未见异常，不能输出“从日志中可以看出”这类日志分析结论；只在明确错误日志存在时描述具体错误信号。",
			"日志分析涉及业务内容时必须脱敏：不要输出车次、订单号、用户号、手机号、证件号、库存/座位明细、token、session、请求参数原值或业务字段原值；只保留错误类型、服务/组件、时间窗口和影响方向。",
			"existingAlertAnalyses 只作为辅助信息，不要原样罗列；如果所有来源都没有明确方向，输出“现有证据暂不能归因，需要结合告警指标或明确错误日志继续确认。”",
			"issue.rootCause 控制在 40-120 个中文字符。",
			"当 evidence 包含“请求北极星”时，rootCause 必须读取网络耗时、CPU 执行耗时、调度等待耗时、TPS 变化和慢调用样本线索。",
			"请求北极星表示单次请求平均耗时分布，不是线程级北极星；不要写成线程、CPU 火焰图或线程等待结论。",
			"如果请求北极星显示网络耗时、CPU 执行耗时或调度等待耗时在告警窗口异常升高，应作为延时根因方向之一；如果整体趋势无明显异常但存在超过绝对阈值的慢调用样本，只能建议下钻这些样本，不能直接判定整体资源耗时异常。",
			"P4 invalidAlert 表示告警持续触发但已有分析反复判定无效，应按告警规则诊断处理，分析指标、表达式、阈值、标签口径或同比基线问题，不能当作服务故障处置。",
			"P4 invalidAlert 的 rootCause 应指向规则口径或指标口径问题；actionPlan 应聚焦规则诊断、RED 指标对比和告警规则修正；除非输入证据明确支持，不要输出扩容、重启等业务止血动作。",
			"actionPlan 生成 2-4 个顺序合理、互相呼应、可执行的步骤。",
			"actionPlan 不是从 0 开始排查；必须承接 issue.analysis、issue.metricAnalysis、issue.rootCause、existingReportFindings、existingAlertAnalyses 和 evidence 中已经形成的判断。",
			"每个 actionPlan 步骤只能做两类事情：沿着已有分析继续定位尚未闭合的问题，或基于已有根因/指标/日志/链路结论执行缓解、修复、验证和复发观察。",
			"不要输出泛化起手式排查，例如“检查服务状态、查看日志、查看监控、确认告警详情”；除非已有分析明确指向这些对象，并且 action 写清楚要验证的具体信号。",
			"existingReportActions 是参考来源，不能原样复制或机械追加；应结合当前健康状态、告警是否仍触发、指标和明确错误日志证据重新组织。",
			"保留仍适用的排查目标和验证要点，删除与当前状态矛盾、过期或不可执行的动作；如果原报告建议偏单次告警事件，需要上升到当前风险对象。",
			"actionPlan 优先处理仍在触发或用户影响最大的方向，再做根因确认、恢复验证和复发观察。",
			"不要输出泛泛的“持续关注”，必须说明关注对象、验证信号和风险；target 不得使用 issue.key、aggregationKey 或 activeAlert/recoveredAlert/invalidAlert 这类内部标识。",
			"actionPlan.phase 控制在 4-12 个中文字符；target 控制在 4-40 个中文字符；action 控制在 30-80 个中文字符；validation 和 risk 控制在 20-70 个中文字符。",
			"actionPlan.validation 必须使用“确认/检查/观察/回看/对比”等动作表达；不要写成“日志无异常”“指标已恢复”这类已经成立的事实结论。",
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
		systemHealthWarnCtx(ctx, "system-health AI HTTP request marshal failed",
			zap.String("endpoint", openAIResponsesEndpoint(baseURL)),
			zap.String("model", model),
			zap.Int("inputPayloadBytes", len(inputPayload)),
			zap.Error(err),
		)
		return systemHealthAISummaryResponse{}, err
	}
	systemHealthInfoCtx(ctx, "system-health AI raw request payload",
		zap.String("endpoint", openAIResponsesEndpoint(baseURL)),
		zap.String("model", model),
		zap.Int("inputPayloadBytes", len(inputPayload)),
		zap.Int("requestBodyBytes", len(requestBody)),
		zap.String("rawInputPayload", string(inputPayload)),
		zap.String("rawRequestBody", string(requestBody)),
	)
	requestContext := context.Background()
	if ctx != nil {
		requestContext = ctx.GetContext()
	}
	requestContext, cancel := context.WithTimeout(requestContext, time.Duration(timeoutSecond)*time.Second)
	defer cancel()

	endpoint := openAIResponsesEndpoint(baseURL)
	req, err := http.NewRequestWithContext(requestContext, http.MethodPost, endpoint, bytes.NewReader(requestBody))
	if err != nil {
		systemHealthWarnCtx(ctx, "system-health AI HTTP request build failed",
			zap.String("endpoint", endpoint),
			zap.String("model", model),
			zap.Int("inputPayloadBytes", len(inputPayload)),
			zap.Int("requestBodyBytes", len(requestBody)),
			zap.Error(err),
		)
		return systemHealthAISummaryResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apiKey))

	callStartedAt := time.Now()
	systemHealthInfoCtx(ctx, "system-health AI HTTP request started",
		zap.String("endpoint", endpoint),
		zap.String("model", model),
		zap.Int("inputPayloadBytes", len(inputPayload)),
		zap.Int("requestBodyBytes", len(requestBody)),
		zap.Int("timeoutSecond", timeoutSecond),
		zap.Strings("requestIssueKeys", systemHealthAIIssueInputKeys(payload.Issues)),
	)
	resp, err := (&http.Client{Timeout: time.Duration(timeoutSecond) * time.Second}).Do(req)
	if err != nil {
		systemHealthWarnCtx(ctx, "system-health AI HTTP request failed",
			zap.String("endpoint", endpoint),
			zap.String("model", model),
			zap.Int("inputPayloadBytes", len(inputPayload)),
			zap.Int("requestBodyBytes", len(requestBody)),
			zap.Bool("timeout", isTimeoutError(err)),
			zap.Duration("elapsed", time.Since(callStartedAt)),
			zap.Error(err),
		)
		return systemHealthAISummaryResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		systemHealthWarnCtx(ctx, "system-health AI HTTP request bad status",
			zap.String("endpoint", endpoint),
			zap.String("model", model),
			zap.Int("httpStatus", resp.StatusCode),
			zap.Int("inputPayloadBytes", len(inputPayload)),
			zap.Int("requestBodyBytes", len(requestBody)),
			zap.Int("responseBodyBytes", len(body)),
			zap.String("rawResponseBody", string(body)),
			zap.String("responseBodyPreview", truncateText(strings.TrimSpace(string(body)), 220)),
			zap.Duration("elapsed", time.Since(callStartedAt)),
		)
		return systemHealthAISummaryResponse{}, fmt.Errorf("AI summary API returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	rawResponseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		systemHealthWarnCtx(ctx, "system-health AI response read failed",
			zap.String("endpoint", endpoint),
			zap.String("model", model),
			zap.Int("httpStatus", resp.StatusCode),
			zap.Duration("elapsed", time.Since(callStartedAt)),
			zap.Error(err),
		)
		return systemHealthAISummaryResponse{}, err
	}
	var responsesResp openAIResponsesResponse
	if err := json.Unmarshal(rawResponseBody, &responsesResp); err != nil {
		systemHealthWarnCtx(ctx, "system-health AI response decode failed",
			zap.String("endpoint", endpoint),
			zap.String("model", model),
			zap.Int("httpStatus", resp.StatusCode),
			zap.String("responseParseStage", "decode_openai_response"),
			zap.Int("responseBodyBytes", len(rawResponseBody)),
			zap.String("rawResponseBody", string(rawResponseBody)),
			zap.Duration("elapsed", time.Since(callStartedAt)),
			zap.Error(err),
		)
		return systemHealthAISummaryResponse{}, err
	}
	outputText := extractOpenAIResponsesText(responsesResp)
	systemHealthInfoCtx(ctx, "system-health AI raw response payload",
		zap.String("endpoint", endpoint),
		zap.String("model", model),
		zap.Int("httpStatus", resp.StatusCode),
		zap.String("responseID", responsesResp.ID),
		zap.Int("responseBodyBytes", len(rawResponseBody)),
		zap.Int("outputTextBytes", len(outputText)),
		zap.String("rawResponseBody", string(rawResponseBody)),
		zap.String("rawOutputText", outputText),
	)
	if responsesResp.Error != nil {
		systemHealthWarnCtx(ctx, "system-health AI response API error",
			zap.String("endpoint", endpoint),
			zap.String("model", model),
			zap.Int("httpStatus", resp.StatusCode),
			zap.String("apiErrorType", responsesResp.Error.Type),
			zap.String("apiErrorMessage", truncateText(responsesResp.Error.Message, 220)),
			zap.Duration("elapsed", time.Since(callStartedAt)),
		)
		return systemHealthAISummaryResponse{}, fmt.Errorf("AI response error: %s", responsesResp.Error.Message)
	}
	aiResp, err := parseSystemHealthAIResponseText(outputText)
	if err != nil {
		systemHealthWarnCtx(ctx, "system-health AI response parse failed",
			zap.String("endpoint", endpoint),
			zap.String("model", model),
			zap.Int("httpStatus", resp.StatusCode),
			zap.String("responseID", responsesResp.ID),
			zap.String("responseParseStage", "parse_summary_schema"),
			zap.Int("outputTextBytes", len(outputText)),
			zap.Duration("elapsed", time.Since(callStartedAt)),
			zap.Error(err),
		)
		return systemHealthAISummaryResponse{}, err
	}
	aiResp.WorkflowRunID = firstNonEmpty(aiResp.WorkflowRunID, responsesResp.ID)
	systemHealthInfoCtx(ctx, "system-health AI HTTP request completed",
		zap.String("endpoint", endpoint),
		zap.String("model", model),
		zap.Int("httpStatus", resp.StatusCode),
		zap.String("responseID", responsesResp.ID),
		zap.Int("outputTextBytes", len(outputText)),
		zap.Int("responseIssueCount", len(aiResp.Issues)),
		zap.Bool("summaryPresent", strings.TrimSpace(aiResp.Summary) != ""),
		zap.Duration("elapsed", time.Since(callStartedAt)),
	)
	return aiResp, nil
}

func effectiveSystemHealthAITimeoutSecond(timeoutSecond int) int {
	if timeoutSecond <= 0 {
		return 10
	}
	return timeoutSecond
}

func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

func openAIResponsesEndpoint(baseURL string) string {
	return strings.TrimRight(strings.TrimSpace(baseURL), "/") + "/responses"
}

func systemHealthAIResponseSchema() map[string]any {
	actionPlanItemSchema := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"phase", "target", "action", "validation", "risk"},
		"properties": map[string]any{
			"phase":      map[string]any{"type": "string", "minLength": 2, "maxLength": 16},
			"target":     map[string]any{"type": "string", "minLength": 2, "maxLength": 60},
			"action":     map[string]any{"type": "string", "minLength": 10, "maxLength": 120},
			"validation": map[string]any{"type": "string", "minLength": 8, "maxLength": 100},
			"risk":       map[string]any{"type": "string", "minLength": 8, "maxLength": 100},
		},
	}
	issueSchema := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"key", "title", "analysis", "rootCause", "metricAnalysis", "actionPlan"},
		"properties": map[string]any{
			"key":            map[string]any{"type": "string", "minLength": 1},
			"title":          map[string]any{"type": "string", "minLength": 4, "maxLength": 64},
			"analysis":       map[string]any{"type": "string", "minLength": 20, "maxLength": 240},
			"rootCause":      map[string]any{"type": "string", "minLength": 0, "maxLength": 180},
			"metricAnalysis": map[string]any{"type": "string", "minLength": 0, "maxLength": 160},
			"actionPlan":     map[string]any{"type": "array", "minItems": 2, "maxItems": 4, "items": actionPlanItemSchema},
		},
	}
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"summary", "issues"},
		"properties": map[string]any{
			"summary": map[string]any{"type": "string", "minLength": 20, "maxLength": 140},
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
			Key:                    issue.Key,
			Priority:               issue.Priority,
			Title:                  issue.Title,
			AlertNames:             uniqueStrings(input.alertNames),
			Status:                 issue.Status,
			Scope:                  issue.Scope,
			Impact:                 issue.Impact,
			ActiveAlerts:           issue.ActiveAlerts,
			StartedAt:              issue.StartedAt,
			DataSources:            issue.DataSources,
			MetricNarrativeContext: buildMetricAnalysisForIssueContext(issue.Scope, issue.Status, issue.ActiveAlerts, issue.Evidence),
			Evidence:               buildSystemHealthAIEvidenceInput(issue.Evidence),
			ExistingAlertAnalyses:  buildSystemHealthAIAlertAnalysisInput(input.events),
			ExistingReportFindings: buildSystemHealthAIReportFindingInput(input.alertReportFindings),
			ExistingReportActions:  buildSystemHealthAIReportActionInput(input.alertReportActions),
		})
	}
	return request
}

func buildSystemHealthAIEvidenceInput(evidence []response.SystemHealthEvidence) []systemHealthAIEvidenceInput {
	const maxAIEvidenceItems = 8
	items := make([]systemHealthAIEvidenceInput, 0, minInt(len(evidence), maxAIEvidenceItems))
	for _, item := range evidence {
		if len(items) >= maxAIEvidenceItems {
			break
		}
		if strings.TrimSpace(item.Result) == "" {
			continue
		}
		items = append(items, systemHealthAIEvidenceInput{
			Type:             item.Type,
			Source:           item.Source,
			Severity:         item.Severity,
			Result:           truncateText(item.Result, 360),
			Metric:           buildSystemHealthAIMetricInput(item.Metric),
			PolarisBreakdown: item.PolarisBreakdown,
		})
	}
	return items
}

func buildSystemHealthAIMetricInput(metric *response.SystemHealthMetricEvidence) *systemHealthAIMetricInput {
	if metric == nil {
		return nil
	}
	timeRange := metricTimeRangeText(metric)
	return &systemHealthAIMetricInput{
		MetricName:     metricUserFacingLabel(metric),
		RuleName:       metric.RuleName,
		SeriesMatched:  metric.SeriesMatched,
		SeriesReturned: metric.SeriesReturned,
		PointCount:     metric.PointCount,
		FirstValue:     metric.FirstValue,
		LastValue:      metric.LastValue,
		AverageValue:   metric.AverageValue,
		MinValue:       metric.MinValue,
		MaxValue:       metric.MaxValue,
		MaxDeviation:   metric.MaxDeviation,
		Delta:          metric.Delta,
		ChangePercent:  metric.ChangePercent,
		Trend:          metricTrendPhrase(metric),
		StartAt:        timeRange.StartAt,
		EndAt:          timeRange.EndAt,
		PeakAt:         timeRange.PeakAt,
		MinAt:          timeRange.MinAt,
		EventScoped:    metric.EventScoped,
		Message:        truncateText(metric.Message, 160),
	}
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
		alertDirection := cleanUserFacingReportText(event.AlertDirection)
		analyzeError := cleanUserFacingReportText(event.AnalyzeErr)
		checkOutput := cleanUserFacingReportText(event.Output)
		if alertDirection == "" && strings.TrimSpace(event.AnalyzeRunID) == "" && analyzeError == "" && checkOutput == "" {
			continue
		}
		items = append(items, systemHealthAIAlertAnalysisInput{
			EventID:        event.EventID,
			AlertID:        event.AlertID,
			AlertName:      event.Name,
			AlertDirection: truncateText(alertDirection, 220),
			AnalyzeRunID:   event.AnalyzeRunID,
			AnalyzeError:   truncateText(analyzeError, 220),
			CheckOutput:    truncateText(checkOutput, 260),
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
			AlertName:  action.AlertName,
			Level:      action.Level,
			Phase:      truncateText(action.Phase, 80),
			Target:     truncateText(action.Target, 120),
			Action:     truncateText(action.Action, 220),
			Validation: truncateText(action.Validation, 180),
			Risk:       truncateText(action.Risk, 180),
		})
	}
	return items
}

func buildSystemHealthAIReportFindingInput(findings []systemHealthAlertReportFinding) []systemHealthAIReportFindingInput {
	items := make([]systemHealthAIReportFindingInput, 0, minInt(len(findings), maxEvidenceSamples))
	for _, finding := range findings {
		if len(items) >= maxEvidenceSamples {
			break
		}
		rootCause := cleanAlertReportDirectionalText(finding.RootCause)
		rootCauseOther := cleanAlertReportDirectionalTexts(finding.RootCauseOther, 6, 260)
		overviewReason := cleanAlertReportDirectionalText(finding.AnalyzeProcess)
		detail := cleanAlertReportDirectionalText(finding.Detail)
		if rootCause == "" && len(rootCauseOther) == 0 && overviewReason == "" && detail == "" && len(finding.Evidence) == 0 {
			continue
		}
		items = append(items, systemHealthAIReportFindingInput{
			AlertName:      finding.AlertName,
			RootCause:      truncateText(rootCause, 260),
			RootCauseOther: rootCauseOther,
			AnalyzeProcess: truncateText(overviewReason, 1200),
			OverviewReason: truncateText(overviewReason, 1200),
			Detail:         truncateText(detail, 1200),
			OverviewDetail: truncateText(detail, 1200),
			ReportEvidence: buildSystemHealthAIReportEvidenceInput(finding.Evidence),
		})
	}
	return items
}

func buildSystemHealthAIReportEvidenceInput(evidence []systemHealthAlertReportEvidence) []systemHealthAIReportEvidenceInput {
	items := make([]systemHealthAIReportEvidenceInput, 0, minInt(len(evidence), 4))
	for _, item := range evidence {
		if len(items) >= 4 {
			break
		}
		summary := displayableAlertReportEvidenceSummary(item.Kind, item.Summary)
		if summary == "" {
			continue
		}
		items = append(items, systemHealthAIReportEvidenceInput{
			Kind:    item.Kind,
			Summary: truncateText(summary, 220),
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
		alertDirection := cleanUserFacingReportText(event.AlertDirection)
		analyzeError := cleanUserFacingReportText(event.AnalyzeErr)
		checkOutput := cleanUserFacingReportText(event.Output)
		if alertDirection == "" && strings.TrimSpace(event.AnalyzeRunID) == "" && analyzeError == "" && checkOutput == "" {
			continue
		}
		key := strings.Join([]string{
			event.EventID,
			event.AlertID,
			event.Name,
			alertDirection,
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
			AlertDirection: truncateText(alertDirection, 300),
			AnalyzeRunID:   event.AnalyzeRunID,
			AnalyzeError:   truncateText(analyzeError, 220),
			CheckOutput:    truncateText(checkOutput, 260),
			Validity:       firstNonEmpty(event.Validity, event.IsValid),
		})
	}
	return items
}

func (h *handler) queryAlertReportActions(ctx core.Context, input reportInput, startTime time.Time, endTime time.Time) []systemHealthAlertReportAction {
	actions, _, _ := h.queryAlertReportData(ctx, input, startTime, endTime)
	return actions
}

func (h *handler) queryAlertReportData(ctx core.Context, input reportInput, startTime time.Time, endTime time.Time) ([]systemHealthAlertReportAction, []systemHealthAlertReportFinding, []systemHealthAlertReportTopologyRole) {
	if h == nil || h.chRepo == nil || ctx == nil || len(input.events) == 0 {
		systemHealthWarnCtx(ctx, "system-health alert report lookup skipped",
			zap.String("issueKey", input.key),
			zap.String("aggregationKey", input.aggregationKey),
			zap.String("reason", "missing_repository_context_or_events"),
			zap.Bool("hasHandler", h != nil),
			zap.Int("eventCount", len(input.events)),
		)
		return nil, nil, nil
	}
	actions := make([]systemHealthAlertReportAction, 0, maxEvidenceSamples)
	findings := make([]systemHealthAlertReportFinding, 0, maxEvidenceSamples)
	topologyRoles := make([]systemHealthAlertReportTopologyRole, 0, maxEvidenceSamples)
	seenEvents := make(map[string]struct{}, len(input.events))
	reportStartTime := startTime.Add(-time.Hour)
	reportEndTime := endTime.Add(time.Hour)
	checked := 0
	hitCount := 0
	missCount := 0
	queryFailedCount := 0
	emptyEventIDCount := 0
	duplicateEventCount := 0
	lookupLimitHit := false
	sampleLimitHit := false
	lookupStartedAt := time.Now()
	systemHealthInfoCtx(ctx, "system-health alert report lookup started",
		zap.String("issueKey", input.key),
		zap.String("aggregationKey", input.aggregationKey),
		zap.String("scope", input.scope),
		zap.Int("eventCount", len(input.events)),
		zap.Int("maxLookups", maxAlertReportLookups),
		zap.Int("maxEvidenceSamples", maxEvidenceSamples),
		zap.Time("reportStartTime", reportStartTime),
		zap.Time("reportEndTime", reportEndTime),
	)
	eventByReportEventID := make(map[string]alertmodel.AEventWithWRecord, len(input.events)*2)
	eventsByAlertID := make(map[string]alertmodel.AEventWithWRecord, len(input.events))
	for _, event := range input.events {
		for _, candidate := range alertReportLookupEventIDCandidates(event) {
			eventByReportEventID[candidate.AlertEventID] = event
		}
		if alertID := strings.TrimSpace(event.AlertID); alertID != "" {
			if _, ok := eventsByAlertID[alertID]; !ok {
				eventsByAlertID[alertID] = event
			}
		}
	}
	seenReports := make(map[string]struct{}, maxEvidenceSamples)
	appendReport := func(report *model.AlertEventReport, event alertmodel.AEventWithWRecord, lookupMethodsTried []string, elapsed time.Duration) bool {
		if report == nil {
			return false
		}
		reportKey := alertReportDedupKey(report)
		if reportKey != "" {
			if _, ok := seenReports[reportKey]; ok {
				return false
			}
			seenReports[reportKey] = struct{}{}
		}
		reportAlertEventID := strings.TrimSpace(report.Tags["alertEventId"])
		if matchedEvent, ok := eventByReportEventID[reportAlertEventID]; ok {
			event = matchedEvent
		}
		topologyRole := alertReportTopologyRole(report, event)
		if topologyRole.HasTopology || topologyRole.Reason != "" {
			systemHealthInfoCtx(ctx, "system-health alert report topology role parsed",
				zap.String("issueKey", input.key),
				zap.String("aggregationKey", input.aggregationKey),
				zap.String("eventID", strings.TrimSpace(event.EventID)),
				zap.String("alertID", event.AlertID),
				zap.String("alertName", event.Name),
				zap.String("reportAlertEventID", reportAlertEventID),
				zap.String("topologyRole", topologyRole.Role),
				zap.String("topologyReason", topologyRole.Reason),
				zap.String("currentService", topologyRole.CurrentService),
				zap.String("currentEndpoint", topologyRole.CurrentEndpoint),
				zap.Int("parentCount", topologyRole.ParentCount),
				zap.Int("childCount", topologyRole.ChildCount),
				zap.Strings("parentNames", topologyRole.ParentNames),
				zap.Strings("childNames", topologyRole.ChildNames),
				zap.Bool("hasTopology", topologyRole.HasTopology),
			)
		}
		eventFindings := alertReportFindings(report, event.Name, topologyRole)
		eventActions := parseAlertReportSuggestActions(report.Suggest, event.Name)
		hitCount++
		systemHealthInfoCtx(ctx, "system-health alert report lookup hit",
			zap.String("issueKey", input.key),
			zap.String("aggregationKey", input.aggregationKey),
			zap.String("eventID", strings.TrimSpace(event.EventID)),
			zap.String("alertID", event.AlertID),
			zap.String("alertName", event.Name),
			zap.Strings("lookupMethodsTried", lookupMethodsTried),
			zap.String("reportAlertEventID", reportAlertEventID),
			zap.Int("suggestRawCount", len(report.Suggest)),
			zap.Int("actionCount", len(eventActions)),
			zap.Int("findingCount", len(eventFindings)),
			zap.Int("evidenceCount", countAlertReportEvidence(eventFindings)),
			zap.Bool("hasRootCause", strings.TrimSpace(report.RootCause) != "" || len(report.RootCauseOther) > 0),
			zap.Bool("hasOverview", strings.TrimSpace(report.OverviewReason) != "" || strings.TrimSpace(report.OverviewDetail) != ""),
			zap.Bool("hasTopology", strings.TrimSpace(report.Topology) != ""),
			zap.Int("dataCount", len(report.Data)),
			zap.Duration("elapsed", elapsed),
		)
		findings = append(findings, eventFindings...)
		findings = uniqueAlertReportFindings(findings)
		if topologyRole.HasTopology || topologyRole.Reason != "" {
			topologyRoles = append(topologyRoles, topologyRole)
			topologyRoles = uniqueAlertReportTopologyRoles(topologyRoles)
		}
		actions = append(actions, eventActions...)
		actions = uniqueAlertReportActions(actions)
		return true
	}
	for _, alertID := range alertIDsFromEvents(input.events) {
		if len(actions) >= maxEvidenceSamples && len(findings) >= maxEvidenceSamples {
			sampleLimitHit = true
			break
		}
		attemptStartedAt := time.Now()
		systemHealthInfoCtx(ctx, "system-health alert report lookup by alert_id started",
			zap.String("issueKey", input.key),
			zap.String("aggregationKey", input.aggregationKey),
			zap.String("alertID", alertID),
			zap.Time("reportStartTime", reportStartTime),
			zap.Time("reportEndTime", reportEndTime),
		)
		reports, err := h.chRepo.GetAlertEventReportsByAlertID(ctx, reportStartTime, reportEndTime, alertID)
		if err != nil {
			reason, failed := alertReportLookupFailureReason(err)
			if failed {
				queryFailedCount++
			} else {
				missCount++
			}
			systemHealthWarnCtx(ctx, "system-health alert report lookup by alert_id missed",
				zap.String("issueKey", input.key),
				zap.String("aggregationKey", input.aggregationKey),
				zap.String("alertID", alertID),
				zap.String("reason", reason),
				zap.Error(err),
				zap.Duration("elapsed", time.Since(attemptStartedAt)),
			)
			continue
		}
		event := eventsByAlertID[alertID]
		appended := 0
		for i := range reports {
			if len(actions) >= maxEvidenceSamples && len(findings) >= maxEvidenceSamples {
				sampleLimitHit = true
				break
			}
			if appendReport(&reports[i], event, []string{"alert_id_workflow_join"}, time.Since(attemptStartedAt)) {
				appended++
			}
		}
		systemHealthInfoCtx(ctx, "system-health alert report lookup by alert_id completed",
			zap.String("issueKey", input.key),
			zap.String("aggregationKey", input.aggregationKey),
			zap.String("alertID", alertID),
			zap.Int("reportCount", len(reports)),
			zap.Int("appendedReportCount", appended),
			zap.Duration("elapsed", time.Since(attemptStartedAt)),
		)
	}
	for _, event := range input.events {
		if len(actions) >= maxEvidenceSamples && len(findings) >= maxEvidenceSamples {
			sampleLimitHit = true
			break
		}
		if checked >= maxAlertReportLookups {
			lookupLimitHit = true
			break
		}
		eventID := strings.TrimSpace(event.EventID)
		if eventID == "" {
			emptyEventIDCount++
			systemHealthWarnCtx(ctx, "system-health alert report lookup skipped for event",
				zap.String("issueKey", input.key),
				zap.String("aggregationKey", input.aggregationKey),
				zap.String("alertID", event.AlertID),
				zap.String("alertName", event.Name),
				zap.String("reason", "empty_event_id"),
			)
			continue
		}
		if _, ok := seenEvents[eventID]; ok {
			duplicateEventCount++
			continue
		}
		seenEvents[eventID] = struct{}{}
		checked++
		eventLookupStartedAt := time.Now()
		lookupCandidates := alertReportLookupEventIDCandidates(event)
		lookupMethodsTried := make([]string, 0, len(lookupCandidates)+1)
		systemHealthInfoCtx(ctx, "system-health alert report lookup event started",
			zap.String("issueKey", input.key),
			zap.String("aggregationKey", input.aggregationKey),
			zap.String("eventID", eventID),
			zap.String("alertID", event.AlertID),
			zap.String("alertName", event.Name),
			zap.String("workflowInputEventID", strings.TrimSpace(event.Input)),
			zap.String("analyzeRunID", strings.TrimSpace(event.AnalyzeRunID)),
			zap.String("workflowRunID", strings.TrimSpace(event.WorkflowRunID)),
			zap.String("lookupCandidates", systemHealthCompactJSON(lookupCandidates)),
		)
		var report *model.AlertEventReport
		var lastLookupErr error
		queryFailed := false
		for _, candidate := range lookupCandidates {
			lookupMethodsTried = append(lookupMethodsTried, candidate.Method)
			attemptStartedAt := time.Now()
			systemHealthInfoCtx(ctx, "system-health alert report lookup attempt started",
				zap.String("issueKey", input.key),
				zap.String("aggregationKey", input.aggregationKey),
				zap.String("eventID", eventID),
				zap.String("alertID", event.AlertID),
				zap.String("alertName", event.Name),
				zap.String("lookupMethod", candidate.Method),
				zap.String("lookupAlertEventID", candidate.AlertEventID),
			)
			candidateReport, err := h.chRepo.GetAlertEventReport(ctx, reportStartTime, reportEndTime, map[string]string{"alertEventId": candidate.AlertEventID})
			if err == nil && candidateReport != nil {
				report = candidateReport
				systemHealthInfoCtx(ctx, "system-health alert report lookup attempt hit",
					zap.String("issueKey", input.key),
					zap.String("aggregationKey", input.aggregationKey),
					zap.String("eventID", eventID),
					zap.String("alertID", event.AlertID),
					zap.String("alertName", event.Name),
					zap.String("lookupMethod", candidate.Method),
					zap.String("lookupAlertEventID", candidate.AlertEventID),
					zap.String("reportAlertEventID", strings.TrimSpace(report.Tags["alertEventId"])),
					zap.Duration("elapsed", time.Since(attemptStartedAt)),
				)
				break
			}
			lastLookupErr = err
			reason, failed := alertReportLookupFailureReason(err)
			queryFailed = queryFailed || failed
			systemHealthWarnCtx(ctx, "system-health alert report lookup attempt missed",
				zap.String("issueKey", input.key),
				zap.String("aggregationKey", input.aggregationKey),
				zap.String("eventID", eventID),
				zap.String("alertID", event.AlertID),
				zap.String("alertName", event.Name),
				zap.String("lookupMethod", candidate.Method),
				zap.String("lookupAlertEventID", candidate.AlertEventID),
				zap.String("reason", reason),
				zap.Error(err),
				zap.Duration("elapsed", time.Since(attemptStartedAt)),
			)
		}
		analyzeRunID := strings.TrimSpace(event.AnalyzeRunID)
		if report == nil && analyzeRunID != "" {
			lookupMethodsTried = append(lookupMethodsTried, "analyze_run_id")
			attemptStartedAt := time.Now()
			systemHealthInfoCtx(ctx, "system-health alert report lookup attempt started",
				zap.String("issueKey", input.key),
				zap.String("aggregationKey", input.aggregationKey),
				zap.String("eventID", eventID),
				zap.String("alertID", event.AlertID),
				zap.String("alertName", event.Name),
				zap.String("lookupMethod", "analyze_run_id"),
				zap.String("analyzeRunID", analyzeRunID),
			)
			candidateReport, err := h.chRepo.GetAlertEventReportByAnalyzeRunID(ctx, reportStartTime, reportEndTime, analyzeRunID)
			if err == nil && candidateReport != nil {
				report = candidateReport
				systemHealthInfoCtx(ctx, "system-health alert report lookup attempt hit",
					zap.String("issueKey", input.key),
					zap.String("aggregationKey", input.aggregationKey),
					zap.String("eventID", eventID),
					zap.String("alertID", event.AlertID),
					zap.String("alertName", event.Name),
					zap.String("lookupMethod", "analyze_run_id"),
					zap.String("analyzeRunID", analyzeRunID),
					zap.String("reportAlertEventID", strings.TrimSpace(report.Tags["alertEventId"])),
					zap.Duration("elapsed", time.Since(attemptStartedAt)),
				)
			} else {
				lastLookupErr = err
				reason, failed := alertReportLookupFailureReason(err)
				queryFailed = queryFailed || failed
				systemHealthWarnCtx(ctx, "system-health alert report lookup attempt missed",
					zap.String("issueKey", input.key),
					zap.String("aggregationKey", input.aggregationKey),
					zap.String("eventID", eventID),
					zap.String("alertID", event.AlertID),
					zap.String("alertName", event.Name),
					zap.String("lookupMethod", "analyze_run_id"),
					zap.String("analyzeRunID", analyzeRunID),
					zap.String("reason", reason),
					zap.Error(err),
					zap.Duration("elapsed", time.Since(attemptStartedAt)),
				)
			}
		}
		if report == nil {
			fields := []zap.Field{
				zap.String("issueKey", input.key),
				zap.String("aggregationKey", input.aggregationKey),
				zap.String("eventID", eventID),
				zap.String("alertID", event.AlertID),
				zap.String("alertName", event.Name),
				zap.String("workflowInputEventID", strings.TrimSpace(event.Input)),
				zap.String("analyzeRunID", analyzeRunID),
				zap.Strings("lookupMethodsTried", lookupMethodsTried),
				zap.Time("reportStartTime", reportStartTime),
				zap.Time("reportEndTime", reportEndTime),
				zap.Duration("elapsed", time.Since(eventLookupStartedAt)),
			}
			if queryFailed {
				queryFailedCount++
				fields = append(fields, zap.String("reason", "query_failed"), zap.Error(lastLookupErr))
			} else {
				missCount++
				fields = append(fields, zap.String("reason", "report_not_found"), zap.Error(lastLookupErr))
			}
			systemHealthWarnCtx(ctx, "system-health alert report lookup missed", fields...)
			continue
		}
		appendReport(report, event, lookupMethodsTried, time.Since(eventLookupStartedAt))
	}
	if len(actions) > maxEvidenceSamples {
		actions = actions[:maxEvidenceSamples]
	}
	if len(findings) > maxEvidenceSamples {
		findings = findings[:maxEvidenceSamples]
	}
	if len(topologyRoles) > maxEvidenceSamples {
		topologyRoles = topologyRoles[:maxEvidenceSamples]
	}
	systemHealthInfoCtx(ctx, "system-health alert report lookup completed",
		zap.String("issueKey", input.key),
		zap.String("aggregationKey", input.aggregationKey),
		zap.Int("eventCount", len(input.events)),
		zap.Int("uniqueEvents", len(seenEvents)),
		zap.Int("checkedEvents", checked),
		zap.Int("hitCount", hitCount),
		zap.Int("missCount", missCount),
		zap.Int("queryFailedCount", queryFailedCount),
		zap.Int("emptyEventIDCount", emptyEventIDCount),
		zap.Int("duplicateEventCount", duplicateEventCount),
		zap.Bool("lookupLimitHit", lookupLimitHit),
		zap.Bool("sampleLimitHit", sampleLimitHit),
		zap.Int("finalActionCount", len(actions)),
		zap.Int("finalFindingCount", len(findings)),
		zap.Int("finalTopologyRoleCount", len(topologyRoles)),
		zap.Duration("elapsed", time.Since(lookupStartedAt)),
	)
	return actions, findings, topologyRoles
}

type alertReportLookupCandidate struct {
	Method       string `json:"method"`
	AlertEventID string `json:"alertEventId"`
}

func alertReportLookupEventIDCandidates(event alertmodel.AEventWithWRecord) []alertReportLookupCandidate {
	candidates := make([]alertReportLookupCandidate, 0, 2)
	seen := make(map[string]struct{}, 2)
	appendCandidate := func(method string, eventID string) {
		eventID = strings.TrimSpace(eventID)
		if eventID == "" {
			return
		}
		if _, ok := seen[eventID]; ok {
			return
		}
		seen[eventID] = struct{}{}
		candidates = append(candidates, alertReportLookupCandidate{
			Method:       method,
			AlertEventID: eventID,
		})
	}
	appendCandidate("event_id", event.EventID)
	appendCandidate("workflow_input", event.Input)
	return candidates
}

func alertReportDedupKey(report *model.AlertEventReport) string {
	if report == nil {
		return ""
	}
	reportAlertEventID := strings.TrimSpace(report.Tags["alertEventId"])
	if reportAlertEventID != "" {
		return reportAlertEventID + "@" + report.OverviewTimestamp.Format(time.RFC3339Nano)
	}
	return strings.Join([]string{
		report.OverviewTimestamp.Format(time.RFC3339Nano),
		truncateText(report.OverviewDetail, 120),
		truncateText(report.OverviewReason, 120),
		truncateText(report.RootCause, 120),
	}, "\x00")
}

func alertReportLookupFailureReason(err error) (string, bool) {
	if err == nil {
		return "report_not_found", false
	}
	lower := strings.ToLower(err.Error())
	if strings.Contains(lower, "no event found") || strings.Contains(lower, "no alert report found") {
		return "report_not_found", false
	}
	if strings.Contains(lower, "empty analyze_run_id") {
		return "empty_analyze_run_id", false
	}
	return "query_failed", true
}

func countAlertReportEvidence(findings []systemHealthAlertReportFinding) int {
	count := 0
	for _, finding := range findings {
		count += len(finding.Evidence)
	}
	return count
}

func alertReportOverviewDetailCount(findings []systemHealthAlertReportFinding) int {
	count := 0
	for _, finding := range uniqueAlertReportFindings(findings) {
		if cleanAlertReportDirectionalText(finding.Detail) != "" {
			count++
		}
	}
	return count
}

func alertReportOverviewReasonCount(findings []systemHealthAlertReportFinding) int {
	count := 0
	for _, finding := range uniqueAlertReportFindings(findings) {
		if cleanAlertReportDirectionalText(finding.AnalyzeProcess) != "" {
			count++
		}
	}
	return count
}

func alertReportFindings(report *model.AlertEventReport, alertName string, topologyRoles ...systemHealthAlertReportTopologyRole) []systemHealthAlertReportFinding {
	if report == nil {
		return nil
	}
	rootCauseParts := make([]string, 0, 1+len(report.RootCauseOther))
	if rootCause := cleanAlertReportDirectionalText(report.RootCause); rootCause != "" {
		rootCauseParts = append(rootCauseParts, rootCause)
	}
	rootCauseOther := make([]string, 0, len(report.RootCauseOther))
	for _, raw := range report.RootCauseOther {
		if text := cleanAlertReportDirectionalText(alertReportJSONText(raw)); text != "" {
			rootCauseParts = append(rootCauseParts, text)
			rootCauseOther = append(rootCauseOther, truncateText(text, 260))
		}
	}
	finding := systemHealthAlertReportFinding{
		AlertName:      alertName,
		RootCause:      truncateText(strings.Join(uniqueStrings(rootCauseParts), "；"), 360),
		RootCauseOther: uniqueStrings(rootCauseOther),
		AnalyzeProcess: truncateText(cleanAlertReportDirectionalText(report.OverviewReason), 1200),
		Detail:         truncateText(cleanAlertReportDirectionalText(report.OverviewDetail), 1200),
		Evidence:       buildAlertReportEvidenceSummaries(report, topologyRoles...),
	}
	if finding.RootCause == "" && len(finding.RootCauseOther) == 0 && finding.AnalyzeProcess == "" && finding.Detail == "" && len(finding.Evidence) == 0 {
		return nil
	}
	return []systemHealthAlertReportFinding{finding}
}

func buildAlertReportEvidenceSummaries(report *model.AlertEventReport, topologyRoles ...systemHealthAlertReportTopologyRole) []systemHealthAlertReportEvidence {
	if report == nil {
		return nil
	}
	grouped := make(map[string][]string)
	appendSummary := func(kind string, summary string) {
		kind = strings.TrimSpace(kind)
		summary = displayableAlertReportEvidenceSummary(kind, summary)
		if kind == "" || summary == "" {
			return
		}
		grouped[kind] = uniqueStrings(append(grouped[kind], summary))
	}

	for _, raw := range report.Data {
		item := alertReportJSONMap(raw)
		text := alertReportEvidenceText(raw)
		if text == "" {
			continue
		}
		appendSummary(classifyAlertReportEvidence(item, text), text)
	}
	topologySummary := ""
	if len(topologyRoles) > 0 {
		topologySummary = strings.TrimSpace(topologyRoles[0].Summary)
	}
	if topologySummary == "" {
		topologySummary = alertReportTopologySummary(report.Topology)
	}
	if topologySummary != "" {
		appendSummary("apm", topologySummary)
	}
	if reportAnalysis := alertReportAnalysisSummary(report); reportAnalysis != "" {
		appendSummary("alert", reportAnalysis)
	}

	order := []string{"logs", "metrics", "alert"}
	result := make([]systemHealthAlertReportEvidence, 0, len(grouped))
	for _, kind := range order {
		summaries := grouped[kind]
		if len(summaries) == 0 {
			continue
		}
		result = append(result, systemHealthAlertReportEvidence{
			Kind:    kind,
			Summary: truncateText(strings.Join(summaries[:minInt(len(summaries), 2)], "；"), 360),
		})
	}
	return result
}

func alertReportJSONMap(raw string) map[string]any {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var item map[string]any
	if err := json.Unmarshal([]byte(raw), &item); err != nil {
		return nil
	}
	return item
}

func alertReportEvidenceText(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	item := alertReportJSONMap(raw)
	if item == nil {
		return truncateText(cleanUserFacingReportText(raw), 260)
	}
	values := make([]string, 0, 4)
	collectReportEvidenceText(item, 0, &values)
	return truncateText(strings.Join(uniqueStrings(values), "；"), 320)
}

func collectReportEvidenceText(value any, depth int, values *[]string) {
	if depth > 3 || len(*values) >= 6 || value == nil {
		return
	}
	switch typed := value.(type) {
	case string:
		text := cleanUserFacingReportText(typed)
		if isUsefulReportEvidenceText(text) {
			*values = append(*values, text)
		}
	case []any:
		for _, item := range typed {
			collectReportEvidenceText(item, depth+1, values)
			if len(*values) >= 6 {
				return
			}
		}
	case map[string]any:
		for _, key := range []string{"summary", "conclusion", "result", "analysis", "detail", "description", "content", "message", "reason", "title", "name"} {
			if raw, ok := typed[key]; ok {
				collectReportEvidenceText(raw, depth+1, values)
				if len(*values) >= 6 {
					return
				}
			}
		}
		for key, raw := range typed {
			if isMetadataReportKey(key) {
				continue
			}
			collectReportEvidenceText(raw, depth+1, values)
			if len(*values) >= 6 {
				return
			}
		}
	}
}

func isUsefulReportEvidenceText(text string) bool {
	text = cleanUserFacingReportText(text)
	if len([]rune(text)) < 4 {
		return false
	}
	lower := strings.ToLower(text)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return false
	}
	return true
}

func isMetadataReportKey(key string) bool {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "id", "eventid", "alertid", "timestamp", "time", "starttime", "endtime", "traceid", "spanid", "runid", "workflowrunid":
		return true
	default:
		return false
	}
}

func classifyAlertReportEvidence(item map[string]any, text string) string {
	signatureParts := []string{text}
	for _, key := range []string{"type", "source", "category", "name", "title", "kind", "module"} {
		if value := stringValueFromMap(item, key); value != "" {
			signatureParts = append(signatureParts, value)
		}
	}
	signature := strings.ToLower(strings.Join(signatureParts, " "))
	switch {
	case strings.Contains(signature, "日志") || strings.Contains(signature, "log"):
		return "logs"
	case strings.Contains(signature, "apm") || strings.Contains(signature, "trace") || strings.Contains(signature, "span") || strings.Contains(signature, "链路") || strings.Contains(signature, "拓扑") || strings.Contains(signature, "调用"):
		return "apm"
	case strings.Contains(signature, "指标") || strings.Contains(signature, "metric") || strings.Contains(signature, "red") || strings.Contains(signature, "prometheus") || strings.Contains(signature, "pql") || strings.Contains(signature, "latency") || strings.Contains(signature, "duration") || strings.Contains(signature, "error_rate"):
		return "metrics"
	case strings.Contains(signature, "告警") || strings.Contains(signature, "alert") || strings.Contains(signature, "分析") || strings.Contains(signature, "有效"):
		return "alert"
	default:
		return "alert"
	}
}

func normalizeReportEvidenceSummary(text string) string {
	text = cleanUserFacingReportText(text)
	if text == "" {
		return ""
	}
	replacements := []struct {
		old string
		new string
	}{
		{"raw_tags", "告警标签"},
		{"span_trace", "链路样本"},
		{"Prometheus", "指标系统"},
		{"prometheus", "指标系统"},
		{"PQL", "指标查询"},
		{"pql", "指标查询"},
	}
	for _, replacement := range replacements {
		text = strings.ReplaceAll(text, replacement.old, replacement.new)
	}
	return truncateText(text, 320)
}

func displayableAlertReportEvidenceSummary(kind string, summary string) string {
	kind = strings.ToLower(strings.TrimSpace(kind))
	if kind == "apm" {
		return ""
	}
	summary = normalizeReportEvidenceSummary(summary)
	if summary == "" {
		return ""
	}
	if kind == "logs" {
		return cleanDisplayableLogEvidenceSummary(summary)
	}
	return summary
}

func cleanDisplayableLogEvidenceSummary(summary string) string {
	summary = stripLogEvidencePlaceholders(summary)
	if summary == "" || !hasConcreteLogEvidence(summary) {
		return ""
	}
	return truncateText(redactLogBusinessContent(summary), 320)
}

func redactLogBusinessContent(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	replacements := []struct {
		pattern string
		value   string
	}{
		{`\b[A-Z][0-9]{2,5}(?:\s*[~～-]\s*[A-Z]?[0-9]{2,5})?\b`, "车次已脱敏"},
		{`\b1[3-9]\d{9}\b`, "手机号已脱敏"},
		{`(?i)\b[A-Z0-9._%+\-]+@[A-Z0-9.\-]+\.[A-Z]{2,}\b`, "邮箱已脱敏"},
		{`(?i)\b(token|authorization|session|cookie|password|passwd|secret)\s*[:=]\s*["']?[^"',;\s}]+`, `${1}=凭据已脱敏`},
		{`\b(orderId|order_id|userId|user_id|passengerId|passenger_id|ticketId|ticket_id|trainNo|train_no|seatNo|seat_no|routeId|route_id)\s*[:=]\s*["']?[^"',;\s}]+`, `${1}=已脱敏`},
		{`(订单|用户|乘客|车次|座位|席位|库存|余额|票号|身份证|证件号|手机号|路线|站点|始发站|终点站)\s*[:=：]\s*["“']?[^"',，；;\s}”]+`, `${1}=已脱敏`},
		{`\b\d{8,}\b`, "长数字已脱敏"},
	}
	for _, replacement := range replacements {
		text = regexp.MustCompile(replacement.pattern).ReplaceAllString(text, replacement.value)
	}
	return text
}

func stripLogEvidencePlaceholders(summary string) string {
	summary = strings.TrimSpace(summary)
	if summary == "" {
		return ""
	}
	replacer := strings.NewReplacer(
		"服务xx在xx时间段内出现了xxx日志错误", " ",
		"服务 xx 在 xx 时间段内出现了 xxx 日志错误", " ",
		"服务xx在xx时间段内出现xxx日志错误", " ",
		"服务 xx 在 xx 时间段内出现 xxx 日志错误", " ",
		"Let's get this organized!", " ",
		"Let''s get this organized!", " ",
		"Let’s get this organized!", " ",
	)
	summary = replacer.Replace(summary)
	summary = strings.TrimSpace(strings.Trim(summary, "：:;；,.，。\"'“”‘’ "))
	return strings.Join(strings.Fields(summary), " ")
}

func hasConcreteLogEvidence(summary string) bool {
	text := strings.TrimSpace(summary)
	if text == "" {
		return false
	}
	lower := strings.ToLower(text)
	negativeMarkers := []string{
		"无日志数据", "无可用日志", "日志为空", "未查询到日志", "没有查询到日志",
		"未检索到日志", "没有检索到日志", "未发现日志", "未发现异常日志",
		"未发现错误日志", "未发现慢请求", "无持续慢请求", "no log", "not found",
		"正常日志", "日志正常", "info 为主", "info为主", "无明显报错", "未见明显报错",
		"未持续新增", "没有持续新增", "无异常堆积", "恢复后未持续新增",
	}
	for _, marker := range negativeMarkers {
		if strings.Contains(lower, strings.ToLower(marker)) {
			return false
		}
	}
	if containsAnyText(lower, "error", "exception", "timeout", "fatal", "panic", "traceback", "5xx", "4xx") {
		return true
	}
	if containsAnyText(text, "异常", "错误", "超时", "堆栈", "报错", "失败", "慢请求", "错误日志") {
		return true
	}
	return false
}

func alertReportTopologySummary(raw string) string {
	item := alertReportJSONMap(raw)
	if item == nil {
		return ""
	}
	names := make([]string, 0, 6)
	collectTopologyNames(item, 0, &names)
	names = uniqueStrings(names)
	if len(names) == 0 {
		return "方向报告包含调用拓扑，用于确认告警目标与上下游链路关系。"
	}
	return fmt.Sprintf("方向报告的调用拓扑涉及 %s，用于确认告警目标与上下游链路关系。", strings.Join(names[:minInt(len(names), 5)], "、"))
}

func alertReportTopologyRole(report *model.AlertEventReport, event alertmodel.AEventWithWRecord) systemHealthAlertReportTopologyRole {
	role := systemHealthAlertReportTopologyRole{
		AlertName:          event.Name,
		ReportAlertEventID: "",
		Role:               "unknown",
	}
	if report == nil {
		role.Reason = "report_missing"
		return role
	}
	role.ReportAlertEventID = strings.TrimSpace(report.Tags["alertEventId"])
	raw := strings.TrimSpace(report.Topology)
	if raw == "" {
		role.Reason = "topology_missing"
		return role
	}
	role.HasTopology = true
	var item map[string]any
	if err := json.Unmarshal([]byte(raw), &item); err != nil {
		role.Reason = "topology_json_parse_failed"
		role.Summary = "方向报告包含调用拓扑，但拓扑内容无法解析，暂不能确认该接口是否为入口。"
		return role
	}
	if len(item) == 0 {
		role.Reason = "topology_empty"
		role.Summary = "方向报告的调用拓扑为空，暂不能确认该接口是否为入口。"
		return role
	}

	current := mapValueFromMap(item, "current")
	if current == nil {
		role.Reason = "topology_current_missing"
		role.Summary = "方向报告包含调用拓扑，但缺少当前告警接口节点，暂不能确认该接口是否为入口。"
		return role
	}
	role.CurrentService = topologyNodeService(current)
	role.CurrentEndpoint = topologyNodeEndpoint(current)
	if role.CurrentService == "" {
		role.CurrentService = firstNonEmpty(report.OverviewTags["service"], report.OverviewTags["serviceName"], report.OverviewTags["svc_name"], alertServiceName(event))
	}
	if role.CurrentEndpoint == "" {
		role.CurrentEndpoint = firstNonEmpty(report.OverviewTags["endpoint"], report.OverviewTags["content_key"], alertEndpoint(event))
	}
	if !topologyCurrentMatchesEvent(role.CurrentService, role.CurrentEndpoint, report, event) {
		role.Reason = "topology_current_mismatch"
		role.Summary = "方向报告拓扑中的当前节点与告警对象不一致，暂不使用该拓扑判断该接口是否为入口。"
		return role
	}

	parents, parentsExists := topologyRelationValue(item, "parents", "parent", "upstreams", "upstream", "parentNodes", "upstreamNodes")
	children, _ := topologyRelationValue(item, "children", "child", "downstreams", "downstream", "childNodes", "downstreamNodes")
	role.ParentCount = topologyNodeCount(parents)
	role.ChildCount = topologyNodeCount(children)
	role.ParentNames = collectTopologyNodeDisplayNames(parents, 4)
	role.ChildNames = collectTopologyNodeDisplayNames(children, 4)
	if !parentsExists {
		role.Reason = "topology_parents_missing"
		role.Summary = topologyRoleUnknownSummary(role)
		return role
	}
	if role.ParentCount == 0 {
		role.Role = "entry"
		role.Reason = "topology_current_has_no_parents"
		role.Summary = topologyRoleEntrySummary(role)
		return role
	}
	role.Role = "internalEndpoint"
	role.Reason = "topology_current_has_upstream_parents"
	role.Summary = topologyRoleInternalEndpointSummary(role)
	return role
}

func mapValueFromMap(item map[string]any, key string) map[string]any {
	if item == nil {
		return nil
	}
	raw, ok := item[key]
	if !ok || raw == nil {
		return nil
	}
	if typed, ok := raw.(map[string]any); ok {
		return typed
	}
	return nil
}

func topologyRelationValue(item map[string]any, keys ...string) (any, bool) {
	if item == nil {
		return nil, false
	}
	for _, key := range keys {
		if raw, ok := item[key]; ok {
			return raw, true
		}
	}
	return nil, false
}

func topologyNodeService(item map[string]any) string {
	return firstNonEmpty(
		stringValueFromMap(item, "service"),
		stringValueFromMap(item, "serviceName"),
		stringValueFromMap(item, "svc_name"),
		stringValueFromMap(item, "service_name"),
		stringValueFromMap(item, "application"),
		stringValueFromMap(item, "app"),
	)
}

func topologyNodeEndpoint(item map[string]any) string {
	return firstNonEmpty(
		stringValueFromMap(item, "endpoint"),
		stringValueFromMap(item, "contentKey"),
		stringValueFromMap(item, "content_key"),
		stringValueFromMap(item, "route"),
		stringValueFromMap(item, "uri"),
		stringValueFromMap(item, "path"),
		stringValueFromMap(item, "url"),
	)
}

func topologyCurrentMatchesEvent(currentService string, currentEndpoint string, report *model.AlertEventReport, event alertmodel.AEventWithWRecord) bool {
	expectedService := firstNonEmpty(report.OverviewTags["service"], report.OverviewTags["serviceName"], report.OverviewTags["svc_name"], alertServiceName(event))
	expectedEndpoint := firstNonEmpty(report.OverviewTags["endpoint"], report.OverviewTags["content_key"], alertEndpoint(event))
	if expectedService != "" && currentService != "" && !sameTopologyToken(expectedService, currentService) {
		return false
	}
	if expectedEndpoint != "" && currentEndpoint != "" && !sameTopologyToken(expectedEndpoint, currentEndpoint) {
		return false
	}
	return currentService != "" || currentEndpoint != ""
}

func sameTopologyToken(left string, right string) bool {
	left = normalizeTopologyToken(left)
	right = normalizeTopologyToken(right)
	if left == "" || right == "" {
		return false
	}
	return left == right
}

func normalizeTopologyToken(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "")
	value = strings.Trim(value, "/")
	return value
}

func topologyNodeCount(value any) int {
	switch typed := value.(type) {
	case nil:
		return 0
	case []any:
		count := 0
		for _, item := range typed {
			count += topologyNodeCount(item)
		}
		return count
	case map[string]any:
		if len(typed) == 0 {
			return 0
		}
		return 1
	case string:
		if strings.TrimSpace(typed) == "" {
			return 0
		}
		return 1
	default:
		return 1
	}
}

func collectTopologyNodeDisplayNames(value any, limit int) []string {
	if limit <= 0 || value == nil {
		return nil
	}
	names := make([]string, 0, limit)
	var collect func(any)
	collect = func(raw any) {
		if len(names) >= limit || raw == nil {
			return
		}
		switch typed := raw.(type) {
		case []any:
			for _, item := range typed {
				collect(item)
				if len(names) >= limit {
					return
				}
			}
		case map[string]any:
			name := topologyNodeDisplayName(typed)
			if name != "" {
				names = append(names, name)
			}
		case string:
			if text := strings.TrimSpace(typed); isUsefulTopologyName(text) {
				names = append(names, text)
			}
		}
	}
	collect(value)
	return uniqueStrings(names)
}

func topologyNodeDisplayName(item map[string]any) string {
	service := topologyNodeService(item)
	endpoint := topologyNodeEndpoint(item)
	if service != "" && endpoint != "" {
		return service + " / " + endpoint
	}
	if service != "" {
		return service
	}
	if endpoint != "" {
		return endpoint
	}
	return firstNonEmpty(stringValueFromMap(item, "name"), stringValueFromMap(item, "target"), stringValueFromMap(item, "source"))
}

func topologyRoleEntrySummary(role systemHealthAlertReportTopologyRole) string {
	target := topologyRoleTarget(role)
	childText := ""
	if role.ChildCount > 0 {
		childText = fmt.Sprintf("，下游关联 %d 个调用对象", role.ChildCount)
		if len(role.ChildNames) > 0 {
			childText += "（" + strings.Join(role.ChildNames[:minInt(len(role.ChildNames), 3)], "、") + "）"
		}
	}
	return fmt.Sprintf("方向报告拓扑显示%s未发现上游调用方%s；若延时复发，优先核对该请求的延时、错误率和主要下游耗时。", target, childText)
}

func topologyRoleInternalEndpointSummary(role systemHealthAlertReportTopologyRole) string {
	target := topologyRoleTarget(role)
	parentText := fmt.Sprintf("存在 %d 个上游调用方", role.ParentCount)
	if len(role.ParentNames) > 0 {
		parentText += "（" + strings.Join(role.ParentNames[:minInt(len(role.ParentNames), 3)], "、") + "）"
	}
	childText := ""
	if role.ChildCount > 0 {
		childText = fmt.Sprintf("，并继续调用 %d 个下游对象", role.ChildCount)
		if len(role.ChildNames) > 0 {
			childText += "（" + strings.Join(role.ChildNames[:minInt(len(role.ChildNames), 3)], "、") + "）"
		}
	}
	return fmt.Sprintf("方向报告拓扑显示%s%s%s，更适合作为服务接口/链路中间节点分析，而不是直接作为入口侧接口处理。", target, parentText, childText)
}

func topologyRoleUnknownSummary(role systemHealthAlertReportTopologyRole) string {
	target := topologyRoleTarget(role)
	return fmt.Sprintf("方向报告包含%s的调用拓扑，但缺少上游调用方信息，暂不能仅凭该报告确认该接口是否位于入口侧。", target)
}

func topologyRoleTarget(role systemHealthAlertReportTopologyRole) string {
	if role.CurrentService != "" && role.CurrentEndpoint != "" {
		return role.CurrentService + " / " + role.CurrentEndpoint
	}
	if role.CurrentService != "" {
		return role.CurrentService
	}
	if role.CurrentEndpoint != "" {
		return role.CurrentEndpoint
	}
	return "当前告警接口"
}

func collectTopologyNames(value any, depth int, names *[]string) {
	if depth > 4 || len(*names) >= 8 || value == nil {
		return
	}
	switch typed := value.(type) {
	case string:
		text := strings.TrimSpace(typed)
		if isUsefulTopologyName(text) {
			*names = append(*names, text)
		}
	case []any:
		for _, item := range typed {
			collectTopologyNames(item, depth+1, names)
			if len(*names) >= 8 {
				return
			}
		}
	case map[string]any:
		for _, key := range []string{"service", "serviceName", "name", "endpoint", "contentKey", "url", "target", "source"} {
			if raw, ok := typed[key]; ok {
				collectTopologyNames(raw, depth+1, names)
				if len(*names) >= 8 {
					return
				}
			}
		}
		for key, raw := range typed {
			if isMetadataReportKey(key) {
				continue
			}
			collectTopologyNames(raw, depth+1, names)
			if len(*names) >= 8 {
				return
			}
		}
	}
}

func isUsefulTopologyName(text string) bool {
	text = strings.TrimSpace(text)
	if len([]rune(text)) < 3 || len([]rune(text)) > 90 {
		return false
	}
	lower := strings.ToLower(text)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return false
	}
	for _, keyword := range []string{"service", "/api/", "post ", "get ", "put ", "delete ", "入口", "下游", "上游"} {
		if strings.Contains(lower, keyword) || strings.Contains(text, keyword) {
			return true
		}
	}
	return strings.Contains(text, "-")
}

func alertReportAnalysisSummary(report *model.AlertEventReport) string {
	parts := []string{
		cleanAlertReportDirectionalText(report.OverviewReason),
		cleanAlertReportDirectionalText(report.OverviewDetail),
		cleanAlertReportDirectionalText(report.RootCause),
	}
	for _, raw := range report.RootCauseOther {
		if text := cleanAlertReportDirectionalText(alertReportJSONText(raw)); text != "" {
			parts = append(parts, text)
		}
	}
	parts = uniqueStrings(parts)
	if len(parts) == 0 {
		return ""
	}
	return truncateText(strings.Join(parts[:minInt(len(parts), 3)], "；"), 360)
}

func alertReportJSONText(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	var item map[string]any
	if err := json.Unmarshal([]byte(raw), &item); err != nil {
		return truncateText(cleanUserFacingReportText(raw), 180)
	}
	for _, key := range []string{"cause", "rootCause", "reason", "summary", "description", "detail", "content", "message"} {
		if value := stringValueFromMap(item, key); value != "" {
			return truncateText(cleanUserFacingReportText(value), 180)
		}
	}
	return truncateText(raw, 180)
}

func uniqueAlertReportFindings(findings []systemHealthAlertReportFinding) []systemHealthAlertReportFinding {
	seen := make(map[string]struct{}, len(findings))
	result := make([]systemHealthAlertReportFinding, 0, len(findings))
	for _, finding := range findings {
		finding.AlertName = strings.TrimSpace(finding.AlertName)
		finding.RootCause = cleanAlertReportDirectionalText(finding.RootCause)
		finding.RootCauseOther = cleanAlertReportDirectionalTexts(finding.RootCauseOther, 6, 260)
		finding.AnalyzeProcess = cleanAlertReportDirectionalText(finding.AnalyzeProcess)
		finding.Detail = cleanAlertReportDirectionalText(finding.Detail)
		finding.Evidence = uniqueAlertReportEvidence(finding.Evidence)
		if finding.RootCause == "" && len(finding.RootCauseOther) == 0 && finding.AnalyzeProcess == "" && finding.Detail == "" && len(finding.Evidence) == 0 {
			continue
		}
		key := strings.ToLower(strings.Join([]string{finding.AlertName, finding.RootCause, strings.Join(finding.RootCauseOther, "；"), finding.AnalyzeProcess, finding.Detail, alertReportEvidenceKey(finding.Evidence)}, "\x00"))
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, finding)
	}
	return result
}

func uniqueAlertReportEvidence(evidence []systemHealthAlertReportEvidence) []systemHealthAlertReportEvidence {
	seen := make(map[string]struct{}, len(evidence))
	result := make([]systemHealthAlertReportEvidence, 0, len(evidence))
	for _, item := range evidence {
		item.Kind = strings.TrimSpace(item.Kind)
		item.Summary = cleanUserFacingReportText(item.Summary)
		if item.Kind == "" || item.Summary == "" {
			continue
		}
		key := strings.ToLower(item.Kind + "\x00" + item.Summary)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, item)
	}
	return result
}

func alertReportEvidenceKey(evidence []systemHealthAlertReportEvidence) string {
	parts := make([]string, 0, len(evidence))
	for _, item := range evidence {
		parts = append(parts, item.Kind+":"+item.Summary)
	}
	return strings.Join(parts, "\x00")
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
			stringValueFromMap(item, "建议"),
			stringValueFromMap(item, "动作"),
			stringValueFromMap(item, "行动"),
			stringValueFromMap(item, "处置"),
			stringValueFromMap(item, "处理建议"),
			stringValueFromMap(item, "建议行动"),
			stringValueFromMap(item, "title"),
			stringValueFromMap(item, "标题"),
			stringValueFromMap(item, "content"),
			stringValueFromMap(item, "内容"),
			stringValueFromMap(item, "description"),
			stringValueFromMap(item, "描述"),
			stringValueFromMap(item, "说明"),
			stringValueFromMap(item, "suggestion"),
		)
		if actionText == "" {
			continue
		}
		actions = append(actions, systemHealthAlertReportAction{
			AlertName:  alertName,
			Level:      strings.ToLower(firstNonEmpty(stringValueFromMap(item, "level"), stringValueFromMap(item, "级别"), stringValueFromMap(item, "优先级"))),
			Phase:      truncateText(firstNonEmpty(stringValueFromMap(item, "phase"), stringValueFromMap(item, "stage"), stringValueFromMap(item, "阶段"), stringValueFromMap(item, "步骤"), stringValueFromMap(item, "环节")), 80),
			Target:     truncateText(firstNonEmpty(stringValueFromMap(item, "target"), stringValueFromMap(item, "object"), stringValueFromMap(item, "scope"), stringValueFromMap(item, "目标"), stringValueFromMap(item, "对象"), stringValueFromMap(item, "范围"), stringValueFromMap(item, "排查对象")), 120),
			Action:     truncateText(actionText, 220),
			Validation: truncateText(firstNonEmpty(stringValueFromMap(item, "validation"), stringValueFromMap(item, "verification"), stringValueFromMap(item, "validate"), stringValueFromMap(item, "expected"), stringValueFromMap(item, "check"), stringValueFromMap(item, "验证"), stringValueFromMap(item, "验证方式"), stringValueFromMap(item, "验收"), stringValueFromMap(item, "确认项"), stringValueFromMap(item, "检查项"), stringValueFromMap(item, "预期"), stringValueFromMap(item, "预期结果")), 220),
			Risk:       truncateText(firstNonEmpty(stringValueFromMap(item, "risk"), stringValueFromMap(item, "caution"), stringValueFromMap(item, "impact"), stringValueFromMap(item, "风险"), stringValueFromMap(item, "注意"), stringValueFromMap(item, "影响"), stringValueFromMap(item, "注意事项")), 220),
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

func alertEventDistributionFields(events []alertmodel.AEventWithWRecord) []zap.Field {
	statusCounts := make(map[string]int)
	validityCounts := make(map[string]int)
	severityCounts := make(map[string]int)
	sourceCounts := make(map[string]int)
	kindCounts := make(map[string]int)
	services := make(map[string]struct{})
	endpoints := make(map[string]struct{})
	alertIDs := make(map[string]struct{})
	eventIDs := make(map[string]struct{})
	alertNames := make(map[string]struct{})
	aggregationKeys := make(map[string]struct{})
	for _, event := range events {
		statusCounts[firstNonEmpty(normalizeStatus(event.Status), "unknown")]++
		validityCounts[strings.ToLower(firstNonEmpty(event.Validity, event.IsValid, "unknown"))]++
		severityCounts[firstNonEmpty(normalizeSeverity(event.Severity), "unknown")]++
		sourceCounts[firstNonEmpty(event.Source, event.SourceID, "unknown")]++
		kind := classifyIssueKind(event)
		if kind == "" {
			kind = "unclassified"
		}
		kindCounts[kind]++
		if service := alertServiceName(event); service != "" {
			services[service] = struct{}{}
		}
		if endpoint := alertEndpoint(event); endpoint != "" {
			endpoints[endpoint] = struct{}{}
		}
		if event.AlertID != "" {
			alertIDs[event.AlertID] = struct{}{}
		}
		if event.EventID != "" {
			eventIDs[event.EventID] = struct{}{}
		}
		if event.Name != "" {
			alertNames[event.Name] = struct{}{}
		}
		normalized := normalizeAlertEvent(event)
		if normalized.AggregationKey != "" {
			aggregationKeys[normalized.AggregationKey] = struct{}{}
		}
	}
	return []zap.Field{
		zap.String("statusCounts", systemHealthCompactJSON(statusCounts)),
		zap.String("validityCounts", systemHealthCompactJSON(validityCounts)),
		zap.String("severityCounts", systemHealthCompactJSON(severityCounts)),
		zap.String("sourceCounts", systemHealthCompactJSON(sourceCounts)),
		zap.String("kindCounts", systemHealthCompactJSON(kindCounts)),
		zap.Int("firingCount", statusCounts[alertmodel.StatusFiring]),
		zap.Int("resolvedCount", statusCounts[alertmodel.StatusResolved]),
		zap.Int("uniqueEventCount", len(eventIDs)),
		zap.Int("uniqueAlertInstanceCount", len(alertIDs)),
		zap.Int("uniqueAlertNameCount", len(alertNames)),
		zap.Int("uniqueServiceCount", len(services)),
		zap.Int("uniqueEndpointCount", len(endpoints)),
		zap.Int("uniqueAggregationKeyCount", len(aggregationKeys)),
		zap.Strings("sampleAlertNames", firstNStrings(stringSetKeys(alertNames), 8)),
		zap.Strings("sampleAggregationKeys", firstNStrings(stringSetKeys(aggregationKeys), 8)),
	}
}

func stringSetKeys(values map[string]struct{}) []string {
	if len(values) == 0 {
		return nil
	}
	result := make([]string, 0, len(values))
	for value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func firstNStrings(values []string, limit int) []string {
	if len(values) <= limit {
		return values
	}
	return values[:limit]
}

func evidenceTypes(evidence []response.SystemHealthEvidence) []string {
	result := make([]string, 0, len(evidence))
	for _, item := range evidence {
		result = append(result, item.Type)
	}
	return uniqueStrings(result)
}

func systemHealthAIIssueInputKeys(issues []systemHealthAIIssueInput) []string {
	result := make([]string, 0, len(issues))
	for _, issue := range issues {
		result = append(result, issue.Key)
	}
	return uniqueStrings(result)
}

func systemHealthAIIssueOutputKeys(issues []systemHealthAIIssueOutput) []string {
	result := make([]string, 0, len(issues))
	for _, issue := range issues {
		result = append(result, issue.Key)
	}
	return uniqueStrings(result)
}

func missingAIIssueKeys(inputs []systemHealthAIIssueInput, outputs map[string]systemHealthAIIssueOutput) []string {
	result := make([]string, 0)
	for _, issue := range inputs {
		key := strings.TrimSpace(issue.Key)
		if key == "" {
			continue
		}
		if _, ok := outputs[key]; !ok {
			result = append(result, key)
		}
	}
	return result
}

func prometheusWarnings(warnings v1.Warnings) []string {
	result := make([]string, 0, len(warnings))
	for _, warning := range warnings {
		result = append(result, fmt.Sprint(warning))
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

	queryStartedAt := time.Now()
	systemHealthInfoCtx(ctx, "system-health real alert event query started",
		zap.Int64("groupId", groupID),
		zap.Int64("startTimeMicros", req.StartTime),
		zap.Int64("endTimeMicros", req.EndTime),
		zap.Int("currentPage", req.Pagination.CurrentPage),
		zap.Int("pageSize", req.Pagination.PageSize),
		zap.String("sortBy", req.SortBy),
		zap.Int("subGroupIDCount", len(req.SubGroupIDs)),
		zap.Strings("subGroupIDs", req.SubGroupIDs),
	)
	events, total, err := h.chRepo.GetAlertEventWithWorkflowRecord(ctx, req, 0)
	if err != nil {
		systemHealthWarnCtx(ctx, "system-health real alert event query failed",
			zap.Int64("groupId", groupID),
			zap.Int64("startTimeMicros", req.StartTime),
			zap.Int64("endTimeMicros", req.EndTime),
			zap.Int("subGroupIDCount", len(req.SubGroupIDs)),
			zap.Duration("elapsed", time.Since(queryStartedAt)),
			zap.Error(err),
		)
		return events, err
	}
	fields := []zap.Field{
		zap.Int64("groupId", groupID),
		zap.Int64("totalCount", total),
		zap.Int("eventCount", len(events)),
		zap.Duration("elapsed", time.Since(queryStartedAt)),
	}
	fields = append(fields, alertEventDistributionFields(events)...)
	systemHealthInfoCtx(ctx, "system-health real alert event query completed", fields...)
	return events, err
}

func (h *handler) queryAlertManagerActiveEvents(ctx core.Context, groupID int64) ([]alertmodel.AEventWithWRecord, error) {
	if strings.TrimSpace(h.alertManagerURL) == "" {
		systemHealthInfoCtx(ctx, "system-health alertmanager active alert query skipped",
			zap.Int64("groupId", groupID),
			zap.String("reason", "not_configured"),
		)
		return nil, nil
	}
	endpoint := strings.TrimRight(h.alertManagerURL, "/") + "/api/v2/alerts"
	queryStartedAt := time.Now()
	systemHealthInfoCtx(ctx, "system-health alertmanager active alert query started",
		zap.Int64("groupId", groupID),
		zap.String("endpoint", endpoint),
	)
	req, err := http.NewRequestWithContext(ctx.GetContext(), http.MethodGet, endpoint, nil)
	if err != nil {
		systemHealthWarnCtx(ctx, "system-health alertmanager active alert query failed",
			zap.Int64("groupId", groupID),
			zap.String("endpoint", endpoint),
			zap.String("reason", "build_request_failed"),
			zap.Duration("elapsed", time.Since(queryStartedAt)),
			zap.Error(err),
		)
		return nil, err
	}
	resp, err := (&http.Client{Timeout: 8 * time.Second}).Do(req)
	if err != nil {
		systemHealthWarnCtx(ctx, "system-health alertmanager active alert query failed",
			zap.Int64("groupId", groupID),
			zap.String("endpoint", endpoint),
			zap.String("reason", "request_failed"),
			zap.Duration("elapsed", time.Since(queryStartedAt)),
			zap.Error(err),
		)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		systemHealthWarnCtx(ctx, "system-health alertmanager active alert query failed",
			zap.Int64("groupId", groupID),
			zap.String("endpoint", endpoint),
			zap.Int("httpStatus", resp.StatusCode),
			zap.String("reason", "bad_status"),
			zap.Duration("elapsed", time.Since(queryStartedAt)),
		)
		return nil, fmt.Errorf("alertmanager returned status %d", resp.StatusCode)
	}

	var alerts []alertManagerAPIAlert
	if err := json.NewDecoder(resp.Body).Decode(&alerts); err != nil {
		systemHealthWarnCtx(ctx, "system-health alertmanager active alert query failed",
			zap.Int64("groupId", groupID),
			zap.String("endpoint", endpoint),
			zap.Int("httpStatus", resp.StatusCode),
			zap.String("reason", "decode_failed"),
			zap.Duration("elapsed", time.Since(queryStartedAt)),
			zap.Error(err),
		)
		return nil, err
	}

	allowedGroupIDs := allowedAlertGroupIDs(groupID)
	events := make([]alertmodel.AEventWithWRecord, 0, len(alerts))
	filteredByState := 0
	filteredByGroup := 0
	missingAlertName := 0
	filteredOther := 0
	for _, alert := range alerts {
		if alert.Status.State != "" && alert.Status.State != "active" && alert.Status.State != alertmodel.StatusFiring {
			filteredByState++
			continue
		}
		if len(allowedGroupIDs) > 0 {
			if _, ok := allowedGroupIDs[alertGroupIDLabel(alert.Labels)]; !ok {
				filteredByGroup++
				continue
			}
		}
		if strings.TrimSpace(alert.Labels["alertname"]) == "" {
			missingAlertName++
			continue
		}
		event, ok := alertManagerAlertToEvent(alert, nil)
		if ok {
			events = append(events, event)
		} else {
			filteredOther++
		}
	}
	systemHealthInfoCtx(ctx, "system-health alertmanager active alert query completed",
		zap.Int64("groupId", groupID),
		zap.String("endpoint", endpoint),
		zap.Int("httpStatus", resp.StatusCode),
		zap.Int("rawAlertCount", len(alerts)),
		zap.Int("filteredAlertCount", len(alerts)-len(events)),
		zap.Int("filteredByStateCount", filteredByState),
		zap.Int("filteredByGroupCount", filteredByGroup),
		zap.Int("missingAlertNameCount", missingAlertName),
		zap.Int("filteredOtherCount", filteredOther),
		zap.Int("eventCount", len(events)),
		zap.String("allowedGroupIDs", systemHealthCompactJSON(stringSetKeys(allowedGroupIDs))),
		zap.Duration("elapsed", time.Since(queryStartedAt)),
	)
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

func (h *handler) queryAlertHistoryStats(ctx core.Context, groupID int64, events []alertmodel.AEventWithWRecord, endTime time.Time) map[string]clickhouse.AlertEventHistoryStats {
	alertIDs := alertIDsFromEvents(events)
	if h == nil || h.chRepo == nil || ctx == nil || len(alertIDs) == 0 {
		systemHealthWarnCtx(ctx, "system-health alert history lookup skipped",
			zap.Int64("groupId", groupID),
			zap.Int("eventCount", len(events)),
			zap.Int("alertIDCount", len(alertIDs)),
			zap.Bool("hasHandler", h != nil),
			zap.String("reason", "missing_repository_context_or_alert_ids"),
		)
		return nil
	}
	startTime := endTime.Add(-72 * time.Hour)
	queryStartedAt := time.Now()
	systemHealthInfoCtx(ctx, "system-health alert history lookup started",
		zap.Int64("groupId", groupID),
		zap.Int("eventCount", len(events)),
		zap.Int("alertIDCount", len(alertIDs)),
		zap.Time("startTime", startTime),
		zap.Time("endTime", endTime),
	)
	stats, err := h.chRepo.GetAlertEventHistoryStatsByAlertIDs(ctx, startTime, endTime, alertIDs)
	if err != nil {
		systemHealthWarnCtx(ctx, "system-health alert history lookup failed",
			zap.Int64("groupId", groupID),
			zap.Int("alertIDCount", len(alertIDs)),
			zap.Time("startTime", startTime),
			zap.Time("endTime", endTime),
			zap.Duration("elapsed", time.Since(queryStartedAt)),
			zap.Error(err),
		)
		return nil
	}
	result := make(map[string]clickhouse.AlertEventHistoryStats, len(stats))
	for _, item := range stats {
		if strings.TrimSpace(item.AlertID) == "" {
			continue
		}
		result[item.AlertID] = item
	}
	systemHealthInfoCtx(ctx, "system-health alert history lookup completed",
		zap.Int64("groupId", groupID),
		zap.Int("alertIDCount", len(alertIDs)),
		zap.Int("historyStatCount", len(result)),
		zap.Time("startTime", startTime),
		zap.Time("endTime", endTime),
		zap.Duration("elapsed", time.Since(queryStartedAt)),
	)
	return result
}

func (h *handler) queryAlertHistorySegments(ctx core.Context, groupID int64, events []alertmodel.AEventWithWRecord, endTime time.Time) map[string][]clickhouse.AlertEventRawSegment {
	alertIDs := alertIDsFromEvents(events)
	if h == nil || h.chRepo == nil || ctx == nil || len(alertIDs) == 0 {
		systemHealthWarnCtx(ctx, "system-health alert history segment lookup skipped",
			zap.Int64("groupId", groupID),
			zap.Int("eventCount", len(events)),
			zap.Int("alertIDCount", len(alertIDs)),
			zap.Bool("hasHandler", h != nil),
			zap.String("reason", "missing_repository_context_or_alert_ids"),
		)
		return nil
	}
	startTime := endTime.Add(-72 * time.Hour)
	queryStartedAt := time.Now()
	systemHealthInfoCtx(ctx, "system-health alert history segment lookup started",
		zap.Int64("groupId", groupID),
		zap.Int("eventCount", len(events)),
		zap.Int("alertIDCount", len(alertIDs)),
		zap.Time("startTime", startTime),
		zap.Time("endTime", endTime),
	)
	segments, err := h.chRepo.GetAlertEventRawSegmentsByAlertIDs(ctx, startTime, endTime, alertIDs)
	if err != nil {
		systemHealthWarnCtx(ctx, "system-health alert history segment lookup failed",
			zap.Int64("groupId", groupID),
			zap.Int("alertIDCount", len(alertIDs)),
			zap.Time("startTime", startTime),
			zap.Time("endTime", endTime),
			zap.Duration("elapsed", time.Since(queryStartedAt)),
			zap.Error(err),
		)
		return nil
	}
	result := make(map[string][]clickhouse.AlertEventRawSegment, len(alertIDs))
	for _, item := range segments {
		alertID := strings.TrimSpace(item.AlertID)
		if alertID == "" {
			continue
		}
		result[alertID] = append(result[alertID], item)
	}
	systemHealthInfoCtx(ctx, "system-health alert history segment lookup completed",
		zap.Int64("groupId", groupID),
		zap.Int("alertIDCount", len(alertIDs)),
		zap.Int("historySegmentCount", len(segments)),
		zap.Int("historySegmentAlertIDCount", len(result)),
		zap.Time("startTime", startTime),
		zap.Time("endTime", endTime),
		zap.Duration("elapsed", time.Since(queryStartedAt)),
	)
	return result
}

func alertIDsFromEvents(events []alertmodel.AEventWithWRecord) []string {
	seen := make(map[string]struct{}, len(events))
	result := make([]string, 0, len(events))
	for _, event := range events {
		alertID := strings.TrimSpace(event.AlertID)
		if alertID == "" {
			continue
		}
		if _, ok := seen[alertID]; ok {
			continue
		}
		seen[alertID] = struct{}{}
		result = append(result, alertID)
	}
	return result
}

func alertHistorySegmentsForInput(events []alertmodel.AEventWithWRecord, history map[string][]clickhouse.AlertEventRawSegment) []clickhouse.AlertEventRawSegment {
	if len(events) == 0 || len(history) == 0 {
		return nil
	}
	result := make([]clickhouse.AlertEventRawSegment, 0, len(events))
	seen := make(map[string]struct{}, len(events))
	for _, event := range events {
		alertID := strings.TrimSpace(event.AlertID)
		if alertID == "" {
			continue
		}
		if _, ok := seen[alertID]; ok {
			continue
		}
		seen[alertID] = struct{}{}
		result = append(result, history[alertID]...)
	}
	sort.SliceStable(result, func(i, j int) bool {
		left := result[i].CreateTime
		if left.IsZero() {
			left = result[i].ReceivedTime
		}
		right := result[j].CreateTime
		if right.IsZero() {
			right = result[j].ReceivedTime
		}
		return left.Before(right)
	})
	return result
}

func alertHistoryStatsForInput(events []alertmodel.AEventWithWRecord, history map[string]clickhouse.AlertEventHistoryStats) []clickhouse.AlertEventHistoryStats {
	if len(events) == 0 || len(history) == 0 {
		return nil
	}
	result := make([]clickhouse.AlertEventHistoryStats, 0, len(events))
	seen := make(map[string]struct{}, len(events))
	for _, event := range events {
		alertID := strings.TrimSpace(event.AlertID)
		if alertID == "" {
			continue
		}
		if _, ok := seen[alertID]; ok {
			continue
		}
		seen[alertID] = struct{}{}
		if stats, ok := history[alertID]; ok {
			result = append(result, stats)
		}
	}
	return result
}

func (h *handler) buildInputsFromRealEvents(ctx core.Context, groupID int64, events []alertmodel.AEventWithWRecord, startTime time.Time, endTime time.Time, capabilities telemetryCapabilityStatus) []reportInput {
	buildStartedAt := time.Now()
	systemHealthInfoCtx(ctx, "system-health report input grouping started",
		zap.Int64("groupId", groupID),
		zap.Int("eventCount", len(events)),
		zap.Time("startTime", startTime),
		zap.Time("endTime", endTime),
	)
	grouped := make(map[string]*reportInput)
	droppedNoKind := 0
	kindCounts := make(map[string]int)
	scopeKindCounts := make(map[string]int)
	for _, event := range events {
		kind := classifyIssueKind(event)
		if kind == "" {
			droppedNoKind++
			continue
		}
		kindCounts[kind]++
		normalized := normalizeAlertEvent(event)
		normalized.IssueKind = kind
		scope, scopeKind := normalizedAggregationScope(normalized, event)
		scopeKindCounts[scopeKind]++
		aggregationKey := normalized.AggregationKey
		if aggregationKey == "" {
			aggregationKey = scopeKind + "::" + scope
		}
		key := aggregationKey
		input, exists := grouped[key]
		if !exists {
			input = &reportInput{
				key:                  stableKey(kind + "::" + aggregationKey),
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
		} else {
			input.kind = mergeIssueKind(input.kind, kind)
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
	systemHealthInfoCtx(ctx, "system-health report input grouping completed",
		zap.Int64("groupId", groupID),
		zap.Int("eventCount", len(events)),
		zap.Int("droppedNoKindCount", droppedNoKind),
		zap.Int("groupedInputCount", len(grouped)),
		zap.String("kindCounts", systemHealthCompactJSON(kindCounts)),
		zap.String("scopeKindCounts", systemHealthCompactJSON(scopeKindCounts)),
		zap.Duration("elapsed", time.Since(buildStartedAt)),
	)

	inputs := make([]reportInput, 0, len(grouped))
	alertHistoryStatsByID := h.queryAlertHistoryStats(ctx, groupID, events, endTime)
	alertHistorySegmentsByID := h.queryAlertHistorySegments(ctx, groupID, events, endTime)
	alertHistoryStartTime := endTime.Add(-72 * time.Hour)
	for _, input := range grouped {
		inputStartedAt := time.Now()
		finalizeReportInput(input, endTime)
		input.alertNames = uniqueStrings(input.alertNames)
		input.alertHistoryStats = alertHistoryStatsForInput(input.events, alertHistoryStatsByID)
		input.alertHistorySegments = alertHistorySegmentsForInput(input.events, alertHistorySegmentsByID)
		input.alertHistoryStartTime = alertHistoryStartTime
		input.alertHistoryEndTime = endTime
		input.key = stableKey(input.kind + "::" + input.aggregationKey)
		input.title = buildAggregatedIssueTitle(*input)
		alertIDCount := len(alertIDsFromEvents(input.events))
		systemHealthInfoCtx(ctx, "system-health report input enrichment started",
			zap.Int64("groupId", groupID),
			zap.String("issueKey", input.key),
			zap.String("aggregationKey", input.aggregationKey),
			zap.String("kind", input.kind),
			zap.String("scopeKind", input.scopeKind),
			zap.String("scope", input.scope),
			zap.Int("eventCount", len(input.events)),
			zap.Int("deduplicatedAlerts", input.deduplicatedAlerts),
			zap.Int("repeatCount", input.repeatCount),
			zap.Int("alertIDCount", alertIDCount),
		)
		if len(input.alertHistorySegments) == 0 {
			reason := "history_segments_not_found"
			if alertIDCount == 0 {
				reason = "alert_id_missing"
			} else if len(input.alertHistoryStats) > 0 {
				reason = "history_segments_empty_stats_available"
			}
			systemHealthWarnCtx(ctx, "system-health alert trend segment fallback",
				zap.Int64("groupId", groupID),
				zap.String("issueKey", input.key),
				zap.String("aggregationKey", input.aggregationKey),
				zap.String("scope", input.scope),
				zap.Int("eventCount", len(input.events)),
				zap.Int("alertIDCount", alertIDCount),
				zap.Int("alertHistoryStats", len(input.alertHistoryStats)),
				zap.String("reason", reason),
			)
		}
		input.alertReportActions, input.alertReportFindings, input.alertReportTopologyRoles = h.queryAlertReportData(ctx, *input, startTime, endTime)
		if len(input.alertReportActions) > 0 || len(input.alertReportFindings) > 0 || len(input.alertReportTopologyRoles) > 0 {
			input.availableCalcSources = append(input.availableCalcSources, "alertReport")
			input.displaySources = append(input.displaySources, sourceLabel("alertReport"))
		}
		applyAlertReportTopologyRoleToInput(ctx, input)
		input.evidence = h.buildRealEvidence(ctx, *input, groupID, startTime, endTime, capabilities)
		input.displaySources = uniqueStrings(append(input.displaySources, displaySourcesFromEvidence(input.evidence)...))
		input.impact = buildRealImpact(*input)
		systemHealthInfoCtx(ctx, "system-health report input enrichment completed",
			zap.Int64("groupId", groupID),
			zap.String("issueKey", input.key),
			zap.String("aggregationKey", input.aggregationKey),
			zap.String("title", input.title),
			zap.String("kind", input.kind),
			zap.String("scopeKind", input.scopeKind),
			zap.Int("eventCount", len(input.events)),
			zap.Int("rawEventCount", input.rawEventCount),
			zap.Int("deduplicatedAlerts", input.deduplicatedAlerts),
			zap.Int("activeAlerts", input.activeAlerts),
			zap.Int("repeatCount", input.repeatCount),
			zap.Int("alertHistoryStats", len(input.alertHistoryStats)),
			zap.Int("alertHistorySegments", len(input.alertHistorySegments)),
			zap.Int("alertReportActions", len(input.alertReportActions)),
			zap.Int("alertReportFindings", len(input.alertReportFindings)),
			zap.Int("alertReportTopologyRoles", len(input.alertReportTopologyRoles)),
			zap.String("alertReportTopologyRole", dominantAlertReportTopologyRole(input.alertReportTopologyRoles).Role),
			zap.Int("evidenceCount", len(input.evidence)),
			zap.Strings("evidenceTypes", evidenceTypes(input.evidence)),
			zap.Strings("displaySources", input.displaySources),
			zap.Duration("elapsed", time.Since(inputStartedAt)),
		)
		inputs = append(inputs, *input)
	}
	systemHealthInfoCtx(ctx, "system-health report input build completed",
		zap.Int64("groupId", groupID),
		zap.Int("inputCount", len(inputs)),
		zap.Duration("elapsed", time.Since(buildStartedAt)),
	)
	return inputs
}

func displaySourcesFromEvidence(evidence []response.SystemHealthEvidence) []string {
	sources := make([]string, 0, 1)
	for _, item := range evidence {
		switch {
		case item.Type == "Trace链路样本" || strings.Contains(item.Type, "APM链路"):
			sources = append(sources, sourceLabel("trace"))
		case strings.Contains(item.Type, "日志"):
			sources = append(sources, sourceLabel("logs"))
		case strings.Contains(item.Type, "请求北极星"):
			sources = append(sources, sourceLabel("requestPolaris"))
		case strings.Contains(item.Type, "指标趋势"):
			sources = append(sources, sourceLabel("redMetrics"))
		case strings.Contains(item.Type, "方向报告"):
			sources = append(sources, sourceLabel("alertReport"))
		}
	}
	return uniqueStrings(sources)
}

func applyAlertReportTopologyRoleToInput(ctx core.Context, input *reportInput) {
	if input == nil {
		return
	}
	role := dominantAlertReportTopologyRole(input.alertReportTopologyRoles)
	if role.Role == "" || role.Role == "unknown" {
		if role.Reason != "" {
			systemHealthInfoCtx(ctx, "system-health alert report topology role not applied",
				zap.String("issueKey", input.key),
				zap.String("aggregationKey", input.aggregationKey),
				zap.String("scopeKind", input.scopeKind),
				zap.String("scope", input.scope),
				zap.String("topologyRole", role.Role),
				zap.String("topologyReason", role.Reason),
			)
		}
		return
	}
	oldScopeKind := input.scopeKind
	oldScope := input.scope
	oldCriticality := input.objectCriticality
	switch role.Role {
	case "entry":
		if input.scopeKind != "service" && role.CurrentService != "" && role.CurrentEndpoint != "" {
			input.scope = role.CurrentService + " / " + role.CurrentEndpoint
		}
		if input.scopeKind != "service" && role.CurrentEndpoint != "" {
			input.scopeKind = "entry"
		}
		input.objectCriticality = math.Max(input.objectCriticality, 90)
	case "internalEndpoint":
		if input.scopeKind != "service" && role.CurrentService != "" && role.CurrentEndpoint != "" {
			input.scope = role.CurrentService + " / " + role.CurrentEndpoint
		}
		if input.scopeKind == "entry" {
			input.scopeKind = "serviceEndpoint"
		}
		input.objectCriticality = math.Max(input.objectCriticality, 85)
	}
	input.mergeReasons = buildMergeReasons(*input)
	if reason := topologyRoleMergeReason(role); reason != "" {
		input.mergeReasons = uniqueStrings(append(input.mergeReasons, reason))
	}
	systemHealthInfoCtx(ctx, "system-health alert report topology role applied",
		zap.String("issueKey", input.key),
		zap.String("aggregationKey", input.aggregationKey),
		zap.String("oldScopeKind", oldScopeKind),
		zap.String("newScopeKind", input.scopeKind),
		zap.String("oldScope", oldScope),
		zap.String("newScope", input.scope),
		zap.Float64("oldObjectCriticality", oldCriticality),
		zap.Float64("newObjectCriticality", input.objectCriticality),
		zap.String("topologyRole", role.Role),
		zap.String("topologyReason", role.Reason),
		zap.String("currentService", role.CurrentService),
		zap.String("currentEndpoint", role.CurrentEndpoint),
		zap.Int("parentCount", role.ParentCount),
		zap.Int("childCount", role.ChildCount),
	)
}

func topologyRoleMergeReason(role systemHealthAlertReportTopologyRole) string {
	switch role.Role {
	case "entry":
		return "方向报告拓扑确认入口侧接口"
	case "internalEndpoint":
		return "方向报告拓扑显示存在上游调用方"
	default:
		return ""
	}
}

func dominantAlertReportTopologyRole(roles []systemHealthAlertReportTopologyRole) systemHealthAlertReportTopologyRole {
	roles = uniqueAlertReportTopologyRoles(roles)
	for _, role := range roles {
		if role.Role == "internalEndpoint" {
			return role
		}
	}
	for _, role := range roles {
		if role.Role == "entry" {
			return role
		}
	}
	for _, role := range roles {
		if role.Role != "" {
			return role
		}
	}
	return systemHealthAlertReportTopologyRole{}
}

func uniqueAlertReportTopologyRoles(roles []systemHealthAlertReportTopologyRole) []systemHealthAlertReportTopologyRole {
	if len(roles) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(roles))
	result := make([]systemHealthAlertReportTopologyRole, 0, len(roles))
	for _, role := range roles {
		key := strings.Join([]string{
			role.Role,
			role.Reason,
			role.CurrentService,
			role.CurrentEndpoint,
			strings.Join(role.ParentNames, ","),
			strings.Join(role.ChildNames, ","),
		}, "\x00")
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, role)
	}
	return result
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
	if event.ServiceName != "" {
		return "service::" + event.ServiceName
	}
	if event.ObjectName != "" {
		return "object::" + event.ObjectType + "::" + event.ObjectName
	}
	return "group::" + firstNonEmpty(event.Group, "unknown")
}

func normalizedAggregationScope(event normalizedAlertEvent, rawEvent alertmodel.AEventWithWRecord) (string, string) {
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
	input.effectiveAlerts = 0
	input.invalidAlerts = 0
	input.activeAlerts = 0
	input.repeatCount = 0
	input.durationSeconds = 0
	input.alertSeverities = input.alertSeverities[:0]
	invalidEvents := normalizedEventsMatchingKind(input.normalizedEvents, input.kind, func(kind string) bool {
		return kind == issueKindInvalidAlert
	})
	effectiveEvents := input.normalizedEvents
	if input.kind != issueKindInvalidAlert {
		effectiveEvents = normalizedEventsMatchingKind(input.normalizedEvents, input.kind, func(kind string) bool {
			return kind != issueKindInvalidAlert
		})
	}
	if len(effectiveEvents) == 0 {
		effectiveEvents = input.normalizedEvents
	}
	effectiveInstances := deduplicateAlertInstances(effectiveEvents, endTime)
	input.effectiveAlerts = len(effectiveInstances)
	input.invalidAlerts = len(deduplicateAlertInstances(invalidEvents, endTime))
	activeInstances := deduplicateAlertInstances(normalizedEventsMatchingKind(input.normalizedEvents, input.kind, func(kind string) bool {
		return kind == issueKindActiveAlert
	}), endTime)
	input.activeAlerts = len(activeInstances)

	for _, instance := range input.deduplicatedInstances {
		input.repeatCount += maxInt(instance.RepeatCount-1, 0)
		input.durationSeconds = math.Max(input.durationSeconds, instance.DurationSeconds)
	}
	for _, instance := range effectiveInstances {
		input.alertSeverities = append(input.alertSeverities, instance.MaxSeverity)
	}
	if input.deduplicatedAlerts == 0 {
		input.deduplicatedAlerts = len(input.events)
		input.effectiveAlerts = input.deduplicatedAlerts
		if input.kind == issueKindActiveAlert {
			input.activeAlerts = len(input.events)
		}
	}
	if len(input.alertSeverities) == 0 {
		for _, instance := range input.deduplicatedInstances {
			input.alertSeverities = append(input.alertSeverities, instance.MaxSeverity)
		}
	}
	input.mergeReasons = buildMergeReasons(*input)
}

func normalizedEventsMatchingKind(events []normalizedAlertEvent, fallbackKind string, matches func(string) bool) []normalizedAlertEvent {
	result := make([]normalizedAlertEvent, 0, len(events))
	for _, event := range events {
		kind := event.IssueKind
		if kind == "" {
			kind = fallbackKind
		}
		if matches(kind) {
			result = append(result, event)
		}
	}
	return result
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
		reasons = append(reasons, "同一入口侧接口")
	} else if input.scopeKind == "serviceEndpoint" {
		reasons = append(reasons, "同一服务接口")
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
	if input.kind != issueKindInvalidAlert && input.invalidAlerts > 0 {
		reasons = append(reasons, "包含告警规则诊断")
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

func mergeIssueKind(current string, next string) string {
	if current == "" {
		return next
	}
	if issueKindPriority(next) > issueKindPriority(current) {
		return next
	}
	return current
}

func issueKindPriority(kind string) int {
	switch kind {
	case issueKindActiveAlert:
		return 4
	case issueKindInspection:
		return 3
	case issueKindRecoveredAlert:
		return 2
	case issueKindInvalidAlert:
		return 1
	default:
		return 0
	}
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
	if kind == issueKindInvalidAlert {
		return "疑似告警规则问题：" + event.Name
	}
	return event.Name
}

func buildAggregatedIssueTitle(input reportInput) string {
	baseTitle := buildRiskIssueTitle(input)
	if input.kind == issueKindInspection {
		return truncateText(baseTitle, 48)
	}
	if input.kind == issueKindInvalidAlert {
		return truncateText(baseTitle, 48)
	}
	if input.kind == issueKindRecoveredAlert {
		return truncateText("已恢复："+baseTitle, 48)
	}
	return truncateText(baseTitle, 48)
}

func buildRiskIssueTitle(input reportInput) string {
	target := issueTitleTarget(input)
	problem := issueTitleProblem(input)
	switch problem {
	case "请求延时":
		if issueTitleSuggestsIncrease(input) {
			return target + " 请求延时上升"
		}
		return target + " 存在请求延时问题"
	case "错误异常":
		if issueTitleSuggestsIncrease(input) {
			return target + " 错误异常上升"
		}
		return target + " 存在错误异常问题"
	case "资源压力":
		return target + " 存在资源压力问题"
	case "可用性":
		return target + " 存在可用性问题"
	case "告警规则":
		return target + " 存在告警规则问题"
	case "巡检指标":
		return target + " 存在巡检指标异常"
	default:
		alertName := firstString(uniqueStrings(input.alertNames))
		if alertName != "" && !strings.Contains(alertName, "告警汇总") {
			return target + " 存在" + alertName + "问题"
		}
		return target + " 存在健康风险"
	}
}

func issueTitleTarget(input reportInput) string {
	target := strings.TrimSpace(input.scope)
	if target == "" {
		target = firstNonEmpty(input.aggregationKey, input.title, "当前影响目标")
	}
	return target
}

func issueTitleProblem(input reportInput) string {
	if input.kind == issueKindInspection {
		return "巡检指标"
	}
	if input.kind == issueKindInvalidAlert {
		return "告警规则"
	}
	text := strings.ToLower(strings.Join(append([]string{
		input.title,
		input.scope,
		input.impact,
	}, input.alertNames...), " "))
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

func issueTitleSuggestsIncrease(input reportInput) bool {
	text := strings.ToLower(strings.Join(append([]string{input.title}, input.alertNames...), " "))
	return strings.Contains(text, "升高") ||
		strings.Contains(text, "上升") ||
		strings.Contains(text, "增加") ||
		strings.Contains(text, "变高") ||
		strings.Contains(text, "increase") ||
		strings.Contains(text, "high")
}

func firstString(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func buildAlertEventEvidenceResult(input reportInput) string {
	if input.kind == issueKindInvalidAlert {
		return buildInvalidAlertEventEvidenceResult(input)
	}
	alertNames := uniqueStrings(input.alertNames)
	directions := alertDirections(input.events)
	scopeLabel := inputScopeLabel(input.scopeKind)
	symptom := issueSymptomText(scoredIssue{input: input})
	parts := []string{
		fmt.Sprintf("%s在当前窗口%s，%s。", scopeLabel, alertEventFrequencyText(len(input.events), input.repeatCount, input.deduplicatedAlerts, symptom), alertRecoveryText(input)),
	}
	if len(alertNames) > 1 {
		parts = append(parts, fmt.Sprintf("多类%s告警共同指向同一%s，说明风险不是单条告警孤立触发。", symptom, inputScopeNoun(input.scopeKind)))
	}
	if input.repeatCount > 0 {
		parts = append(parts, "同一告警对象存在反复触发，恢复后仍需要重点观察是否复发。")
	} else if len(input.events) >= 3 && input.activeAlerts == 0 {
		parts = append(parts, "虽然当前已恢复，但窗口内出现频率偏高，建议纳入恢复观察。")
	}
	if input.durationSeconds > 0 {
		parts = append(parts, "最长一次持续约 "+formatDurationCN(input.durationSeconds)+"。")
	}
	if len(directions) > 0 {
		parts = append(parts, "已有告警分析指向："+strings.Join(directions, "、")+"。")
	}
	if diagnostic := ruleDiagnosticEvidenceText(input); diagnostic != "" {
		parts = append(parts, diagnostic)
	}
	return truncateText(strings.Join(parts, ""), 360)
}

func buildAlertHistoryEvidenceResult(input reportInput) string {
	if len(input.alertHistoryStats) == 0 {
		return ""
	}
	stats := input.alertHistoryStats
	sort.SliceStable(stats, func(i, j int) bool {
		if stats[i].RecordCount == stats[j].RecordCount {
			return stats[i].LastSeenAt.After(stats[j].LastSeenAt)
		}
		return stats[i].RecordCount > stats[j].RecordCount
	})
	totalRecords := uint64(0)
	totalFiring := uint64(0)
	totalResolved := uint64(0)
	maxDuration := 0.0
	totalDuration := 0.0
	var firstSeen time.Time
	var lastSeen time.Time
	latestStatus := ""
	for _, item := range stats {
		totalRecords += item.RecordCount
		totalFiring += item.FiringCount
		totalResolved += item.ResolvedCount
		maxDuration = math.Max(maxDuration, item.MaxDurationSeconds)
		totalDuration += item.TotalDurationSeconds
		if !item.FirstSeenAt.IsZero() && (firstSeen.IsZero() || item.FirstSeenAt.Before(firstSeen)) {
			firstSeen = item.FirstSeenAt
		}
		if !item.LastSeenAt.IsZero() && (lastSeen.IsZero() || item.LastSeenAt.After(lastSeen)) {
			lastSeen = item.LastSeenAt
			latestStatus = item.LatestStatus
		}
	}
	if totalRecords == 0 {
		return ""
	}
	scopeLabel := inputScopeLabel(input.scopeKind)
	parts := []string{
		fmt.Sprintf("过去 3 天相同告警共出现 %d 次", totalRecords),
	}
	if len(stats) > 1 {
		parts[0] = fmt.Sprintf("过去 3 天该%s涉及的相同告警共出现 %d 次，覆盖 %d 个告警实例", inputScopeNoun(input.scopeKind), totalRecords, len(stats))
	}
	if totalFiring > 0 || totalResolved > 0 {
		parts = append(parts, fmt.Sprintf("其中触发记录 %d 次、恢复记录 %d 次", totalFiring, totalResolved))
	}
	if !firstSeen.IsZero() && !lastSeen.IsZero() {
		parts = append(parts, fmt.Sprintf("首次出现于 %s，最近出现于 %s", formatEventTime(firstSeen), formatEventTime(lastSeen)))
	}
	if maxDuration > 0 {
		parts = append(parts, "最长一次持续约 "+formatDurationCN(maxDuration))
	}
	if totalDuration > 0 && totalRecords > 1 {
		parts = append(parts, "累计持续约 "+formatDurationCN(totalDuration))
	}
	if latestStatus != "" {
		parts = append(parts, "最近状态为 "+userFacingAlertStatus(latestStatus))
	}
	if totalRecords >= 10 {
		parts = append(parts, scopeLabel+"存在明显反复触发，需要重点关注复发规律和规则口径")
	} else if totalRecords > 1 {
		parts = append(parts, scopeLabel+"近期并非首次出现，恢复观察时需要关注是否再次复发")
	}
	return truncateText(strings.Join(parts, "；")+"。", 360)
}

func buildAlertTrend(input reportInput) *response.SystemHealthAlertTrend {
	if len(input.alertHistorySegments) == 0 && len(input.alertHistoryStats) == 0 {
		return nil
	}
	windowEnd := input.alertHistoryEndTime
	if windowEnd.IsZero() {
		windowEnd = latestAlertHistoryTime(input)
	}
	if windowEnd.IsZero() {
		windowEnd = parseDisplayTime(input.updatedAt)
	}
	if windowEnd.IsZero() {
		windowEnd = time.Now()
	}
	windowStart := input.alertHistoryStartTime
	if windowStart.IsZero() {
		windowStart = windowEnd.Add(-72 * time.Hour)
	}

	segments := buildAlertTrendSegments(input, windowStart, windowEnd)
	stats := buildAlertTrendStats(input, segments)
	if stats.occurrenceCount == 0 && len(segments) == 0 {
		return nil
	}
	scopeLabel := alertTrendScopeLabel(input.scopeKind)
	symptom := issueSymptomText(scoredIssue{input: input})
	typeText := alertTrendTypeText(stats.alertTypes, symptom)
	status := "当前已恢复"
	if input.activeAlerts > 0 {
		status = fmt.Sprintf("仍有 %d 个告警处于触发状态", input.activeAlerts)
	}

	summaryParts := []string{
		fmt.Sprintf("最近 3 天，%s相关告警出现 %d 次", scopeLabel, stats.occurrenceCount),
	}
	if typeText != "" {
		summaryParts = append(summaryParts, "主要表现为"+typeText)
	}
	if stats.longestDurationSeconds > 0 {
		summaryParts = append(summaryParts, "最长一次持续约 "+formatDurationCN(stats.longestDurationSeconds))
	}
	summaryParts = append(summaryParts, status)

	activityText := alertTrendActivityText(stats.occurrenceCount, len(segments), input.activeAlerts)
	problem := alertTrendProblemText(scopeLabel, stats.occurrenceCount, stats.longestDurationSeconds, input.activeAlerts)
	return &response.SystemHealthAlertTrend{
		Window:                 "最近 3 天",
		WindowText:             "最近 3 天",
		StartTime:              formatAlertTrendTime(windowStart),
		EndTime:                formatAlertTrendTime(windowEnd),
		OccurrenceCount:        stats.occurrenceCount,
		EventCount:             stats.occurrenceCount,
		TotalCount:             stats.occurrenceCount,
		RawEventCount:          len(segments),
		LongestDurationSeconds: stats.longestDurationSeconds,
		MaxDurationSeconds:     stats.longestDurationSeconds,
		TotalDurationSeconds:   stats.totalDurationSeconds,
		Summary:                truncateText(strings.Join(summaryParts, "；")+"。", 260),
		ActivityText:           activityText,
		Problem:                problem,
		AlertTypes:             stats.alertTypes,
		Segments:               segments,
	}
}

type systemHealthAlertTrendStats struct {
	occurrenceCount        int
	longestDurationSeconds float64
	totalDurationSeconds   float64
	alertTypes             []response.SystemHealthAlertTrendType
}

type alertTrendEventMeta struct {
	Name     string
	Severity string
}

type alertTrendTimelineSegment struct {
	AlertID         string
	AlertName       string
	Severity        string
	Status          string
	Start           time.Time
	End             time.Time
	Count           int
	DurationSeconds float64
	IsActive        bool
}

func buildAlertTrendSegments(input reportInput, windowStart time.Time, windowEnd time.Time) []response.SystemHealthAlertTrendSegment {
	metaByAlertID := alertTrendEventMetaByID(input.events)
	timeline := buildAlertTrendTimelineSegments(input.alertHistorySegments, metaByAlertID, windowStart, windowEnd)
	timeline = mergeAlertTrendTimelineSegmentsForRiskPoint(timeline)
	segments := make([]response.SystemHealthAlertTrendSegment, 0, len(timeline))
	for _, item := range timeline {
		if item.Start.IsZero() && item.End.IsZero() {
			continue
		}
		alertName := strings.TrimSpace(item.AlertName)
		severity := strings.TrimSpace(item.Severity)
		alertType := alertTrendType(alertName)
		status := normalizeStatus(item.Status)
		summary := alertTrendSegmentSummary(alertName, alertType, status, item.DurationSeconds)
		segments = append(segments, response.SystemHealthAlertTrendSegment{
			AlertID:         strings.TrimSpace(item.AlertID),
			AlertName:       alertName,
			RuleName:        alertName,
			Title:           firstNonEmpty(alertName, alertType, "相关告警"),
			Summary:         summary,
			Type:            alertType,
			Severity:        severity,
			Status:          status,
			StartTime:       formatAlertTrendTime(item.Start),
			StartedAt:       formatAlertTrendTime(item.Start),
			EndTime:         formatAlertTrendTime(item.End),
			EndedAt:         formatAlertTrendTime(item.End),
			DurationSeconds: item.DurationSeconds,
			Count:           item.Count,
			EventCount:      item.Count,
			AlertCount:      item.Count,
			IsActive:        item.IsActive,
		})
	}
	if len(segments) > 0 {
		return segments
	}
	return buildAlertTrendSegmentsFromStats(input, metaByAlertID, windowStart, windowEnd)
}

func mergeAlertTrendTimelineSegmentsForRiskPoint(segments []alertTrendTimelineSegment) []alertTrendTimelineSegment {
	if len(segments) <= 1 {
		return segments
	}
	sort.SliceStable(segments, func(i, j int) bool {
		if segments[i].Start.Equal(segments[j].Start) {
			return segments[i].End.Before(segments[j].End)
		}
		return segments[i].Start.Before(segments[j].Start)
	})

	merged := make([]alertTrendTimelineSegment, 0, len(segments))
	current := segments[0]
	for _, next := range segments[1:] {
		if !alertTrendTimelineSegmentsOverlap(current, next) {
			merged = append(merged, current)
			current = next
			continue
		}
		current = mergeAlertTrendTimelineSegment(current, next)
	}
	merged = append(merged, current)
	return merged
}

func alertTrendTimelineSegmentsOverlap(left alertTrendTimelineSegment, right alertTrendTimelineSegment) bool {
	if left.Start.IsZero() || left.End.IsZero() || right.Start.IsZero() || right.End.IsZero() {
		return false
	}
	return !right.Start.After(left.End)
}

func mergeAlertTrendTimelineSegment(left alertTrendTimelineSegment, right alertTrendTimelineSegment) alertTrendTimelineSegment {
	if !right.Start.IsZero() && (left.Start.IsZero() || right.Start.Before(left.Start)) {
		left.Start = right.Start
	}
	if right.End.After(left.End) {
		left.End = right.End
	}
	left.Count += maxInt(right.Count, 1)
	left.DurationSeconds = left.End.Sub(left.Start).Seconds()
	if left.DurationSeconds <= 0 {
		left.DurationSeconds = 60
	}
	left.AlertID = mergeAlertTrendText(left.AlertID, right.AlertID, ",")
	left.AlertName = mergeAlertTrendText(left.AlertName, right.AlertName, "、")
	left.Severity = maxSeverity([]string{left.Severity, right.Severity})
	left.IsActive = left.IsActive || right.IsActive
	if left.IsActive {
		left.Status = alertmodel.StatusFiring
	} else if normalizeStatus(left.Status) == "" || normalizeStatus(left.Status) == "unknown" {
		left.Status = right.Status
	} else if normalizeStatus(right.Status) == alertmodel.StatusResolved {
		left.Status = alertmodel.StatusResolved
	}
	return left
}

func mergeAlertTrendText(left string, right string, sep string) string {
	values := make([]string, 0, 4)
	for _, value := range []string{left, right} {
		for _, part := range strings.Split(value, sep) {
			part = strings.TrimSpace(part)
			if part != "" {
				values = append(values, part)
			}
		}
	}
	return strings.Join(uniqueStrings(values), sep)
}

func buildAlertTrendTimelineSegments(rawSegments []clickhouse.AlertEventRawSegment, metaByAlertID map[string]alertTrendEventMeta, windowStart time.Time, windowEnd time.Time) []alertTrendTimelineSegment {
	if len(rawSegments) == 0 {
		return nil
	}
	byAlertID := make(map[string][]clickhouse.AlertEventRawSegment, len(rawSegments))
	for _, raw := range rawSegments {
		alertID := strings.TrimSpace(raw.AlertID)
		if alertID == "" {
			continue
		}
		byAlertID[alertID] = append(byAlertID[alertID], raw)
	}
	alertIDs := make([]string, 0, len(byAlertID))
	for alertID := range byAlertID {
		alertIDs = append(alertIDs, alertID)
	}
	sort.Strings(alertIDs)

	result := make([]alertTrendTimelineSegment, 0, len(rawSegments))
	for _, alertID := range alertIDs {
		records := byAlertID[alertID]
		sort.SliceStable(records, func(i, j int) bool {
			left := alertTrendTransitionTime(records[i])
			right := alertTrendTransitionTime(records[j])
			if left.Equal(right) {
				return alertTrendRecordStart(records[i]).Before(alertTrendRecordStart(records[j]))
			}
			return left.Before(right)
		})

		var open *alertTrendTimelineSegment
		for _, raw := range records {
			status := normalizeStatus(raw.Status)
			start := alertTrendRecordStart(raw)
			recovery := alertTrendRecordRecovery(raw)
			name := strings.TrimSpace(raw.Name)
			severity := strings.TrimSpace(raw.Severity)
			if meta, ok := metaByAlertID[alertID]; ok {
				if name == "" {
					name = meta.Name
				}
				if severity == "" {
					severity = meta.Severity
				}
			}

			switch status {
			case alertmodel.StatusFiring:
				if start.IsZero() {
					start = alertTrendTransitionTime(raw)
				}
				if open == nil {
					open = &alertTrendTimelineSegment{
						AlertID:   alertID,
						AlertName: name,
						Severity:  severity,
						Status:    alertmodel.StatusFiring,
						Start:     start,
						Count:     1,
						IsActive:  true,
					}
				} else {
					if !start.IsZero() && (open.Start.IsZero() || start.Before(open.Start)) {
						open.Start = start
					}
					open.Count++
				}
				if open.AlertName == "" {
					open.AlertName = name
				}
				if open.Severity == "" {
					open.Severity = severity
				}
			case alertmodel.StatusResolved:
				if recovery.IsZero() {
					recovery = alertTrendTransitionTime(raw)
				}
				if open == nil {
					if start.IsZero() {
						start = recovery
					}
					open = &alertTrendTimelineSegment{
						AlertID:   alertID,
						AlertName: name,
						Severity:  severity,
						Status:    alertmodel.StatusResolved,
						Start:     start,
						Count:     1,
					}
				} else {
					open.Status = alertmodel.StatusResolved
					open.IsActive = false
					open.Count++
					if !start.IsZero() && (open.Start.IsZero() || start.Before(open.Start)) {
						open.Start = start
					}
					if open.AlertName == "" {
						open.AlertName = name
					}
					if open.Severity == "" {
						open.Severity = severity
					}
				}
				open.End = recovery
				result = appendAlertTrendTimelineSegment(result, open, windowStart, windowEnd)
				open = nil
			default:
				start, end, _ := alertTrendSegmentTimes(raw, windowStart, windowEnd)
				result = appendAlertTrendTimelineSegment(result, &alertTrendTimelineSegment{
					AlertID:   alertID,
					AlertName: name,
					Severity:  severity,
					Status:    status,
					Start:     start,
					End:       end,
					Count:     1,
				}, windowStart, windowEnd)
			}
		}
		if open != nil {
			open.Status = alertmodel.StatusFiring
			open.IsActive = true
			open.End = windowEnd
			result = appendAlertTrendTimelineSegment(result, open, windowStart, windowEnd)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Start.Equal(result[j].Start) {
			return result[i].AlertID < result[j].AlertID
		}
		return result[i].Start.Before(result[j].Start)
	})
	return result
}

func appendAlertTrendTimelineSegment(result []alertTrendTimelineSegment, segment *alertTrendTimelineSegment, windowStart time.Time, windowEnd time.Time) []alertTrendTimelineSegment {
	if segment == nil {
		return result
	}
	start := segment.Start
	end := segment.End
	if start.IsZero() {
		start = windowStart
	}
	if end.IsZero() || end.Before(start) || end.Equal(start) {
		end = start.Add(time.Minute)
	}
	if !windowStart.IsZero() && start.Before(windowStart) {
		start = windowStart
	}
	if !windowEnd.IsZero() && end.After(windowEnd) {
		end = windowEnd
	}
	if !windowStart.IsZero() && end.Before(windowStart) {
		return result
	}
	if !windowEnd.IsZero() && start.After(windowEnd) {
		return result
	}
	if !end.After(start) {
		end = start.Add(time.Minute)
		if !windowEnd.IsZero() && end.After(windowEnd) {
			end = windowEnd
		}
	}
	duration := end.Sub(start).Seconds()
	if duration <= 0 {
		duration = 60
	}
	segment.Start = start
	segment.End = end
	segment.DurationSeconds = duration
	if segment.Count <= 0 {
		segment.Count = 1
	}
	return append(result, *segment)
}

func alertTrendTransitionTime(raw clickhouse.AlertEventRawSegment) time.Time {
	if normalizeStatus(raw.Status) == alertmodel.StatusResolved {
		if recovery := alertTrendRecordRecovery(raw); !recovery.IsZero() {
			return recovery
		}
	}
	if start := alertTrendRecordStart(raw); !start.IsZero() {
		return start
	}
	return firstNonZeroTime(raw.ReceivedTime, raw.UpdateTime, raw.EndTime)
}

func alertTrendRecordStart(raw clickhouse.AlertEventRawSegment) time.Time {
	return firstNonZeroTime(raw.CreateTime, raw.ReceivedTime, raw.UpdateTime)
}

func alertTrendRecordRecovery(raw clickhouse.AlertEventRawSegment) time.Time {
	if !raw.EndTime.IsZero() && raw.EndTime.After(time.Date(1971, 1, 1, 0, 0, 0, 0, raw.EndTime.Location())) {
		return raw.EndTime
	}
	return firstNonZeroTime(raw.UpdateTime, raw.ReceivedTime)
}

func buildAlertTrendSegmentsFromStats(input reportInput, metaByAlertID map[string]alertTrendEventMeta, windowStart time.Time, windowEnd time.Time) []response.SystemHealthAlertTrendSegment {
	segments := make([]response.SystemHealthAlertTrendSegment, 0, len(input.alertHistoryStats))
	for _, stat := range input.alertHistoryStats {
		if stat.RecordCount == 0 {
			continue
		}
		start := stat.LastSeenAt
		if start.IsZero() {
			start = stat.FirstSeenAt
		}
		duration := stat.MaxDurationSeconds
		if duration > 0 && !start.IsZero() {
			start = start.Add(-time.Duration(duration) * time.Second)
		}
		if start.IsZero() {
			start = windowStart
		}
		end := stat.LastSeenAt
		if end.IsZero() || end.Before(start) {
			end = start.Add(time.Duration(math.Max(duration, 60)) * time.Second)
		}
		if end.After(windowEnd) {
			end = windowEnd
		}
		if start.Before(windowStart) {
			start = windowStart
		}
		alertName := ""
		severity := ""
		if meta, ok := metaByAlertID[strings.TrimSpace(stat.AlertID)]; ok {
			alertName = meta.Name
			severity = meta.Severity
		}
		alertType := alertTrendType(alertName)
		status := normalizeStatus(stat.LatestStatus)
		count := int(stat.RecordCount)
		segments = append(segments, response.SystemHealthAlertTrendSegment{
			AlertID:         strings.TrimSpace(stat.AlertID),
			AlertName:       alertName,
			RuleName:        alertName,
			Title:           firstNonEmpty(alertName, alertType, "相关告警"),
			Summary:         alertTrendSegmentSummary(alertName, alertType, status, duration),
			Type:            alertType,
			Severity:        severity,
			Status:          status,
			StartTime:       formatAlertTrendTime(start),
			StartedAt:       formatAlertTrendTime(start),
			EndTime:         formatAlertTrendTime(end),
			EndedAt:         formatAlertTrendTime(end),
			DurationSeconds: duration,
			Count:           count,
			EventCount:      count,
			AlertCount:      count,
			IsActive:        status == alertmodel.StatusFiring,
		})
	}
	return segments
}

func alertTrendSegmentTimes(raw clickhouse.AlertEventRawSegment, windowStart time.Time, windowEnd time.Time) (time.Time, time.Time, float64) {
	start := raw.CreateTime
	if start.IsZero() {
		start = firstNonZeroTime(raw.ReceivedTime, raw.UpdateTime)
	}
	end := raw.EndTime
	status := normalizeStatus(raw.Status)
	if end.IsZero() || end.Before(start) || end.Equal(start) {
		switch {
		case status == alertmodel.StatusFiring && !windowEnd.IsZero():
			end = windowEnd
		case !raw.UpdateTime.IsZero() && raw.UpdateTime.After(start):
			end = raw.UpdateTime
		case !raw.ReceivedTime.IsZero() && raw.ReceivedTime.After(start):
			end = raw.ReceivedTime
		}
	}
	duration := raw.DurationSeconds
	if duration <= 0 && !start.IsZero() && !end.IsZero() && end.After(start) {
		duration = end.Sub(start).Seconds()
	}
	if !start.IsZero() && (end.IsZero() || !end.After(start)) {
		end = start.Add(time.Duration(math.Max(duration, 60)) * time.Second)
	}
	if !windowEnd.IsZero() && end.After(windowEnd) {
		end = windowEnd
	}
	if duration <= 0 && !start.IsZero() && !end.IsZero() && end.After(start) {
		duration = end.Sub(start).Seconds()
	}
	if duration <= 0 {
		duration = 60
	}
	_ = windowStart
	return start, end, duration
}

func buildAlertTrendStats(input reportInput, segments []response.SystemHealthAlertTrendSegment) systemHealthAlertTrendStats {
	stats := systemHealthAlertTrendStats{}
	typeCounts := make(map[string]int)
	for _, stat := range input.alertHistoryStats {
		stats.occurrenceCount += int(stat.RecordCount)
		stats.longestDurationSeconds = math.Max(stats.longestDurationSeconds, stat.MaxDurationSeconds)
		stats.totalDurationSeconds += stat.TotalDurationSeconds
	}
	if stats.occurrenceCount == 0 {
		for _, segment := range segments {
			count := maxInt(segment.Count, 1)
			stats.occurrenceCount += count
			stats.longestDurationSeconds = math.Max(stats.longestDurationSeconds, segment.DurationSeconds)
			stats.totalDurationSeconds += segment.DurationSeconds * float64(count)
		}
	}
	for _, segment := range segments {
		typeName := firstNonEmpty(segment.Type, alertTrendType(segment.AlertName), "告警信号")
		typeCounts[typeName] += maxInt(segment.Count, 1)
	}
	if len(typeCounts) == 0 {
		for _, alertName := range uniqueStrings(input.alertNames) {
			typeCounts[alertTrendType(alertName)]++
		}
	}
	stats.alertTypes = alertTrendTypesFromCounts(typeCounts)
	return stats
}

func alertTrendTypesFromCounts(typeCounts map[string]int) []response.SystemHealthAlertTrendType {
	types := make([]response.SystemHealthAlertTrendType, 0, len(typeCounts))
	for name, count := range typeCounts {
		if strings.TrimSpace(name) == "" || count <= 0 {
			continue
		}
		types = append(types, response.SystemHealthAlertTrendType{Name: name, Count: count})
	}
	sort.SliceStable(types, func(i, j int) bool {
		if types[i].Count == types[j].Count {
			return types[i].Name < types[j].Name
		}
		return types[i].Count > types[j].Count
	})
	if len(types) > 5 {
		return types[:5]
	}
	return types
}

func alertTrendEventMetaByID(events []alertmodel.AEventWithWRecord) map[string]alertTrendEventMeta {
	result := make(map[string]alertTrendEventMeta, len(events))
	for _, event := range events {
		alertID := strings.TrimSpace(event.AlertID)
		if alertID == "" {
			continue
		}
		meta := result[alertID]
		if meta.Name == "" {
			meta.Name = strings.TrimSpace(event.Name)
		}
		if meta.Severity == "" {
			meta.Severity = strings.TrimSpace(event.Severity)
		}
		result[alertID] = meta
	}
	return result
}

func alertTrendType(value string) string {
	text := strings.ToLower(strings.TrimSpace(value))
	switch {
	case strings.Contains(text, "latency"), strings.Contains(text, "duration"), strings.Contains(text, "timeout"), strings.Contains(text, "slow"), strings.Contains(text, "p95"), strings.Contains(text, "p99"), strings.Contains(text, "延时"), strings.Contains(text, "延迟"), strings.Contains(text, "超时"), strings.Contains(text, "慢"):
		return "请求延时"
	case strings.Contains(text, "error"), strings.Contains(text, "exception"), strings.Contains(text, "fail"), strings.Contains(text, "5xx"), strings.Contains(text, "4xx"), strings.Contains(text, "错误"), strings.Contains(text, "异常"), strings.Contains(text, "失败"):
		return "错误率"
	case strings.Contains(text, "log"), strings.Contains(text, "日志"):
		return "日志异常"
	case strings.Contains(text, "cpu"), strings.Contains(text, "memory"), strings.Contains(text, "oom"), strings.Contains(text, "disk"), strings.Contains(text, "load"), strings.Contains(text, "资源"), strings.Contains(text, "内存"), strings.Contains(text, "磁盘"), strings.Contains(text, "负载"):
		return "资源压力"
	case strings.Contains(text, "available"), strings.Contains(text, "unavailable"), strings.Contains(text, "down"), strings.Contains(text, "不可用"):
		return "可用性"
	default:
		return "告警信号"
	}
}

func alertTrendTypeText(types []response.SystemHealthAlertTrendType, fallback string) string {
	names := make([]string, 0, len(types)+1)
	for _, item := range types {
		if item.Name == "" || item.Name == "告警信号" {
			continue
		}
		names = append(names, item.Name)
	}
	if len(names) == 0 && strings.TrimSpace(fallback) != "" {
		names = append(names, strings.TrimSpace(fallback))
	}
	if len(names) == 0 {
		return ""
	}
	return strings.Join(uniqueStrings(names)[:minInt(len(uniqueStrings(names)), 3)], "、")
}

func alertTrendActivityText(occurrenceCount int, segmentCount int, activeAlerts int) string {
	switch {
	case activeAlerts > 0 && occurrenceCount >= 10:
		return "仍在触发且近期高频出现"
	case activeAlerts > 0:
		return "当前仍在触发"
	case occurrenceCount >= 10 || segmentCount >= 8:
		return "近期高频出现"
	case occurrenceCount >= 3 || segmentCount >= 3:
		return "近期多次出现"
	case occurrenceCount > 1:
		return "近期重复出现"
	default:
		return "近期少量出现"
	}
}

func alertTrendProblemText(scopeLabel string, occurrenceCount int, longestDurationSeconds float64, activeAlerts int) string {
	if activeAlerts > 0 {
		return fmt.Sprintf("%s仍有告警触发，需要优先确认当前业务请求是否受影响。", scopeLabel)
	}
	if occurrenceCount >= 10 {
		return fmt.Sprintf("%s近期反复出现同类告警，恢复后仍需要观察是否再次集中出现。", scopeLabel)
	}
	if longestDurationSeconds >= 30*60 {
		return fmt.Sprintf("%s曾出现持续时间较长的告警，建议结合指标趋势和告警分析确认是否已经稳定恢复。", scopeLabel)
	}
	return fmt.Sprintf("%s当前未持续告警，但近期出现过相关信号，建议纳入恢复观察。", scopeLabel)
}

func alertTrendScopeLabel(scopeKind string) string {
	switch scopeKind {
	case "entry", "serviceEndpoint":
		return "该接口"
	case "service":
		return "该服务"
	case "node":
		return "该节点"
	case "pod":
		return "该实例"
	case "database":
		return "该数据库"
	default:
		return "该影响目标"
	}
}

func alertTrendSegmentSummary(alertName string, alertType string, status string, durationSeconds float64) string {
	subject := firstNonEmpty(alertName, alertType, "相关告警")
	parts := []string{subject}
	if alertType != "" && alertType != "告警信号" && !strings.Contains(subject, alertType) {
		parts = append(parts, "表现为"+alertType)
	}
	if durationSeconds > 0 {
		parts = append(parts, "持续约"+formatDurationCN(durationSeconds))
	}
	if status != "" {
		parts = append(parts, userFacingAlertStatus(status))
	}
	return strings.Join(parts, "，")
}

func latestAlertHistoryTime(input reportInput) time.Time {
	var latest time.Time
	consider := func(value time.Time) {
		if !value.IsZero() && (latest.IsZero() || value.After(latest)) {
			latest = value
		}
	}
	for _, segment := range input.alertHistorySegments {
		consider(segment.ReceivedTime)
		consider(segment.UpdateTime)
		consider(segment.EndTime)
	}
	for _, stat := range input.alertHistoryStats {
		consider(stat.LastSeenAt)
	}
	return latest
}

func formatAlertTrendTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(time.RFC3339Nano)
}

func firstNonZeroTime(values ...time.Time) time.Time {
	for _, value := range values {
		if !value.IsZero() {
			return value
		}
	}
	return time.Time{}
}

func userFacingAlertStatus(status string) string {
	switch normalizeStatus(status) {
	case alertmodel.StatusFiring:
		return "仍在触发"
	case alertmodel.StatusResolved:
		return "已恢复"
	default:
		return firstNonEmpty(status, "未知")
	}
}

func buildInvalidAlertEventEvidenceResult(input reportInput) string {
	scopeLabel := inputScopeLabel(input.scopeKind)
	count := maxInt(input.invalidAlerts, input.deduplicatedAlerts)
	if count <= 0 {
		count = len(input.events)
	}
	parts := []string{
		fmt.Sprintf("%s在当前窗口触发过告警，但已有分析更倾向将其作为告警规则诊断信号处理。", scopeLabel),
		"需要对比告警指标与服务 RED 指标是否同步变化，并核对表达式、阈值、标签口径和同比基线。",
	}
	if count > 1 || input.repeatCount > 0 {
		parts = append(parts, "同一目标反复出现时，应优先判断规则噪声是否偏高，而不是直接按服务故障处理。")
	}
	return truncateText(strings.Join(parts, ""), 360)
}

func ruleDiagnosticEvidenceText(input reportInput) string {
	if input.kind == issueKindInvalidAlert || input.invalidAlerts <= 0 {
		return ""
	}
	count := input.invalidAlerts
	return fmt.Sprintf("同时有 %d 个同目标告警实例更适合作为告警规则诊断信号，处置时需要同步核对规则口径，避免规则噪声干扰服务风险判断。", count)
}

func inputScopeLabel(scopeKind string) string {
	switch scopeKind {
	case "entry":
		return "该入口侧接口"
	case "serviceEndpoint":
		return "该服务接口"
	case "service":
		return "该服务"
	case "node":
		return "该节点"
	case "pod":
		return "该实例"
	case "database":
		return "该数据库"
	default:
		return "该影响目标"
	}
}

func inputScopeNoun(scopeKind string) string {
	switch scopeKind {
	case "entry":
		return "入口侧接口"
	case "serviceEndpoint":
		return "服务接口"
	case "service":
		return "服务"
	case "node":
		return "节点"
	case "pod":
		return "实例"
	case "database":
		return "数据库"
	default:
		return "影响目标"
	}
}

func alertEventFrequencyText(eventCount int, repeatCount int, instanceCount int, symptom string) string {
	switch {
	case eventCount >= 5 || repeatCount >= 3:
		return "多次反复出现" + symptom + "告警"
	case eventCount >= 2 || instanceCount >= 2:
		return "多次出现" + symptom + "告警"
	case eventCount == 1:
		return "单次出现" + symptom + "告警"
	default:
		return "识别到" + symptom + "告警"
	}
}

func alertRecoveryText(input reportInput) string {
	if input.activeAlerts > 0 {
		return "当前仍有告警未恢复，需要优先确认影响是否仍在扩大"
	}
	if len(input.events) >= 3 || input.repeatCount > 0 {
		return "当前已恢复，但窗口内触发较频繁，需要确认恢复是否稳定"
	}
	return "当前已恢复，建议结合后续窗口确认是否复发"
}

func buildAlertEventPatternSummary(input reportInput, alertNames []string, directions []string) string {
	parts := make([]string, 0, 5)
	eventCount := len(input.events)
	if eventCount == 0 {
		return ""
	}
	if input.repeatCount > 0 {
		parts = append(parts, fmt.Sprintf("同一告警对象反复出现%d次", input.repeatCount))
	} else if eventCount > 1 && input.deduplicatedAlerts <= 1 {
		parts = append(parts, "同一影响目标出现多次同类事件")
	}
	if len(alertNames) > 1 {
		parts = append(parts, fmt.Sprintf("同一影响目标出现%d类告警", len(alertNames)))
	}
	if input.deduplicatedAlerts > 1 {
		parts = append(parts, fmt.Sprintf("覆盖%d个告警对象", input.deduplicatedAlerts))
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
		if direction := cleanUserFacingReportText(event.AlertDirection); direction != "" {
			values = append(values, direction)
		}
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
	if historyResult := buildAlertHistoryEvidenceResult(input); historyResult != "" {
		evidence = append(evidence, response.SystemHealthEvidence{
			Type:     "告警历史",
			Source:   sourceLabel("alerts"),
			Severity: maxSeverity(input.alertSeverities),
			Query:    "alert_event.alert_id last_3d",
			Result:   historyResult,
		})
	}

	if reportEvidence := buildAlertReportEvidence(input.alertReportFindings, maxSeverity(input.alertSeverities)); len(reportEvidence) > 0 {
		evidence = append(evidence, reportEvidence...)
	}
	if trendEvidence := h.alertRuleMetricTrendEvidence(ctx, input, groupID, startTime, endTime); trendEvidence.Result != "" {
		evidence = append(evidence, trendEvidence)
	}
	if requestPolarisEvidence := h.requestPolarisEvidence(ctx, input, groupID, startTime, endTime); requestPolarisEvidence.Result != "" {
		evidence = append(evidence, requestPolarisEvidence)
	}
	if objectStatusEvidence := h.relatedObjectStatusEvidence(ctx, input); objectStatusEvidence.Result != "" {
		evidence = append(evidence, objectStatusEvidence)
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

func buildAlertReportEvidence(findings []systemHealthAlertReportFinding, severity string) []response.SystemHealthEvidence {
	grouped := make(map[string][]string)
	for _, finding := range findings {
		for _, item := range finding.Evidence {
			if strings.TrimSpace(item.Kind) == "" || strings.TrimSpace(item.Summary) == "" {
				continue
			}
			summary := displayableAlertReportEvidenceSummary(item.Kind, item.Summary)
			if summary == "" {
				continue
			}
			grouped[item.Kind] = uniqueStrings(append(grouped[item.Kind], summary))
		}
	}
	order := []string{"logs", "metrics", "alert"}
	result := make([]response.SystemHealthEvidence, 0, len(order))
	for _, kind := range order {
		summaries := grouped[kind]
		if len(summaries) == 0 {
			continue
		}
		result = append(result, response.SystemHealthEvidence{
			Type:     alertReportEvidenceType(kind),
			Source:   sourceLabel("alertReport"),
			Severity: severity,
			Query:    alertReportEvidenceQuery(kind),
			Result:   alertReportEvidenceResult(kind, strings.Join(summaries[:minInt(len(summaries), 3)], "；")),
		})
	}
	return result
}

func alertReportEvidenceType(kind string) string {
	switch kind {
	case "apm":
		return "方向报告-APM链路"
	case "logs":
		return "方向报告-日志分析"
	case "metrics":
		return "方向报告-指标趋势"
	case "alert":
		return "方向报告-告警事件分析"
	default:
		return "方向报告-补充证据"
	}
}

func alertReportEvidenceQuery(kind string) string {
	switch kind {
	case "apm":
		return "alert_report.topology + alert_report.data"
	case "logs", "metrics", "alert":
		return "alert_report.data"
	default:
		return "alert_report"
	}
}

func alertReportEvidenceResult(kind string, summary string) string {
	summary = normalizeReportEvidenceSummary(summary)
	switch kind {
	case "apm":
		return "方向报告中的 APM 链路分析显示：" + summary
	case "logs":
		return "方向报告中的日志分析显示：" + summary
	case "metrics":
		return "方向报告中的指标趋势显示：" + summary
	case "alert":
		return "方向报告中的告警事件分析显示：" + summary
	default:
		return "方向报告补充证据显示：" + summary
	}
}

func (h *handler) requestPolarisEvidence(ctx core.Context, input reportInput, groupID int64, startTime time.Time, endTime time.Time) response.SystemHealthEvidence {
	evidence := response.SystemHealthEvidence{
		Type:     "请求北极星指标趋势",
		Source:   sourceLabel("requestPolaris"),
		Severity: maxSeverity(input.alertSeverities),
		Query:    "kindling_profiling_request_resource_duration",
	}
	if !isLatencyIssueInput(input) {
		return evidence
	}
	if h == nil || h.promRepo == nil || len(input.events) == 0 {
		systemHealthWarnCtx(ctx, "system-health request polaris skipped",
			zap.String("issueKey", input.key),
			zap.String("aggregationKey", input.aggregationKey),
			zap.Bool("hasHandler", h != nil),
			zap.Bool("hasPromRepo", h != nil && h.promRepo != nil),
			zap.Int("eventCount", len(input.events)),
			zap.String("reason", "missing_prometheus_or_events"),
		)
		return evidence
	}

	event, serviceName, contentKey, clusterID := requestPolarisTarget(input.events)
	if strings.TrimSpace(serviceName) == "" {
		systemHealthWarnCtx(ctx, "system-health request polaris skipped",
			zap.String("issueKey", input.key),
			zap.String("aggregationKey", input.aggregationKey),
			zap.String("reason", "service_name_missing"),
		)
		return evidence
	}

	step := trendQueryStep(startTime, endTime)
	interval := prometheus.VecFromDuration(step)
	specs := requestPolarisMetricSpecs(serviceName, contentKey, clusterID, interval, step)
	trends := make([]requestPolarisMetricTrend, 0, len(specs))
	queryStartedAt := time.Now()
	for _, spec := range specs {
		trend := h.queryRequestPolarisMetricTrend(ctx, spec, step, startTime, endTime)
		if trend.PointCount > 0 || trend.QueryErr != "" {
			trends = append(trends, trend)
		}
	}
	if len(trends) == 0 {
		systemHealthWarnCtx(ctx, "system-health request polaris unavailable",
			zap.String("issueKey", input.key),
			zap.String("aggregationKey", input.aggregationKey),
			zap.String("serviceName", serviceName),
			zap.String("contentKey", contentKey),
			zap.String("clusterID", clusterID),
			zap.String("reason", "no_metric_points"),
			zap.Duration("elapsed", time.Since(queryStartedAt)),
		)
		return evidence
	}

	latencyTrend := requestPolarisTrendByKey(trends, "latency")
	topTrend := firstNonNilRequestPolarisTrend(latencyTrend, topRequestPolarisTrend(trends))
	if topTrend != nil {
		evidence.Metric = buildMetricEvidence(alertMetricTrendQueryResult{
			MetricName:     topTrend.MetricName,
			Expression:     topTrend.Expression,
			Step:           step,
			StartTime:      startTime,
			EndTime:        endTime,
			SeriesReturned: topTrend.SeriesCount,
			MatchedSeries:  topTrend.Series,
			Stats:          topTrend.Stats,
			EventScoped:    strings.TrimSpace(contentKey) != "",
			Message:        requestPolarisMetricMessage(*topTrend),
		})
	}
	evidence.PolarisBreakdown = buildRequestPolarisBreakdownEvidence(input.scope, serviceName, contentKey, startTime, endTime, trends)

	abnormal := requestPolarisAbnormalTrends(trends)
	var slowSpans []clickhouse.QueryTraceResult
	var slowSpanTotal int64
	var slowSpanErr error
	if len(abnormal) == 0 && h.chRepo != nil {
		slowSpans, slowSpanTotal, slowSpanErr = h.queryRequestPolarisSlowSpans(ctx, groupID, startTime, endTime, serviceName, contentKey, clusterID)
	}
	evidence.Result = buildRequestPolarisEvidenceResult(input.scope, serviceName, contentKey, trends, abnormal, slowSpans, slowSpanTotal, slowSpanErr)
	systemHealthInfoCtx(ctx, "system-health request polaris completed",
		zap.String("issueKey", input.key),
		zap.String("aggregationKey", input.aggregationKey),
		zap.String("eventID", event.EventID),
		zap.String("alertID", event.AlertID),
		zap.String("alertName", event.Name),
		zap.String("serviceName", serviceName),
		zap.String("contentKey", contentKey),
		zap.String("clusterID", clusterID),
		zap.Int("trendCount", len(trends)),
		zap.Int("abnormalTrendCount", len(abnormal)),
		zap.Int64("slowSpanTotal", slowSpanTotal),
		zap.Bool("hasSlowSpanError", slowSpanErr != nil),
		zap.Duration("elapsed", time.Since(queryStartedAt)),
	)
	return evidence
}

func isLatencyIssueInput(input reportInput) bool {
	text := strings.ToLower(strings.Join(append([]string{input.title, input.scope, input.impact}, input.alertNames...), " "))
	return strings.Contains(text, "latency") ||
		strings.Contains(text, "duration") ||
		strings.Contains(text, "timeout") ||
		strings.Contains(text, "slow") ||
		strings.Contains(text, "p95") ||
		strings.Contains(text, "p99") ||
		strings.Contains(text, "延时") ||
		strings.Contains(text, "延迟") ||
		strings.Contains(text, "耗时") ||
		strings.Contains(text, "超时") ||
		strings.Contains(text, "慢")
}

func requestPolarisTarget(events []alertmodel.AEventWithWRecord) (alertmodel.AEventWithWRecord, string, string, string) {
	for _, event := range events {
		rawTags := rawTagsFromEvent(event)
		serviceName := firstNonEmpty(
			rawTags["service_name"],
			rawTags["serviceName"],
			rawTags["svc_name"],
			event.GetServiceNameTag(),
		)
		if strings.TrimSpace(serviceName) == "" {
			continue
		}
		return event,
			serviceName,
			firstNonEmpty(rawTags["content_key"], rawTags["endpoint"]),
			firstNonEmpty(rawTags["cluster_id"], rawTags["clusterId"])
	}
	return alertmodel.AEventWithWRecord{}, "", "", ""
}

func requestPolarisMetricSpecs(serviceName string, contentKey string, clusterID string, interval string, step time.Duration) []requestPolarisMetricTrend {
	selector := requestPolarisSelector(serviceName, contentKey, clusterID)
	stepSeconds := math.Max(step.Seconds(), 1)
	return []requestPolarisMetricTrend{
		{
			Key:          "net",
			MetricName:   "请求网络平均耗时",
			Unit:         "ms/request",
			ResourceKind: "网络耗时",
			Expression: fmt.Sprintf(
				`(sum by (content_key, svc_name) (increase(kindling_profiling_net_duration_nanoseconds_sum{%s}[%s])) / sum by (content_key, svc_name) (increase(kindling_profiling_net_duration_nanoseconds_count{%s}[%s]))) / 1000000`,
				selector, interval, selector, interval,
			),
		},
		{
			Key:          "epoll",
			MetricName:   "请求调度等待平均耗时",
			Unit:         "ms/request",
			ResourceKind: "调度等待耗时",
			Expression: fmt.Sprintf(
				`(sum by (content_key, svc_name) (increase(kindling_profiling_epoll_duration_nanoseconds_sum{%s}[%s])) / sum by (content_key, svc_name) (increase(kindling_profiling_epoll_duration_nanoseconds_count{%s}[%s]))) / 1000000`,
				selector, interval, selector, interval,
			),
		},
		{
			Key:          "cpu",
			MetricName:   "请求CPU执行平均耗时",
			Unit:         "ms/request",
			ResourceKind: "CPU执行耗时",
			Expression: fmt.Sprintf(
				`(sum by (content_key, svc_name) (increase(kindling_profiling_cpu_duration_nanoseconds_sum{%s}[%s])) / sum by (content_key, svc_name) (increase(kindling_span_trace_duration_nanoseconds_count{%s}[%s]))) / 1000000`,
				selector, interval, selector, interval,
			),
		},
		{
			Key:          "latency",
			MetricName:   "请求平均耗时",
			Unit:         "ms/request",
			ResourceKind: "请求耗时",
			Expression: fmt.Sprintf(
				`(sum by (content_key, svc_name) (increase(kindling_span_trace_duration_nanoseconds_sum{%s}[%s])) / sum by (content_key, svc_name) (increase(kindling_span_trace_duration_nanoseconds_count{%s}[%s]))) / 1000000`,
				selector, interval, selector, interval,
			),
		},
		{
			Key:          "tps",
			MetricName:   "请求吞吐量",
			Unit:         "req/s",
			ResourceKind: "TPS",
			Expression: fmt.Sprintf(
				`sum by (content_key, svc_name) (increase(kindling_span_trace_duration_nanoseconds_count{%s}[%s])) / %.0f`,
				selector, interval, stepSeconds,
			),
		},
	}
}

func requestPolarisSelector(serviceName string, contentKey string, clusterID string) string {
	matchers := []string{fmt.Sprintf(`svc_name=~"%s"`, prometheus.EscapeRegexp(serviceName))}
	if strings.TrimSpace(contentKey) != "" {
		matchers = append(matchers, fmt.Sprintf(`content_key=~"%s"`, prometheus.EscapeRegexp(contentKey)))
	} else {
		matchers = append(matchers, `content_key=~".*"`)
	}
	if strings.TrimSpace(clusterID) != "" {
		matchers = append(matchers, fmt.Sprintf(`cluster_id=~"%s"`, prometheus.EscapeRegexp(clusterID)))
	}
	return strings.Join(matchers, ",")
}

func (h *handler) queryRequestPolarisMetricTrend(ctx core.Context, spec requestPolarisMetricTrend, step time.Duration, startTime time.Time, endTime time.Time) requestPolarisMetricTrend {
	trend := spec
	results, err := h.queryRawMetricSeries(ctx, startTime, endTime, spec.Expression, step)
	if err != nil {
		trend.QueryErr = err.Error()
		return trend
	}
	stats, ok := calculateMetricTrendStats(results)
	if !ok {
		trend.QueryErr = "no_usable_points"
		return trend
	}
	trend.Stats = stats
	trend.Series = results
	trend.SeriesCount = len(results)
	trend.PointCount = stats.PointCount
	trend.Abnormal, trend.Reason = requestPolarisTrendAbnormalReason(trend)
	return trend
}

func requestPolarisTrendAbnormalReason(trend requestPolarisMetricTrend) (bool, string) {
	if trend.PointCount == 0 {
		return false, ""
	}
	if trend.Key == "tps" {
		if trend.Stats.ChangePercent != nil && *trend.Stats.ChangePercent >= requestPolarisIncreaseThreshold*100 && trend.Stats.Delta > 0 {
			return true, fmt.Sprintf("TPS 升高约 %s%%", formatMetricFloat(*trend.Stats.ChangePercent))
		}
		return false, ""
	}
	if trend.Key == "latency" {
		return false, ""
	}
	if trend.Stats.MaxValue >= requestPolarisResourceThresholdMS {
		return true, fmt.Sprintf("%s最高达到 %sms", trend.ResourceKind, formatMetricFloat(trend.Stats.MaxValue))
	}
	if trend.Stats.ChangePercent != nil && *trend.Stats.ChangePercent >= requestPolarisIncreaseThreshold*100 && trend.Stats.Delta > 0 {
		return true, fmt.Sprintf("%s升高约 %s%%", trend.ResourceKind, formatMetricFloat(*trend.Stats.ChangePercent))
	}
	return false, ""
}

func requestPolarisAbnormalTrends(trends []requestPolarisMetricTrend) []requestPolarisMetricTrend {
	result := make([]requestPolarisMetricTrend, 0, len(trends))
	for _, trend := range trends {
		if trend.Abnormal && trend.Key != "tps" {
			result = append(result, trend)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Stats.MaxValue > result[j].Stats.MaxValue
	})
	return result
}

func topRequestPolarisTrend(trends []requestPolarisMetricTrend) *requestPolarisMetricTrend {
	var top *requestPolarisMetricTrend
	for index := range trends {
		trend := &trends[index]
		if trend.PointCount == 0 || trend.Key == "tps" {
			continue
		}
		if top == nil || (trend.Abnormal && !top.Abnormal) || trend.Stats.MaxValue > top.Stats.MaxValue {
			top = trend
		}
	}
	return top
}

func firstNonNilRequestPolarisTrend(trends ...*requestPolarisMetricTrend) *requestPolarisMetricTrend {
	for _, trend := range trends {
		if trend != nil && trend.PointCount > 0 {
			return trend
		}
	}
	return nil
}

func requestPolarisMetricMessage(trend requestPolarisMetricTrend) string {
	return fmt.Sprintf("%s在窗口内首值=%s、末值=%s、最高=%s，单位=%s。%s",
		trend.MetricName,
		formatMetricFloat(trend.Stats.FirstValue),
		formatMetricFloat(trend.Stats.LastValue),
		formatMetricFloat(trend.Stats.MaxValue),
		trend.Unit,
		firstNonEmpty(trend.Reason, "未达到请求北极星异常阈值"),
	)
}

func (h *handler) queryRequestPolarisSlowSpans(ctx core.Context, groupID int64, startTime time.Time, endTime time.Time, serviceName string, contentKey string, clusterID string) ([]clickhouse.QueryTraceResult, int64, error) {
	thresholdMicros := int64(requestPolarisSlowSpanThresholdMS * 1000)
	req := &request.GetTracePageListRequest{
		GroupID:   groupID,
		StartTime: startTime.UnixMicro(),
		EndTime:   endTime.UnixMicro(),
		Service:   []string{serviceName},
		EndPoint:  contentKey,
		PageNum:   1,
		PageSize:  20,
		Filters: []*request.ComplexSpanTraceFilter{
			{
				SpanTraceFilter: &request.SpanTraceFilter{
					Key:       "duration",
					DataType:  request.U64Column,
					Operation: request.OpGreaterThan,
					Value:     []string{strconv.FormatInt(thresholdMicros, 10)},
				},
			},
		},
	}
	if clusterID != "" {
		req.ClusterIDs = []string{clusterID}
	}
	return h.chRepo.GetTracePageList(ctx, req)
}

func buildRequestPolarisBreakdownEvidence(scope string, serviceName string, contentKey string, startTime time.Time, endTime time.Time, trends []requestPolarisMetricTrend) *response.SystemHealthPolarisBreakdownEvidence {
	if len(trends) == 0 {
		return nil
	}
	breakdown := &response.SystemHealthPolarisBreakdownEvidence{
		ServiceName: serviceName,
		ContentKey:  contentKey,
		Target:      requestPolarisTargetText(scope, serviceName, contentKey),
		StartTime:   startTime.UnixMicro(),
		EndTime:     endTime.UnixMicro(),
		Resources:   make([]response.SystemHealthPolarisBreakdownItemEvidence, 0, 3),
	}
	for _, key := range []string{"latency", "net", "cpu", "epoll", "tps"} {
		trend := requestPolarisTrendByKey(trends, key)
		if trend == nil {
			continue
		}
		item := requestPolarisBreakdownItem(*trend)
		switch key {
		case "latency":
			breakdown.Latency = &item
		case "tps":
			breakdown.TPS = &item
		default:
			breakdown.Resources = append(breakdown.Resources, item)
		}
	}
	if breakdown.Latency == nil && len(breakdown.Resources) == 0 && breakdown.TPS == nil {
		return nil
	}
	return breakdown
}

func requestPolarisBreakdownItem(trend requestPolarisMetricTrend) response.SystemHealthPolarisBreakdownItemEvidence {
	points, pointsTruncated := aggregateAlertMetricSeriesPoints(trend.Series, maxMetricEvidencePoints)
	return response.SystemHealthPolarisBreakdownItemEvidence{
		Key:             trend.Key,
		Name:            trend.MetricName,
		ResourceKind:    trend.ResourceKind,
		Unit:            trend.Unit,
		PointCount:      trend.PointCount,
		FirstValue:      trend.Stats.FirstValue,
		LastValue:       trend.Stats.LastValue,
		AverageValue:    trend.Stats.AverageValue,
		MinValue:        trend.Stats.MinValue,
		MaxValue:        trend.Stats.MaxValue,
		Delta:           trend.Stats.Delta,
		ChangePercent:   trend.Stats.ChangePercent,
		Abnormal:        trend.Abnormal,
		Reason:          trend.Reason,
		Points:          points,
		PointLimit:      maxMetricEvidencePoints,
		PointsTruncated: pointsTruncated,
	}
}

func buildRequestPolarisEvidenceResult(scope string, serviceName string, contentKey string, trends []requestPolarisMetricTrend, abnormal []requestPolarisMetricTrend, slowSpans []clickhouse.QueryTraceResult, slowSpanTotal int64, slowSpanErr error) string {
	targetDetail := requestPolarisTargetText(scope, serviceName, contentKey)
	tpsText := requestPolarisTPSText(trends)
	latencyTrend := requestPolarisTrendByKey(trends, "latency")
	if len(abnormal) > 0 {
		top := abnormal[0]
		parts := []string{
			fmt.Sprintf("请求北极星显示，%s 在告警窗口内%s异常，最高 %sms，说明单次请求耗时可能被%s放大", targetDetail, top.ResourceKind, formatMetricFloat(top.Stats.MaxValue), top.ResourceKind),
		}
		if top.Stats.ChangePercent != nil {
			parts = append(parts, fmt.Sprintf("窗口变化约 %s%%", formatSignedMetricFloat(*top.Stats.ChangePercent)))
		}
		if tpsText != "" {
			parts = append(parts, tpsText)
		}
		if len(abnormal) > 1 {
			parts = append(parts, "同时可见"+requestPolarisTrendBrief(abnormal[1]))
		}
		return truncateText(strings.Join(parts, "；")+"。", 520)
	}
	latencyText := ""
	if latencyTrend != nil && latencyTrend.PointCount > 0 {
		latencyText = fmt.Sprintf("请求平均耗时从 %sms 到 %sms，最高 %sms；", formatMetricFloat(latencyTrend.Stats.FirstValue), formatMetricFloat(latencyTrend.Stats.LastValue), formatMetricFloat(latencyTrend.Stats.MaxValue))
	}
	targetText := targetDetail
	if targetText != "" {
		targetText += "，"
	}
	if slowSpanErr != nil {
		return truncateText(fmt.Sprintf("请求北极星显示，%s%s网络、CPU执行、调度等待等资源分项未见超过 %.0fms/请求或明显升高；慢调用样本检索失败：%s。%s", targetText, latencyText, requestPolarisResourceThresholdMS, slowSpanErr.Error(), tpsText), 560)
	}
	if slowSpanTotal > 0 {
		return truncateText(fmt.Sprintf("请求北极星显示，%s%s网络、CPU执行、调度等待等资源分项未见超过 %.0fms/请求或明显升高；但检索到 %d 条超过 %.0fms 的慢调用样本，%s。该结果用于定位慢调用样本，不代表整体资源耗时已异常。%s",
			targetText,
			latencyText,
			requestPolarisResourceThresholdMS,
			slowSpanTotal,
			requestPolarisSlowSpanThresholdMS,
			summarizeTraceSamples(slowSpans),
			tpsText,
		), 560)
	}
	return truncateText(fmt.Sprintf("请求北极星显示，%s%s网络、CPU执行、调度等待等资源分项未见超过 %.0fms/请求或明显升高，也未检索到超过 %.0fms 的慢调用样本；当前不把资源分项耗时作为主要根因。%s",
		targetText,
		latencyText,
		requestPolarisResourceThresholdMS,
		requestPolarisSlowSpanThresholdMS,
		tpsText,
	), 520)
}

func requestPolarisTrendByKey(trends []requestPolarisMetricTrend, key string) *requestPolarisMetricTrend {
	for index := range trends {
		if trends[index].Key == key && trends[index].PointCount > 0 {
			return &trends[index]
		}
	}
	return nil
}

func requestPolarisTargetText(scope string, serviceName string, contentKey string) string {
	target := firstNonEmpty(scope, serviceName, contentKey, "该请求")
	details := make([]string, 0, 2)
	if serviceName != "" && !strings.Contains(target, serviceName) {
		details = append(details, "服务 "+serviceName)
	}
	if contentKey != "" && !strings.Contains(target, contentKey) {
		details = append(details, "请求路径 "+contentKey)
	}
	if len(details) == 0 {
		return target
	}
	return target + "（" + strings.Join(details, "，") + "）"
}

func requestPolarisTPSText(trends []requestPolarisMetricTrend) string {
	for _, trend := range trends {
		if trend.Key != "tps" || trend.PointCount == 0 {
			continue
		}
		direction := "基本持平"
		if trend.Stats.Delta > 0 {
			direction = "升高"
		} else if trend.Stats.Delta < 0 {
			direction = "回落"
		}
		change := ""
		if trend.Stats.ChangePercent != nil {
			change = fmt.Sprintf("，变化约 %s%%", formatSignedMetricFloat(*trend.Stats.ChangePercent))
		}
		return fmt.Sprintf("同窗口 TPS 从 %s 到 %s req/s，整体%s%s", formatMetricFloat(trend.Stats.FirstValue), formatMetricFloat(trend.Stats.LastValue), direction, change)
	}
	return ""
}

func requestPolarisTrendBrief(trend requestPolarisMetricTrend) string {
	return fmt.Sprintf("%s最高 %sms", trend.ResourceKind, formatMetricFloat(trend.Stats.MaxValue))
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
		systemHealthWarnCtx(ctx, "system-health metric trend skipped",
			zap.String("issueKey", input.key),
			zap.String("aggregationKey", input.aggregationKey),
			zap.String("reason", "no_events"),
		)
		return evidence
	}
	trend, err := h.queryAlertMetricTrend(ctx, input.events[:minInt(maxEvidenceSamples, len(input.events))], groupID, startTime, endTime)
	if err != nil {
		evidence.Result = err.Error()
		systemHealthWarnCtx(ctx, "system-health metric trend fallback",
			zap.String("issueKey", input.key),
			zap.String("aggregationKey", input.aggregationKey),
			zap.String("scope", input.scope),
			zap.String("reason", "query_failed_or_unavailable"),
			zap.Error(err),
		)
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
		systemHealthInfoCtx(ctx, "system-health metric trend matched event series",
			zap.String("issueKey", input.key),
			zap.String("aggregationKey", input.aggregationKey),
			zap.String("ruleGroup", trend.Rule.Group),
			zap.String("ruleName", trend.Rule.Alert),
			zap.Int("seriesReturned", trend.SeriesReturned),
			zap.Int("seriesMatched", trend.Stats.SeriesCount),
			zap.Int("pointCount", trend.Stats.PointCount),
			zap.Bool("eventScoped", trend.EventScoped),
			zap.Duration("step", trend.Step),
			zap.Any("matchLabels", trend.MatchLabels),
		)
		return evidence
	}
	evidence.Result = truncateText(fmt.Sprintf(
		"已匹配告警规则 %s/%s；规则查询返回 %d 条 series，但当前窗口内没有命中告警事件对应的 series，未展示规则整体聚合值。",
		trend.Rule.Group,
		trend.Rule.Alert,
		trend.SeriesReturned,
	), 420)
	systemHealthWarnCtx(ctx, "system-health metric trend used global boundary only",
		zap.String("issueKey", input.key),
		zap.String("aggregationKey", input.aggregationKey),
		zap.String("ruleGroup", trend.Rule.Group),
		zap.String("ruleName", trend.Rule.Alert),
		zap.Int("seriesReturned", trend.SeriesReturned),
		zap.Int("pointCount", trend.Stats.PointCount),
		zap.Bool("eventScoped", trend.EventScoped),
		zap.String("reason", trend.Message),
	)
	return evidence
}

func (h *handler) queryAlertMetricTrend(ctx core.Context, events []alertmodel.AEventWithWRecord, groupID int64, startTime time.Time, endTime time.Time) (alertMetricTrendQueryResult, error) {
	if len(events) == 0 {
		return alertMetricTrendQueryResult{}, fmt.Errorf("未提供告警事件，无法反查 PQL 和指标趋势。")
	}
	event := events[0]
	queryStartedAt := time.Now()
	systemHealthInfoCtx(ctx, "system-health metric trend query started",
		zap.Int64("groupId", groupID),
		zap.Int("eventCount", len(events)),
		zap.String("eventID", event.EventID),
		zap.String("alertID", event.AlertID),
		zap.String("alertName", event.Name),
		zap.Time("startTime", startTime),
		zap.Time("endTime", endTime),
	)
	rule, found := h.findAlertRuleForEvent(ctx, event, groupID)
	if !found || rule == nil {
		systemHealthWarnCtx(ctx, "system-health metric trend query failed",
			zap.Int64("groupId", groupID),
			zap.String("eventID", event.EventID),
			zap.String("alertID", event.AlertID),
			zap.String("alertName", event.Name),
			zap.String("reason", "alert_rule_not_found"),
			zap.Duration("elapsed", time.Since(queryStartedAt)),
		)
		return alertMetricTrendQueryResult{}, fmt.Errorf("未找到告警规则：告警事件=%s，无法反查 PQL 和指标趋势。", event.Name)
	}
	expr := strings.TrimSpace(rule.Expr)
	if expr == "" {
		systemHealthWarnCtx(ctx, "system-health metric trend query failed",
			zap.Int64("groupId", groupID),
			zap.String("eventID", event.EventID),
			zap.String("alertName", event.Name),
			zap.String("ruleGroup", rule.Group),
			zap.String("ruleName", rule.Alert),
			zap.String("reason", "empty_rule_expr"),
			zap.Duration("elapsed", time.Since(queryStartedAt)),
		)
		return alertMetricTrendQueryResult{}, fmt.Errorf("已匹配告警规则 %s/%s，但规则未配置 PQL。", rule.Group, rule.Alert)
	}
	if h.promRepo == nil {
		systemHealthWarnCtx(ctx, "system-health metric trend query failed",
			zap.Int64("groupId", groupID),
			zap.String("eventID", event.EventID),
			zap.String("alertName", event.Name),
			zap.String("ruleGroup", rule.Group),
			zap.String("ruleName", rule.Alert),
			zap.String("reason", "prometheus_repository_not_configured"),
			zap.Duration("elapsed", time.Since(queryStartedAt)),
		)
		return alertMetricTrendQueryResult{}, fmt.Errorf("已匹配告警规则 %s/%s；Prometheus 仓库未初始化，无法查询窗口趋势。", rule.Group, rule.Alert)
	}

	step := trendQueryStep(startTime, endTime)
	systemHealthInfoCtx(ctx, "system-health metric trend raw query prepared",
		zap.Int64("groupId", groupID),
		zap.String("eventID", event.EventID),
		zap.String("alertName", event.Name),
		zap.String("ruleGroup", rule.Group),
		zap.String("ruleName", rule.Alert),
		zap.String("ruleSource", rule.Labels["source"]),
		zap.Int("exprLength", len(expr)),
		zap.String("exprHash", systemHealthShortHash(expr)),
		zap.Duration("step", step),
	)
	results, err := h.queryRawMetricSeries(ctx, startTime, endTime, expr, step)
	if err != nil {
		systemHealthWarnCtx(ctx, "system-health metric trend query failed",
			zap.Int64("groupId", groupID),
			zap.String("eventID", event.EventID),
			zap.String("alertName", event.Name),
			zap.String("ruleGroup", rule.Group),
			zap.String("ruleName", rule.Alert),
			zap.String("ruleSource", rule.Labels["source"]),
			zap.String("reason", "prometheus_query_failed"),
			zap.Duration("elapsed", time.Since(queryStartedAt)),
			zap.Error(err),
		)
		return alertMetricTrendQueryResult{}, fmt.Errorf("已匹配告警规则 %s/%s；查询窗口趋势失败：%w", rule.Group, rule.Alert, err)
	}
	for _, candidate := range events {
		matched := filterSeriesByAlertEvent(results, candidate)
		stats, ok := calculateMetricTrendStats(matched)
		if !ok {
			continue
		}
		systemHealthInfoCtx(ctx, "system-health metric trend query completed",
			zap.Int64("groupId", groupID),
			zap.String("eventID", candidate.EventID),
			zap.String("alertName", candidate.Name),
			zap.String("ruleGroup", rule.Group),
			zap.String("ruleName", rule.Alert),
			zap.String("ruleSource", rule.Labels["source"]),
			zap.Int("seriesReturned", len(results)),
			zap.Int("seriesMatched", len(matched)),
			zap.Int("pointCount", stats.PointCount),
			zap.Bool("eventScoped", true),
			zap.Duration("elapsed", time.Since(queryStartedAt)),
		)
		return alertMetricTrendQueryResult{
			Rule:           rule,
			Expression:     expr,
			Step:           step,
			StartTime:      startTime,
			EndTime:        endTime,
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
		systemHealthWarnCtx(ctx, "system-health metric trend query failed",
			zap.Int64("groupId", groupID),
			zap.String("eventID", event.EventID),
			zap.String("alertName", event.Name),
			zap.String("ruleGroup", rule.Group),
			zap.String("ruleName", rule.Alert),
			zap.String("ruleSource", rule.Labels["source"]),
			zap.Int("seriesReturned", len(results)),
			zap.String("reason", "no_usable_points"),
			zap.Duration("elapsed", time.Since(queryStartedAt)),
		)
		return alertMetricTrendQueryResult{}, fmt.Errorf("已匹配告警规则 %s/%s；规则查询未返回可用时序点。", rule.Group, rule.Alert)
	}
	systemHealthWarnCtx(ctx, "system-health metric trend query completed without event scoped series",
		zap.Int64("groupId", groupID),
		zap.String("eventID", event.EventID),
		zap.String("alertName", event.Name),
		zap.String("ruleGroup", rule.Group),
		zap.String("ruleName", rule.Alert),
		zap.String("ruleSource", rule.Labels["source"]),
		zap.Int("seriesReturned", len(results)),
		zap.Int("pointCount", stats.PointCount),
		zap.Bool("eventScoped", false),
		zap.Duration("elapsed", time.Since(queryStartedAt)),
	)
	return alertMetricTrendQueryResult{
		Rule:           rule,
		Expression:     expr,
		Step:           step,
		StartTime:      startTime,
		EndTime:        endTime,
		SeriesReturned: len(results),
		Stats:          stats,
		EventScoped:    false,
		Message:        "规则查询只返回全局或不可匹配的聚合序列，未能用当前告警事件标签定位到指定 series。",
	}, nil
}

func buildMetricEvidence(trend alertMetricTrendQueryResult) *response.SystemHealthMetricEvidence {
	series, pointsTruncated := buildMetricSeriesEvidence(trend.MatchedSeries, maxMetricEvidenceSeries, maxMetricEvidencePoints)
	ruleGroup := ""
	ruleName := ""
	if trend.Rule != nil {
		ruleGroup = trend.Rule.Group
		ruleName = trend.Rule.Alert
	}
	var startTimeMicros int64
	var endTimeMicros int64
	if !trend.StartTime.IsZero() {
		startTimeMicros = trend.StartTime.UnixMicro()
	}
	if !trend.EndTime.IsZero() {
		endTimeMicros = trend.EndTime.UnixMicro()
	}
	metric := &response.SystemHealthMetricEvidence{
		RuleGroup:       ruleGroup,
		RuleName:        ruleName,
		MetricName:      firstNonEmpty(trend.MetricName, inferMetricNameFromRule(ruleName, trend.Expression)),
		Expression:      trend.Expression,
		Step:            trend.Step.String(),
		StartTime:       startTimeMicros,
		EndTime:         endTimeMicros,
		SeriesReturned:  trend.SeriesReturned,
		SeriesMatched:   trend.Stats.SeriesCount,
		PointCount:      trend.Stats.PointCount,
		FirstValue:      trend.Stats.FirstValue,
		LastValue:       trend.Stats.LastValue,
		AverageValue:    trend.Stats.AverageValue,
		MinValue:        trend.Stats.MinValue,
		MaxValue:        trend.Stats.MaxValue,
		MaxDeviation:    trend.Stats.MaxDeviation,
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

func aggregateAlertMetricSeriesPoints(series []alertMetricSeries, pointLimit int) ([]response.SystemHealthMetricDataPoint, bool) {
	if len(series) == 0 || pointLimit <= 0 {
		return nil, false
	}
	aggregated := make(map[int64]float64)
	for _, item := range series {
		for _, point := range item.Values {
			if point.TimeStamp <= 0 || math.IsNaN(point.Value) || math.IsInf(point.Value, 0) {
				continue
			}
			aggregated[point.TimeStamp] += point.Value
		}
	}
	if len(aggregated) == 0 {
		return nil, false
	}
	timestamps := make([]int64, 0, len(aggregated))
	for timestamp := range aggregated {
		timestamps = append(timestamps, timestamp)
	}
	sort.Slice(timestamps, func(i, j int) bool { return timestamps[i] < timestamps[j] })
	points := make([]prometheus.Points, 0, len(timestamps))
	for _, timestamp := range timestamps {
		points = append(points, prometheus.Points{
			TimeStamp: timestamp,
			Value:     aggregated[timestamp],
		})
	}
	return downsampleMetricPoints(points, pointLimit)
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
	lookupStartedAt := time.Now()
	systemHealthInfoCtx(ctx, "system-health alert rule lookup started",
		zap.Int64("groupId", groupID),
		zap.String("eventID", event.EventID),
		zap.String("alertID", event.AlertID),
		zap.String("alertName", event.Name),
		zap.String("eventGroup", event.Group),
	)
	if rule, found := h.findVMAlertRuleForEvent(ctx, event); found {
		systemHealthInfoCtx(ctx, "system-health alert rule lookup completed",
			zap.Int64("groupId", groupID),
			zap.String("eventID", event.EventID),
			zap.String("alertName", event.Name),
			zap.String("ruleGroup", rule.Group),
			zap.String("ruleName", rule.Alert),
			zap.String("ruleSource", rule.Labels["source"]),
			zap.String("matchedBy", "vmalert"),
			zap.Duration("elapsed", time.Since(lookupStartedAt)),
		)
		return rule, true
	}
	if h.k8sRepo == nil || strings.TrimSpace(event.Name) == "" {
		systemHealthWarnCtx(ctx, "system-health alert rule lookup failed",
			zap.Int64("groupId", groupID),
			zap.String("eventID", event.EventID),
			zap.String("alertName", event.Name),
			zap.Bool("k8sRepoConfigured", h.k8sRepo != nil),
			zap.String("reason", "missing_k8s_repository_or_alert_name"),
			zap.Duration("elapsed", time.Since(lookupStartedAt)),
		)
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
		rule, found := h.findAlertRuleByName(ctx, event, id, index == 0)
		if found {
			systemHealthInfoCtx(ctx, "system-health alert rule lookup completed",
				zap.Int64("groupId", groupID),
				zap.Int64("lookupGroupId", id),
				zap.String("eventID", event.EventID),
				zap.String("alertName", event.Name),
				zap.String("ruleGroup", rule.Group),
				zap.String("ruleName", rule.Alert),
				zap.String("ruleSource", firstNonEmpty(rule.Labels["source"], "k8s-alert-rules")),
				zap.String("matchedBy", "k8s_rule_name"),
				zap.Duration("elapsed", time.Since(lookupStartedAt)),
			)
			return rule, true
		}
	}
	systemHealthWarnCtx(ctx, "system-health alert rule lookup failed",
		zap.Int64("groupId", groupID),
		zap.String("eventID", event.EventID),
		zap.String("alertName", event.Name),
		zap.String("groupIDsTried", systemHealthCompactJSON(uniqueInt64(groupIDs))),
		zap.String("reason", "rule_not_found"),
		zap.Duration("elapsed", time.Since(lookupStartedAt)),
	)
	return nil, false
}

func (h *handler) findVMAlertRuleForEvent(ctx core.Context, event alertmodel.AEventWithWRecord) (*request.AlertRule, bool) {
	if strings.TrimSpace(h.vmAlertURL) == "" || strings.TrimSpace(event.Name) == "" {
		systemHealthInfoCtx(ctx, "system-health vmalert rule lookup skipped",
			zap.String("eventID", event.EventID),
			zap.String("alertName", event.Name),
			zap.Bool("hasVMAlertURL", strings.TrimSpace(h.vmAlertURL) != ""),
			zap.String("reason", "missing_vmalert_url_or_alert_name"),
		)
		return nil, false
	}
	if rule, found := h.findVMAlertActiveRuleForEvent(ctx, event); found {
		return rule, true
	}
	return h.findVMAlertRuleDefinitionForEvent(ctx, event)
}

func (h *handler) findVMAlertActiveRuleForEvent(ctx core.Context, event alertmodel.AEventWithWRecord) (*request.AlertRule, bool) {
	endpoint := strings.TrimRight(h.vmAlertURL, "/") + "/api/v1/alerts"
	lookupStartedAt := time.Now()
	systemHealthInfoCtx(ctx, "system-health vmalert active rule lookup started",
		zap.String("endpoint", endpoint),
		zap.String("eventID", event.EventID),
		zap.String("alertName", event.Name),
	)
	req, err := http.NewRequestWithContext(ctx.GetContext(), http.MethodGet, endpoint, nil)
	if err != nil {
		systemHealthWarnCtx(ctx, "system-health vmalert active rule lookup failed",
			zap.String("endpoint", endpoint),
			zap.String("eventID", event.EventID),
			zap.String("alertName", event.Name),
			zap.String("reason", "build_request_failed"),
			zap.Duration("elapsed", time.Since(lookupStartedAt)),
			zap.Error(err),
		)
		return nil, false
	}
	resp, err := (&http.Client{Timeout: 8 * time.Second}).Do(req)
	if err != nil {
		systemHealthWarnCtx(ctx, "system-health vmalert active rule lookup failed",
			zap.String("endpoint", endpoint),
			zap.String("eventID", event.EventID),
			zap.String("alertName", event.Name),
			zap.String("reason", "request_failed"),
			zap.Duration("elapsed", time.Since(lookupStartedAt)),
			zap.Error(err),
		)
		return nil, false
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		systemHealthWarnCtx(ctx, "system-health vmalert active rule lookup failed",
			zap.String("endpoint", endpoint),
			zap.String("eventID", event.EventID),
			zap.String("alertName", event.Name),
			zap.Int("httpStatus", resp.StatusCode),
			zap.String("reason", "bad_status"),
			zap.Duration("elapsed", time.Since(lookupStartedAt)),
		)
		return nil, false
	}

	var apiResp vmAlertAPIAlertsResp
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		systemHealthWarnCtx(ctx, "system-health vmalert active rule lookup failed",
			zap.String("endpoint", endpoint),
			zap.String("eventID", event.EventID),
			zap.String("alertName", event.Name),
			zap.Int("httpStatus", resp.StatusCode),
			zap.String("reason", "decode_failed"),
			zap.Duration("elapsed", time.Since(lookupStartedAt)),
			zap.Error(err),
		)
		return nil, false
	}
	candidateNameCount := 0
	filteredStateCount := 0
	filteredLabelCount := 0
	emptyExprCount := 0
	for _, alert := range apiResp.Data.Alerts {
		if strings.TrimSpace(alert.Name) != strings.TrimSpace(event.Name) {
			continue
		}
		candidateNameCount++
		if alert.State != "" && alert.State != alertmodel.StatusFiring {
			filteredStateCount++
			continue
		}
		if !labelsMatchAlertEvent(alert.Labels, event) {
			filteredLabelCount++
			continue
		}
		expr := strings.TrimSpace(firstNonEmpty(alert.Expression, alert.Query))
		if expr == "" {
			emptyExprCount++
			continue
		}
		rule := &request.AlertRule{
			Group: firstNonEmpty(alert.Labels["alertgroup"], alert.Labels["group"], alert.GroupID),
			Alert: event.Name,
			Expr:  expr,
			Labels: map[string]string{
				"group_id": alert.GroupID,
				"rule_id":  alert.ID,
				"source":   "vmalert-api",
			},
		}
		systemHealthInfoCtx(ctx, "system-health vmalert active rule lookup completed",
			zap.String("endpoint", endpoint),
			zap.String("eventID", event.EventID),
			zap.String("alertName", event.Name),
			zap.Int("httpStatus", resp.StatusCode),
			zap.Int("candidateRuleCount", candidateNameCount),
			zap.String("ruleGroup", rule.Group),
			zap.String("ruleName", rule.Alert),
			zap.String("matchedBy", "name_labels_state"),
			zap.Duration("elapsed", time.Since(lookupStartedAt)),
		)
		return rule, true
	}
	systemHealthWarnCtx(ctx, "system-health vmalert active rule lookup not matched",
		zap.String("endpoint", endpoint),
		zap.String("eventID", event.EventID),
		zap.String("alertName", event.Name),
		zap.Int("httpStatus", resp.StatusCode),
		zap.Int("rawAlertCount", len(apiResp.Data.Alerts)),
		zap.Int("candidateRuleCount", candidateNameCount),
		zap.Int("filteredStateCount", filteredStateCount),
		zap.Int("filteredLabelCount", filteredLabelCount),
		zap.Int("emptyExprCount", emptyExprCount),
		zap.Duration("elapsed", time.Since(lookupStartedAt)),
	)
	return nil, false
}

func (h *handler) findVMAlertRuleDefinitionForEvent(ctx core.Context, event alertmodel.AEventWithWRecord) (*request.AlertRule, bool) {
	endpoint := strings.TrimRight(h.vmAlertURL, "/") + "/api/v1/rules"
	lookupStartedAt := time.Now()
	systemHealthInfoCtx(ctx, "system-health vmalert rule definition lookup started",
		zap.String("endpoint", endpoint),
		zap.String("eventID", event.EventID),
		zap.String("alertName", event.Name),
	)
	req, err := http.NewRequestWithContext(ctx.GetContext(), http.MethodGet, endpoint, nil)
	if err != nil {
		systemHealthWarnCtx(ctx, "system-health vmalert rule definition lookup failed",
			zap.String("endpoint", endpoint),
			zap.String("eventID", event.EventID),
			zap.String("alertName", event.Name),
			zap.String("reason", "build_request_failed"),
			zap.Duration("elapsed", time.Since(lookupStartedAt)),
			zap.Error(err),
		)
		return nil, false
	}
	resp, err := (&http.Client{Timeout: 8 * time.Second}).Do(req)
	if err != nil {
		systemHealthWarnCtx(ctx, "system-health vmalert rule definition lookup failed",
			zap.String("endpoint", endpoint),
			zap.String("eventID", event.EventID),
			zap.String("alertName", event.Name),
			zap.String("reason", "request_failed"),
			zap.Duration("elapsed", time.Since(lookupStartedAt)),
			zap.Error(err),
		)
		return nil, false
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		systemHealthWarnCtx(ctx, "system-health vmalert rule definition lookup failed",
			zap.String("endpoint", endpoint),
			zap.String("eventID", event.EventID),
			zap.String("alertName", event.Name),
			zap.Int("httpStatus", resp.StatusCode),
			zap.String("reason", "bad_status"),
			zap.Duration("elapsed", time.Since(lookupStartedAt)),
		)
		return nil, false
	}

	var apiResp vmAlertAPIRulesResp
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		systemHealthWarnCtx(ctx, "system-health vmalert rule definition lookup failed",
			zap.String("endpoint", endpoint),
			zap.String("eventID", event.EventID),
			zap.String("alertName", event.Name),
			zap.Int("httpStatus", resp.StatusCode),
			zap.String("reason", "decode_failed"),
			zap.Duration("elapsed", time.Since(lookupStartedAt)),
			zap.Error(err),
		)
		return nil, false
	}

	rawTags := rawTagsFromEvent(event)
	var fallback *request.AlertRule
	totalRuleCount := 0
	candidateNameCount := 0
	emptyExprCount := 0
	for _, group := range apiResp.Data.Groups {
		for _, rule := range group.Rules {
			totalRuleCount++
			if strings.TrimSpace(rule.Name) != strings.TrimSpace(event.Name) {
				continue
			}
			candidateNameCount++
			expr := strings.TrimSpace(firstNonEmpty(rule.Query, rule.Expr))
			if expr == "" {
				emptyExprCount++
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
				systemHealthInfoCtx(ctx, "system-health vmalert rule definition lookup completed",
					zap.String("endpoint", endpoint),
					zap.String("eventID", event.EventID),
					zap.String("alertName", event.Name),
					zap.Int("httpStatus", resp.StatusCode),
					zap.Int("groupCount", len(apiResp.Data.Groups)),
					zap.Int("totalRuleCount", totalRuleCount),
					zap.Int("candidateRuleCount", candidateNameCount),
					zap.String("ruleGroup", candidate.Group),
					zap.String("ruleName", candidate.Alert),
					zap.String("matchedBy", "name_labels_or_group"),
					zap.Duration("elapsed", time.Since(lookupStartedAt)),
				)
				return candidate, true
			}
		}
	}
	if fallback != nil {
		systemHealthWarnCtx(ctx, "system-health vmalert rule definition fallback matched by name",
			zap.String("endpoint", endpoint),
			zap.String("eventID", event.EventID),
			zap.String("alertName", event.Name),
			zap.Int("httpStatus", resp.StatusCode),
			zap.Int("groupCount", len(apiResp.Data.Groups)),
			zap.Int("totalRuleCount", totalRuleCount),
			zap.Int("candidateRuleCount", candidateNameCount),
			zap.Int("emptyExprCount", emptyExprCount),
			zap.String("ruleGroup", fallback.Group),
			zap.String("ruleName", fallback.Alert),
			zap.String("matchedBy", "name_fallback"),
			zap.Duration("elapsed", time.Since(lookupStartedAt)),
		)
		return fallback, true
	}
	systemHealthWarnCtx(ctx, "system-health vmalert rule definition lookup not matched",
		zap.String("endpoint", endpoint),
		zap.String("eventID", event.EventID),
		zap.String("alertName", event.Name),
		zap.Int("httpStatus", resp.StatusCode),
		zap.Int("groupCount", len(apiResp.Data.Groups)),
		zap.Int("totalRuleCount", totalRuleCount),
		zap.Int("candidateRuleCount", candidateNameCount),
		zap.Int("emptyExprCount", emptyExprCount),
		zap.Duration("elapsed", time.Since(lookupStartedAt)),
	)
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

func (h *handler) findAlertRuleByName(ctx core.Context, event alertmodel.AEventWithWRecord, groupID int64, syncNow bool) (*request.AlertRule, bool) {
	filter := &request.AlertRuleFilter{Alert: event.Name}
	if groupID > 0 {
		filter.GroupID = groupID
	}
	lookupStartedAt := time.Now()
	rules, totalCount := h.k8sRepo.GetAlertRules("", filter, nil, syncNow)
	systemHealthInfoCtx(ctx, "system-health k8s alert rule lookup returned",
		zap.String("eventID", event.EventID),
		zap.String("alertName", event.Name),
		zap.Int64("lookupGroupId", groupID),
		zap.Bool("syncNow", syncNow),
		zap.Int("ruleCount", len(rules)),
		zap.Int("totalCount", totalCount),
		zap.Duration("elapsed", time.Since(lookupStartedAt)),
	)
	if len(rules) == 0 && !syncNow {
		syncLookupStartedAt := time.Now()
		rules, totalCount = h.k8sRepo.GetAlertRules("", filter, nil, true)
		systemHealthInfoCtx(ctx, "system-health k8s alert rule lookup returned",
			zap.String("eventID", event.EventID),
			zap.String("alertName", event.Name),
			zap.Int64("lookupGroupId", groupID),
			zap.Bool("syncNow", true),
			zap.Int("ruleCount", len(rules)),
			zap.Int("totalCount", totalCount),
			zap.Duration("elapsed", time.Since(syncLookupStartedAt)),
		)
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
	queryStartedAt := time.Now()
	systemHealthInfoCtx(ctx, "system-health prometheus range query started",
		zap.Time("startTime", startTime),
		zap.Time("endTime", endTime),
		zap.Duration("step", step),
		zap.Int("exprLength", len(expr)),
		zap.String("exprHash", systemHealthShortHash(expr)),
	)
	value, warnings, err := h.promRepo.GetApi().QueryRange(ctx.GetContext(), expr, v1.Range{
		Start: startTime,
		End:   endTime,
		Step:  step,
	})
	if err != nil {
		systemHealthWarnCtx(ctx, "system-health prometheus range query failed",
			zap.Time("startTime", startTime),
			zap.Time("endTime", endTime),
			zap.Duration("step", step),
			zap.Int("exprLength", len(expr)),
			zap.String("exprHash", systemHealthShortHash(expr)),
			zap.Strings("promWarnings", prometheusWarnings(warnings)),
			zap.Duration("elapsed", time.Since(queryStartedAt)),
			zap.Error(err),
		)
		return nil, err
	}
	matrix, ok := value.(prommodel.Matrix)
	if !ok {
		systemHealthWarnCtx(ctx, "system-health prometheus range query failed",
			zap.Time("startTime", startTime),
			zap.Time("endTime", endTime),
			zap.Duration("step", step),
			zap.Int("exprLength", len(expr)),
			zap.String("exprHash", systemHealthShortHash(expr)),
			zap.String("valueType", fmt.Sprintf("%T", value)),
			zap.Strings("promWarnings", prometheusWarnings(warnings)),
			zap.String("reason", "unexpected_value_type"),
			zap.Duration("elapsed", time.Since(queryStartedAt)),
		)
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
	pointCount := 0
	for _, item := range series {
		pointCount += len(item.Values)
	}
	systemHealthInfoCtx(ctx, "system-health prometheus range query completed",
		zap.Time("startTime", startTime),
		zap.Time("endTime", endTime),
		zap.Duration("step", step),
		zap.Int("exprLength", len(expr)),
		zap.String("exprHash", systemHealthShortHash(expr)),
		zap.String("valueType", fmt.Sprintf("%T", value)),
		zap.Strings("promWarnings", prometheusWarnings(warnings)),
		zap.Int("seriesReturned", len(series)),
		zap.Int("pointCount", pointCount),
		zap.Bool("emptySeries", len(series) == 0),
		zap.Duration("elapsed", time.Since(queryStartedAt)),
	)
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
	sumValue := 0.0
	for _, timestamp := range timestamps {
		value := aggregated[timestamp]
		sumValue += value
		minValue = math.Min(minValue, value)
		maxValue = math.Max(maxValue, value)
	}
	averageValue := sumValue / float64(len(timestamps))
	maxDeviation := math.Max(math.Abs(maxValue-averageValue), math.Abs(minValue-averageValue))
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
		AverageValue:  averageValue,
		MinValue:      minValue,
		MaxValue:      maxValue,
		MaxDeviation:  maxDeviation,
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
		parts = append(parts, "当前状态："+alertStatusText(event.Status))
	}
	return strings.Join(parts, "，")
}

func alertStatusText(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case alertmodel.StatusFiring, "active":
		return "仍在触发"
	case alertmodel.StatusResolved, "recovering", "recovered":
		return "已恢复"
	default:
		return strings.TrimSpace(status)
	}
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
		return fmt.Sprintf("%d 个去重有效告警实例仍在触发，影响目标为 %s；最长持续 %s%s。", input.activeAlerts, input.scope, formatDurationCN(input.durationSeconds), ruleDiagnosticImpactSuffix(input))
	case issueKindInspection:
		return fmt.Sprintf("真实巡检/异常检测事件仍在触发，影响目标为 %s%s。", input.scope, ruleDiagnosticImpactSuffix(input))
	case issueKindRecoveredAlert:
		return fmt.Sprintf("%d 个去重有效告警实例已恢复，影响目标 %s 进入恢复观察%s。", effectiveAlertCount(input), input.scope, ruleDiagnosticImpactSuffix(input))
	case issueKindInvalidAlert:
		return fmt.Sprintf("%d 个告警实例更适合作为告警规则诊断信号，影响目标为 %s；优先判断告警指标、表达式或阈值是否与服务 RED 指标背离。", maxInt(input.invalidAlerts, input.deduplicatedAlerts), input.scope)
	default:
		return fmt.Sprintf("影响目标为 %s。", input.scope)
	}
}

func effectiveAlertCount(input reportInput) int {
	if input.effectiveAlerts > 0 {
		return input.effectiveAlerts
	}
	return input.deduplicatedAlerts
}

func ruleDiagnosticImpactSuffix(input reportInput) string {
	if input.kind == issueKindInvalidAlert || input.invalidAlerts <= 0 {
		return ""
	}
	return fmt.Sprintf("；另有 %d 个同目标告警实例作为规则诊断信号一并合并", input.invalidAlerts)
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
		completenessGap := dataCompletenessGap(input.availableCalcSources)
		pathRisk := 0.3684*entryImpact + 0.2632*alertSeverity + 0.2105*redDegradation + 0.1579*propagation
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
			input:               input,
			entryImpactScore:    entryImpact,
			alertSeverityScore:  alertSeverity,
			redDegradationScore: redDegradation,
			propagationScore:    propagation,
			dataCompletenessGap: completenessGap,
			pathRisk:            pathRisk,
			priority:            priority,
			status:              status,
		})
	}

	sort.Slice(scored, func(i, j int) bool {
		if scored[i].priority != scored[j].priority {
			return priorityRank(scored[i].priority) < priorityRank(scored[j].priority)
		}
		if scored[i].pathRisk != scored[j].pathRisk {
			return scored[i].pathRisk > scored[j].pathRisk
		}
		if scored[i].input.activeAlerts != scored[j].input.activeAlerts {
			return scored[i].input.activeAlerts > scored[j].input.activeAlerts
		}
		if scored[i].input.deduplicatedAlerts != scored[j].input.deduplicatedAlerts {
			return scored[i].input.deduplicatedAlerts > scored[j].input.deduplicatedAlerts
		}
		if scored[i].input.rawEventCount != scored[j].input.rawEventCount {
			return scored[i].input.rawEventCount > scored[j].input.rawEventCount
		}
		return scored[i].input.key < scored[j].input.key
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
			MetricAnalysis:        buildMetricAnalysisForIssueContext(item.input.scope, item.status, item.input.activeAlerts, item.input.evidence),
			EBPFRole:              buildEBPFRole(item),
			Impact:                item.input.impact,
			DataSources:           item.input.displaySources,
			Actions:               actionTexts(actionPlan),
			ActionPlan:            actionPlan,
			ExistingAlertAnalyses: buildIssueExistingAlertAnalyses(item.input.events),
			AlertTrend:            buildAlertTrend(item.input),
			Evidence:              item.input.evidence,
		})
	}
	return issues
}

func logSystemHealthIssueDiagnostics(ctx core.Context, scored []scoredIssue, issues []response.SystemHealthIssue) {
	for index, item := range scored {
		var issue response.SystemHealthIssue
		if index < len(issues) {
			issue = issues[index]
		}
		rootCause, rootCauseSource, rootCauseReason := buildRootCauseDecision(item)
		metricStatus, metricReason := metricTrendDiagnostic(item.input.evidence)
		reportEvidenceKinds := alertReportEvidenceKinds(item.input.alertReportFindings)
		topologyRole := dominantAlertReportTopologyRole(item.input.alertReportTopologyRoles)
		actionPlanSource := "system_default"
		if len(alertReportActionPlanSteps(item.input.alertReportActions, item.input.scope)) > 0 {
			actionPlanSource = "alert_report+system_default"
		}
		alertTrendSegmentCount := 0
		alertTrendOccurrenceCount := 0
		if issue.AlertTrend != nil {
			alertTrendSegmentCount = len(issue.AlertTrend.Segments)
			alertTrendOccurrenceCount = issue.AlertTrend.OccurrenceCount
		}
		systemHealthInfoCtx(ctx, "system-health issue diagnostics",
			zap.String("issueKey", item.input.key),
			zap.String("responseKey", issue.Key),
			zap.String("aggregationKey", item.input.aggregationKey),
			zap.String("kind", item.input.kind),
			zap.String("priority", item.priority),
			zap.String("status", item.status),
			zap.String("title", item.input.title),
			zap.String("scope", item.input.scope),
			zap.String("scopeKind", item.input.scopeKind),
			zap.Int("rawEventCount", item.input.rawEventCount),
			zap.Int("deduplicatedAlerts", item.input.deduplicatedAlerts),
			zap.Int("activeAlerts", item.input.activeAlerts),
			zap.Int("repeatCount", item.input.repeatCount),
			zap.Strings("mergeReasons", item.input.mergeReasons),
			zap.Strings("alertNames", uniqueStrings(item.input.alertNames)),
			zap.String("rootCauseSource", rootCauseSource),
			zap.String("rootCauseFallbackReason", rootCauseReason),
			zap.Bool("hasRootCause", rootCause != ""),
			zap.Int("alertReportActions", len(item.input.alertReportActions)),
			zap.Int("alertReportFindings", len(item.input.alertReportFindings)),
			zap.Int("alertReportTopologyRoles", len(item.input.alertReportTopologyRoles)),
			zap.Int("alertHistoryStats", len(item.input.alertHistoryStats)),
			zap.Int("alertHistorySegments", len(item.input.alertHistorySegments)),
			zap.Int("alertTrendSegments", alertTrendSegmentCount),
			zap.Int("alertTrendOccurrenceCount", alertTrendOccurrenceCount),
			zap.String("alertReportTopologyRole", topologyRole.Role),
			zap.String("alertReportTopologyReason", topologyRole.Reason),
			zap.Int("alertReportTopologyParentCount", topologyRole.ParentCount),
			zap.Int("alertReportTopologyChildCount", topologyRole.ChildCount),
			zap.Strings("alertReportEvidenceKinds", reportEvidenceKinds),
			zap.String("actionPlanSource", actionPlanSource),
			zap.Int("actionPlanSteps", len(issue.ActionPlan)),
			zap.String("metricTrendStatus", metricStatus),
			zap.String("metricTrendReason", metricReason),
			zap.Strings("dataSources", item.input.displaySources),
			zap.Strings("missingData", missingSourceLabels(item.input.availableCalcSources)),
		)
	}
}

func metricTrendDiagnostic(evidence []response.SystemHealthEvidence) (string, string) {
	for _, item := range evidence {
		if !strings.Contains(item.Type, "告警指标趋势") {
			continue
		}
		if item.Metric == nil {
			if strings.TrimSpace(item.Result) == "" {
				return "missing", "metric_evidence_empty"
			}
			return "unavailable", item.Result
		}
		if item.Metric.EventScoped {
			return "event_scoped", item.Metric.Message
		}
		return "global_only", item.Metric.Message
	}
	return "missing", "metric_evidence_not_built"
}

func alertReportEvidenceKinds(findings []systemHealthAlertReportFinding) []string {
	kinds := make([]string, 0, 4)
	for _, finding := range findings {
		for _, item := range finding.Evidence {
			if strings.TrimSpace(item.Kind) == "" {
				continue
			}
			kinds = append(kinds, item.Kind)
		}
	}
	return uniqueStrings(kinds)
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
	reportPlan := alertReportActionPlanSteps(item.input.alertReportActions, item.input.scope)
	if len(reportPlan) > 0 {
		return mergeActionPlanSteps(reportPlan, buildDefaultActionPlan(item))
	}
	return buildDefaultActionPlan(item)
}

func buildDefaultActionPlan(item scoredIssue) []response.SystemHealthActionPlanStep {
	switch item.input.kind {
	case issueKindRecoveredAlert:
		return buildRecoveredAlertActionPlan(item)
	case issueKindInspection:
		return appendRuleDiagnosticActionStep(item, []response.SystemHealthActionPlanStep{
			{Phase: "异常确认", Target: "确认巡检异常是否影响业务入口。", Action: "查看该影响目标相关指标趋势和关联服务状态。", Validation: "指标回到阈值内或确认无业务影响。", Risk: "巡检异常可能提前于正式告警出现。", Source: "system"},
		})
	case issueKindInvalidAlert:
		return []response.SystemHealthActionPlanStep{
			{Phase: "规则诊断", Target: "判断告警指标是否误报。", Action: "对比告警触发指标与同窗口服务 RED 指标，确认请求量、错误率和延迟是否同步异常。", Validation: "RED 指标无明显变化且告警持续被识别为无效时，进入告警规则修正流程。", Risk: "直接屏蔽告警可能掩盖真实异常，应先确认指标口径。", Source: "system"},
			{Phase: "规则修正", Target: "降低无效告警反复触发。", Action: "检查告警表达式、标签匹配、阈值、同比基线和聚合维度，必要时调整规则或静默条件。", Validation: "规则调整后无效告警数量下降，真实业务异常仍能触发有效告警。", Risk: "阈值放宽过度会降低告警灵敏度。", Source: "system"},
		}
	default:
		return appendRuleDiagnosticActionStep(item, []response.SystemHealthActionPlanStep{
			{Phase: "影响控制", Target: "优先处理仍在触发的有效告警。", Action: "进入关联告警，确认告警目标、级别、最近触发时间和影响范围。", Validation: "有效 firing 告警数量下降，关键目标状态恢复。", Risk: "只关闭告警不代表业务影响已经恢复。", Source: "system"},
			{Phase: "恢复验证", Target: "确认问题不再扩散或复发。", Action: "持续观察该影响目标的告警状态和相关指标。", Validation: "观察窗口内无新增有效告警。", Risk: "下游影响可能滞后于告警恢复。", Source: "system"},
		})
	}
}

func buildRecoveredAlertActionPlan(item scoredIssue) []response.SystemHealthActionPlanStep {
	scope := firstNonEmpty(item.input.scope, "当前影响目标")
	symptom := issueSymptomText(item)
	steps := []response.SystemHealthActionPlanStep{
		{
			Phase:      "复发观察",
			Target:     scope,
			Action:     fmt.Sprintf("回看 %s 在当前窗口内的告警触发和恢复时间，确认是否集中在同一对象反复出现。", scope),
			Validation: "确认后续观察窗口内同一对象没有新的 firing 告警，且告警间隔没有继续缩短。",
			Risk:       fmt.Sprintf("恢复态%s告警可能在流量波动或下游抖动时再次触发。", symptom),
			Source:     "system",
		},
	}

	steps = append(steps, recoveredMetricActionStep(item, scope))
	if traceEvidenceAvailable(item.input.evidence) {
		steps = append(steps, response.SystemHealthActionPlanStep{
			Phase:      "链路回看",
			Target:     scope,
			Action:     fmt.Sprintf("回看 %s 告警触发时段的 Trace 样本，确认慢调用集中在入口处理、下游调用还是资源等待。", scope),
			Validation: "确认恢复后同一入口不再出现集中慢调用或错误链路样本。",
			Risk:       "只看告警恢复可能遗漏下游调用偶发变慢，需结合链路样本确认恢复质量。",
			Source:     "system",
		})
	} else if direction := firstString(alertDirections(item.input.events)); direction != "" {
		steps = append(steps, response.SystemHealthActionPlanStep{
			Phase:      "报告复核",
			Target:     scope,
			Action:     fmt.Sprintf("按已有告警分析指向的“%s”复核相关服务或依赖，确认恢复后该方向不再出现异常信号。", truncateText(direction, 80)),
			Validation: "确认告警分析指向的服务、依赖或指标在恢复观察窗口内没有继续恶化。",
			Risk:       "已有分析方向如果未复核，重复告警可能会被误判为单次抖动。",
			Source:     "system",
		})
	}

	return normalizeActionPlanSteps(appendRuleDiagnosticActionStep(item, steps))
}

func appendRuleDiagnosticActionStep(item scoredIssue, steps []response.SystemHealthActionPlanStep) []response.SystemHealthActionPlanStep {
	if item.input.kind == issueKindInvalidAlert || item.input.invalidAlerts <= 0 {
		return steps
	}
	scope := firstNonEmpty(item.input.scope, "当前影响目标")
	steps = append(steps, response.SystemHealthActionPlanStep{
		Phase:      "规则口径核对",
		Target:     scope,
		Action:     fmt.Sprintf("同一影响目标还有 %d 个告警实例更适合作为规则诊断信号；对比告警指标与服务 RED 指标，核对表达式、阈值、标签匹配和同比基线是否合理。", item.input.invalidAlerts),
		Validation: "确认规则诊断信号没有掩盖真实业务异常；如 RED 指标无同步退化且告警仍反复出现，进入告警规则修正流程。",
		Risk:       "只处理服务风险、不核对规则口径，后续仍可能因同一规则噪声反复触发。",
		Source:     "system",
	})
	return steps
}

func recoveredMetricActionStep(item scoredIssue, scope string) response.SystemHealthActionPlanStep {
	if strings.Contains(strings.Join(item.input.alertNames, " "), "昨天") || strings.Contains(item.input.title, "昨天") {
		return response.SystemHealthActionPlanStep{
			Phase:      "基线核对",
			Target:     scope,
			Action:     fmt.Sprintf("核对 %s 与昨日同窗口的请求量和延时基线，确认告警是否由低流量或基线差异放大。", scope),
			Validation: "确认同比基线、当前绝对延时和请求量变化能够解释本次恢复态告警。",
			Risk:       "同比类告警容易受昨日基线影响，未核对基线可能误判风险程度。",
			Source:     "system",
		}
	}
	if usableMetricEvidenceAvailable(item.input.evidence) {
		return response.SystemHealthActionPlanStep{
			Phase:      "指标复核",
			Target:     scope,
			Action:     fmt.Sprintf("对比 %s 的告警指标、入口 RED 延时、请求量和错误率，确认恢复后指标没有继续处于高位。", scope),
			Validation: "确认告警指标回落后，入口延时、请求量和错误率没有同步退化。",
			Risk:       "单个告警指标恢复不代表业务影响完全消失，仍需确认 RED 指标一致恢复。",
			Source:     "system",
		}
	}
	if metricEvidencePresent(item.input.evidence) {
		return response.SystemHealthActionPlanStep{
			Phase:      "指标补齐",
			Target:     scope,
			Action:     fmt.Sprintf("补齐 %s 的指标查询链路，优先确认告警规则表达式、标签匹配和入口 RED 指标是否可回看。", scope),
			Validation: "确认同一对象能查到延时、请求量和错误率趋势，再用指标验证恢复是否稳定。",
			Risk:       "指标趋势不可用时，只能依赖告警事件判断恢复状态，容易漏掉持续退化。",
			Source:     "system",
		}
	}
	return response.SystemHealthActionPlanStep{
		Phase:      "指标补齐",
		Target:     scope,
		Action:     fmt.Sprintf("补齐 %s 的入口延时、请求量和错误率趋势，用于验证恢复是否稳定。", scope),
		Validation: "确认恢复观察窗口内关键 RED 指标没有继续退化。",
		Risk:       "缺少指标趋势时，恢复判断主要依赖告警状态，无法充分判断用户侧影响。",
		Source:     "system",
	}
}

func normalizeActionPlanSteps(steps []response.SystemHealthActionPlanStep) []response.SystemHealthActionPlanStep {
	result := make([]response.SystemHealthActionPlanStep, 0, len(steps))
	for _, step := range steps {
		result = append(result, normalizeActionPlanStep(step))
	}
	return result
}

func metricEvidencePresent(evidence []response.SystemHealthEvidence) bool {
	for _, item := range evidence {
		if strings.Contains(item.Type, "指标趋势") && !strings.Contains(item.Type, "请求北极星") {
			return true
		}
	}
	return false
}

func usableMetricEvidenceAvailable(evidence []response.SystemHealthEvidence) bool {
	for _, item := range evidence {
		if strings.Contains(item.Type, "告警指标趋势") && item.Metric != nil && item.Metric.PointCount > 0 {
			return true
		}
	}
	return false
}

func traceEvidenceAvailable(evidence []response.SystemHealthEvidence) bool {
	for _, item := range evidence {
		if strings.Contains(item.Type, "Trace") || strings.Contains(item.Type, "APM链路") {
			return true
		}
	}
	return false
}

func alertReportActionPlanSteps(actions []systemHealthAlertReportAction, scope string) []response.SystemHealthActionPlanStep {
	steps := make([]response.SystemHealthActionPlanStep, 0, len(actions))
	for _, action := range actions {
		actionText := strings.TrimSpace(action.Action)
		if actionText == "" {
			continue
		}
		target := firstNonEmpty(action.Target, action.AlertName, scope)
		steps = append(steps, normalizeActionPlanStep(response.SystemHealthActionPlanStep{
			Phase:      firstNonEmpty(action.Phase, alertReportActionPhase(action.Level)),
			Target:     target,
			Action:     actionText,
			Validation: firstNonEmpty(action.Validation, "执行后回看关联告警、关键指标和报告证据，确认风险已收敛。"),
			Risk:       firstNonEmpty(action.Risk, "来自原告警分析报告，执行前需结合当前系统健康状态确认适用性。"),
			Source:     "alertReport",
		}))
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
			Source: "ai",
		})
	}
	return steps
}

func markActionPlanSource(steps []response.SystemHealthActionPlanStep, source string) []response.SystemHealthActionPlanStep {
	result := make([]response.SystemHealthActionPlanStep, 0, len(steps))
	for _, step := range steps {
		if strings.TrimSpace(step.Source) == "" {
			step.Source = source
		}
		result = append(result, step)
	}
	return result
}

func normalizeAIActionPlanTargets(steps []response.SystemHealthActionPlanStep, issue response.SystemHealthIssue) []response.SystemHealthActionPlanStep {
	result := make([]response.SystemHealthActionPlanStep, 0, len(steps))
	for _, step := range steps {
		target := strings.TrimSpace(step.Target)
		if isInternalIssueKeyTarget(target, issue) {
			step.Target = strings.TrimSpace(issue.Scope)
		}
		result = append(result, step)
	}
	return result
}

func isInternalIssueKeyTarget(target string, issue response.SystemHealthIssue) bool {
	target = strings.TrimSpace(target)
	if target == "" {
		return false
	}
	normalized := strings.ToLower(target)
	if issue.Key != "" && normalized == strings.ToLower(strings.TrimSpace(issue.Key)) {
		return true
	}
	if issue.AggregationKey != "" && normalized == strings.ToLower(stableKey(issue.AggregationKey)) {
		return true
	}
	return regexp.MustCompile(`^(activealert|recoveredalert|invalidalert|inspection)-(entry|service|object|group)-[a-z0-9-]+$`).MatchString(normalized)
}

func mergeActionPlanSteps(primary []response.SystemHealthActionPlanStep, secondary []response.SystemHealthActionPlanStep) []response.SystemHealthActionPlanStep {
	merged := make([]response.SystemHealthActionPlanStep, 0, len(primary)+len(secondary))
	seen := make(map[string]struct{}, len(primary)+len(secondary))
	appendStep := func(step response.SystemHealthActionPlanStep) {
		step = normalizeActionPlanStep(step)
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

func normalizeActionPlanStep(step response.SystemHealthActionPlanStep) response.SystemHealthActionPlanStep {
	step.Phase = strings.TrimSpace(step.Phase)
	step.Target = strings.TrimSpace(step.Target)
	step.Action = strings.TrimSpace(step.Action)
	step.Validation = normalizeActionPlanValidation(step.Validation)
	step.Risk = strings.TrimSpace(step.Risk)
	step.Source = strings.TrimSpace(step.Source)
	return step
}

func normalizeActionPlanValidation(validation string) string {
	text := strings.TrimSpace(validation)
	if text == "" {
		return ""
	}
	text = strings.TrimSpace(strings.TrimPrefix(text, "验证："))
	text = strings.TrimSpace(strings.TrimPrefix(text, "验证:"))
	text = ensureChinesePeriod(text)
	normalized := strings.TrimSuffix(text, "。")
	if normalized == "日志无持续慢请求或异常堆积" || normalized == "日志中无持续慢请求或异常堆积" {
		return "检查日志中是否仍存在持续慢请求或异常堆积，并确认恢复后不再新增。"
	}
	if startsWithAny(text, []string{"确认", "检查", "观察", "回看", "查看", "验证", "对比", "持续观察", "执行后"}) {
		return text
	}
	return "确认" + text
}

func ensureChinesePeriod(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	if strings.HasSuffix(text, "。") || strings.HasSuffix(text, "；") || strings.HasSuffix(text, "！") || strings.HasSuffix(text, "？") || strings.HasSuffix(text, ".") || strings.HasSuffix(text, ";") || strings.HasSuffix(text, "!") || strings.HasSuffix(text, "?") {
		return text
	}
	return text + "。"
}

func startsWithAny(text string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(text, prefix) {
			return true
		}
	}
	return false
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
	if input.scopeKind == "entry" || input.scopeKind == "serviceEndpoint" || input.scopeKind == "service" {
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

func dataCompletenessGap(available []string) float64 {
	missing := len(missingSourceLabels(available))
	if len(expectedSources) == 0 {
		return 0
	}
	return math.Min(float64(missing)*100/float64(len(expectedSources)), 100)
}

func sourceLabel(source string) string {
	if label, ok := sourceLabels[source]; ok {
		return label
	}
	return source
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
		EntryImpactScore:    int(math.Round(item.entryImpactScore)),
		AlertSeverityScore:  int(math.Round(item.alertSeverityScore)),
		RedDegradationScore: int(math.Round(item.redDegradationScore)),
		PropagationScore:    int(math.Round(item.propagationScore)),
		PathRisk:            int(math.Round(item.pathRisk)),
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
		"入口/目标影响 " + strconvRound(item.entryImpactScore),
		"告警严重度 " + strconvRound(item.alertSeverityScore),
		"RED退化 " + strconvRound(item.redDegradationScore),
		"传播风险 " + strconvRound(item.propagationScore),
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
		return "活跃告警仍在触发，且窗口内出现反复事件"
	}
	if activeCount > 0 {
		return "活跃告警尚未清零"
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
		return fmt.Sprintf("%s%s%s告警%s，目前告警已清零，相关影响目标处于恢复观察", eventPrefix, reportScopePatternText(scored), reportSymptomText(scored), repeatText)
	}
	return "当前窗口未发现持续触发的核心告警，系统处于观察状态"
}

func reportScopePatternText(scored []scoredIssue) string {
	countByKind := map[string]int{}
	for _, item := range scored {
		countByKind[item.input.scopeKind]++
	}
	if countByKind["entry"] > 0 && countByKind["serviceEndpoint"] > 0 && len(scored) > 3 {
		return "多入口/多服务接口"
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
	if countByKind["serviceEndpoint"] > 0 {
		if len(scored) > 1 {
			return "多服务接口"
		}
		return "单服务接口"
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
		return "多目标"
	}
	return "单目标"
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
	if item.dataCompletenessGap == 0 {
		return []string{"真实告警事件已覆盖", "未使用本地内置样例数据"}
	}
	return []string{
		"真实数据源覆盖不完整，仅作为数据边界展示，不参与健康度计算",
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
		return "该影响目标存在持续触发但被识别为无效的告警，优先按 P4 规则诊断处理：对比告警触发指标与服务 RED 指标，确认是否为告警表达式、阈值、标签匹配或同比基线问题。"
	}
	state := "当前仍在触发"
	if item.input.activeAlerts == 0 {
		state = "当前已恢复，进入恢复观察"
	}
	symptom := "健康风险"
	for _, name := range item.input.alertNames {
		if strings.Contains(name, "延时") || strings.Contains(name, "延迟") || strings.Contains(strings.ToLower(name), "latency") {
			symptom = "延时类告警"
			break
		}
		if strings.Contains(name, "CPU") || strings.Contains(name, "内存") || strings.Contains(strings.ToLower(name), "memory") {
			symptom = "资源类告警"
			break
		}
	}
	metricDirection := buildMetricAnalysisForIssueContext(item.input.scope, item.status, item.input.activeAlerts, item.input.evidence)
	if metricDirection == "" {
		metricDirection = buildRootCause(item)
	}
	if metricDirection != "" {
		return fmt.Sprintf("%s 出现过有效%s，%s。指标/eBPF 方向指向：%s", item.input.scope, symptom, state, truncateText(metricDirection, 180))
	}
	return fmt.Sprintf("%s 出现过有效%s，%s。需要结合 eBPF/RED 指标和调用链继续确认入口、下游或资源方向。", item.input.scope, symptom, state)
}

func buildRootCause(item scoredIssue) string {
	rootCause, _, _ := buildRootCauseDecision(item)
	return rootCause
}

func buildRootCauseDecision(item scoredIssue) (string, string, string) {
	if item.input.kind == issueKindInvalidAlert {
		return "告警本身持续触发但已有分析判定无效，根因优先考虑告警表达式、阈值、标签匹配或指标口径问题。", "invalid_alert_rule_diagnostic", "issue_kind_invalid_alert"
	}
	if reportRootCause := buildRootCauseFromAlertReports(item.input.alertReportFindings); reportRootCause != "" {
		return reportRootCause, "alert_report", fmt.Sprintf("alert_report_findings=%d", len(uniqueAlertReportFindings(item.input.alertReportFindings)))
	}
	if requestPolarisRootCause := buildRootCauseFromRequestPolaris(item.input.evidence); requestPolarisRootCause != "" {
		return requestPolarisRootCause, "request_polaris", "alert_report_root_cause_missing_request_polaris_abnormal"
	}
	directions := alertDirections(item.input.events)
	if len(directions) == 0 {
		return "", "empty", "no_alert_report_findings_and_no_alert_direction"
	}
	roots := make([]string, 0, len(directions))
	for _, direction := range directions {
		if root := rootCauseFromAlertDirection(direction); root != "" {
			roots = append(roots, root)
		}
	}
	roots = uniqueStrings(roots)
	if len(roots) == 0 {
		return "", "empty", "alert_direction_cleaned_empty"
	}
	if len(roots) == 1 {
		return "综合告警分析和事件表现，可能根因集中在：" + truncateText(roots[0], 180), "alert_direction_fallback", "alert_report_root_cause_missing"
	}
	return "综合告警分析和事件表现，可能根因包括：" + truncateText(strings.Join(roots[:minInt(len(roots), 3)], "；"), 220), "alert_direction_fallback", "alert_report_root_cause_missing"
}

func buildRootCauseFromAlertReports(findings []systemHealthAlertReportFinding) string {
	findings = uniqueAlertReportFindings(findings)
	if len(findings) == 0 {
		return ""
	}
	rootCauses := make([]string, 0, len(findings))
	details := make([]string, 0, len(findings))
	for _, finding := range findings {
		if rootCause := cleanAlertReportDirectionalText(finding.RootCause); rootCause != "" {
			rootCauses = append(rootCauses, rootCause)
		}
		rootCauses = append(rootCauses, cleanAlertReportDirectionalTexts(finding.RootCauseOther, 6, 260)...)
		if analyzeProcess := cleanAlertReportDirectionalText(finding.AnalyzeProcess); analyzeProcess != "" {
			details = append(details, analyzeProcess)
		} else if detail := cleanAlertReportDirectionalText(finding.Detail); detail != "" {
			details = append(details, detail)
		}
	}
	rootCauses = uniqueStrings(rootCauses)
	details = uniqueStrings(details)
	parts := make([]string, 0, 2)
	if len(rootCauses) > 0 {
		parts = append(parts, "告警分析报告判断："+truncateText(strings.Join(rootCauses[:minInt(len(rootCauses), 3)], "；"), 260))
	}
	if len(details) > 0 {
		parts = append(parts, "报告详情："+truncateText(strings.Join(details[:minInt(len(details), 2)], "；"), 220))
	}
	return truncateText(strings.Join(parts, "。"), 420)
}

func buildRootCauseFromRequestPolaris(evidence []response.SystemHealthEvidence) string {
	for _, item := range evidence {
		if !strings.Contains(item.Type, "请求北极星") {
			continue
		}
		result := cleanUserFacingReportText(item.Result)
		if result == "" {
			continue
		}
		if strings.Contains(result, "未见网络耗时") || strings.Contains(result, "未见请求资源耗时") || strings.Contains(result, "未见超过") || strings.Contains(result, "检索失败") || strings.Contains(result, "不把请求资源耗时作为主要根因") || strings.Contains(result, "不把资源分项耗时作为主要根因") || strings.Contains(result, "不代表全局资源耗时已异常") || strings.Contains(result, "不代表整体资源耗时已异常") {
			continue
		}
		if !strings.Contains(result, "异常") && !strings.Contains(result, "升高") && !strings.Contains(result, "达到") && !strings.Contains(result, "放大") {
			continue
		}
		return "请求北极星显示单次请求资源耗时存在异常波动，可能根因集中在：" + truncateText(result, 260)
	}
	return ""
}

type metricTimeRangeSummary struct {
	StartAt string
	EndAt   string
	PeakAt  string
	MinAt   string
}

func buildMetricAnalysis(evidence []response.SystemHealthEvidence) string {
	return buildMetricAnalysisForIssueContext("该影响目标", "", 0, evidence)
}

func buildMetricAnalysisForIssueContext(scope string, status string, activeAlerts int, evidence []response.SystemHealthEvidence) string {
	for _, item := range evidence {
		if !strings.Contains(item.Type, "告警指标趋势") || item.Metric == nil {
			continue
		}
		metric := item.Metric
		subject := strings.TrimSpace(scope)
		if subject == "" {
			subject = "该影响目标"
		}
		metricLabel := metricUserFacingLabel(metric)
		if metric.PointCount <= 0 {
			return fmt.Sprintf("%s的%s暂未获取到足够采样点，暂时无法判断指标是在升高还是回落。", subject, metricLabel)
		}
		if !metric.EventScoped {
			return fmt.Sprintf("%s的%s未能精确匹配到当前告警事件对应的时序，当前趋势只能作为背景参考，不能直接证明该影响目标仍异常。", subject, metricLabel)
		}
		changeText := ""
		if metric.ChangePercent != nil {
			changeText = "，变化约 " + formatSignedMetricFloat(*metric.ChangePercent) + "%"
		}
		timeText := "在当前窗口内"
		if rangeText := metricTimeRangePhrase(metric); rangeText != "" {
			timeText = "在 " + rangeText
		}
		peakText := metricPeakPhrase(metric)
		meaning := metricRiskMeaning(metric, status, activeAlerts)
		switch {
		case metric.Delta < 0:
			return fmt.Sprintf("%s的%s显示，%s从 %s 回落到 %s%s%s；%s",
				subject,
				metricLabel,
				timeText,
				formatMetricFloat(metric.FirstValue),
				formatMetricFloat(metric.LastValue),
				changeText,
				peakText,
				meaning,
			)
		case metric.Delta > 0:
			return fmt.Sprintf("%s的%s显示，%s从 %s 升高到 %s%s%s；%s",
				subject,
				metricLabel,
				timeText,
				formatMetricFloat(metric.FirstValue),
				formatMetricFloat(metric.LastValue),
				changeText,
				peakText,
				meaning,
			)
		default:
			return fmt.Sprintf("%s的%s显示，%s基本持平，数值约为 %s%s；%s",
				subject,
				metricLabel,
				timeText,
				formatMetricFloat(metric.LastValue),
				peakText,
				meaning,
			)
		}
	}
	return ""
}

func metricUserFacingLabel(metric *response.SystemHealthMetricEvidence) string {
	if metric == nil {
		return "相关指标"
	}
	if metricName := strings.TrimSpace(metric.MetricName); metricName != "" {
		return metricName
	}
	return inferMetricNameFromRule(metric.RuleName, metric.Expression)
}

func inferMetricNameFromRule(ruleName string, expression string) string {
	signatureText := strings.Join([]string{
		strings.TrimSpace(ruleName),
		strings.TrimSpace(expression),
	}, " ")
	signature := strings.ToLower(signatureText)
	dayOverDay := containsAnyText(signatureText, "相比昨天", "较昨天", "比昨天", "昨日", "昨天", "日同比") ||
		strings.Contains(signature, "offset 24h") ||
		strings.Contains(signature, "day_over_day") ||
		strings.Contains(signature, "dayoverday") ||
		strings.Contains(signature, "dod")

	logMetric := containsAnyText(signatureText, "日志") || strings.Contains(signature, "log")
	latencyMetric := containsAnyText(signatureText, "平均请求延时", "平均延时", "请求延时", "请求延迟", "延时", "延迟", "耗时", "超时", "慢请求") ||
		strings.Contains(signature, "latency") ||
		strings.Contains(signature, "duration") ||
		strings.Contains(signature, "response_time") ||
		strings.Contains(signature, "request_time") ||
		strings.Contains(signature, "rt_") ||
		strings.Contains(signature, "p95") ||
		strings.Contains(signature, "p99")
	errorMetric := containsAnyText(signatureText, "请求错误率", "错误率", "错误", "异常", "失败") ||
		strings.Contains(signature, "error") ||
		strings.Contains(signature, "exception") ||
		strings.Contains(signature, "fail") ||
		strings.Contains(signature, "5xx") ||
		strings.Contains(signature, "4xx")

	switch {
	case logMetric && errorMetric && dayOverDay:
		return "日志错误数日同比"
	case logMetric && errorMetric:
		return "日志错误数"
	case latencyMetric && dayOverDay:
		return "请求延时日同比"
	case latencyMetric && (containsAnyText(signatureText, "平均请求延时", "平均延时") || strings.Contains(signature, "avg") || strings.Contains(signature, "average") || strings.Contains(signature, "mean")):
		return "平均请求延时"
	case latencyMetric:
		return "请求延时"
	case errorMetric && dayOverDay:
		return "请求错误率日同比"
	case errorMetric:
		return "请求错误率"
	case strings.Contains(signature, "cpu"):
		if dayOverDay {
			return "CPU 使用率日同比"
		}
		return "CPU 使用率"
	case strings.Contains(signature, "memory") || containsAnyText(signatureText, "内存"):
		if dayOverDay {
			return "内存使用率日同比"
		}
		return "内存使用率"
	case strings.Contains(signature, "disk") || containsAnyText(signatureText, "磁盘"):
		if dayOverDay {
			return "磁盘使用率日同比"
		}
		return "磁盘使用率"
	default:
		return "相关指标"
	}
}

func containsAnyText(text string, patterns ...string) bool {
	for _, pattern := range patterns {
		if pattern != "" && strings.Contains(text, pattern) {
			return true
		}
	}
	return false
}

func metricTrendPhrase(metric *response.SystemHealthMetricEvidence) string {
	if metric == nil || metric.PointCount <= 0 {
		return "无法判断"
	}
	switch {
	case metric.Delta < 0:
		return "整体回落"
	case metric.Delta > 0:
		return "整体升高"
	default:
		return "基本持平"
	}
}

func metricRiskMeaning(metric *response.SystemHealthMetricEvidence, status string, activeAlerts int) string {
	trend := metricTrendPhrase(metric)
	active := activeAlerts > 0 || strings.EqualFold(status, "active") || strings.EqualFold(status, "firing")
	if active {
		switch trend {
		case "整体回落":
			return "虽然指标已有缓解，但告警仍在触发，需要继续确认是否已经脱离影响区间"
		case "整体升高":
			return "告警仍在触发且指标继续抬升，说明当前风险仍有指标支撑，需要优先确认用户请求是否受影响"
		default:
			return "告警仍在触发，指标未明显回落，需要结合 RED 和调用链确认影响是否持续"
		}
	}
	switch trend {
	case "整体回落":
		return "随着指标回落且告警已清零，本次风险更适合作为恢复观察和复发监控"
	case "整体升高":
		return "虽然当前告警已清零，但窗口内指标仍有抬升，需要观察后续是否复发"
	default:
		return "当前告警已清零，指标未继续放大，后续重点观察是否再次触发"
	}
}

func metricTimeRangeText(metric *response.SystemHealthMetricEvidence) metricTimeRangeSummary {
	points := aggregateMetricEvidencePoints(metric)
	if len(points) == 0 {
		return metricTimeRangeSummary{}
	}
	start := points[0]
	end := points[len(points)-1]
	peak := start
	minPoint := start
	for _, point := range points {
		if point.Value > peak.Value {
			peak = point
		}
		if point.Value < minPoint.Value {
			minPoint = point
		}
	}
	return metricTimeRangeSummary{
		StartAt: formatMetricTimestamp(start.Timestamp),
		EndAt:   formatMetricTimestamp(end.Timestamp),
		PeakAt:  formatMetricTimestamp(peak.Timestamp),
		MinAt:   formatMetricTimestamp(minPoint.Timestamp),
	}
}

func metricTimeRangePhrase(metric *response.SystemHealthMetricEvidence) string {
	timeRange := metricTimeRangeText(metric)
	if timeRange.StartAt == "" || timeRange.EndAt == "" {
		return ""
	}
	if timeRange.StartAt == timeRange.EndAt {
		return timeRange.StartAt + " 附近"
	}
	return timeRange.StartAt + " 到 " + timeRange.EndAt
}

func metricPeakPhrase(metric *response.SystemHealthMetricEvidence) string {
	timeRange := metricTimeRangeText(metric)
	if timeRange.PeakAt == "" || metric == nil || metric.PointCount <= 0 {
		return ""
	}
	if math.Abs(metric.MaxValue-metric.FirstValue) < 1e-9 && math.Abs(metric.MaxValue-metric.LastValue) < 1e-9 {
		return ""
	}
	return "，峰值约 " + formatMetricFloat(metric.MaxValue) + "（" + timeRange.PeakAt + "）"
}

func aggregateMetricEvidencePoints(metric *response.SystemHealthMetricEvidence) []response.SystemHealthMetricDataPoint {
	if metric == nil || len(metric.Series) == 0 {
		return nil
	}
	aggregated := make(map[int64]float64)
	for _, series := range metric.Series {
		for _, point := range series.Points {
			if point.Timestamp <= 0 || math.IsNaN(point.Value) || math.IsInf(point.Value, 0) {
				continue
			}
			aggregated[point.Timestamp] += point.Value
		}
	}
	if len(aggregated) == 0 {
		return nil
	}
	timestamps := make([]int64, 0, len(aggregated))
	for timestamp := range aggregated {
		timestamps = append(timestamps, timestamp)
	}
	sort.Slice(timestamps, func(i, j int) bool { return timestamps[i] < timestamps[j] })
	points := make([]response.SystemHealthMetricDataPoint, 0, len(timestamps))
	for _, timestamp := range timestamps {
		points = append(points, response.SystemHealthMetricDataPoint{
			Timestamp: timestamp,
			Value:     aggregated[timestamp],
		})
	}
	return points
}

func formatMetricTimestamp(timestamp int64) string {
	if timestamp <= 0 {
		return ""
	}
	var value time.Time
	switch {
	case timestamp > 1e15:
		value = time.Unix(0, timestamp*int64(time.Microsecond))
	case timestamp > 1e12:
		value = time.UnixMilli(timestamp)
	case timestamp > 1e9:
		value = time.Unix(timestamp, 0)
	default:
		return ""
	}
	return value.In(time.Local).Format("2006-01-02 15:04")
}

func containsMetricQuerySyntax(text string) bool {
	normalized := strings.ToLower(strings.TrimSpace(text))
	if normalized == "" {
		return false
	}
	for _, marker := range []string{
		"pql",
		"promql",
		"query=",
		"expr=",
		"查询表达式",
		"prometheus 查询",
		"rate(",
		"irate(",
		"increase(",
		"sum by",
		"avg by",
		"max by",
		"min by",
		"histogram_quantile(",
		"last_over_time(",
		"avg_over_time(",
		"max_over_time(",
		"min_over_time(",
	} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}

func rootCauseFromAlertDirection(direction string) string {
	root := cleanUserFacingReportText(direction)
	if root == "" {
		return ""
	}
	for {
		trimmed := root
		for _, prefix := range []string{"告警发生时分析指向：", "告警发生时分析指向:", "告警发生时分析指向", "已有告警分析指向：", "已有告警分析指向:", "已有告警分析指向", "疑似", "可能是", "可能为", "可能", "初步判断", "分析认为", "根因可能是", "根因可能为"} {
			trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
		}
		if trimmed == root {
			break
		}
		root = trimmed
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

func cleanUserFacingReportText(value string) string {
	text := strings.TrimSpace(value)
	if text == "" {
		return ""
	}
	text = html.UnescapeString(text)
	text = strings.ReplaceAll(text, "\\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = markdownCodeFencePattern.ReplaceAllString(text, " ")
	text = htmlDetailsBlockPattern.ReplaceAllString(text, " ")
	text = htmlTagPattern.ReplaceAllString(text, " ")
	text = strings.NewReplacer(
		"**", "",
		"__", "",
		"`", "",
		"&nbsp;", " ",
		"\u00a0", " ",
	).Replace(text)

	lines := strings.Split(text, "\n")
	kept := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(strings.TrimLeft(line, "-*#> \t"))
		line = stripInternalReasoningPrefix(line)
		line = truncateAtInternalReasoningMarker(line)
		if line == "" || isInternalReasoningLine(line) || isInternalReportHeadingOnly(line) {
			continue
		}
		kept = append(kept, line)
	}
	text = strings.Join(kept, " ")
	text = strings.Join(strings.Fields(text), " ")
	text = stripInternalReasoningPrefix(text)
	text = truncateAtInternalReasoningMarker(text)
	if isInternalReasoningLine(text) || isInternalReportHeadingOnly(text) {
		return ""
	}
	return text
}

func cleanAlertReportDirectionalText(value string) string {
	text := cleanUserFacingReportText(value)
	if text == "" || isNonDiagnosticDirectionText(text) {
		return ""
	}
	return text
}

func cleanAlertReportDirectionalTexts(values []string, limit int, textLimit int) []string {
	items := make([]string, 0, minInt(len(values), limit))
	for _, value := range values {
		if len(items) >= limit {
			break
		}
		text := cleanAlertReportDirectionalText(value)
		if text == "" {
			continue
		}
		items = append(items, truncateText(text, textLimit))
	}
	return uniqueStrings(items)
}

func isNonDiagnosticDirectionText(value string) bool {
	text := strings.TrimSpace(strings.Trim(value, "：:;；,.，。 "))
	if text == "" {
		return false
	}
	compact := strings.ToLower(strings.Join(strings.Fields(text), ""))
	markers := []string{
		"无明确故障方向",
		"无明显故障方向",
		"没有明确故障方向",
		"没有明显故障方向",
		"未发现明确故障方向",
		"未发现明显故障方向",
		"未发现故障方向",
		"未定位到明确故障方向",
		"未定位到明确根因",
		"无法判断故障方向",
		"无法明确故障方向",
		"无法判断根因",
		"无法明确根因",
		"暂无法判断故障方向",
		"暂无法定位根因",
		"暂未定位到根因",
		"暂无明确根因",
		"无明确根因",
	}
	for _, marker := range markers {
		if strings.Contains(compact, marker) {
			return true
		}
	}
	return false
}

func ensureEBPFRootCauseContext(value string) string {
	text := strings.TrimSpace(value)
	if text == "" || strings.Contains(strings.ToLower(text), "ebpf") || strings.Contains(text, "eBPF") {
		return text
	}
	return "基于 eBPF 数据的方向报告分析，" + text
}

func isInternalReportHeadingOnly(value string) bool {
	text := strings.TrimSpace(strings.Trim(value, "：:;；,.，。 "))
	switch text {
	case "告警发生时分析指向", "已有告警分析指向", "告警分析指向", "分析指向":
		return true
	default:
		return false
	}
}

func stripInternalReasoningPrefix(value string) string {
	text := strings.TrimSpace(value)
	if text == "" || !cjkTextPattern.MatchString(text) {
		return text
	}
	lower := strings.ToLower(text)
	prefixMarkers := []string{
		"summarizing",
		"summary",
		"i need",
		"i should",
		"i'll",
		"let me",
		"based on",
		"the user",
		"we need",
	}
	markerIndex := -1
	for _, marker := range prefixMarkers {
		index := strings.Index(lower, marker)
		if index < 0 {
			continue
		}
		if markerIndex < 0 || index < markerIndex {
			markerIndex = index
		}
	}
	if markerIndex < 0 {
		return text
	}
	firstCJK := cjkTextPattern.FindStringIndex(text)
	if len(firstCJK) != 2 || firstCJK[0] <= markerIndex {
		return text
	}
	return strings.TrimSpace(text[firstCJK[0]:])
}

func truncateAtInternalReasoningMarker(value string) string {
	text := strings.TrimSpace(value)
	if text == "" {
		return ""
	}
	lower := strings.ToLower(text)
	markers := []string{
		"thinking",
		"extracting fault",
		"chain-of-thought",
		"internal reasoning",
		"i need",
		"i should",
		"i'll",
		"we need",
		"let's craft",
		"final answer",
	}
	markerIndex := -1
	for _, marker := range markers {
		index := strings.Index(lower, marker)
		if index < 0 {
			continue
		}
		if markerIndex < 0 || index < markerIndex {
			markerIndex = index
		}
	}
	if markerIndex <= 0 {
		return text
	}
	return strings.TrimSpace(strings.TrimRight(text[:markerIndex], "：:;；,.，。 "))
}

func isInternalReasoningLine(value string) bool {
	text := strings.TrimSpace(value)
	if text == "" {
		return false
	}
	lower := strings.ToLower(text)
	if cjkTextPattern.MatchString(text) {
		return strings.Contains(lower, "<details") || strings.Contains(lower, "</details>")
	}
	internalMarkers := []string{
		"thinking",
		"summarizing",
		"extracting fault",
		"chain-of-thought",
		"internal reasoning",
		"i need a concise",
		"i need to",
		"i should",
		"i need answer",
		"we need answer",
		"we need to answer",
		"let's craft",
		"final answer",
	}
	for _, marker := range internalMarkers {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return strings.Contains(lower, "<details") || strings.Contains(lower, "</details>")
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
