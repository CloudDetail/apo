# Alert Rules And Incident Events Flows

## Main Flow 1
告警规则管理流程：

1. `App.js` 在非登录页启动时调用 `getRuleGroupLabelApi -> /api/alerts/rule/groups`，结果写入 `groupLabelReducer`。
2. 用户进入 `/alerts/rule`，`AlertsRule` 读取 `groupLabelSelectOptions` 和 `dataGroupId`。
3. 页面调用 `getAlertRulesApi -> /api/alerts/rule/list` 获取规则分页列表。
4. 页面调用 `getAlertRulesStatusApi -> /alert/api/v1/rules` 获取外部规则状态，并按 `group-alert` 构造状态 map。
5. 用户新增或编辑规则时，`ModifyAlertRuleModal` 将表单转成 `AlertRule`，调用 `/api/alerts/rule/add` 或 `/api/alerts/rule`。
6. 后端 `AddAlertRule` 或 `UpdateAlertRule` 调用 `checkOrFillGroupsLabel` 校验 group 与 label，再按 `groupId` 补充 `labels["groupId"]`。
7. 后端通过 `k8sApi.AddAlertRule` 或 `k8sApi.UpdateAlertRule` 写入规则配置。
8. 前端收到成功结果后提示保存成功并刷新规则列表。

## Main Flow 2
告警通知 receiver 管理流程：

1. 用户进入 `/alerts/notify`，`AlertsNotify` 调用 `/api/alerts/alertmanager/receiver/list` 获取 receiver 列表。
2. 用户新增或编辑 receiver 时，`ModifyAlertNotifyModal` 根据 `notifyType` 构造 `amConfigReceiver`。
3. email 类型会组装 `emailConfigs`，并把 `smarthost` 与 `smarthostPort` 拼成 `smarthost`。
4. webhook 类型会组装 `webhookConfigs`，支持 basic auth、bearer token 和 header。
5. dingtalk 类型会组装 `dingTalkConfigs`，包含 `url` 与 `secret`。
6. wechat 类型会组装 `wechatConfigs`，并过滤没有 `toUser`、`toParty`、`toTag` 的配置。
7. 后端 `service_amconfig_receiver.go` 根据 `AlertReceiver.Enabled` 选择内置 receiver 或外部 AlertManager config；外部 dingtalk 会额外写入 `ding_talk_config`。
8. 前端收到成功结果后关闭弹窗并刷新 receiver 列表。

## Main Flow 3
告警事件处置流程：

1. 用户进入 `/alerts/events`，页面读取 `timeRange`、`dataGroupId` 和当前过滤器。
2. 页面调用 `/api/alerts/event/list`，请求包含 `startTime`、`endTime`、`pagination`、`filters`、`groupId`。
3. 后端按 `groupId` 计算 `SubGroupIDs`，从 ClickHouse 查询 `alert_event` 和 `workflow_records` 聚合结果。
4. 后端为每条事件补充 `WorkflowParams`、`duration`、`validity`、`tagsDisplay`，并返回 `counts`、`alertCheckId`。
5. 前端展示统计卡片、饼图、过滤器和事件表格。
6. 用户点击 `cause` 时，前端调用 `/api/alerts/events/classify` 获取 workflow id，再打开 `/dify/app/{workflowId}/run-once?...`。
7. 用户点击有效性结果时，前端使用 `alertCheckId` 和 `workflowRunId` 打开 `/dify/app/{alertCheckId}/logs/{workflowRunId}`。
8. 用户点击详情时，前端进入 `/alerts/events/detail/:alertId/:eventId` 并调用 `/api/alerts/event/detail`。
9. 后端返回当前事件和同一 `alertId` 的历史事件，前端展示详情、历史、通知状态、静默入口和定位入口。

## Main Flow 4
报告查看流程：

1. 在 `/alerts/events` 中，如果某条事件有 `alertDirection`，前端渲染 `viewReport` 入口。
2. 用户点击后，页面打开 `/#/report?alertEventId={event.id}&startTime={startTime}&endTime={endTime}`。
3. `frontend/src/pro/views/report/index.tsx` 读取 query，构造 `tags = { alertEventId }`。
4. 报告页调用 `getReportApi -> /api/alerts/events/report/get`。
5. 后端在 `alert_report` 中按扩展后的时间范围和 tags 查找报告，并转换成报告页面需要的 `reportData`。

## Edge Cases

| Case | Implemented Behavior | To Be Confirmed |
|---|---|---|
| 事件列表当前页超过总页数 | `AlertEventsPage` 计算 `totalPages`，若当前页过大则把 `pageIndex` 调整到最后一页。 | 无。 |
| 已恢复事件的 workflow 参数时间 | 后端对 resolved 事件使用 `EndTime - 15m` 到 `EndTime`。 | 15 分钟窗口是否为业务约定需要确认。 |
| 未恢复事件的 workflow 参数时间 | 后端优先使用 `LastCheckAt - 15m` 到 `LastCheckAt`，否则使用 `UpdateTime - 15m` 到 `UpdateTime`。 | `validity` 与 `LastCheckAt` 的优先级是否稳定需要确认。 |
| 详情请求未带时间范围 | 后端使用当前事件 `UpdateTime - 2h` 到当前时间。 | 该默认窗口是否面向所有告警类型。 |
| 静默能力关闭 | `enableInnerReceiver == false` 时静默 service 返回 `inner alert is not open`。 | 前端没有按该能力开关隐藏 `SilentAlert`。 |
| dingtalk 外部 receiver | 外部模式下后端先创建内部 webhook，再保存 DingTalk URL/Secret 到 DB。 | 失败回滚只删除 k8s receiver，DB 写入前后的事务边界需要确认。 |
| 过滤动态字段 | 后端只允许匹配 `allowFilterKey` 的字段。 | 前端传入的 `groupId` 当前未参与过滤值查询。 |

## Failure Paths

| Failure Path | Evidence | Result |
|---|---|---|
| 请求绑定失败 | `ShouldBindJSON` / `ShouldBindQuery` | handler 返回 `code.ParamBindError`。 |
| 规则 group 与 label 不匹配 | `checkOrFillGroupsLabel` | 返回 `AlertGroupAndLabelMismatchError`。 |
| Dify 分类 workflow 失败 | `AlertEventClassify` | 返回 `workflow run failed` 或 Dify 调用错误，前端 workflow modal 可能没有 URL。 |
| 报告重复新增 | `AddAlertEventReport` | 返回 `AddAlertEventReportError`。 |
| dingtalk receiver 内容为空 | `AddAMReceiversForExternalAM` / `UpdateAMReceiverForExternalAM` | 返回 `AlertManagerEmptyReceiver`。 |
| 静默能力不可用 | `service_alert_slient_config.go` | 返回普通错误字符串 `inner alert is not open`。 |
| 手动 resolve 失败 | `ManualResolveLatestAlertEventByAlertID` | handler 返回 `code.GetAlertSlienceError`，错误码命名与 resolve 语义不完全匹配。 |
| 外部 `/alert/api/v1/rules` 失败 | `AlertsRule.loadAlertStates` | 前端 catch 后只输出 console，规则列表仍可能展示 `unknown` 状态。 |

## Permission Constraints

| Constraint | Evidence | Notes |
|---|---|---|
| Frontend route guard | `AuthRouter` | 包装存在，但实际权限检查被注释，默认允许。 |
| Menu visibility | `UserContext` + `AppSidebar` | 用户能否看到菜单由后端权限配置决定。 |
| License gating | `router_api.go` | `/api/alerts` 整体应用 `LicMiddleware`。 |
| Auth gating | `router_api.go` | 规则、receiver、静默、resolve、filter 接口在 `AuthMiddleware` 后注册。 |
| License-only API | `router_api.go` | 事件列表、详情、报告、AlertManager 输入和 DingTalk 输出在 `AuthMiddleware` 前注册。 |
| Data group UI | `ossRoutes` | `/alerts/rule` 和 `/alerts/events` 显示数据组选择。 |
| Data group backend | `AlertEventList` | 事件列表用 `GroupID` 展开 `SubGroupIDs`。 |
| Filter data group | `SearchAlertEventFilterValuesRequest` | 前端传 `groupId`，后端结构当前没有对应字段。 |
