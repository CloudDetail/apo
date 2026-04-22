# Trace Exploration And Profiling Open Questions

## Business Semantics That Cannot Be Fully Inferred From Code

1. `To Be Confirmed`: `/trace/full` 作为纯 Jaeger iframe 页面是过渡方案还是最终产品形态。
2. `To Be Confirmed`: profiling/flame graph 接口虽然归在 `/api/trace`，但当前主前端未消费，它们在业务上是“本模块一部分”还是“共享后端接口层”。
3. `To Be Confirmed`: `faultTypeList` 在前端的 `slow`、`error`、`normal`、`slowAndError` 是否与后端或产品定义完全一致，目前只能从筛选拼装逻辑逆推出语义。

## Legacy Or Dead-Code Suspicion

1. `To Be Confirmed`: `frontend/src/oss/views/trace/TraceMoreFilters.jsx` 当前未观察到任何引用，疑似旧版筛选面板残留。
2. `To Be Confirmed`: `frontend/src/oss/views/trace/mock.js` 中的 `traceTableMock` 只被当作渲染开关引用，显示存在遗留调试代码。
3. `To Be Confirmed`: `frontend/src/_nav.js` 仍保留静态 trace 菜单，而当前主菜单体系已经依赖后端权限配置，可能是旧导航残留。

## Missing Frontend-Backend Alignment

1. `To Be Confirmed`: 当前前端只使用 `/api/trace/pagelist`，而后端还提供 `/api/trace/pagelist/filters`、`/api/trace/pagelist/filter/value`、`/api/trace/info`、profiling/flame graph 等接口，存在明显未接线能力。
2. `To Be Confirmed`: 前端 `JaegerIframeModal` 直接走 `/jaeger/trace/<traceId>?uiEmbed=v0`，没有使用 `/api/trace/info`。
3. `To Be Confirmed`: `GetTracePageListRequest` 支持 `Instance`，但当前前端在 `FaultSiteTrace.jsx` 中注释掉了 `instance` 传参，只传 `containerId`、`nodeName`、`pid`。
4. `To Be Confirmed`: `GetTraceFiltersRequest.StartTime` 的 form tag 是 `statTime`，可能影响基于 query 参数的调用对齐。

## Permission Uncertainty

1. `To Be Confirmed`: `/api/trace/onoffcpu`、`/api/trace/flame`、`/api/trace/flame/process` 当前只受 `LicMiddleware` 约束，不走 `AuthMiddleware`，需要确认是否符合预期。
2. `To Be Confirmed`: `/api/trace/pagelist/filters`、`/api/trace/pagelist/filter/value`、`/api/trace/info` 虽然要求认证，但没有看到与 `groupId` 对应的数据范围限制。
3. `To Be Confirmed`: `/trace/full` 直接 iframe `/jaeger/search` 的访问控制，是否完全受当前系统会话与代理配置保护，需要结合部署配置确认。
4. `To Be Confirmed`: 前端 `AuthRouter` 当前实际校验被注释，页面进入权限更多依赖菜单配置和后端接口拦截。

## Data Model Uncertainty

1. `To Be Confirmed`: `span_trace` 表里的 `flags`、`labels`、`metrics` 字段完整语义当前只能从查询与展示字段反推。
2. `To Be Confirmed`: `GetTraceFilters` 的过滤键缓存策略是 48 小时，但是否足以反映动态字段变化，需要结合运行场景判断。
3. `To Be Confirmed`: `GetProcessFlameGraphData` 合并多个线程 flame graph 时，如果 `labels` 不一致仅记录 warning，不清楚这是否允许产生语义混合的结果。
