# Trace Exploration And Profiling Source Map

## Scope
本文档先固定 `trace-exploration-and-profiling` 模块的代码证据，不预先发明需求。当前模块实际包含三条线：

- `/trace/fault-site` 故障现场链路检索页
- `/trace/full` Jaeger 搜索页的 iframe 承载入口
- `/api/trace` 下的 Trace 列表、Jaeger 明细代理、profiling / flame graph 后端接口

## Relevant Source Directories

| Layer | Path | Notes |
|---|---|---|
| Frontend | `frontend/src/oss/views/trace` | 本模块主要前端目录。 |
| Frontend | `frontend/src/oss/views/trace/component` | Jaeger iframe 弹窗与 Trace 状态标签。 |
| Frontend | `frontend/src/oss/contexts/LogsTraceFilterContext.tsx` | `/trace/fault-site` 复用的筛选状态容器。 |
| Frontend | `frontend/src/oss/components/Filter/LogsTraceFilter.jsx` | Trace 页面实际复用的筛选组件。 |
| Frontend | `frontend/src/core/api/trace.js` | 当前前端唯一 trace 专属 API wrapper。 |
| Frontend | `frontend/src/constants.js` | `DefaultTraceFilters` 定义。 |
| Frontend | `frontend/src/core/components/AuthRouter.tsx` | Trace 路由同样经由统一 guard 包装，但实际校验逻辑被注释。 |
| Frontend | `frontend/src/core/components/TimeLine/index.jsx` | 共享时间线组件可跳到 `#/trace/fault-site`。 |
| Frontend | `frontend/src/oss/views/serviceInfo/component/infoUni/TimeLine.jsx` | Service 详情页时间线跳转到 Trace 页面。 |
| Backend | `backend/pkg/router/router_api.go` | `/api/trace` 路由注册入口。 |
| Backend | `backend/pkg/api/trace` | Trace handler 层。 |
| Backend | `backend/pkg/services/trace` | Trace 查询、Jaeger 单链路、profiling 和 flame graph service。 |
| Backend | `backend/pkg/model/request/trace_req.go` | Trace 与 profiling 请求结构。 |
| Backend | `backend/pkg/model/request/span_trace.go` | Trace 复杂筛选结构与过滤运算符定义。 |
| Backend | `backend/pkg/model/response/trace_resp.go` | Trace 查询、Jaeger、profiling 响应结构。 |
| Backend | `backend/pkg/model/response/span_trace.go` | Trace 过滤器响应结构。 |
| Backend | `backend/pkg/repository/clickhouse` | `span_trace`、`profiling_event`、`flame_graph` 查询来源。 |
| Backend | `backend/pkg/repository/jaeger` | 通过 HTTP 读取 Jaeger trace 明细。 |
| Backend | `backend/pkg/services/data` | `/api/trace/pagelist` 的数据组权限校验依赖。 |

## Route Files

| Side | File | Observed Routes |
|---|---|---|
| Frontend | `frontend/src/oss/routes.js` | `/trace/fault-site`, `/trace/full` |
| Backend | `backend/pkg/router/router_api.go` | `/api/trace/onoffcpu`, `/api/trace/flame`, `/api/trace/flame/process`, `/api/trace/pagelist`, `/api/trace/pagelist/filters`, `/api/trace/pagelist/filter/value`, `/api/trace/info` |

## Page Files

| Type | File | Notes |
|---|---|---|
| Route Page | `frontend/src/oss/views/trace/FaultSiteTracePage.tsx` | 用 `LogsTraceFilterProvider` 包装故障现场 Trace 页。 |
| Route Page | `frontend/src/oss/views/trace/FullTrace.jsx` | 直接嵌入 `/jaeger/search`。 |
| Page Core | `frontend/src/oss/views/trace/FaultSiteTrace.jsx` | 故障现场 Trace 主页面，负责筛选、分页、Jaeger 弹窗、跳转日志页。 |
| Component | `frontend/src/oss/views/trace/component/JaegerIframeModal.jsx` | 单条 Trace 的 Jaeger iframe 弹窗。 |
| Component | `frontend/src/oss/views/trace/component/TraceErrorType.jsx` | slow / error / normal 状态标签。 |
| Legacy Candidate | `frontend/src/oss/views/trace/TraceMoreFilters.jsx` | 当前代码库中未观察到引用。 |
| Legacy Candidate | `frontend/src/oss/views/trace/mock.js` | `traceTableMock` 仅被当作渲染开关引用。 |

## API Wrapper Files

| File | Wrapper | API Path | Notes |
|---|---|---|---|
| `frontend/src/core/api/trace.js` | `getTracePageListApi` | `/api/trace/pagelist` | 当前前端唯一直接消费的 trace API。 |
| `frontend/src/core/api/serviceInfo.js` | `getServiceTraceLogsApi` | `/api/service/trace/logs` | 属于 service 详情侧时间线能力，不并入本模块主体实现。 |
| `frontend/src/core/api/serviceInfo.js` | `getTraceMetricsApi` | `/api/service/trace/metrics` | 同上，归 `service-observability`。 |

## Controller Or Router Files

| Type | File | Notes |
|---|---|---|
| Router | `backend/pkg/router/router_api.go` | `/api/trace` route group 先套 `LicMiddleware`，然后只对部分接口再加 `AuthMiddleware`。 |
| Handler Root | `backend/pkg/api/trace/handler.go` | 组装 `traceService` 与 `dataService`。 |
| Handler | `backend/pkg/api/trace/func_gettracepagelist.go` | `/api/trace/pagelist`，带 `CheckGroupPermission`。 |
| Handler | `backend/pkg/api/trace/func_gettracefilters.go` | `/api/trace/pagelist/filters`。 |
| Handler | `backend/pkg/api/trace/func_gettracefiltervalue.go` | `/api/trace/pagelist/filter/value`。 |
| Handler | `backend/pkg/api/trace/func_getsingletraceinfo.go` | `/api/trace/info`，返回 Jaeger 单链路明细。 |
| Handler | `backend/pkg/api/trace/func_getonoffcpu.go` | `/api/trace/onoffcpu`，profiling 事件。 |
| Handler | `backend/pkg/api/trace/func_getflamegraphdata.go` | `/api/trace/flame`，span 级 flame graph。 |
| Handler | `backend/pkg/api/trace/func_getprocessflamegraph.go` | `/api/trace/flame/process`，进程级 flame graph。 |

## Service Files

| Service Area | Files | Notes |
|---|---|---|
| Fault Site Trace List | `backend/pkg/services/trace/service_get_trace_page_list.go` | 故障现场 Trace 分页列表，支持数据组自动补 service 范围。 |
| Filter Keys | `backend/pkg/services/trace/service_get_trace_filters.go` | 获取可用筛选项。 |
| Filter Values | `backend/pkg/services/trace/service_get_trace_filter_values.go` | 获取指定筛选字段的候选值。 |
| Single Trace | `backend/pkg/services/trace/service_get_single_trace_info.go` | 通过 Jaeger Repo 拉取单个 trace。 |
| Profiling Events | `backend/pkg/services/trace/service_get_on_off_cpu.go` | 读取 `profiling_event`。 |
| Span Flame Graph | `backend/pkg/services/trace/service_get_flame_data.go` | 读取 span 级 flame graph。 |
| Process Flame Graph | `backend/pkg/services/trace/service_get_process_flame_data.go` | 合并多个线程 flame graph 为进程级结果。 |

## Model Entity Schema Files

| Type | File | Notes |
|---|---|---|
| Request | `backend/pkg/model/request/trace_req.go` | Trace 列表、单条 Trace、profiling、flame graph 请求。 |
| Request | `backend/pkg/model/request/span_trace.go` | `SpanTraceFilter`、`ComplexSpanTraceFilter`、`Operation`、`DataType`。 |
| Response | `backend/pkg/model/response/trace_resp.go` | Trace 列表、Jaeger、profiling、flame graph 响应。 |
| Response | `backend/pkg/model/response/span_trace.go` | Trace 过滤器键和值响应。 |
| Repository Types | `backend/pkg/repository/clickhouse/dao_span_trace.go` | `QueryTraceResult` 与筛选 SQL 逻辑。 |
| Repository Types | `backend/pkg/repository/clickhouse/dao_profilling_event.go` | `ProfilingEvent`。 |
| Repository Types | `backend/pkg/repository/clickhouse/dao_flame_graph.go` | `FlameGraphData`。 |

## Important Shared Dependencies

| Dependency | File | Notes |
|---|---|---|
| Shared Filter UI | `frontend/src/oss/components/Filter/LogsTraceFilter.jsx` | Trace 故障现场页与日志故障现场页共享筛选器。 |
| Shared Trace/Logs Context | `frontend/src/oss/contexts/LogsTraceFilterContext.tsx` | Trace 页面没有独立 Context。 |
| Data Group State | `frontend/src/core/store/reducers/dataGroupReducer.js` | `/trace/fault-site` 使用 `dataGroupId`。 |
| Menu And Route Permission | `frontend/src/core/contexts/UserContext.js`, `frontend/src/core/components/AuthRouter.tsx` | 菜单与路由权限依赖统一机制，但前端守卫当前未真正生效。 |
| Data Group Permission Service | `backend/pkg/services/data` | `/api/trace/pagelist` 的权限检查依赖它。 |
| Shared Group Scope Cache | `backend/pkg/services/common/group_filter.go`, `data_group_store_v2.go` | `GetTracePageList` 依赖 `DataGroupStorage` 补全 service 范围。 |
| Cross Module Jump | `frontend/src/oss/views/serviceInfo/component/infoUni/TimeLine.jsx` | service 详情时间线可以跳到 `#/trace/fault-site`。 |
| Jaeger Repo | `backend/pkg/repository/jaeger` | `/api/trace/info` 通过它代理 Jaeger trace。 |

## Evidence Notes

1. 当前前端只直接调用了 `/api/trace/pagelist`；`/api/trace/pagelist/filters`、`/api/trace/pagelist/filter/value`、`/api/trace/info`、profiling/flame graph 接口没有在当前主前端中观察到消费点。
2. `/trace/full` 并没有通过 `/api/trace/info` 拉数据，而是直接 iframe 到 `/jaeger/search`。
3. `router_api.go` 中有明确注释：`/api/trace/onoffcpu`、`/api/trace/flame`、`/api/trace/flame/process` “used by another project. DO NOT Auth.”，说明 profiling 能力当前更像后端共享接口而非本前端主入口。
4. `frontend/src/_nav.js` 仍保留静态 trace 菜单，但当前实际菜单体系主要依赖 `UserContext` 从后端权限配置拉取，存在旧导航残留迹象。
