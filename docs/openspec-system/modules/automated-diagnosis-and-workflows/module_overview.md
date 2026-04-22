# Automated Diagnosis And Workflows Module Overview

## Module Goal
本模块负责把已发生的告警和异常转换成自动化诊断与 workflow 执行入口。前端包括 `/workflows` 的 Dify 工作流控制台入口、`/alert-analyze` 的告警影响面分析、`/alert-analyze/info` 的下游异常拓扑分析和突变检测配置；后端包括 `/api/alerts/*` 下的 `alertanalyze` 路由与 `repository/dify` 中的 workflow 执行能力。

## Scope
当前代码证据支持以下范围：

- `/workflows` 页面通过 iframe 打开 Dify 应用控制台。
- Dify 匿名登录、token 注入与 token 刷新监听。
- `/alert-analyze` 告警事件列表、受影响入口列表、故障贡献度排序、异常列表弹窗。
- `/alert-analyze/info` 下游拓扑分析、异常趋势、delta 播放、节点详情。
- 突变检测快速配置、执行记录、规则记录与同步告警规则。
- 异常 span 查询。
- 后端根据告警 tags 反推 service / endpoint，并结合拓扑与实例信息聚合下游异常。
- Dify async alert check、自动 analyze workflow 和 alert classify 调用。

当前代码不应纳入本模块主体范围的能力：

- `alert-rules-and-incident-events`：告警规则、receiver、事件详情、静默、resolve。
- `conversational-diagnosis-and-reporting`：`/alert-analyze/chatAI/:service/:endpoint`、报告页面和聊天页面主体。
- `service-observability`：服务总览、拓扑、基础 service 查询 API。
- `trace-exploration-and-profiling`：trace 详情、trace 页面和 flame graph。

## Frontend Entry

| Entry Type | Evidence | Notes |
|---|---|---|
| Route | `frontend/src/oss/routes.js` | `/workflows`。 |
| Route | `frontend/src/pro/routes.js` | `/alert-analyze`、`/alert-analyze/info`。 |
| Page | `frontend/src/oss/views/workflows/index.tsx` | 工作流控制台入口页。 |
| Page | `frontend/src/pro/views/alertAnalyze/index.jsx` | 告警分析页。 |
| Page | `frontend/src/pro/views/serviceAnalyze/index.jsx` | 下游异常拓扑分析页。 |
| Context | `AlertAnalyzeContext`, `ServiceAnalyzeContext` | 页面主状态管理。 |
| API Wrapper | `frontend/src/core/api/workflows.ts`, `frontend/src/pro/api/alertAnalyze.js` | 本模块主要前端 API 封装。 |

## Backend Entry

| Entry Type | Evidence | Notes |
|---|---|---|
| Router Group | `backend/pkg/router/router_api.go` | `alertanalyze` routes 注册在 `/api/alerts` 下，并位于 `AuthMiddleware()` 之后。 |
| Handler Package | `backend/pkg/api/alertanalyze` | 本模块主要 handler。 |
| Service Package | `backend/pkg/services/alertanalyze` | 告警分析、下游异常、突变检测逻辑。 |
| Workflow Repo | `backend/pkg/repository/dify` | Dify workflow 调用和异步工作流执行。 |
| Models | `backend/pkg/model/request/*`, `response/*`, `workflow_records.go`, `mutation_detect.go` | 请求响应、workflow record 和突变检测模型。 |

## Core Concepts

| Concept | Explanation |
|---|---|
| `WorkflowsIframe` | 通过 iframe 打开 `/dify/apps`，并把 `access_token`、`refresh_token` 拼到 URL 上。 |
| `AlertAnalyzeContext` | `/alert-analyze` 的状态中心，负责告警列表、受影响入口、拓扑、趋势和 delta 数据。 |
| `ServiceAnalyzeContext` | `/alert-analyze/info` 的状态中心，负责指定 `service + endpoint` 的拓扑与异常分析。 |
| `ImpactAlertEvent` | 后端为告警分析页构造的告警事件结构，包含 `impactEndpoints`、`impactEntrys` 和 `isAnalyzed`。 |
| `AlertRelatedEntry` | 影响面分析返回的入口对象，包含 `relatedAlertRate`。 |
| `DescendantAnormal*` | 下游异常的计数、事件记录和趋势返回结构。 |
| `DetectMutation` | 突变检测执行记录和规则记录模型。 |
| `WorkflowRecord` | Dify workflow 执行记录，包含 `workflowRunId`、`output`、`alertDirection`、`analyzeRunId` 等字段。 |
| `AlertEventClassify` | 告警分组和告警名映射到具体 analyze workflow 的缓存结果。 |
| `WorkflowParams` | 自动 analyze workflow 的参数载体，由告警事件和其关联服务推导。 |

## Related Files

- `frontend/src/oss/routes.js`
- `frontend/src/pro/routes.js`
- `frontend/src/core/api/workflows.ts`
- `frontend/src/pro/api/alertAnalyze.js`
- `frontend/src/oss/views/workflows/index.tsx`
- `frontend/src/oss/views/workflows/workflowsIframe.tsx`
- `frontend/src/pro/views/alertAnalyze/index.jsx`
- `frontend/src/pro/views/alertAnalyze/components/index.jsx`
- `frontend/src/pro/views/alertAnalyze/components/AlertEvent/index.jsx`
- `frontend/src/pro/views/alertAnalyze/components/ImpactList/index.jsx`
- `frontend/src/pro/views/alertAnalyze/components/AlertRuleQuickSet/index.jsx`
- `frontend/src/pro/views/alertAnalyze/components/AnomalyDetectionRecord/index.jsx`
- `frontend/src/pro/views/serviceAnalyze/index.jsx`
- `frontend/src/pro/views/serviceAnalyze/DescendantTopology/index.jsx`
- `frontend/src/pro/contexts/AlertAnalyzeContext.js`
- `frontend/src/pro/contexts/ServiceAnalyzeContext.tsx`
- `backend/pkg/router/router_api.go`
- `backend/pkg/api/alertanalyze/handler.go`
- `backend/pkg/services/alertanalyze/service.go`
- `backend/pkg/services/alertanalyze/endpoints_map.go`
- `backend/pkg/services/alertanalyze/service_get_alert_events_svc.go`
- `backend/pkg/services/alertanalyze/service_get_alerts_impact.go`
- `backend/pkg/services/alertanalyze/service_get_descendant_contributaion.go`
- `backend/pkg/services/alertanalyze/service_get_descendant_anormal.go`
- `backend/pkg/services/alertanalyze/service_get_descendant_anormal_trend.go`
- `backend/pkg/services/alertanalyze/service_get_descendant_anormal_delta.go`
- `backend/pkg/services/alertanalyze/service_defects_detect.go`
- `backend/pkg/repository/dify/dao.go`
- `backend/pkg/repository/dify/async_workflow.go`
- `backend/pkg/repository/dify/workflow_worker.go`

## Dependencies

| Dependency Type | Dependency | Why It Matters |
|---|---|---|
| Frontend Shared | `timeRangeReducer` | 分析页和服务分析页都依赖时间范围。 |
| Frontend Shared | `dataGroupReducer` | 多数分析请求都会传 `groupId`。 |
| Frontend Shared | `ChartsContext` | 影响面列表内嵌小图。 |
| Frontend Shared | `DataSourceFilter` | 影响面列表的过滤 UI。 |
| Backend Shared | `clickhouse.Repo` | 告警事件、异常 span、检测执行记录主要查询源。 |
| Backend Shared | `prometheus.Repo` | 查询 active instances、数据库下游、执行 mutation check。 |
| Backend Shared | `polarisanalyzer.Repo` | 下游贡献度排序。 |
| Backend Shared | `kubernetes.Repo` | 突变检测同步告警规则。 |
| External Boundary | `repository/dify` | workflow 分类、alert check 和 analyze 执行。 |
| Cross Module | `conversational-diagnosis-and-reporting` | 影响面列表会打开 `/alert-analyze/chatAI/:service/:endpoint`。 |

## Risks And Unknowns

1. `getAlertDescendantApi` 前端 wrapper 与后端实际 `/api/alerts/descendant/anormal/list` 路径不一致。
2. `WorkflowsIframe` 只在匿名用户且本地无 token 时触发匿名登录，非匿名用户 token 生命周期如何保证需要确认。
3. Dify 相关 `/dify/console/api/*` 前端调用未在当前主后端路由中观察到实现，应作为外部边界处理。
4. `alertanalyze` 的很多响应类型放在 `response/alert_input.go` 和 `response/service_resp.go` 中，模型归属较混杂。
5. `TriggerAdapterUpdate` 属于数据接入模块，不应和本模块混淆，但其输出会影响 trace API 分析边界。
