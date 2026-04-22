# Conversational Diagnosis And Reporting Source Map

## Scope
本文档只固定 `conversational-diagnosis-and-reporting` 模块在当前仓库中的真实代码证据。该模块覆盖会话式诊断界面、故障报告展示、告警事件报告消费、自定义数据视图配置，以及与外部 AI 会话服务、`originx` 根因分析服务之间的前端对接边界。

## Relevant Source Directories

| Layer | Path | Notes |
|---|---|---|
| Frontend | `frontend/src/pro/views/chatAI` | `/alert-analyze/chatAI/:service/:endpoint` 页面，负责流式 AI 诊断、可疑节点确认和数据视图选择。 |
| Frontend | `frontend/src/pro/views/causeReport` | `/cause/report/:traceId/:spanId` 与 `/cause/error-report/:traceId` 页面，负责展示 `originx` 返回的根因报告与拓扑。 |
| Frontend | `frontend/src/pro/views/report` | `/report` 页面，负责消费已有的告警事件/Agent 报告快照。 |
| Frontend | `frontend/src/pro/views/agent` | `/agent` 页面，基于 AG-UI 客户端展示通用对话式诊断入口。 |
| Frontend | `frontend/src/pro/components/AGUIChat` | Agent 会话 UI、消息渲染、Findings 展示、模型选择和快捷问题入口。 |
| Frontend | `frontend/src/pro/contexts/ChatAIContext.tsx` | ChatAI 页面状态、流式请求控制、可疑节点与数据视图选择状态。 |
| Frontend | `frontend/src/pro/contexts/CauseReportContext.tsx` | Cause report 页面状态、`originx` 数据获取与拓扑预处理。 |
| Frontend | `frontend/src/pro/contexts/AGUIContext.tsx` | Agent 会话、线程历史、模型列表、数据平面就绪检查和 AG-UI 事件流处理。 |
| Frontend | `frontend/src/pro/api/chatAI.ts` | `/api/chat`、`/api/alertchat` 和 `/ai/api/download` wrapper，当前前两个 wrapper 未见页面调用。 |
| Frontend | `frontend/src/pro/api/conversation.ts` | `/ai/api/v1/conversations/*` 历史消息与 findings 数据 wrapper。 |
| Frontend | `frontend/src/pro/api/cause.ts` | `/originx/api/*` wrapper。 |
| Frontend | `frontend/src/pro/api/report.ts` | `/api/alerts/events/report/get` wrapper。 |
| Frontend | `frontend/src/pro/api/custom.ts` | `/api/custom/*` wrapper。 |
| Frontend | `frontend/src/pro/api/modelSetting.ts` | `/ai/api/v1/model-setting/*` wrapper。 |
| Backend | `backend/pkg/api/custom` | 本仓库内与会话诊断直接对应的后端 API 包，负责自定义数据视图模板。 |
| Backend | `backend/pkg/services/custom` | 自定义数据视图 CRUD service。 |
| Backend | `backend/pkg/repository/database/dao_custom_query_type.go` | `custom_query_service` 持久化。 |
| Backend | `backend/pkg/api/alerts/func_getalerteventreport.go` | `/api/alerts/events/report/get`，供 `/report` 页面读取报告。 |
| Backend | `backend/pkg/api/alerts/func_addalerteventreport.go` | `/api/alerts/events/report/add`，供其他链路写入报告快照。 |
| Backend | `backend/pkg/services/alerts/service_alert_event_report.go` | 从 ClickHouse `alert_report` 读取报告。 |
| Backend | `backend/pkg/services/alerts/service_add_alert_event_report.go` | 向 `alert_report` 写入报告。 |
| Backend | `backend/pkg/repository/clickhouse/dao_alert_event_report.go` | `alert_report` 查询与插入。 |
| External Boundary | `/ai/api/*` | 当前仓库前端大量调用，但主 Go 后端未提供对应实现。 |
| External Boundary | `/originx/api/*` | 当前仓库前端调用的根因/trace 查询边界，主 Go 后端未提供实现。 |

## Route Files

| Side | File | Observed Routes |
|---|---|---|
| Frontend | `frontend/src/pro/routes.js` | `/cause/report/:traceId/:spanId`, `/cause/error-report/:traceId`, `/alert-analyze/chatAI/:service/:endpoint`, `/report`, `/agent` |
| Frontend Shared | `frontend/src/routes.js` | `/user` 等基础路由与本模块无直接页面注册关系，但 `AuthRouter` 在这里被统一复用。 |
| Backend | `backend/pkg/router/router_api.go` | `/api/custom/create`, `/api/custom/update`, `/api/custom/get`, `/api/custom/delete`, `/api/alerts/events/report/get`, `/api/alerts/events/report/add` |
| External Boundary | `frontend/vite.config.mjs` | 开发态只显式代理 `/ai/api` 到 `http://127.0.0.1:8000`；`/originx/api` 未见同文件代理配置。 |

## Page Files

| Type | File | Notes |
|---|---|---|
| Route Page | `frontend/src/pro/views/chatAI/index.tsx` | 仅包裹 `ChatAIProvider` 与 `ChatAICom`。 |
| Chat Container | `frontend/src/pro/views/chatAI/ChatAICom/index.tsx` | 负责 `/ai/api/alertchat` 初始流式分析、`/ai/api/chat` 追问和 `CauseConfirm` 展开。 |
| Conversation | `frontend/src/pro/views/chatAI/ChatAICom/Conversation/ConversationCom.tsx` | 按人类/AI消息渲染对话块。 |
| File Download | `frontend/src/pro/views/chatAI/ChatAICom/Conversation/HumanMessage.tsx` | 通过 `/ai/api/download` 下载对话返回的附件。 |
| Cause Confirm | `frontend/src/pro/views/chatAI/ChatAICom/CauseConfirm/index.tsx` | 二步式确认可疑节点和数据视图。 |
| Custom View Config | `frontend/src/pro/views/chatAI/ChatAICom/CauseConfirm/DataViewConfirmStep/CustomDataViewTab.tsx` | 调用 `/api/custom/*` 保存自定义数据视图模板。 |
| Cause Report Route | `frontend/src/pro/views/causeReport/index.tsx` | 仅包裹 `CauseReportProvider` 与 `CauseReportCom`。 |
| Cause Report Shell | `frontend/src/pro/views/causeReport/CauseReportCom/index.tsx` | 根据 `category` 展示报告头、拓扑与故障分析过程。 |
| Cause Report Header | `frontend/src/pro/views/causeReport/CauseReportCom/ReportHeader/index.tsx` | 展示 traceId、当前节点、受影响入口和结论。 |
| Cause Report Content | `frontend/src/pro/views/causeReport/CauseReportCom/ReportContent/index.tsx` | 渲染 `reportData.reports`。 |
| Report Page | `frontend/src/pro/views/report/index.tsx` | 根据 URL 中 `alertEventId` 或 `agentChatId` 请求报告快照。 |
| Agent Page | `frontend/src/pro/views/agent/index.tsx` | 用 `AGUIProvider` 驱动 `AGUIChat`。 |
| AGUI Chat | `frontend/src/pro/components/AGUIChat/index.tsx` | 管理 threadId 历史加载、清空、发送问题。 |
| Quick Entry | `frontend/src/pro/components/AGUIChat/components/QuickQuestions.tsx` | 根据数据平面状态和默认模型显示引导、快捷问题和设置入口。 |
| Model Setting UI | `frontend/src/pro/components/AGUIChat/components/AgentModel/index.tsx` | 读取并切换默认模型，弹出 provider 配置卡片。 |

## API Wrapper Files

| File | Wrapper | API Path | Notes |
|---|---|---|---|
| `frontend/src/pro/api/chatAI.ts` | `alertchatApi` | `/api/alertchat` | wrapper 存在，当前主页面改用 `streamFetch('/ai/api/alertchat')`。 |
| `frontend/src/pro/api/chatAI.ts` | `chatApi` | `/api/chat` | wrapper 存在，当前主页面改用 `streamFetch('/ai/api/chat?...')`。 |
| `frontend/src/pro/api/chatAI.ts` | `downloadFileApi` | `/ai/api/download` | `HumanMessage.tsx` 实际使用。 |
| `frontend/src/pro/api/conversation.ts` | `getConversationMessagesApi` | `/ai/api/v1/conversations/messages` | `AGUIContext.loadHistoryMessages` 实际使用。 |
| `frontend/src/pro/api/conversation.ts` | `getFindingsDataApi` | `/ai/api/v1/conversations/findings-data` | wrapper 存在，当前主要证据仍在 context/组件侧预留。 |
| `frontend/src/pro/api/cause.ts` | `getMutatedtracesApi` | `/originx/api/query/mutatedtraceswithspan/:traceId/:spanId` | `CauseReportContext.fetchSlowData` 使用。 |
| `frontend/src/pro/api/cause.ts` | `getFlowsApi` | `/originx/api/flows/span/run/:traceId/:spanId` | `CauseReportContext.fetchSlowData` 使用。 |
| `frontend/src/pro/api/cause.ts` | `getErrorTraceLogApi` | `/originx/api/query/logs/:traceId/:spanId` | 报告节点详情的错误日志使用。 |
| `frontend/src/pro/api/cause.ts` | `getErrortracesApi` | `/originx/api/query/errortraces/:traceId` | `CauseReportContext.getErrorTraces` 使用。 |
| `frontend/src/pro/api/cause.ts` | `getErrorFlowsApi` | `/originx/api/flows/error/:traceId/:spanId` | 错误报告流程使用。 |
| `frontend/src/pro/api/cause.ts` | `getEntryEndpointApi` | `/originx/api/query/entryendpoint` | `serviceAnalyze` 中的 FaultReportModal 使用，是本模块的上游入口桥接。 |
| `frontend/src/pro/api/report.ts` | `getReportApi` | `api/alerts/events/report/get` | `/report` 页面读取报告快照。 |
| `frontend/src/pro/api/custom.ts` | `getCustomList` | `/api/custom/get` | 读取自定义数据视图模板列表。 |
| `frontend/src/pro/api/custom.ts` | `creatCustomDataView` | `/api/custom/create` | 保存新模板。 |
| `frontend/src/pro/api/custom.ts` | `updateCustomDataView` | `/api/custom/update` | 更新模板。 |
| `frontend/src/pro/api/custom.ts` | `deleteCustomDataView` | `/api/custom/delete` | 删除模板。 |
| `frontend/src/pro/api/modelSetting.ts` | `getModelListApi` | `/ai/api/v1/model-setting/list` | Agent 模型列表读取。 |
| `frontend/src/pro/api/modelSetting.ts` | `setDefaultModelApi` | `/ai/api/v1/model-setting/default` | Agent 默认模型切换。 |

## Controller Or Router Files

| Type | File | Notes |
|---|---|---|
| Custom Handler | `backend/pkg/api/custom/handler.go` | `/api/custom/*` 接口集合。 |
| Custom CRUD | `backend/pkg/api/custom/func_getcustomquery.go`, `func_createcustomquery.go`, `func_updatecustomquery.go`, `func_deletecustomquery.go` | 读取、创建、更新、删除自定义数据视图模板。 |
| Alert Report Handler | `backend/pkg/api/alerts/func_getalerteventreport.go` | `/api/alerts/events/report/get`。 |
| Alert Report Save | `backend/pkg/api/alerts/func_addalerteventreport.go` | `/api/alerts/events/report/add`。 |

## Service Files

| Service Area | Files | Notes |
|---|---|---|
| Custom Query Service | `backend/pkg/services/custom/service.go` | 模块内自定义数据视图 service 入口。 |
| Custom CRUD | `service_get_custom.go`, `service_creat_custom.go`, `service_update_custom.go`, `service_delete_custom.go` | 把前端 `type[]` 与 `custom_query_service.querytype` 逗号串互转。 |
| Alert Report Read | `backend/pkg/services/alerts/service_alert_event_report.go` | 从 ClickHouse `alert_report` 读取单份报告。 |
| Alert Report Write | `backend/pkg/services/alerts/service_add_alert_event_report.go` | 如果同标签窗口内已有报告则拒绝重复写入。 |

## Model Entity Schema Files

| Type | File | Notes |
|---|---|---|
| Custom Query Table | `backend/pkg/repository/database/dao_custom_query_type.go` | `QueryServiceType.TableName()` 为 `custom_query_service`。 |
| Custom Request | `backend/pkg/model/request/custom_req.go` | `CreateCustomRequest`、`UpdateCustomRequest`、`DeleteCustomRequest`。 |
| Custom Response | `backend/pkg/model/response/custom_resp.go` | `GetCustomResponse.Customs` 和 `CustonType`。 |
| Alert Report Request | `backend/pkg/model/request/alert_event.go` | `GetAlertEventReportRequest`、`AddAlertEventReportRequest`。 |
| Alert Report Response | `backend/pkg/model/response/alert_event.go` | `GetAlertEventReportResponse`、`ReportData`。 |
| Alert Report Storage | `backend/pkg/repository/clickhouse/dao_alert_event_report.go` | 直接读写 `alert_report`。 |

## Important Shared Dependencies

| Dependency | File | Notes |
|---|---|---|
| Shared Request Utility | `frontend/src/core/utils/request.js` | 为 `/api/*`、`/ai/api/*`、`/originx/api/*` 提供统一 token 与语言头处理。 |
| Route Guard | `frontend/src/core/components/AuthRouter.tsx` | 本模块各路由都经 `AuthRouter` 包装，但实际 `useEffect` 权限检查被注释。 |
| User/Menu Context | `frontend/src/core/contexts/UserContext.js` | Agent 与平台入口菜单最终受后端 `permission/config` 结果影响。 |
| Time Range Store | `frontend/src/core/store/reducers/timeRangeReducer` | ChatAI、Report、CauseReport 都依赖全局时间范围。 |
| License Context | `frontend/src/pro/contexts/LicenseContext.tsx` | 平台级 license 信息会在布局层影响本模块可见提示。 |
| Dataplane Check | `frontend/src/pro/api/dataplane.ts` | `AGUIContext.checkPermission` 使用 `/api/dataplane/data/check` 判断 Agent 前置条件。 |
| Dify/Workflow Token Refresh | `frontend/src/core/utils/request.js` | 登录刷新时顺带刷新 `difyToken`，影响 workflow/AI 相关会话边界。 |

## Evidence Notes

1. `ChatAICom` 和 `ChatAIContext` 当前主流程都直接使用 `streamFetch('/ai/api/*')`，并未复用 `chatAI.ts` 里的 `/api/chat` 与 `/api/alertchat` wrapper。
2. `/report` 页面唯一稳定的本仓库后端映射是 `/api/alerts/events/report/get`；`agentChatId` 的生产来源在当前仓库中未完全呈现。
3. `/agent` 页面依赖 `AGUIClient`、`@ag-ui/client` 和 `/ai/api/v1/srechat`，属于外部 Agent 服务边界，不应误写成主 Go 后端实现。
4. `frontend/vite.config.mjs` 只显式代理 `/ai/api`，未看到 `/originx/api` 的开发代理定义；`originx` 边界应标记为 To Be Confirmed。
5. `PreUseInspection.tsx` 里仍引用 `/#/agent-datasource`，但当前实际业务路由是 `/dataplane`，说明存在旧入口残留。

