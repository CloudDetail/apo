// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"time"
)

// AlertEvent With Workflow record
type AEventWithWRecord struct {
	AlertEvent

	LastStatus string `json:"lastStatus,omitempty" ch:"last_status"`

	WorkflowDetail
	NotifyDetail
}

// FIXME add importance field
type WorkflowDetail struct {
	WorkflowRunID string `json:"workflowRunId" ch:"workflow_run_id"`

	WorkflowID   string `json:"workflowId" ch:"workflow_id"`
	WorkflowName string `json:"workflowName" ch:"workflow_name"`

	Validity    string    `json:"validity" ch:"validity"`
	Input       string    `json:"input" ch:"input"`
	Output      string    `json:"output" ch:"output"`
	RoundedTime time.Time `json:"-" ch:"rounded_time"`
	Importance  uint8     `json:"-" ch:"importance"`
	LastCheckAt time.Time `json:"lastCheckAt" ch:"last_check_at"`

	Duration string `json:"duration" ch:"-"`

	WorkflowParams WorkflowParams `json:"workflowParams"`

	// Deprecated: use [Validity] instead, will remove after 1.7.x
	IsValid string `json:"isValid" ch:"is_valid"`

	AlertDirection string `json:"alertDirection" ch:"alert_direction"`
	AnalyzeRunID   string `json:"analyzeRunId" ch:"analyze_run_id"`
	AnalyzeErr     string `json:"analyzeErr" ch:"analyze_err"`
}

type NotifyDetail struct {
	SendSuccess string    `json:"notifySuccess" ch:"notify_success"`
	SendFailed  string    `json:"notifyFailed" ch:"notify_failed"`
	NotifyAt    time.Time `json:"notifyAt" ch:"notify_at"`
}

type WorkflowParams struct {
	StartTime int64 `json:"startTime" form:"startTime"`
	EndTime   int64 `json:"endTime" form:"endTime"`

	NodeName string `json:"nodeName" form:"nodeName"`
	NodeIp   string `json:"nodeIp" form:"nodeIp"`

	Params  string `json:"params"`
	Edition string `json:"edition"`
}

type AlertAnalyzeWorkflowParams struct {
	Node      string `json:"node,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Pod       string `json:"pod,omitempty"`
	Service   string `json:"service,omitempty"`
	Endpoint  string `json:"endpoint,omitempty"`
	Pid       string `json:"pid,omitempty"`
	AlertName string `json:"alertName,omitempty"`

	Detail      string `json:"detail,omitempty"`
	ContainerID string `json:"containerId,omitempty"`

	Tags         map[string]string `json:"tags,omitempty"`
	RawTags      map[string]any    `json:"raw_tags,omitempty"`
	AlertEventId string            `json:"alertEventId,omitempty"`

	ReportOutputContract *AlertReportOutputContract `json:"reportOutputContract,omitempty"`
}

type AlertReportOutputContract struct {
	OverviewSummary       string `json:"overview.summary"`
	OverviewDetail        string `json:"overview.detail"`
	OverviewReason        string `json:"overview.reason"`
	RootCauseCause        string `json:"rootCauseAnalysis.cause"`
	RootCauseOther        string `json:"rootCauseAnalysis.other"`
	Suggest               string `json:"suggest"`
	Evidence              string `json:"evidence"`
	ReportView            string `json:"reportView"`
	AnalyzeProcess        string `json:"analyzeProcess"`
	FirstScreenConstraint string `json:"firstScreenConstraint"`
}

func AttachReportOutputContract(params *AlertAnalyzeWorkflowParams) {
	if params == nil {
		return
	}
	params.ReportOutputContract = &AlertReportOutputContract{
		OverviewSummary:       "当前 APO 写入接口不持久化 overview.summary；首屏标题内容必须写入 overview.detail。",
		OverviewDetail:        "首屏大标题来源，标题由工作流生成，前端不会拼接 rootCause。只输出一句话，必须包含简单异常症状；证据不足时使用初步定位、倾向于、更可能等表述，禁止把宽泛方向写成定案根因；不要写证据、排除过程、处置建议、Markdown 标题或列表。",
		OverviewReason:        "首屏标题下方描述来源。按当前现象、支持判断、当前结论三段概括主要依据、已排除方向和待确认项；不要把下游、网络、资源等宽泛方向写成已完全确认。",
		RootCauseCause:        "根因结论。表达为故障对象/方向 + 异常现象 + 导致的业务影响；证据不足时保持初步定位语气，避免只写网络异常、资源异常等宽泛候选方向。",
		RootCauseOther:        "排查方向列表。major 表示快速判断原因，minor 表示待验证方向，normal 表示已排除方向；normal.reason 只写简单排除原因。",
		Suggest:               "处置建议列表。summary 写处置建议短标题，用于列表扫读，必须基于 action 重新改写而不是截断；action 写完整可执行动作；verification、trigger、risk 有依据才写，没有验证方式时不要编造；verificationSummary 写验证短文案；summary 和 verificationSummary 不要用省略号、Markdown 或列表。",
		Evidence:              "支撑指标/关键证据来源。必须优先输出与 rootCauseAnalysis.other 中 major 根因方向直接相关的 metrics/red/log/traces 证据；例如 CPU/代码/日志打印/重试导致的 CPU 方向异常，应输出 CPU 使用率、错误日志数、重试次数等根因证据指标；downstream/dependency 根因应输出下游服务、下游接口耗时、错误率、超时、慢调用分布、下游数据库/MySQL/SQL 查询耗时、锁等待、连接池等待、下游调用异常日志或慢 trace 等证据。不要只输出通用服务 RED 指标；RED 只允许在 downstream/dependency/traffic/thirdparty 方向作为根因证据，cpu/code/log/retry/net/storage/runtime/infra 方向不能用 RED 兜底。metrics 证据需包含 name、type=metrics、unit、description、data[].chart.chartData；如果根因只能由日志或 trace 证明，例如 ResourceAccessException、SocketTimeoutException、Read timed out，必须输出真实 log/traces 证据；没有真实 chartData 时不要编造。",
		ReportView:            "新版 /report 展示增量输出。不要修改已有节点、已有输出字段或已有字段结构。输出 headline、conclusionSummary、summaryCards、faultLocation、rootCauseEvidenceRefs、rootCauseMetricRefs、businessImpact、timeline、analysisSteps、actionTitles、evidenceHighlights。timeline 使用 timestamp/title/summary 结构：timestamp 必须是真实微秒时间戳，title 写短标题，summary 写一句话事件说明；有明确时间点才输出 timeline，禁止把 timestamp 改成 time 或把 summary 改成 detail。analysisSteps 使用 title/summary 结构；没有明确时间点时 timeline 输出空数组，把无时间点排查内容写入 analysisSteps，不要编造 timestamp。rootCauseEvidenceRefs 是首屏支撑证据唯一来源，必须引用真实 evidence 下标，支持 metrics/red/log/traces；引用 metrics/red 时该 evidence 必须有真实 chartData，指标名称/描述/legend/unit 要匹配 major 根因方向；引用 log/traces 时 title/summary 必须直接描述 major 根因证据。type 别名归一：disk 按 storage，file 按 runtime/storage，memory 按 runtime/infra，system 按 infra/runtime/cpu/memory，config 按 config/dependency/log，self 按 code/log/runtime。各类型首选支撑证据：storage/disk 使用 DB 查询耗时、慢 SQL、磁盘 IO、磁盘使用率、连接池等待、SQL 错误；runtime/file/memory 使用 GC、堆内存、线程、协程、FD、文件描述符、OOM；infra/system 使用 Pod 重启、OOMKill、节点资源、容器状态、K8s Event、CPU、内存；config 使用配置拉取失败、配置中心错误、DB 连接失败、连接串/凭据错误、初始化失败日志；self 使用应用异常、错误日志、panic、stack、runtime 异常。downstream/dependency 根因必须优先引用下游服务、具体接口耗时、错误率、超时、慢调用分布、下游数据库/MySQL/SQL 查询耗时、锁等待、连接池等待、下游调用异常日志或慢 trace 等证据，例如 ts-travel-plan-service 的 POST /api/v1/travelplanservice/travelPlan/cheapest 耗时，ts-config-service 访问 train-ticket-mysql-master-0 时 SQL 查询变慢、锁等待或连接池等待升高，或 requesttemplate-demo 调用 localhost:58082 /api/jpa-demo/sleep 出现 ResourceAccessException、SocketTimeoutException、Read timed out。rootCauseMetricRefs 仅作为兼容字段，有根因指标时可同步填入同一 metrics/red evidence；不能替代 rootCauseEvidenceRefs。evidenceHighlights 只服务关键证据摘要，不是首屏支撑证据来源；可以摘要根因证据或排除证据，但正常或已排除证据必须明确写成排除依据，不能替代 rootCauseEvidenceRefs，不能伪装成支撑指标。没有匹配根因的支撑证据时宁可不输出 rootCauseEvidenceRefs/rootCauseMetricRefs，不要用无关 RED 指标或正常日志误导。faultLocation 只写简单易懂的异常对象；businessImpact 没有明确业务意义或无意义时输出 null 或省略；actionTitles.title 必须是 agent 改写的简短标题，不要让 APO 截断生成。",
		AnalyzeProcess:        "分析过程原文字段。结构化展示时间线请写入 reportView.timeline.timestamp；无时间点时输出短步骤，不要复述首屏结论，不要为了展示伪造 HH:MM。",
		FirstScreenConstraint: "首屏信息分工：overview.detail 只讲症状+诊断判断语气，overview.reason 只讲归因依据+排除过程，三个诊断卡片再展示具体症状、快速判断原因和已排除项，处置建议在独立面板展示，避免重复。",
	}
}
