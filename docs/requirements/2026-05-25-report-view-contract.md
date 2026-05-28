# /report reportView 增量契约

## 目标

新版 `/report` 只消费已有报告快照和新增的 `reportView` 展示契约。已有 Dify 节点输出、已有 APO 报告字段和已有字段语义不变。

已有快照字段继续表示报告生成时的原始诊断结果：

- `reportType`
- `overview`
- `topology`
- `rootCauseAnalysis`
- `suggest`
- `analyzeProcess`
- `evidence`

新增 `reportView` 只服务新版 UI 展示。它由 Dify/Agent 增量生成，APO 后端负责保存和透传，前端负责按规则展示和过滤。后端不从长文本抽取标题、结论、业务影响、时间线或处置建议短标题。

## 页面结构

首屏不展示报告 ID、生成时间、分析窗口。

| 区块 | 展示内容 | 来源 | 缺失时行为 |
|---|---|---|---|
| 主标题 | 一句话说明发生了什么和倾向根因 | `reportView.headline` | 回退 `overview.detail` 原文 |
| 结论摘要 | 一句话说明主要依据和不确定性 | `reportView.conclusionSummary` | 回退 `overview.reason` 原文 |
| 故障位置 | 简单对象，如错误位置、慢点位置、异常对象 | `reportView.faultLocation` | 隐藏该区块 |
| 支撑指标 | 0 到 3 个根因证据指标 | `reportView.rootCauseMetricRefs` + `evidence` 过滤 | 显示“未返回根因证据指标” |
| 诊断卡片 | 告警症状、快速判断原因、已排除方向 | `reportView.summaryCards` | 回退旧字段原文 |

后续模块：

| 区块 | 展示内容 | 来源 | 缺失时行为 |
|---|---|---|---|
| 业务影响 | 有实际意义的影响摘要、入口、服务列表 | `reportView.businessImpact` | 无意义或为空时隐藏 |
| 事件时间线 / 排查步骤 | 有时间点展示时间线，无时间点展示步骤 | `reportView.timeline` / `reportView.analysisSteps` | 回退 `analyzeProcess` 原文 |
| 处置建议 | 简化标题、原始动作、验证和风险 | `reportView.actionTitles` 引用 `suggest` | 标题回退 `suggest.action` 原文，不截断造标题 |
| 关键证据 | 证据摘要和原始证据 | `reportView.evidenceHighlights` 引用 `evidence` | 原始证据仍展示，结论取 `evidence.description` |
| 排查方向 | major/minor/normal 候选与排除项 | `rootCauseAnalysis.other` | 原文展示 |

## reportView JSON

`reportView` 是新增顶层字段，不修改已有字段：

```json
{
  "reportView": {
    "version": "v1",
    "headline": "应用错误日志异常增长，更可能为日志打印或重试逻辑导致 CPU 异常",
    "conclusionSummary": "CPU 使用率与错误日志数在告警窗口同步升高，RED 仅作为影响现象，不作为根因证据。",
    "faultLocation": {
      "label": "异常位置",
      "target": "checkout-service",
      "summary": "CPU 异常集中在该服务实例。"
    },
    "summaryCards": [
      { "type": "symptom", "title": "告警症状", "summary": "错误日志数在告警窗口内异常增长。" },
      { "type": "rootCause", "title": "快速判断原因", "summary": "CPU 与错误日志同步升高，倾向日志打印或重试逻辑。" },
      { "type": "excluded", "title": "已排除方向", "summary": "未看到下游 5xx 或网络超时同步升高。" }
    ],
    "rootCauseMetricRefs": [
      { "evidenceIndex": 0, "title": "CPU 使用率", "summary": "45% -> 96%" },
      { "evidenceIndex": 1, "title": "错误日志数", "summary": "30 -> 55377" }
    ],
    "businessImpact": null,
    "timeline": [
      { "timestamp": 1779811334000000, "title": "错误日志数开始升高", "summary": "同一错误模板持续出现。" }
    ],
    "analysisSteps": [],
    "actionTitles": [
      { "suggestIndex": 0, "title": "先降低异常日志输出" }
    ],
    "evidenceHighlights": [
      { "evidenceIndex": 0, "title": "CPU 使用率持续升高", "summary": "告警窗口内 CPU 峰值接近 96%。" }
    ]
  }
}
```

## 支撑指标规则

首屏右上角“支撑指标”只展示根因证据指标。

1. 主根因方向取 `rootCauseAnalysis.other[level=major].type`。
2. 优先展示 `evidence[type=metrics]`，并且必须包含真实 `chartData`。
3. 指标的 `name`、`description`、`legend`、`unit` 至少一个能匹配主根因方向。
4. RED 指标只允许在 `downstream`、`dependency`、`traffic`、`thirdparty` 下作为证据。
5. `cpu`、`code`、`log`、`retry`、`net`、`storage`、`runtime`、`infra` 下不允许用 RED 兜底。
6. 没有匹配指标时不展示指标卡，显示“未返回根因证据指标”。

## Dify 输出要求

Dify 只新增输出，不修改已有节点和已有输出结构。

- 新增 `reportView` 生成节点或新增最终聚合输出。
- `headline`、`conclusionSummary`、`summaryCards`、`faultLocation`、`actionTitles`、`evidenceHighlights` 必须由 agent 写成展示文案。
- `rootCauseMetricRefs` 必须引用真实 `evidence` 下标；不能引用没有 `chartData` 的证据。
- `rootCauseMetricRefs` 的指标名称、描述、legend、unit 要匹配主根因方向；没有匹配根因的指标时宁可不输出，不要用无关 RED 指标误导。
- RED 只允许在 `downstream`、`dependency`、`traffic`、`thirdparty` 方向作为根因证据，`cpu`、`code`、`log`、`retry`、`net`、`storage`、`runtime`、`infra` 方向不能用 RED 兜底。
- `faultLocation` 只写简单易懂的异常对象和一句依据，不输出复杂链路。
- `businessImpact` 没有明确业务意义时输出 `null` 或省略。
- `timeline[]` 使用 `{ "timestamp": number, "title": string, "summary": string }`，`timestamp` 是真实微秒时间戳；没有明确时间点时不要编造，改写入 `analysisSteps[]`。
- `analysisSteps[]` 使用 `{ "title": string, "summary": string }`，不输出 `timestamp`。
- `actionTitles[].title` 必须是简短动作标题，不要由 APO 前端截断。
- `evidence.description` 仍然写原始证据结论，供原始证据区域展示。

## APO 实现边界

APO 后端：

- 接收 `AddAlertEventReportRequest.reportView`。
- 不改 ClickHouse 表结构；写入时把 `reportView` 序列化到内部隐藏 tag。
- 读取时从内部隐藏 tag 还原为 `snapshot.reportView`。
- `overview.tags` 对外过滤内部隐藏字段。
- 不生成结论型 `headline`、`businessImpact`、`timeline`、`action title`。

APO 前端：

- 优先消费 `snapshot.reportView`。
- 按支撑指标规则从 `evidence` 中筛选根因证据指标。
- `timeline[].timestamp` 由前端转成本地展示时间，`summary` 映射为详情文本；兼容旧的 `time/detail` 字段。
- 保留旧报告字段的直接展示兜底，但不硬抽取、不硬拼结论。
- 业务影响无意义时隐藏。
