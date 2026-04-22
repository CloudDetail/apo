# Trace Exploration And Profiling Flows

## Main Flow 1

### Fault Site Trace Search

1. 用户进入 `/trace/fault-site`，页面由 `FaultSiteTracePage` 装配 `LogsTraceFilterProvider`。
2. 用户通过 `LogsTraceFilter type="trace"` 输入 cluster、namespace、service、instance、`traceId`、duration、fault type 和时间范围。
3. `FaultSiteTrace` 从共享上下文与 `dataGroupReducer` 读取筛选条件，组装 `/api/trace/pagelist` 请求。
4. 后端 `GetTracePageList` 先执行 `CheckGroupPermission`。
5. 若请求只给了 `groupId` 而没有 service，`service_get_trace_page_list.go` 会从 `DataGroupStorage` 自动补 service 范围。
6. 后端查询 ClickHouse `span_trace`，返回分页结果。
7. 前端把结果渲染为表格，并在操作列提供跳转日志页与打开 Jaeger 的入口。

## Main Flow 2

### Trace Detail Viewing

1. 用户在 `/trace/fault-site` 表格中点击某个 `traceId`。
2. 前端打开 `JaegerIframeModal`。
3. 弹窗直接 iframe 到 `/jaeger/trace/<traceId>?uiEmbed=v0`。

并行存在的另一条明细访问路径：

1. 用户进入 `/trace/full`。
2. 页面直接嵌入 `/jaeger/search`。
3. 实际搜索与明细浏览都在 Jaeger UI 内完成，而不是通过当前 React 页面自行请求数据。

## Main Flow 3

### Profiling And Flame Graph Retrieval

1. 外部调用方访问 `/api/trace/onoffcpu`、`/api/trace/flame` 或 `/api/trace/flame/process`。
2. 路由层只要求 `LicMiddleware`，不走 `AuthMiddleware`。
3. 后端从 ClickHouse 读取 `profiling_event` 或 `flame_graph`。
4. 对于 `/api/trace/flame/process`，service 会把多个线程 flame graph 合并成一个进程级 `flamebearer`。
5. 结果返回给调用方，当前主前端仓库中没有观察到对应消费页面。

## Edge Cases

1. 当 `groupId > 0` 且未传 service 时，后端会自动补齐可见 service 列表。
2. 当 namespace 为 `VM_NS` 时，后端会把它归一为空字符串，这同时作用于顶层 namespace 和嵌套筛选条件。
3. `faultTypeList` 在前端会被编码成带 `mergeSep: 'OR'` 的嵌套筛选结构，而不是简单的单字段枚举。
4. 若时间线跳入 `#/trace/fault-site` 时设置了 `sessionStorage.openJaegerModalAfterLoad`，页面首轮加载后会自动打开 Jaeger 弹窗。
5. `GetFlameGraphData` 如果查到多条结果，service 不会报错，只会记录 warning。

## Failure Paths

1. 请求参数绑定失败时，handler 直接返回 `ParamBindError`。
2. `/api/trace/pagelist` 权限检查失败时，会返回空分页结构。
3. Jaeger HTTP 调用失败时，`/api/trace/info` 会直接返回错误。
4. profiling / flame graph 查询失败时，handler 会返回相应的 `code.GetOnOffCPUError` 或 `code.GetFlameGraphError`。
5. 当前前端直接嵌入 Jaeger iframe，因此如果 Jaeger UI 或代理路径不可用，页面会直接表现为 iframe 加载失败，而不是本前端组件级错误提示。

## Permission Constraints

1. `/api/trace/pagelist` 有显式 `CheckGroupPermission`。
2. `/trace/fault-site` 页面也会把 `groupId` 传给后端，形成前后端一致的数据组约束。
3. `/api/trace/pagelist/filters`、`/api/trace/pagelist/filter/value`、`/api/trace/info` 受 `AuthMiddleware` 保护，但当前未观察到与 `groupId` 对应的数据范围拦截。
4. `/api/trace/onoffcpu`、`/api/trace/flame`、`/api/trace/flame/process` 当前只走 `LicMiddleware`，这是本模块最需要单独标注的权限特征。
