# Trace Exploration And Profiling Module Overview

## Module Goal
本模块围绕 Trace 排障与 Trace profiling 暴露两层能力：

- 面向业务排障的 `/trace/fault-site`，按 service、namespace、instance、`traceId`、duration、fault type 等条件检索故障现场链路，并可直接打开 Jaeger 详情。
- 面向链路浏览的 `/trace/full`，直接承载 Jaeger 搜索页。
- 面向性能剖析的 `/api/trace/onoffcpu`、`/api/trace/flame`、`/api/trace/flame/process`，为 profiling 与 flame graph 提供后端接口。

因此它不是纯前端页面模块，而是“原生 Trace 检索页 + Jaeger 直连入口 + profiling 后端接口”的组合体。

## Scope
当前代码证据支持以下范围：

- `/trace/fault-site` 故障现场链路分页检索
- 通过 iframe 打开 Jaeger 单条 Trace
- `/trace/full` 直接嵌入 Jaeger 搜索页
- `/api/trace/pagelist`、`/api/trace/pagelist/filters`、`/api/trace/pagelist/filter/value`
- `/api/trace/info`
- `/api/trace/onoffcpu`
- `/api/trace/flame`
- `/api/trace/flame/process`

当前代码不应纳入本模块主体范围的能力：

- `/api/service/trace/metrics` 与 `/api/service/trace/logs` 的 service 详情侧 Trace 面板逻辑
- `/logs/*` 页面主体逻辑
- service 拓扑、依赖与实例面板

## Frontend Entry

| Entry Type | Evidence | Notes |
|---|---|---|
| Route | `frontend/src/oss/routes.js` | 注册 `/trace/fault-site` 与 `/trace/full`。 |
| Page | `frontend/src/oss/views/trace/FaultSiteTracePage.tsx` | 故障现场 Trace 页入口。 |
| Page | `frontend/src/oss/views/trace/FullTrace.jsx` | Jaeger 搜索页承载入口。 |
| Shared Context | `frontend/src/oss/contexts/LogsTraceFilterContext.tsx` | `/trace/fault-site` 没有独立 trace context，而是复用 logs/trace 共享筛选状态。 |
| Shared Filter | `frontend/src/oss/components/Filter/LogsTraceFilter.jsx` | 故障现场 Trace 页的筛选 UI。 |

## Backend Entry

| Entry Type | Evidence | Notes |
|---|---|---|
| Router Group | `backend/pkg/router/router_api.go` | `/api/trace` 先套 `LicMiddleware`，之后对部分接口再加 `AuthMiddleware`。 |
| Handler Package | `backend/pkg/api/trace` | 请求绑定、默认分页、错误处理与权限拦截。 |
| Service Package | `backend/pkg/services/trace` | Trace 列表、Jaeger 单链路、profiling、flame graph。 |
| Repository | `backend/pkg/repository/clickhouse` | `span_trace`、`profiling_event`、`flame_graph` 的核心查询来源。 |
| Repository | `backend/pkg/repository/jaeger` | `/api/trace/info` 的实际数据来源。 |

## Core Concepts

| Concept | Explanation |
|---|---|
| `fault-site trace` | 围绕故障时间点筛出异常链路，支持 slow / error / normal 过滤。 |
| `full trace` | 当前实现不是自建页面，而是 Jaeger 搜索页的 iframe 容器。 |
| `SpanTraceFilter` | Trace 列表筛选字段定义，带 `parentField`、`dataType`、`operation`。 |
| `ComplexSpanTraceFilter` | 支持 `mergeSep` 和 `subFilters` 的嵌套筛选结构。 |
| `QueryTraceResult` | Trace 列表返回项，包含 `flags`、`labels`、`metrics`、`thresholdValue`、`spanId` 等字段。 |
| `profiling_event` | 记录 on/off CPU 与线程执行消耗的 ClickHouse 数据源。 |
| `flame_graph` | 存储 span 级或线程级 flamebearer 数据的 ClickHouse 数据源。 |
| `Jaeger iframe` | 前端当前查看 Trace 详情的主要方式，而不是先调用 `/api/trace/info` 再渲染。 |

## Related Files

- `frontend/src/oss/views/trace/FaultSiteTrace.jsx`
- `frontend/src/oss/views/trace/FaultSiteTracePage.tsx`
- `frontend/src/oss/views/trace/FullTrace.jsx`
- `frontend/src/oss/views/trace/component/JaegerIframeModal.jsx`
- `frontend/src/oss/views/trace/component/TraceErrorType.jsx`
- `frontend/src/oss/components/Filter/LogsTraceFilter.jsx`
- `frontend/src/oss/contexts/LogsTraceFilterContext.tsx`
- `frontend/src/core/api/trace.js`
- `backend/pkg/api/trace/handler.go`
- `backend/pkg/services/trace/service_get_trace_page_list.go`
- `backend/pkg/services/trace/service_get_trace_filters.go`
- `backend/pkg/services/trace/service_get_single_trace_info.go`
- `backend/pkg/services/trace/service_get_on_off_cpu.go`
- `backend/pkg/services/trace/service_get_flame_data.go`
- `backend/pkg/services/trace/service_get_process_flame_data.go`
- `backend/pkg/model/request/trace_req.go`
- `backend/pkg/model/request/span_trace.go`
- `backend/pkg/model/response/trace_resp.go`
- `backend/pkg/model/response/span_trace.go`

## Dependencies

| Dependency Type | Dependency | Why It Matters |
|---|---|---|
| Frontend Shared | `LogsTraceFilterContext` + `LogsTraceFilter` | Trace 故障现场页与日志故障现场页共用同一筛选模型。 |
| Frontend Shared | `dataGroupReducer` | `/trace/fault-site` 查询时会带 `groupId`。 |
| Frontend Shared | `AuthRouter` + `UserContext` | 页面访问与菜单显隐走统一权限壳层。 |
| Backend Shared | `data.Service.CheckGroupPermission` | `/api/trace/pagelist` 的数据组权限检查依赖它。 |
| Backend Shared | `common.DataGroupStorage` | 当仅传 `groupId` 而未传 service 时，后端会自动补齐可见 service 范围。 |
| Repository | `clickhouse.Repo` | Trace 列表、筛选项、profiling、flame graph 全依赖它。 |
| Repository | `jaeger.JaegerRepo` | `/api/trace/info` 通过它向 Jaeger 发 HTTP 请求。 |
| Cross Module | `service-observability` | service 详情时间线会跳到 `#/trace/fault-site`。 |
| External Surface | Jaeger UI | `/trace/full` 与 `JaegerIframeModal` 直接嵌入其页面。 |

## Risks And Unknowns

1. 当前前端没有消费 `/api/trace/info` 与 profiling/flame graph 接口，模块的前后端覆盖面并不对称。
2. `/api/trace/onoffcpu`、`/api/trace/flame`、`/api/trace/flame/process` 明确不走 `AuthMiddleware`，权限边界需要特别确认。
3. `/trace/full` 直接 iframe `/jaeger/search`，意味着部分 Trace 浏览体验和权限模型可能不经过当前前端应用层。
4. `TraceMoreFilters.jsx`、`traceTableMock`、`_nav.js` 中的静态 trace 菜单都显示存在旧实现或残留分支。
