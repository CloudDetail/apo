# Conversational Diagnosis And Reporting Flows

## Main Flow 1
### Chat-Based Root Cause Analysis

1. 用户进入 `/alert-analyze/chatAI/:service/:endpoint`。
2. `ChatAIContext` 从路由参数和全局时间范围生成分析上下文。
3. `ChatAICom` 通过 `streamFetch('/ai/api/alertchat')` 请求初始分析流。
4. 前端持续解析 SSE 文本块，把 `content`、`reportUrlList` 和 `nodeList` 写入对话状态。
5. 当流中出现 `nodeList` 时，`changeSuspectedNodeList` 会标记根节点并展开 `CauseConfirm`。
6. 用户在 `DataViewConfirmStep` 中选择默认或自定义数据视图。
7. 若使用自定义模板，则先调用 `/api/custom/create|update`；随后 `analysisNode` 调用 `streamFetch('/ai/api/nodeinfo')` 做二次分析。

## Main Flow 2
### Report Consumption

1. 用户进入 `/report?startTime=...&endTime=...&alertEventId=...` 或携带 `agentChatId`。
2. 页面调用 `/api/alerts/events/report/get`。
3. 后端 `alerts` service 从 `alert_report` 表按标签读取最近一条记录。
4. 前端根据 `reportType` 和 `reportData` 渲染 `Summary`、`AffectedEndpoints`、`RootCauseAnalysis`、`RecommendedAction`、`AnalysisProcess`、`SupportingEvidence`。

## Edge Cases

| Case | Current Behavior |
|---|---|
| Chat 流返回 JSON 而不是 SSE | `ChatAICom.getChartData` 会尝试 `response.json()`，若 `code === 'B0302'` 则标记为未配置。 |
| 报告接口无数据 | `/report` 页面最终显示 `Empty`。 |
| Cause report 数据缺失 | `CauseReportContext` 把 `dataMissing` 置为 `true`，页面显示空态文案。 |
| Cause report 超时 | 若 `originx` 返回 `status === 'failure'` 且消息包含 `timeout`，页面显示 `reportTimeout`。 |
| Agent 无默认模型 | `QuickQuestions` 显示模型配置警告并引导打开 `AgentModel`。 |

## Failure Paths

| Failure Path | Evidence |
|---|---|
| Agent 未就绪 | `AGUIContext.checkPermission` 把 `hasPermission` 设为 `false`，`QuickQuestions` 展示安装引导。 |
| 数据平面不存在 | `AGUIContext.checkPermission` 允许继续使用，但展示 warning。 |
| `/originx/api/*` 查询失败 | `CauseReportContext` 捕获错误，仅打印日志并结束 loading。 |
| 自定义模板操作失败 | `CustomDataViewTab` 当前仅 `console.log(error)`，失败反馈较弱。 |
| 下载附件失败 | `HumanMessage` 没有显式错误提示。 |

## Permission Constraints

1. 页面路由都经 `AuthRouter` 包装，但当前 `checkRoutePermission` 的实际执行被注释，因此不能把它写成已生效的严格前端拦截。
2. `/agent` 页面虽然可路由进入，但真正可操作还取决于 `/api/dataplane/data/check`、默认模型是否已配置，以及外部 Agent 服务是否在线。
3. `/api/custom/*` 与 `/api/alerts/events/report/*` 当前 router 只观察到 `LicMiddleware()`，未见单独 `AuthMiddleware()` 包裹这两个子组的直接定义，需要结合上层 group 继续确认。

