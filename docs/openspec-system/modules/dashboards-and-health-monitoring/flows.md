# Dashboards And Health Monitoring Flows

## Main Flow 1

### Dashboard Rendering

1. 用户进入 `/system-dashboard`、`/basic-dashboard`、`/application-dashboard`、`/middleware-dashboard` 或 `/mysql-dashboard`。
2. 页面渲染 `IframeDashboard`。
3. `IframeDashboard` 在 `userReducer.menuItems` 中按 `dashboardKey` 查找目标菜单项。
4. 组件读取 `dashItem.router.page.url` 作为基础地址。
5. 组件根据当前时间范围追加 `from`、`to` 参数。
6. 组件根据当前主题追加 `theme` 参数。
7. 最终渲染 iframe，实际可视化内容来自 Grafana 或类似页面。

## Main Flow 2

### Service Health Scoring

1. 用户进入 `/healthy-service`。
2. 页面从 `timeRangeReducer` 和 `dataGroupReducer` 读取查询条件。
3. 页面调用 `/api/service/ryglight`。
4. 后端先做 `CheckGroupPermission`。
5. 后端基于 `groupId`、serviceName、endpointName、namespace 构造 PQL 过滤条件。
6. 后端查询延时同比、错误率同比、日志错误同比、实例数和告警计数。
7. 后端把各项检查转成 `RYGScoreDetail`，汇总出 `score`、`percentScore` 和 `status`。
8. 前端表格展示健康灯颜色、百分比分和评分详情。

## Main Flow 3

### Monitor Status Retrieval

1. 调用方访问 `/api/service/monitor/status`。
2. 后端根据时间范围调用 `prometheus.PQLMonitorStatus`。
3. 返回 `monitorName` 与 `isAlive` 列表。

当前代码中未观察到主前端消费这条链路。

## Edge Cases

1. 如果某个同比指标查不到昨日数据，健康评分会进入“跳过检查”或“昨日同时期无错误”的分支，而不是直接报错。
2. 如果 service 实例数小于 2，`replica` 检查项会直接给 0 分。
3. `MiddlewareDashboard` 除 iframe 外还会在页面顶部额外展示一个文档提示框。
4. dashboard URL 依赖 `menuItems` 提供；如果菜单里没有对应 `dashboardKey`，`IframeDashboard` 不会生成有效 `src`。

## Failure Paths

1. `/api/service/ryglight` 参数绑定失败时，handler 直接返回 `ParamBindError`。
2. `/api/service/ryglight` 权限失败时，返回空 `serviceList`。
3. `/api/service/monitor/status` 即使 Prometheus 查询失败，当前 service 也可能返回空结果而不是显式错误。
4. 如果 `IframeDashboard` 没有拿到有效 `router.page.url`，页面会渲染空 iframe 或不渲染预期内容。
5. `MysqlDashboard` 当前传参名与 `IframeDashboard` 不一致，存在页面无法正确生成 `src` 的风险。

## Permission Constraints

1. 健康页 `/healthy-service` 在前端路由中显式标记了 `showDataGroup: true`。
2. `/api/service/ryglight` 在后端会检查 `groupId` 对应的数据组权限。
3. `/api/service/monitor/status` 当前没有看到等价的数据组过滤逻辑。
4. dashboard 页面表面上只渲染 iframe，但其入口展示仍受菜单权限配置控制。
