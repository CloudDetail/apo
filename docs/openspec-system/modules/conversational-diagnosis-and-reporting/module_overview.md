# Conversational Diagnosis And Reporting Module Overview

## Module Goal
本模块负责把故障诊断结果以“对话”和“报告”两种方式呈现给用户。前端提供 `/alert-analyze/chatAI/:service/:endpoint`、`/cause/report/:traceId/:spanId`、`/cause/error-report/:traceId`、`/report` 和 `/agent` 页面；本仓库后端实际承接的接口主要是 `/api/custom/*` 自定义数据视图模板管理，以及 `/api/alerts/events/report/get|add` 报告快照读写。

## Scope
当前代码证据支持以下范围：

- ChatAI 页面拉起流式故障分析，选择可疑节点后继续请求附加数据视图。
- Cause report 页面读取 `originx` 返回的慢调用/错误根因链路，并在前端组装拓扑与步骤说明。
- Report 页面按 `alertEventId` 或 `agentChatId` 读取已生成报告快照。
- Agent 页面通过 AG-UI 协议连接外部 SRE Agent 服务，支持快捷问题、线程历史和 findings/tool display 渲染。
- 自定义数据视图模板的创建、读取、更新、删除。
- 告警事件报告快照在 ClickHouse `alert_report` 中的读取与新增。

当前代码不应纳入本模块主体范围的能力：

- `automated-diagnosis-and-workflows`：告警事件分析流程、impact、异常贡献度、workflow iframe。
- `dataplane-and-topology-governance`：`/api/dataplane/*`、数据平面接入、服务匹配规则。
- `service-observability` / `trace-exploration-and-profiling` / `logs-exploration-and-parsing`：观测原始查询页面。
- `platform-administration`：license 页面本身属于平台配置，但 license 状态会横切影响本模块入口。

## Frontend Entry

| Entry Type | Evidence | Notes |
|---|---|---|
| Routes | `frontend/src/pro/routes.js` | `/cause/report/:traceId/:spanId`、`/cause/error-report/:traceId`、`/alert-analyze/chatAI/:service/:endpoint`、`/report`、`/agent`。 |
| Chat AI Page | `frontend/src/pro/views/chatAI/index.tsx` | `ChatAIProvider` 包裹的页面入口。 |
| Cause Report Page | `frontend/src/pro/views/causeReport/index.tsx` | `CauseReportProvider` 包裹的页面入口。 |
| Report Page | `frontend/src/pro/views/report/index.tsx` | 报告快照展示页。 |
| Agent Page | `frontend/src/pro/views/agent/index.tsx` | AG-UI 对话入口。 |
| Shared Agent UI | `frontend/src/pro/components/AGUIChat` | Agent 会话与 findings 展示主体。 |
| State Context | `ChatAIContext.tsx` / `CauseReportContext.tsx` / `AGUIContext.tsx` | 页面级状态核心。 |
| API Wrappers | `frontend/src/pro/api/*.ts` | 连接本仓库后端与外部服务边界。 |

## Backend Entry

| Entry Type | Evidence | Notes |
|---|---|---|
| Router Group | `backend/pkg/router/router_api.go` | `/api/custom/*` 与 `/api/alerts/events/report/*`。 |
| Custom Handler | `backend/pkg/api/custom` | 自定义数据视图 CRUD。 |
| Custom Service | `backend/pkg/services/custom` | 模板持久化逻辑。 |
| Report Handler | `backend/pkg/api/alerts/func_getalerteventreport.go`, `func_addalerteventreport.go` | 报告快照读写。 |
| Report Service | `backend/pkg/services/alerts/service_alert_event_report.go`, `service_add_alert_event_report.go` | 直接操作 ClickHouse `alert_report`。 |
| Repository | `backend/pkg/repository/database/dao_custom_query_type.go`, `backend/pkg/repository/clickhouse/dao_alert_event_report.go` | 模块内主要持久化边界。 |
| External Boundary | `/ai/api/*`, `/originx/api/*` | 当前前端强依赖，但主 Go 后端未实现。 |

## Core Concepts

| Concept | Explanation |
|---|---|
| `custom_query_service` | 自定义数据视图模板表，记录 `customName` 与逗号拼接后的 `querytype`。 |
| `alert_report` | 已生成报告快照存储表，供 `/report` 页面消费。 |
| `ChatAIProvider` | 管理当前入口服务、可疑节点列表、对话内容、流式请求控制器和分析阶段状态。 |
| `CauseReportProvider` | 负责把 `originx` 返回的 trace/cause 数据转换为报告页可消费的 `topologyData` 与 `reportData`。 |
| `AGUIContext` | 对 `@ag-ui/client` 进行适配，保存消息、run 状态、findings 工具调用、默认模型和线程历史。 |
| `agentReady` / `dataPlaneExist` | Agent 页面在前端通过 `/api/dataplane/data/check` 得到的前置状态。 |
| `reportType` / `reportData` | `/api/alerts/events/report/get` 返回的快照类型与报告正文。 |

## Related Files

- `frontend/src/pro/routes.js`
- `frontend/src/pro/views/chatAI/index.tsx`
- `frontend/src/pro/views/chatAI/ChatAICom/index.tsx`
- `frontend/src/pro/views/chatAI/ChatAICom/CauseConfirm/DataViewConfirmStep/index.tsx`
- `frontend/src/pro/views/chatAI/ChatAICom/CauseConfirm/DataViewConfirmStep/CustomDataViewTab.tsx`
- `frontend/src/pro/views/causeReport/index.tsx`
- `frontend/src/pro/views/causeReport/CauseReportCom/index.tsx`
- `frontend/src/pro/views/report/index.tsx`
- `frontend/src/pro/views/agent/index.tsx`
- `frontend/src/pro/components/AGUIChat/index.tsx`
- `frontend/src/pro/components/AGUIChat/components/QuickQuestions.tsx`
- `frontend/src/pro/components/AGUIChat/components/AgentModel/index.tsx`
- `frontend/src/pro/contexts/ChatAIContext.tsx`
- `frontend/src/pro/contexts/CauseReportContext.tsx`
- `frontend/src/pro/contexts/AGUIContext.tsx`
- `frontend/src/pro/api/chatAI.ts`
- `frontend/src/pro/api/conversation.ts`
- `frontend/src/pro/api/cause.ts`
- `frontend/src/pro/api/report.ts`
- `frontend/src/pro/api/custom.ts`
- `frontend/src/pro/api/modelSetting.ts`
- `backend/pkg/api/custom/handler.go`
- `backend/pkg/api/custom/func_getcustomquery.go`
- `backend/pkg/api/custom/func_createcustomquery.go`
- `backend/pkg/api/custom/func_updatecustomquery.go`
- `backend/pkg/api/custom/func_deletecustomquery.go`
- `backend/pkg/services/custom/service.go`
- `backend/pkg/services/custom/service_get_custom.go`
- `backend/pkg/services/custom/service_creat_custom.go`
- `backend/pkg/services/custom/service_update_custom.go`
- `backend/pkg/services/custom/service_delete_custom.go`
- `backend/pkg/api/alerts/func_getalerteventreport.go`
- `backend/pkg/api/alerts/func_addalerteventreport.go`
- `backend/pkg/services/alerts/service_alert_event_report.go`
- `backend/pkg/services/alerts/service_add_alert_event_report.go`
- `backend/pkg/repository/database/dao_custom_query_type.go`
- `backend/pkg/repository/clickhouse/dao_alert_event_report.go`

## Dependencies

| Dependency Type | Dependency | Why It Matters |
|---|---|---|
| Frontend Shared | `AuthRouter` | 本模块路由都经统一包装，但实际 route permission 检查当前被注释。 |
| Frontend Shared | `timeRangeReducer` | ChatAI、CauseReport、Report 都依赖统一时间范围。 |
| Frontend Shared | `UserContext` + `AppSidebar` | `/agent` 等入口最终由菜单配置和 edition 过滤控制。 |
| Cross Module | `dataplane-and-topology-governance` | Agent 页通过 `/api/dataplane/data/check` 判断是否允许继续使用。 |
| Cross Module | `automated-diagnosis-and-workflows` | `/alert-analyze/info` 页面中的故障报告弹窗会跳转进入本模块路由。 |
| Backend Shared | `request.js` | 对 `/api/*`、`/ai/api/*`、`/originx/api/*` 注入认证和语言头。 |
| External Service | `/ai/api/*` | ChatAI、Agent、模型配置、历史线程的主体后端。 |
| External Service | `/originx/api/*` | 根因报告与 trace 相关原始分析数据来源。 |

## Risks And Unknowns

1. `/ai/api/*` 与 `/originx/api/*` 在主 Go 后端中没有对应 router 注册，属于外部边界，语义需要持续确认。
2. `chatAI.ts` 中 `/api/chat` 与 `/api/alertchat` wrapper 当前没有被主页面实际使用，存在旧接口残留嫌疑。
3. `frontend/vite.config.mjs` 仅代理 `/ai/api`，未见 `/originx/api` 开发代理，开发态路由链路需要确认。
4. `PreUseInspection.tsx` 仍使用 `/#/agent-datasource` 老路径，而当前主路由是 `/dataplane`。
5. `/report` 页面同时接受 `alertEventId` 与 `agentChatId`，但当前仓库里 `agentChatId` 的完整产生链路未完全呈现。

