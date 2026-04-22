# Alert Rules And Incident Events Backend

## API Summary

### License Only APIs

以下接口在 `backend/pkg/router/router_api.go` 中注册在 `alertApi.Use(middlewares.AuthMiddleware())` 之前。它们仍受 `/api/alerts` group 的 `LicMiddleware` 控制，但当前代码证据未显示它们受 `AuthMiddleware` 控制。

| API Path | Handler | Service | Purpose |
|---|---|---|---|
| `/api/alerts/event/list` | `alerts.AlertEventList` | `alerts.AlertEventList` | 查询告警事件列表、统计、workflow id 和有效性信息。 |
| `/api/alerts/event/detail` | `alerts.AlertEventDetail` | `alerts.AlertDetail` | 查询当前事件与同一 `alertId` 的历史事件。 |
| `/api/alerts/events/classify` | `alerts.AlertEventClassify` | `alerts.AlertEventClassify` | 调用 Dify 分类 workflow，返回具体 workflow id。 |
| `/api/alerts/events/report/get` | `alerts.GetAlertEventReport` | `alerts.GetAlertEventReport` | 按时间范围和 tags 查询 `alert_report`。 |
| `/api/alerts/events/report/add` | `alerts.AddAlertEventReport` | `alerts.AddAlertEventReport` | 写入 `alert_report`。 |
| `/api/alerts/inputs/alertmanager` | `alerts.InputAlertManager` | `alertinput.ProcessAlertEvents` | 接收 AlertManager 事件输入，当前实际委托到告警接入服务。 |
| `/api/alerts/outputs/dingtalk/:uuid` | `alerts.ForwardToDingTalk` | `alerts.ForwardToDingTalk` | 将 AlertManager webhook 转发到 DingTalk。 |

### Auth Protected APIs

以下接口注册在 `alertApi.Use(middlewares.AuthMiddleware())` 之后，当前代码证据显示它们受 `LicMiddleware` 与 `AuthMiddleware` 双重控制。

| API Path | Handler | Service | Purpose |
|---|---|---|---|
| `/api/alerts/rules/file` GET | `alerts.GetAlertRuleFile` | `alerts.GetAlertRuleFile` | 读取规则文件内容。 |
| `/api/alerts/rules/file` POST | `alerts.UpdateAlertRuleFile` | `alerts.UpdateAlertRuleFile` | 更新规则文件内容。 |
| `/api/alerts/rule/groups` | `alerts.GetGroupList` | `alerts.GetGroupList` | 返回 group label 映射。 |
| `/api/alerts/rule/metrics` | `alerts.GetMetricPQL` | `alerts.GetMetricPQL` | 返回 quick alert rule metrics。 |
| `/api/alerts/rule/list` | `alerts.GetAlertRules` | `alerts.GetAlertRules` | 告警规则分页列表。 |
| `/api/alerts/rule` POST | `alerts.UpdateAlertRule` | `alerts.UpdateAlertRule` | 更新告警规则。 |
| `/api/alerts/rule` DELETE | `alerts.DeleteAlertRule` | `alerts.DeleteAlertRule` | 删除告警规则。 |
| `/api/alerts/rule/add` | `alerts.AddAlertRule` | `alerts.AddAlertRule` | 新增告警规则。 |
| `/api/alerts/rule/available` | `alerts.CheckAlertRule` | `alerts.CheckAlertRule` | 检查告警规则可用性。 |
| `/api/alerts/alertmanager/receiver/list` | `alerts.GetAlertManagerConfigReceiver` | `alerts.GetAMConfigReceivers` | 通知 receiver 分页列表。 |
| `/api/alerts/alertmanager/receiver/add` | `alerts.AddAlertManagerConfigReceiver` | `alerts.AddAMConfigReceiver` | 新增通知 receiver。 |
| `/api/alerts/alertmanager/receiver` POST | `alerts.UpdateAlertManagerConfigReceiver` | `alerts.UpdateAMConfigReceiver` | 更新通知 receiver。 |
| `/api/alerts/alertmanager/receiver` DELETE | `alerts.DeleteAlertManagerConfigReceiver` | `alerts.DeleteAMConfigReceiver` | 删除通知 receiver。 |
| `/api/alerts/slient` GET | `alerts.GetAlertSlienceConfig` | `alerts.GetSlienceConfigByAlertID` | 读取静默配置。 |
| `/api/alerts/slient/list` | `alerts.ListAlertSlienceConfig` | `alerts.ListSlienceConfig` | 静默配置列表。 |
| `/api/alerts/slient` DELETE | `alerts.RemoveAlertSlienceConfig` | `alerts.RemoveSlienceConfigByAlertID` | 删除静默配置。 |
| `/api/alerts/slient` POST | `alerts.SetAlertSlienceConfig` | `alerts.SetSlienceConfigByAlertID` | 保存静默配置。 |
| `/api/alerts/resolve` | `alerts.MarkAlertResolvedManually` | `alerts.ManualResolveLatestAlertEventByAlertID` | 手动 resolve 最新告警事件。 |
| `/api/alerts/filter/keys` | `alerts.GetAlertEventStaticFilters` | `alerts.GetStaticFilterKeys` | 返回静态过滤项。 |
| `/api/alerts/filter/labelkeys` | `alerts.GetAlertEventLabelFilterKeys` | `alerts.GetAlertEventFilterLabelKeys` | 返回动态 label 过滤字段。 |
| `/api/alerts/filter/values` | `alerts.SearchAlertEventFilterValues` | `alerts.GetAlertEventFilterValues` | 返回某过滤字段的可选值。 |

## Controllers Or Routers

| File | Role |
|---|---|
| `backend/pkg/router/router_api.go` | 注册 `/api/alerts`，先应用 `LicMiddleware`，再在部分 route 注册后追加 `AuthMiddleware`。 |
| `backend/pkg/api/alerts/handler.go` | handler 工厂，初始化 `alerts.Service`、`alertinput.Service`、`serviceoverview.Service`。 |
| `backend/pkg/api/alerts/func_*.go` | 每个接口一个 handler 文件，统一执行请求绑定、service 调用和 `c.Payload` / `c.AbortWithError`。 |

当前 handler 模式比较稳定：

1. JSON body 接口使用 `ShouldBindJSON`。
2. GET query 接口使用 `ShouldBindQuery`。
3. 成功使用 `c.Payload(...)`。
4. 参数错误使用 `code.ParamBindError`。
5. 业务错误按场景使用 `code.GetAlertEventListError`、`code.UpdateAMConfigReceiverError`、`code.SetAlertSlienceError` 等。

## Services

| Service Area | Files | Behavior |
|---|---|---|
| Event List | `service_alert_event_list.go` | 根据 `groupId` 计算子数据组，调用 `GetAlertEventWithWorkflowRecord` 与 `GetAlertEventCounts`，补充 `WorkflowParams`、`duration`、`validity` 和 tag display。 |
| Event Detail | `service_alert_event_detail.go` | 查询当前事件；若请求未给时间范围，则以事件 `UpdateTime` 前两小时到当前时间作为历史范围；查询同一 `alertId` 的历史事件并支持定位分页。 |
| Event Classify | `service_alert_event_classify.go` | 调用 Dify `AlertClassify` workflow，缓存 `alertGroup + alertName` 到 `workflowId` 的结果。 |
| Event Report | `service_alert_event_report.go`, `service_add_alert_event_report.go` | 查询和写入 `alert_report`；新增前先按 `Overview.TimeStamp ± 1h` 与 tags 检查重复。 |
| Event Filters | `service_alert_event_filters.go` | 从 ClickHouse 获取静态过滤项、动态 label keys 和 values。 |
| Rule List | `service_get_alert_rule.go` | 从 `k8sApi.GetAlertRules` 读取规则并返回分页。 |
| Rule CRUD | `service_add_alert_rule.go`, `service_update_alert_rule.go`, `service_check_alert_rule.go` | 新增/更新前校验 group 与 label 映射，按需把 `groupId` 写入 `labels["groupId"]`。 |
| Rule Options | `service_get_group_list.go`, `service_get_metric_pql.go` | 按语言返回 group label，并从 `quick_alert_rule_metric` 返回 PromQL 快捷项。 |
| Receiver CRUD | `service_amconfig_receiver.go` | 根据 `AlertReceiver.Enabled` 选择内置 receiver 或外部 AlertManager config；外部模式下 dingtalk receiver 额外持久化到 `ding_talk_config`。 |
| Silence | `service_alert_slient_config.go` | 仅在 `enableInnerReceiver` 为真时可读写静默，否则返回 `inner alert is not open`。 |
| DingTalk Output | `service_forward_to_dingtalk.go` | 通过 uuid 查询 `ding_talk_config`，构造通知并调用 DingTalk webhook。 |
| Deprecated Input | `inputs_alertmanager.go` | `Service.InputAlertManager` 注释已标记 deprecated，handler 当前没有直接使用该 service 方法。 |

## Models And Schemas

| Type | File | Notes |
|---|---|---|
| `GetAlertRuleRequest` | `backend/pkg/model/request/alert_input.go` | 规则列表请求，包含 `AlertRuleFilter` 和 `PageParam`。 |
| `AlertRule` | `backend/pkg/model/request/alert_input.go` | 规则配置结构。 |
| `UpdateAlertManagerConfigReceiver` | `backend/pkg/model/request/alert_input.go` | receiver 更新请求。 |
| `AlertEventSearchRequest` | `backend/pkg/model/request/alert_event.go` | 事件列表请求，包含时间范围、排序、分页、filters、`GroupID`、`SubGroupIDs`。 |
| `GetAlertDetailRequest` | `backend/pkg/model/request/alert_event.go` | 详情请求，包含 `AlertID`、`EventID`、时间范围、分页和 `LocateEvent`。 |
| `SearchAlertEventFilterValuesRequest` | `backend/pkg/model/request/alert_event.go` | 过滤值请求，当前没有 `GroupID`。 |
| `SetAlertSlienceConfigRequest` | `backend/pkg/model/request/alert_slience.go` | 静默保存请求。 |
| `AlertEventSearchResponse` | `backend/pkg/model/response/alert_event.go` | 返回 `events`、`pagination`、`alertEventAnalyzeWorkflowId`、`alertCheckId`、`counts`。 |
| `GetAlertDetailResponse` | `backend/pkg/model/response/alert_event.go` | 返回当前事件、历史事件、分页、workflow id 和 `locateIndex`。 |
| `AEventWithWRecord` | `backend/pkg/model/integration/alert/alert_event_processed.go` | 告警事件与 workflow/notify 信息的聚合结构。 |
| `AlertEventReport` | `backend/pkg/model/alert_report.go` | 报告持久化模型。 |
| `DingTalkConfig` | `backend/pkg/model/amconfig/notifiers.go` | `ding_talk_config` 表模型。 |
| `AlertSlienceConfig` | `backend/pkg/model/amconfig/slienceconfig/slience.go` | `alert_slients` 表模型。 |

## Database Mapping

| Storage | Evidence | Notes |
|---|---|---|
| ClickHouse `alert_event` | `dao_alert_event_with_workflow.go`, `dao_alert_event_detail.go`, `dao_alert_event_filter.go` | 事件列表、详情、过滤均以 `alert_event` 为核心。 |
| ClickHouse `workflow_records` | `dao_alert_event_with_workflow.go`, `dao_alert_event_detail.go`, `dao_alert_event_filter.go` | 用于计算 `validity`、`workflowRunId`、`alertDirection` 等。 |
| ClickHouse `alert_notify_record` | `dao_alert_event_detail.go` | 详情历史事件中展示通知发送状态。 |
| ClickHouse `alert_report` | `dao_alert_event_report.go` | 报告读取和写入。 |
| Database `ding_talk_config` | `amconfig.DingTalkConfig.TableName()`, `repository/database/alert_manager.go` | 外部 AlertManager 模式下保存 DingTalk webhook URL、secret、uuid。 |
| Database `quick_alert_rule_metric` | `dao_quick_alert_rule_metric.go` | 规则编辑器 quick metrics。 |
| Database `target_tags` | `model/integration/alert/target_tag.go`, `dao_list_alert_target_tags.go` | 事件 tag display 国际化展示名。 |
| Database `alert_slients` | `model/amconfig/slienceconfig/slience.go` | 内置 receiver 模式下的静默配置表。 |
| Kubernetes ConfigMap | `repository/kubernetes/dao.go`, `dao_alert_rule.go`, `dao_alert_manager_config.go` | 默认规则文件 `alert-rules.yaml` 和 AlertManager config 文件 `alertmanager.yaml`。 |

## Validation And Rules

| Rule | Evidence | Notes |
|---|---|---|
| JSON Binding | 多数 `backend/pkg/api/alerts/func_*.go` | POST/DELETE handler 通常用 `ShouldBindJSON`。 |
| Query Binding | `func_getalertslientconfig.go` 等 | GET 接口用 `ShouldBindQuery`。 |
| Rule Required Fields | `AlertRule` | `group` 和 `alert` 带 `binding:"required"`。 |
| Group Label Consistency | `checkOrFillGroupsLabel` | 若 group 与 label 不匹配，返回 `AlertGroupAndLabelMismatchError`。 |
| Data Group Label Injection | `AddAlertRule`, `UpdateAlertRule` | `GroupID != 0` 时写入 `labels["groupId"]`。 |
| Receiver Mode Switch | `AlertReceiver.Enabled` | 内置模式使用 `receivers`，外部模式使用 `k8sApi`，dingtalk 特殊写入 DB。 |
| Silence Mode Restriction | `service_alert_slient_config.go` | 非内置 receiver 模式直接返回错误。 |
| Filter Key Safety | `dao_alert_event_filter.go` | 对 `labels.*`、`tags.*`、`workflow.*` 的动态字段做正则限制。 |
| Report Duplicate Check | `AddAlertEventReport` | 同一 tags 与时间窗口已有报告时返回错误。 |
| Manual Resolve Validation | `ManualResolveLatestAlertEventByAlertID` | 代码中有 `TODO valid alertID`，当前只转发到 ClickHouse。 |

## Error Handling

| Pattern | Evidence | Notes |
|---|---|---|
| Parameter Error | `c.AbortWithError(http.StatusBadRequest, code.ParamBindError, err)` | handler 通用绑定失败路径。 |
| Business Error | 各 handler 的 `code.*Error` | 例如 `GetAlertEventListError`、`SearchAlertFilterValueError`、`UpdateAMConfigReceiverError`。 |
| Frontend Catch | `AlertsRule`, `AlertsNotify`, `AlertEventsPage` | 多数请求失败后只关闭 loading 或输出 console。 |
| Dify Classify Failure | `service_alert_event_classify.go` | workflow 返回非 `succeeded` 时返回 `workflow run failed`。 |
| Empty Receiver | `service_amconfig_receiver.go` | dingtalk receiver 为空时返回 `AlertManagerEmptyReceiver`。 |
| Missing Inner Receiver | `service_alert_slient_config.go` | 非内置模式返回 `inner alert is not open`。 |

## Notes

1. `/api/alerts` route group 同时包含 `alertanalyze` 相关接口，这些接口应归入 `automated-diagnosis-and-workflows`，本模块只记录它们与 `/api/alerts` namespace 的共存事实。
2. `func_alerteventlist.go` 的 swagger 注释写的是 `/api/alerts/events/list`，实际路由注册是 `/api/alerts/event/list`。
3. `Service.InputAlertManager` 标记 deprecated，但 `func_inputalertmanager.go` 当前通过 `inputService.ProcessAlertEvents` 接收 AlertManager 数据。
4. `GetAlertEventReport` 查询时间范围扩大为 `startTime - 1h` 到 `endTime + 1h`，这是代码实现行为，但业务原因需要后续确认。
