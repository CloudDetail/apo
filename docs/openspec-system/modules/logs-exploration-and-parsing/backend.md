# Logs Exploration And Parsing Backend

## API Summary

### Fault Site APIs

| API | Handler | Service | Notes |
|---|---|---|---|
| `POST /api/log/fault/pagelist` | `GetFaultLogPageList` | `GetFaultLogPageList` | 故障现场分页列表，带数据组权限检查。 |
| `POST /api/log/fault/content` | `GetFaultLogContent` | `GetFaultLogContent` | 单次故障现场日志内容查询。 |

### Full Query APIs

| API | Handler | Service | Notes |
|---|---|---|---|
| `POST /api/log/query` | `QueryLog` | `QueryLog` | 全量日志检索。 |
| `POST /api/log/chart` | `GetLogChart` | `GetLogChart` | 直方图与总数统计。 |
| `POST /api/log/index` | `GetLogIndex` | `GetLogIndex` | 字段值分布分析。 |
| `POST /api/log/context` | `QueryLogContext` | `QueryLogContext` | 单条日志上下文。 |
| `GET /api/log/table` | `GetLogTableInfo` | `GetLogTableInfo` | 解析规则树与外部表树的基础信息。 |

### Parse Rule APIs

| API | Handler | Service | Notes |
|---|---|---|---|
| `GET /api/log/rule/service` | `GetServiceRoute` | `GetServiceRoute` | 根据 service 反推 `routeRule`。 |
| `GET /api/log/rule/get` | `GetLogParseRule` | `GetLogParseRule` | 读取解析规则详情。 |
| `POST /api/log/rule/add` | `AddLogParseRule` | `AddLogParseRule` | 新增解析规则。 |
| `POST /api/log/rule/update` | `UpdateLogParseRule` | `UpdateLogParseRule` | 更新解析规则。 |
| `DELETE /api/log/rule/delete` | `DeleteLogParseRule` | `DeleteLogParseRule` | 删除解析规则。 |

### External Table APIs

| API | Handler | Service | Notes |
|---|---|---|---|
| `GET /api/log/other` | `OtherTable` | `OtherTable` | 浏览可登记外部表。 |
| `GET /api/log/other/table` | `OtherTableInfo` | `OtherTableInfo` | 获取指定外部表列定义。 |
| `POST /api/log/other/add` | `AddOtherTable` | `AddOtherTable` | 把外部表登记到 `otherlogtable`。 |
| `DELETE /api/log/other/delete` | `DeleteOtherTable` | `DeleteOtherTable` | 删除外部表登记。 |

## Controllers Or Routers

| File | Responsibility |
|---|---|
| `backend/pkg/router/router_api.go` | 把全部日志能力挂到 `/api/log`，统一应用 `AuthMiddleware` 与 `LicMiddleware`。 |
| `backend/pkg/api/log/handler.go` | 组装 `logService`、`dataService`，并在启动阶段执行 `InitParseLogTable`。 |
| `backend/pkg/api/log/func_getfaultloglist.go` | 故障现场分页入口，包含 `CheckGroupPermission`。 |
| `backend/pkg/api/log/func_getfaultlogcontent.go` | 故障现场内容入口，当前未做 service 级权限校验。 |
| `backend/pkg/api/log/func_querylog.go` | 全量查询入口，设置默认 `pageNum`、`pageSize`、`query`、`timeField`、`logField`。 |
| `backend/pkg/api/log/func_getlogchart.go` | 直方图入口，设置默认 `query`、`timeField`、`logField`。 |
| `backend/pkg/api/log/func_getlogindex.go` | 字段索引入口，设置默认 `query`、`timeField`、`logField`。 |
| `backend/pkg/api/log/func_querylogcontext.go` | 上下文入口。 |
| `backend/pkg/api/log/func_getlogtableinfo.go` | 解析规则树与外部表树入口。 |
| `backend/pkg/api/log/func_getserviceroute.go` | service 到 route rule 的桥接入口。 |
| `backend/pkg/api/log/func_getlogparserule.go` | 读取解析规则。 |
| `backend/pkg/api/log/func_addparserule.go` | 新增解析规则。 |
| `backend/pkg/api/log/func_updatelogparserule.go` | 更新解析规则。 |
| `backend/pkg/api/log/func_deleteparserule.go` | 删除解析规则。 |
| `backend/pkg/api/log/func_othertable.go` | 可登记外部表列表。 |
| `backend/pkg/api/log/func_othertableinfo.go` | 外部表字段定义。 |
| `backend/pkg/api/log/func_addothertable.go` | 外部表登记入口，调用 `req.FillerValue()`。 |
| `backend/pkg/api/log/func_deleteothertable.go` | 外部表删除入口。 |

## Services

### Fault Site

| File | Behavior |
|---|---|
| `backend/pkg/services/log/service_get_fault_log_page_list.go` | 把请求转换成 `clickhouse.FaultLogQuery`；当 `groupId > 0` 且前端未传 service 时，会按数据组自动补齐可见 service 列表。 |
| `backend/pkg/services/log/service_get_fault_log_contents.go` | 先调用 `QueryApplicationLogs`；若结果为空，再构造 `raw_logs` 查询条件并调用 `QueryAllLogsInOrder`。 |

### Full Logs Query And Analysis

| File | Behavior |
|---|---|
| `backend/pkg/services/log/service_query_log.go` | `offset > 10000` 时先用 histogram 回算偏移；对内置解析日志表拆分 `Tags` 和 `LogFields`；对外部表把全部字段放入 `HiddenFields` / `LogFields`。 |
| `backend/pkg/services/log/service_query_log_context.go` | 从 `logtableinfo.Fields` 读取日志字段定义，然后对前后日志做 `log2item` 转换。 |
| `backend/pkg/services/log/service_get_log_chart.go` | 读取直方图后补齐空桶，并回填 `Count`。 |
| `backend/pkg/services/log/service_log_index.go` | 统计字段值分布并按数量降序返回。 |
| `backend/pkg/services/log/service_get_log_table_info.go` | 读取全部解析规则与已登记外部表，并组装为前端树结构所需的 `parses` / `instances`。 |

### Parse Rule Lifecycle

| File | Behavior |
|---|---|
| `backend/pkg/services/log/service_get_log_parse_rule.go` | 读取 `logtableinfo`；若记录不存在，返回默认 `all_logs` 规则。 |
| `backend/pkg/services/log/service_add_log_parse_rule.go` | 创建 ClickHouse 表；更新 Vector ConfigMap；写入 `logtableinfo`。非结构化规则会从 regex 中推导字段。 |
| `backend/pkg/services/log/service_update_log_parse_rule.go` | 更新 ClickHouse 表结构；更新 Vector ConfigMap；更新 `logtableinfo`。 |
| `backend/pkg/services/log/service_delete_log_parse_rule.go` | 删除 Vector 配置；删除 ClickHouse 表；删除 `logtableinfo` 记录。 |
| `backend/pkg/services/log/service_get_service_route.go` | 根据最近 7 天活跃实例名推导 `k8s.pod.name` 的前缀集合。 |
| `backend/pkg/services/log/vector/factory.go` | 直接编辑 `route_logs`、`parse_<name>`、`to_<name>` 等 Vector 配置节点。 |

### External Table Lifecycle

| File | Behavior |
|---|---|
| `backend/pkg/services/log/service_other_log.go` | 从 ClickHouse 枚举所有可见数据库与表。 |
| `backend/pkg/services/log/service_other_log_info.go` | 查询指定表的列名与列类型。 |
| `backend/pkg/services/log/service_add_other_table.go` | 只把登记信息写入 `otherlogtable`，不创建真实表。 |
| `backend/pkg/services/log/service_delete_other_table.go` | 从 `otherlogtable` 删除登记信息。 |

## Models And Schemas

| Type | File | Notes |
|---|---|---|
| Request | `backend/pkg/model/request/log_req.go` | `GetFaultLogPageListRequest`、`GetFaultLogContentRequest`。 |
| Request | `backend/pkg/model/request/log_query.go` | `LogQueryRequest`、`LogIndexRequest`、`LogQueryContextRequest`。 |
| Request | `backend/pkg/model/request/log_parse.go` | 解析规则相关请求。 |
| Request | `backend/pkg/model/request/log_other.go` | 外部表相关请求。 |
| Request | `backend/pkg/model/request/log_table.go` | 解析日志表结构和默认值。 |
| Response | `backend/pkg/model/response/log_resp.go` | 故障现场分页与内容响应。 |
| Response | `backend/pkg/model/response/log_query.go` | 全量查询与上下文响应。 |
| Response | `backend/pkg/model/response/log_chart.go` | 直方图响应。 |
| Response | `backend/pkg/model/response/log_index.go` | 字段索引响应。 |
| Response | `backend/pkg/model/response/log_parse.go` | 解析规则与 `routeRule` 响应。 |
| Response | `backend/pkg/model/response/log_other.go` | 外部表列表、列定义与增删响应。 |

## Database Mapping

| Storage | Entity | Mapping |
|---|---|---|
| Relational DB | `logtableinfo` | `dao_logtableinfo.go` 中的 `LogTableInfo`，保存 `database`、`tablename`、`parsename`、`routerule`、`parserule`、`parseinfo`、`service`、`structured`。 |
| Relational DB | `otherlogtable` | `dao_othertableinfo.go` 中的 `OtherLogTable`，保存 `database`、`tablename`、`cluster`、`timefield`、`logfield`、`instance`。 |
| ClickHouse | `raw_logs` | 默认解析表初始化与故障现场 fallback 查询的基础表。 |
| ClickHouse | `logs_<parseName>` | 解析规则新增时创建的结构化日志表。 |
| Vector Sink | `logs_<parseName>_buffer` | `vector/factory.go` 中把 sink 目标表设置为 `<table>_buffer`。 |

## Validation And Rules

1. `/api/log` route group 统一套用了 `AuthMiddleware` 与 `LicMiddleware`。
2. `GetFaultLogPageList` 在 handler 层默认 `pageNum=1`、`pageSize=10`，并调用 `CheckGroupPermission`。
3. `QueryLog` 在 handler 层默认 `pageNum=1`、`pageSize=10`、`query="(1='1')"`、`timeField="timestamp"`、`logField="content"`。
4. `GetLogChart` 与 `GetLogIndex` 也会在 handler 层补全默认 `query`、`timeField`、`logField`。
5. `AddOtherTableRequest.FillerValue()` 会在未传值时填充默认 `cluster`、`database`、`instance`。
6. 解析规则新增/更新时，若 `isStructured == false`，后端会从 regex 的命名捕获组里推导字段定义；若为结构化日志，则忽略 `parseRule`。
7. `GetServiceRoute` 并不是直接查日志规则表，而是通过最近 7 天 Prometheus 活跃实例推导 pod 前缀。

## Error Handling

1. 请求绑定失败统一通过 `c.AbortWithError(http.StatusBadRequest, code.ParamBindError, err)` 返回。
2. `GetFaultLogPageList` 权限失败不会返回 HTTP 403 风格的空体，而是携带空 `pagination` 与空 `list` 的响应体。
3. `QueryLog` 在 service 层会把查询失败或无结果写入 `response.LogQueryResponse.Err`，然后仍返回 200 风格 payload。
4. `OtherTableInfo`、`AddOtherTable`、`DeleteOtherTable` 也使用响应体里的 `error` 字段表达部分错误，而不是一律抛 HTTP 错误。
5. 解析规则新增/更新/删除链路跨多个外部系统，若中途失败，当前代码未见统一补偿。

## Notes

1. `handler.New(...)` 在服务启动时执行 `InitParseLogTable(core.EmptyCtx(), req)`，说明默认日志表初始化不是懒加载行为。
2. `service_log_table.go` 中默认解析规则的 `defaultRouteRule` 固定为 `starts_with(string!(."k8s.pod.name"), "apo")`，具有明显环境假设。
3. `GetFaultLogContent` 的 TODO 已明确指出当前 `service` 未参与数据权限检查。
