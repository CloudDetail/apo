# Trace Exploration And Profiling Backend

## API Summary

### Fault Site Trace APIs

| API | Handler | Service | Notes |
|---|---|---|---|
| `POST /api/trace/pagelist` | `GetTracePageList` | `GetTracePageList` | 故障现场 Trace 分页检索。 |
| `GET /api/trace/pagelist/filters` | `GetTraceFilters` | `GetTraceFilters` | 读取可用 Trace 筛选字段。 |
| `POST /api/trace/pagelist/filter/value` | `GetTraceFilterValue` | `GetTraceFilterValues` | 读取指定筛选字段候选值。 |
| `GET /api/trace/info` | `GetSingleTraceInfo` | `GetSingleTraceID` | 代理 Jaeger 单条 Trace 明细。 |

### Profiling And Flame Graph APIs

| API | Handler | Service | Notes |
|---|---|---|---|
| `GET /api/trace/onoffcpu` | `GetOnOffCPU` | `GetOnOffCPU` | 读取 `profiling_event`。 |
| `GET /api/trace/flame` | `GetFlameGraphData` | `GetFlameGraphData` | 读取 span 级 flame graph。 |
| `GET /api/trace/flame/process` | `GetProcessFlameGraph` | `GetProcessFlameGraphData` | 合并线程 flame graph 为进程级结果。 |

## Controllers Or Routers

| File | Responsibility |
|---|---|
| `backend/pkg/router/router_api.go` | 定义 `/api/trace` 路由组，并把 profiling/flame graph 与 pagelist/info 接口拆成不同认证层。 |
| `backend/pkg/api/trace/handler.go` | 组装 `traceService` 与 `dataService`。 |
| `backend/pkg/api/trace/func_gettracepagelist.go` | 绑定 `GetTracePageListRequest`，做 `CheckGroupPermission`，补默认分页。 |
| `backend/pkg/api/trace/func_gettracefilters.go` | 绑定 `GetTraceFiltersRequest`。 |
| `backend/pkg/api/trace/func_gettracefiltervalue.go` | 绑定 `GetTraceFilterValueRequest`，把微秒时间转为 `time.Time`。 |
| `backend/pkg/api/trace/func_getsingletraceinfo.go` | 绑定 `GetSingleTraceInfoRequest` 并返回 Jaeger 明细字符串。 |
| `backend/pkg/api/trace/func_getonoffcpu.go` | 绑定 `GetOnOffCPURequest`。 |
| `backend/pkg/api/trace/func_getflamegraphdata.go` | 绑定 `GetFlameDataRequest`。 |
| `backend/pkg/api/trace/func_getprocessflamegraph.go` | 绑定 `GetProcessFlameGraphRequest`。 |

## Services

### Fault Site Trace Search

| File | Behavior |
|---|---|
| `backend/pkg/services/trace/service_get_trace_page_list.go` | 当 `groupId > 0` 且前端未传 service 时，会按数据组自动补齐可见 service；同时会把 `VM_NS` 归一为空字符串。 |
| `backend/pkg/services/trace/service_get_trace_filters.go` | 获取可用筛选字段，内部依赖 ClickHouse 缓存逻辑。 |
| `backend/pkg/services/trace/service_get_trace_filter_values.go` | 获取某个筛选字段可选值。 |
| `backend/pkg/services/trace/service_get_single_trace_info.go` | 通过 `jaegerRepo.GetSingleTrace(traceId)` 拉取 Jaeger 原始明细。 |

### Profiling And Flame Graph

| File | Behavior |
|---|---|
| `backend/pkg/services/trace/service_get_on_off_cpu.go` | 从 `profiling_event` 表读取 on/off CPU 事件。 |
| `backend/pkg/services/trace/service_get_flame_data.go` | 读取 span 级 flame graph；若结果多于一条会打 warning。 |
| `backend/pkg/services/trace/service_get_process_flame_data.go` | 拉取同一进程多个线程的 flame graph，合并成一个进程级 `flamebearer`。 |

## Models And Schemas

| Type | File | Notes |
|---|---|---|
| Request | `backend/pkg/model/request/trace_req.go` | Trace 列表、Jaeger 明细、profiling、flame graph 请求。 |
| Request | `backend/pkg/model/request/span_trace.go` | `SpanTraceFilter`、`ComplexSpanTraceFilter`、`Operation`、`DataType`。 |
| Response | `backend/pkg/model/response/trace_resp.go` | Trace 列表、Jaeger、profiling、flame graph 响应。 |
| Response | `backend/pkg/model/response/span_trace.go` | Trace 筛选字段和候选值响应。 |
| Repository Type | `backend/pkg/repository/clickhouse/dao_span_trace.go` | `QueryTraceResult` 与筛选 SQL 拼装。 |
| Repository Type | `backend/pkg/repository/clickhouse/dao_profilling_event.go` | `ProfilingEvent`。 |
| Repository Type | `backend/pkg/repository/clickhouse/dao_flame_graph.go` | `FlameGraphData`。 |

## Database Mapping

| Storage | Entity | Mapping |
|---|---|---|
| ClickHouse | `span_trace` | `dao_span_trace.go` 中的 Trace 列表与筛选查询。 |
| ClickHouse | `profiling_event` | `dao_profilling_event.go` 中的 profiling 事件查询。 |
| ClickHouse | `flame_graph` | `dao_flame_graph.go` 中的 flame graph 查询。 |
| Jaeger HTTP API | `/jaeger/api/traces/<traceId>` | `backend/pkg/repository/jaeger/trace.go` 中的单条 Trace 拉取。 |
| Relational DB | 无专属 trace 业务表证据 | 当前代码中 Trace 模块没有观察到独立的 trace 元数据表。 |

## Validation And Rules

1. `/api/trace` 路由组先统一应用 `LicMiddleware()`。
2. `router_api.go` 中明确规定 `/api/trace/onoffcpu`、`/api/trace/flame`、`/api/trace/flame/process` “DO NOT Auth”，即仅受 license 控制，不走 `AuthMiddleware`。
3. `POST /api/trace/pagelist` 在 handler 层会做 `CheckGroupPermission`，并补默认 `pageNum=1`、`pageSize=10`。
4. `GetTracePageList` service 在 `groupId > 0 && len(service)==0` 时，会按数据组权限自动补齐 service 列表。
5. `GetTracePageList` 会把请求中的 `VM_NS` 转换为空字符串；对嵌套筛选里的 namespace 也做相同归一。
6. `GetTraceFilters` 通过 ClickHouse repo 获取 filter key；若 `needUpdate == false`，repo 会优先使用缓存，超过 48 小时才刷新。
7. `ComplexSpanTraceFilter` 支持 `mergeSep` 和 `subFilters`，实际 SQL 在 `dao_span_trace.go` 中通过递归 `extractSpanFilter` 拼装。
8. `/api/trace/info` 返回类型是 `GetSingleTraceInfoResponse string`，当前没有额外结构化转换。

## Error Handling

1. 请求绑定失败统一通过 `AbortWithError(..., code.ParamBindError, err)` 返回。
2. `/api/trace/pagelist` 权限失败时不会只返回错误码，而是附带空 `list` 与空 `pagination` 的响应体。
3. `GetTraceFilters` 与 `GetTraceFilterValues` 在 service 层遇错时会尽量返回空数组结构，再把 error 往上抛。
4. `GetFlameGraphData` 若没有查询到数据，可返回零值响应而不是明确错误。
5. `GetProcessFlameGraphData` 在合并 flame graph 时，如果 `labels` 不一致或 `TID != 0`，会写 warning 日志，但不会直接报错中断。

## Notes

1. 当前主前端没有直接消费 `/api/trace/info`，也没有消费 profiling/flame graph 接口。
2. `GetTraceFiltersRequest` 的 `StartTime` form tag 是 `statTime`，与字段名 `startTime` 不一致，语义需要特别留意。
3. Trace 模块在后端上同时服务于本项目前端和“another project”，边界已经在路由注释中明确体现。
