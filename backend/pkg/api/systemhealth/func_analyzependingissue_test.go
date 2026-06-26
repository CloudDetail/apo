// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package systemhealth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func TestCallSystemHealthPendingIssueAIUsesMetricContext(t *testing.T) {
	var gotRequest openAIResponsesRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/responses" {
			t.Fatalf("expected /responses path, got %s", r.URL.Path)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-key" {
			t.Fatalf("expected bearer auth, got %q", auth)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotRequest); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		output := `{
			"analysis":"order-api 的错误率已超过异常阈值，当前应优先确认失败请求是否集中在核心接口。",
			"impact":"请求失败会直接影响订单系统核心链路。",
			"actionPlan":[
				{"phase":"定位失败","target":"order-api","action":"按接口和状态码聚合最近窗口失败请求，确认错误是否集中在单一接口。","validation":"确认错误率回落且失败请求不再集中增长。","risk":"若只看平均值可能遗漏局部接口失败。"},
				{"phase":"验证恢复","target":"订单系统","action":"处理后回看错误率和活跃告警指标，确认系统健康分同步恢复。","validation":"观察连续两个窗口错误率保持稳定。","risk":"短时恢复后仍可能再次波动。"}
			],
			"validation":"确认错误率和活跃告警指标回到稳定区间。",
			"risk":"当前分析只基于系统健康指标，未包含日志和 Trace 下钻证据。",
			"dataBoundary":"仅基于当前待处理指标上下文生成。"
		}`
		_ = json.NewEncoder(w).Encode(openAIResponsesResponse{
			ID:         "resp-pending-1",
			OutputText: output,
		})
	}))
	defer server.Close()

	payload := systemHealthPendingIssueAIRequest{
		Window: "15m",
		Issue: response.SystemHealthPendingIssue{
			GroupID:      101,
			SystemName:   "订单系统",
			ServiceName:  "order-api",
			MetricKey:    healthMetricErrorRate,
			MetricLabel:  "错误率",
			Status:       healthStatusRed,
			Value:        "5.70%",
			Note:         "错误率超过异常阈值",
			ServiceScore: 35,
			SystemScore:  38,
		},
	}

	aiResp, err := callSystemHealthPendingIssueAI(nil, server.URL, "test-key", "gpt-test", 3, payload)
	if err != nil {
		t.Fatalf("call pending issue AI: %v", err)
	}

	if gotRequest.Text.Format.Name != "system_health_pending_issue_analysis" {
		t.Fatalf("expected pending issue schema, got %q", gotRequest.Text.Format.Name)
	}
	if !strings.Contains(gotRequest.Instructions, "服务指标异常") {
		t.Fatalf("expected service metric issue instructions, got %q", gotRequest.Instructions)
	}
	if !strings.Contains(gotRequest.Instructions, "不逐一分析多个系统差异") {
		t.Fatalf("expected instructions to avoid per-system analysis, got %q", gotRequest.Instructions)
	}
	if !strings.Contains(gotRequest.Input, "订单系统") || !strings.Contains(gotRequest.Input, "5.70%") {
		t.Fatalf("expected metric context in AI input, got %s", gotRequest.Input)
	}
	if aiResp.WorkflowRunID != "resp-pending-1" {
		t.Fatalf("expected workflow id from response id, got %q", aiResp.WorkflowRunID)
	}
	if aiResp.Analysis == "" || len(aiResp.ActionPlan) != 2 || aiResp.DataBoundary == "" {
		t.Fatalf("expected structured pending issue AI output, got %+v", aiResp)
	}
}

func TestCallSystemHealthPendingIssueAIWithChatCompletionsWireAPIUsesMessageContent(t *testing.T) {
	var gotRequest map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("expected /chat/completions path, got %s", r.URL.Path)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-key" {
			t.Fatalf("expected bearer auth, got %q", auth)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotRequest); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		responseFormat, _ := gotRequest["response_format"].(map[string]any)
		if responseFormat["type"] != "json_object" {
			t.Fatalf("expected json_object response format, got %+v", responseFormat)
		}
		messages, _ := gotRequest["messages"].([]any)
		if len(messages) != 2 {
			t.Fatalf("expected system and user messages, got %+v", gotRequest["messages"])
		}
		userMessage, _ := messages[1].(map[string]any)
		userContent, _ := userMessage["content"].(string)
		if !strings.Contains(userContent, "order-api") || !strings.Contains(userContent, "5.70%") {
			t.Fatalf("expected metric context in chat completion input, got %s", userContent)
		}
		maxTokens, ok := gotRequest["max_tokens"].(float64)
		if !ok || int(maxTokens) != defaultSystemHealthPendingIssueAIChatMaxTokens {
			t.Fatalf("expected default max_tokens %d, got %#v", defaultSystemHealthPendingIssueAIChatMaxTokens, gotRequest["max_tokens"])
		}

		output := `{
			"analysis":"order-api has a sustained error-rate pending issue that needs operator triage.",
			"impact":"Users may see failed order requests until the error rate returns to baseline.",
			"actionPlan":[
				{"phase":"Locate","target":"order-api","action":"Group recent failed requests by endpoint and status code to find the concentrated failure path.","validation":"The failing endpoint and status code group is identified.","risk":"Broad aggregation can hide a narrow endpoint spike."},
				{"phase":"Verify","target":"order-api","action":"After mitigation, compare the error-rate metric with the previous analysis window.","validation":"Error rate stays below the alert threshold for two windows.","risk":"Short recovery windows can miss recurring failures."}
			],
			"validation":"Confirm the metric is below threshold for consecutive windows.",
			"risk":"Ignoring the issue can prolong user-visible request failures.",
			"dataBoundary":"Only pending issue metric context was provided; logs and traces were not provided."
		}`
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": "chatcmpl-pending-1",
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"role":    "assistant",
						"content": output,
					},
				},
			},
		})
	}))
	defer server.Close()

	payload := systemHealthPendingIssueAIRequest{
		Window: "15m",
		Issue: response.SystemHealthPendingIssue{
			GroupID:      101,
			SystemName:   "order-system",
			ServiceName:  "order-api",
			MetricKey:    healthMetricErrorRate,
			MetricLabel:  "error rate",
			Status:       healthStatusRed,
			Value:        "5.70%",
			Note:         "error rate exceeds threshold",
			ServiceScore: 35,
			SystemScore:  38,
		},
	}

	aiResp, err := callSystemHealthPendingIssueAIWithWireAPI(nil, server.URL, "test-key", "glm-4-flash", 3, 0, "chat_completions", payload)
	if err != nil {
		t.Fatalf("call pending issue AI with chat completions: %v", err)
	}

	if aiResp.WorkflowRunID != "chatcmpl-pending-1" {
		t.Fatalf("expected workflow id from chat completion id, got %q", aiResp.WorkflowRunID)
	}
	if aiResp.Analysis == "" || len(aiResp.ActionPlan) != 2 || aiResp.DataBoundary == "" {
		t.Fatalf("expected structured pending issue AI output, got %+v", aiResp)
	}
}

func TestSystemHealthPendingIssueAIWireAPIsShareBusinessInstructions(t *testing.T) {
	var responsesInstructions string
	var chatSystemPrompt string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		output := `{
			"analysis":"order-api has a pending issue that needs immediate triage.",
			"impact":"Users may see failed order requests until the issue recovers.",
			"actionPlan":[
				{"phase":"Locate","target":"order-api","action":"Group recent failures by endpoint and status code to identify the affected path.","validation":"The affected endpoint and status code are identified.","risk":"Wide grouping can hide a narrow failure spike."},
				{"phase":"Verify","target":"order-api","action":"Compare the metric after mitigation with the previous analysis window.","validation":"The metric stays below threshold for two windows.","risk":"Short recovery windows may miss recurring failures."}
			],
			"validation":"Confirm the metric returns below threshold for consecutive windows.",
			"risk":"Ignoring the issue can prolong user-visible request failures.",
			"dataBoundary":"Only pending issue metric context was provided; logs and traces were not provided."
		}`

		switch r.URL.Path {
		case "/responses":
			var req openAIResponsesRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("decode responses request: %v", err)
			}
			responsesInstructions = req.Instructions
			_ = json.NewEncoder(w).Encode(openAIResponsesResponse{
				ID:         "resp-prompt",
				OutputText: output,
			})
		case "/chat/completions":
			var req map[string]any
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("decode chat completions request: %v", err)
			}
			messages, _ := req["messages"].([]any)
			if len(messages) > 0 {
				systemMessage, _ := messages[0].(map[string]any)
				chatSystemPrompt, _ = systemMessage["content"].(string)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": "chat-prompt",
				"choices": []map[string]any{
					{
						"message": map[string]any{
							"role":    "assistant",
							"content": output,
						},
					},
				},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	payload := systemHealthPendingIssueAIRequest{
		Window: "15m",
		Issue: response.SystemHealthPendingIssue{
			GroupID:     101,
			ServiceName: "order-api",
			MetricKey:   healthMetricErrorRate,
			Status:      healthStatusRed,
			Value:       "5.70%",
			Note:        "error rate exceeds threshold",
		},
	}

	if _, err := callSystemHealthPendingIssueAI(nil, server.URL, "test-key", "gpt-test", 3, payload); err != nil {
		t.Fatalf("call responses pending issue AI: %v", err)
	}
	if _, err := callSystemHealthPendingIssueAIWithWireAPI(nil, server.URL, "test-key", "glm-test", 3, 0, "chat_completions", payload); err != nil {
		t.Fatalf("call chat completions pending issue AI: %v", err)
	}

	responsesInstructions = strings.TrimSpace(responsesInstructions)
	chatSystemPrompt = strings.TrimSpace(chatSystemPrompt)
	if responsesInstructions == "" {
		t.Fatal("expected responses instructions to be captured")
	}
	if !strings.Contains(chatSystemPrompt, responsesInstructions) {
		t.Fatalf("expected chat completions system prompt to include shared responses instructions\nresponses: %q\nchat: %q", responsesInstructions, chatSystemPrompt)
	}
}

func TestSystemHealthPendingIssueAIResponseSchemaRequiresAllObjectProperties(t *testing.T) {
	assertStrictObjectSchema(t, systemHealthPendingIssueAIResponseSchema(), "root")
}

func TestSystemHealthPendingIssueAnalysisCacheReturnsCachedSuccessfulResult(t *testing.T) {
	cache := newSystemHealthPendingIssueAnalysisCache(time.Minute, 8)
	now := time.Date(2026, 6, 22, 10, 0, 0, 0, time.UTC)
	issue := response.SystemHealthPendingIssue{
		GroupID:     101,
		ServiceName: "order-api",
		MetricKey:   healthMetricErrorRate,
		Status:      healthStatusRed,
		Value:       "5.70%",
		Note:        "错误率超过异常阈值",
	}
	key := systemHealthPendingIssueAnalysisCacheKey("15m", issue)
	expected := response.SystemHealthPendingIssueAnalyzeResp{
		AIStatus: "succeeded",
		Issue:    issue,
		Analysis: "第一次 AI 分析结果",
	}

	cache.set(key, expected, now)
	cached, ok := cache.get(key, now.Add(30*time.Second))

	if !ok {
		t.Fatal("expected pending issue analysis cache hit")
	}
	if cached.Analysis != expected.Analysis || cached.Issue.Value != expected.Issue.Value {
		t.Fatalf("unexpected cached analysis: %+v", cached)
	}
}

func TestSystemHealthPendingIssueAnalysisCacheSkipsFailedResult(t *testing.T) {
	cache := newSystemHealthPendingIssueAnalysisCache(time.Minute, 8)
	now := time.Date(2026, 6, 22, 10, 0, 0, 0, time.UTC)
	issue := response.SystemHealthPendingIssue{
		GroupID:     101,
		ServiceName: "order-api",
		MetricKey:   healthMetricErrorRate,
		Status:      healthStatusRed,
		Value:       "5.70%",
	}
	key := systemHealthPendingIssueAnalysisCacheKey("15m", issue)

	cache.set(key, response.SystemHealthPendingIssueAnalyzeResp{
		AIStatus: "failed",
		Issue:    issue,
		Error:    "AI timeout",
	}, now)

	if _, ok := cache.get(key, now.Add(30*time.Second)); ok {
		t.Fatal("expected failed pending issue analysis to bypass cache")
	}
}

func TestSystemHealthPendingIssueAnalysisCacheKeyIncludesWindowAndMetricValue(t *testing.T) {
	issue := response.SystemHealthPendingIssue{
		GroupID:     101,
		ServiceName: "order-api",
		MetricKey:   healthMetricErrorRate,
		Status:      healthStatusRed,
		Value:       "5.70%",
		Note:        "错误率超过异常阈值",
	}

	baseKey := systemHealthPendingIssueAnalysisCacheKey("15m", issue)
	changedWindowKey := systemHealthPendingIssueAnalysisCacheKey("30m", issue)
	issue.Value = "2.00%"
	changedValueKey := systemHealthPendingIssueAnalysisCacheKey("15m", issue)

	if baseKey == changedWindowKey {
		t.Fatal("expected cache key to include analysis window")
	}
	if baseKey == changedValueKey {
		t.Fatal("expected cache key to include metric value")
	}
}

func assertStrictObjectSchema(t *testing.T, schema map[string]any, path string) {
	t.Helper()

	if schema["type"] == "object" {
		properties, _ := schema["properties"].(map[string]any)
		if len(properties) > 0 {
			requiredValues, ok := schema["required"].([]string)
			if !ok {
				t.Fatalf("%s schema required must be []string", path)
			}
			required := make(map[string]struct{}, len(requiredValues))
			for _, name := range requiredValues {
				required[name] = struct{}{}
			}

			missing := make([]string, 0)
			for name := range properties {
				if _, ok := required[name]; !ok {
					missing = append(missing, name)
				}
			}
			sort.Strings(missing)
			if len(missing) > 0 {
				t.Fatalf("%s schema required is missing properties: %s", path, strings.Join(missing, ","))
			}
		}
	}

	properties, _ := schema["properties"].(map[string]any)
	for name, raw := range properties {
		child, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		assertStrictObjectSchema(t, child, path+"."+name)
		if items, ok := child["items"].(map[string]any); ok {
			assertStrictObjectSchema(t, items, path+"."+name+"[]")
		}
	}
}
