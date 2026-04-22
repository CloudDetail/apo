# Alert Rules And Incident Events Frontend

## Pages

| Page | File | Description |
|---|---|---|
| `/alerts/rule` | `frontend/src/oss/views/alerts/AlertsRule.jsx` | 告警规则列表页，支持按 group 和 alert 名搜索、分页、新增、编辑、删除，并从外部规则状态接口补充状态 tag。 |
| `/alerts/notify` | `frontend/src/oss/views/alerts/AlertsNotify.jsx` | 告警通知 receiver 管理页，支持按名称搜索、分页、新增、编辑、删除。 |
| `/alerts/events` | `frontend/src/oss/views/alertEvents/index.tsx` | 告警事件列表页，展示 firing/resolved 统计、降噪相关统计、动态过滤、事件表格、workflow 入口和报告跳转。 |
| `/alerts/events/detail/:alertId/:eventId` | `frontend/src/oss/views/alertEvents/detail/index.tsx` | 告警事件详情页，展示当前事件、同一 `alertId` 历史事件、静默入口、定位按钮和手动 resolve。 |
| `/report` | `frontend/src/pro/views/report/index.tsx` | 不属于本模块主页面，但由告警事件列表通过 `alertEventId` 跳转，并调用本模块报告 API。 |

## Routes

| Route | Route File | Notes |
|---|---|---|
| `/alerts/rule` | `frontend/src/oss/routes.js` | `hideSystemTimeRangePicker: true`，`showDataGroup: true`。 |
| `/alerts/notify` | `frontend/src/oss/routes.js` | `hideSystemTimeRangePicker: true`，当前未启用 `showDataGroup`。 |
| `/alerts/events` | `frontend/src/oss/routes.js` | `showDataGroup: true`，使用全局时间范围。 |
| `/alerts/events/detail/:alertId/:eventId` | `frontend/src/oss/routes.js` | `hideSystemTimeRangePicker: true`，从 URL params 读取 `alertId`、`eventId`。 |

## Menus

| Menu Aspect | Evidence | Notes |
|---|---|---|
| Menu Items | `backend/pkg/repository/database/init_menu_item.go` | `alertsRule` -> `/alerts/rule`，`alertsNotify` -> `/alerts/notify`，`alertEvents` -> `/alerts/events`，`alertEventDetail` -> `/alerts/events/detail/:alertID/:eventID`。 |
| Parent Menu | `backend/pkg/repository/database/init_menu_item.go` | `alertsRule`、`alertsNotify`、`alertEvents` 归属父级 `alerts`。 |
| Feature Mapping | `backend/pkg/repository/database/init_feature_mapping.go` | `告警规则`、`告警通知`、`告警事件` 绑定对应菜单，`告警事件` 额外绑定详情 route。 |
| Frontend Render | `frontend/src/core/contexts/UserContext.js`, `frontend/src/core/components/AppSidebar.js` | 菜单最终由 `/api/permission/config` 返回的 `menuItems` 渲染。 |
| Legacy Navigation | `frontend/src/_nav.js` | 保留 `alerts` 组和 `alertsRule`、`alertsNotify` 静态项，但当前侧边栏主要使用服务端返回的 `menuItems`。 |

## Key Components

| Component | File | Role |
|---|---|---|
| `AlertsRule` | `frontend/src/oss/views/alerts/AlertsRule.jsx` | 规则列表、搜索、状态映射、分页和新增/编辑/删除入口。 |
| `ModifyAlertRuleModal` | `frontend/src/oss/views/alerts/modal/ModifyAlertRuleModal.jsx` | 规则表单，负责组装 `alertRule`、处理 `labels`、`annotations`、`for` 和 `groupId`。 |
| `ALertConditionCom` | `frontend/src/oss/views/alerts/modal/ALertConditionCom.jsx` | PromQL 编辑器，加载 `quick_alert_rule_metric` 返回的 quick metrics。 |
| `DataGroupTreeSelector` | `frontend/src/core/components/DataGroup/DataGroupTreeSelector` | 规则弹窗内选择 `groupId`。 |
| `AlertsNotify` | `frontend/src/oss/views/alerts/AlertsNotify.jsx` | 通知 receiver 列表、搜索、分页和新增/编辑/删除入口。 |
| `ModifyAlertNotifyModal` | `frontend/src/oss/views/alerts/modal/ModifyAlertNotifyModal.jsx` | 组装 email、webhook、dingtalk、wechat receiver 请求结构。 |
| `AlertEventsPage` | `frontend/src/oss/views/alertEvents/index.tsx` | 告警事件列表顶层容器，管理统计、过滤、分页、workflow modal。 |
| `Filter` | `frontend/src/oss/views/alertEvents/components/Filter/index.tsx` | 动态过滤器容器，支持静态字段和 label 字段扩展。 |
| `CheckboxFilter` | `frontend/src/oss/views/alertEvents/components/Filter/CheckboxFilter.tsx` | 基于 `/api/alerts/filter/values` 加载可选值。 |
| `LabelKeyFilter` | `frontend/src/oss/views/alertEvents/components/Filter/LabelKeyFilter.tsx` | 支持 `labels.*`、`tags.*` 或 workflow 字段过滤。 |
| `AlertInfoCom` | `frontend/src/oss/views/alertEvents/components/AlertInfoCom.tsx` | 统一渲染告警级别、状态、持续时间、tag display、有效性。 |
| `CurrentEventDetail` | `frontend/src/oss/views/alertEvents/detail/CurrentEventDetail.tsx` | 当前事件详情和 workflow run-once/logs modal。 |
| `Events` | `frontend/src/oss/views/alertEvents/detail/Events.tsx` | 同一 `alertId` 的历史事件表格和 workflow logs modal。 |
| `SilentAlert` | `frontend/src/oss/views/alertEvents/detail/SilentAlert.tsx` | 静默读取、时间范围展示和保存。 |
| `WorkflowsIframe` | `frontend/src/oss/views/workflows/workflowsIframe` | 嵌入 `/dify/app/...` 页面。 |

## State Management

| State Area | Evidence | Notes |
|---|---|---|
| Group Label | `frontend/src/App.js`, `frontend/src/core/store/reducers/groupLabelReducer.js` | 应用启动时调用 `/api/alerts/rule/groups`，结果供 `AlertsRule` 和 `ModifyAlertRuleModal` 使用。 |
| Data Group | `frontend/src/core/store/reducers/dataGroupReducer.js` | `AlertsRule`、`AlertEventsPage`、过滤组件读取 `dataGroupId`。 |
| Time Range | `frontend/src/core/store/reducers/timeRangeReducer.js` | `AlertEventsPage` 与过滤组件读取 `startTime`、`endTime`。 |
| Local Table State | `AlertsRule`, `AlertsNotify`, `AlertEventsPage`, `Events` | 分页、loading、搜索条件、弹窗状态均为组件本地状态。 |
| Filter State | `AlertEventsPage` | 默认过滤为 `status=firing` 和 `validity=valid,other`，后续由 `Filter` 组件更新。 |
| Workflow Modal State | `AlertEventsPage`, `CurrentEventDetail`, `Events` | 本地保存 `modalOpen` 与 `workflowUrl`。 |
| Report State | `frontend/src/pro/views/report/index.tsx` | 报告页按 URL query 构造 `tags` 后本地保存 `reportData` 与 `reportType`。 |

## API Usage

| Frontend Action | Wrapper | API Path | Notes |
|---|---|---|---|
| 加载规则列表 | `getAlertRulesApi` | `/api/alerts/rule/list` | 请求包含 `currentPage`、`pageSize`、`alert`、`group`、`groupId`。 |
| 加载规则运行状态 | `getAlertRulesStatusApi` | `/alert/api/v1/rules` | 返回 groups/rules/state，当前属于外部边界。 |
| 新增规则 | `addRuleApi` | `/api/alerts/rule/add` | `ModifyAlertRuleModal` 组装 `alertRule` 和 `groupId`。 |
| 更新规则 | `updateRuleApi` | `/api/alerts/rule` | 编辑时额外传 `oldGroup`、`oldAlert`。 |
| 删除规则 | `deleteRuleApi` | `/api/alerts/rule` | 传 `group`、`alert`、`groupId`。 |
| 加载 group label | `getRuleGroupLabelApi` | `/api/alerts/rule/groups` | 在 `App.js` 启动阶段调用。 |
| 加载 quick metrics | `getRuleMetricsApi` | `/api/alerts/rule/metrics` | `ALertConditionCom` 用于 PromQL 快捷选择。 |
| 加载 receiver 列表 | `getAlertmanagerListApi` | `/api/alerts/alertmanager/receiver/list` | `AlertsNotify` 传 `currentPage`、`pageSize`、`name`、`refreshCache`。 |
| 新增 receiver | `addAlertNotifyApi` | `/api/alerts/alertmanager/receiver/add` | 支持 email、webhook、dingtalk、wechat。 |
| 更新 receiver | `updateAlertNotifyApi` | `/api/alerts/alertmanager/receiver` | 编辑时传 `oldName`。 |
| 删除 receiver | `deleteAlertNotifyApi` | `/api/alerts/alertmanager/receiver` | dingtalk 删除时额外传 `type: 'dingtalk'`。 |
| 加载事件列表 | `getAlertEventsApi` | `/api/alerts/event/list` | 请求包含时间范围、分页、filters、`groupId`。 |
| 加载静态过滤项 | `getAlertsFilterKeysApi` | `/api/alerts/filter/keys` | 返回 `severity`、`status`、`validity` 等静态项。 |
| 加载动态 label key | `getAlertsFilterLabelKeysApi` | `/api/alerts/filter/labelkeys` | 请求包含 `startTime`、`endTime`。 |
| 加载过滤值 | `getAlertsFilterValuesApi` | `/api/alerts/filter/values` | 前端传 `searchKey`、`startTime`、`endTime`、`groupId`。 |
| 获取 workflow id | `getAlertWorkflowIdApi` | `/api/alerts/events/classify` | 按 `alertGroup`、`alertName` 获取 `workflowId`。 |
| 加载事件详情 | `getAlertEventDetailApi` | `/api/alerts/event/detail` | 请求包含 `alertId`、`eventId`、分页和 `locateEvent`。 |
| 读取静默配置 | `getAlertSilentConfigApi` | `/api/alerts/slient` | query 参数为 `alertId`。 |
| 保存静默配置 | `saveAlertSilentConfigApi` | `/api/alerts/slient` | body 包含 `alertId`、`forDuration`。 |
| 手动 resolve | `resolveAlertApi` | `/api/alerts/resolve` | body 包含 `alertId`。 |
| 打开报告页 | direct `window.open` | `/#/report?alertEventId=...` | 由报告模块继续读取报告 API。 |
| 加载报告 | `getReportApi` | `/api/alerts/events/report/get` | `frontend/src/pro/views/report/index.tsx` 使用。 |

## User Interaction Flow

### Flow A
用户进入 `/alerts/rule` 后，页面从 `groupLabelReducer` 读取 group 下拉项，并按 `dataGroupId`、搜索条件和分页调用 `/api/alerts/rule/list`。列表返回后再调用 `/alert/api/v1/rules` 构造 `group-alert -> state` 映射，用 tag 显示 `firing`、`pending`、`inactive` 或 `unknown`。

### Flow B
用户新增或编辑告警规则时，`ModifyAlertRuleModal` 将表单转成 `AlertRule`，调用 `/api/alerts/rule/add` 或 `/api/alerts/rule`。保存成功后关闭弹窗并刷新规则表。

### Flow C
用户进入 `/alerts/notify` 后，页面调用 `/api/alerts/alertmanager/receiver/list` 拉取 receiver。新增或编辑时，`ModifyAlertNotifyModal` 根据 `notifyType` 组装不同字段：`emailConfigs`、`webhookConfigs`、`dingTalkConfigs`、`wechatConfigs`。

### Flow D
用户进入 `/alerts/events` 后，页面按时间范围、数据组、分页和过滤器调用 `/api/alerts/event/list`。后端返回事件、统计、`alertCheckId` 后，前端展示统计卡片、表格、有效性状态和操作按钮。点击 `cause` 会先调用 `/api/alerts/events/classify` 获取 workflow id，再打开 `/dify/app/{workflowId}/run-once?...`。

### Flow E
用户点击某条事件的详情按钮后跳转到 `/alerts/events/detail/:alertId/:eventId`。详情页调用 `/api/alerts/event/detail`，展示当前事件和同类历史事件；用户可读取/保存 `/api/alerts/slient`，也可调用 `/api/alerts/resolve` 手动 resolve 最新事件。

## Permissions

| Aspect | Evidence | Notes |
|---|---|---|
| Route Wrapper | `frontend/src/oss/routes.js` + `frontend/src/core/components/AuthRouter.tsx` | 本模块四个主 route 均被 `AuthRouter` 包装。 |
| Actual Frontend Route Check | `frontend/src/core/components/AuthRouter.tsx` | `checkRoutePermission` 的 `useEffect` 当前被注释，前端侧默认 `isAllowed=true`。 |
| Menu Visibility | `frontend/src/core/contexts/UserContext.js` + `frontend/src/core/components/AppSidebar.js` | 菜单是否展示由后端权限配置返回。 |
| Data Scope UI | `frontend/src/oss/routes.js` | `/alerts/rule` 和 `/alerts/events` 设置 `showDataGroup: true`。 |
| Backend Auth Boundary | `backend/pkg/router/router_api.go` | `/api/alerts` 内部一部分接口在 `AuthMiddleware` 前注册，一部分在之后注册，前端不能单独推断后端权限。 |

## Notes

1. `frontend/src/oss/views/alerts/index.js` 的 tabs 页面当前未被 `ossRoutes` 直接使用，不能作为主入口描述。
2. `reloadAlertRulesApi` wrapper 存在，但当前主页面未见调用。
3. `getAlertRulesStatusApi` 和 `reloadAlertRulesApi` 使用 `/alert/*`，应作为外部 AlertManager/VMAlert 边界，而不是本仓库主后端 `/api/alerts`。
4. `CheckboxFilter` 和 `LabelKeyFilter` 传 `groupId` 给 `/api/alerts/filter/values`，但后端请求结构当前未包含 `GroupID` 字段。
5. `AlertEventsPage` 用 `window.open('/#/report?...')` 连接报告页面，报告页细节属于 `conversational-diagnosis-and-reporting`，本模块只记录跳转和 API 输出边界。
