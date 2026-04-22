# Alert Rules And Incident Events Open Questions

## Business Semantics That Cannot Be Fully Inferred From Code

1. `validity` 的业务含义需要确认。当前 ClickHouse 查询把 workflow `output = 'false'` 映射为 `valid`，`output = 'true'` 映射为 `invalid`，语义依赖 Dify workflow 输出约定。
2. `alertDirection` 出现时前端展示 `viewReport`，但 `alertDirection` 的生成规则在本模块中没有完整业务解释。
3. `AlertEventClassify` 根据 `alertGroup` 和 `alertName` 返回 workflow id，但不同告警类型如何映射到不同 workflow 需要从 Dify 配置或业务文档确认。
4. 事件列表默认过滤 `status=firing` 与 `validity=valid,other`，其中 `other` 在后端静态过滤项或查询语义中的含义需要确认。
5. 告警静默的业务范围需要确认。当前 `SilentAlert` 只按 `alertId` 保存 `forDuration`，是否影响同一 `alertId` 的所有未来事件需要确认。
6. 手动 resolve 的业务边界需要确认。后端方法名是 `ManualResolveLatestAlertEventByAlertID`，当前代码注释有 `TODO valid alertID`。

## Legacy Or Dead-Code Suspicion

1. `frontend/src/oss/views/alerts/index.js` 是告警规则/通知 tabs 容器，但当前 `frontend/src/oss/routes.js` 直接使用 `AlertsRule` 和 `AlertsNotify`，未见 `/alerts` 页面入口。
2. `frontend/src/core/api/alerts.js` 中 `reloadAlertRulesApi` 当前未见主页面调用。
3. `AlertsRule.jsx` 中同时存在 `loadRulesData` / `init` 和 `fetchData` 两组加载逻辑，当前 `useDebounce` 主要走 `loadRulesData` 或 `init`，`fetchData` 的实际入口需要确认。
4. `backend/pkg/services/alerts/inputs_alertmanager.go` 的 `InputAlertManager` 注释标记 deprecated，handler 当前改为调用 `alertinput.ProcessAlertEvents`。
5. `backend/pkg/api/alerts/func_alerteventlist.go` 的 swagger 注释写 `/api/alerts/events/list`，实际路由注册是 `/api/alerts/event/list`。
6. `backend/pkg/api/alerts/func_addalerteventreport.go` 的 swagger 注释与实际 router method 需要继续核对。

## Missing Frontend Backend Alignment

1. 前端 `getAlertsFilterValuesApi` 传 `groupId`，但 `SearchAlertEventFilterValuesRequest` 当前没有 `GroupID` 字段，过滤值查询可能没有按数据组过滤。
2. 前端 route 参数是 `/alerts/events/detail/:alertId/:eventId`，后端初始化元数据中使用 `/alerts/events/detail/:alertID/:eventID`，大小写不一致是否影响权限匹配需要确认。
3. `getAlertRulesStatusApi` 使用 `/alert/api/v1/rules`，当前主后端未见对应 route，规则状态来源需要确认部署时代理到哪里。
4. `getAlertRulesApi` 传 `groupId` 给 `/api/alerts/rule/list`，后端 `AlertRuleFilter.GroupID` 存在，但 `k8sApi.GetAlertRules` 内部如何应用该字段需要继续追踪。
5. `deleteRuleApi` 前端传 `groupId`，但 `DeleteAlertRuleRequest` 当前没有 `GroupID` 字段，删除规则时数据组是否应参与校验需要确认。
6. `getAlertEventReport` 前端报告页可按 `agentChatId` 查询，但本模块事件页只按 `alertEventId` 跳转；`agentChatId` 属于报告/agent 模块语义。

## Permission Uncertainty

1. `/api/alerts/event/list`、`/api/alerts/event/detail`、`/api/alerts/events/classify`、`/api/alerts/events/report/*`、`/api/alerts/inputs/alertmanager`、`/api/alerts/outputs/dingtalk/:uuid` 注册在 `AuthMiddleware` 前，是否有意开放需要确认。
2. 前端所有相关 route 被 `AuthRouter` 包装，但 `AuthRouter` 内实际调用 `getRouterPermissionApi` 的逻辑被注释，前端路由权限目前只依赖菜单是否可见。
3. `alertEventDetail` 是初始化菜单项又是 feature route，是否应该出现在菜单或仅作为隐藏 route 需要确认。
4. `/api/alerts/filter/*` 接口在后端受 `AuthMiddleware`，但前端的事件列表主体接口当前只受 license，这种分层是否符合产品预期需要确认。
5. AlertManager 输入和 DingTalk 输出接口若需被外部服务调用，缺少 auth 可能是设计；但需要明确是否通过网络层或 secret 保护。

## Data Model Uncertainty

1. `alert_slients`、`Slience`、`slient` 拼写贯穿模型、接口和响应字段，当前必须保持不变，但是否计划迁移为 `silent` 需要确认。
2. `AlertSlienceConfig` 中 gorm tag `primryKey` 拼写疑似错误，是否影响自动迁移或主键识别需要确认。
3. `dao_alert_event_report.go` 的 `SQL_GET_ALERT_EVENT_REPORT` select 列中 `rootcause` 出现两次，是否是笔误需要确认。
4. `AEventWithWRecord.IsValid` 已标记 deprecated，但前端组件仍围绕 `validity` 和部分 `isValid` 命名展示，后续字段兼容期需要确认。
5. `CheckAlertRule` 返回 `Available = find`，如果 `find` 表示规则已存在，则字段语义可能与“可用”相反。
6. `AlertEventClassify` 的 LRU cache key 是 `alertGroup + alertName`，没有分隔符，理论上存在拼接碰撞风险。
