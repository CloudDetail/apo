# Logs Exploration And Parsing Module Overview

## Module Goal
本模块提供两类互补的日志能力：

- 面向排障回放的 `/logs/fault-site`，围绕故障时间点、`traceId`、service、instance 拉取现场日志分页与上下文内容。
- 面向日志库运营的 `/logs/full`，围绕全量检索、直方图、字段索引分析、解析规则维护与外部日志表接入展开。

它既是日志查询入口，也是日志结构治理入口。

## Scope
当前代码证据支持以下范围：

- `/logs/fault-site` 故障现场日志分页与单条内容查看
- `/logs/full` 全量日志查询、分页、直方图、字段索引分析
- 单条日志上下文查看
- 解析规则读取、新增、更新、删除
- 基于 service 自动推断日志路由规则
- 外部日志表浏览、登记、删除
- 默认解析日志表 `raw_logs` / `all_logs` 的后端启动初始化

当前代码不应纳入本模块主体范围的能力：

- `/trace/*` 页面主体逻辑
- `/service/info` 中的日志时间线与错误实例面板主体逻辑
- 健康监控页或仪表盘页面
- 告警规则与告警事件治理

## Frontend Entry

| Entry Type | Evidence | Notes |
|---|---|---|
| Route | `frontend/src/oss/routes.js` | 注册 `/logs/fault-site` 与 `/logs/full`。 |
| Page | `frontend/src/oss/views/logs/FaultSiteLogsPage.tsx` | 故障现场日志页入口。 |
| Page | `frontend/src/oss/views/logs/FullLogsPage.jsx` | 全量日志页入口。 |
| Context | `frontend/src/oss/contexts/LogsTraceFilterContext.tsx` | 故障现场日志筛选状态。 |
| Context | `frontend/src/core/contexts/LogsContext.js` | 全量日志查询状态与方法集合。 |
| Redux | `frontend/src/core/store/reducers/logsReducer.js` | 全量日志全局状态存储。 |
| Shared Filter | `frontend/src/oss/components/Filter/LogsTraceFilter.jsx` | 故障现场日志查询条件入口。 |

## Backend Entry

| Entry Type | Evidence | Notes |
|---|---|---|
| Router Group | `backend/pkg/router/router_api.go` | 统一挂在 `/api/log` 下，并套用 `AuthMiddleware` 与 `LicMiddleware`。 |
| Handler Package | `backend/pkg/api/log` | 负责请求绑定、默认值、错误处理。 |
| Service Package | `backend/pkg/services/log` | 负责日志查询、解析规则、外部表管理。 |
| Vector Helper | `backend/pkg/services/log/vector` | 负责把解析规则同步到 Vector ConfigMap。 |
| DAO Models | `backend/pkg/repository/database/dao_logtableinfo.go`, `backend/pkg/repository/database/dao_othertableinfo.go` | 保存解析规则与外部表接入元数据。 |

## Core Concepts

| Concept | Explanation |
|---|---|
| `fault-site logs` | 面向故障时刻的日志回放视图，查询入口带 `groupId` 与数据源筛选。 |
| `full logs` | 面向全量日志检索与分析的视图，围绕表、字段与查询语句运转。 |
| `parse rule` | 把原始日志路由到 `logs_<parseName>` 并提取字段的配置集合。 |
| `route rule` | 解析规则命中条件，最终写入 Vector 的 `route_logs` 配置。 |
| `logtableinfo` | 解析日志表的元数据表，保存 `ParseName`、`RouteRule`、`ParseRule`、`Service`、`IsStructured` 等信息。 |
| `otherlogtable` | 外部日志表登记表，保存外部表的 `database`、`tablename`、`timefield`、`instance` 等元数据。 |
| `defaultFields` / `hiddenFields` | `/logs/full` 中把非日志字段和日志解析字段拆开显示，用于字段树与显隐控制。 |
| `context` | 以某条日志为中心，向前后查询相邻日志。当前前端只对内置解析日志表开放。 |

## Related Files

- `frontend/src/oss/views/logs/FaultSiteLogs.jsx`
- `frontend/src/oss/views/logs/component/FullLogs/index.jsx`
- `frontend/src/oss/views/logs/component/FullLogs/component/ConfigLogRuleModal/index.jsx`
- `frontend/src/oss/views/logs/component/FullLogs/component/ConfigTableModal/index.jsx`
- `frontend/src/core/contexts/LogsContext.js`
- `frontend/src/core/store/reducers/logsReducer.js`
- `frontend/src/core/api/logs.js`
- `backend/pkg/api/log/handler.go`
- `backend/pkg/services/log/service_get_fault_log_page_list.go`
- `backend/pkg/services/log/service_query_log.go`
- `backend/pkg/services/log/service_add_log_parse_rule.go`
- `backend/pkg/services/log/service_update_log_parse_rule.go`
- `backend/pkg/services/log/service_delete_log_parse_rule.go`
- `backend/pkg/services/log/service_add_other_table.go`
- `backend/pkg/services/log/vector/factory.go`
- `backend/pkg/repository/database/dao_logtableinfo.go`
- `backend/pkg/repository/database/dao_othertableinfo.go`

## Dependencies

| Dependency Type | Dependency | Why It Matters |
|---|---|---|
| Frontend Shared | `DataSourceFilter` | `/logs/fault-site` 的 cluster/namespace/service/instance 选择依赖共享数据源筛选器。 |
| Frontend Shared | `timeRangeReducer` | `/logs/full` 依赖共享系统时间范围。 |
| Frontend Shared | `UserContext` + `AuthRouter` | 菜单显隐与路由进入都复用统一权限基础设施。 |
| Cross Module | `core/api/service.getServiceListApi` | 解析规则配置弹窗要读取 service 列表。 |
| Backend Shared | `data.Service.CheckGroupPermission` | 故障现场日志分页会按数据组做权限检查。 |
| Repository | `clickhouse.Repo` | 实际日志查询、外部表元信息、字段索引与上下文都依赖它。 |
| Repository | `database.Repo` | 解析规则与外部表元数据都持久化在这里。 |
| Repository | `kubernetes.Repo` | 解析规则生命周期会修改 Vector ConfigMap。 |
| Repository | `prometheus.Repo` | `GetServiceRoute` 通过活跃实例推断 `k8s.pod.name` 路由规则。 |
| Cross Module | `service-observability` | 该模块的错误实例与日志面板会跳转到 `#/logs/fault-site`。 |
| Cross Module | `trace-exploration-and-profiling` | `LogsTraceFilter` 与 `LogsTraceFilterContext` 是共享筛选模型的一部分。 |

## Risks And Unknowns

1. `/logs/full` 当前前后端代码中未观察到 `groupId` 或等价的数据组范围约束，语义需要 `To Be Confirmed`。
2. `GetFaultLogContentRequest.ServiceName` 在 handler 中明确标注未用于权限校验，说明故障现场内容接口存在已知权限缺口。
3. 解析规则新增、更新、删除跨越 ClickHouse、Kubernetes ConfigMap 与数据库三类资源，当前未见事务式回滚。
4. `GetServiceRoute` 通过实例名拆分 pod 前缀推断 `k8s.pod.name`，对命名规范依赖较强。
