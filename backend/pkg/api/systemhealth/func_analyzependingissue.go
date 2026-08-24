// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package systemhealth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"go.uber.org/zap"
)

type systemHealthPendingIssueAIRequest struct {
	Window string                            `json:"window"`
	Issue  response.SystemHealthPendingIssue `json:"issue"`
}

type systemHealthPendingIssueAIResponse struct {
	WorkflowRunID string                                `json:"workflowRunId"`
	Analysis      string                                `json:"analysis"`
	Impact        string                                `json:"impact"`
	ActionPlan    []response.SystemHealthActionPlanStep `json:"actionPlan"`
	Validation    string                                `json:"validation"`
	Risk          string                                `json:"risk"`
	DataBoundary  string                                `json:"dataBoundary"`
}

const defaultSystemHealthPendingIssueAIChatMaxTokens = 1024
const minSystemHealthPendingIssueAIChatRetryMaxTokens = 4096
const pendingIssueActionTextLimit = 60
const pendingIssueActionCheckTextLimit = 40
const pendingIssueActionTargetTextLimit = 60

var pendingIssueFixedActionPhases = []string{"定位", "排查", "处置"}

type openAIChatCompletionsResponse struct {
	ID      string                        `json:"id"`
	Choices []openAIChatCompletionsChoice `json:"choices"`
	Error   *openAIResponsesAPIError      `json:"error,omitempty"`
}

type openAIChatCompletionsChoice struct {
	Message      openAIChatCompletionsMessage `json:"message"`
	FinishReason string                       `json:"finish_reason"`
}

type openAIChatCompletionsMessage struct {
	Content any `json:"content"`
}

type systemHealthPendingIssueAnalysisCache struct {
	mutex      sync.RWMutex
	ttl        time.Duration
	maxEntries int
	entries    map[string]systemHealthPendingIssueAnalysisCacheEntry
}

type systemHealthPendingIssueAnalysisCacheEntry struct {
	resp      response.SystemHealthPendingIssueAnalyzeResp
	createdAt time.Time
	expiresAt time.Time
}

func newSystemHealthPendingIssueAnalysisCache(ttl time.Duration, maxEntries int) *systemHealthPendingIssueAnalysisCache {
	if maxEntries <= 0 {
		maxEntries = 512
	}
	return &systemHealthPendingIssueAnalysisCache{
		ttl:        ttl,
		maxEntries: maxEntries,
		entries:    make(map[string]systemHealthPendingIssueAnalysisCacheEntry),
	}
}

func (c *systemHealthPendingIssueAnalysisCache) get(key string, now time.Time) (response.SystemHealthPendingIssueAnalyzeResp, bool) {
	if c == nil || key == "" {
		return response.SystemHealthPendingIssueAnalyzeResp{}, false
	}
	c.mutex.RLock()
	entry, ok := c.entries[key]
	c.mutex.RUnlock()
	if !ok || (!entry.expiresAt.IsZero() && now.After(entry.expiresAt)) {
		return response.SystemHealthPendingIssueAnalyzeResp{}, false
	}
	return cloneSystemHealthPendingIssueAnalyzeResp(entry.resp), true
}

func (c *systemHealthPendingIssueAnalysisCache) set(key string, resp response.SystemHealthPendingIssueAnalyzeResp, now time.Time) {
	if c == nil || key == "" || resp.AIStatus != "succeeded" {
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for existingKey, entry := range c.entries {
		if !entry.expiresAt.IsZero() && now.After(entry.expiresAt) {
			delete(c.entries, existingKey)
		}
	}
	if len(c.entries) >= c.maxEntries {
		var oldestKey string
		var oldestCreatedAt time.Time
		for existingKey, entry := range c.entries {
			if oldestKey == "" || entry.createdAt.Before(oldestCreatedAt) {
				oldestKey = existingKey
				oldestCreatedAt = entry.createdAt
			}
		}
		delete(c.entries, oldestKey)
	}

	expiresAt := time.Time{}
	if c.ttl > 0 {
		expiresAt = now.Add(c.ttl)
	}
	c.entries[key] = systemHealthPendingIssueAnalysisCacheEntry{
		resp:      cloneSystemHealthPendingIssueAnalyzeResp(resp),
		createdAt: now,
		expiresAt: expiresAt,
	}
}

func systemHealthPendingIssueAnalysisCacheKey(window string, issue response.SystemHealthPendingIssue) string {
	raw, _ := json.Marshal(struct {
		Window string                            `json:"window"`
		Issue  response.SystemHealthPendingIssue `json:"issue"`
	}{
		Window: window,
		Issue:  issue,
	})
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func cloneSystemHealthPendingIssueAnalyzeResp(resp response.SystemHealthPendingIssueAnalyzeResp) response.SystemHealthPendingIssueAnalyzeResp {
	if resp.ActionPlan != nil {
		resp.ActionPlan = append([]response.SystemHealthActionPlanStep(nil), resp.ActionPlan...)
	}
	return resp
}

func (h *handler) AnalyzePendingIssue() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.AnalyzeSystemHealthPendingIssueRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(http.StatusBadRequest, code.ParamBindError, err)
			return
		}

		resp, err := h.analyzePendingIssue(c, req)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, code.ServerError, err)
			return
		}
		c.Payload(resp)
	}
}

func (h *handler) analyzePendingIssue(ctx core.Context, req *request.AnalyzeSystemHealthPendingIssueRequest) (response.SystemHealthPendingIssueAnalyzeResp, error) {
	issue, window, err := h.resolvePendingIssue(ctx, req)
	if err != nil {
		return response.SystemHealthPendingIssueAnalyzeResp{}, err
	}
	cacheKey := systemHealthPendingIssueAnalysisCacheKey(window, issue)
	now := time.Now()
	if cached, ok := h.pendingIssueAnalysisCache.get(cacheKey, now); ok {
		return cached, nil
	}

	aiConfig := config.Get().AI
	endpoint := systemHealthPendingIssueAIEndpoint(aiConfig.BaseURL, aiConfig.WireAPI)
	result := response.SystemHealthPendingIssueAnalyzeResp{
		AIStatus: "not_configured",
		Source:   "ai-api",
		Endpoint: endpoint,
		Issue:    issue,
	}
	if !aiConfig.Enable || strings.TrimSpace(aiConfig.BaseURL) == "" || strings.TrimSpace(aiConfig.Model) == "" {
		result.Source = "not_configured"
		result.Error = "AI 未配置"
		return result, nil
	}
	wireAPI := normalizeSystemHealthAIWireAPI(aiConfig.WireAPI)
	if wireAPI == "" {
		result.AIStatus = "failed"
		result.Error = "unsupported AI wire API: " + aiConfig.WireAPI
		return result, nil
	}

	aiResp, err := callSystemHealthPendingIssueAIWithWireAPI(
		ctx,
		aiConfig.BaseURL,
		resolveSystemHealthAIAPIKey(aiConfig.APIKey, aiConfig.EnvKey),
		aiConfig.Model,
		aiConfig.TimeoutSecond,
		aiConfig.MaxTokens,
		wireAPI,
		systemHealthPendingIssueAIRequest{
			Window: window,
			Issue:  issue,
		},
	)
	if err != nil {
		result.AIStatus = "failed"
		result.Error = truncateText(err.Error(), 220)
		return result, nil
	}

	result.AIStatus = "succeeded"
	result.WorkflowRunID = aiResp.WorkflowRunID
	result.Analysis = cleanUserFacingReportText(aiResp.Analysis)
	result.Impact = cleanUserFacingReportText(aiResp.Impact)
	result.ActionPlan = sanitizePendingIssueActionPlan(aiResp.ActionPlan, issue)
	result.Validation = cleanUserFacingReportText(aiResp.Validation)
	result.Risk = cleanUserFacingReportText(aiResp.Risk)
	result.DataBoundary = cleanUserFacingReportText(aiResp.DataBoundary)
	h.pendingIssueAnalysisCache.set(cacheKey, result, now)
	return result, nil
}

func normalizeSystemHealthAIWireAPI(wireAPI string) string {
	switch strings.ToLower(strings.TrimSpace(wireAPI)) {
	case "", "responses":
		return "responses"
	case "chat", "chat_completions", "chat-completions", "chat/completions":
		return "chat_completions"
	default:
		return ""
	}
}

func systemHealthPendingIssueAIEndpoint(baseURL string, wireAPI string) string {
	if normalizeSystemHealthAIWireAPI(wireAPI) == "chat_completions" {
		return openAIChatCompletionsEndpoint(baseURL)
	}
	return openAIResponsesEndpoint(baseURL)
}

func (h *handler) resolvePendingIssue(ctx core.Context, req *request.AnalyzeSystemHealthPendingIssueRequest) (response.SystemHealthPendingIssue, string, error) {
	detail, err := h.buildSystemDetail(ctx, &request.GetSystemHealthSystemDetailRequest{
		GroupID:      req.GroupID,
		RelativeTime: req.RelativeTime,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
	})
	if err != nil {
		return response.SystemHealthPendingIssue{}, "", err
	}
	for _, issue := range buildSystemHealthPendingIssues([]response.SystemHealthSystem{detail.System}, 0) {
		if issue.ServiceName == req.ServiceName && issue.MetricKey == req.MetricKey {
			return issue, detail.Window, nil
		}
	}
	return response.SystemHealthPendingIssue{}, detail.Window, fmt.Errorf("pending issue not found or recovered: groupId=%d service=%s metric=%s", req.GroupID, req.ServiceName, req.MetricKey)
}

func sanitizePendingIssueActionPlan(steps []response.SystemHealthActionPlanStep, issue response.SystemHealthPendingIssue) []response.SystemHealthActionPlanStep {
	cleaned := make([]response.SystemHealthActionPlanStep, 0, len(pendingIssueFixedActionPhases))
	for _, step := range steps {
		action := cleanUserFacingReportText(step.Action)
		validation := cleanUserFacingReportText(step.Validation)
		risk := cleanUserFacingReportText(step.Risk)
		if action == "" && validation == "" && risk == "" {
			continue
		}
		cleaned = append(cleaned, response.SystemHealthActionPlanStep{
			Target:     cleanUserFacingReportText(step.Target),
			Action:     action,
			Validation: validation,
			Risk:       risk,
		})
		if len(cleaned) >= len(pendingIssueFixedActionPhases) {
			break
		}
	}

	result := make([]response.SystemHealthActionPlanStep, 0, len(pendingIssueFixedActionPhases))
	for index, phase := range pendingIssueFixedActionPhases {
		step := response.SystemHealthActionPlanStep{}
		if index < len(cleaned) {
			step = cleaned[index]
		}
		result = append(result, response.SystemHealthActionPlanStep{
			Phase:      phase,
			Target:     truncateText(firstNonEmpty(step.Target, issue.ServiceName, issue.SystemName, "当前服务"), pendingIssueActionTargetTextLimit),
			Action:     truncateText(firstNonEmpty(step.Action, defaultPendingIssueActionStepText(phase, issue)), pendingIssueActionTextLimit),
			Validation: truncateText(firstNonEmpty(step.Validation, defaultPendingIssueActionValidationText(phase, issue)), pendingIssueActionCheckTextLimit),
			Risk:       truncateText(firstNonEmpty(step.Risk, defaultPendingIssueActionRiskText(phase)), pendingIssueActionCheckTextLimit),
			Source:     "ai",
		})
	}
	return result
}

func defaultPendingIssueActionStepText(phase string, issue response.SystemHealthPendingIssue) string {
	metricLabel := firstNonEmpty(issue.MetricLabel, "异常指标")
	switch phase {
	case "定位":
		return fmt.Sprintf("确认 %s 当前值 %s 和异常窗口。", metricLabel, firstNonEmpty(issue.Value, "是否仍异常"))
	case "排查":
		return "对比关联指标、日志或 Trace，缩小到服务、接口或实例。"
	case "处置":
		return fmt.Sprintf("按确认原因处理后复查 %s，必要时回滚最近变更。", metricLabel)
	default:
		return "按当前异常上下文继续确认和处理。"
	}
}

func defaultPendingIssueActionValidationText(phase string, issue response.SystemHealthPendingIssue) string {
	metricLabel := firstNonEmpty(issue.MetricLabel, "指标")
	switch phase {
	case "定位":
		return "确认异常对象和时间窗口。"
	case "排查":
		return "确认主要异常来源。"
	case "处置":
		return fmt.Sprintf("%s 回到稳定区间。", metricLabel)
	default:
		return "确认问题状态清晰。"
	}
}

func defaultPendingIssueActionRiskText(phase string) string {
	switch phase {
	case "定位":
		return "对象不清会扩大排查范围。"
	case "排查":
		return "证据不足可能误判原因。"
	case "处置":
		return "未验证恢复可能复发。"
	default:
		return "处理不完整会延长影响。"
	}
}

func callSystemHealthPendingIssueAIWithWireAPI(ctx core.Context, baseURL string, apiKey string, model string, timeoutSecond int, maxTokens int, wireAPI string, payload systemHealthPendingIssueAIRequest) (systemHealthPendingIssueAIResponse, error) {
	switch normalizeSystemHealthAIWireAPI(wireAPI) {
	case "responses":
		return callSystemHealthPendingIssueAI(ctx, baseURL, apiKey, model, timeoutSecond, payload)
	case "chat_completions":
		return callSystemHealthPendingIssueAIChatCompletions(ctx, baseURL, apiKey, model, timeoutSecond, maxTokens, payload)
	default:
		return systemHealthPendingIssueAIResponse{}, fmt.Errorf("unsupported AI wire API: %s", wireAPI)
	}
}

func systemHealthPendingIssueAIInstructions() string {
	return strings.Join([]string{
		"你是系统健康中心的服务指标异常分析助手。",
		"任务：根据输入中的服务、指标状态、当前值和指标说明，给出普通 SRE 用户可执行的事件说明和处理建议。",
		"如果输入包含多个系统名，只把它们当作该服务异常出现的范围提示，不逐一分析多个系统差异。",
		"只能基于输入中的真实字段生成内容；不得编造日志、Trace、告警事件、变更记录、根因或未提供的上下文。",
		"如果输入证据不足以判断根因，应在 dataBoundary 中说明数据边界，并仍按固定三步输出偏验证和定位的 actionPlan。",
		"输出必须严格符合请求中的 json_schema。",
		"analysis 字段用于生成“事件”段落，说明该指标当前发生了什么以及为什么需要处理，控制在 40-140 个中文字符；正文不要输出“分析：”“事件：”等段落标题。",
		"impact 说明可能影响，控制在 20-100 个中文字符。",
		"actionPlan 必须且只能生成 3 个步骤，phase 固定依次为“定位”“排查”“处置”，不要使用其他步骤名。",
		"每个 action 必须可执行，控制在 60 个中文字符内；定位确认对象和异常窗口，排查缩小指标/日志/Trace/资源证据，处置说明处理和复查。",
		"validation 必须说明如何确认恢复，控制在 40 个中文字符内。",
		"risk 必须说明忽略或误判该问题的风险，控制在 40 个中文字符内。",
		"dataBoundary 必须说明本次分析只使用了哪些输入，缺少哪些下钻证据。",
	}, "\n")
}

func callSystemHealthPendingIssueAI(ctx core.Context, baseURL string, apiKey string, model string, timeoutSecond int, payload systemHealthPendingIssueAIRequest) (systemHealthPendingIssueAIResponse, error) {
	timeoutSecond = effectiveSystemHealthAITimeoutSecond(timeoutSecond)
	if strings.TrimSpace(apiKey) == "" {
		return systemHealthPendingIssueAIResponse{}, fmt.Errorf("missing AI API key")
	}

	inputPayload, err := json.Marshal(payload)
	if err != nil {
		return systemHealthPendingIssueAIResponse{}, err
	}
	requestBody, err := json.Marshal(openAIResponsesRequest{
		Model:        strings.TrimSpace(model),
		Instructions: systemHealthPendingIssueAIInstructions(),
		Input:        string(inputPayload),
		Text: openAIResponsesTextConfig{
			Format: openAIResponsesTextFormat{
				Type:   "json_schema",
				Name:   "system_health_pending_issue_analysis",
				Strict: true,
				Schema: systemHealthPendingIssueAIResponseSchema(),
			},
		},
	})
	if err != nil {
		return systemHealthPendingIssueAIResponse{}, err
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
		return systemHealthPendingIssueAIResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apiKey))

	startedAt := time.Now()
	resp, err := (&http.Client{Timeout: time.Duration(timeoutSecond) * time.Second}).Do(req)
	if err != nil {
		systemHealthWarnCtx(ctx, "system-health pending issue AI HTTP request failed",
			zap.String("endpoint", endpoint),
			zap.Bool("timeout", isTimeoutError(err)),
			zap.Duration("elapsed", time.Since(startedAt)),
			zap.Error(err),
		)
		return systemHealthPendingIssueAIResponse{}, err
	}
	defer resp.Body.Close()
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return systemHealthPendingIssueAIResponse{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return systemHealthPendingIssueAIResponse{}, fmt.Errorf("AI pending issue API returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(rawBody)))
	}

	var responsesResp openAIResponsesResponse
	if err := json.Unmarshal(rawBody, &responsesResp); err != nil {
		return systemHealthPendingIssueAIResponse{}, err
	}
	if responsesResp.Error != nil {
		return systemHealthPendingIssueAIResponse{}, fmt.Errorf("AI response error: %s", responsesResp.Error.Message)
	}
	outputText := extractOpenAIResponsesText(responsesResp)
	aiResp, err := parseSystemHealthPendingIssueAIResponseText(outputText)
	if err != nil {
		return systemHealthPendingIssueAIResponse{}, err
	}
	aiResp.WorkflowRunID = firstNonEmpty(aiResp.WorkflowRunID, responsesResp.ID)
	return aiResp, nil
}

func effectiveSystemHealthPendingIssueAIChatMaxTokens(maxTokens int) int {
	if maxTokens > 0 {
		return maxTokens
	}
	return defaultSystemHealthPendingIssueAIChatMaxTokens
}

func callSystemHealthPendingIssueAIChatCompletions(ctx core.Context, baseURL string, apiKey string, model string, timeoutSecond int, maxTokens int, payload systemHealthPendingIssueAIRequest) (systemHealthPendingIssueAIResponse, error) {
	timeoutSecond = effectiveSystemHealthAITimeoutSecond(timeoutSecond)
	if strings.TrimSpace(apiKey) == "" {
		return systemHealthPendingIssueAIResponse{}, fmt.Errorf("missing AI API key")
	}

	inputPayload, err := json.Marshal(payload)
	if err != nil {
		return systemHealthPendingIssueAIResponse{}, err
	}
	schemaBytes, err := json.Marshal(systemHealthPendingIssueAIResponseSchema())
	if err != nil {
		return systemHealthPendingIssueAIResponse{}, err
	}
	requestPayload := map[string]any{
		"model": strings.TrimSpace(model),
		"messages": []map[string]string{
			{
				"role": "system",
				"content": strings.Join([]string{
					systemHealthPendingIssueAIInstructions(),
					"Chat Completions 兼容接口必须只返回一个 JSON 对象，不要使用 Markdown 代码块。",
					"本次请求的 json_schema 如下：" + string(schemaBytes),
				}, "\n"),
			},
			{
				"role":    "user",
				"content": string(inputPayload),
			},
		},
		"response_format": map[string]string{
			"type": "json_object",
		},
		"stream": false,
	}

	requestContext := context.Background()
	if ctx != nil {
		requestContext = ctx.GetContext()
	}
	requestContext, cancel := context.WithTimeout(requestContext, time.Duration(timeoutSecond)*time.Second)
	defer cancel()

	endpoint := openAIChatCompletionsEndpoint(baseURL)
	client := &http.Client{Timeout: time.Duration(timeoutSecond) * time.Second}
	requestMaxTokens := effectiveSystemHealthPendingIssueAIChatMaxTokens(maxTokens)
	for attempt := 0; ; attempt++ {
		requestPayload["max_tokens"] = requestMaxTokens
		requestBody, err := json.Marshal(requestPayload)
		if err != nil {
			return systemHealthPendingIssueAIResponse{}, err
		}
		req, err := http.NewRequestWithContext(requestContext, http.MethodPost, endpoint, bytes.NewReader(requestBody))
		if err != nil {
			return systemHealthPendingIssueAIResponse{}, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apiKey))

		startedAt := time.Now()
		resp, err := client.Do(req)
		if err != nil {
			systemHealthWarnCtx(ctx, "system-health pending issue AI chat completions HTTP request failed",
				zap.String("endpoint", endpoint),
				zap.Bool("timeout", isTimeoutError(err)),
				zap.Duration("elapsed", time.Since(startedAt)),
				zap.Error(err),
			)
			return systemHealthPendingIssueAIResponse{}, err
		}
		rawBody, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			return systemHealthPendingIssueAIResponse{}, readErr
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return systemHealthPendingIssueAIResponse{}, fmt.Errorf("AI pending issue chat completions API returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(rawBody)))
		}

		var chatResp openAIChatCompletionsResponse
		if err := json.Unmarshal(rawBody, &chatResp); err != nil {
			return systemHealthPendingIssueAIResponse{}, err
		}
		if chatResp.Error != nil {
			return systemHealthPendingIssueAIResponse{}, fmt.Errorf("AI response error: %s", chatResp.Error.Message)
		}
		outputText := extractOpenAIChatCompletionsText(chatResp)
		aiResp, err := parseSystemHealthPendingIssueAIResponseText(outputText)
		if err == nil {
			aiResp.WorkflowRunID = firstNonEmpty(aiResp.WorkflowRunID, chatResp.ID)
			return aiResp, nil
		}
		if !shouldRetryOpenAIChatCompletionsResponse(chatResp, err) {
			return systemHealthPendingIssueAIResponse{}, err
		}
		if attempt == 1 {
			return systemHealthPendingIssueAIResponse{}, fmt.Errorf("AI pending issue chat response truncated after reaching max_tokens=%d: %w", requestMaxTokens, err)
		}
		requestMaxTokens *= 2
		if requestMaxTokens < minSystemHealthPendingIssueAIChatRetryMaxTokens {
			requestMaxTokens = minSystemHealthPendingIssueAIChatRetryMaxTokens
		}
	}
}

func shouldRetryOpenAIChatCompletionsResponse(resp openAIChatCompletionsResponse, parseErr error) bool {
	if openAIChatCompletionsResponseWasTruncated(resp) {
		return true
	}
	var syntaxErr *json.SyntaxError
	return errors.As(parseErr, &syntaxErr) && strings.Contains(strings.ToLower(syntaxErr.Error()), "unexpected end")
}

func openAIChatCompletionsResponseWasTruncated(resp openAIChatCompletionsResponse) bool {
	for _, choice := range resp.Choices {
		if strings.EqualFold(strings.TrimSpace(choice.FinishReason), "length") {
			return true
		}
	}
	return false
}

func openAIChatCompletionsEndpoint(baseURL string) string {
	endpoint := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if strings.HasSuffix(endpoint, "/chat/completions") {
		return endpoint
	}
	return endpoint + "/chat/completions"
}

func extractOpenAIChatCompletionsText(resp openAIChatCompletionsResponse) string {
	for _, choice := range resp.Choices {
		switch content := choice.Message.Content.(type) {
		case string:
			if strings.TrimSpace(content) != "" {
				return strings.TrimSpace(content)
			}
		case []any:
			parts := make([]string, 0, len(content))
			for _, item := range content {
				itemMap, ok := item.(map[string]any)
				if !ok {
					continue
				}
				if text, ok := itemMap["text"].(string); ok && strings.TrimSpace(text) != "" {
					parts = append(parts, strings.TrimSpace(text))
				}
			}
			if len(parts) > 0 {
				return strings.Join(parts, "")
			}
		case map[string]any:
			if raw, err := json.Marshal(content); err == nil {
				return string(raw)
			}
		}
	}
	return ""
}

func parseSystemHealthPendingIssueAIResponseText(text string) (systemHealthPendingIssueAIResponse, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return systemHealthPendingIssueAIResponse{}, fmt.Errorf("AI response text is empty")
	}
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)
	var aiResp systemHealthPendingIssueAIResponse
	if err := json.Unmarshal([]byte(text), &aiResp); err != nil {
		return systemHealthPendingIssueAIResponse{}, fmt.Errorf("failed to parse pending issue AI response JSON: %w", err)
	}
	return aiResp, nil
}

func systemHealthPendingIssueAIResponseSchema() map[string]any {
	actionPlanItemSchema := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"phase", "target", "action", "validation", "risk"},
		"properties": map[string]any{
			"phase":      map[string]any{"type": "string", "enum": pendingIssueFixedActionPhases},
			"target":     map[string]any{"type": "string", "minLength": 2, "maxLength": pendingIssueActionTargetTextLimit},
			"action":     map[string]any{"type": "string", "minLength": 8, "maxLength": pendingIssueActionTextLimit},
			"validation": map[string]any{"type": "string", "minLength": 6, "maxLength": pendingIssueActionCheckTextLimit},
			"risk":       map[string]any{"type": "string", "minLength": 6, "maxLength": pendingIssueActionCheckTextLimit},
		},
	}
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"analysis", "impact", "actionPlan", "validation", "risk", "dataBoundary"},
		"properties": map[string]any{
			"analysis":     map[string]any{"type": "string", "minLength": 20, "maxLength": 180},
			"impact":       map[string]any{"type": "string", "minLength": 10, "maxLength": 120},
			"actionPlan":   map[string]any{"type": "array", "minItems": 3, "maxItems": 3, "items": actionPlanItemSchema},
			"validation":   map[string]any{"type": "string", "minLength": 8, "maxLength": 100},
			"risk":         map[string]any{"type": "string", "minLength": 8, "maxLength": 100},
			"dataBoundary": map[string]any{"type": "string", "minLength": 8, "maxLength": 140},
		},
	}
}
