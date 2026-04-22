# Trace Exploration And Profiling Dependencies

## Upstream Dependencies

| Dependency | Evidence | Why It Matters |
|---|---|---|
| `data-scope-governance` | `dataGroupReducer`, `CheckGroupPermission`, `DataGroupStorage` | `/trace/fault-site` 的查询范围与数据权限直接依赖数据组治理。 |
| `identity-access-and-organization` | `AuthRouter`, `UserContext`, `AuthMiddleware` | 页面访问、菜单展示和 API 认证依赖统一权限体系。 |
| `logs-exploration-and-parsing` | `LogsTraceFilterContext`, `LogsTraceFilter` | Trace 故障现场页直接复用日志模块的筛选上下文和组件。 |
| Jaeger Runtime | `FullTrace.jsx`, `JaegerIframeModal.jsx`, `jaegerRepo` | 当前 Trace 详情与全量浏览都强依赖 Jaeger。 |
| ClickHouse Trace Storage | `span_trace`, `profiling_event`, `flame_graph` | Trace 列表与 profiling 数据的主要数据源。 |

## Downstream Dependencies

| Consumer | Evidence | Dependency Form |
|---|---|---|
| `service-observability` | `frontend/src/oss/views/serviceInfo/component/infoUni/TimeLine.jsx` | service 详情时间线把用户导向 `#/trace/fault-site`。 |
| 外部项目 | `router_api.go` 中 “These APIs are used by another project. DO NOT Auth.” | profiling/flame graph 接口被其他项目直接消费。 |

## Shared Components

| Shared Component | File | Usage |
|---|---|---|
| `LogsTraceFilter` | `frontend/src/oss/components/Filter/LogsTraceFilter.jsx` | `/trace/fault-site` 的主筛选组件。 |
| `LogsTraceFilterContext` | `frontend/src/oss/contexts/LogsTraceFilterContext.tsx` | 故障现场 Trace 页状态容器。 |
| `BasicCard` | `src/core/components/Card/BasicCard` | 故障现场 Trace 页布局容器。 |
| `BasicTable` | `src/core/components/Table/basicTable` | Trace 列表渲染。 |
| `LoadingSpinner` | `src/core/components/Spinner` | Trace 列表加载态。 |
| `TraceErrorType` | `frontend/src/oss/views/trace/component/TraceErrorType.jsx` | 状态标签组件，也可视为本模块内部共享部件。 |

## Shared Utilities

| Utility | File | Usage |
|---|---|---|
| `request` | `frontend/src/core/utils/request` | `getTracePageListApi` 的请求入口。 |
| `convertTime` | `src/core/utils/time` | 列表时间与持续时间展示。 |
| `TimestampToISO` | `src/core/utils/time` | 从 Trace 页跳转日志页时拼接时间参数。 |
| `DefaultTraceFilters` | `frontend/src/constants.js` | 构造 duration / namespace / fault type 过滤条件。 |
| `react-use/useDebounce` | `FaultSiteTrace.jsx` | 控制筛选请求频率。 |

## External Services

| External Service | Evidence | Purpose |
|---|---|---|
| ClickHouse | `dao_span_trace.go`, `dao_profilling_event.go`, `dao_flame_graph.go` | 提供 Trace 列表、profiling 事件与 flame graph 数据。 |
| Jaeger UI | `/jaeger/search`, `/jaeger/trace/<traceId>?uiEmbed=v0` | 提供前端当前实际 Trace 浏览界面。 |
| Jaeger HTTP API | `backend/pkg/repository/jaeger/trace.go` | 提供 `/api/trace/info` 的数据源。 |
| Relational Database | `dbRepo.GetScopeIDsSelectedByGroupID(...)` | 虽无 trace 专属业务表，但数据组权限补全依赖它。 |

## Cross-Module Linkage Notes

1. `trace-exploration-and-profiling` 与 `logs-exploration-and-parsing` 在故障现场页共享一套筛选器与状态模型，文档应显式互相引用。
2. `service-observability` 文档应把时间线跳转 `#/trace/fault-site` 标为跨模块入口，而不是把 trace 故障现场逻辑并入自身。
3. profiling/flame graph 虽归在本模块，但当前主前端未消费，应在系统文档里视为“后端能力先行”的接口面。
