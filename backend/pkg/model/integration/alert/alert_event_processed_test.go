package alert

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAttachReportOutputContractToAlertAnalyzeWorkflowParams(t *testing.T) {
	params := AlertAnalyzeWorkflowParams{
		AlertName: "checkout error rate high",
	}

	AttachReportOutputContract(&params)

	payload, err := json.Marshal(params)
	if err != nil {
		t.Fatal(err)
	}

	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatal(err)
	}

	contract, ok := raw["reportOutputContract"].(map[string]any)
	if !ok {
		t.Fatalf("reportOutputContract missing from workflow params: %s", payload)
	}

	overviewDetail, _ := contract["overview.detail"].(string)
	if !strings.Contains(overviewDetail, "一句话") ||
		!strings.Contains(overviewDetail, "异常症状") ||
		!strings.Contains(overviewDetail, "标题由工作流生成") ||
		!strings.Contains(overviewDetail, "证据不足") ||
		!strings.Contains(overviewDetail, "初步定位") ||
		strings.Contains(overviewDetail, "必须包含简单异常症状和根因结论") {
		t.Fatalf("overview.detail contract is not explicit enough: %q", overviewDetail)
	}

	overviewSummary, _ := contract["overview.summary"].(string)
	if !strings.Contains(overviewSummary, "不持久化") ||
		!strings.Contains(overviewSummary, "overview.detail") {
		t.Fatalf("overview.summary contract does not explain the persisted field: %q", overviewSummary)
	}

	overviewReason, _ := contract["overview.reason"].(string)
	if !strings.Contains(overviewReason, "主要依据") ||
		!strings.Contains(overviewReason, "已排除") {
		t.Fatalf("overview.reason contract is not explicit enough: %q", overviewReason)
	}

	rootCauseOther, _ := contract["rootCauseAnalysis.other"].(string)
	if !strings.Contains(rootCauseOther, "normal") ||
		!strings.Contains(rootCauseOther, "已排除") {
		t.Fatalf("rootCauseAnalysis.other contract is not explicit enough: %q", rootCauseOther)
	}

	suggest, _ := contract["suggest"].(string)
	if !strings.Contains(suggest, "verification") ||
		!strings.Contains(suggest, "不要编造") ||
		!strings.Contains(suggest, "summary") ||
		!strings.Contains(suggest, "短标题") ||
		strings.Contains(suggest, "首屏处置建议摘要卡") ||
		!strings.Contains(suggest, "verificationSummary") ||
		!strings.Contains(suggest, "不要用省略号") {
		t.Fatalf("suggest contract is not explicit enough: %q", suggest)
	}

	evidence, _ := contract["evidence"].(string)
	if !strings.Contains(evidence, "major") ||
		!strings.Contains(evidence, "根因") ||
		!strings.Contains(evidence, "CPU") ||
		!strings.Contains(evidence, "RED") ||
		!strings.Contains(evidence, "不能用 RED 兜底") ||
		!strings.Contains(evidence, "不要只输出") {
		t.Fatalf("evidence contract is not explicit enough: %q", evidence)
	}

	reportView, _ := contract["reportView"].(string)
	if !strings.Contains(reportView, "增量输出") ||
		!strings.Contains(reportView, "不要修改已有") ||
		!strings.Contains(reportView, "rootCauseEvidenceRefs") ||
		!strings.Contains(reportView, "首屏支撑证据唯一来源") ||
		!strings.Contains(reportView, "metrics/red/log/traces") ||
		!strings.Contains(reportView, "type 别名归一") ||
		!strings.Contains(reportView, "disk 按 storage") ||
		!strings.Contains(reportView, "file 按 runtime/storage") ||
		!strings.Contains(reportView, "memory 按 runtime/infra") ||
		!strings.Contains(reportView, "system 按 infra/runtime/cpu/memory") ||
		!strings.Contains(reportView, "config 按 config/dependency/log") ||
		!strings.Contains(reportView, "self 按 code/log/runtime") ||
		!strings.Contains(reportView, "rootCauseMetricRefs") ||
		!strings.Contains(reportView, "兼容") ||
		!strings.Contains(reportView, "真实 evidence 下标") ||
		!strings.Contains(reportView, "下游服务") ||
		!strings.Contains(reportView, "接口耗时") ||
		!strings.Contains(reportView, "MySQL") ||
		!strings.Contains(reportView, "锁等待") ||
		!strings.Contains(reportView, "连接池等待") ||
		!strings.Contains(reportView, "配置拉取失败") ||
		!strings.Contains(reportView, "DB 连接失败") ||
		!strings.Contains(reportView, "OOMKill") ||
		!strings.Contains(reportView, "文件描述符") ||
		!strings.Contains(reportView, "Read timed out") ||
		!strings.Contains(reportView, "evidenceHighlights 只服务关键证据摘要") ||
		!strings.Contains(reportView, "正常或已排除证据") ||
		!strings.Contains(reportView, "businessImpact") ||
		!strings.Contains(reportView, "无意义") ||
		!strings.Contains(reportView, "timeline 使用 timestamp/title/summary") ||
		!strings.Contains(reportView, "analysisSteps 使用 title/summary") ||
		!strings.Contains(reportView, "没有匹配根因的支撑证据") ||
		!strings.Contains(reportView, "faultLocation") ||
		!strings.Contains(reportView, "actionTitles") ||
		!strings.Contains(reportView, "简短标题") {
		t.Fatalf("reportView contract is not explicit enough: %q", reportView)
	}

	firstScreenConstraint, _ := contract["firstScreenConstraint"].(string)
	if !strings.Contains(firstScreenConstraint, "三个诊断卡片") ||
		!strings.Contains(firstScreenConstraint, "处置建议在独立面板展示") ||
		strings.Contains(firstScreenConstraint, "四个卡片") {
		t.Fatalf("firstScreenConstraint still describes the old action card layout: %q", firstScreenConstraint)
	}
}
