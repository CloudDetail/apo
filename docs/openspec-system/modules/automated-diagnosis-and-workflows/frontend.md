# Automated Diagnosis And Workflows Frontend

## Pages

| Page | Route | Description |
|---|---|---|
| `WorkflowsPage` | `/workflows` | Dify 工作流控制台入口页。 |
| `AlertAnalyzePage` | `/alert-analyze` | 告警分析主页面，展示告警事件列表与受影响入口分析。 |
| `ServiceAnalyzePage` | `/alert-analyze/info` | 下游异常拓扑分析页面。 |

## Routes

| File | Route | Component | Notes |
|---|---|---|---|
| `frontend/src/oss/routes.js` | `/workflows` | `WorkflowsPage` | 菜单项 `workflows`。 |
| `frontend/src/pro/routes.js` | `/alert-analyze` | `AlertAnalyzePage` | `showDataGroup: true`。 |
| `frontend/src/pro/routes.js` | `/alert-analyze/info` | `ServiceAnalyzePage` | `showDataGroup: 'view'`。 |

## Menus

| Source | Key | Route | Notes |
|---|---|---|---|
| `backend/pkg/repository/database/init_menu_item.go` | `workflows` | `/workflows` | 菜单项已初始化。 |
| `backend/pkg/repository/database/init_feature_mapping.go` | `工作流` | `workflows` | feature 与菜单项存在映射。 |
| `backend/pkg/repository/database/init_feature_mapping.go` | `告警事件` | `/alerts/events/detail/:alertID/:eventID` | 与本模块下游分析相关，但不等于本模块入口。 |

## Key Components

| Component | File | Implemented Behavior |
|---|---|---|
| `WorkflowsIframe` | `frontend/src/oss/views/workflows/workflowsIframe.tsx` | 检查本地 `difyToken` / `difyRefreshToken`，必要时匿名登录，并监听 iframe 发回的 `refresh_token` 消息。 |
| `AlertAnalyzeCom` | `frontend/src/pro/views/alertAnalyze/components/index.jsx` | 上下分栏，顶部按钮打开突变检测配置和检测记录弹窗。 |
| `AlertEventTable` | `components/AlertEvent/index.jsx` | 展示告警事件、raw tags、受影响应用和受影响入口。 |
| `ImpactList` | `components/ImpactList/index.jsx` | 展示入口列表、延迟/错误率/TPS 图、故障贡献度和三个动作按钮。 |
| `AlertRuleQuickSet` | `components/AlertRuleQuickSet/index.jsx` | 从后端拉预定义指标，生成 `DetectMutationRequest` 并提交。 |
| `AnomalyDetectionRecord` | `components/AnomalyDetectionRecord/index.jsx` | 分割视图展示检测执行记录和规则记录。 |
| `DescendantTopology` | `frontend/src/pro/views/serviceAnalyze/DescendantTopology/index.jsx` | 拓扑图、趋势图、异常类型过滤、播放条和节点详情。 |

## State Management

| State | File | Notes |
|---|---|---|
| `AlertAnalyzeContext` | `frontend/src/pro/contexts/AlertAnalyzeContext.js` | 管理 `alertEvents`、`eventImpactList`、`selectedEndpointInfo`、`topologyData`、`anormalTypes`、`delta*` 等。 |
| `ServiceAnalyzeContext` | `frontend/src/pro/contexts/ServiceAnalyzeContext.tsx` | 管理 `service`、`endpoint`、`clusterIds`、拓扑和 delta/趋势数据。 |
| Redux Time Range | `timeRangeReducer` | 两个页面都依赖。 |
| Redux Data Group | `dataGroupReducer` | 两个页面都向后端传 `groupId`。 |
| Redux Group Labels | `groupLabelReducer` | 告警事件表展示 group label。 |
| Local Storage | `WorkflowsIframe` | 保存 `difyToken` 和 `difyRefreshToken`。 |

## API Usage

| UI Action | Wrapper | API Path | Notes |
|---|---|---|---|
| 匿名登录 Dify | `workflowAnonymousLoginApi` | `/dify/console/api/login/apo/anonymous` | `WorkflowsIframe` 当前实际调用。 |
| 查询告警分析事件 | `getServiceAlertEventsApi` | `core/api/service` | `AlertAnalyzeContext.fetchData()` 使用。 |
| 影响面分析 | `getAlertEventImpactApi` | `/api/alerts/event/impact` | 右侧入口列表主数据。 |
| 下游趋势 | `getAlertDescendantTrendApi` | `/api/alerts/descendant/anormal/trend` | 生成趋势图。 |
| 下游 delta | `getAlertDescendantDeltaApi` | `/api/alerts/descendant/anormal/delta` | 生成 origin / triggered / final 数据。 |
| 贡献度排序 | `getAlertDescendantContributionApi` | `/api/alerts/descendant/anormal/contribution` | `ImpactList` 嵌套列调用。 |
| 下游异常列表 | `getAlertDescendantEventApi` | `/api/alerts/descendant/anormal/list` | 异常列表弹窗。 |
| 异常 span | `getFailureReportApi` | `/api/alerts/anomaly-span/list` | 节点详情里的故障报告弹窗使用。 |
| 预定义指标 | `getMetricsApi` | `/api/alerts/detect/mutation/metrics` | 突变检测快速指标。 |
| 提交突变检测 | `addMutationRule` | `api/alerts/detect/mutation/add` | 保存突变检测配置。 |
| 执行记录 | `getDetectionRecordApi` | `/api/alerts/detect/mutation/exec-list` | 检测记录列表。 |
| 规则记录 | `getMutationList` | `/api/alerts/detect/mutation/rule-list` | 规则记录列表。 |

## User Interaction Flow

1. 用户进入 `/alert-analyze`。
2. `AlertAnalyzeContext` 先通过 `getServiceAlertEventsApi` 拉取告警事件列表。
3. 用户点击一键分析后，`getAlertEventImpact` 拉取受影响入口列表。
4. `ImpactList` 为当前页入口列表额外请求 charts 数据，并可继续请求贡献度数据。
5. 用户可以从 `ImpactList` 打开下游异常拓扑页 `#/alert-analyze/info?...`、打开异常列表弹窗，或打开 `/alert-analyze/chatAI/...`。
6. 在 `/alert-analyze/info` 页面，`ServiceAnalyzeContext` 根据 query 中的 `service`、`endpoint`、`clusterIds` 请求拓扑、趋势和 delta 数据。
7. 用户可以在 `AlertRuleQuickSet` 弹窗中配置并提交突变检测；也可以在 `AnomalyDetectionRecord` 弹窗中查看执行记录和规则记录。
8. `/workflows` 页面单独提供 Dify 控制台入口，用于查看工作流配置。

## Permissions

| Area | Evidence | Notes |
|---|---|---|
| Route Guard | `AuthRouter` | `/workflows`、`/alert-analyze`、`/alert-analyze/info` 都被包装。 |
| Menu Visibility | `init_menu_item.go`, `init_feature_mapping.go` | `/workflows` 菜单来源于后端初始化。 |
| Backend Middleware | `router_api.go` | `alertanalyze` routes 在 `/api/alerts` 的 `AuthMiddleware()` 之后。 |

## Notes

- `ImpactList` 打开的 `/alert-analyze/chatAI/:service/:endpoint` 不是本模块主体，应在会话式诊断文档中继续展开。
- `WorkflowsIframe` 的 token 刷新依赖 iframe `postMessage`，页面本身没有显式刷新 token API 调用。
- 告警分析页会复用 `service` 模块的接口和图表上下文，因此它不是一个完全独立的前端子系统。
