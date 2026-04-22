# Alert Rules And Incident Events Module Overview

## Module Goal
本模块围绕“告警治理”和“事件处置”提供已实现能力：用户可以维护告警规则、维护通知 receiver、查看告警事件列表、进入某个告警事件详情、查看同一 `alertId` 的历史事件、触发 Dify workflow、查看 workflow 结果、设置静默并手动 resolve 最新事件。

## Scope
当前代码证据支持以下范围：

- 告警规则列表、搜索、新增、编辑、删除和状态展示。
- 告警规则 group label 与 quick metrics 辅助编辑。
- AlertManager receiver 列表、新增、编辑、删除，覆盖 email、webhook、dingtalk、wechat 配置结构。
- 告警事件列表、统计、分页、过滤、有效性展示、workflow run-once 入口和 workflow logs 入口。
- 告警事件详情、同一 `alertId` 历史事件、通知发送状态、当前事件定位。
- 告警静默读取和保存，接口路径保持 `slient`。
- 手动 resolve 最新告警事件。
- AlertManager 输入 `/api/alerts/inputs/alertmanager` 与 DingTalk 输出 `/api/alerts/outputs/dingtalk/:uuid` 的后端接入点。
- 告警报告读取和新增的后端接口，以及报告页通过 `alertEventId` 消费报告的前端入口。

当前代码不应纳入本模块主体范围的能力：

- `alert-sources-and-enrichment`：`/api/alertinput/*`、告警来源、schema、cluster、enrich rule 管理。
- `automated-diagnosis-and-workflows`：`/alert-analyze`、异常贡献度、突变检测、`alertanalyze` service 主体。
- `conversational-diagnosis-and-reporting`：报告页面的展示细节、chat、agent、custom data view 主体。
- `service-observability` 中的 `/api/service/alert/sample/events`。
- 外部 `/alert/api/v1/rules` 与 `/alert/-/reload` 的服务端实现，当前仓库主后端未观察到对应 route。

## Frontend Entry

| Entry Type | Evidence | Notes |
|---|---|---|
| Route | `frontend/src/oss/routes.js` | 注册 `/alerts/rule`、`/alerts/notify`、`/alerts/events`、`/alerts/events/detail/:alertId/:eventId`。 |
| Page | `frontend/src/oss/views/alerts/AlertsRule.jsx` | 告警规则管理页。 |
| Page | `frontend/src/oss/views/alerts/AlertsNotify.jsx` | 告警通知 receiver 管理页。 |
| Page | `frontend/src/oss/views/alertEvents/index.tsx` | 告警事件列表页。 |
| Page | `frontend/src/oss/views/alertEvents/detail/index.tsx` | 告警事件详情页。 |
| API Wrapper | `frontend/src/core/api/alerts.js` | 本模块主要前端 API 封装。 |
| Shared State | `groupLabelReducer`, `dataGroupReducer`, `timeRangeReducer` | 分别提供 group label、数据组、时间范围。 |
| Cross Page | `frontend/src/pro/views/report/index.tsx` | 报告页通过 `getReportApi` 读取告警事件报告。 |

## Backend Entry

| Entry Type | Evidence | Notes |
|---|---|---|
| Router Group | `backend/pkg/router/router_api.go` | `/api/alerts` route group，统一使用 `LicMiddleware`，中途追加 `AuthMiddleware`。 |
| Handler Package | `backend/pkg/api/alerts` | 本模块主要 controller/handler。 |
| Service Package | `backend/pkg/services/alerts` | 规则、receiver、事件、报告、静默、转发等业务逻辑。 |
| Request Schemas | `backend/pkg/model/request/alert_input.go`, `alert_event.go`, `alert_slience.go` | 请求结构分散在三个文件。 |
| Response Schemas | `backend/pkg/model/response/alert_input.go`, `alert_event.go`, `alert_slience.go` | 响应结构分散在三个文件。 |
| Repositories | `clickhouse`, `kubernetes`, `database`, `dify` | 事件查询、规则配置、receiver 持久化和 workflow 边界。 |
| Integration Service | `backend/pkg/services/integration/alert` | `/api/alerts/inputs/alertmanager` 当前实际委托到这里处理。 |

## Core Concepts

| Concept | Explanation |
|---|---|
| `AlertRule` | 告警规则配置对象，包含 `group`、`alert`、`expr`、`for`、`labels`、`annotations`，通过 `k8sApi` 写入规则配置。 |
| `AlertManagerConfigReceiver` | 通知 receiver 配置，前端支持 email、webhook、dingtalk、wechat 四类表单结构。 |
| `AlertEvent` | 告警事件记录，后端主要从 ClickHouse `alert_event` 查询，并和 workflow/notify 记录聚合。 |
| `alertId` | 同一告警对象的聚合标识，详情页按 `alertId` 查询同类历史事件。 |
| `eventId` | 单次告警事件记录标识，详情页通过 `alertId + eventId` 定位当前事件。 |
| `validity` | workflow 结果派生出的有效性状态，前端展示 `valid`、`invalid`、`unknown`、`skipped`、`failed`。 |
| `alertCheckId` | Dify alert check workflow id，前端用于拼接 `/dify/app/{alertCheckId}/logs/{workflowRunId}`。 |
| `WorkflowParams` | 后端为前端 run-once 入口构造的参数，包含时间范围、节点和序列化的告警上下文。 |
| `slient` / `Slience` | 当前实现中用于静默能力的接口和响应字段拼写，文档保持原样。 |
| `DataGroup` | 告警规则和事件查询的共享数据范围上下文，前端传 `groupId`。 |
| `quick_alert_rule_metric` | 告警规则编辑器的 quick metrics 数据表。 |
| `target_tags` | 告警事件 tag display 的展示名来源。 |

## Related Files

- `frontend/src/oss/routes.js`
- `frontend/src/core/api/alerts.js`
- `frontend/src/oss/views/alerts/AlertsRule.jsx`
- `frontend/src/oss/views/alerts/AlertsNotify.jsx`
- `frontend/src/oss/views/alerts/modal/ModifyAlertRuleModal.jsx`
- `frontend/src/oss/views/alerts/modal/ALertConditionCom.jsx`
- `frontend/src/oss/views/alerts/modal/ModifyAlertNotifyModal.jsx`
- `frontend/src/oss/views/alertEvents/index.tsx`
- `frontend/src/oss/views/alertEvents/detail/index.tsx`
- `frontend/src/oss/views/alertEvents/detail/CurrentEventDetail.tsx`
- `frontend/src/oss/views/alertEvents/detail/Events.tsx`
- `frontend/src/oss/views/alertEvents/detail/SilentAlert.tsx`
- `frontend/src/oss/views/alertEvents/components/Filter/index.tsx`
- `frontend/src/pro/api/report.ts`
- `backend/pkg/router/router_api.go`
- `backend/pkg/api/alerts/handler.go`
- `backend/pkg/services/alerts/service.go`
- `backend/pkg/services/alerts/service_alert_event_list.go`
- `backend/pkg/services/alerts/service_alert_event_detail.go`
- `backend/pkg/services/alerts/service_amconfig_receiver.go`
- `backend/pkg/services/alerts/service_add_alert_rule.go`
- `backend/pkg/services/alerts/service_update_alert_rule.go`
- `backend/pkg/services/alerts/service_alert_slient_config.go`
- `backend/pkg/model/request/alert_input.go`
- `backend/pkg/model/request/alert_event.go`
- `backend/pkg/model/request/alert_slience.go`
- `backend/pkg/model/response/alert_event.go`
- `backend/pkg/repository/clickhouse/dao_alert_event_with_workflow.go`
- `backend/pkg/repository/clickhouse/dao_alert_event_detail.go`
- `backend/pkg/repository/clickhouse/dao_alert_event_filter.go`
- `backend/pkg/repository/clickhouse/dao_alert_event_report.go`
- `backend/pkg/repository/database/alert_manager.go`
- `backend/pkg/repository/database/dao_quick_alert_rule_metric.go`
- `backend/pkg/repository/database/init_menu_item.go`

## Dependencies

| Dependency Type | Dependency | Why It Matters |
|---|---|---|
| Frontend Shared | `AuthRouter` | 所有本模块 route 都被包装，但当前实际 route 权限检查逻辑被注释。 |
| Frontend Shared | `UserContext` + `AppSidebar` | 菜单可见性来自 `/api/permission/config`。 |
| Frontend Shared | `dataGroupReducer` | 规则列表和事件列表都依赖 `dataGroupId`。 |
| Frontend Shared | `timeRangeReducer` | 事件列表、统计和过滤值依赖 `startTime`、`endTime`。 |
| Frontend Shared | `groupLabelReducer` | 规则 group 下拉和编辑回填依赖应用启动时加载的 group label。 |
| Frontend Shared | `WorkflowsIframe` | 事件列表和详情直接嵌入 Dify workflow 页面。 |
| Backend Shared | `LicMiddleware` | `/api/alerts` group 整体受 license 控制。 |
| Backend Shared | `AuthMiddleware` | 只覆盖中途之后注册的 `/api/alerts` 子路由。 |
| Repository | `clickhouse.Repo` | 告警事件、workflow record、通知记录、报告的主要读写源。 |
| Repository | `kubernetes.Repo` | 告警规则和外部 AlertManager config 的配置读写源。 |
| Repository | `database.Repo` | DingTalk receiver、quick metrics、target tags 的存储源。 |
| Repository | `dify.DifyRepo` | workflow id、cache minutes 和分类 workflow 调用。 |
| Cross Module | `alert-sources-and-enrichment` | `/api/alerts/inputs/alertmanager` 当前实际转到 `alertinput.ProcessAlertEvents`。 |
| Cross Module | `conversational-diagnosis-and-reporting` | 报告页读取本模块的 `/api/alerts/events/report/get`。 |

## Risks And Unknowns

1. `/api/alerts` route group 在 `AuthMiddleware` 前后分段注册，事件列表、详情、报告和输入输出接口只观察到 `LicMiddleware`，权限边界需要重点确认。
2. 前端过滤组件传 `groupId`，但后端 `SearchAlertEventFilterValuesRequest` 当前未定义 `GroupID`，事件过滤值可能没有按数据组收敛。
3. `slient` / `Slience` 是当前代码契约，但存在拼写风险，后续不能在文档或代码生成中擅自改成 `silent`。
4. `getAlertRulesStatusApi` 使用 `/alert/api/v1/rules`，当前主后端未见该 route，规则状态来源需要按外部边界记录。
5. `CheckAlertRule` 返回字段名是 `available`，但 service 直接把 `CheckAlertRule` 的 `find` 结果赋给 `Available`，若 `find` 表示“已存在”，语义可能反向。
6. `frontend/src/oss/views/alerts/index.js` 存在旧 tabs 入口，但当前路由未使用，可能是遗留入口。
