# Logs Exploration And Parsing Dependencies

## Upstream Dependencies

| Dependency | Evidence | Why It Matters |
|---|---|---|
| `data-scope-governance` | `DataSourceFilter` + `dataGroupReducer` + `CheckGroupPermission` | `/logs/fault-site` 的筛选范围与数据权限依赖数据组治理能力。 |
| `identity-access-and-organization` | `UserContext`、`/api/permission/config`、`AuthMiddleware` | 菜单显隐与登录后访问基础能力依赖统一权限体系。 |
| `service-observability` | `ConfigLogRuleModal` 调 `getServiceListApi` | 解析规则配置页需要 service 列表作为辅助输入。 |
| `trace-exploration-and-profiling` | `LogsTraceFilter` / `LogsTraceFilterContext` 共享筛选模型 | 日志与 trace 的故障现场筛选组件明显复用同一套概念模型。 |
| Shared Time Range Infrastructure | `timeRangeReducer` | `/logs/full` 依赖统一系统时间范围。 |

## Downstream Dependencies

| Consumer | Evidence | Dependency Form |
|---|---|---|
| `service-observability` | `frontend/src/oss/views/serviceInfo/component/infoUni/ErrorInstance/ErrorInstanceInfo.jsx`、`LogsInfo.jsx` | 通过页面跳转把用户导向 `#/logs/fault-site`，把本模块作为后续排障入口。 |
| 用户侧运维流程 | `/logs/full` 中的规则管理与外部表接入 | 本模块不仅被消费，还会反向影响日志采集与结构化治理。 |

目前没有在代码中看到更多前端模块直接嵌入 `/logs/full` 内部组件；已确认的直接下游主要是跳转到 `/logs/fault-site` 的排障链路。

## Shared Components

| Shared Component | File | Usage |
|---|---|---|
| `DataSourceFilter` | `frontend/src/core/components/Filter/DataSourceFilter.tsx` | `/logs/fault-site` 的 cluster/namespace/service/instance 筛选。 |
| `DateTimeRangePickerCom` | `frontend/src/core/components/DateTime/DateTimeRangePickerCom` | 故障现场日志页内置时间范围选择。 |
| `BasicCard` | `frontend/src/core/components/Card/BasicCard` | 两个页面的主要容器。 |
| `LoadingSpinner` | `frontend/src/core/components/Spinner` | 故障现场日志页、全量日志页、上下文弹窗都使用。 |
| `notify` | `frontend/src/core/utils/notify` | 解析规则与外部表增删成功提示。 |
| `CodeMirrorSearch` | `frontend/src/oss/views/logs/component/FullLogs/component/SerarchBar/RawLogQuery/CodeMirrorSearch` | `/logs/full` 查询输入控件。 |

## Shared Utilities

| Utility | File | Usage |
|---|---|---|
| `request` | `frontend/src/core/utils/request` | 所有前端 API wrapper 复用。 |
| `convertTime` | `src/core/utils/time` | 故障现场时间点、全量日志时间戳展示。 |
| `swTraceIDToTraceID` | `src/core/utils/trace` | `LogsTraceFilter` 里 trace 模式的共享能力，表明筛选组件与 trace 模块耦合。 |
| `react-use/useDebounce` | `FaultSiteLogs.jsx`, `FullLogs/index.jsx` | 减少频繁筛选导致的重复请求。 |

## External Services

| External Service | Evidence | Purpose |
|---|---|---|
| ClickHouse | `service_get_fault_log_page_list.go`, `service_query_log.go`, `service_other_log.go` | 存储和查询故障现场日志、全量日志与外部表元数据。 |
| Kubernetes Vector ConfigMap | `service_add_log_parse_rule.go`, `service_update_log_parse_rule.go`, `service_delete_log_parse_rule.go`, `vector/factory.go` | 解析规则生命周期会直接修改采集与路由配置。 |
| Prometheus | `service_get_service_route.go` | 通过活跃实例信息推断 route rule。 |
| Relational Database | `dao_logtableinfo.go`, `dao_othertableinfo.go` | 保存解析规则元数据和外部表登记元数据。 |

## Cross-Module Linkage Notes

1. 前端文档应把 `/logs/fault-site` 与 `service-observability` 的跳转关系单独写清，避免误以为它是一个孤立页面。
2. 后端文档应把解析规则生命周期与 `data-ingestion-and-oneagent` 或采集侧能力显式关联，因为这里会直接修改 Vector 配置，而不是单纯读查询数据。
3. `LogsTraceFilter` 的共享实现意味着日志与 trace 文档需要互相引用，而不是分别把筛选器当成独立实现。
