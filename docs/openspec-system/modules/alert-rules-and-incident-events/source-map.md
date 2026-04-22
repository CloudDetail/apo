# Alert Rules And Incident Events Source Map

## Scope
本文档先做证据提取，不直接扩写业务结论。目标是把 `alert-rules-and-incident-events` 模块在当前仓库中的真实代码落点固定下来，作为后续模块文档的依据。

## Relevant Source Directories

| Layer | Path | Notes |
|---|---|---|
| Frontend | `frontend/src/oss/views/alerts` | `/alerts/rule` 与 `/alerts/notify` 页面主目录，覆盖告警规则和告警通知 receiver 管理。 |
| Frontend | `frontend/src/oss/views/alertEvents` | `/alerts/events` 与 `/alerts/events/detail/:alertId/:eventId` 页面主目录，覆盖告警事件列表、详情、过滤、静默和工作流入口。 |
| Frontend | `frontend/src/core/api/alerts.js` | 告警规则、通知、事件、过滤、静默和手动 resolve 的主要 API wrapper。 |
| Frontend | `frontend/src/pro/api/report.ts` | 报告页读取 `/api/alerts/events/report/get`，属于本模块后端对报告模块的输出边界。 |
| Frontend | `frontend/src/pro/views/report` | 告警事件列表会跳转到 `/#/report?alertEventId=...`，报告页再消费告警报告 API。 |
| Frontend | `frontend/src/App.js` | 应用启动后调用 `getRuleGroupLabelApi`，把告警规则 group label 写入 redux。 |
| Frontend | `frontend/src/core/store/reducers/groupLabelReducer.js` | 保存 `groupLabel` 与 `groupLabelSelectOptions`，供告警规则页和弹窗使用。 |
| Frontend | `frontend/src/core/store/reducers/timeRangeReducer.js` | 告警事件列表和过滤条件依赖全局时间范围。 |
| Frontend | `frontend/src/core/store/reducers/dataGroupReducer.js` | 告警规则列表、事件列表和过滤组件依赖 `dataGroupId`。 |
| Backend | `backend/pkg/router/router_api.go` | `/api/alerts` route group 注册入口，同时包含本模块和 `alertanalyze` 的部分接口。 |
| Backend | `backend/pkg/api/alerts` | 本模块主要 handler 包。 |
| Backend | `backend/pkg/services/alerts` | 本模块主要 service 包。 |
| Backend | `backend/pkg/model/request/alert_input.go` | 告警规则、AlertManager receiver、AlertManager 输入请求结构。 |
| Backend | `backend/pkg/model/request/alert_event.go` | 告警事件列表、详情、报告、过滤、手动 resolve 请求结构。 |
| Backend | `backend/pkg/model/request/alert_slience.go` | 告警静默相关请求结构。 |
| Backend | `backend/pkg/model/response/alert_input.go` | 告警规则、通知 receiver、group label、quick metrics 响应结构。 |
| Backend | `backend/pkg/model/response/alert_event.go` | 告警事件列表、详情、分类和报告响应结构。 |
| Backend | `backend/pkg/model/response/alert_slience.go` | 告警静默响应结构。 |
| Backend | `backend/pkg/model/integration/alert` | 新告警事件、事件聚合、workflow 参数、alert source 相关模型。 |
| Backend | `backend/pkg/model/amconfig` | AlertManager receiver、notifier config 与 `DingTalkConfig`。 |
| Backend | `backend/pkg/model/amconfig/slienceconfig` | `AlertSlienceConfig` 与 `alert_slients` 表映射。 |
| Backend | `backend/pkg/repository/clickhouse` | `alert_event`、`workflow_records`、`alert_notify_record`、`alert_report` 的查询与写入。 |
| Backend | `backend/pkg/repository/kubernetes` | 告警规则文件与 AlertManager config 文件的读写。 |
| Backend | `backend/pkg/repository/database` | `ding_talk_config`、`quick_alert_rule_metric`、菜单/权限初始化和集成标签查询。 |
| Backend | `backend/pkg/receiver` | 内置 AlertManager receiver 与静默能力的后端依赖。 |

## Route Files

| Side | File | Observed Routes |
|---|---|---|
| Frontend | `frontend/src/oss/routes.js` | `/alerts/rule`, `/alerts/notify`, `/alerts/events`, `/alerts/events/detail/:alertId/:eventId` |
| Backend | `backend/pkg/router/router_api.go` | `/api/alerts/event/list`, `/api/alerts/event/detail`, `/api/alerts/events/classify`, `/api/alerts/events/report/get`, `/api/alerts/events/report/add`, `/api/alerts/inputs/alertmanager`, `/api/alerts/outputs/dingtalk/:uuid`, `/api/alerts/rules/file`, `/api/alerts/rule/groups`, `/api/alerts/rule/metrics`, `/api/alerts/rule/list`, `/api/alerts/rule`, `/api/alerts/rule/add`, `/api/alerts/rule/available`, `/api/alerts/alertmanager/receiver/list`, `/api/alerts/alertmanager/receiver/add`, `/api/alerts/alertmanager/receiver`, `/api/alerts/slient`, `/api/alerts/slient/list`, `/api/alerts/resolve`, `/api/alerts/filter/keys`, `/api/alerts/filter/labelkeys`, `/api/alerts/filter/values` |
| External Or Adjacent | `frontend/src/core/api/alerts.js` | `/alert/api/v1/rules`, `/alert/-/reload`，当前更像外部 AlertManager 或 VMAlert 边界，不是主后端 `/api/alerts`。 |
| Cross Module | `frontend/src/oss/views/alertEvents/index.tsx` | 打开 `/#/report?alertEventId=...`，由报告页面消费 `/api/alerts/events/report/get`。 |

## Page Files

| Type | File | Notes |
|---|---|---|
| Route Page | `frontend/src/oss/views/alerts/AlertsRule.jsx` | `/alerts/rule` 页面，负责告警规则列表、搜索、新增、编辑、删除和状态展示。 |
| Route Page | `frontend/src/oss/views/alerts/AlertsNotify.jsx` | `/alerts/notify` 页面，负责 AlertManager receiver 列表、搜索、新增、编辑、删除。 |
| Legacy Or Composite Page | `frontend/src/oss/views/alerts/index.js` | 旧的 tabs 容器存在，但当前 `ossRoutes` 直接路由到 `AlertsRule` 和 `AlertsNotify`。 |
| Modal | `frontend/src/oss/views/alerts/modal/ModifyAlertRuleModal.jsx` | 告警规则新增/编辑弹窗，组装 `alertRule` 并调用新增或更新接口。 |
| Modal Component | `frontend/src/oss/views/alerts/modal/ALertConditionCom.jsx` | PromQL 编辑器和 quick metrics 下拉，调用 `/api/alerts/rule/metrics`。 |
| Modal | `frontend/src/oss/views/alerts/modal/ModifyAlertNotifyModal.jsx` | 通知 receiver 新增/编辑弹窗，组装 email、webhook、dingtalk、wechat config。 |
| Notification Form | `frontend/src/oss/views/alerts/modal/compontents/*` | 各类通知渠道表单子组件。 |
| Route Page | `frontend/src/oss/views/alertEvents/index.tsx` | `/alerts/events` 页面，负责事件列表、统计、过滤、Dify workflow 入口和报告跳转。 |
| Route Page | `frontend/src/oss/views/alertEvents/detail/index.tsx` | `/alerts/events/detail/:alertId/:eventId` 页面，负责当前事件详情与同一 alert 的历史事件。 |
| Detail Component | `frontend/src/oss/views/alertEvents/detail/CurrentEventDetail.tsx` | 当前告警事件详情、workflow 结果和 run-once 入口。 |
| Detail Component | `frontend/src/oss/views/alertEvents/detail/Events.tsx` | 同一 `alertId` 下的历史事件表格。 |
| Detail Component | `frontend/src/oss/views/alertEvents/detail/SilentAlert.tsx` | 告警静默读取与保存。 |
| Shared Component | `frontend/src/oss/views/alertEvents/components/AlertInfoCom.tsx` | 告警级别、状态、详情标签、有效性状态等展示组件。 |
| Filter Component | `frontend/src/oss/views/alertEvents/components/Filter/*` | 事件过滤器、静态字段、动态 label 字段、值选择和表达式匹配。 |
| Chart Component | `frontend/src/oss/views/alertEvents/PieChart.tsx` | 告警 firing/resolved 统计图。 |

## API Wrapper Files

| File | Wrapper | API Path | Notes |
|---|---|---|---|
| `frontend/src/core/api/alerts.js` | `getAlertRulesApi` | `api/alerts/rule/list` | 告警规则分页列表。 |
| `frontend/src/core/api/alerts.js` | `getAlertRulesStatusApi` | `/alert/api/v1/rules` | 前端用于补充规则 firing/pending/inactive 状态，当前不是主后端 `/api/alerts`。 |
| `frontend/src/core/api/alerts.js` | `reloadAlertRulesApi` | `/alert/-/reload` | wrapper 存在，当前未见主要页面调用。 |
| `frontend/src/core/api/alerts.js` | `updateRuleApi` | `api/alerts/rule` | 更新告警规则。 |
| `frontend/src/core/api/alerts.js` | `addRuleApi` | `/api/alerts/rule/add` | 新增告警规则。 |
| `frontend/src/core/api/alerts.js` | `deleteRuleApi` | `api/alerts/rule` | 删除告警规则。 |
| `frontend/src/core/api/alerts.js` | `getRuleGroupLabelApi` | `api/alerts/rule/groups` | 应用启动时加载规则 group label。 |
| `frontend/src/core/api/alerts.js` | `getRuleMetricsApi` | `api/alerts/rule/metrics` | PromQL quick metrics。 |
| `frontend/src/core/api/alerts.js` | `getAlertmanagerListApi` | `/api/alerts/alertmanager/receiver/list` | 通知 receiver 列表。 |
| `frontend/src/core/api/alerts.js` | `addAlertNotifyApi` | `/api/alerts/alertmanager/receiver/add` | 新增通知 receiver。 |
| `frontend/src/core/api/alerts.js` | `updateAlertNotifyApi` | `/api/alerts/alertmanager/receiver` | 更新通知 receiver。 |
| `frontend/src/core/api/alerts.js` | `deleteAlertNotifyApi` | `/api/alerts/alertmanager/receiver` | 删除通知 receiver。 |
| `frontend/src/core/api/alerts.js` | `getAlertEventsApi` | `/api/alerts/event/list` | 告警事件列表。 |
| `frontend/src/core/api/alerts.js` | `getAlertEventDetailApi` | `/api/alerts/event/detail` | 告警事件详情和历史。 |
| `frontend/src/core/api/alerts.js` | `getAlertWorkflowIdApi` | `/api/alerts/events/classify` | 根据 `alertGroup` 与 `alertName` 获取 workflow id。 |
| `frontend/src/core/api/alerts.js` | `getAlertSilentConfigApi` | `/api/alerts/slient` | 读取告警静默配置。 |
| `frontend/src/core/api/alerts.js` | `deleteAlertSilentConfigApi` | `/api/alerts/slient` | 删除告警静默配置，当前主 detail 页面未见调用。 |
| `frontend/src/core/api/alerts.js` | `saveAlertSilentConfigApi` | `/api/alerts/slient` | 保存告警静默配置。 |
| `frontend/src/core/api/alerts.js` | `resolveAlertApi` | `/api/alerts/resolve` | 手动 resolve 最新告警事件。 |
| `frontend/src/core/api/alerts.js` | `getAlertsFilterKeysApi` | `/api/alerts/filter/keys` | 获取静态过滤项。 |
| `frontend/src/core/api/alerts.js` | `getAlertsFilterLabelKeysApi` | `/api/alerts/filter/labelkeys` | 获取动态 label 过滤字段。 |
| `frontend/src/core/api/alerts.js` | `getAlertsFilterValuesApi` | `/api/alerts/filter/values` | 获取某过滤字段的可选值。 |
| `frontend/src/pro/api/report.ts` | `getReportApi` | `api/alerts/events/report/get` | 报告页面按 `alertEventId` 或 `agentChatId` 读取报告。 |

## Controller Or Router Files

| Type | File | Notes |
|---|---|---|
| Router | `backend/pkg/router/router_api.go` | `/api/alerts` 统一使用 `LicMiddleware`；部分接口在 `AuthMiddleware` 前注册，后续接口在 `AuthMiddleware` 后注册。 |
| Handler Factory | `backend/pkg/api/alerts/handler.go` | 创建 `alerts.Service`、`alertinput.Service` 和 `serviceoverview.Service`。 |
| Event Handler | `backend/pkg/api/alerts/func_alerteventlist.go` | `/api/alerts/event/list`。 |
| Event Handler | `backend/pkg/api/alerts/func_alerteventdetail.go` | `/api/alerts/event/detail`。 |
| Event Handler | `backend/pkg/api/alerts/func_alerteventclassify.go` | `/api/alerts/events/classify`。 |
| Event Handler | `backend/pkg/api/alerts/func_getalerteventreport.go` | `/api/alerts/events/report/get`。 |
| Event Handler | `backend/pkg/api/alerts/func_addalerteventreport.go` | `/api/alerts/events/report/add`。 |
| Rule Handler | `backend/pkg/api/alerts/func_getalertrules.go` | `/api/alerts/rule/list`。 |
| Rule Handler | `backend/pkg/api/alerts/func_addalertrule.go`, `func_updatealertrule.go`, `func_deletealertrule.go`, `func_checkalertrule.go` | 告警规则 CRUD 与可用性检查。 |
| Rule Config Handler | `backend/pkg/api/alerts/func_getalertrulefile.go`, `func_updatealertrulefile.go` | 规则文件级别读写，当前主前端页面未见直接消费。 |
| Option Handler | `backend/pkg/api/alerts/func_getgrouplist.go`, `func_getmetricpql.go` | group label 与 quick metrics。 |
| Receiver Handler | `backend/pkg/api/alerts/func_getalertmanagerconfigreceiver.go`, `func_addalertmanagerconfigreceiver.go`, `func_updatealertmanagerconfigreceiver.go`, `func_deletealertmanagerconfigreceiver.go` | AlertManager receiver CRUD。 |
| Silence Handler | `backend/pkg/api/alerts/func_getalertslientconfig.go`, `func_listalertslientconfig.go`, `func_setalertslientconfig.go`, `func_removealertslientconfig.go` | 告警静默配置。 |
| Filter Handler | `backend/pkg/api/alerts/func_getalerteventstaticfilters.go`, `func_getalerteventlabelfilterkeys.go`, `func_searchalerteventfiltervalues.go` | 事件过滤元数据与值。 |
| Integration Handler | `backend/pkg/api/alerts/func_inputalertmanager.go` | `/api/alerts/inputs/alertmanager`，当前实现转到 `alertinput.ProcessAlertEvents`。 |
| Output Handler | `backend/pkg/api/alerts/func_forwardtodingtalk.go` | `/api/alerts/outputs/dingtalk/:uuid`，接收 AlertManager webhook 后转发 DingTalk。 |

## Service Files

| Service Area | Files | Notes |
|---|---|---|
| Event List | `backend/pkg/services/alerts/service_alert_event_list.go` | 从 ClickHouse 查询事件与 workflow record，补齐 `WorkflowParams`、`validity`、`counts`、tag display。 |
| Event Detail | `backend/pkg/services/alerts/service_alert_event_detail.go` | 查询当前事件、同 `alertId` 历史事件、定位分页和 workflow 参数。 |
| Event Filters | `backend/pkg/services/alerts/service_alert_event_filters.go` | 转发到 ClickHouse 获取静态过滤项、label keys 和 values。 |
| Event Classify | `backend/pkg/services/alerts/service_alert_event_classify.go` | 调用 Dify `AlertClassify` workflow，按 `alertGroup + alertName` 做一小时 LRU 缓存。 |
| Event Report | `backend/pkg/services/alerts/service_alert_event_report.go`, `service_add_alert_event_report.go` | 从 `alert_report` 读取或写入报告数据。 |
| Rule CRUD | `backend/pkg/services/alerts/service_get_alert_rule.go`, `service_add_alert_rule.go`, `service_update_alert_rule.go`, `service_check_alert_rule.go` | 通过 `k8sApi` 管理告警规则，并校验 group 与 label 映射。 |
| Rule Options | `backend/pkg/services/alerts/service_get_group_list.go`, `service_get_metric_pql.go` | group label 国际化与 quick metrics。 |
| Receiver CRUD | `backend/pkg/services/alerts/service_amconfig_receiver.go` | 根据 `AlertReceiver.Enabled` 在内置 receiver 与外部 AlertManager config 之间切换。 |
| Silence | `backend/pkg/services/alerts/service_alert_slient_config.go` | 仅 `enableInnerReceiver` 为真时通过 `receivers` 管理静默。 |
| DingTalk Forward | `backend/pkg/services/alerts/service_forward_to_dingtalk.go` | 从 `ding_talk_config` 读取 URL/Secret 并发送通知。 |
| Deprecated Input | `backend/pkg/services/alerts/inputs_alertmanager.go`, `alertmanager.go` | 保留 AlertManager 转换逻辑，但 `Service.InputAlertManager` 注释标记为 deprecated。 |
| Notification Builder | `backend/pkg/services/alerts/notification/*` | 构造 DingTalk 转发内容。 |

## Model Entity Schema Files

| Type | File | Notes |
|---|---|---|
| Request | `backend/pkg/model/request/alert_input.go` | 告警规则、AlertManager receiver、AlertManager 输入、检测规则请求。 |
| Request | `backend/pkg/model/request/alert_event.go` | 事件查询、详情、报告、过滤、手动 resolve 请求。 |
| Request | `backend/pkg/model/request/alert_slience.go` | 静默读取、保存、删除请求。 |
| Response | `backend/pkg/model/response/alert_input.go` | 规则列表、receiver 列表、group label、quick metrics 返回体。 |
| Response | `backend/pkg/model/response/alert_event.go` | 事件列表、详情、workflow id、过滤项、报告返回体。 |
| Response | `backend/pkg/model/response/alert_slience.go` | `slience` 与 `sliences` 返回字段。 |
| Domain Model | `backend/pkg/model/integration/alert/alert.go` | `Alert` 与 `alert_id` 聚合模型。 |
| Domain Model | `backend/pkg/model/integration/alert/alert_event.go` | `AlertEvent`、`FastAlertID`、`ToAMAlert`。 |
| Domain Model | `backend/pkg/model/integration/alert/alert_event_processed.go` | `AEventWithWRecord`、`WorkflowDetail`、`NotifyDetail`、`WorkflowParams`。 |
| Domain Model | `backend/pkg/model/alert_event.go` | 旧 `AlertEvent` 模型，注释指向 `alert_event` 表。 |
| Report Model | `backend/pkg/model/alert_report.go` | `AlertEventReport` 与报告结构。 |
| Receiver Model | `backend/pkg/model/amconfig/notifiers.go` | `Receiver` 子类型和 `DingTalkConfig`，其中 `DingTalkConfig.TableName()` 为 `ding_talk_config`。 |
| Silence Model | `backend/pkg/model/amconfig/slienceconfig/slience.go` | `AlertSlienceConfig.TableName()` 为 `alert_slients`。 |
| Tag Model | `backend/pkg/model/integration/alert/target_tag.go` | `TargetTag.TableName()` 为 `target_tags`。 |
| Quick Metrics Model | `backend/pkg/repository/database/dao_quick_alert_rule_metric.go` | `AlertMetricsData.TableName()` 为 `quick_alert_rule_metric`。 |

## Important Shared Dependencies

| Dependency | File | Notes |
|---|---|---|
| Route Guard | `frontend/src/core/components/AuthRouter.tsx` | 相关 route 都被 `AuthRouter` 包装，但实际 `checkRoutePermission` 的 `useEffect` 当前被注释。 |
| Menu Permission | `frontend/src/core/contexts/UserContext.js`, `frontend/src/core/components/AppSidebar.js` | 侧边栏来自 `/api/permission/config` 返回的 `menuItems`，不是纯前端硬编码。 |
| Menu Metadata | `backend/pkg/repository/database/init_menu_item.go`, `init_feature_mapping.go`, `init_router.go` | 初始化 `alertsRule`、`alertsNotify`、`alertEvents`、`alertEventDetail` 与对应 feature。 |
| Request Utility | `frontend/src/core/utils/request.js` | 所有 wrapper 共用请求、token、错误处理。 |
| Notification Utility | `frontend/src/core/utils/notify` | 规则和通知保存/删除成功后提示。 |
| Time Utility | `frontend/src/core/utils/time` | 事件时间展示统一转换。 |
| Data Group State | `frontend/src/core/store/reducers/dataGroupReducer.js` | 规则列表、事件列表和过滤组件读取 `dataGroupId`。 |
| Time Range State | `frontend/src/core/store/reducers/timeRangeReducer.js` | 事件列表和过滤值查询读取 `startTime`、`endTime`。 |
| Group Label State | `frontend/src/core/store/reducers/groupLabelReducer.js` | 规则 group 下拉和编辑回填依赖。 |
| Workflow Iframe | `frontend/src/oss/views/workflows/workflowsIframe` | 告警事件页面直接嵌入 `/dify/app/...`。 |
| Backend Middleware | `backend/pkg/middleware/lic_middleware.go`, `auth_middleware.go` | `/api/alerts` 先整体应用 license，再局部应用 auth。 |
| ClickHouse Repo | `backend/pkg/repository/clickhouse` | 告警事件、报告、workflow record、通知记录主要查询源。 |
| Kubernetes Repo | `backend/pkg/repository/kubernetes` | 管理告警规则和 AlertManager config。 |
| Database Repo | `backend/pkg/repository/database` | 管理 DingTalk receiver、quick metrics、target tags 等。 |
| Dify Repo | `backend/pkg/repository/dify` | 分类 workflow、alert check/analyze workflow id、cache minutes。 |
| Integration Alert Service | `backend/pkg/services/integration/alert` | `/api/alerts/inputs/alertmanager` 当前实际委托到 `ProcessAlertEvents`。 |

## Evidence Notes

1. `/api/alerts` route group 中，`event/list`、`event/detail`、`events/classify`、`events/report/*`、`inputs/alertmanager`、`outputs/dingtalk/:uuid` 注册在 `AuthMiddleware` 之前；规则、通知、静默、过滤和 resolve 注册在 `AuthMiddleware` 之后。
2. `frontend/src/core/api/alerts.js` 同时访问 `/api/alerts/*` 和 `/alert/*`，后者在当前主后端路由中未见注册，应作为外部 AlertManager/VMAlert 边界处理。
3. 前端过滤组件向 `/api/alerts/filter/values` 传递 `groupId`，但 `SearchAlertEventFilterValuesRequest` 当前未定义 `GroupID` 字段。
4. 静默相关接口和模型使用 `slient` / `Slience` 拼写，这是当前实现契约，文档中必须保持原样。
5. `frontend/src/oss/views/alerts/index.js` 是 tabs 容器，但当前 `frontend/src/oss/routes.js` 没有直接使用 `/alerts` 作为页面入口。
