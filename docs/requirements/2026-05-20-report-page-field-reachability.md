# /report 页面字段可达性说明

## 目标

本文档对应新版 `/report` 事故处置简报页面，用来说明页面每个信息字段是否真实可达、应从哪里取得、需要怎样实现。页面本身只展示用户决策所需信息，不展示本文档中的“可达性”“实现路径”等研发说明。

## 可达性分级

| 分级 | 含义 |
|---|---|
| 已可用 | 当前 APO 已有接口或 `alert_report` 快照能够返回，前端可直接消费或做轻量适配。 |
| 后端可补 | 当前 `/api/alerts/events/report/get` 未直接返回，但 APO 后端已有数据源或相邻接口，能通过聚合接口补齐。 |
| 生产侧补充 | 需要报告生产器、AgentSRE 或 OriginX 在生成报告时输出更结构化的字段。 |
| 暂不展示 | 数据虽然可能推断，但不稳定或解释成本高，不适合放到当前页面。 |

## 页面字段矩阵

| 页面位置 | 字段 | 当前 mock 示例 | 可达性 | 数据来源 | 实现方式 |
|---|---|---:|---|---|---|
| 顶部状态 | 报告 ID | `INC-2026-0520-014` | 已可用 | `alert_report.overview_tags.alertEventId` 或查询 tags | 当前 `/api/alerts/events/report/get` 已按 tags 查快照，可直接映射。 |
| 顶部状态 | 生成时间 | `2026-05-20 14:24` | 已可用 | `alert_report.overview_timestamp` | 当前返回 `overview.timestamp`，前端格式化即可。 |
| 顶部状态 | 分析窗口 | `2026-05-20 14:03 - 2026-05-20 14:24` | 已可用 / 生产侧补充 | `alert_report.evidence[].chartData`、报告生产器 timeRange | 短期可从 evidence chartData 的最早/最晚时间推断；正式应由报告生产器输出明确 `analysisWindow`。 |
| 顶部状态 | 故障级别 | `主要故障` | 后端可补 | 告警事件 severity、reportType、影响面和阈值 | 短期可由 `reportType=slow/error/infra` 粗略映射；正式应从告警事件 severity 或报告生产器输出。 |
| 顶部状态 | 当前状态 | `仍异常` | 后端可补 | 告警事件状态 + 当前窗口 RED 指标 | 按 `alertEventId` 查询事件是否恢复，再对入口服务/endpoint 做当前 RED 查询确认。 |
| 顶部状态 | 当前状态更新时间 | `2026-05-20 14:29` | 后端可补 | 告警事件更新时间、当前 RED 查询窗口结束时间 | 该字段属于实时状态，不改写报告正文，只更新“当前告警状态”区域。 |
| 顶部状态 | 持续时长 | `18 分钟` | 后端可补 | 告警开始/恢复时间、异常 span 窗口 | 用事件 start/end 或 anomaly span 的最早/最晚时间计算；未恢复时用当前时间。 |
| 顶部状态 | 根因置信度 | `82%` | 生产侧补充 | AgentSRE evidence scoring、报告生产器评分 | 推荐由 AgentSRE 输出结构化 score；后端可按证据类型、时序一致性、排除项数量做兜底评分。 |
| 第一屏标题 | 已定位结论 | `order-service 订单查询慢 SQL 导致 checkout 下单链路延迟升高` | 已可用 | `alert_report.rootcause` | 当前 `rootCauseAnalysis.cause` 可直接展示，标题表达确定性结论，不写“可能/疑似”。 |
| 第一屏摘要 | 影响与建议摘要 | `checkout 与 submit-order 入口 P95...` | 已可用 | `overview.detail`、`overview.reason`、Agent 总结 | 当前快照已有 overview 文本；如果生产器分字段输出，前端不再拼接长文本。 |
| 第一屏右侧 | 当前告警状态 | `当前仍异常，未确认恢复` | 后端可补 | 告警事件状态 + 当前窗口 RED 指标 | 该区域只展示可变状态，不展示处置动作，避免与“立即行动”重复。 |
| 第一屏右侧 | 当前状态摘要 | `最近 5 分钟 checkout P95 仍高于 500ms...` | 后端可补 | 当前 RED 查询、告警事件状态 | 用于说明现在是否仍在异常/恢复/已恢复；不回写报告快照内容。 |
| 报告时影响 | 影响范围 | `支付链路 checkout、submit-order...` | 已可用 | `overview.detail` | 当前快照已提供；后续可由影响面接口生成更稳定结构。 |
| 报告时影响 | 受影响入口 | `checkout`、`submit-order` | 后端可补 | `/api/alerts/event/impact`、RED endpoint 查询 | 按报告分析窗口查询影响入口，再用 RED 趋势确认是否真实异常。 |
| 报告时影响 | 关键指标变化 | `P95 320ms -> 1.8s`、`错误率 0.6% -> 4.2%` | 已可用 / 后端可补 | `alert_report.evidence.chartData`、`/api/dataplane/redcharts`、`/api/service/redcharts` | 快照含 chartData 时直接转换；缺失时按 service、endpoint、报告分析窗口查询 RED。 |
| 影响链路 | 拓扑节点与异常边 | `checkout-service -> order-service -> mysql-orders` | 已可用 / 后端可补 | `alert_report.topology`、OriginX trace/flow、dataplane topology | 快照有 topology 时直接展示；没有时按 traceId/spanId 或 service 时间窗生成。 |
| 事件时间线 | 异常开始、触发、定位、建议生成 | `14:03...14:21...` | 已可用 / 生产侧补充 | `analyzeProcess`、告警事件、AgentSRE tool timeline | 当前可解析 `analyzeProcess` 文本；正式应由生产器输出结构化 timeline。 |
| 行动详情 | 分类动作、验证指标、风险 | 回滚、扩容、索引、观察 | 已可用 / 生产侧补充 | `suggest`、AgentSRE runbook 输出 | 当前 action 可用；category、verification、risk 应由报告生产器结构化输出。 |
| 证据详情 | 指标/日志/trace 证据摘要 | 指标变化、慢 SQL、拓扑 | 已可用 / 后端可补 | `alert_report.evidence`、`/api/log/fault/*`、`/api/service/trace/logs`、`/originx/api/*` | 快照展示摘要；详情下钻需要后端按 traceId/spanId/service/timeRange 补查。 |
| 已排除项 | 非根因说明 | `CPU 峰值 61%，未达到瓶颈` | 已可用 | `rootCauseAnalysis.other[level=normal]` | 取 normal candidate 作为已排除原因。 |

## 不放在页面上的内容

| 内容 | 原因 | 推荐位置 |
|---|---|---|
| 字段可达性、来源说明、实现路径 | 用户处置故障时不需要理解研发链路，会降低决策效率。 | 本文档、接口设计文档、PR 描述。 |
| 每个证据的内部查询 SQL/PQL | 过于底层，且不同 provider 实现不同。 | 证据详情弹窗或调试模式。 |
| 置信度算法细节 | 算法解释成本高，且需要稳定口径。 | AgentSRE/报告生产器 contract。 |
| 所有指标曲线 | 第一屏会过载。 | 证据详情或下钻页面。 |
| 独立的“为什么这么判断/关键证据”模块 | 会和首屏根因结论、下方根因证据重复；当前没有稳定结构化推理字段支撑独立展示。 | 根因证据详情内部，未来可做 tab。 |

## 后续接口建议

建议新增或扩展一个聚合后的报告视图接口，避免前端在页面上拼多条业务接口：

```ts
interface IncidentReportView {
  snapshot: AlertEventReportResponse
  analysisWindow?: {
    startTime: number
    endTime: number
  }
  status?: 'active' | 'recovering' | 'resolved' | 'unknown'
  statusUpdatedAt?: number
  currentStatusSummary?: string
  duration?: string
  severity?: 'major' | 'minor' | 'normal'
  confidence?: number
  affectedEntrypoints?: string[]
  metricDeltas?: Array<{
    label: string
    before: number
    current: number
    unit: string
    target?: string
  }>
  timeline?: Array<{
    time: string
    title: string
    detail: string
  }>
}
```

短期可以继续使用 mock，但 mock 字段应只选本文档标为“已可用”“后端可补”“生产侧补充”的字段，不使用没有明确来源的装饰性信息。
