# Automated Diagnosis And Workflows Source Map

## Scope
本文档固定 `automated-diagnosis-and-workflows` 模块的实际代码证据。该模块覆盖 Dify workflow 控制台入口、告警分析页、下游异常拓扑分析页、突变检测配置与记录，以及后端 `alertanalyze` 服务和 `dify` 仓储边界。

## Relevant Source Directories

| Layer | Path | Notes |
|---|---|---|
| Frontend | `frontend/src/oss/views/workflows` | `/workflows` 页面，通过 iframe 打开 `/dify/apps`。 |
| Frontend | `frontend/src/core/api/workflows.ts` | Dify console 登录和匿名登录 wrapper。 |
| Frontend | `frontend/src/pro/views/alertAnalyze` | `/alert-analyze` 告警分析页面。 |
| Frontend | `frontend/src/pro/views/serviceAnalyze` | `/alert-analyze/info` 下游异常拓扑分析页面。 |
| Frontend | `frontend/src/pro/api/alertAnalyze.js` | 告警分析和突变检测相关 API wrapper。 |
| Frontend | `frontend/src/pro/contexts/AlertAnalyzeContext.js` | `/alert-analyze` 页面状态管理与数据拉取。 |
| Frontend | `frontend/src/pro/contexts/ServiceAnalyzeContext.tsx` | `/alert-analyze/info` 页面状态管理与数据拉取。 |
| Backend | `backend/pkg/router/router_api.go` | `/api/alerts/*` 中属于 `alertanalyze` 的 route 注册位置。 |
| Backend | `backend/pkg/api/alertanalyze` | 本模块主要 handler 包。 |
| Backend | `backend/pkg/services/alertanalyze` | 影响面、贡献度、下游异常、趋势、delta、突变检测、异常 span 查询。 |
| Backend | `backend/pkg/repository/dify` | workflow 执行、alert classify、async alert check、自动 analyze 的外部服务仓储。 |
| Backend | `backend/pkg/model/request/alert_analyze.go` | 告警分析请求模型。 |
| Backend | `backend/pkg/model/request/service_req.go` | 下游异常分析、趋势、delta、异常 span 等请求模型。 |
| Backend | `backend/pkg/model/response/alert_input.go`, `service_resp.go` | 告警分析与下游异常返回模型。 |
| Backend | `backend/pkg/model/workflow_records.go` | workflow record 和 `AlertEventClassify` 模型。 |
| Backend | `backend/pkg/model/mutation_detect.go` | 预定义指标表达式和 `DetectMutation` 记录模型。 |

## Route Files

| Side | File | Observed Routes |
|---|---|---|
| Frontend | `frontend/src/oss/routes.js` | `/workflows` |
| Frontend | `frontend/src/pro/routes.js` | `/alert-analyze`, `/alert-analyze/info` |
| Frontend Downstream | `frontend/src/pro/routes.js` | `/alert-analyze/chatAI/:service/:endpoint`，由分析页按钮打开，但主体属于会话式诊断模块。 |
| Backend | `backend/pkg/router/router_api.go` | `/api/alerts/alert/events`, `/api/alerts/event/impact`, `/api/alerts/descendant/anormal/contribution`, `/api/alerts/descendant/anormal/list`, `/api/alerts/descendant/anormal/trend`, `/api/alerts/descendant/anormal/delta`, `/api/alerts/anomaly-span/list`, `/api/alerts/detect/mutation/metrics`, `/api/alerts/detect/mutation/add`, `/api/alerts/detect/mutation/exec-list`, `/api/alerts/detect/mutation/rule-list` |
| External Or Adjacent | `frontend/src/core/api/workflows.ts` | `/dify/console/api/login/apo`, `/dify/console/api/logout`, `/dify/console/api/login/apo/anonymous` |

## Page Files

| Type | File | Notes |
|---|---|---|
| Route Page | `frontend/src/oss/views/workflows/index.tsx` | `/workflows` 页面，嵌入 `WorkflowsIframe src="/dify/apps"`。 |
| Shared Iframe | `frontend/src/oss/views/workflows/workflowsIframe.tsx` | 负责匿名登录、token 注入和 token 刷新监听。 |
| Route Page | `frontend/src/pro/views/alertAnalyze/index.jsx` | `/alert-analyze` 页面容器，包裹 `AlertAnalyzeProvider`。 |
| Main Composite | `frontend/src/pro/views/alertAnalyze/components/index.jsx` | 页面主体，左侧告警事件列表，右侧影响面分析或一键分析空态。 |
| Alert Events | `frontend/src/pro/views/alertAnalyze/components/AlertEvent/index.jsx` | 告警事件列表、标签展开和分页。 |
| Impact List | `frontend/src/pro/views/alertAnalyze/components/ImpactList/index.jsx` | 受影响入口列表、图表、故障贡献度、下钻按钮。 |
| Mutation Quick Set | `frontend/src/pro/views/alertAnalyze/components/AlertRuleQuickSet/index.jsx` | 弹窗配置突变检测规则。 |
| Detection Record | `frontend/src/pro/views/alertAnalyze/components/AnomalyDetectionRecord/index.jsx` | 执行记录和规则记录双面板。 |
| Route Page | `frontend/src/pro/views/serviceAnalyze/index.jsx` | `/alert-analyze/info` 页面容器，包裹 `ServiceAnalyzeProvider`。 |
| Topology Composite | `frontend/src/pro/views/serviceAnalyze/DescendantTopology/index.jsx` | 下游拓扑、时间线、异常类型过滤、播放和节点详情。 |

## API Wrapper Files

| File | Wrapper | API Path | Notes |
|---|---|---|---|
| `frontend/src/core/api/workflows.ts` | `workflowLoginApi` | `/dify/console/api/login/apo` | wrapper 存在，当前工作流页未直接调用。 |
| `frontend/src/core/api/workflows.ts` | `workflowLogoutApi` | `/dify/console/api/logout` | wrapper 存在，当前工作流页未直接调用。 |
| `frontend/src/core/api/workflows.ts` | `workflowAnonymousLoginApi` | `/dify/console/api/login/apo/anonymous` | `WorkflowsIframe` 当前实际调用。 |
| `frontend/src/pro/api/alertAnalyze.js` | `getAlertEventImpactApi` | `/api/alerts/event/impact` | 影响面分析。 |
| `frontend/src/pro/api/alertAnalyze.js` | `getAlertDescendantApi` | `/api/alerts/descendant/anormal` | wrapper 路径与后端实际 `/list` 不一致。 |
| `frontend/src/pro/api/alertAnalyze.js` | `getAlertDescendantDeltaApi` | `/api/alerts/descendant/anormal/delta` | 下游异常 delta。 |
| `frontend/src/pro/api/alertAnalyze.js` | `getAlertDescendantTrendApi` | `/api/alerts/descendant/anormal/trend` | 下游异常趋势。 |
| `frontend/src/pro/api/alertAnalyze.js` | `getAlertDescendantContributionApi` | `/api/alerts/descendant/anormal/contribution` | 下游故障贡献度。 |
| `frontend/src/pro/api/alertAnalyze.js` | `getAlertDescendantEventApi` | `/api/alerts/descendant/anormal/list` | 某入口下游异常事件列表。 |
| `frontend/src/pro/api/alertAnalyze.js` | `getFailureReportApi` | `/api/alerts/anomaly-span/list` | 按 service / endpoint / reason 查询异常 span。 |
| `frontend/src/pro/api/alertAnalyze.js` | `getMetricsApi` | `/api/alerts/detect/mutation/metrics` | 突变检测预定义指标。 |
| `frontend/src/pro/api/alertAnalyze.js` | `addMutationRule` | `api/alerts/detect/mutation/add` | 添加突变检测规则。 |
| `frontend/src/pro/api/alertAnalyze.js` | `getDetectionRecordApi` | `/api/alerts/detect/mutation/exec-list` | 获取执行记录。 |
| `frontend/src/pro/api/alertAnalyze.js` | `getMutationList` | `/api/alerts/detect/mutation/rule-list` | 获取规则列表。 |

## Controller Or Router Files

| Type | File | Notes |
|---|---|---|
| Router | `backend/pkg/router/router_api.go` | `alertanalyze` routes 注册在 `/api/alerts` group 且位于 `AuthMiddleware()` 之后。 |
| Handler Factory | `backend/pkg/api/alertanalyze/handler.go` | 构造 `alertanalyze.Service`、`serviceoverview.Service` 和 `service.Service`。 |
| Impact Handler | `func_getalertimpact.go` | `/api/alerts/event/impact`。 |
| Alert Events Handler | `func_getalertevents.go` | `/api/alerts/alert/events`。 |
| Descendant Handler | `func_getdescendantcontribution.go`, `func_getdescendantanormal.go`, `func_getanormaltrendbyentry.go`, `func_getdescendantanormaldelta.go` | 下游异常贡献度、列表、趋势、delta。 |
| Anomaly Span Handler | `func_getanomalyspan.go` | `/api/alerts/anomaly-span/list`。 |
| Mutation Handler | `func_getpredefineddetectexpr.go`, `func_adddefectsdetect.go`, `func_getdefectdetectexeclist.go`, `func_getdefectdetectrulelist.go` | 突变检测指标、执行、记录、规则。 |

## Service Files

| Service Area | Files | Notes |
|---|---|---|
| Service Construction | `backend/pkg/services/alertanalyze/service.go` | 聚合 ClickHouse、Prometheus、Polaris、Kubernetes、Database 仓储。 |
| Alert Events | `service_get_alert_events_svc.go` | 查询 firing + mutation 告警，并推断 impact services 和 entry endpoints。 |
| Impact Analysis | `service_get_alerts_impact.go` | 统计入口与告警关联率。 |
| Descendant Contribution | `service_get_descendant_contributaion.go` | 通过 Polaris 相关性排序下游节点。 |
| Descendant Alerts | `service_get_descendant_anormal.go` | 汇总入口下游节点的异常事件。 |
| Descendant Trend | `service_get_descendant_anormal_trend.go` | 生成趋势 chart 数据。 |
| Descendant Delta | `service_get_descendant_anormal_delta.go` | 生成 origin / triggered / final 异常计数和事件记录。 |
| Anomaly Span | `service_get_anomaly_span.go` | 查询用于故障报告的异常 span。 |
| Mutation Detection | `service_predefined_detect_expr_part.go`, `service_defects_detect.go`, `service_defects_exec_list.go`, `service_defects_rule_list.go` | 预定义指标、执行突变检测、写入 ClickHouse、同步告警规则、查询执行记录和规则。 |
| Endpoint Mapping | `endpoints_map.go` | 从告警 tags 推断 service / endpoint。 |
| Dify Workflow | `backend/pkg/repository/dify/*` | Dify workflow 执行、classify、async alert check、自动 analyze。 |

## Important Shared Dependencies

| Dependency | File | Notes |
|---|---|---|
| `getServiceAlertEventsApi` | `core/api/service` | `AlertAnalyzeContext` 拉取告警事件列表时复用 service 域接口。 |
| `getServiceRelationApi` | `core/api/service` | `AlertAnalyzeContext` 和 `ServiceAnalyzeContext` 拉取服务拓扑。 |
| `DataSourceFilter` | `src/core/components/Filter/DataSourceFilter` | 影响面列表支持 service / endpoint / namespace / cluster 过滤。 |
| `ChartsContext` | `src/core/contexts/ChartsContext` | 影响面列表的 latency / errorRate / tps 图元依赖。 |
| `timeRangeReducer` | `src/core/store/reducers/timeRangeReducer` | 告警分析和服务分析共享时间范围。 |
| `dataGroupReducer` | Redux | 告警分析和服务分析都向后端传 `groupId`。 |
| `groupLabelReducer` | Redux | 告警事件表展示 group label。 |
| `repository/dify` | backend | 工作流自动分类、alert check 和 analyze 的外部工作流边界。 |

## Evidence Notes

1. `WorkflowsIframe` 当前只在 `user.username === 'anonymous'` 且本地没有 token 时调用 `workflowAnonymousLoginApi`。
2. `ImpactList` 会打开 `#/alert-analyze/info?...` 和 `#/alert-analyze/chatAI/...`；后者应归入会话式诊断模块。
3. `frontend/src/pro/api/alertAnalyze.js` 中 `getAlertDescendantApi` 路径是 `/api/alerts/descendant/anormal`，而当前后端实际注册的是 `/api/alerts/descendant/anormal/list`。
4. `func_updatecluster.go` 的空错误分支属于数据接入模块，这里不归入本模块。
5. `router_api.go` 中 `alertanalyze` routes 位于 `/api/alerts` group 的 `AuthMiddleware()` 之后，与事件列表接口的权限层次不同。
