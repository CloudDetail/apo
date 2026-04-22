# Conversational Diagnosis And Reporting Frontend

## Pages

| Page | Route | Responsibility |
|---|---|---|
| `ChatAIPage` | `/alert-analyze/chatAI/:service/:endpoint` | 发起流式 AI 诊断、展示对话、确认可疑节点、补充数据视图。 |
| `CauseReportPage` | `/cause/report/:traceId/:spanId` | 展示慢调用报告。 |
| `CauseReportPage` | `/cause/error-report/:traceId` | 展示错误报告。 |
| `Report` | `/report` | 读取已存储的报告快照并折叠展示摘要、分析过程和证据。 |
| `AgentPage` | `/agent` | 提供通用 SRE Agent 会话入口。 |

## Routes

| Route | File | Notes |
|---|---|---|
| `/alert-analyze/chatAI/:service/:endpoint` | `frontend/src/pro/routes.js` | `showDataGroup: 'view'`。 |
| `/cause/report/:traceId/:spanId` | `frontend/src/pro/routes.js` | `hideSystemTimeRangePicker: true`。 |
| `/cause/error-report/:traceId` | `frontend/src/pro/routes.js` | `hideSystemTimeRangePicker: true`。 |
| `/report` | `frontend/src/pro/routes.js` | `hideSystemTimeRangePicker: true`。 |
| `/agent` | `frontend/src/pro/routes.js` | `hideSystemTimeRangePicker: true`。 |

## Menus

| Entry | Evidence | Notes |
|---|---|---|
| `agent` | `frontend/src/core/components/AppSidebar.js` | 仅在 `VITE_APP_CODE_VERSION === 'EE'` 时加入顶部特殊菜单区。 |
| `/report` feature route | `backend/pkg/repository/database/init_feature_mapping.go` | 映射到“故障报告” feature，但未看到独立左侧 menu key。 |
| `causeReport` / `chatAI` deep link | `frontend/src/pro/routes.js` | 当前更像从其他页面跳入的业务深链接，而非独立主菜单。 |

## Key Components

| Component | File | Role |
|---|---|---|
| `ChatAICom` | `frontend/src/pro/views/chatAI/ChatAICom/index.tsx` | 管理初始分析流、追问流、对话拼装和自动滚动。 |
| `CauseConfirm` | `frontend/src/pro/views/chatAI/ChatAICom/CauseConfirm/index.tsx` | 二步式确认根因节点与可视化数据类型。 |
| `DataViewConfirmStep` | `frontend/src/pro/views/chatAI/ChatAICom/CauseConfirm/DataViewConfirmStep/index.tsx` | 读取、删除、选择自定义数据视图模板。 |
| `CustomDataViewTab` | `frontend/src/pro/views/chatAI/ChatAICom/CauseConfirm/DataViewConfirmStep/CustomDataViewTab.tsx` | 创建或更新自定义模板后立即调用 `analysisNode`。 |
| `CauseReportCom` | `frontend/src/pro/views/causeReport/CauseReportCom/index.tsx` | 报告壳层。 |
| `ReportHeader` / `ReportContent` | `frontend/src/pro/views/causeReport/CauseReportCom/*` | 负责头部概览和步骤内容渲染。 |
| `AGUIChat` | `frontend/src/pro/components/AGUIChat/index.tsx` | Agent 对话入口。 |
| `QuickQuestions` | `frontend/src/pro/components/AGUIChat/components/QuickQuestions.tsx` | 根据权限状态展示快捷提问、安装引导和模型配置入口。 |
| `AgentModel` | `frontend/src/pro/components/AGUIChat/components/AgentModel/index.tsx` | 展示并更新默认模型。 |

## State Management

| State Area | File | Notes |
|---|---|---|
| Chat AI page state | `ChatAIContext.tsx` | 保存 `service`、`endpoint`、`causeNode`、`suspectedNodeList`、`dialogue`、`streamType` 等。 |
| Cause report state | `CauseReportContext.tsx` | 保存 `traceId`、`spanId`、`category`、`topologyData`、`reportData`、`isTimeOut`。 |
| Agent session state | `AGUIContext.tsx` | 保存 `messages`、`currentThreadId`、`isRunning`、`findingsToolCalls`、`defaultModel`。 |
| Global time range | `timeRangeReducer` | `ChatAIContext`、`Report` 和 `CauseReportContext` 均直接读取。 |
| License status | `LicenseContext.tsx` | 布局层会显示 license banner，但不是本模块业务状态。 |

## API Usage

| Frontend Action | API | Backend/Boundary |
|---|---|---|
| ChatAI 初始分析 | `streamFetch('/ai/api/alertchat')` | 外部 AI 服务 |
| ChatAI 追问 | `streamFetch('/ai/api/chat?...')` | 外部 AI 服务 |
| ChatAI 节点二次分析 | `streamFetch('/ai/api/nodeinfo')` | 外部 AI 服务 |
| 对话附件下载 | `/ai/api/download` | 外部 AI 服务 |
| Agent 历史线程加载 | `/ai/api/v1/conversations/messages` | 外部 AI 服务 |
| Agent 模型列表 | `/ai/api/v1/model-setting/list` | 外部 AI 服务 |
| Agent 默认模型切换 | `/ai/api/v1/model-setting/default` | 外部 AI 服务 |
| CauseReport 慢调用拓扑 | `/originx/api/query/mutatedtraceswithspan/:traceId/:spanId` | 外部 `originx` 服务 |
| CauseReport 慢调用流程 | `/originx/api/flows/span/run/:traceId/:spanId` | 外部 `originx` 服务 |
| CauseReport 错误拓扑 | `/originx/api/query/errortraces/:traceId` | 外部 `originx` 服务 |
| CauseReport 错误流程 | `/originx/api/flows/error/:traceId/:spanId` | 外部 `originx` 服务 |
| 报告节点错误日志 | `/originx/api/query/logs/:traceId/:spanId` | 外部 `originx` 服务 |
| 自定义视图列表 | `/api/custom/get` | 本仓库后端 |
| 新增/更新/删除自定义视图 | `/api/custom/create`, `/api/custom/update`, `/api/custom/delete` | 本仓库后端 |
| 读取报告快照 | `/api/alerts/events/report/get` | 本仓库后端 |
| Agent 前置检测 | `/api/dataplane/data/check` | `dataplane-and-topology-governance` 模块 |

## User Interaction Flow

1. 用户从 `/alert-analyze/info` 或其他入口跳到 `/alert-analyze/chatAI/:service/:endpoint`。
2. `ChatAIContext` 读取当前时间范围和路由参数，`ChatAICom` 发起 `/ai/api/alertchat` 流式请求。
3. AI 返回 `nodeList` 后，前端展开 `CauseConfirm`，让用户确认可疑节点与数据视图类型。
4. 用户如果选择自定义数据视图，会先通过 `/api/custom/*` 管理模板，再触发 `analysisNode` 请求 `/ai/api/nodeinfo`。
5. 如果用户在 `/agent` 进入通用会话，`AGUIContext` 先通过 `/api/dataplane/data/check`、`/ai/api/v1/model-setting/list` 做前置检查，再发起 AG-UI 事件流。
6. 当已有报告快照时，`/report` 通过 `/api/alerts/events/report/get` 渲染摘要、根因、建议与证据。
7. 当用户查看慢调用或错误根因时，`CauseReportContext` 直接读取 `originx` 数据并在前端重建报告结构。

## Permissions

| Area | Evidence | Interpretation |
|---|---|---|
| Route wrapper | `AuthRouter` | 所有页面都被包装，但实际 `checkRoutePermission` 的 `useEffect` 被注释。 |
| Agent menu visibility | `AppSidebar.js` | `agent` 仅在企业版菜单白名单中出现。 |
| Agent usability | `AGUIContext.checkPermission` | 依赖 `/api/dataplane/data/check` 返回的 `agentReady` 和 `dataPlaneExist`。 |
| Report deep link | `init_feature_mapping.go` | `/report` 映射到“故障报告” feature。 |

## Notes

1. `chatAI.ts` 中 `/api/chat` 与 `/api/alertchat` wrapper 目前没有替代主页面里的 `streamFetch` 逻辑。
2. `getFindingsDataApi` 已定义，但当前主证据仍集中在 AG-UI 事件流和消息历史，单独 findings API 的实际消费范围待确认。
3. `PreUseInspection.tsx` 仍保留旧路径 `/#/agent-datasource`，与当前 `/dataplane` 路由不一致。

## Findings Consumer Notes

- `/agent` 页面消费 findings node 时，应优先使用 canonical fields：`title`、`status`、`summaryConclusion`、`content`、`metrics`、`table`、`logs`、`traces`、`unsupportedReason`；其中详情面板文本区只渲染一个 canonical 文本块，优先 `summaryConclusion`，缺失时回退到 `content`。
- rollout 期间即使 AgentSRE 继续返回历史 alias，前端也不应把这些 alias 当作主字段语义。
- `unsupported` 节点在当前批次中保持 clickable，但 consumer UI 可以把它的视觉样式降到与 `normal` 等价，避免分散用户对异常数据的注意力；这种视觉策略不改变 `nodeKind: unsupported` 的语义。
- 详情面板继续遵守单一 canonical 文本槽位规则；对于 `unsupported` 节点，如果 producer 没有提供 `summaryConclusion` / `content`，consumer 不强制插入 unavailable banner。
- consumer adapter 必须保留 producer 最终写回的状态语义；如果 AgentSRE 已明确返回 `unsupported + normal`，前端不得再强行改成 `warning`。
- 当前已落地的 canonical `unsupportedReason` 本地化映射包括：`no-downstream-services`、`red-metrics-unavailable`、`red-query-failed`、`no-valid-evidence`。
- 如果树节点与右侧 evidence 表达矛盾，应优先视为 producer projection 问题，而不是前端自行补造结论。
