# Dashboards And Health Monitoring Open Questions

## Business Semantics That Cannot Be Fully Inferred From Code

1. `To Be Confirmed`: `/healthy-service` 的 RYG 健康评分是否面向运维诊断、SRE 巡检，还是面向业务方的服务健康摘要，目前只能从评分逻辑推断。
2. `To Be Confirmed`: `/api/service/monitor/status` 与 `/healthy-service` 的关系是否本应更紧密，目前代码里二者并未在前端接通。

## Legacy Or Dead-Code Suspicion

1. `To Be Confirmed`: `frontend/src/_nav.js` 仍保留 dashboard 和 healthy 的静态菜单定义，而实际菜单体系已经依赖后端权限配置。
2. `To Be Confirmed`: `IframeDashboard.jsx` 中包含较多 iframe 内部 DOM 操作和 `MutationObserver` 逻辑，且首个 `useEffect` 的清理函数里误写成 `addEventListener`，显示存在历史遗留代码痕迹。

## Missing Frontend-Backend Alignment

1. `To Be Confirmed`: `ServiceHealthyPage` 调 `/api/service/ryglight` 时传的是 `dataGroupId`，而后端 `GetRygLightRequest` 绑定字段是 `groupId`。
2. `To Be Confirmed`: `GetRygLightRequest` 支持 `endpointName` 和 `namespace`，但当前健康页只渲染了 `serviceName` 输入框。
3. `To Be Confirmed`: `MysqlDashboard` 和 `pro/components/DataViewCom/MetricsDashboard.jsx` 向 `IframeDashboard` 传的是 `src`，但组件接口定义的是 `srcProp`。
4. `To Be Confirmed`: `/api/service/monitor/status` 当前未观察到主前端消费点。

## Permission Uncertainty

1. `To Be Confirmed`: dashboard 页面真正访问的 Grafana / Jaeger URL 是否完全受当前系统权限控制，需要结合网关与部署代理确认。
2. `To Be Confirmed`: `/api/service/monitor/status` 没有看到 `groupId` 过滤逻辑，是否允许全局查看所有监控项需要确认。
3. `To Be Confirmed`: 前端 `AuthRouter` 当前未真正执行路由权限校验，因此 dashboard 页面入口是否完全依赖菜单控制与后端接口拦截，需要单独说明。

## Data Model Uncertainty

1. `To Be Confirmed`: `service_ryg_light.go` 对不同评分项采用固定阈值 `10/20/50`、`5/10/20` 等，这些阈值当前看起来是硬编码规则，不清楚是否应可配置。
2. `To Be Confirmed`: `GetMonitorStatus` 把 `1.0` 映射为 `isAlive=true`，其余值都视为 false，但当前没有区分 pending / maintenance 等更细状态。
